package teammember

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrUnexpectedStatus = errors.New("team membership returned an unexpected successful HTTP status")
	directoryID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// Service manages native conversationMember resources, independently of backing
// group relationships. Construction is local and acquires no credentials.
type Service struct{ scope *client.Scope }

// NewService binds the resource to the root client's Graph session.
func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

func (s *Service) local() error {
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, err := s.scope.Session().Info()
	return err
}

func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.local()
}

func memberRequest(method, teamID string, memberID MembershipID) (client.Request, error) {
	if !directoryID.MatchString(teamID) {
		return client.Request{}, client.ErrRequest
	}
	path := []string{"v1.0", "teams", teamID, "members"}
	if memberID != "" {
		id := string(memberID)
		if strings.TrimSpace(id) == "" || id == "." || id == ".." || strings.ContainsRune(id, '\\') || strings.ContainsFunc(id, unicode.IsControl) {
			return client.Request{}, client.ErrRequest
		}
		path = append(path, id)
	}
	req := client.Request{Method: method, Path: path, Header: http.Header{"Accept": {"application/json"}}}
	if method == http.MethodGet {
		req.Replay = client.ReplaySafe
	}
	return req, nil
}

func validRoles(roles *[]Role) bool {
	return roles == nil || *roles != nil && (len(*roles) == 0 || len(*roles) == 1 && (*roles)[0] == Owner)
}

func (s *Service) resource(ctx context.Context, req client.Request, expected int) (*Member, *resty.Response, error) {
	response, err := s.scope.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	if response.StatusCode() != expected {
		return nil, response, ErrUnexpectedStatus
	}
	var value Member
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// Create adds an AAD conversation member once. A nil roles pointer omits roles;
// an explicit empty array adds an ordinary member. Guest is server-assigned and
// cannot be supplied as a role. Application guest addition is unsupported by
// Graph; actual eligibility and access are enforced by the service, with no
// identity fallback or extra directory lookup. HTTP 201 is not an async handle.
func (s *Service) Create(ctx context.Context, teamID string, value CreateRequest) (*Member, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	req, err := memberRequest(http.MethodPost, teamID, "")
	if err != nil {
		return nil, nil, err
	}
	if !directoryID.MatchString(string(value.UserID)) || !validRoles(value.Roles) {
		return nil, nil, client.ErrRequest
	}
	environment, err := s.scope.Session().Environment()
	if err != nil {
		return nil, nil, err
	}
	req.Body = struct {
		Type  string  `json:"@odata.type"`
		Roles *[]Role `json:"roles,omitempty"`
		User  string  `json:"user@odata.bind"`
	}{Type: "#microsoft.graph.aadUserConversationMember", Roles: value.Roles,
		User: environment.Endpoints.Graph + "/v1.0/users('" + string(value.UserID) + "')"}
	return s.resource(ctx, req, http.StatusCreated)
}

// GetByID reads the native opaque membership identity. It neither accepts a
// directory-user selector nor decodes or synthesizes a membership ID.
func (s *Service) GetByID(ctx context.Context, teamID string, id MembershipID) (*Member, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if id == "" {
		return nil, nil, client.ErrRequest
	}
	req, err := memberRequest(http.MethodGet, teamID, id)
	if err != nil {
		return nil, nil, err
	}
	return s.resource(ctx, req, http.StatusOK)
}

// UpdateByID patches roles and returns the updated member. Omitted roles remain
// omitted; explicit [] demotes, ["owner"] promotes. Null, guest, member and
// unknown write qualifiers are rejected locally. Graph enforces guest promotion
// restrictions and the caller's authority; no pre-read or deletion is hidden.
func (s *Service) UpdateByID(ctx context.Context, teamID string, id MembershipID, value UpdateRequest) (*Member, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if id == "" || !validRoles(value.Roles) {
		return nil, nil, client.ErrRequest
	}
	req, err := memberRequest(http.MethodPatch, teamID, id)
	if err != nil {
		return nil, nil, err
	}
	req.Body = value
	return s.resource(ctx, req, http.StatusOK)
}

// DeleteByID removes only the addressed conversation membership. It never
// deletes a directory user, team, backing group or group relationship directly.
func (s *Service) DeleteByID(ctx context.Context, teamID string, id MembershipID) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, client.ErrRequest
	}
	req, err := memberRequest(http.MethodDelete, teamID, id)
	if err != nil {
		return nil, err
	}
	response, err := s.scope.Do(ctx, req, nil)
	if err == nil && response.StatusCode() != http.StatusNoContent {
		err = ErrUnexpectedStatus
	}
	return response, err
}

// PromoteOwner updates the member's role without changing its identity.
func (s *Service) PromoteOwner(ctx context.Context, teamID string, id MembershipID) (*Member, *resty.Response, error) {
	roles := []Role{Owner}
	return s.UpdateByID(ctx, teamID, id, UpdateRequest{Roles: &roles})
}

// DemoteOwner retains membership and explicitly removes the owner qualifier.
func (s *Service) DemoteOwner(ctx context.Context, teamID string, id MembershipID) (*Member, *resty.Response, error) {
	roles := []Role{}
	return s.UpdateByID(ctx, teamID, id, UpdateRequest{Roles: &roles})
}

// ListOptions selects the documented members collection query parameters.
// Zero Top accepts the server default of 100; the documented maximum is 999.
type ListOptions struct {
	Filter string
	Select []string
	Top    int
}

// ListPage is one native collection page. NextLink is opaque, untrusted metadata;
// returning it does not follow it or establish that the collection is complete.
type ListPage struct {
	Items     []Member
	NextLink  string
	RawFields map[string]json.RawMessage
}

func listRequest(teamID string, options ListOptions) (client.Request, error) {
	req, err := memberRequest(http.MethodGet, teamID, "")
	if err != nil {
		return client.Request{}, err
	}
	if options.Top < 0 || options.Top > 999 {
		return client.Request{}, client.ErrRequest
	}
	query := make(url.Values)
	if options.Filter != "" {
		query.Set("$filter", options.Filter)
	}
	if options.Top != 0 {
		query.Set("$top", strconv.Itoa(options.Top))
	}
	if len(options.Select) != 0 {
		for _, field := range options.Select {
			if strings.TrimSpace(field) == "" || strings.ContainsAny(field, ",\r\n") {
				return client.Request{}, client.ErrRequest
			}
		}
		query.Set("$select", strings.Join(options.Select, ","))
	}
	req.Query = query
	return req, nil
}

// List reads one page, retaining the next link without following it. It does
// not join backing-group owners or project PowerShell TeamUserSummary records.
func (s *Service) List(ctx context.Context, teamID string, options ListOptions) (*ListPage, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	req, err := listRequest(teamID, options)
	if err != nil {
		return nil, nil, err
	}
	response, err := s.scope.Do(ctx, req, nil)
	if err != nil {
		return nil, response, err
	}
	if response.StatusCode() != http.StatusOK {
		return nil, response, ErrUnexpectedStatus
	}
	content, err := decodePage(response.Body())
	if err != nil {
		return nil, response, err
	}
	page := &ListPage{Items: content.Items, NextLink: content.Next}
	if json.Unmarshal(response.Body(), &page.RawFields) != nil {
		return nil, response, client.ErrDecode
	}
	return page, response, nil
}

func decodePage(body []byte) (client.PageContent[Member], error) {
	return client.DecodeCollection[Member](body, "value", "@odata.nextLink")
}

// NewPager builds a lazy bounded traversal without credentials or HTTP. Style
// must be zero or LinkPaging and Parameter empty; only MaxPages/MaxItems apply.
// The shared pager validates every opaque next-link origin before acquisition.
func (s *Service) NewPager(teamID string, options ListOptions, limits client.PagingOptions) (*client.Pager[Member], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	req, err := listRequest(teamID, options)
	if err != nil {
		return nil, err
	}
	if limits.Style != 0 && limits.Style != client.LinkPaging || limits.Parameter != "" || limits.ExpectedStatus != 0 && limits.ExpectedStatus != http.StatusOK {
		return nil, client.ErrPagination
	}
	limits.Style = client.LinkPaging
	limits.ExpectedStatus = http.StatusOK
	return client.NewPager(s.scope, req, decodePage, limits)
}

// ListAll explicitly collects native pages. On failure or a traversal bound its
// partial items and last response remain available, with a non-nil error.
func (s *Service) ListAll(ctx context.Context, teamID string, options ListOptions, limits client.PagingOptions) (client.ListResult[Member], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Member]{}, err
	}
	pager, err := s.NewPager(teamID, options, limits)
	if err != nil {
		return client.ListResult[Member]{}, err
	}
	return pager.Collect(ctx)
}
