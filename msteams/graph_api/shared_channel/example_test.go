package sharedchannel_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	share "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/shared_channel"
)

// Compile-only examples borrow the application's configured client and perform
// no authentication or tenant operation during ordinary tests.
func ExampleService_ListAll() {
	list := func(ctx context.Context, sdk *msteams.Client, host, channel string) (client.ListResult[share.SharedWithTeam], error) {
		return sdk.GraphAPI.ChannelShares.ListAll(ctx, host, channel, client.PagingOptions{MaxPages: 16, MaxItems: 1000})
	}
	_ = list
}
