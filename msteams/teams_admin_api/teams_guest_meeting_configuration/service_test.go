package teamsguestmeetingconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_meeting_configuration"
)

const collection = "/Skype.Policy/configurations/TeamsGuestMeetingConfiguration"
const selected = collection + "/configuration/Global"
const get = "Get-CsTeamsGuestMeetingConfiguration"
const set = "Set-CsTeamsGuestMeetingConfiguration"

func TestUnitUpdateUsesActualPreparationAndSparsePatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	run(t, contracttest.Case{Name: "update", Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("GET", selected, set, nil, 200, string(before)),
		compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","AllowMeetNow":false}`), 204, string(readResponseFixture(t, "update_204.body"))),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{AllowMeetNow: policy.Set(false)}})
		if e == nil && (v == nil || v.Before == nil || v.Before.AllowMeetNow == nil || !*v.Before.AllowMeetNow || v.ReadResponse == nil || v.ReadResponse.StatusCode() != 200 || !v.PatchInvoked || !v.HTTPAccepted) {
			t.Fatal("preparation or mutation acceptance receipt lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}

// Every pinned Set parameter binds by its exact name. False and empty values are
// sent rather than dropped, explicit null is retained and Force keeps the
// compiled string form.
func TestUnitEverySettingNullFalseEmptyAndForceRemainExplicit(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		settings   resource.Settings
	}{
		{"all-bound", `{"Identity":"Global","AllowIPVideo":false,"ScreenSharingMode":"SingleApplication","AllowMultipleScreenshare":false,"AllowMeetNow":true,"LiveCaptionsEnabledType":"Disabled","AllowTranscription":true,"AllowParticipantGiveRequestControl":false,"AllowExternalParticipantGiveRequestControl":true,"Force":"False"}`, resource.Settings{
			AllowIPVideo: policy.Set(false), ScreenSharingMode: policy.Set("SingleApplication"), AllowMultipleScreenshare: policy.Set(false), AllowMeetNow: policy.Set(true),
			LiveCaptionsEnabledType: policy.Set("Disabled"), AllowTranscription: policy.Set(true), AllowParticipantGiveRequestControl: policy.Set(false), AllowExternalParticipantGiveRequestControl: policy.Set(true),
		}},
		{"null", `{"Identity":"Global","AllowTranscription":null,"LiveCaptionsEnabledType":null,"Force":"False"}`, resource.Settings{AllowTranscription: policy.Null[bool](), LiveCaptionsEnabledType: policy.Null[string]()}},
		{"empty", `{"Identity":"Global","ScreenSharingMode":"","Force":"False"}`, resource.Settings{ScreenSharingMode: policy.Set("")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(tc.body), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: tc.settings, Controls: policy.Controls{Force: new(false)}})
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitFilterUsesActualCollection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"global", "Glo?al", 1}, {"missing", "missing", 0}, {"literal-whitespace", " ", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: get, ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, get, nil, 200, string(readResponseFixture(t, "list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				values, r, e := sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations.ListFiltered(ctx, tc.pattern)
				if e == nil && len(values) != tc.count {
					t.Fatal("local filter result differs")
				}
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}
