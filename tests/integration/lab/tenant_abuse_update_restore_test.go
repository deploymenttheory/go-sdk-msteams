package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_tenant_abuse_configuration"
)

type tenantAbuseConfigurationCleanup struct{ service *resource.Service }

func (tenantAbuseConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}
func (tenantAbuseConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}
func (h tenantAbuseConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var original struct {
		Value *uint64 `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&original) != nil || d.Decode(new(any)) != io.EOF || original.Value == nil {
		return lab.ErrUnresolved
	}
	current, _, e := h.service.GetByIdentity(ctx, "Global")
	if e != nil {
		return e
	}
	if current != nil && current.LowSeatLimit != nil && *current.LowSeatLimit == *original.Value {
		return nil
	}
	if _, _, e := h.service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{LowSeatLimit: policy.Set(*original.Value)}}); e != nil {
		return e
	}
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.LowSeatLimit != nil && *v.LowSeatLimit == *original.Value, e
	})
}
func tenantAbuseProjection(v *resource.Configuration) map[string]any {
	return map[string]any{"LowSeatLimit": v.LowSeatLimit, "TrialTenantValidation": v.TrialTenantValidation, "SkipValidation": v.SkipValidation, "CreateThreadThresholdPerSeat": v.CreateThreadThresholdPerSeat, "AddMemberThresholdPerSeat": v.AddMemberThresholdPerSeat, "SendMessageThresholdPerSeat": v.SendMessageThresholdPerSeat, "CreateThreadThresholdForTrialTenants": v.CreateThreadThresholdForTrialTenants, "AddMemberThresholdForTrialTenants": v.AddMemberThresholdForTrialTenants, "SendMessageThresholdForTrialTenants": v.SendMessageThresholdForTrialTenants, "BlockedTenants": v.BlockedTenants}
}

// TestAccTenantAbuseConfigurationUpdateRestore refuses mutation without a readable
// non-null original. It remains executable for an explicitly authorized identity.
func TestAccTenantAbuseConfigurationUpdateRestore(t *testing.T) {
	cfg, e := lab.Load()
	if e != nil {
		t.Fatal("invalid tenant-abuse mutation configuration")
	}
	if e := cfg.Require(lab.Mutation); errors.Is(e, lab.ErrDisabled) {
		t.Skip("tenant-abuse update requires explicit mutation opt-in")
	} else if e != nil {
		t.Fatal("invalid tenant-abuse mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, r, e := s.GetByIdentity(ctx, "Global")
	if e != nil || before == nil || before.LowSeatLimit == nil {
		logAccessResponse(t, r, e)
		t.Fatal("tenant-abuse original UInt64 unavailable; mutation refused")
	}
	journal, e := lab.NewJournal(cfg.JournalDirectory, binding)
	if e != nil {
		t.Fatal("private tenant-abuse journal unavailable")
	}
	handlers := map[string]lab.Handler{"tenant-abuse-configuration": tenantAbuseConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, e := journal.Recover(ctx, handlers); e != nil {
			t.Error("tenant-abuse restoration unresolved; private journal retained")
		}
		if e := journal.Close(); e != nil {
			t.Error("tenant-abuse journal close failed")
		}
	})
	compare := func(v *resource.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !sameTemplateProjection(tenantAbuseProjection(v), tenantConfigurationOracle(t, ctx, "tenant-abuse", tenant)) {
			t.Fatal("Go and PowerShell tenant-abuse getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, e := json.Marshal(map[string]uint64{"value": *before.LowSeatLimit})
	if e != nil {
		t.Fatal("tenant-abuse original could not be journaled")
	}
	entry, e := journal.PlanRestore("tenant-abuse-configuration", "Global", snapshot)
	if e != nil {
		t.Fatal("tenant-abuse restore intent not journaled")
	}
	want := *before.LowSeatLimit
	if want == ^uint64(0) {
		want--
	} else {
		want++
	}
	_, r, e = s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{LowSeatLimit: policy.Set(want)}})
	writeResponseFixture(t, "teams_tenant_abuse_configuration", "update", r, nil)
	if e != nil {
		logAccessResponse(t, r, e)
		t.Fatal("tenant-abuse update failed; restoration intent retained")
	}
	if e := journal.Applied(entry, "Global"); e != nil {
		t.Fatal("tenant-abuse update receipt not journaled")
	}
	expected := tenantAbuseProjection(before)
	expected["LowSeatLimit"] = &want
	var updated *resource.Configuration
	if e := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		matched := e == nil && v != nil && sameTemplateProjection(expected, tenantAbuseProjection(v))
		if matched {
			updated = v
			writeResponseFixture(t, "teams_tenant_abuse_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); e != nil {
		t.Fatal("tenant-abuse UInt64 persistence or omission failed")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, e := journal.Recover(ctx, handlers); e != nil {
		t.Fatal("tenant-abuse original restoration failed")
	}
	restored, r, e := s.GetByIdentity(ctx, "Global")
	if e != nil || restored == nil || !sameTemplateProjection(tenantAbuseProjection(before), tenantAbuseProjection(restored)) {
		t.Fatal("tenant-abuse original was not independently restored")
	}
	writeResponseFixture(t, "teams_tenant_abuse_configuration", "restored_readback", r, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}
func TestUnitTenantAbuseRecoveryRejectsInvalidUnsignedSnapshots(t *testing.T) {
	t.Parallel()
	for _, entry := range []lab.Entry{
		{ResourceID: "other", Snapshot: json.RawMessage(`{"value":1}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":-1}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":18446744073709551616}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":1,"other":true}`)},
	} {
		if !errors.Is((tenantAbuseConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported recovery target or UInt64 snapshot accepted")
		}
	}
}
