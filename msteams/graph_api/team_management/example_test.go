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
