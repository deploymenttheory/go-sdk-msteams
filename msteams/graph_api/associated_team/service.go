// Package associatedteam provides the user's associated-team view, including
// teams associated through shared channels. Team lifecycle uses its own service.
package associatedteam

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrCurrentUser = errors.New("associated-team current-user view requires delegated authentication")
var directoryID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Team preserves the associated-team identity, host tenant and future metadata.
type Team struct {
	ID          string                     `json:"id"`
	DisplayName *string                    `json:"displayName,omitempty"`
	TenantID    *string                    `json:"tenantId,omitempty"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *Team) UnmarshalJSON(data []byte) error {
	type snapshot Team
	var value snapshot
	if json.Unmarshal(data, &value) != nil || strings.TrimSpace(value.ID) == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*v = Team(value)
	return nil
}

type Service struct{ scope *client.Scope }

// NewService binds local supporting reads to one Graph scope.
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

func decode(data []byte) (client.PageContent[Team], error) {
	return client.DecodeCollection[Team](data, "value", "@odata.nextLink")
}

// NewPager builds a lazy explicit-user traversal. A current-user view requires
// the separate NewMinePager operation; empty IDs are not silently substituted.
func (s *Service) NewPager(userID string, bounds client.PagingOptions) (*client.Pager[Team], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	if userID == "" {
		return nil, client.ErrRequest
	}
	return s.pager(userID, bounds)
}

// NewMinePager explicitly selects the delegated current user's associated teams.
func (s *Service) NewMinePager(bounds client.PagingOptions) (*client.Pager[Team], error) {
	return s.pager("", bounds)
}

func (s *Service) pager(userID string, bounds client.PagingOptions) (*client.Pager[Team], error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	path := []string{"beta", "me", "teamwork", "associatedTeams"}
	if userID == "" {
		info, err := s.scope.Session().Info()
		if err != nil {
			return nil, err
		}
		if info.Mode != auth.Delegated {
			return nil, ErrCurrentUser
		}
	} else {
		if !directoryID.MatchString(userID) {
			return nil, client.ErrRequest
		}
		path = []string{"beta", "users", userID, "teamwork", "associatedTeams"}
	}
	if bounds.Style != 0 && bounds.Style != client.LinkPaging || bounds.Parameter != "" || bounds.ExpectedStatus != 0 && bounds.ExpectedStatus != http.StatusOK {
		return nil, client.ErrPagination
	}
	bounds.Style, bounds.ExpectedStatus = client.LinkPaging, http.StatusOK
	req := client.Request{Method: http.MethodGet, Path: path, Replay: client.ReplaySafe, Header: http.Header{"Accept": {"application/json"}}}
	return client.NewPager(s.scope, req, decode, bounds)
}

// ListMine completely traverses the caller's associated-team view. Application
// sessions fail locally before credential acquisition or any request.
func (s *Service) ListMine(ctx context.Context, bounds client.PagingOptions) (client.ListResult[Team], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Team]{}, err
	}
	pager, err := s.NewMinePager(bounds)
	if err != nil {
		return client.ListResult[Team]{}, err
	}
	return pager.Collect(ctx)
}

// ListForUser addresses an explicit directory user UUID in this Graph scope.
// UUID input establishes a selector, not verified existence or membership.
func (s *Service) ListForUser(ctx context.Context, userID string, bounds client.PagingOptions) (client.ListResult[Team], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Team]{}, err
	}
	if userID == "" {
		return client.ListResult[Team]{}, client.ErrRequest
	}
	return s.list(ctx, userID, bounds)
}

func (s *Service) list(ctx context.Context, userID string, bounds client.PagingOptions) (client.ListResult[Team], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Team]{}, err
	}
	pager, err := s.NewPager(userID, bounds)
	if err != nil {
		return client.ListResult[Team]{}, err
	}
	return pager.Collect(ctx)
}
