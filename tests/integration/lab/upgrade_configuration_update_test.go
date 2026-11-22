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
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_configuration"
)

func upgradeConfigurationProjection(v *configuration.Configuration) map[string]any {
	return map[string]any{"DownloadTeams": v.DownloadTeams, "SfBMeetingJoinUx": v.SfBMeetingJoinUx, "BlockLegacyAuthorization": v.BlockLegacyAuthorization}
}

type upgradeConfigurationCleanup struct{ service *configuration.Service }

func (upgradeConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}
func (upgradeConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}
func (h upgradeConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var snapshot struct {
		Value *bool `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF || snapshot.Value == nil {
		return lab.ErrUnresolved
	}
	before, _, err := h.service.GetByIdentity(ctx, "Global")
	if err != nil {
		return err
	}
	if before != nil && before.DownloadTeams != nil && *before.DownloadTeams == *snapshot.Value {
		return nil
	}
	if _, _, err := h.service.UpdateByIdentity(ctx, "Global", configuration.UpdateRequest{Settings: configuration.Settings{DownloadTeams: policy.Set(*snapshot.Value)}}); err != nil {
		return err
	}
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.DownloadTeams != nil && *v.DownloadTeams == *snapshot.Value, e
	})
}

func TestAccUpgradeConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid upgrade configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("upgrade configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid upgrade configuration mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private upgrade recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"upgrade-configuration": upgradeConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("upgrade configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("upgrade recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || before == nil || before.DownloadTeams == nil {
		logAccessResponse(t, response, err)
		t.Fatal("supported original upgrade setting unavailable; update refused")
	}
	compare := func(v *configuration.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !sameTemplateProjection(upgradeConfigurationProjection(v), tenantConfigurationOracle(t, ctx, "upgrade-configuration", tenant)) {
			t.Fatal("Go and PowerShell upgrade configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]bool{"value": *before.DownloadTeams})
	if err != nil {
		t.Fatal("upgrade original setting could not be journaled")
	}
	entry, err := journal.PlanRestore("upgrade-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("upgrade restoration intent not journaled")
	}
	want := !*before.DownloadTeams
	_, response, err = s.UpdateByIdentity(ctx, "Global", configuration.UpdateRequest{Settings: configuration.Settings{DownloadTeams: policy.Set(want)}})
	writeResponseFixture(t, "teams_upgrade_configuration", "update", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("upgrade configuration update failed")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("upgrade update receipt not journaled")
	}
	var updated *configuration.Configuration
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		expected := upgradeConfigurationProjection(before)
		expected["DownloadTeams"] = &want
		matched := e == nil && v != nil && sameTemplateProjection(expected, upgradeConfigurationProjection(v))
		if matched {
			updated = v
			writeResponseFixture(t, "teams_upgrade_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("upgrade configuration update did not persist")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("upgrade original setting restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || restored == nil || restored.DownloadTeams == nil || !sameTemplateProjection(upgradeConfigurationProjection(before), upgradeConfigurationProjection(restored)) {
		t.Fatal("upgrade original state not independently restored")
	}
	writeResponseFixture(t, "teams_upgrade_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitUpgradeRecoveryRejectsForeignTargetsAndUnsupportedOriginals(t *testing.T) {
	t.Parallel()
	for _, entry := range []lab.Entry{
		{ResourceID: "Tag:foreign", Snapshot: json.RawMessage(`{"value":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":"false"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":1}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":true,"extra":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"value":true} {}`)},
	} {
		if !errors.Is((upgradeConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
}
