package onlinevoicemailvalidationconfiguration_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestPinnedValidationSetterObservations(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/voicemail-validation-models.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			URI      string          `json:"uri"`
			Null     bool            `json:"is_null"`
			Rejected bool            `json:"rejected"`
			Wire     json.RawMessage `json:"wire"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 11 {
		t.Fatal("invalid pinned setter capture")
	}
	accepted := 0
	for _, f := range fixture.Cases {
		field := policy.Set(f.URI)
		if f.Null {
			field = policy.Null[string]()
		}
		input := resource.Settings{AudioFileValidationEnabled: policy.Set(false), AudioFileValidationURI: field}
		if (input.Validate() != nil) != f.Rejected {
			t.Fatal("request validation differs from pinned model setter")
		}
		if f.Rejected {
			continue
		}
		accepted++
		var value resource.Configuration
		if json.Unmarshal(f.Wire, &value) != nil || value.AudioFileValidationEnabled == nil || *value.AudioFileValidationEnabled || (value.AudioFileValidationURI == nil) != f.Null {
			t.Fatal("pinned nullable/false projection changed")
		}
		encoded, err := json.Marshal(value)
		var before, after any
		if err != nil || json.Unmarshal(f.Wire, &before) != nil || json.Unmarshal(encoded, &after) != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("pinned setter snapshot changed")
		}
	}
	if accepted != 6 {
		t.Fatal("pinned setter acceptance count changed")
	}
}
