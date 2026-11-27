package teamsaipolicy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ai_policy"
)

func TestUnitAllPinnedSettingsBindValuesAndNull(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("testdata/settings_values.json")
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		t.Fatal(err)
	}
	for _, nulls := range []bool{false, true} {
		input := map[string]json.RawMessage{}
		for name, value := range values {
			input[name] = value
			if nulls {
				input[name] = json.RawMessage("null")
			}
		}
		encoded, err := json.Marshal(input)
		var settings resource.Settings
		if err != nil || json.Unmarshal(encoded, &settings) != nil {
			t.Fatal("settings rejected pinned types")
		}
		outbound, err := json.Marshal(settings)
		var fields map[string]json.RawMessage
		if err != nil || json.Unmarshal(outbound, &fields) != nil || len(fields) != len(input) {
			t.Fatal("bound settings were lost")
		}
		for name, value := range input {
			if !bytes.Equal(value, fields[name]) {
				t.Fatalf("bound setting %s changed value or numeric precision", name)
			}
		}
		input["Identity"] = json.RawMessage(`"fixture-policy"`)
		input["Force"] = json.RawMessage(`"False"`)
		body, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		for _, operation := range []string{"create", "patch"} {
			method, command, path := "POST", "New-CsTeamsAIPolicy", "/Skype.Policy/configurations/TeamsAIPolicy"
			if operation == "patch" {
				method, command, path = "PATCH", "Set-CsTeamsAIPolicy", path+"/configuration/fixture-policy"
			}
			run(t, testutil.Case{Name: operation + "-all-bound", Steps: []testutil.Exchange{compiledWire(method, path, command, body, 503, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				controls := policy.Controls{Force: new(false)}
				if operation == "create" {
					_, r, e := sdk.TeamsAdminAPI.TeamsAIPolicies.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{Settings: settings, Controls: controls})
					return testutil.Observation{Response: r, Err: e}
				}
				_, r, e := sdk.TeamsAdminAPI.TeamsAIPolicies.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: settings, Controls: controls})
				return testutil.Observation{Response: r, Err: e}
			}})
		}
	}
}
