package teamsupgradeconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_configuration"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Configuration, error) {
		value, _, err := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations.GetByIdentity(ctx, "Global")
		return value, err
	}
	_ = read
}
