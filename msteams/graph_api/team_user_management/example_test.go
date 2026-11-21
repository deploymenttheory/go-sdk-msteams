package teamusermanagement_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

// This example is compile-only. Retain stages on error: membership can have
// succeeded before the owner stage fails.
func ExampleService_Add() {
	ensureOwner := func(ctx context.Context, sdk *msteams.Client, groupID, userID string) (management.ChangeResult, error) {
		return sdk.GraphAPI.TeamUserManagement.Add(ctx, groupID, userID, management.ChangeOptions{Role: management.Owner})
	}
	_ = ensureOwner
}

// demoteOwner removes only the owner reference and leaves membership intact.
func ExampleService_Remove() {
	demoteOwner := func(ctx context.Context, sdk *msteams.Client, groupID, userID string) (management.ChangeResult, error) {
		return sdk.GraphAPI.TeamUserManagement.Remove(ctx, groupID, userID, management.ChangeOptions{Role: management.Owner})
	}
	_ = demoteOwner
}
