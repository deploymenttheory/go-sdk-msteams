package teamoperation_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	operation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_operation"
)

// Compile-only: the application supplies its configured client and observed IDs.
func ExampleService_GetByID() {
	read := func(ctx context.Context, sdk *msteams.Client, teamID, operationID string) (*operation.Operation, error) {
		value, _, err := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, operationID)
		return value, err // A failed remote operation is still a readable snapshot.
	}
	_ = read
}

func ExampleService_WaitByID() {
	wait := func(ctx context.Context, sdk *msteams.Client, teamID, operationID string) (client.OperationSnapshot, error) {
		// A cancelled/failed wait retains its handle, last response and diagnostics.
		return sdk.GraphAPI.TeamOperations.WaitByID(ctx, teamID, operationID, client.WaitOptions{MaxPolls: 8})
	}
	_ = wait
}
