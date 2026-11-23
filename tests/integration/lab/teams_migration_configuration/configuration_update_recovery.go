package teamsmigrationconfigurationacc

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_migration_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

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
