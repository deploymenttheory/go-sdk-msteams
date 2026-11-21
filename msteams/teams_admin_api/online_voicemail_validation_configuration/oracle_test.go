package onlinevoicemailvalidationconfiguration_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitTypedValidationSetterObservations(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			URI      string          `json:"uri"`
			Null     bool            `json:"is_null"`
			Rejected bool            `json:"rejected"`
			Wire     json.RawMessage `json:"wire"`
		} `json:"cases"`
	}{
		Cases: []struct {
			URI      string          `json:"uri"`
			Null     bool            `json:"is_null"`
			Rejected bool            `json:"rejected"`
			Wire     json.RawMessage `json:"wire"`
		}{

			{
				URI:      "",
				Null:     true,
				Rejected: false,
				Wire:     json.RawMessage("{\"AudioFileValidationEnabled\":false,\"AudioFileValidationUri\":null,\"Identity\":\"Global\"}"),
			},

			{
				URI:      "https://example.invalid/validate",
				Null:     false,
				Rejected: false,
				Wire:     json.RawMessage("{\"AudioFileValidationEnabled\":false,\"AudioFileValidationUri\":\"https://example.invalid/validate\",\"Identity\":\"Global\"}"),
			},

			{
				URI:      "HTTPS://example.invalid/validate",
				Null:     false,
				Rejected: false,
				Wire:     json.RawMessage("{\"AudioFileValidationEnabled\":false,\"AudioFileValidationUri\":\"HTTPS://example.invalid/validate\",\"Identity\":\"Global\"}"),
			},

			{
				URI:      "https:// ",
				Null:     false,
				Rejected: false,
				Wire:     json.RawMessage("{\"AudioFileValidationEnabled\":false,\"AudioFileValidationUri\":\"https:// \",\"Identity\":\"Global\"}"),
			},

			{
				URI:      "https://example.invalid/" + strings.Repeat("a", 2048),
				Null:     false,
				Rejected: false,
				Wire:     json.RawMessage("{\"AudioFileValidationEnabled\":false,\"AudioFileValidationUri\":\"https://example.invalid/" + strings.Repeat("a", 2048) + "\",\"Identity\":\"Global\"}"),
			},

			{
				URI:      "https://example.invalid/validate\n",
				Null:     false,
				Rejected: false,
				Wire:     json.RawMessage("{\"AudioFileValidationEnabled\":false,\"AudioFileValidationUri\":\"https://example.invalid/validate\\n\",\"Identity\":\"Global\"}"),
			},

			{
				URI:      "http://example.invalid/validate",
				Null:     false,
				Rejected: true,
			},

			{
				URI:      "validate",
				Null:     false,
				Rejected: true,
			},

			{
				URI:      "",
				Null:     false,
				Rejected: true,
			},

			{
				URI:      "https://",
				Null:     false,
				Rejected: true,
			},

			{
				URI:      "https://example.invalid/val\nidate",
				Null:     false,
				Rejected: true,
			},
		},
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
