package sharedchannel

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	"github.com/go-resty/resty/v2"
)

var (
	ErrUnexpectedStatus = errors.New("shared-channel relationship returned an unexpected successful HTTP status")
	directoryID         = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// Service owns native sharing reads and relationship deletion. Create/update
// remain unsettled service contracts; no speculative request or stub is exposed.
type Service struct{ scope *client.Scope }

// NewService constructs a local resource dependency without credentials or HTTP.
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

func request(method, hostTeamID, channelID, sharedTeamID string) (client.Request, error) {
	if !directoryID.MatchString(hostTeamID) || strings.TrimSpace(channelID) == "" || channelID == "." || channelID == ".." || strings.ContainsRune(channelID, '\\') || strings.ContainsFunc(channelID, unicode.IsControl) {
		return client.Request{}, client.ErrRequest
	}
	path := []string{"beta", "teams", hostTeamID, "channels", channelID, "sharedWithTeams"}
	if sharedTeamID != "" {
		if !directoryID.MatchString(sharedTeamID) {
			return client.Request{}, client.ErrRequest
		}
		path = append(path, sharedTeamID)
	}
	r := client.Request{Method: method, Path: path, Header: http.Header{"Accept": {"application/json"}}}
	if method == http.MethodGet {
		r.Replay = client.ReplaySafe
	}
	return r, nil
}

// GetByID reads one shared-with-team relationship. Host identity is explicit;
// the requested recipient/shared-with ID is never substituted as the host.
func (s *Service) GetByID(ctx context.Context, hostTeamID, channelID, sharedTeamID string) (*SharedWithTeam, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if sharedTeamID == "" {
		return nil, nil, client.ErrRequest
	}
	req, err := request(http.MethodGet, hostTeamID, channelID, sharedTeamID)
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
	var value SharedWithTeam
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// DeleteByID unshares the channel with the addressed team once. This deletes
// the sharing relationship, not a directory user, target team or host channel.
func (s *Service) DeleteByID(ctx context.Context, hostTeamID, channelID, sharedTeamID string) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if sharedTeamID == "" {
		return nil, client.ErrRequest
	}
	req, err := request(http.MethodDelete, hostTeamID, channelID, sharedTeamID)
	if err != nil {
		return nil, err
	}
	response, err := s.scope.Do(ctx, req, nil)
	if err == nil && response.StatusCode() != http.StatusNoContent {
		err = ErrUnexpectedStatus
	}
	return response, err
}

func decodeShares(data []byte) (client.PageContent[SharedWithTeam], error) {
	return client.DecodeCollection[SharedWithTeam](data, "value", "@odata.nextLink")
}
func decodeMembers(data []byte) (client.PageContent[member.Member], error) {
	return client.DecodeCollection[member.Member](data, "value", "@odata.nextLink")
}

func limits(value client.PagingOptions) (client.PagingOptions, error) {
	if value.Style != 0 && value.Style != client.LinkPaging || value.Parameter != "" || value.ExpectedStatus != 0 && value.ExpectedStatus != http.StatusOK {
		return value, client.ErrPagination
	}
	value.Style, value.ExpectedStatus = client.LinkPaging, http.StatusOK
	return value, nil
}

// List reads a single native relationship page and preserves its continuation.
func (s *Service) List(ctx context.Context, hostTeamID, channelID string) (*ListPage, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	req, err := request(http.MethodGet, hostTeamID, channelID, "")
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
	content, err := decodeShares(response.Body())
	if err != nil {
		return nil, response, err
	}
	page := &ListPage{Items: content.Items, NextLink: content.Next}
	if json.Unmarshal(response.Body(), &page.RawFields) != nil {
		return nil, response, client.ErrDecode
	}
	return page, response, nil
}

// NewPager constructs a lazy bounded relationship traversal without I/O.
func (s *Service) NewPager(hostTeamID, channelID string, bounds client.PagingOptions) (*client.Pager[SharedWithTeam], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	req, err := request(http.MethodGet, hostTeamID, channelID, "")
	if err != nil {
		return nil, err
	}
	bounds, err = limits(bounds)
	if err != nil {
		return nil, err
	}
	return client.NewPager(s.scope, req, decodeShares, bounds)
}

// ListAll collects the complete bounded sharing list. Failure retains partial
// items and the last response with a nonnil error.
func (s *Service) ListAll(ctx context.Context, hostTeamID, channelID string, bounds client.PagingOptions) (client.ListResult[SharedWithTeam], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[SharedWithTeam]{}, err
	}
	pager, err := s.NewPager(hostTeamID, channelID, bounds)
	if err != nil {
		return client.ListResult[SharedWithTeam]{}, err
	}
	return pager.Collect(ctx)
}

// NewAllowedMembersPager addresses the exact shared-with-team allowedMembers
// collection. It does not substitute a direct channel roster or target team roster.
func (s *Service) NewAllowedMembersPager(hostTeamID, channelID, sharedTeamID string, bounds client.PagingOptions) (*client.Pager[member.Member], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	if sharedTeamID == "" {
		return nil, client.ErrRequest
	}
	req, err := request(http.MethodGet, hostTeamID, channelID, sharedTeamID)
	if err != nil {
		return nil, err
	}
	req.Path = append(req.Path, "allowedMembers")
	bounds, err = limits(bounds)
	if err != nil {
		return nil, err
	}
	return client.NewPager(s.scope, req, decodeMembers, bounds)
}

// ListAllowedMembers collects native access records for the addressed sharing
// relationship. It preserves returned membership/user/tenant
// identities without inferring that the members belong to the host tenant.
func (s *Service) ListAllowedMembers(ctx context.Context, hostTeamID, channelID, sharedTeamID string, bounds client.PagingOptions) (client.ListResult[member.Member], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[member.Member]{}, err
	}
	pager, err := s.NewAllowedMembersPager(hostTeamID, channelID, sharedTeamID, bounds)
	if err != nil {
		return client.ListResult[member.Member]{}, err
	}
	return pager.Collect(ctx)
}
