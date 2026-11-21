package grouprelationship_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	relation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
)

// Compile-only: caller supplies a configured client and directory identities.
// This intentionally exposes the completed member stage if ownership fails.
func ExampleMemberService_AddReference() {
	addMemberThenOwner := func(ctx context.Context, sdk *msteams.Client, groupID string, userID relation.DirectoryObjectID) (bool, error) {
		if _, err := sdk.GraphAPI.GroupMembers.AddReference(ctx, groupID, relation.Reference{ID: userID}); err != nil {
			return false, err
		}
		_, err := sdk.GraphAPI.GroupOwners.AddReference(ctx, groupID, relation.Reference{ID: userID})
		return true, err
	}
	_ = addMemberThenOwner
}
