package policies_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	guest "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_meeting_configuration"
	meeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
)

// Each read reuses a retained primitive acceptance response. The cross-family
// sequence, intermediate outcomes and caller recovery are authored offline.
// HTTP acceptance/readback replay is not actual persistence or a transaction.
func TestUnitMeetingAndGuestConfigurationFailuresKeepCallerRecoveryIndependent(t *testing.T) {
	t.Parallel()
	mb := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_meeting_configuration/testdata/get_global.json")
	mu := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_meeting_configuration/testdata/update_readback.json")
	mr := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_meeting_configuration/testdata/restored_readback.json")
	gb := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_guest_meeting_configuration/testdata/get_global.json")
	gu := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_guest_meeting_configuration/testdata/update_readback.json")
	gr := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_guest_meeting_configuration/testdata/restored_readback.json")
	denial := pol91ResponseFixture(t, "msteams/teams_admin_api/privacy_configuration/testdata/get_global_403.json")
	var before, updated meeting.Configuration
	var guestBefore, guestUpdated guest.Configuration
	if json.Unmarshal(mb, &before) != nil || json.Unmarshal(mu, &updated) != nil || json.Unmarshal(gb, &guestBefore) != nil || json.Unmarshal(gu, &guestUpdated) != nil || before.CustomFooterText == nil || *before.CustomFooterText != "" || before.HelpURL != nil || updated.CustomFooterText == nil || updated.HelpURL == nil || guestBefore.AllowMeetNow == nil || !*guestBefore.AllowMeetNow || guestUpdated.AllowMeetNow == nil || *guestUpdated.AllowMeetNow {
		t.Fatal("permanent preparation/readback contract unavailable")
	}
	body := func(v map[string]any) json.RawMessage {
		t.Helper()
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	for _, tc := range []struct {
		name            string
		status          int
		reply, observed []byte
		want            error
		accepted        bool
		observedEnabled bool
	}{
		{"denied", 403, denial, gb, auth.ErrForbidden, false, true},
		{"accepted-malformed", 200, []byte(`{"AllowMeetNow":{}}`), gu, client.ErrDecode, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := contracttest.Case{Name: "meeting-guest-" + tc.name + "-recovery", Family: "pol91", Scope: "service", Source: "tests/contract/policies/meeting_configuration_recovery_test.go", Mappings: []contracttest.Mapping{
				{Command: "Get-CsTeamsMeetingConfiguration", ParameterSet: "Identity"}, {Command: "Set-CsTeamsMeetingConfiguration", ParameterSet: "Identity"},
				{Command: "Get-CsTeamsGuestMeetingConfiguration", ParameterSet: "Identity"}, {Command: "Set-CsTeamsGuestMeetingConfiguration", ParameterSet: "Identity"},
			}, Steps: []contracttest.Exchange{
				pol91ConfigurationWire("TeamsMeetingConfiguration", "GET", "Set-CsTeamsMeetingConfiguration", nil, 200, mb),
				pol91ConfigurationWire("TeamsMeetingConfiguration", "PATCH", "Set-CsTeamsMeetingConfiguration", body(map[string]any{"Identity": "Global", "CustomFooterText": updated.CustomFooterText, "HelpURL": updated.HelpURL}), 204, nil),
				pol91ConfigurationWire("TeamsGuestMeetingConfiguration", "GET", "Set-CsTeamsGuestMeetingConfiguration", nil, 200, gb),
				pol91ConfigurationWire("TeamsGuestMeetingConfiguration", "PATCH", "Set-CsTeamsGuestMeetingConfiguration", json.RawMessage(`{"Identity":"Global","AllowMeetNow":false}`), tc.status, tc.reply),
				pol91ConfigurationWire("TeamsGuestMeetingConfiguration", "GET", "Get-CsTeamsGuestMeetingConfiguration", nil, 200, tc.observed),
				pol91ConfigurationWire("TeamsGuestMeetingConfiguration", "PATCH", "Set-CsTeamsGuestMeetingConfiguration", json.RawMessage(`{"Identity":"Global","AllowMeetNow":true}`), 204, nil),
				pol91ConfigurationWire("TeamsGuestMeetingConfiguration", "GET", "Get-CsTeamsGuestMeetingConfiguration", nil, 200, gr),
				pol91ConfigurationWire("TeamsMeetingConfiguration", "PATCH", "Set-CsTeamsMeetingConfiguration", json.RawMessage(`{"Identity":"Global","CustomFooterText":"","HelpURL":null}`), 204, nil),
				pol91ConfigurationWire("TeamsMeetingConfiguration", "GET", "Get-CsTeamsMeetingConfiguration", nil, 200, mr),
			}, Want: contracttest.Expected{Status: 200, Body: mr}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				meetings, guests := sdk.TeamsAdminAPI.TeamsMeetingConfigurations, sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations
				first, response, err := meetings.UpdateByIdentity(ctx, "Global", meeting.UpdateRequest{ScalarSettings: meeting.ScalarSettings{CustomFooterText: policy.Set(*updated.CustomFooterText), HelpURL: policy.Set(*updated.HelpURL)}})
				if err != nil || response == nil || response.StatusCode() != 204 || first == nil || first.Before == nil || first.Before.Identity == nil || *first.Before.Identity != "Global" || first.Before.CustomFooterText == nil || *first.Before.CustomFooterText != "" || first.Before.HelpURL != nil || first.ReadResponse == nil || !bytes.Equal(first.ReadResponse.Body(), mb) || !first.PatchInvoked || !first.HTTPAccepted || first.After != nil {
					t.Fatal("meeting acceptance lost typed preparation or fabricated persistence")
				}
				failed, response, err := guests.UpdateByIdentity(ctx, "Global", guest.UpdateRequest{Settings: guest.Settings{AllowMeetNow: policy.Set(false)}})
				if !errors.Is(err, tc.want) || response == nil || response.StatusCode() != tc.status || !bytes.Equal(response.Body(), tc.reply) || failed == nil || failed.Before == nil || failed.Before.AllowMeetNow == nil || !*failed.Before.AllowMeetNow || failed.ReadResponse == nil || !bytes.Equal(failed.ReadResponse.Body(), gb) || !failed.PatchInvoked || failed.HTTPAccepted != tc.accepted || failed.After != nil {
					t.Fatal("later guest failure confused families or lost its actual receipt/acceptance")
				}
				observed, response, err := guests.GetByIdentity(ctx, "Global")
				if err != nil || response == nil || observed == nil || observed.AllowMeetNow == nil || *observed.AllowMeetNow != tc.observedEnabled || !bytes.Equal(response.Body(), tc.observed) {
					t.Fatal("explicit reconciliation read lost its own observed state")
				}
				value, response, err := guests.PatchByIdentity(ctx, "Global", guest.PatchRequest{Settings: guest.Settings{AllowMeetNow: policy.Set(*guestBefore.AllowMeetNow)}})
				if err != nil || response == nil || response.StatusCode() != 204 || value != nil {
					t.Fatal("caller-selected guest restoration failed or fabricated a value")
				}
				restoredGuest, response, err := guests.GetByIdentity(ctx, "Global")
				if err != nil || response == nil || restoredGuest == nil || restoredGuest.AllowMeetNow == nil || !*restoredGuest.AllowMeetNow {
					t.Fatal("separate guest restoration readback unavailable")
				}
				restoredValue, response, err := meetings.PatchByIdentity(ctx, "Global", meeting.PatchRequest{Settings: meeting.Settings{ScalarSettings: meeting.ScalarSettings{CustomFooterText: policy.Set(*before.CustomFooterText), HelpURL: policy.Null[string]()}}})
				if err != nil || response == nil || response.StatusCode() != 204 || restoredValue != nil {
					t.Fatal("caller-selected meeting restoration failed")
				}
				readback, response, err := meetings.GetByIdentity(ctx, "Global")
				if err == nil && (response == nil || readback == nil || readback.HelpURL != nil || readback.CustomFooterText == nil || *readback.CustomFooterText != "" || !first.HTTPAccepted || first.After != nil || !bytes.Equal(first.Before.RawJSON, bytes.TrimSpace(mb)) || failed.HTTPAccepted != tc.accepted || !bytes.Equal(failed.Before.RawJSON, bytes.TrimSpace(gb)) || !*restoredGuest.AllowMeetNow) {
					t.Fatal("restoration mutated earlier snapshots or confused receipt and effective state")
				}
				return contracttest.Observation{Response: response, Err: err}
			}}
			result, err := contracttest.Run(t.Context(), c)
			if err != nil || !result.Passed() {
				t.Fatalf("meeting/guest recovery differs: %v %+v", err, result.Differences())
			}
		})
	}
}
