package region_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/region"
)

func ExampleService_GetForUser() {
	read := func(ctx context.Context, sdk *msteams.Client, userID string) (*region.Location, error) {
		value, _, err := sdk.TeamsAdminAPI.Regions.GetForUser(ctx, userID)
		return value, err
	}
	_ = read
}
