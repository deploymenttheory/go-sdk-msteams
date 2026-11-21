package team_test

import (
	"context"
	"fmt"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
)

func ExampleSettings() {
	allowGiphy := false
	settings := team.Settings{FunSettings: &team.FunSettings{AllowGiphy: &allowGiphy}}
	fmt.Println(*settings.FunSettings.AllowGiphy, settings.MemberSettings == nil)
	// Output: false true
}

// This compile-only example performs no authentication or tenant operations in
// ordinary tests. A real application supplies its own bound client and context.
func ExampleService_Create() {
	createAndWait := func(ctx context.Context, sdk *msteams.Client, ownerID string) (client.OperationSnapshot, error) {
		accepted, err := sdk.GraphAPI.Teams.Create(ctx, team.CreateRequest{
			DisplayName: "SDK lab fixture",
			Members:     []team.InitialMember{{UserID: ownerID, Owner: true}},
		})
		if err != nil {
			// Keep accepted.Response for reconciliation if submission reached
			// Graph but its handle could not be validated. Do not resubmit.
			return client.OperationSnapshot{Response: accepted.Response}, err
		}
		// Persist accepted.TeamID/Operation.Snapshot().Handle privately before
		// waiting. A local timeout does not cancel the remote create.
		return accepted.Operation.Wait(ctx, client.WaitOptions{})
	}
	_ = createAndWait
}

func ExampleService_DeleteByID() {
	deleteOwnedTeam := func(ctx context.Context, sdk *msteams.Client, ownedGroupID string) error {
		// Deletes the backing group and its associated resources. Establish
		// fixture ownership before invoking; a failed delete is not replayed.
		_, err := sdk.GraphAPI.Teams.DeleteByID(ctx, ownedGroupID)
		return err
	}
	_ = deleteOwnedTeam
}

func ExampleService_UploadPicture() {
	upload := func(ctx context.Context, sdk *msteams.Client, groupID string, jpeg io.ReadCloser) error {
		// UploadPicture owns and closes jpeg, including validation failure.
		// Supply a valid JPEG; bytes are preserved without image conversion.
		_, err := sdk.GraphAPI.Teams.UploadPicture(ctx, groupID, jpeg)
		return err
	}
	_ = upload
}
