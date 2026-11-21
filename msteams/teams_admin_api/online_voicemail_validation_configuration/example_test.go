package onlinevoicemailvalidationconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func ExampleService_GetByIdentity() {
	read := func(ctx context.Context, sdk *msteams.Client) (*resource.Configuration, error) {
		value, _, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.GetByIdentity(ctx, "Global")
		return value, err
	}
	_ = read
}

func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client) error {
		_, _, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{AudioFileValidationEnabled: policy.Set(false)}})
		// URI remains unbound. Global changes require a caller's explicit restore plan.
		return err
	}
	_ = update
}
