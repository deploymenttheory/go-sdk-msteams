package teamsupgradeconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_configuration"
)

const collection = "/Skype.Policy/configurations/TeamsUpgradeConfiguration"
const selected = collection + "/configuration/Global"
const set = "Set-CsTeamsUpgradeConfiguration"

func TestUnitUpdateUsesActualPreparationAndSparsePatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	after := readResponseFixture(t, "update_readback.json")
	var initial, changed resource.Configuration
	if json.Unmarshal(before, &initial) != nil || json.Unmarshal(after, &changed) != nil || initial.DownloadTeams == nil || changed.DownloadTeams == nil || *initial.DownloadTeams == *changed.DownloadTeams {
		t.Fatal("actual fixtures must show a non-default change")
	}
	body, e := json.Marshal(map[string]any{"Identity": "Global", "DownloadTeams": *changed.DownloadTeams})
	if e != nil {
		t.Fatal(e)
	}
	run(t, testutil.Case{Name: "sparse-update", Steps: []testutil.Exchange{
		compiledWire("GET", selected, set, nil, 200, string(before)), compiledWire("PATCH", selected, set, body, 204, string(readResponseFixture(t, "update_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsUpgradeConfiguration", nil, 200, string(after)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations
		receipt, r, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{DownloadTeams: policy.Set(*changed.DownloadTeams)}})
		if e != nil {
			return testutil.Observation{Response: r, Err: e}
		}
		if receipt == nil || receipt.Before == nil || receipt.ReadResponse == nil || !receipt.PatchInvoked || !receipt.HTTPAccepted || receipt.Before.DownloadTeams == nil || *receipt.Before.DownloadTeams != *initial.DownloadTeams {
			t.Fatal("receipt lost preparation or acceptance")
		}
		value, r, e := s.GetByIdentity(ctx, "Global")
		if e == nil && (value.DownloadTeams == nil || *value.DownloadTeams != *changed.DownloadTeams) {
			t.Fatal("actual non-default typed value lost")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}
func TestUnitAllSettingsPreserveNullFalseOpaqueAndForce(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"DownloadTeams":false,"BlockLegacyAuthorization":false,"SfBMeetingJoinUx":"future-value"}`, `{"DownloadTeams":null,"BlockLegacyAuthorization":null,"SfBMeetingJoinUx":null}`, `{"SfBMeetingJoinUx":""}`} {
		var settings resource.Settings
		var body map[string]any
		if json.Unmarshal([]byte(raw), &settings) != nil || json.Unmarshal([]byte(raw), &body) != nil {
			t.Fatal("invalid source vector")
		}
		body["Identity"] = "Global"
		body["Force"] = "False"
		encoded, e := json.Marshal(body)
		if e != nil {
			t.Fatal(e)
		}
		run(t, testutil.Case{Name: "source-bindings", Steps: []testutil.Exchange{compiledWire("PATCH", selected, set, encoded, 503, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: settings, Controls: policy.Controls{Force: new(false)}})
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}
