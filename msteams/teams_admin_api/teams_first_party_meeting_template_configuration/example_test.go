package teamsfirstpartymeetingtemplateconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Configuration, error) {
		value, _, err := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.GetByIdentity(ctx, "Global")
		return value, err
	}
	_ = read
}

func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client) error {
		_, _, err := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{
			ScalarSettings: resource.ScalarSettings{Description: policy.Set("Lab first-party templates")},
		})
		return err
	}
	_ = update
}
