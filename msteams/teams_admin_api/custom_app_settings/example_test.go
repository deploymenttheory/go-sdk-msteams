package customappsettings_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own client and context.
func ExampleService_Get() {
	sideloading := func(ctx context.Context, sdk *msteams.Client) (bool, error) {
		settings, _, err := sdk.TeamsAdminAPI.CustomAppSettings.Get(ctx)
		if err != nil {
			return false, err
		}
		return settings.SideloadedAppsInteractionEnabled(), nil
	}
	_ = sideloading
}

// SetSideloadedAppsInteraction follows Set-CsTeamsSettingsCustomApp: it reads
// the settings and writes them back with the new flag.
func ExampleService_SetSideloadedAppsInteraction() {
	disable := func(ctx context.Context, sdk *msteams.Client) error {
		_, err := sdk.TeamsAdminAPI.CustomAppSettings.SetSideloadedAppsInteraction(ctx, false)
		return err
	}
	_ = disable
}

// Read the public cmdlet's flag projection without strict full-document decoding.
func ExampleService_GetSideloadedAppsInteraction() {
	read := func(ctx context.Context, sdk *msteams.Client) (bool, error) {
		enabled, _, err := sdk.TeamsAdminAPI.CustomAppSettings.GetSideloadedAppsInteraction(ctx)
		return enabled, err // A false value with an error is not a disabled observation.
	}
	_ = read // Compiles without selecting credentials or reading a tenant.
}
