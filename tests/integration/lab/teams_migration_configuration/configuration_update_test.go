package teamsmigrationconfigurationacc

import (
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccMigrationConfigurationUpdateRestore(t *testing.T) {
	labkit.RunSingletonUpdate(t, migrationSingleton)
}
