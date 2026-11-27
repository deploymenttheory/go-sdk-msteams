package teamsmeetingconfigurationacc

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	meeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// Non-default acceptance values. Both are inert display strings: a footer
// line on meeting invitations and the join-page help link.
const (
	acceptanceFooter  = "msteams-sdk acceptance footer"
	acceptanceHelpURL = "https://example.invalid/msteams-sdk-help"
)

// TestAccMeetingConfigurationUpdateRestore changes only CustomFooterText and
// HelpURL on Global through Set's read-then-PATCH composition, verifies the
// other 19 getters are preserved, and restores the journaled originals
// (including an explicit null) with an independent read-back.
func TestAccMeetingConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid meeting configuration enablement")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting configuration update requires explicit mutation enablement")
	} else if err != nil {
		t.Fatal("invalid meeting configuration mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsMeetingConfigurations
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private meeting configuration recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"meeting-configuration": meetingConfigurationCleanup{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("meeting configuration restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("meeting configuration recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || before == nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("original meeting configuration unavailable; update refused")
	}
	if (before.CustomFooterText != nil && *before.CustomFooterText == acceptanceFooter) || (before.HelpURL != nil && *before.HelpURL == acceptanceHelpURL) {
		t.Fatal("meeting configuration already holds acceptance values; recover the journal first")
	}
	compare := func(v *meeting.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.SameProjection(meetingProjection(v), meetingOracle(t, ctx, tenant)) {
			t.Fatal("Go and PowerShell meeting configuration getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]*string{"CustomFooterText": before.CustomFooterText, "HelpURL": before.HelpURL})
	if err != nil {
		t.Fatal("meeting configuration originals could not be journaled")
	}
	entry, err := journal.PlanRestore("meeting-configuration", "Global", snapshot)
	if err != nil {
		t.Fatal("meeting configuration restoration intent not journaled")
	}
	footer, help := acceptanceFooter, acceptanceHelpURL
	_, response, err = s.UpdateByIdentity(ctx, "Global", meeting.UpdateRequest{ScalarSettings: meeting.ScalarSettings{
		CustomFooterText: policy.Set(footer),
		HelpURL:          policy.Set(help),
	}})
	labkit.WriteResponseFixture(t, "teams_meeting_configuration", "update", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting configuration update failed; restoration journal retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("meeting configuration update receipt not journaled")
	}
	want := maps.Clone(meetingProjection(before))
	want["CustomFooterText"], want["HelpURL"] = &footer, &help
	var updated *meeting.Configuration
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, e := s.GetByIdentity(ctx, "Global")
		matched := e == nil && v != nil && labkit.SameProjection(want, meetingProjection(v))
		if matched {
			updated = v
			labkit.WriteResponseFixture(t, "teams_meeting_configuration", "update_readback", r, nil)
		}
		return matched, e
	}); err != nil {
		t.Fatal("meeting configuration update or omitted-field preservation did not converge")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("meeting configuration original settings restoration failed")
	}
	restored, response, err := s.GetByIdentity(ctx, "Global")
	if err != nil || restored == nil || !labkit.SameProjection(meetingProjection(before), meetingProjection(restored)) {
		t.Fatal("meeting configuration original state not independently restored")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_configuration", "restored_readback", response, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitMeetingRecoveryRejectsForeignTargetsAndMalformedSnapshots(t *testing.T) {
	t.Parallel()
	for _, entry := range []lab.Entry{
		{ResourceID: "Tag:foreign", Snapshot: json.RawMessage(`{"CustomFooterText":"","HelpURL":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"CustomFooterText":""}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"HelpURL":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"CustomFooterText":1,"HelpURL":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"CustomFooterText":"","HelpURL":false}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"CustomFooterText":"","HelpURL":null,"extra":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"CustomFooterText":"","HelpURL":null} {}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`null`)},
	} {
		if !errors.Is((meetingConfigurationCleanup{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported restoration reached the service")
		}
	}
	for _, raw := range []string{`{"CustomFooterText":"","HelpURL":null}`, `{"CustomFooterText":null,"HelpURL":"https://example.invalid"}`} {
		if _, ok := decodeMeetingSnapshot([]byte(raw)); !ok {
			t.Fatal("valid nullable snapshot rejected")
		}
	}
}

func TestUnitMeetingOracleRejectsMixedAndIncompleteOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000001"
	settings := map[string]any{}
	for _, name := range meetingSettings {
		settings[name] = nil
	}
	complete, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": "Global", "settings": settings})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeMeetingOracle(complete, tenant); err != nil {
		t.Fatal("complete oracle report rejected")
	}
	delete(settings, "ReportMeeting")
	incomplete, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": "Global", "settings": settings})
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range [][]byte{
		incomplete,
		append(append([]byte(nil), complete...), []byte(" {}")...),
		[]byte(strings.Replace(string(complete), tenant, "00000000-0000-4000-8000-000000000002", 1)),
		[]byte(strings.Replace(string(complete), `"success":true`, `"success":false`, 1)),
		[]byte(strings.Replace(string(complete), `"identity":"Global"`, `"identity":"Tag:other"`, 1)),
		[]byte(`{"success":true,"unexpected":1}`),
	} {
		if _, err := decodeMeetingOracle(raw, tenant); err == nil {
			t.Fatal("unverified oracle report accepted")
		}
	}
}
