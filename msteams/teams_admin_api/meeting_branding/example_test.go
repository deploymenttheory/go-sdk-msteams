package meetingbranding_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func ExampleNewBackgroundImage() {
	image, err := resource.NewBackgroundImage("fixture-image", 0, "Fixture", "https://example.invalid/background.png", resource.BackgroundImageOptions{})
	if err != nil {
		return
	}

	_, _ = resource.Assets(image)
}

func ExampleService_Create() {
	create := func(ctx context.Context, sdk *msteams.Client, name string) error {
		_, _, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.Create(ctx, resource.CreateRequest{Identity: name, Settings: resource.Settings{Fields: resource.Fields{EnableMeetingOptionsThemeOverride: policy.Set(false)}}})
		return err
	}
	_ = create
}

func ExampleService_UpdateByIdentity() {
	update := func(ctx context.Context, sdk *msteams.Client, name string) error {
		clear, err := resource.ReplaceAssets[resource.BackgroundImage]()
		if err != nil {
			return err
		}
		_, _, err = sdk.TeamsAdminAPI.MeetingBrandingPolicies.UpdateByIdentity(ctx, name, resource.UpdateRequest{MeetingBackgroundImages: clear})
		return err
	}
	_ = update
}
