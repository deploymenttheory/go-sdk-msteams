package privacyconfiguration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
)

func ExampleService_PatchByIdentity() {

	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	_, _, _ = sdk.TeamsAdminAPI.PrivacyConfigurations.PatchByIdentity(context.Background(), "Global", resource.PatchRequest{Settings: resource.Settings{EnablePrivacyMode: policy.Set(false), PublishLocationDataDefault: policy.Null[bool]()}})
}

func ExampleService_UpdateByIdentity() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}

	result, _, err := sdk.TeamsAdminAPI.PrivacyConfigurations.UpdateByIdentity(context.Background(), "Global", resource.UpdateRequest{Settings: resource.Settings{AutoInitiateContacts: policy.Set(false)}})
	_, _ = result, err
}
