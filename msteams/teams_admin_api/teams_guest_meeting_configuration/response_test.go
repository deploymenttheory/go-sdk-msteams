package teamsguestmeetingconfiguration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_meeting_configuration"
)

// typedSettings independently decodes the eight getters from a real response.
type typedSettings struct {
	AllowIPVideo                               *bool   `json:"AllowIPVideo"`
	ScreenSharingMode                          *string `json:"ScreenSharingMode"`
	AllowMultipleScreenshare                   *bool   `json:"AllowMultipleScreenshare"`
	AllowMeetNow                               *bool   `json:"AllowMeetNow"`
	LiveCaptionsEnabledType                    *string `json:"LiveCaptionsEnabledType"`
	AllowTranscription                         *bool   `json:"AllowTranscription"`
	AllowParticipantGiveRequestControl         *bool   `json:"AllowParticipantGiveRequestControl"`
	AllowExternalParticipantGiveRequestControl *bool   `json:"AllowExternalParticipantGiveRequestControl"`
}

func project(v *resource.Configuration) typedSettings {
	return typedSettings{v.AllowIPVideo, v.ScreenSharingMode, v.AllowMultipleScreenshare, v.AllowMeetNow, v.LiveCaptionsEnabledType, v.AllowTranscription, v.AllowParticipantGiveRequestControl, v.AllowExternalParticipantGiveRequestControl}
}

func responseFixtureSteps(t *testing.T, path string, status int, body []byte) []contracttest.Exchange {
	t.Helper()
	const origin = "https://api.interfaces.records.teams.microsoft.com"
	var steps []contracttest.Exchange
	return append(steps, contracttest.Exchange{
		Request: contracttest.WireRequest{Method: http.MethodGet, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-CmdletName": {"Get-CsTeamsGuestMeetingConfiguration"}, "MPACmdlet": {"true"}, "X-MS-Target-Uri": nil}},
		Reply:   contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body},
	})
}

func TestUnitGetGlobalResponseFixture(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global.json")
	var expected typedSettings
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	result, err := contracttest.Run(t.Context(), contracttest.Case{
		Name: "get-global", Family: "teams_guest_meeting_configuration", Scope: "service", Source: "msteams/teams_admin_api/teams_guest_meeting_configuration/testdata/get_global.json",
		Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsGuestMeetingConfiguration", ParameterSet: "Identity"}},
		Steps:    responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsGuestMeetingConfiguration/configuration/Global", 200, raw),
		Want:     contracttest.Expected{Status: http.StatusOK, Body: raw, JSON: raw},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations.GetByIdentity(ctx, "Global")
			if err == nil {
				if value == nil || value.Identity == nil || *value.Identity != "Global" || !reflect.DeepEqual(project(value), expected) {
					t.Fatal("typed identity or settings differ from the real response")
				}
				return contracttest.Observation{Value: value.RawFields, Response: response}
			}
			return contracttest.Observation{Response: response, Err: err}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("permanent response fixture differs: %v %+v", err, result.Differences())
	}
}

func TestUnitListResponseFixture(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	var expected []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &expected); err != nil || expected == nil {
		t.Fatal("actual list fixture must retain a JSON array")
	}
	wantSettings := make(map[string]typedSettings, len(expected))
	for _, snapshot := range expected {
		var item struct {
			Identity string `json:"Identity"`
			typedSettings
		}
		encoded, err := json.Marshal(snapshot)
		if err != nil || json.Unmarshal(encoded, &item) != nil || item.Identity == "" {
			t.Fatal("invalid typed list fixture projection")
		}
		wantSettings[item.Identity] = item.typedSettings
	}
	// The native List contract stably places Global before other identities.
	slices.SortStableFunc(expected, func(a, b map[string]json.RawMessage) int {
		var left, right string
		if json.Unmarshal(a["Identity"], &left) != nil || json.Unmarshal(b["Identity"], &right) != nil {
			t.Fatal("real list fixture has no comparable identity")
		}
		globalLeft, globalRight := strings.EqualFold(left, "Global"), strings.EqualFold(right, "Global")
		if globalLeft == globalRight {
			return 0
		}
		if globalLeft {
			return -1
		}
		return 1
	})
	want, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	result, err := contracttest.Run(t.Context(), contracttest.Case{
		Name: "list", Family: "teams_guest_meeting_configuration", Scope: "service", Source: "msteams/teams_admin_api/teams_guest_meeting_configuration/testdata/list.json",
		Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsGuestMeetingConfiguration", ParameterSet: "Identity"}},
		Steps:    responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsGuestMeetingConfiguration", 200, raw),
		Want:     contracttest.Expected{Status: http.StatusOK, Body: raw, JSON: want},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations.List(ctx)
			if err == nil {
				if len(values) != len(expected) {
					t.Fatal("real list item count changed")
				}
				for _, value := range values {
					if value == nil || value.Identity == nil || *value.Identity == "" {
						t.Fatal("list lost its typed resource identity")
					}
					settings, present := wantSettings[*value.Identity]
					if !present || !reflect.DeepEqual(project(value), settings) {
						t.Fatal("typed settings differ from the real list response")
					}
				}
			}
			return contracttest.Observation{Value: values, Response: response, Err: err}
		},
	})
	if err != nil || !result.Passed() {
		t.Fatalf("permanent response fixture differs: %v %+v", err, result.Differences())
	}
}

// The update and restoration read-backs are the actual persisted states; the
// toggled AllowMeetNow differs while every other getter is preserved.
func TestUnitUpdateReadBacksRetainOmittedSettings(t *testing.T) {
	t.Parallel()
	decode := func(name string) typedSettings {
		var v resource.Configuration
		if err := json.Unmarshal(readResponseFixture(t, name), &v); err != nil {
			t.Fatal(err)
		}
		return project(&v)
	}
	before, updated, restored := decode("get_global.json"), decode("update_readback.json"), decode("restored_readback.json")
	if before.AllowMeetNow == nil || updated.AllowMeetNow == nil || *before.AllowMeetNow == *updated.AllowMeetNow {
		t.Fatal("update read-back did not change AllowMeetNow")
	}
	updated.AllowMeetNow = before.AllowMeetNow
	if !reflect.DeepEqual(before, updated) || !reflect.DeepEqual(before, restored) {
		t.Fatal("omitted settings or restoration differ from the original")
	}
}
