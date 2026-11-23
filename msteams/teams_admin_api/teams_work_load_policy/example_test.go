package teamsworkloadpolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_load_policy"
)

func ExampleService_Create() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.TeamsWorkLoadPolicies.Create(context.Background(), resource.CreateRequest{Identity: "sdk-owned-policy", Settings: resource.Settings{AllowCalling: new(false)}})
}
