package teamsmeetingbroadcastconfigurationacc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	broadcast "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// labSupportURL is a reserved documentation address written only for the
// duration of the update and replaced by the journaled original.
const labSupportURL = "https://sdk-lab.example.com/broadcast-support"

// sdnUnchanged reports whether the SDN provider settings the test never binds
// are identical; the token is compared by presence only and never logged.
func sdnUnchanged(a, b *broadcast.Configuration) bool {
	left, right := meetingBroadcastSDNProjection(a), meetingBroadcastSDNProjection(b)
	for _, name := range []string{"SupportURL", "AllowSdnProviderForBroadcastMeeting"} {
		delete(left, name)
		delete(right, name)
	}
	return labkit.SameProjection(left, right)
}

func TestAccMeetingBroadcastConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid meeting broadcast configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting broadcast configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid meeting broadcast configuration mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private meeting broadcast recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"meeting-broadcast-configuration": meetingBroadcastConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("meeting broadcast configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("meeting broadcast recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	read := func() *broadcast.Configuration {
		v, response, err := s.GetByIdentity(ctx, "Global", broadcast.ReadOptions{ExposeSDNConfigurationJsonBlob: true})
		if err != nil || v == nil || v.AllowSdnProviderForBroadcastMeeting == nil || v.SupportURL == nil {
			labkit.LogAccessResponse(t, response, err)
			t.Fatal("supported meeting broadcast settings unavailable")
		}
		return v
	}
	compare := func(v *broadcast.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		pinned := meetingBroadcastOracle(t, ctx, tenant)
		if !labkit.SameProjection(meetingBroadcastProjection(v), pinned.Settings) || !labkit.SameProjection(meetingBroadcastSDNProjection(v), pinned.SDN) {
			t.Fatal("Go and PowerShell meeting broadcast configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	before := read()
	if *before.SupportURL == labSupportURL {
		t.Fatal("meeting broadcast SupportURL already holds the lab value; update refused")
	}
	compare(before)
	snapshot, err := json.Marshal(meetingBroadcastSnapshot{Allow: before.AllowSdnProviderForBroadcastMeeting, SupportURL: before.SupportURL})
	if err != nil {
		t.Fatal("meeting broadcast original settings could not be journaled")
	}
	entry, err := journal.PlanRestore("meeting-broadcast-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("meeting broadcast restoration intent not journaled")
	}
	// AllowSdnProviderForBroadcastMeeting stays unbound: the service rejects
	// enabling it without an SDN provider (40013 SkypecastSdnMissingParameter),
	// and the lab has no provider to configure. Its omission is verified instead.
	wantURL := labSupportURL
	response, err := s.UpdateGlobal(ctx, broadcast.UpdateRequest{SupportURL: &wantURL})
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_configuration", "update", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting broadcast configuration update failed; restoration journal retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("meeting broadcast update receipt not journaled")
	}
	want := meetingBroadcastSnapshot{Allow: before.AllowSdnProviderForBroadcastMeeting, SupportURL: &wantURL}
	var updated *broadcast.Configuration
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global", broadcast.ReadOptions{})
		matched := e == nil && want.matches(v)
		if matched {
			labkit.WriteResponseFixture(t, "teams_meeting_broadcast_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("meeting broadcast configuration update did not persist")
	}
	updated = read()
	if !want.matches(updated) || !sdnUnchanged(before, updated) {
		t.Fatal("meeting broadcast update changed unbound SDN settings")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("meeting broadcast original settings restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global", broadcast.ReadOptions{})
	original := meetingBroadcastSnapshot{Allow: before.AllowSdnProviderForBroadcastMeeting, SupportURL: before.SupportURL}
	if err != nil || !original.matches(restored) {
		t.Fatal("meeting broadcast original state not independently restored")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_configuration", "restored_readback", response, nil)
	restored = read()
	if !original.matches(restored) || !sdnUnchanged(before, restored) {
		t.Fatal("meeting broadcast restoration changed SDN settings")
	}
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitMeetingBroadcastRecoveryRejectsForeignTargetsAndUnsupportedOriginals(t *testing.T) {
	t.Parallel()
	for _, entry := range []lab.Entry{
		{ResourceID: "Tag:foreign", Snapshot: json.RawMessage(`{"allow_sdn_provider":false,"support_url":"https://example.invalid"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_sdn_provider":false}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"support_url":"https://example.invalid"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_sdn_provider":null,"support_url":"https://example.invalid"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_sdn_provider":"false","support_url":"https://example.invalid"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_sdn_provider":false,"support_url":1}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_sdn_provider":false,"support_url":"https://example.invalid","sdn_api_token":"x"}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"allow_sdn_provider":false,"support_url":"https://example.invalid"} {}`)},
	} {
		if !errors.Is((meetingBroadcastConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
}
