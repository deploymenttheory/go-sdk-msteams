package teamsfirstpartymeetingtemplateconfiguration_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
)

func TestUnitEveryTypedGetterUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"get_global.json", "update_readback.json", "restored_readback.json"} {
		t.Run(file, func(t *testing.T) {
			raw := readResponseFixture(t, file)
			var value resource.Configuration
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			var expected struct {
				Identity    *string `json:"Identity"`
				Description *string `json:"Description"`
				Templates   []struct {
					Identity, Description, Name, Category *string
					Options                               []struct {
						Identity, Value, Name *string
						IsLocked, IsHidden    *bool
					} `json:"TeamsMeetingOptions"`
				} `json:"TeamsMeetingTemplates"`
			}
			if json.Unmarshal(raw, &expected) != nil || len(expected.Templates) == 0 {
				t.Fatal("tenant response lacks a populated first-party template list")
			}
			if !reflect.DeepEqual(value.Identity, expected.Identity) || !reflect.DeepEqual(value.Description, expected.Description) || len(value.TeamsMeetingTemplates) != len(expected.Templates) {
				t.Fatal("typed root getters differ from acceptance response")
			}
			for i, want := range expected.Templates {
				got := value.TeamsMeetingTemplates[i]
				if got == nil || !reflect.DeepEqual(got.Identity, want.Identity) || !reflect.DeepEqual(got.Description, want.Description) || !reflect.DeepEqual(got.Name, want.Name) || !reflect.DeepEqual(got.Category, want.Category) || len(got.TeamsMeetingOptions) != len(want.Options) || got.RawJSON == nil {
					t.Fatal("typed template getters differ from acceptance response")
				}
				for j, option := range want.Options {
					o := got.TeamsMeetingOptions[j]
					if o == nil || !reflect.DeepEqual(o.Identity, option.Identity) || !reflect.DeepEqual(o.Value, option.Value) || !reflect.DeepEqual(o.Name, option.Name) || !reflect.DeepEqual(o.IsLocked, option.IsLocked) || !reflect.DeepEqual(o.IsHidden, option.IsHidden) {
						t.Fatal("typed option getters differ from acceptance response")
					}
				}
			}
			if value.RawFields["ConfigMetadata"] == nil {
				t.Fatal("unknown response fields were not retained")
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

func TestUnitModelsRedactFormatting(t *testing.T) {
	t.Parallel()
	for _, v := range []any{resource.Configuration{}, resource.Template{}, resource.Option{}, resource.UpdateResult{}, resource.TemplateEdit{}} {
		if s, ok := v.(interface{ GoString() string }); !ok || s.GoString() == "" {
			t.Fatal("model formatting is not redacted")
		}
	}
}

func jsonEqual(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
