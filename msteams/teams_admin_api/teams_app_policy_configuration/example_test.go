package teamsapppolicyconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	config "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_policy_configuration"
)

// This compiles a caller-owned workflow without contacting a tenant in go test.
func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*config.Configuration, error) {
		value, _, err := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.GetByIdentity(ctx, "Global")
		return value, err
	}
	_ = read
}

// The caller must explicitly authorize a tenant-wide setting change. The
// closure is not invoked during go test; HTTP acceptance is not persistence.
func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client) (*config.UpdateResult, error) {
		result, _, err := sdk.TeamsAdminAPI.TeamsAppPolicyConfigurations.UpdateByIdentity(ctx, "Global", config.UpdateRequest{Settings: config.Settings{AppCatalogURI: policy.Null[string]()}})
		return result, err
	}
	_ = update
}
