package associatedteam_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	associated "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/associated_team"
)

func ExampleService_ListForUser() {
	list := func(ctx context.Context, sdk *msteams.Client, userID string) (client.ListResult[associated.Team], error) {
		return sdk.GraphAPI.AssociatedTeams.ListForUser(ctx, userID, client.PagingOptions{MaxPages: 16})
	}
	_ = list
}
