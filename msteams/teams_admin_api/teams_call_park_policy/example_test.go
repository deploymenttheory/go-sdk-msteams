package teamscallparkpolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
)

func ExampleService_Create() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.TeamsCallParkPolicies.Create(context.Background(), resource.CreateRequest{Identity: "Tag:Fixture", Settings: resource.Settings{Description: new(""), AllowCallPark: new(false), ParkTimeoutSeconds: new(int64(120)), PickupRangeStart: new(int64(100)), PickupRangeEnd: new(int64(199))}})
}
func ExampleService_UpdateGlobal() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Bound zero stays present; omitted fields do not acquire defaults.
	_, _ = sdk.TeamsAdminAPI.TeamsCallParkPolicies.UpdateGlobal(context.Background(), resource.Settings{ParkTimeoutSeconds: new(int64(0))})
}
