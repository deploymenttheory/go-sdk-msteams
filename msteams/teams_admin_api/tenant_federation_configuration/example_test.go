package tenantfederationconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	tenant "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
)

func ExampleService_UpdateGlobal() {
	update := func(ctx context.Context, sdk *msteams.Client) error {
		no := false
		_, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.UpdateGlobal(ctx, tenant.UpdateRequest{Settings: tenant.Settings{AllowFederatedUsers: &no}, AllowedDomains: tenant.NewAllowList("example.invalid")})
		return e
	}
	_ = update
}
func ExampleNewAllowAllKnownDomains() { choice := tenant.NewAllowAllKnownDomains(); _ = choice }
