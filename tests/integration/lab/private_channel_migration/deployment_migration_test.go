package privatechannelmigrationacc

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccPrivateChannelMigrationReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
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
		t.Fatalf("private-channel migration status read failed (HTTP %d)", status)
	}
	if value.TenantID == nil || !strings.EqualFold(*value.TenantID, tenant) {
		t.Fatal("migration status tenant does not match selected identity")
	}
	labkit.WriteResponseFixture(t, "private_channel_migration", "tenant_status", response, nil)
	t.Log("private migration bound-tenant status read: passed")

	t.Log("tenant private-channel migration observation read; completion/permissions remain separately qualified")
}

// TestAccPrivateChannelMigrationAuthentication runs only in the credential-owning
// auth partition. Certificate material never reaches a service partition.
func TestAccPrivateChannelMigrationAuthentication(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	if os.Getenv("MSTEAMS_LAB_AUTH_METHOD") != "all" {
		t.Fatal("private migration backend conformance requires the auth partition")
	}
	if !labkit.OracleEnabled(t) {
		t.Fatal("private migration backend conformance requires pinned PowerShell")
	}
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	var observed *labkit.PrivateMigrationObservation
	t.Run("certificate-oracle", func(t *testing.T) {
		t.Setenv("MSTEAMS_LAB_AUTH_METHOD", "certificate")
		_, binding, _ := labkit.SDKClient(t)
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.DirectoryID.MatchString(tenant) {
			t.Fatal("canonical migration tenant binding required")
		}
		observed = labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "private-migration", ID: tenant}).PrivateMigration
		if observed == nil || observed.TenantID == nil || !strings.EqualFold(*observed.TenantID, tenant) {
			t.Fatal("pinned migration status tenant does not match selected identity")
		}
	})
	for _, method := range []string{"client-secret", "certificate"} {
		t.Run(method, func(t *testing.T) {
			t.Setenv("MSTEAMS_LAB_AUTH_METHOD", method)
			sdk, binding, _ := labkit.SDKClient(t)
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
				t.Fatalf("private-channel migration status read failed (HTTP %d)", status)
			}
			if value.TenantID == nil || !strings.EqualFold(*value.TenantID, tenant) {
				t.Fatal("migration status tenant does not match selected identity")
			}
			t.Log("private migration " + method + " backend read: passed")
			want := labkit.PrivateMigrationObservation{TenantID: value.TenantID, Status: value.MigrationStatus, Details: value.Details}
			if observed == nil || !reflect.DeepEqual(want, *observed) {
				t.Fatal("PowerShell and Go selected private migration fields differ")
			}
			t.Log("PowerShell 8.1.0 / Go private migration read parity: passed")
		})
	}
}
