package teamsguestmeetingconfigurationacc

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	guestmeeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_meeting_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccGuestMeetingConfigurationUpdateRestore toggles only AllowMeetNow on
// Global through Set's read-then-PATCH composition, verifies the other seven
// getters are preserved, and restores the journaled original value.
func TestAccGuestMeetingConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid guest meeting configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("guest meeting configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid guest meeting configuration mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private guest meeting recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"guest-meeting-configuration": guestMeetingConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("guest meeting configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("guest meeting recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || before == nil || before.AllowMeetNow == nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("supported original guest meeting setting unavailable; update refused")
	}
	compare := func(v *guestmeeting.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.SameProjection(guestMeetingProjection(v), guestMeetingOracle(t, ctx, tenant)) {
			t.Fatal("Go and PowerShell guest meeting configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]bool{"value": *before.AllowMeetNow})
	if err != nil {
		t.Fatal("guest meeting original setting could not be journaled")
	}
	entry, err := journal.PlanRestore("guest-meeting-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("guest meeting restoration intent not journaled")
	}
	changed := !*before.AllowMeetNow
	_, response, err = s.UpdateByIdentity(ctx, "Global", guestmeeting.UpdateRequest{Settings: guestmeeting.Settings{AllowMeetNow: policy.Set(changed)}})
	labkit.WriteResponseFixture(t, "teams_guest_meeting_configuration", "update", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("guest meeting configuration update failed; restoration journal retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("guest meeting update receipt not journaled")
	}
	want := maps.Clone(guestMeetingProjection(before))
	want["AllowMeetNow"] = &changed
	var updated *guestmeeting.Configuration
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		matched := e == nil && v != nil && labkit.SameProjection(want, guestMeetingProjection(v))
		if matched {
			updated = v
			labkit.WriteResponseFixture(t, "teams_guest_meeting_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("guest meeting configuration update or omitted-field preservation did not converge")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("guest meeting original setting restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || restored == nil || !labkit.SameProjection(guestMeetingProjection(before), guestMeetingProjection(restored)) {
		t.Fatal("guest meeting original state not independently restored")
	}
	labkit.WriteResponseFixture(t, "teams_guest_meeting_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitGuestMeetingRecoveryRejectsForeignTargetsAndUnsupportedOriginals(t *testing.T) {
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
		if !errors.Is((guestMeetingConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
}
