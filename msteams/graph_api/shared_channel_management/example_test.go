package sharedchannelmanagement_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/shared_channel_management"
)

func ExampleService_ListAllowedUsers() {
	owners := func(ctx context.Context, sdk *msteams.Client, host, channel, sharedTeam string) (management.AllowedUsersResult, error) {
		return sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, host, channel, sharedTeam, management.AllowedUsersOptions{Role: management.Owner, Traversal: client.PagingOptions{MaxPages: 16, MaxItems: 1000}, UserLookupMaxPages: 8})
	}
	_ = owners
}
