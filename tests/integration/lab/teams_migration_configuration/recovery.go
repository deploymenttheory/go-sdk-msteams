package teamsmigrationconfigurationacc

import (
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// RecoveryHandlers returns journal recovery handlers for the owned
// teams_migration_configuration resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"migration-configuration": labkit.ConfigurationRestore{SingletonConfiguration: migrationSingleton(sdk)},
	}
}
