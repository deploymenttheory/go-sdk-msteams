package team

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// NewJoinedPager constructs a lazy traversal of one directory user's direct
// team memberships. This API does not support initial OData query parameters.
// Its projections remain distinct from full team settings and shared-channel
// associated teams. User ID must be a directory UUID, not a membership ID.
func (s *Service) NewJoinedPager(userID string, limits client.PagingOptions) (*client.Pager[Team], error) {
	if err := s.ready(context.Background()); err != nil {
		return nil, err
	}
	if !directoryID.MatchString(userID) {
		return nil, client.ErrRequest
	}
	if limits.Style != 0 && limits.Style != client.LinkPaging || limits.Parameter != "" || limits.ExpectedStatus != 0 && limits.ExpectedStatus != http.StatusOK {
		return nil, client.ErrPagination
	}
	limits.Style, limits.ExpectedStatus = client.LinkPaging, http.StatusOK
	req := request(http.MethodGet, "users", userID, "joinedTeams")
	req.Replay = client.ReplaySafe
	return client.NewPager(s.scope, req, decodePage, limits)
}

// ListJoined collects all direct joined-team pages within explicit bounds. A
// failure retains partial basic projections; it never reports a complete list.
// Supporting UPN/current-user resolution is a separate operation/composition.
func (s *Service) ListJoined(ctx context.Context, userID string, limits client.PagingOptions) (client.ListResult[Team], error) {
	if err := s.ready(ctx); err != nil {
		return client.ListResult[Team]{}, err
	}
	//nolint:contextcheck // Pager construction is local; Collect receives the operation context.
	pager, err := s.NewJoinedPager(userID, limits)
	if err != nil {
		return client.ListResult[Team]{}, err
	}
	return pager.Collect(ctx)
}
