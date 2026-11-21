package externalaccesspolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Policy, error) {
		v, _, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.GetByIdentity(ctx, "Global")
		return v, e
	}
	_ = read
}
func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client) error {
		_, _, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Fields: resource.Fields{EnableFederationAccess: policy.Set(false)}})
		return e
	}
	_ = update
}
