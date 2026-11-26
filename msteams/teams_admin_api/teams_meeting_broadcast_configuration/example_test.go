package teamsmeetingbroadcastconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Configuration, error) {
		value, _, err := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.GetByIdentity(ctx, "Global", resource.ReadOptions{ExposeSDNConfigurationJsonBlob: true})
		return value, err
	}
	_ = read
}

func ExampleService_UpdateGlobal() {
	update := func(ctx context.Context, sdk *msteams.Client) error {
		url := "https://support.example.com/broadcast"
		_, err := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.UpdateGlobal(ctx, resource.UpdateRequest{SupportURL: &url})
		return err
	}
	_ = update
}
