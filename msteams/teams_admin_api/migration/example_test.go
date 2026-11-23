package migration_test

import (
	"context"
	"errors"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/migration"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleService_ListMeetingMigrationStatuses() {
	failed := func(ctx context.Context, cfg *auth.Config) (count int, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return 0, err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		since := time.Now().Add(-7 * 24 * time.Hour)
		state := migration.StateFailed
		values, _, err := sdk.TeamsAdminAPI.Migration.ListMeetingMigrationStatuses(ctx, &migration.StatusOptions{State: &state, StartTime: &since})
		return len(values), err
	}
	_ = failed
}

// StartMeetingMigration recreates the user's meetings and cannot be undone.
// Check the status first and never retry blindly after a lost response.
func ExampleService_StartMeetingMigration() {
	start := func(ctx context.Context, sdk *msteams.Client, user string) error {
		s := sdk.TeamsAdminAPI.Migration
		pending, _, err := s.ListMeetingMigrationStatuses(ctx, &migration.StatusOptions{Identity: &user, State: new(migration.StatePending)})
		if err != nil || len(pending) != 0 {
			return err
		}
		target := "Teams"
		_, _, err = s.StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{Identity: user, TargetMeetingType: &target})
		return err
	}
	_ = start
}
