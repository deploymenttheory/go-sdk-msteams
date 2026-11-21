package onlinevoicemailvalidationconfiguration_test

import (
	"encoding/json"
	"reflect"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestValidationConfigurationRetainsWirePresenceAndFutureFields(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"Identity":"Global","Future":false}`, `{"Identity":"Global","AudioFileValidationEnabled":false,"AudioFileValidationUri":null}`, `{"Identity":"Global","audiofilevalidationuri":"HTTPS://example.invalid/validate","Future":{"a":[]}}`} {
		var value resource.Configuration
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("valid configuration rejected")
		}
		encoded, err := json.Marshal(value)
		var before, after any
		if err != nil || json.Unmarshal([]byte(raw), &before) != nil || json.Unmarshal(encoded, &after) != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("configuration presence/future data changed")
		}
	}
}

func TestValidationURIReadAndRequestRules(t *testing.T) {
	t.Parallel()
	for _, uri := range []string{"", "http://example.invalid/validate", "relative", "https://", "https://example.invalid/val\nidate"} {
		if (resource.Settings{AudioFileValidationURI: policy.Set(uri)}).Validate() == nil {
			t.Fatal("invalid URI request accepted")
		}
		encoded, err := json.Marshal(map[string]any{"audiofilevalidationuri": uri})
		var value resource.Configuration
		if err != nil || json.Unmarshal(encoded, &value) == nil {
			t.Fatal("invalid URI read accepted")
		}
	}
	for _, uri := range []string{"https://example.invalid/validate", "HTTPS:// ", "https://example.invalid/validate\n"} {
		if (resource.Settings{AudioFileValidationURI: policy.Set(uri)}).Validate() != nil {
			t.Fatal("source-compatible URI rejected")
		}
	}
	if (resource.Settings{AudioFileValidationURI: policy.Null[string](), AudioFileValidationEnabled: policy.Set(false)}).Validate() != nil {
		t.Fatal("nullable URI/false request rejected")
	}
	var value resource.Configuration
	if json.Unmarshal([]byte(`{"AudioFileValidationUri":"http://example.invalid/a","AudioFileValidationUri":"https://example.invalid/b"}`), &value) == nil {
		t.Fatal("duplicate key bypassed setter validation")
	}
}
