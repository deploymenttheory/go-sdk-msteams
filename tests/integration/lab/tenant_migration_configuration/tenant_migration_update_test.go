package tenantmigrationconfigurationacc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	tenantmigration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_migration_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccTenantMigrationConfigurationUpdateRestore temporarily disables the
// Meeting Migration Service and re-enables it. It refuses to run when the
// original is disabled: enabling MMS from a disabled original could let the
// service act on pending meeting updates, so the test only ever moves
// enabled -> disabled -> enabled, which cannot start a migration.
func TestAccTenantMigrationConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid tenant migration configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("tenant migration configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid tenant migration configuration mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.TenantMigrationConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private tenant migration recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"tenant-migration-configuration": tenantMigrationConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("tenant migration configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("tenant migration recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || !tenantMigrationComplete(before) {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("original tenant migration setting unavailable; update refused")
	}
	if !*before.MeetingMigrationEnabled {
		t.Fatal("Meeting Migration Service is disabled; enabling it is not a safe lab write, update refused")
	}
	compare := func(v *tenantmigration.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.SameProjection(tenantMigrationProjection(v), tenantMigrationOracle(t, ctx, tenant)) {
			t.Fatal("Go and PowerShell tenant migration configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	original := tenantMigrationSnapshot{MeetingMigrationEnabled: before.MeetingMigrationEnabled}
	snapshot, err := json.Marshal(original)
	if err != nil {
		t.Fatal("tenant migration original setting could not be journaled")
	}
	entry, err := journal.PlanRestore("tenant-migration-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("tenant migration restoration intent not journaled")
	}
	disabled := false
	want := tenantMigrationSnapshot{MeetingMigrationEnabled: &disabled}
	response, err = s.UpdateGlobal(ctx, tenantmigration.UpdateRequest{MeetingMigrationEnabled: want.MeetingMigrationEnabled})
	labkit.WriteResponseFixture(t, "tenant_migration_configuration", "update", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("tenant migration configuration update failed; restoration journal retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("tenant migration update receipt not journaled")
	}
	var updated *tenantmigration.Configuration
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		matched := e == nil && want.matches(v)
		if matched {
			updated = v
			labkit.WriteResponseFixture(t, "tenant_migration_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("tenant migration configuration update did not persist")
	}
	if updated.Identity == nil || !strings.EqualFold(*updated.Identity, "Global") {
		t.Fatal("tenant migration update changed the configuration identity")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("tenant migration original setting restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || !original.matches(restored) || !labkit.SameProjection(tenantMigrationProjection(before), tenantMigrationProjection(restored)) {
		t.Fatal("tenant migration original state not independently restored")
	}
	labkit.WriteResponseFixture(t, "tenant_migration_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitTenantMigrationRecoveryRejectsForeignTargetsAndUnsafeOriginals(t *testing.T) {
	t.Parallel()
	for _, entry := range []lab.Entry{
		{ResourceID: "Tag:foreign", Snapshot: json.RawMessage(`{"meeting_migration_enabled":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"meeting_migration_enabled":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"meeting_migration_enabled":false}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"meeting_migration_enabled":"true"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"meeting_migration_enabled":true,"extra":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"meeting_migration_enabled":true} {}`)},
	} {
		if !errors.Is((tenantMigrationConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
}
