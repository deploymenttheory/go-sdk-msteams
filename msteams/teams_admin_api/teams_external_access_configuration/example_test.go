package teamsexternalaccessconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_external_access_configuration"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Configuration, error) {
		v, _, e := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.GetByIdentity(ctx, "Global")
		return v, e
	}
	_ = read
}
func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client) error {
		_, _, e := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Fields: resource.Fields{BlockExternalUserAccess: policy.Null[bool]()}, BlockedUsers: policy.ReplaceStrings()})
		return e
	}
	_ = update
}
