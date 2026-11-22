package teamstenantabuseconfiguration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_tenant_abuse_configuration"
)

func TestUnitAllPinnedSettingsBindExactValuesAndNull(t *testing.T) {
	t.Parallel()
	values := map[string]json.RawMessage{"LowSeatLimit": json.RawMessage(`18446744073709551615`),
		"TrialTenantValidation":                json.RawMessage(`"future-value"`),
		"SkipValidation":                       json.RawMessage(`""`),
		"CreateThreadThresholdPerSeat":         json.RawMessage(`0`),
		"AddMemberThresholdPerSeat":            json.RawMessage(`9007199254740993`),
		"SendMessageThresholdPerSeat":          json.RawMessage(`18446744073709551615`),
		"CreateThreadThresholdForTrialTenants": json.RawMessage(`0`),
		"AddMemberThresholdForTrialTenants":    json.RawMessage(`1`),
		"SendMessageThresholdForTrialTenants":  json.RawMessage(`9007199254740993`)}
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
		input["BlockedTenants"] = json.RawMessage(`["future-value"]`)
		input["Identity"] = json.RawMessage(`"fixture-policy"`)
		input["Force"] = json.RawMessage(`"False"`)
		body, e := json.Marshal(input)
		if e != nil {
			t.Fatal(e)
		}
		run(t, contracttest.Case{Name: "all-bound", Steps: []contracttest.Exchange{compiledWire("PATCH", "/Skype.Policy/configurations/TeamsTenantAbuseConfiguration/configuration/fixture-policy", "Set-CsTeamsTenantAbuseConfiguration", body, 503, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: settings, BlockedTenants: policy.Strings("future-value")}, Controls: policy.Controls{Force: new(false)}})
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
func TestUnitWriteFieldsMatchPinnedRuntime(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if e != nil {
		t.Fatal(e)
	}
	var baseline struct {
		Commands []struct {
			Name string
			Sets []struct{ Parameters []struct{ Name string } } `json:"parameter_sets"`
		}
	}
	if json.Unmarshal(raw, &baseline) != nil {
		t.Fatal("invalid pinned metadata")
	}
	common := map[string]bool{}
	for _, name := range []string{"Identity", "Force", "Verbose", "Debug", "ErrorAction", "WarningAction", "InformationAction", "ProgressAction", "ErrorVariable", "WarningVariable", "InformationVariable", "OutVariable", "OutBuffer", "PipelineVariable", "WhatIf", "Confirm"} {
		common[name] = true
	}
	typ := reflect.TypeFor[resource.ScalarSettings]()
	expected := []string{"BlockedTenants"}
	for i := 0; i < typ.NumField(); i++ {
		expected = append(expected, typ.Field(i).Name)
	}
	sort.Strings(expected)
	found := false
	for _, command := range baseline.Commands {
		if command.Name != "Set-CsTeamsTenantAbuseConfiguration" {
			continue
		}
		found = true
		for _, set := range command.Sets {
			actual := []string{}
			for _, parameter := range set.Parameters {
				if !common[parameter.Name] {
					actual = append(actual, parameter.Name)
				}
			}
			sort.Strings(actual)
			if !reflect.DeepEqual(actual, expected) {
				t.Fatal("SDK fields differ from pinned parameter set")
			}
		}
	}
	if !found {
		t.Fatal("pinned write command missing")
	}
}
