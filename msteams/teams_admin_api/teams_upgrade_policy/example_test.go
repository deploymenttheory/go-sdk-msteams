package teamsupgradepolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_policy"
)

func ExampleService_Assign() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Omitting MigrateMeetingsToTeams avoids the pinned modern binding defect.
	_, _, _ = sdk.TeamsAdminAPI.TeamsUpgradePolicies.Assign(context.Background(), resource.Assignment{User: &resource.UserAssignment{Identity: "sdk-owned-test-user", PolicyName: "UpgradeToTeams"}})
}
