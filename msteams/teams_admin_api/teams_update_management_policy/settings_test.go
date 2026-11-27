package teamsupdatemanagementpolicy_test

import (
	"bytes"
	"context"
	"encoding/json"

	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_update_management_policy"
)

func TestUnitAllPinnedSettingsBindExactValuesAndNull(t *testing.T) {
	t.Parallel()
	values := map[string]json.RawMessage{"Description": json.RawMessage(`"future-value"`),
		"AllowManagedUpdates":      json.RawMessage(`false`),
		"AllowPreview":             json.RawMessage(`false`),
		"UpdateDayOfWeek":          json.RawMessage(`6`),
		"UpdateTime":               json.RawMessage(`"22:10"`),
		"UpdateTimeOfDay":          json.RawMessage(`"2026-04-12T18:32:44.1234567"`),
		"AllowPublicPreview":       json.RawMessage(`"future-value"`),
		"UseNewTeamsClient":        json.RawMessage(`""`),
		"BlockLegacyAuthorization": json.RawMessage(`false`)}
	for _, nulls := range []bool{false, true} {
		input := map[string]json.RawMessage{}
		for key, value := range values {
			input[key] = value
			if nulls {
				input[key] = json.RawMessage("null")
			}
		}
		raw, e := json.Marshal(input)
		var settings resource.ScalarSettings
		if e != nil || json.Unmarshal(raw, &settings) != nil {
			t.Fatal("pinned scalar types rejected")
		}
		actual, e := json.Marshal(settings)
		var decoded map[string]json.RawMessage
		if e != nil || json.Unmarshal(actual, &decoded) != nil || len(decoded) != len(input) {
			t.Fatal("bound scalar fields lost")
		}
		for key, value := range input {
			if !bytes.Equal(value, decoded[key]) {
				t.Fatal("bound scalar value or numeric precision changed")
			}
		}
		input["DisabledInProductMessages"] = json.RawMessage(`["future-value"]`)
		input["Identity"] = json.RawMessage(`"fixture-policy"`)
		input["Force"] = json.RawMessage(`"False"`)
		body, e := json.Marshal(input)
		if e != nil {
			t.Fatal(e)
		}
		run(t, testutil.Case{Name: "all-bound", Steps: []testutil.Exchange{compiledWire("PATCH", "/Skype.Policy/configurations/TeamsUpdateManagementPolicy/configuration/fixture-policy", "Set-CsTeamsUpdateManagementPolicy", body, 503, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: settings, DisabledInProductMessages: policy.Strings("future-value")}, Controls: policy.Controls{Force: new(false)}})
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}
