package channel_test

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
)

func ExampleUpdateRequest() {
	empty := ""
	patch := channel.UpdateRequest{Description: &empty}
	fmt.Println(patch.DisplayName == nil, patch.Description != nil)
	// Output: true true
}

// Compile-only: the application supplies its configured SDK client and owner.
func ExampleService_Create() {
	create := func(ctx context.Context, sdk *msteams.Client, teamID, ownerID string) (channel.CreateResult, error) {
		result, err := sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{
			DisplayName: "Project", MembershipType: channel.Private, OwnerUserID: &ownerID,
		})
		// Persist result.Channel.ID when present. Shared creation instead returns
		// an accepted Operation, which the application may explicitly Wait on.
		return result, err
	}
	_ = create
}

func ExampleService_ListComplete() {
	list := func(ctx context.Context, sdk *msteams.Client, teamID string) (client.ListResult[channel.Channel], error) {
		return sdk.GraphAPI.Channels.ListComplete(ctx, teamID, channel.ListOptions{},
			client.PagingOptions{MaxPages: 20, MaxItems: 1000})
	}
	_ = list
}

// Compile-only: initial sharing is part of a new shared channel, not an update
// to an existing sharing relationship. The host and owner are caller-selected.
func ExampleService_Create_sharedWithHost() {
	create := func(ctx context.Context, sdk *msteams.Client, teamID, ownerID string) (channel.CreateResult, error) {
		shares := []channel.InitialSharedTeam{{ID: teamID}}
		return sdk.GraphAPI.Channels.Create(ctx, teamID, channel.CreateRequest{
			DisplayName: "Shared project", MembershipType: channel.Shared,
			OwnerUserID: &ownerID, SharedWithTeams: &shares,
		})
	}
	_ = create
}
