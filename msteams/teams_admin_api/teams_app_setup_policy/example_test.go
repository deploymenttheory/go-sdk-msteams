package teamsappsetuppolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

func ExampleNewApp() { app := resource.NewApp("future-app-id"); _ = app }
func ExampleService_UpdateByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Explicit read/edit/PATCH retains the read snapshot on submission failure.
	_, _, _ = sdk.TeamsAdminAPI.TeamsAppSetupPolicies.UpdateByIdentity(context.Background(), "Tag:Fixture", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set("")}, PinnedAppBarApps: resource.ChangePins([]*resource.PinnedApp{resource.NewPinnedApp("future-app-id", 9007199254740993)}, nil)})
}
func ExampleService_AssignUser() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Nil and empty policy names have distinct wire representations.
	_, _, _ = sdk.TeamsAdminAPI.TeamsAppSetupPolicies.AssignUser(context.Background(), "user@example.invalid", new("Fixture"))
}
