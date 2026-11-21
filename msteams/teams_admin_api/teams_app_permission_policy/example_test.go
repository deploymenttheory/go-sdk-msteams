package teamsapppermissionpolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

func ExampleNewCatalogApp() { app := resource.NewCatalogApp("future-catalog-id"); _ = app }
func ExampleService_UpdateByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Explicit read/edit/PATCH retains the read snapshot on submission failure.
	_, _, _ = sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.UpdateByIdentity(context.Background(), "Tag:Fixture", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set("")}, DefaultCatalogApps: resource.ChangeCatalogApps([]*resource.CatalogApp{resource.NewCatalogApp("future-catalog-id")}, nil)})
}
func ExampleService_AssignUser() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Nil and empty policy names have distinct wire representations.
	_, _, _ = sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.AssignUser(context.Background(), "user@example.invalid", new("Fixture"))
}
