package privatechannelmigrationacc

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccPrivateChannelMigrationReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("canonical migration tenant binding required")
	}
	value, response, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
	if err != nil || value == nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Logf("private-channel migration SDK read failed (HTTP %d)", status)
		if oracle {
			// A native denial does not prove application-mode ineligibility.
			// Compare the same pinned command before retaining the native failure.
			labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "private-migration", ID: tenant})
		}
		t.Fatalf("private-channel migration status read failed (HTTP %d)", status)
	}
	if value.TenantID == nil || !strings.EqualFold(*value.TenantID, tenant) {
		t.Fatal("migration status tenant does not match selected identity")
	}
	labkit.WriteResponseFixture(t, "private_channel_migration", "tenant_status", response, nil)
	t.Log("private migration bound-tenant status read: passed")
	if oracle {
		actual := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "private-migration", ID: tenant}).PrivateMigration
		want := labkit.PrivateMigrationObservation{TenantID: value.TenantID, Status: value.MigrationStatus, Details: value.Details}
		if actual == nil || !reflect.DeepEqual(want, *actual) {
			t.Fatal("PowerShell and Go selected private migration fields differ")
		}
		t.Log("PowerShell 8.1.0 / Go private migration read parity: passed")
	}
	t.Log("tenant private-channel migration observation read; completion/permissions remain separately qualified")
}
