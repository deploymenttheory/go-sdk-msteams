package teamscallholdpolicy_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
)

func ExampleService_CreateByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.TeamsCallHoldPolicies.CreateByIdentity(context.Background(), "Tag:Fixture", resource.CreateRequest{Settings: resource.Settings{Description: policy.Set(""), AudioFileID: policy.Set("fixture-audio-file")}})
}
func ExampleService_UpdateByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Explicit Set read/PATCH preserves an omitted AudioFileId.
	_, _, _ = sdk.TeamsAdminAPI.TeamsCallHoldPolicies.UpdateByIdentity(context.Background(), "Tag:Fixture", resource.UpdateRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
}
func ExampleService_AssignToGroup() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	// Generated Grant uses discovery; explicit native priority zero is preserved.
	_, _, _ = sdk.TeamsAdminAPI.TeamsCallHoldPolicies.AssignToGroup(context.Background(), "fixture-group", "Fixture", new(int32(0)))
}
