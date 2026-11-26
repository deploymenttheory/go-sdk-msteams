package teamsfirstpartymeetingtemplateconfiguration_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
)

func TestUnitSettingsMatchEveryPinnedPublicParameter(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	var baseline struct {
		Commands []struct {
			Name string
			Sets []struct {
				Name       string
				Parameters []struct{ Name string }
			} `json:"parameter_sets"`
		}
	}
	if json.Unmarshal(raw, &baseline) != nil {
		t.Fatal("invalid baseline")
	}
	settings := reflect.TypeFor[resource.Settings]()
	response := reflect.TypeFor[resource.Configuration]()
	count, getSets := 0, 0
	for _, command := range baseline.Commands {
		switch command.Name {
		case get:
			for _, set := range command.Sets {
				if set.Name != "Identity" && set.Name != "Filter" {
					t.Fatal("pinned Get parameter sets changed")
				}
				getSets++
			}
		case set:
			if len(command.Sets) != 1 || command.Sets[0].Name != "Identity" {
				t.Fatal("pinned Set parameter sets changed")
			}
			for _, param := range command.Sets[0].Parameters {
				switch param.Name {
				case "Identity", "Force", "Verbose", "Debug", "ErrorAction", "WarningAction", "InformationAction", "ProgressAction", "ErrorVariable", "WarningVariable", "InformationVariable", "OutVariable", "OutBuffer", "PipelineVariable", "WhatIf", "Confirm":
					continue
				}
				if _, ok := settings.FieldByName(param.Name); !ok {
					t.Fatalf("missing setting %s", param.Name)
				}
				if _, ok := response.FieldByName(param.Name); !ok {
					t.Fatalf("missing response getter %s", param.Name)
				}
				count++
			}
		}
	}
	if count != 2 || getSets != 2 {
		t.Fatal("pinned first-party template configuration surface changed")
	}
	for _, tc := range []struct{ name, body string }{{"value", `{"Description":"x"}`}, {"empty", `{"Description":""}`}, {"null", `{"Description":null}`}, {"omitted", `{}`}} {
		t.Run(tc.name, func(t *testing.T) {
			var bound resource.ScalarSettings
			if json.Unmarshal([]byte(tc.body), &bound) != nil {
				t.Fatal("bound Description rejected")
			}
			out, err := json.Marshal(bound)
			if err != nil || !jsonEqual([]byte(tc.body), out) {
				t.Fatal("bound Description value, null or omission changed")
			}
		})
	}
}
