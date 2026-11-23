package teamsupdatemanagementpolicy_test

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
		run(t, contracttest.Case{Name: "all-bound", Steps: []contracttest.Exchange{compiledWire("PATCH", "/Skype.Policy/configurations/TeamsUpdateManagementPolicy/configuration/fixture-policy", "Set-CsTeamsUpdateManagementPolicy", body, 503, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: settings, DisabledInProductMessages: policy.Strings("future-value")}, Controls: policy.Controls{Force: new(false)}})
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
	expected := []string{"DisabledInProductMessages"}
	for i := 0; i < typ.NumField(); i++ {
		expected = append(expected, typ.Field(i).Name)
	}
	sort.Strings(expected)
	found := false
	for _, command := range baseline.Commands {
		if command.Name != "Set-CsTeamsUpdateManagementPolicy" {
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
