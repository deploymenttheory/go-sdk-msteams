package teamsmigrationconfigurationacc

import (
	"context"
	"strings"
	"testing"

	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_migration_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccMigrationConfigurationReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMigrationConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_migration_configuration", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("migration configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("migration configuration identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_migration_configuration", "list", response, nil)
	var selected *migration.Configuration
	for i := range values {
		item := &values[i]
		if item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous migration configuration Global")
			}
			selected = item
		}
	}
	if err != nil || selected == nil || !labkit.SameProjection(migrationConfigurationProjection(value), migrationConfigurationProjection(selected)) {
		t.Fatal("migration configuration read/list getters differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		actual := labkit.TenantConfigurationOracle(t, ctx, "migration", tenant)
		if !labkit.SameProjection(migrationConfigurationProjection(value), actual) {
			t.Fatal("Go and PowerShell migration configuration getters differ")
		}
	}
	t.Log("migration configuration read/list and pinned getter comparison verified")
}
