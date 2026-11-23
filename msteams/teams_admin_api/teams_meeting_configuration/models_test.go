package teamsmeetingconfiguration_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
)

func TestUnitEveryTypedSettingUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"get_global.json", "update_readback.json", "restored_readback.json"} {
		t.Run(file, func(t *testing.T) {
			raw := readResponseFixture(t, file)
			var value resource.Configuration
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			fields := reflect.ValueOf(value)
			typ := fields.Type()
			checked := 0
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				name := field.Tag.Get("json")
				if name == "-" {
					continue
				}
				actual, ok := value.RawFields[name]
				if !ok {
					t.Fatalf("tenant response missing typed field %s", name)
				}
				pointer := reflect.New(field.Type)
				if json.Unmarshal(actual, pointer.Interface()) != nil || !reflect.DeepEqual(fields.Field(i).Interface(), pointer.Elem().Interface()) {
					t.Fatalf("typed field %s differs from acceptance response", name)
				}
				checked++
			}
			if checked != 22 {
				t.Fatal("typed setting count changed")
			}
			snapshot := append([]byte(nil), raw...)
			clear(raw)
			marshaled, err := json.Marshal(value)
			if err != nil || !jsonEqual(snapshot, marshaled) || !bytes.Equal(bytes.TrimSpace(snapshot), value.RawJSON) {
				t.Fatal("model lost owned response snapshot")
			}
		})
	}
}

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
				Parameters []struct {
					Name      string
					Mandatory bool
				}
			} `json:"parameter_sets"`
		}
	}
	if json.Unmarshal(raw, &baseline) != nil {
		t.Fatal("invalid baseline")
	}
	settings := reflect.TypeFor[resource.Settings]()
	response := reflect.TypeFor[resource.Configuration]()
	count, sets := 0, map[string][]string{}
	for _, command := range baseline.Commands {
		for _, parameterSet := range command.Sets {
			if command.Name == get || command.Name == set {
				sets[command.Name] = append(sets[command.Name], parameterSet.Name)
			}
		}
		if command.Name != set {
			continue
		}
		if len(command.Sets) != 1 {
			t.Fatal("pinned Set parameter sets changed")
		}
		for _, param := range command.Sets[0].Parameters {
			switch param.Name {
			case "Identity":
				if !param.Mandatory {
					t.Fatal("pinned Set Identity is no longer mandatory")
				}
				continue
			case "Force", "Verbose", "Debug", "ErrorAction", "WarningAction", "InformationAction", "ProgressAction", "ErrorVariable", "WarningVariable", "InformationVariable", "OutVariable", "OutBuffer", "PipelineVariable", "WhatIf", "Confirm":
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
	if count != 21 || !reflect.DeepEqual(sets[get], []string{"Filter", "Identity"}) || !reflect.DeepEqual(sets[set], []string{"Identity"}) {
		t.Fatal("pinned meeting configuration surface changed")
	}
}

func TestUnitContextElementsKeepReceivedJSONAndNullListIsEmpty(t *testing.T) {
	t.Parallel()
	var value resource.Configuration
	if err := json.Unmarshal([]byte(withContexts(t, `[{"Id":"c7","PublishedName":"n","Published":false,"Extra":[1]},null]`)), &value); err != nil {
		t.Fatal(err)
	}
	items := value.PublishedEntraAuthenticationContexts
	if len(items) != 2 || items[1] != nil || items[0].ID == nil || *items[0].ID != "c7" || items[0].Published == nil || *items[0].Published || items[0].Identity != nil {
		t.Fatal("context elements decoded incorrectly")
	}
	out, err := json.Marshal(items[0])
	if err != nil || string(out) != `{"Id":"c7","PublishedName":"n","Published":false,"Extra":[1]}` {
		t.Fatal("received context element changed on re-send")
	}
	for _, list := range []string{"null", "[]"} {
		var empty resource.Configuration
		if err := json.Unmarshal([]byte(withContexts(t, list)), &empty); err != nil || empty.PublishedEntraAuthenticationContexts == nil || len(empty.PublishedEntraAuthenticationContexts) != 0 {
			t.Fatal("constructor-initialized empty list lost")
		}
	}
	for _, invalid := range []string{`{"PublishedEntraAuthenticationContexts":{}}`, `{"PublishedEntraAuthenticationContexts":[1]}`, `{"ClientAudioPort":-1}`, `{"ClientAudioPort":4294967296}`, `[]`, `null`} {
		var bad resource.Configuration
		if json.Unmarshal([]byte(invalid), &bad) == nil {
			t.Fatalf("invalid response %s accepted", invalid)
		}
	}
}

func TestUnitRedactedFormatting(t *testing.T) {
	t.Parallel()
	text := "Sensitive"
	for _, value := range []interface{ String() string }{resource.Configuration{CustomFooterText: &text}, resource.ScalarSettings{}, resource.Settings{}, resource.UpdateResult{}, resource.PublishedEntraAuthenticationContext{PublishedName: &text}, resource.AuthenticationContextEdit{}} {
		if strings.Contains(value.String(), text) || !strings.Contains(value.String(), "redacted") {
			t.Fatal("formatting exposed configuration values")
		}
	}
}
