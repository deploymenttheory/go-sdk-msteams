package teamsmeetingconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Configuration, error) {
		value, _, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.GetByIdentity(ctx, "Global")
		return value, err
	}
	_ = read
}

func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client) error {
		_, _, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{
			ScalarSettings: resource.ScalarSettings{
				CustomFooterText: policy.Set("Contoso meetings"),
				HelpURL:          policy.Null[string](), // explicit null clears the link
			},
		})
		return err
	}
	_ = update
}
