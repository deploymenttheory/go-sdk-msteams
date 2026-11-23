package teamsmigrationconfigurationacc

import (
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_migration_configuration"
)

func migrationConfigurationProjection(v *migration.Configuration) map[string]any {
	return map[string]any{"EnableLegacyClientInterop": v.EnableLegacyClientInterop}
}
