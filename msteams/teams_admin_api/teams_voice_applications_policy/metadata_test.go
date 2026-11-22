package teamsvoiceapplicationspolicy_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_voice_applications_policy"
)

func TestUnitSettingsMatchPinnedCommandParameters(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline struct {
		Commands []struct {
			Name          string
			ParameterSets []struct{ Parameters []struct{ Name, Type string } } `json:"parameter_sets"`
		}
	}
	if json.Unmarshal(raw, &baseline) != nil {
		t.Fatal("invalid pinned command metadata")
	}
	settings := reflect.TypeFor[resource.Settings]()
	response := reflect.TypeFor[resource.Policy]()
	for _, command := range []string{"New-CsTeamsVoiceApplicationsPolicy", "Set-CsTeamsVoiceApplicationsPolicy"} {
		found := false
		for _, cmd := range baseline.Commands {
			if cmd.Name != command {
				continue
			}
			found = true
			for _, set := range cmd.ParameterSets {
				count := 0
				for _, param := range set.Parameters {
					switch param.Name {
					case "Identity", "Force", "Verbose", "Debug", "ErrorAction", "WarningAction", "InformationAction", "ProgressAction", "ErrorVariable", "WarningVariable", "InformationVariable", "OutVariable", "OutBuffer", "PipelineVariable", "WhatIf", "Confirm":
						continue
					}
					field, ok := settings.FieldByName(param.Name)
					if !ok || strings.Split(field.Tag.Get("json"), ",")[0] != param.Name {
						t.Fatalf("missing setting %s", param.Name)
					}
					read, ok := response.FieldByName(param.Name)
					if !ok || read.Type.Kind() != reflect.Pointer {
						t.Fatalf("missing nullable read field %s", param.Name)
					}
					want := reflect.String
					switch {
					case strings.Contains(param.Type, "System.Boolean"):
						want = reflect.Bool
					case strings.Contains(param.Type, "System.UInt32"):
						want = reflect.Uint32
					case strings.Contains(param.Type, "System.Int64"):
						want = reflect.Int64
					}
					if read.Type.Elem().Kind() != want {
						t.Fatalf("wrong baseline type for %s", param.Name)
					}
					count++
				}
				if count != settings.NumField() {
					t.Fatal("settings and pinned command surface differ")
				}
			}
		}
		if !found {
			t.Fatal("missing baseline command")
		}
	}
}
