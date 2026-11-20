package team

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrUnexpectedStatus = errors.New("team operation returned an unexpected successful HTTP status")
	ErrArchiveOption    = errors.New("SharePoint archive option requires delegated authentication")
	errModel            = errors.New("team response has no identity")
	directoryID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	bindingID           = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Service manages native Graph teams. It uses the same session as the root
// client; construction acquires no credentials and starts no requests.
type Service struct{ scope *client.Scope }

func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, err := s.scope.Session().Info()
	return err
}

func request(method string, path ...string) client.Request {
	return client.Request{Method: method, Path: append([]string{"v1.0"}, path...),
		Header: http.Header{"Accept": {"application/json"}}}
}

func (s *Service) do(ctx context.Context, req client.Request, status int) (*resty.Response, error) {
	response, err := s.scope.Do(ctx, req, nil)
	if err == nil && response.StatusCode() != status {
		err = ErrUnexpectedStatus
	}
	return response, err
}

// GetByID reads one native team. Group properties are not separately hydrated.
func (s *Service) GetByID(ctx context.Context, id string) (*Team, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !directoryID.MatchString(id) {
		return nil, nil, client.ErrRequest
	}
	req := request(http.MethodGet, "teams", id)
	req.Replay = client.ReplaySafe
	response, err := s.do(ctx, req, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeTeam(response.Body())
	return value, response, err
}

// UpdateByID sparsely patches settings. Explicit false/empty containers survive.
// It does not patch the backing group or perform a follow-up read.
func (s *Service) UpdateByID(ctx context.Context, id string, settings Settings) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !directoryID.MatchString(id) {
		return nil, client.ErrRequest
	}
	req := request(http.MethodPatch, "teams", id)
	req.Body = settings
	return s.do(ctx, req, http.StatusNoContent)
}

// DeleteByID deletes the backing Microsoft 365 group, including the team and its
// associated group resources. There is no native DELETE /teams endpoint. The
// caller owns lifecycle recovery; a failed/ambiguous delete is never replayed.
func (s *Service) DeleteByID(ctx context.Context, id string) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !directoryID.MatchString(id) {
		return nil, client.ErrRequest
	}
	return s.do(ctx, request(http.MethodDelete, "groups", id), http.StatusNoContent)
}

// CreateFromGroup teamifies an existing group with an owner. A successful 201
// returns a team; SharePoint provisioning may still be incomplete. Group
// replication errors are returned directly. No group is created/deleted, no
// propagation retry is hidden, and no async-create contract is inferred here.
func (s *Service) CreateFromGroup(ctx context.Context, groupID string, settings Settings) (*Team, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !directoryID.MatchString(groupID) {
		return nil, nil, client.ErrRequest
	}
	req := request(http.MethodPut, "groups", groupID, "team")
	req.Body = settings
	response, err := s.do(ctx, req, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeTeam(response.Body())
	return value, response, err
}

// ListOptions maps only the documented native teams collection parameters.
// Filter is an OData expression, not a display-name search shortcut. Limits
// bound traversal without reporting a truncated result as complete.
type ListOptions struct {
	Filter    string
	Select    []string
	Top       int
	SkipToken string
	Count     *bool
	MaxPages  int
	MaxItems  int
}

// NewPager constructs a lazy, bounded native list. It acquires no token. Only
// basic properties are populated by Graph; use GetByID for individual settings.
func (s *Service) NewPager(options ListOptions) (*client.Pager[Team], error) {
	if err := s.ready(context.Background()); err != nil {
		return nil, err
	}
	if options.Top < 0 {
		return nil, client.ErrRequest
	}
	query := make(url.Values)
	for key, value := range map[string]string{"$filter": options.Filter, "$skiptoken": options.SkipToken} {
		if value != "" {
			query.Set(key, value)
		}
	}
	if len(options.Select) != 0 {
		for _, field := range options.Select {
			if strings.TrimSpace(field) == "" || strings.ContainsAny(field, ",\r\n") {
				return nil, client.ErrRequest
			}
		}
		query.Set("$select", strings.Join(options.Select, ","))
	}
	if options.Top != 0 {
		query.Set("$top", strconv.Itoa(options.Top))
	}
	if options.Count != nil {
		query.Set("$count", strconv.FormatBool(*options.Count))
	}
	req := request(http.MethodGet, "teams")
	req.Query, req.Replay = query, client.ReplaySafe
	return client.NewPager(s.scope, req, decodePage,
		client.PagingOptions{Style: client.LinkPaging, MaxPages: options.MaxPages, MaxItems: options.MaxItems, ExpectedStatus: http.StatusOK})
}

// List collects all native pages within the supplied bounds. On error its
// result retains partial items and the last response; it is not a complete list.
func (s *Service) List(ctx context.Context, options ListOptions) (client.ListResult[Team], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Team]{}, err
	}
	//nolint:contextcheck // Pager construction is local; Collect below receives the operation context.
	pager, err := s.NewPager(options)
	if err != nil {
		return client.ListResult[Team]{}, err
	}
	return pager.Collect(ctx)
}

func (s *Service) mode() (auth.Mode, error) {
	info, err := s.scope.Session().Info()
	return info.Mode, err
}
