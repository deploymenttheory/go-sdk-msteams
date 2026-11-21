package grouprelationship

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// MemberService manages direct backing-group member references. It never
// substitutes native conversation membership operations or resolves user names.
type MemberService struct{ relation *relationship }

// NewMemberService binds a lazy member relationship service to Graph.
func NewMemberService(scope *client.Scope) *MemberService {
	return &MemberService{relation: &relationship{scope: scope, name: "members", defaultKind: DirectoryObjects}}
}

func (s *MemberService) resource() *relationship {
	if s == nil {
		return nil
	}
	return s.relation
}

// AddReference adds one directory relationship, returning the actual 204 outcome.
// A received 400 is not silently treated as a duplicate; no preflight or replay occurs.
func (s *MemberService) AddReference(ctx context.Context, groupID string, value Reference) (*resty.Response, error) {
	return s.resource().addReference(ctx, groupID, value)
}

// RemoveReference always deletes /members/{id}/$ref, never the directory object.
func (s *MemberService) RemoveReference(ctx context.Context, groupID string, id DirectoryObjectID) (*resty.Response, error) {
	return s.resource().removeReference(ctx, groupID, id)
}

// List reads one page of direct members, retaining limited/unknown object fields.
func (s *MemberService) List(ctx context.Context, groupID string, options ListOptions) (*ListPage, *resty.Response, error) {
	return s.resource().list(ctx, groupID, options)
}

// NewPager constructs a lazy bounded traversal without acquiring credentials.
func (s *MemberService) NewPager(groupID string, options ListOptions, limits client.PagingOptions) (*client.Pager[DirectoryObject], error) {
	return s.resource().newPager(groupID, options, limits)
}

// ListAll explicitly collects pages; a non-nil error leaves a partial prefix.
func (s *MemberService) ListAll(ctx context.Context, groupID string, options ListOptions, limits client.PagingOptions) (client.ListResult[DirectoryObject], error) {
	return s.resource().listAll(ctx, groupID, options, limits)
}
