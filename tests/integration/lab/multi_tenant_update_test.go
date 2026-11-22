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
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_multi_tenant_organization_configuration"
)

type multiTenantConfigurationCleanup struct{ service *configuration.Service }

func (multiTenantConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}
func (multiTenantConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}
func (h multiTenantConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var snapshot struct {
		Value *string `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF || snapshot.Value == nil || (*snapshot.Value != "Enabled" && *snapshot.Value != "Disabled") {
		return lab.ErrUnresolved
	}
	before, _, err := h.service.GetByIdentity(ctx, "Global")
	if err != nil {
		return err
	}
	if before != nil && before.CopilotFromHomeTenant != nil && *before.CopilotFromHomeTenant == *snapshot.Value {
		return nil
	}
	if _, _, err := h.service.UpdateByIdentity(ctx, "Global", configuration.UpdateRequest{Settings: configuration.Settings{CopilotFromHomeTenant: policy.Set(*snapshot.Value)}}); err != nil {
		return err
	}
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.CopilotFromHomeTenant != nil && *v.CopilotFromHomeTenant == *snapshot.Value, e
	})
}

func TestAccMultiTenantOrganizationConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid multi-tenant configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("multi-tenant configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid multi-tenant configuration mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private multi-tenant recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"multi-tenant-configuration": multiTenantConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("multi-tenant configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("multi-tenant recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || before == nil || before.CopilotFromHomeTenant == nil || (*before.CopilotFromHomeTenant != "Enabled" && *before.CopilotFromHomeTenant != "Disabled") {
		logAccessResponse(t, response, err)
		t.Fatal("supported original multi-tenant setting unavailable; update refused")
	}
	compare := func(v *configuration.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !sameTemplateProjection(map[string]any{"CopilotFromHomeTenant": v.CopilotFromHomeTenant}, tenantConfigurationOracle(t, ctx, "multi-tenant", tenant)) {
			t.Fatal("Go and PowerShell multi-tenant configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]string{"value": *before.CopilotFromHomeTenant})
	if err != nil {
		t.Fatal("multi-tenant original setting could not be journaled")
	}
	entry, err := journal.PlanRestore("multi-tenant-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("multi-tenant restoration intent not journaled")
	}
	want := "Disabled"
	if *before.CopilotFromHomeTenant == want {
		want = "Enabled"
	}
	_, response, err = s.UpdateByIdentity(ctx, "Global", configuration.UpdateRequest{Settings: configuration.Settings{CopilotFromHomeTenant: policy.Set(want)}})
	writeResponseFixture(t, "teams_multi_tenant_organization_configuration", "update", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("multi-tenant configuration update failed")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("multi-tenant update receipt not journaled")
	}
	var updated *configuration.Configuration
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		matched := e == nil && v != nil && v.CopilotFromHomeTenant != nil && *v.CopilotFromHomeTenant == want
		if matched {
			updated = v
			writeResponseFixture(t, "teams_multi_tenant_organization_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("multi-tenant configuration update did not persist")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("multi-tenant original setting restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || restored == nil || restored.CopilotFromHomeTenant == nil || *restored.CopilotFromHomeTenant != *before.CopilotFromHomeTenant {
		t.Fatal("multi-tenant original state not independently restored")
	}
	writeResponseFixture(t, "teams_multi_tenant_organization_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitMultiTenantRecoveryRejectsForeignTargetsAndUnsupportedOriginals(t *testing.T) {
	t.Parallel()
	for _, entry := range []lab.Entry{
		{ResourceID: "Tag:foreign", Snapshot: json.RawMessage(`{"value":"Enabled"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":false}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":"Unknown"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":"Enabled","extra":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":"Enabled"} {}`)},
	} {
		if !errors.Is((multiTenantConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
}
