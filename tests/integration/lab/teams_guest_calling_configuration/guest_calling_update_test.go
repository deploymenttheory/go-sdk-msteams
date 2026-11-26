package teamsguestcallingconfigurationacc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	guestcalling "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_calling_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccGuestCallingConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid guest calling configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("guest calling configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid guest calling configuration mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsGuestCallingConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private guest calling recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"guest-calling-configuration": guestCallingConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("guest calling configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("guest calling recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || before == nil || before.AllowPrivateCalling == nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("supported original guest calling setting unavailable; update refused")
	}
	compare := func(v *guestcalling.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.SameProjection(guestCallingProjection(v), guestCallingOracle(t, ctx, tenant)) {
			t.Fatal("Go and PowerShell guest calling configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]bool{"value": *before.AllowPrivateCalling})
	if err != nil {
		t.Fatal("guest calling original setting could not be journaled")
	}
	entry, err := journal.PlanRestore("guest-calling-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("guest calling restoration intent not journaled")
	}
	want := !*before.AllowPrivateCalling
	response, err = s.UpdateGlobal(ctx, guestcalling.UpdateRequest{AllowPrivateCalling: &want})
	labkit.WriteResponseFixture(t, "teams_guest_calling_configuration", "update", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("guest calling configuration update failed; restoration journal retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("guest calling update receipt not journaled")
	}
	var updated *guestcalling.Configuration
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		matched := e == nil && v != nil && v.AllowPrivateCalling != nil && *v.AllowPrivateCalling == want
		if matched {
			updated = v
			labkit.WriteResponseFixture(t, "teams_guest_calling_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("guest calling configuration update did not persist")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("guest calling original setting restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || restored == nil || restored.AllowPrivateCalling == nil || *restored.AllowPrivateCalling != *before.AllowPrivateCalling {
		t.Fatal("guest calling original state not independently restored")
	}
	labkit.WriteResponseFixture(t, "teams_guest_calling_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitGuestCallingRecoveryRejectsForeignTargetsAndUnsupportedOriginals(t *testing.T) {
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
		if !errors.Is((guestCallingConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
}
