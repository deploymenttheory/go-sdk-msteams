package channelmember_test

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	channelmember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
)

func ExampleUpdateRequest() {
	roles := []channelmember.Role{}
	demote := channelmember.UpdateRequest{Roles: &roles}
	fmt.Println(demote.Roles != nil, len(*demote.Roles))
	// Output: true 0
}

// This compile-only example performs no authentication or tenant operations in
// ordinary tests. The application supplies an already configured client.
func ExampleService_Create() {
	addAndPromote := func(ctx context.Context, sdk *msteams.Client, teamID, channelID string, userID channelmember.DirectoryUserID) (channelmember.MembershipID, error) {
		member, _, err := sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, channelID, channelmember.CreateRequest{UserID: userID})
		if err != nil {
			return "", err
		}
		// Preserve member.ID for recovery. It is the returned opaque membership
		// identity, not userID. Promotion is a separate, non-atomic stage.
		_, _, err = sdk.GraphAPI.ChannelMembers.PromoteOwner(ctx, teamID, channelID, member.ID)
		return member.ID, err
	}
	_ = addAndPromote
}

func ExampleService_ListAll() {
	list := func(ctx context.Context, sdk *msteams.Client, teamID, channelID string) (client.ListResult[channelmember.Member], error) {
		// On error retain the partial result; never report it as complete.
		return sdk.GraphAPI.ChannelMembers.ListAll(ctx, teamID, channelID, channelmember.ListOptions{Top: 100}, client.PagingOptions{MaxPages: 10, MaxItems: 1000})
	}
	_ = list
}
