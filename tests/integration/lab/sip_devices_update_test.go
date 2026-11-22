package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_sip_devices_configuration"
)

type sipDevicesConfigurationCleanup struct{ service *configuration.Service }

func (sipDevicesConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}
func (sipDevicesConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}
func (h sipDevicesConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
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
	if before != nil && before.BulkSignIn != nil && *before.BulkSignIn == *snapshot.Value {
		return nil
	}
	if _, _, err := h.service.UpdateByIdentity(ctx, "Global", configuration.UpdateRequest{Settings: configuration.Settings{BulkSignIn: policy.Set(*snapshot.Value)}}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.BulkSignIn != nil && *v.BulkSignIn == *snapshot.Value, e
	})
}

func TestAccSipDevicesConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid sip-devices configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("sip-devices configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid sip-devices configuration mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private sip-devices recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"sip-devices-configuration": sipDevicesConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("sip-devices configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("sip-devices recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || before == nil || before.BulkSignIn == nil || (*before.BulkSignIn != "Enabled" && *before.BulkSignIn != "Disabled") {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("supported original sip-devices setting unavailable; update refused")
	}
	compare := func(v *configuration.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.SameProjection(map[string]any{"BulkSignIn": v.BulkSignIn}, labkit.TenantConfigurationOracle(t, ctx, "sip-devices", tenant)) {
			t.Fatal("Go and PowerShell sip-devices configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]string{"value": *before.BulkSignIn})
	if err != nil {
		t.Fatal("sip-devices original setting could not be journaled")
	}
	entry, err := journal.PlanRestore("sip-devices-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("sip-devices restoration intent not journaled")
	}
	want := "Disabled"
	if *before.BulkSignIn == want {
		want = "Enabled"
	}
	_, response, err = s.UpdateByIdentity(ctx, "Global", configuration.UpdateRequest{Settings: configuration.Settings{BulkSignIn: policy.Set(want)}})
	labkit.WriteResponseFixture(t, "teams_sip_devices_configuration", "update", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("sip-devices configuration update failed")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("sip-devices update receipt not journaled")
	}
	var updated *configuration.Configuration
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		matched := e == nil && v != nil && v.BulkSignIn != nil && *v.BulkSignIn == want
		if matched {
			updated = v
			labkit.WriteResponseFixture(t, "teams_sip_devices_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("sip-devices configuration update did not persist")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("sip-devices original setting restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || restored == nil || restored.BulkSignIn == nil || *restored.BulkSignIn != *before.BulkSignIn {
		t.Fatal("sip-devices original state not independently restored")
	}
	labkit.WriteResponseFixture(t, "teams_sip_devices_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitSipDevicesRecoveryRejectsForeignTargetsAndUnsupportedOriginals(t *testing.T) {
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
		if !errors.Is((sipDevicesConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
}
