package teammember_test

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	teammember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_member"
)

func ExampleUpdateRequest() {
	roles := []teammember.Role{}
	demote := teammember.UpdateRequest{Roles: &roles}
	fmt.Println(demote.Roles != nil, len(*demote.Roles))
	// Output: true 0
}

// This compile-only example performs no authentication or tenant operations in
// ordinary tests. The application supplies an already configured client.
func ExampleService_Create() {
	addAndPromote := func(ctx context.Context, sdk *msteams.Client, teamID string, userID teammember.DirectoryUserID) error {
		member, _, err := sdk.GraphAPI.TeamMembers.Create(ctx, teamID, teammember.CreateRequest{UserID: userID})
		if err != nil {
			return err
		}
		// Preserve member.ID for recovery. It is the returned opaque membership
		// identity, not userID. Promotion is a separate, non-atomic stage.
		_, _, err = sdk.GraphAPI.TeamMembers.PromoteOwner(ctx, teamID, member.ID)
		return err
	}
	_ = addAndPromote
}

func ExampleService_ListAll() {
	list := func(ctx context.Context, sdk *msteams.Client, teamID string) (client.ListResult[teammember.Member], error) {
		// On error retain the partial result; never report it as complete.
		return sdk.GraphAPI.TeamMembers.ListAll(ctx, teamID, teammember.ListOptions{Top: 100}, client.PagingOptions{MaxPages: 10, MaxItems: 1000})
	}
	_ = list
}
