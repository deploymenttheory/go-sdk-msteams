package grouprelationship

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// OwnerService manages backing-group owners independently of member references.
// A successful owner write does not establish membership or Teams synchronization.
type OwnerService struct{ relation *relationship }

// NewOwnerService binds a lazy owner relationship service to Graph.
func NewOwnerService(scope *client.Scope) *OwnerService {
	return &OwnerService{relation: &relationship{scope: scope, name: "owners", defaultKind: Users}}
}

func (s *OwnerService) resource() *relationship {
	if s == nil {
		return nil
	}
	return s.relation
}

// AddReference adds an owner using the documented users binding by default. An
// explicit Kind selects a directory-object or service-principal binding instead.
func (s *OwnerService) AddReference(ctx context.Context, groupID string, value Reference) (*resty.Response, error) {
	return s.resource().addReference(ctx, groupID, value)
}

// RemoveReference removes only ownership. Graph enforces its last-user-owner
// restriction; this primitive does not fake a 412 response or remove membership.
func (s *OwnerService) RemoveReference(ctx context.Context, groupID string, id DirectoryObjectID) (*resty.Response, error) {
	return s.resource().removeReference(ctx, groupID, id)
}

// List reads one native owners page. Limited data and service-principal omission
// in the v1.0 endpoint are not evidence of an ordinary user or complete ownership.
func (s *OwnerService) List(ctx context.Context, groupID string, options ListOptions) (*ListPage, *resty.Response, error) {
	return s.resource().list(ctx, groupID, options)
}

// NewPager constructs a lazy same-origin traversal without credentials or HTTP.
func (s *OwnerService) NewPager(groupID string, options ListOptions, limits client.PagingOptions) (*client.Pager[DirectoryObject], error) {
	return s.resource().newPager(groupID, options, limits)
}

// ListAll explicitly traverses the endpoint's pages, preserving a partial prefix
// and last response on error. It cannot correct backend object-type omissions.
func (s *OwnerService) ListAll(ctx context.Context, groupID string, options ListOptions, limits client.PagingOptions) (client.ListResult[DirectoryObject], error) {
	return s.resource().listAll(ctx, groupID, options, limits)
}
