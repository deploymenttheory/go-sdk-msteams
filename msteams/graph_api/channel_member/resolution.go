package channelmember

import (
	"context"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrNotFound  = errors.New("channel membership for directory user was not found")
	ErrAmbiguous = errors.New("channel membership for directory user is ambiguous")
)

// ResolveByUser performs a bounded complete native membership traversal and
// retains the actual opaque membership ID. A nil tenant selector does not assume
// the host tenant; multiple matching users are ambiguous. A selected tenant
// requires a matching returned tenantId. No private membership ID is synthesized.
func (s *Service) ResolveByUser(ctx context.Context, teamID, channelID string, userID DirectoryUserID, tenantID *string, limits client.PagingOptions) (*Member, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !directoryID.MatchString(string(userID)) || tenantID != nil && !directoryID.MatchString(*tenantID) {
		return nil, nil, client.ErrRequest
	}
	ctx, stop, err := s.scope.Session().Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer stop()
	listed, err := s.ListAll(ctx, teamID, channelID, ListOptions{}, limits)
	if err != nil {
		return nil, listed.Response, err
	}
	var found *Member
	for _, member := range listed.Items {
		if member.UserID == nil || !strings.EqualFold(string(*member.UserID), string(userID)) {
			continue
		}
		if tenantID != nil && (member.TenantID == nil || !strings.EqualFold(*member.TenantID, *tenantID)) {
			continue
		}
		if found != nil {
			return nil, listed.Response, ErrAmbiguous
		}
		value := member
		found = &value
	}
	if found == nil {
		return nil, listed.Response, ErrNotFound
	}
	return found, listed.Response, nil
}
