package teamssipdevicesconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_sip_devices_configuration"
)

const collection = "/Skype.Policy/configurations/TeamsSipDevicesConfiguration"
const selected = collection + "/configuration/Global"
const set = "Set-CsTeamsSipDevicesConfiguration"

func TestUnitUpdateUsesActualPreparationAndSparsePatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	run(t, testutil.Case{Name: "update", Steps: []testutil.Exchange{
		compiledWire("GET", selected, set, nil, 200, string(before)),
		compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","BulkSignIn":"Enabled"}`), 204, string(readResponseFixture(t, "update_204.body"))),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{BulkSignIn: policy.Set("Enabled")}})
		if e == nil && (v == nil || v.Before == nil || v.Before.BulkSignIn == nil || *v.Before.BulkSignIn != "Disabled" || v.ReadResponse == nil || v.ReadResponse.StatusCode() != 200 || !v.PatchInvoked || !v.HTTPAccepted) {
			t.Fatal("preparation or mutation acceptance receipt lost")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitNullEmptyAndForceRemainExplicit(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		field      policy.Field[string]
	}{{"null", `{"Identity":"Global","BulkSignIn":null,"Force":"False"}`, policy.Null[string]()}, {"empty", `{"Identity":"Global","BulkSignIn":"","Force":"False"}`, policy.Set("")}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(tc.body), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{BulkSignIn: tc.field}, Controls: policy.Controls{Force: new(false)}})
				return testutil.Observation{Response: r, Err: e}
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
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", collection, "Get-CsTeamsSipDevicesConfiguration", nil, 200, string(readResponseFixture(t, "list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				values, r, e := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations.ListFiltered(ctx, tc.pattern)
				if e == nil && len(values) != tc.count {
					t.Fatal("local filter result differs")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}
