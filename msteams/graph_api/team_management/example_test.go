package teammanagement_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

// This example is compile-only. Retain the result even on
// error: group/team updates are not atomic, and readback can fail after writes.
func ExampleService_UpdateByID() {
	update := func(ctx context.Context, sdk *msteams.Client, id, description string) (management.Result, error) {
		disabled := false
		return sdk.GraphAPI.TeamManagement.UpdateByID(ctx, id, management.UpdateRequest{
			Properties: group.UpdateRequest{Description: &description},
			Settings:   team.Settings{MemberSettings: &team.MemberSettings{AllowCreateUpdateChannels: &disabled}},
		})
	}
	_ = update
}

// Compile-only: the application supplies credentials and owns recovery records.
func ExampleService_SetArchivedState() {
	archive := func(ctx context.Context, sdk *msteams.Client, id string) (management.ArchiveStateResult, error) {
		// Even on error retain the result privately. An accepted remote operation
		// can continue after this wait times out. Reconcile rather than resubmit.
		return sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, id, true, management.ArchiveStateOptions{})
	}
	_ = archive
}

// Compile-only: resolve credentials and persist private recovery data in the
// application. Explicit retention leaves cleanup/reconciliation to the caller.
func ExampleService_Create() {
	create := func(ctx context.Context, sdk *msteams.Client, ownerID string) (management.CreationResult, error) {
		alias := "sdk-lab-fixture"
		return sdk.GraphAPI.TeamManagement.Create(ctx, management.CreateRequest{
			DisplayName: "SDK lab fixture", MailNickname: &alias, Owner: ownerID,
			RetainCreatedGroup: true,
		})
	}
	_ = create
}

func ExampleService_MigrateGroup() {
	migrate := func(ctx context.Context, sdk *msteams.Client, groupID string) (management.CreationResult, error) {
		// This workflow never creates or deletes the supplied backing group.
		return sdk.GraphAPI.TeamManagement.MigrateGroup(ctx, groupID, management.CreationOptions{})
	}
	_ = migrate
}
