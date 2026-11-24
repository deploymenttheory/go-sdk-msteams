package teamsfirstpartymeetingtemplateconfigurationacc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const fixtureFamily = "teams_first_party_meeting_template_configuration"

func TestAccFirstPartyMeetingTemplateConfigurationReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, fixtureFamily, "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("first-party template configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("first-party template configuration identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, fixtureFamily, "list", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("first-party template configuration list failed")
	}
	var selected *resource.Configuration
	for _, item := range values {
		if item != nil && item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous first-party template configuration Global")
			}
			selected = item
		}
	}
	if selected == nil || !labkit.SameProjection(firstPartyTemplateProjection(value), firstPartyTemplateProjection(selected)) {
		t.Fatal("first-party template configuration read/list getters differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("first-party template tenant binding unavailable")
		}
		actual := labkit.MeetingTemplatePolicyOracle(t, ctx, "firstparty", tenant)
		if !labkit.SameProjection(firstPartyTemplateProjection(value), actual.Settings) {
			t.Fatal("Go and PowerShell first-party template configuration getters differ")
		}
	}
	t.Log("first-party template configuration read/list and pinned getter comparison verified")
}

// TestAccFirstPartyMeetingTemplateConfigurationUpdateRestore changes only the
// Global Description, verifies the template list is preserved, and restores the
// original (possibly null) Description through a private recovery journal.
func TestAccFirstPartyMeetingTemplateConfigurationUpdateRestore(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid first-party template mutation configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("first-party template update requires explicit mutation opt-in")
	} else if err != nil {
		t.Fatal("invalid first-party template mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	s := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, r, err := s.GetByIdentity(ctx, "Global")
	if err != nil || before == nil {
		labkit.LogAccessResponse(t, r, err)
		t.Fatal("first-party template original read failed; mutation refused")
	}
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private first-party template journal unavailable")
	}
	handlers := map[string]lab.Handler{"first-party-template-config": firstPartyTemplateDescriptionRestore{s}}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("first-party template restoration unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("first-party template journal close failed")
		}
	})
	compare := func(v *resource.Configuration) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !labkit.SameProjection(firstPartyTemplateProjection(v), labkit.MeetingTemplatePolicyOracle(t, ctx, "firstparty", tenant).Settings) {
			t.Fatal("Go and PowerShell first-party template getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go configuration read parity: passed")
	}
	compare(before)
	snapshot, err := json.Marshal(map[string]*string{"description": before.Description})
	if err != nil {
		t.Fatal("first-party template original could not be journaled")
	}
	entry, err := journal.PlanRestore("first-party-template-config", "Global", snapshot)
	if err != nil {
		t.Fatal("first-party template restore intent not journaled")
	}
	want := "SDK lab non-default updated description"
	if before.Description != nil && *before.Description == want {
		want = "SDK lab non-default initial description"
	}
	_, r, err = s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set(want)}})
	labkit.WriteResponseFixture(t, fixtureFamily, "update", r, nil)
	if err != nil {
		labkit.LogAccessResponse(t, r, err)
		t.Fatal("first-party template update failed; restoration intent retained")
	}
	if err := journal.Applied(entry, "Global"); err != nil {
		t.Fatal("first-party template update receipt not journaled")
	}
	expected := firstPartyTemplateProjection(before)
	expected["Description"] = &want
	var updated *resource.Configuration
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, r, err := s.GetByIdentity(ctx, "Global")
		matched := err == nil && v != nil && labkit.SameProjection(expected, firstPartyTemplateProjection(v))
		if matched {
			updated = v
			labkit.WriteResponseFixture(t, fixtureFamily, "update_readback", r, nil)
		}
		return matched, err
	}); err != nil {
		t.Fatal("first-party template Description persistence or template omission failed")
	}
	t.Log("configuration update/read-back and omission: passed")
	compare(updated)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("first-party template original restoration failed")
	}
	restored, r, err := s.GetByIdentity(ctx, "Global")
	if err != nil || restored == nil || !labkit.SameProjection(firstPartyTemplateProjection(before), firstPartyTemplateProjection(restored)) {
		t.Fatal("first-party template original was not independently restored")
	}
	labkit.WriteResponseFixture(t, fixtureFamily, "restored_readback", r, nil)
	compare(restored)
	t.Log("configuration restoration/read-back: passed")
}

func TestUnitFirstPartyTemplateRecoveryRejectsInvalidSnapshots(t *testing.T) {
	t.Parallel()
	for _, entry := range []lab.Entry{
		{ResourceID: "other", Snapshot: json.RawMessage(`{"description":null}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"description":1}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`{"description":"a","other":true}`)},
		{ResourceID: "Global", Snapshot: json.RawMessage(`null`)},
	} {
		if !errors.Is((firstPartyTemplateDescriptionRestore{}).Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("unsupported recovery target or Description snapshot accepted")
		}
	}
	for _, raw := range []string{`{"description":null}`, `{"description":""}`, `{"description":"x"}`} {
		if _, err := decodeDescriptionSnapshot(lab.Entry{ResourceID: "Global", Snapshot: json.RawMessage(raw)}); err != nil {
			t.Fatal("valid nullable Description snapshot rejected")
		}
	}
}
