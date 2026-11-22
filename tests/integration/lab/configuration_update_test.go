package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	messaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_migration_configuration"
	shiftsApp "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shifts_app_policy"
	"github.com/go-resty/resty/v2"
)

func messagingSingleton(sdk *msteams.Client) labkit.SingletonConfiguration {
	s := sdk.TeamsAdminAPI.TeamsMessagingConfigurations
	return labkit.SingletonConfiguration{Kind: "messaging", Field: "EnableVideoMessageCaptions", Get: func(ctx context.Context) (map[string]any, *resty.Response, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return messagingConfigurationProjection(v), r, nil
	}, Set: func(ctx context.Context, value bool) (*resty.Response, error) {
		_, r, e := s.UpdateByIdentity(ctx, "Global", messaging.UpdateRequest{ScalarSettings: messaging.ScalarSettings{EnableVideoMessageCaptions: policy.Set(value)}})
		return r, e
	}}
}
func migrationSingleton(sdk *msteams.Client) labkit.SingletonConfiguration {
	s := sdk.TeamsAdminAPI.TeamsMigrationConfigurations
	return labkit.SingletonConfiguration{Kind: "migration", Field: "EnableLegacyClientInterop", Get: func(ctx context.Context) (map[string]any, *resty.Response, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return migrationConfigurationProjection(v), r, nil
	}, Set: func(ctx context.Context, value bool) (*resty.Response, error) {
		return s.UpdateGlobal(ctx, migration.UpdateRequest{EnableLegacyClientInterop: &value})
	}}
}

func shiftsAppSingleton(sdk *msteams.Client) labkit.SingletonConfiguration {
	s := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies
	return labkit.SingletonConfiguration{Kind: "shifts-app", Field: "AllowTimeClockLocationDetection", Get: func(ctx context.Context) (map[string]any, *resty.Response, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return map[string]any{"AllowTimeClockLocationDetection": v.AllowTimeClockLocationDetection}, r, nil
	}, Set: func(ctx context.Context, value bool) (*resty.Response, error) {
		return s.UpdateGlobal(ctx, shiftsApp.UpdateRequest{AllowTimeClockLocationDetection: &value})
	}}
}
func TestAccShiftsAppPolicyUpdateRestore(t *testing.T) {
	labkit.RunSingletonUpdate(t, shiftsAppSingleton)
}

func TestAccMessagingConfigurationUpdateRestore(t *testing.T) {
	labkit.RunSingletonUpdate(t, messagingSingleton)
}
func TestAccMigrationConfigurationUpdateRestore(t *testing.T) {
	labkit.RunSingletonUpdate(t, migrationSingleton)
}
