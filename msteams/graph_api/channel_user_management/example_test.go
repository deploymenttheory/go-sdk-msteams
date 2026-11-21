package channelusermanagement_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_user_management"
)

// Compile-only examples borrow the application's configured client. Ordinary
// tests do not authenticate or change any tenant resource.
func ExampleService_Add() {
	addAndPromote := func(ctx context.Context, sdk *msteams.Client, teamID, channelName, user string) (management.ChangeResult, error) {
		added, err := sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, channelName, user, management.ChangeOptions{})
		if err != nil {
			return added, err
		}
		// Persist added.MembershipID. Promotion is a separate, non-atomic step;
		// if it fails, membership created by the first step is still retained.
		return sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, channelName, user, management.ChangeOptions{Role: management.Owner})
	}
	_ = addAndPromote
}

func ExampleService_List() {
	listOwners := func(ctx context.Context, sdk *msteams.Client, teamID, name string) (management.ListResult, error) {
		return sdk.GraphAPI.ChannelUserManagement.List(ctx, teamID, name, management.ListOptions{
			Role: management.Owner, Traversal: client.PagingOptions{MaxPages: 16, MaxItems: 1000}, UserLookupMaxPages: 8,
		})
	}
	_ = listOwners
}
