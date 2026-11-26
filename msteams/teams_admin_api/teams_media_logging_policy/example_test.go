package teamsmedialoggingpolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_logging_policy"
)

func ExampleService_Assign() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Grant the built-in Enabled instance, then clear the direct assignment.
	ctx := context.Background()
	user := resource.UserAssignment{Identity: "sdk-owned-test-user", PolicyName: "Enabled"}
	_, _, _ = sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, resource.Assignment{User: &user})
	user.PolicyName = ""
	_, _, _ = sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, resource.Assignment{User: &user})
}

func ExampleService_ListFiltered() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.ListFiltered(context.Background(), "Enabled")
}
