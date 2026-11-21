package privatechannelmigration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/private_channel_migration"
)

func ExampleService_Get() {
	read := func(ctx context.Context, sdk *msteams.Client) (*migration.Status, error) {
		value, _, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
		return value, err
	}
	_ = read
}
