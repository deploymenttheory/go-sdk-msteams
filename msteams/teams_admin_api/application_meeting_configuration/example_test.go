package applicationmeetingconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_meeting_configuration"
)

func ExampleService_UpdateGlobal() {
	update := func(ctx context.Context, sdk *msteams.Client, appIDs []string) error {
		_, err := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.UpdateGlobal(ctx, configuration.UpdateRequest{AppIDs: configuration.ReplaceAppIDs(appIDs)})
		return err
	}
	_ = update
}

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*configuration.Configuration, error) {
		value, _, err := sdk.TeamsAdminAPI.ApplicationMeetingConfigurations.GetByIdentity(ctx, "Global")
		return value, err
	}
	_ = read
}
