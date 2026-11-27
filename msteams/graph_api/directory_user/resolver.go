// Package directoryuser provides supporting directory reads and identity
// resolution for team compositions. It does not manage directory-user lifecycle.
package directoryuser

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrNotFound         = errors.New("directory user resolution found no match")
	ErrAmbiguous        = errors.New("directory user resolution found multiple identities")
	ErrCurrentUser      = errors.New("current-user lookup requires delegated authentication")
	ErrPending          = errors.New("directory user lookup was accepted but remains pending")
	ErrUnexpectedStatus = errors.New("directory user lookup returned an unexpected successful HTTP status")
	directoryID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// User is a supporting read snapshot. Nullable/future fields remain available;
// a read snapshot is never converted into a directory-user write request.
type User struct {
	ID                string                     `json:"id"`
	DisplayName       *string                    `json:"displayName,omitempty"`
	UserPrincipalName *string                    `json:"userPrincipalName,omitempty"`
	UserType          *string                    `json:"userType,omitempty"`
	Mail              *string                    `json:"mail,omitempty"`
	RawFields         map[string]json.RawMessage `json:"-"`
}

func (u *User) UnmarshalJSON(data []byte) error {
	type snapshot User
	var value snapshot
	if decodeUser(data, &value.ID, &value.DisplayName, &value.UserPrincipalName, &value.UserType, &value.Mail, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*u = User(value)
	return nil
}

// Source identifies how an ID was resolved. Identifier means caller-supplied,
// not verified existence, user type, membership or authority.
type Source string

const (
	Identifier Source = "identifier"
	Direct     Source = "direct"
	Mail       Source = "mail"
	Current    Source = "current"
)

// Resolution exposes a usable ID only on successful resolution. User may retain
// an accepted 202 snapshot for explicit inspection when ErrPending is returned.
type Resolution struct {
	ID     string
	Source Source
	User   *User
}

func (Resolution) String() string     { return "directory user resolution (redacted)" }
func (r Resolution) GoString() string { return r.String() }

// ResolveOptions permits an omitted selector to mean /me only when explicitly
// requested. MaxPages bounds mail fallback; zero uses 16, and partial traversal
// can never establish absence or uniqueness.
type ResolveOptions struct {
	AllowCurrent bool
	MaxPages     int
}

// ReadOptions selects the user properties returned by a supporting read.
type ReadOptions struct{ Select []string }

// Resolver is bound to one Graph scope. Construction is local and lazy; it
// retains no caller context, worker, arbitrary destination or directory cache.
type Resolver struct{ scope *client.Scope }

// NewResolver binds supporting reads to an existing Graph scope without I/O.
func NewResolver(scope *client.Scope) *Resolver { return &Resolver{scope: scope} }

func (r *Resolver) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.scope == nil || r.scope.Session() == nil {
		return client.ErrClosed
	}
	if r.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, err := r.scope.Session().Info()
	return err
}

func validSelector(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".." && !strings.ContainsRune(value, '\\') && !strings.ContainsFunc(value, unicode.IsControl)
}

func request(path ...string) client.Request {
	return client.Request{Method: http.MethodGet, Path: append([]string{"v1.0"}, path...), Replay: client.ReplaySafe, Header: http.Header{"Accept": {"application/json"}}}
}

func selectQuery(options ReadOptions) (url.Values, error) {
	query := make(url.Values)
	if len(options.Select) == 0 {
		return query, nil
	}
	for _, field := range options.Select {
		if strings.TrimSpace(field) == "" || strings.ContainsAny(field, ",\r\n") {
			return nil, client.ErrRequest
		}
	}
	query.Set("$select", strings.Join(options.Select, ","))
	return query, nil
}

func (r *Resolver) read(ctx context.Context, req client.Request) (*User, *resty.Response, error) {
	response, err := r.scope.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	if response.StatusCode() != 200 && response.StatusCode() != 202 {
		return nil, response, ErrUnexpectedStatus
	}
	var value *User
	if len(response.Body()) != 0 {
		var decoded User
		if json.Unmarshal(response.Body(), &decoded) != nil {
			return nil, response, client.ErrDecode
		}
		value = &decoded
	}
	if response.StatusCode() == 202 {
		return value, response, ErrPending
	}
	if value == nil {
		return nil, response, client.ErrDecode
	}
	return value, response, nil
}

// Get reads by directory ID or UPN. It uses the documented OData key form for
// a leading-$ UPN; other selectors are escaped as one path segment. It never
// falls back to mail implicitly. A 202 retains metadata/snapshot with ErrPending.
func (r *Resolver) Get(ctx context.Context, selector string, options ReadOptions) (*User, *resty.Response, error) {
	if err := r.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validSelector(selector) {
		return nil, nil, client.ErrRequest
	}
	req := request("users", selector)
	if strings.HasPrefix(selector, "$") {
		req = request("users('" + strings.ReplaceAll(selector, "'", "''") + "')")
	}
	query, err := selectQuery(options)
	if err != nil {
		return nil, nil, err
	}
	req.Query = query
	return r.read(ctx, req)
}

// Current reads /me only in configured delegated mode, without prompting or
// changing credentials. Applications must supply an explicit owner/user identity.
func (r *Resolver) Current(ctx context.Context, options ReadOptions) (*User, *resty.Response, error) {
	if err := r.ready(ctx); err != nil {
		return nil, nil, err
	}
	info, err := r.scope.Session().Info()
	if err != nil {
		return nil, nil, err
	}
	if info.Mode != auth.Delegated {
		return nil, nil, ErrCurrentUser
	}
	req := request("me")
	query, err := selectQuery(options)
	if err != nil {
		return nil, nil, err
	}
	req.Query = query
	return r.read(ctx, req)
}

// Resolve accepts an explicit UUID without reads. Otherwise it tries the direct
// UPN read; only a received 404 starts the canonical mail-filter lookup. It does
// not use the pinned myorganization alias or pick an arbitrary first mail match.
func (r *Resolver) Resolve(ctx context.Context, selector string, options ResolveOptions) (Resolution, *resty.Response, error) {
	if err := r.ready(ctx); err != nil {
		return Resolution{}, nil, err
	}
	if options.MaxPages < 0 {
		return Resolution{}, nil, client.ErrPagination
	}
	if selector == "" {
		if !options.AllowCurrent {
			return Resolution{}, nil, client.ErrRequest
		}
		user, response, err := r.Current(ctx, ReadOptions{Select: []string{"id"}})
		if err != nil {
			return Resolution{Source: Current, User: user}, response, err
		}
		resolved, resolveErr := resolution(Current, user)
		return resolved, response, resolveErr
	}
	if !validSelector(selector) {
		return Resolution{}, nil, client.ErrRequest
	}
	if directoryID.MatchString(selector) {
		return Resolution{ID: selector, Source: Identifier}, nil, nil
	}
	user, response, err := r.Get(ctx, selector, ReadOptions{Select: []string{"id"}})
	if err == nil {
		resolved, resolveErr := resolution(Direct, user)
		return resolved, response, resolveErr
	}
	if errors.Is(err, ErrPending) {
		return Resolution{Source: Direct, User: user}, response, err
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		return Resolution{}, response, err
	}
	result, fallbackResponse, fallbackErr := r.byMail(ctx, selector, options.MaxPages)
	if fallbackResponse == nil {
		fallbackResponse = response
	}
	return result, fallbackResponse, fallbackErr
}

func resolution(source Source, user *User) (Resolution, error) {
	result := Resolution{Source: source, User: user}
	if user != nil {
		result.ID = user.ID
	}
	if result.ID == "" {
		return result, client.ErrDecode
	}
	return result, nil
}

func (r *Resolver) byMail(ctx context.Context, selector string, maxPages int) (Resolution, *resty.Response, error) {
	if maxPages == 0 {
		maxPages = 16
	}
	req := request("users")
	req.Query = url.Values{"$filter": {"mail eq '" + strings.ReplaceAll(selector, "'", "''") + "'"}, "$select": {"id"}, "$top": {"2"}}
	pager, err := client.NewPager(r.scope, req, func(body []byte) (client.PageContent[User], error) {
		return client.DecodeCollection[User](body, "value", "@odata.nextLink")
	}, client.PagingOptions{Style: client.LinkPaging, MaxPages: maxPages, ExpectedStatus: http.StatusOK})
	if err != nil {
		return Resolution{}, nil, err
	}
	var selected *User
	var response *resty.Response
	for {
		page, ok, err := pager.Next(ctx)
		if page.Response != nil {
			response = page.Response
		}
		if err != nil {
			return Resolution{}, response, err
		}
		if !ok {
			break
		}
		for _, user := range page.Items {
			if user.ID == "" {
				return Resolution{}, response, client.ErrDecode
			}
			if selected == nil {
				copy := user
				selected = &copy
				continue
			}
			// UUID text has case-insensitive spelling; opaque IDs have no
			// evidenced normalization rule and must match exactly.
			if selected.ID != user.ID && !(directoryID.MatchString(selected.ID) && directoryID.MatchString(user.ID) && strings.EqualFold(selected.ID, user.ID)) {
				return Resolution{}, response, ErrAmbiguous
			}
		}
	}
	if selected == nil {
		return Resolution{}, response, ErrNotFound
	}
	resolved, resolveErr := resolution(Mail, selected)
	return resolved, response, resolveErr
}
