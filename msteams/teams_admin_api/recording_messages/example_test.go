package recordingmessages_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

func ExampleNewLocalization() {
	localization, _ := resource.NewLocalization(resource.LocalizationInput{Language: "en-US", InitiatorImplicit: new("Recording is starting.")})
	_ = localization
}
func ExampleService_Create() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	localization, _ := resource.NewLocalization(resource.LocalizationInput{Language: "en-US", InitiatorImplicit: new("Recording is starting.")})

	result, _, err := sdk.TeamsAdminAPI.RecordingMessages.Create(context.Background(), resource.CreateRequest{Localizations: resource.ReplaceLocalizations(localization)})
	_, _ = result, err
}
func ExampleService_UpdateByID() {
	var sdk *msteams.Client
	if sdk == nil {
		return
	}
	localization, _ := resource.NewLocalization(resource.LocalizationInput{Language: "en-US", InitiatorImplicit: new("Recording is starting.")})

	_, _, _ = sdk.TeamsAdminAPI.RecordingMessages.UpdateByID(context.Background(), "00000000-0000-4000-8000-000000000031", resource.UpdateRequest{Localizations: resource.ChangeLocalizations([]*resource.Localization{localization}, nil)})
}
