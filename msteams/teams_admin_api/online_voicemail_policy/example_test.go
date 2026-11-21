package onlinevoicemailpolicy_test

import (
	"context"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func ExampleService_Create() {
	create := func(ctx context.Context, sdk *msteams.Client, identity string) error {
		length, err := resource.TimeSpanFromDuration(90 * time.Second)
		if err != nil {
			return err
		}
		_, _, err = sdk.TeamsAdminAPI.OnlineVoicemailPolicies.Create(ctx, resource.CreateRequest{Identity: identity, Settings: resource.Settings{MaximumRecordingLength: policy.Set(length), EnableTranscription: policy.Set(false)}})
		return err
	}
	_ = create
}

func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client, identity string) error {
		_, _, err := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.UpdateByIdentity(ctx, identity, resource.UpdateRequest{Settings: resource.Settings{MaximumRecordingLength: policy.Null[resource.TimeSpan]()}})
		return err
	}
	_ = update
}

func ExampleService_AssignGroup() {
	assign := func(ctx context.Context, sdk *msteams.Client, groupID, name string) (*resource.AssignmentResult, error) {
		result, _, err := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.AssignGroup(ctx, groupID, policy.Set(name), nil)
		// An optional OperationID supports a later explicit GetAssignmentOperation.
		// Accepted does not establish effective policy or completed propagation.
		return result, err
	}
	_ = assign
}
