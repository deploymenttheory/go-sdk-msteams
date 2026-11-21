package targetingpolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_policy"
)

func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client, identity string) error {
		enabled, empty := "Enabled", ""
		_, err := sdk.TeamsAdminAPI.TargetingPolicies.UpdateByIdentity(ctx, identity, policy.UpdateRequest{CustomTagsMode: &enabled, Description: &empty})
		return err
	}
	_ = update
}
