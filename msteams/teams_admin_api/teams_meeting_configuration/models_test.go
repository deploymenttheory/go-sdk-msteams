package teamsmeetingconfiguration_test

import (
	"bytes"
	"encoding/json"

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
