package channel

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrUnexpectedStatus = errors.New("channel returned an unexpected successful HTTP status")
	ErrProjection       = errors.New("channel host identity is unavailable or inconsistent")
	ErrNotFound         = errors.New("channel display name was not found")
	ErrAmbiguous        = errors.New("channel display name is ambiguous")
	directoryID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// Service manages native channel CRUD and collection operations. Construction
// acquires no credentials and performs no discovery or HTTP requests.
type Service struct{ scope *client.Scope }

// NewService binds the resource to one root client's Graph scope.
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

func (s *Service) begin(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.scope.Session().Begin(ctx)
}

func request(method, teamID, channelID string, collection Collection) (client.Request, error) {
	if !directoryID.MatchString(teamID) {
		return client.Request{}, client.ErrRequest
	}
	if collection == "" {
		collection = Hosted
	}
	if collection != Hosted && collection != All && collection != Incoming {
		return client.Request{}, client.ErrRequest
	}
	path := []string{"beta", "teams", teamID, string(collection)}
	if channelID != "" {
		if strings.TrimSpace(channelID) == "" || channelID == "." || channelID == ".." || strings.ContainsRune(channelID, '\\') || strings.ContainsFunc(channelID, unicode.IsControl) {
			return client.Request{}, client.ErrRequest
		}
		path = append(path, channelID)
	}
	req := client.Request{Method: method, Path: path, Header: http.Header{"Accept": {"application/json"}}}
	if method == http.MethodGet {
		req.Replay = client.ReplaySafe
	}
	return req, nil
}

func readOptions(req *client.Request, options ReadOptions) error {
	req.Query = make(url.Values)
	if len(options.Select) != 0 {
		for _, field := range options.Select {
			if strings.TrimSpace(field) == "" || strings.ContainsAny(field, ",\r\n") {
				return client.ErrRequest
			}
		}
		req.Query.Set("$select", strings.Join(options.Select, ","))
	}
	if options.IncludeUnknownEnumMembers {
		req.Header.Set("Prefer", "include-unknown-enum-members")
	}
	return nil
}

func (s *Service) do(ctx context.Context, req client.Request, status int) (*resty.Response, error) {
	response, err := s.scope.Do(ctx, req, nil)
	if err == nil && response.StatusCode() != status {
		err = ErrUnexpectedStatus
	}
	return response, err
}

// GetByID reads one native channel. Its opaque ID is escaped as one segment;
// response web/OData URLs are retained as data and are never followed.
func (s *Service) GetByID(ctx context.Context, teamID, channelID string, options ReadOptions) (*Channel, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if channelID == "" {
		return nil, nil, client.ErrRequest
	}
	req, err := request(http.MethodGet, teamID, channelID, Hosted)
	if err != nil {
		return nil, nil, err
	}
	if err := readOptions(&req, options); err != nil {
		return nil, nil, err
	}
	response, err := s.do(ctx, req, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var value Channel
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// UpdateByID sends exactly the selected writable fields once. An empty patch
// remains {}, as observed in the pinned client; Graph decides its acceptance.
func (s *Service) UpdateByID(ctx context.Context, teamID, channelID string, value UpdateRequest) (*resty.Response, error) {
	return s.mutate(ctx, http.MethodPatch, teamID, channelID, value)
}

// DeleteByID removes only this channel. It never deletes its backing team.
func (s *Service) DeleteByID(ctx context.Context, teamID, channelID string) (*resty.Response, error) {
	return s.mutate(ctx, http.MethodDelete, teamID, channelID, nil)
}

func (s *Service) mutate(ctx context.Context, method, teamID, channelID string, value any) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if channelID == "" {
		return nil, client.ErrRequest
	}
	req, err := request(method, teamID, channelID, Hosted)
	if err != nil {
		return nil, err
	}
	req.Body = value
	return s.do(ctx, req, http.StatusNoContent)
}

func listRequest(teamID string, options ListOptions) (client.Request, error) {
	req, err := request(http.MethodGet, teamID, "", options.Collection)
	if err != nil {
		return client.Request{}, err
	}
	if err := readOptions(&req, options.ReadOptions); err != nil {
		return client.Request{}, err
	}
	if options.MembershipType != nil {
		if options.Collection == Incoming || !validType(*options.MembershipType) {
			return client.Request{}, client.ErrRequest
		}
		req.Query.Set("$filter", "membershipType eq '"+string(*options.MembershipType)+"'")
	}
	return req, nil
}

func decodePage(data []byte) (client.PageContent[Channel], error) {
	return client.DecodeCollection[Channel](data, "value", "@odata.nextLink")
}

// List reads one hosted/all/incoming page, preserving the untrusted next link.
func (s *Service) List(ctx context.Context, teamID string, options ListOptions) (*ListPage, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	req, err := listRequest(teamID, options)
	if err != nil {
		return nil, nil, err
	}
	response, err := s.do(ctx, req, http.StatusOK)
	if err != nil {
		return nil, response, err
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

// NewPager constructs a lazy origin-validated traversal. Only link paging and
// MaxPages/MaxItems bounds are allowed; it performs no token acquisition.
func (s *Service) NewPager(teamID string, options ListOptions, limits client.PagingOptions) (*client.Pager[Channel], error) {
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
	limits.Style, limits.ExpectedStatus = client.LinkPaging, http.StatusOK
	return client.NewPager(s.scope, req, decodePage, limits)
}

// ListComplete collects a bounded complete collection. On failure the result
// retains partial items and the last response; a nonnil error means incomplete.
func (s *Service) ListComplete(ctx context.Context, teamID string, options ListOptions, limits client.PagingOptions) (client.ListResult[Channel], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Channel]{}, err
	}
	pager, err := s.NewPager(teamID, options, limits)
	if err != nil {
		return client.ListResult[Channel]{}, err
	}
	return pager.Collect(ctx)
}

// ListAll includes hosted and incoming channels, using the distinct allChannels
// endpoint. This name describes the resource collection, not pagination.
func (s *Service) ListAll(ctx context.Context, teamID string, options ListOptions, limits client.PagingOptions) (client.ListResult[Channel], error) {
	options.Collection = All
	return s.ListComplete(ctx, teamID, options, limits)
}

// ListIncoming collects channels shared into this team from other host teams.
func (s *Service) ListIncoming(ctx context.Context, teamID string, options ReadOptions, limits client.PagingOptions) (client.ListResult[Channel], error) {
	return s.ListComplete(ctx, teamID, ListOptions{ReadOptions: options, Collection: Incoming}, limits)
}

func validType(value MembershipType) bool {
	return value == Standard || value == Private || value == Shared
}
