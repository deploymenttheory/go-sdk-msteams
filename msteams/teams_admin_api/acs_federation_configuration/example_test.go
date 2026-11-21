package acsfederationconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/acs_federation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Configuration, error) {
		v, _, e := sdk.TeamsAdminAPI.ACSFederationConfigurations.GetByIdentity(ctx, "Global")
		return v, e
	}
	_ = read
}
func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client) error {
		_, _, e := sdk.TeamsAdminAPI.ACSFederationConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Fields: resource.Fields{EnableAcsUsers: policy.Set(false)}, AllowedAcsResources: policy.ReplaceStrings()})
		return e
	}
	_ = update
}
