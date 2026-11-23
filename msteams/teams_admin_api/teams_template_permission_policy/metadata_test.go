package teamstemplatepermissionpolicy_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestUnitSettingsCoverEveryPinnedWriteParameter(t *testing.T) {
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
		t.Fatal("invalid pinned runtime metadata")
	}
	commands := map[string]bool{"New-CsTeamsTemplatePermissionPolicy": false, "Set-CsTeamsTemplatePermissionPolicy": false}
	wanted := []string{"Description", "HiddenTemplates"}
	common := map[string]bool{}
	for _, name := range []string{"Identity", "Force", "Verbose", "Debug", "ErrorAction", "WarningAction", "InformationAction", "ProgressAction", "ErrorVariable", "WarningVariable", "InformationVariable", "OutVariable", "OutBuffer", "PipelineVariable", "WhatIf", "Confirm"} {
		common[name] = true
	}
	for _, command := range baseline.Commands {
		if _, ok := commands[command.Name]; !ok {
			continue
		}
		commands[command.Name] = true
		for _, set := range command.Sets {
			actual := []string{}
			for _, parameter := range set.Parameters {
				if !common[parameter.Name] {
					actual = append(actual, parameter.Name)
				}
			}
			if !reflect.DeepEqual(actual, wanted) {
				t.Fatal("SDK write fields differ from the pinned parameter set")
			}
		}
	}
	for _, found := range commands {
		if !found {
			t.Fatal("pinned write command missing")
		}
	}
}
