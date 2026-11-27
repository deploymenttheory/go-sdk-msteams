package privacyconfiguration_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
)

func TestUnitPrivacyConfigurationRetainsAllNullableSettingsAndFutureFields(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"EnablePrivacyMode", "AutoInitiateContacts", "PublishLocationDataDefault", "DisplayPublishedPhotoDefault"} {
		for _, value := range []string{"true", "false", "null"} {
			raw := `{"Identity":"fixture-private","` + name + `":` + value + `,"Future":{"a":[],"b":null}}`
			var got resource.Configuration
			if json.Unmarshal([]byte(raw), &got) != nil {
				t.Fatal("valid privacy setting rejected")
			}
			field := reflect.ValueOf(got).FieldByName(name)
			if field.IsNil() != (value == "null") || value != "null" && field.Elem().Bool() != (value == "true") {
				t.Fatal("nullable Boolean projection changed")
			}
			encoded, err := json.Marshal(got)
			var before, after any
			if err != nil || json.Unmarshal([]byte(raw), &before) != nil || json.Unmarshal(encoded, &after) != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("privacy presence/future data changed")
			}
			if strings.Contains(fmt.Sprintf("%v %#v", got, got), "fixture-private") {
				t.Fatal("privacy formatting exposed tenant data")
			}
		}
	}
	var missing resource.Configuration
	if json.Unmarshal([]byte(`{"Identity":"Global"}`), &missing) != nil || len(missing.RawFields) != 1 {
		t.Fatal("missing Boolean fields were materialized")
	}
	for _, raw := range []string{`[]`, `"Global"`, `{"EnablePrivacyMode":"yes"}`, `{"AutoInitiateContacts":{}}`} {
		var got resource.Configuration
		if json.Unmarshal([]byte(raw), &got) == nil {
			t.Fatal("source-rejected scalar/object model accepted")
		}
	}
}

func TestUnitPrivacyBoundSettingsKeepFalseNullAndOmission(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"EnablePrivacyMode", "AutoInitiateContacts", "PublishLocationDataDefault", "DisplayPublishedPhotoDefault"} {
		for _, field := range []policy.Field[bool]{policy.Set(false), policy.Set(true), policy.Null[bool]()} {
			var settings resource.Settings
			reflect.ValueOf(&settings).Elem().FieldByName(name).Set(reflect.ValueOf(field))
			raw, err := json.Marshal(settings)
			var body map[string]json.RawMessage
			if err != nil || json.Unmarshal(raw, &body) != nil || len(body) != 1 {
				t.Fatal("unbound settings were serialized")
			}
			want := "null"
			if !field.IsNull() {
				b, _ := field.Value()
				if b {
					want = "true"
				} else {
					want = "false"
				}
			}
			if string(body[name]) != want {
				t.Fatal("bound nullable Boolean changed")
			}
		}
	}
}
