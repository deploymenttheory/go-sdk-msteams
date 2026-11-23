package configurationdiagnostics_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	cd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/configuration_diagnostics"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleService_GetConfiguration() {
	read := func(ctx context.Context, cfg *auth.Config) (identity *string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		value, _, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.GetConfiguration(ctx, "TeamsFeedbackPolicy", "Global", nil)
		if err != nil {
			return nil, err
		}
		return value.Identity, nil
	}
	_ = read
}

// Create, replace and remove one instance of a configuration type. Set sends
// the complete instance: omitted properties return to their defaults.
func ExampleService_CreateConfiguration() {
	lifecycle := func(ctx context.Context, sdk *msteams.Client) error {
		s := sdk.TeamsAdminAPI.ConfigurationDiagnostics
		name := "Tag:Example"
		if _, err := s.CreateConfiguration(ctx, "TeamsFeedbackPolicy", cd.ConfigurationRequest{Identity: name,
			Settings: map[string]any{"UserInitiatedMode": "Disabled"}}, nil); err != nil {
			return err
		}
		if _, err := s.SetConfiguration(ctx, "TeamsFeedbackPolicy", name, cd.ConfigurationRequest{Identity: name,
			Settings: map[string]any{"UserInitiatedMode": "Enabled"}}, nil); err != nil {
			return err
		}
		_, err := s.RemoveConfiguration(ctx, "TeamsFeedbackPolicy", name)
		return err
	}
	_ = lifecycle
}

// Clearing cache keys is a support operation; tenant administrators are
// refused with auth.ErrForbidden.
func ExampleService_ClearCache() {
	clearKeys := func(ctx context.Context, sdk *msteams.Client, keys []string) ([]string, error) {
		results, _, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.ClearCache(ctx, cd.ClearCacheRequest{KeysToDelete: keys})
		if errors.Is(err, auth.ErrForbidden) {
			return nil, err
		}
		var failed []string
		for _, result := range results {
			failed = append(failed, result.FailedKeys...)
		}
		return failed, err
	}
	_ = clearKeys
}
