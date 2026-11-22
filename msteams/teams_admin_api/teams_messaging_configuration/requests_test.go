package teamsmessagingconfiguration_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
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
			Sets []struct{ Parameters []struct{ Name string } } `json:"parameter_sets"`
		}
	}
	if json.Unmarshal(raw, &baseline) != nil {
		t.Fatal("invalid baseline")
	}
	settings := reflect.TypeFor[resource.Settings]()
	response := reflect.TypeFor[resource.Configuration]()
	count := 0
	for _, command := range baseline.Commands {
		if command.Name != set {
			continue
		}
		if len(command.Sets) != 1 {
			t.Fatal("pinned Set parameter sets changed")
		}
		for _, param := range command.Sets[0].Parameters {
			switch param.Name {
			case "Identity", "Force", "Verbose", "Debug", "ErrorAction", "WarningAction", "InformationAction", "ProgressAction", "ErrorVariable", "WarningVariable", "InformationVariable", "OutVariable", "OutBuffer", "PipelineVariable", "WhatIf", "Confirm":
				continue
			}
			write, ok := settings.FieldByName(param.Name)
			if !ok || strings.Split(write.Tag.Get("json"), ",")[0] != param.Name {
				t.Fatalf("missing setting %s", param.Name)
			}
			if _, ok := response.FieldByName(param.Name); !ok {
				t.Fatalf("missing response getter %s", param.Name)
			}
			count++
		}
	}
	if count != 12 {
		t.Fatal("pinned messaging configuration settings changed")
	}
	for _, nulls := range []bool{false, true} {
		t.Run(map[bool]string{false: "values", true: "nulls"}[nulls], func(t *testing.T) {
			var received map[string]json.RawMessage
			if json.Unmarshal(readResponseFixture(t, "get_global.json"), &received) != nil {
				t.Fatal("invalid actual response")
			}
			input := make(map[string]json.RawMessage)
			scalar := reflect.TypeFor[resource.ScalarSettings]()
			for i := range scalar.NumField() {
				name := scalar.Field(i).Name
				input[name] = received[name]
				if name == "Storyline" {
					input[name] = json.RawMessage("1")
				}
				if nulls {
					input[name] = json.RawMessage("null")
				}
			}
			raw, err := json.Marshal(input)
			var bound resource.ScalarSettings
			if err != nil || json.Unmarshal(raw, &bound) != nil {
				t.Fatal("bound scalar settings rejected")
			}
			out, err := json.Marshal(bound)
			if err != nil || !jsonEqual(raw, out) {
				t.Fatal("bound scalar value, null or omission changed")
			}
		})
	}
}

func TestUnitStorylineRetainsSourceEnumValues(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		want       resource.Storyline
	}{{"named", `"Enabled"`, resource.StorylineEnabled}, {"case", `"disabled"`, resource.StorylineDisabled}, {"numeric", "37", 37}, {"numeric-string", `"-7"`, -7}} {
		t.Run(tc.name, func(t *testing.T) {
			var value resource.Storyline
			if json.Unmarshal([]byte(tc.body), &value) != nil || value != tc.want {
				t.Fatal("source enum representation lost")
			}
		})
	}
	for _, invalid := range []string{`"UnknownName"`, "2147483648", "1.5", "{}", "true"} {
		var value resource.Storyline
		if json.Unmarshal([]byte(invalid), &value) == nil {
			t.Fatal("invalid source enum accepted")
		}
	}
}
