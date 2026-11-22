package teamsvoiceapplicationspolicy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_voice_applications_policy"
)

func TestUnitAllPinnedSettingsBindValuesAndNull(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline struct {
		Commands []struct {
			Name string
			Sets []struct{ Parameters []struct{ Name, Type string } } `json:"parameter_sets"`
		}
	}
	if json.Unmarshal(raw, &baseline) != nil {
		t.Fatal("invalid baseline")
	}
	values := map[string]json.RawMessage{}
	for _, cmd := range baseline.Commands {
		if cmd.Name != "New-CsTeamsVoiceApplicationsPolicy" {
			continue
		}
		for _, param := range cmd.Sets[0].Parameters {
			switch param.Name {
			case "Identity", "Force", "Verbose", "Debug", "ErrorAction", "WarningAction", "InformationAction", "ProgressAction", "ErrorVariable", "WarningVariable", "InformationVariable", "OutVariable", "OutBuffer", "PipelineVariable", "WhatIf", "Confirm":
				continue
			}
			value := `"future-value"`
			switch {
			case strings.Contains(param.Type, "System.Boolean"):
				value = "false"
			}
			values[param.Name] = json.RawMessage(value)
		}
	}
	if len(values) == 0 {
		t.Fatal("missing baseline settings")
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
			method, command, path := "POST", "New-CsTeamsVoiceApplicationsPolicy", "/Skype.Policy/configurations/TeamsVoiceApplicationsPolicy"
			if operation == "patch" {
				method, command, path = "PATCH", "Set-CsTeamsVoiceApplicationsPolicy", path+"/configuration/fixture-policy"
			}
			run(t, contracttest.Case{Name: operation + "-all-bound", Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire(method, path, command, body, 503, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				controls := policy.Controls{Force: new(false)}
				if operation == "create" {
					_, r, e := sdk.TeamsAdminAPI.TeamsVoiceApplicationsPolicies.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{Settings: settings, Controls: controls})
					return contracttest.Observation{Response: r, Err: e}
				}
				_, r, e := sdk.TeamsAdminAPI.TeamsVoiceApplicationsPolicies.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: settings, Controls: controls})
				return contracttest.Observation{Response: r, Err: e}
			}})
		}
	}
}
