package teamsshiftsapppolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shifts_app_policy"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Policy, error) {
		value, _, err := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.GetByIdentity(ctx, "Global")
		return value, err
	}
	_ = read
}
