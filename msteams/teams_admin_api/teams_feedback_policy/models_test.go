package teamsfeedbackpolicy_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

func TestUnitFeedbackPolicyModelsMatchTypedSetters(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Models []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}
	}{
		Models: []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}{

			{
				Name:  "constructor",
				Input: "constructor",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "empty-object",
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "all-settings",
				Input: "{\"UserInitiatedMode\":\"Future-UserInitiatedMode\",\"ReceiveSurveysMode\":\"Future-ReceiveSurveysMode\",\"AllowScreenshotCollection\":false,\"AllowEmailCollection\":false,\"AllowLogCollection\":false,\"EnableFeatureSuggestions\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":false,\"AllowLogCollection\":false,\"AllowScreenshotCollection\":false,\"EnableFeatureSuggestions\":false,\"Identity\":null,\"ReceiveSurveysMode\":\"Future-ReceiveSurveysMode\",\"UserInitiatedMode\":\"Future-UserInitiatedMode\"}"),
				},
			},

			{
				Name:  "UserInitiatedMode-null",
				Input: "{\"UserInitiatedMode\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "UserInitiatedMode-empty",
				Input: "{\"UserInitiatedMode\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":\"\"}"),
				},
			},

			{
				Name:  "UserInitiatedMode-future",
				Input: "{\"UserInitiatedMode\":\"Future-UserInitiatedMode\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":\"Future-UserInitiatedMode\"}"),
				},
			},

			{
				Name:  "ReceiveSurveysMode-null",
				Input: "{\"ReceiveSurveysMode\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "ReceiveSurveysMode-empty",
				Input: "{\"ReceiveSurveysMode\":\"\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":\"\",\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "ReceiveSurveysMode-future",
				Input: "{\"ReceiveSurveysMode\":\"Future-ReceiveSurveysMode\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":\"Future-ReceiveSurveysMode\",\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowScreenshotCollection-null",
				Input: "{\"AllowScreenshotCollection\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowScreenshotCollection-false",
				Input: "{\"AllowScreenshotCollection\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":false,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowScreenshotCollection-true",
				Input: "{\"AllowScreenshotCollection\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":true,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowEmailCollection-null",
				Input: "{\"AllowEmailCollection\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowEmailCollection-false",
				Input: "{\"AllowEmailCollection\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":false,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowEmailCollection-true",
				Input: "{\"AllowEmailCollection\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":true,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowLogCollection-null",
				Input: "{\"AllowLogCollection\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowLogCollection-false",
				Input: "{\"AllowLogCollection\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":false,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "AllowLogCollection-true",
				Input: "{\"AllowLogCollection\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":true,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "EnableFeatureSuggestions-null",
				Input: "{\"EnableFeatureSuggestions\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "EnableFeatureSuggestions-false",
				Input: "{\"EnableFeatureSuggestions\":false}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":false,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "EnableFeatureSuggestions-true",
				Input: "{\"EnableFeatureSuggestions\":true}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":true,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "long-user-mode",
				Input: "{\"UserInitiatedMode\":\"" + strings.Repeat("x", 2000) + "\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":\"" + strings.Repeat("x", 2000) + "\"}"),
				},
			},

			{
				Name:  "case-duplicate",
				Input: "{\"UserInitiatedMode\":\"first\",\"description\":\"second\"}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":\"first\"}"),
				},
			},

			{
				Name:  "duplicate-null",
				Input: "{\"AllowLogCollection\":true,\"allowlogcollection\":null}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Value: json.RawMessage("{\"AllowEmailCollection\":null,\"AllowLogCollection\":null,\"AllowScreenshotCollection\":null,\"EnableFeatureSuggestions\":null,\"Identity\":null,\"ReceiveSurveysMode\":null,\"UserInitiatedMode\":null}"),
				},
			},

			{
				Name:  "invalid-boolean-object",
				Input: "{\"AllowEmailCollection\":{}}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},

			{
				Name:  "invalid-string-array",
				Input: "{\"ReceiveSurveysMode\":[]}",
				Observation: struct {
					Failure *string
					Value   json.RawMessage
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},
		},
	}

	for _, c := range fixture.Models {
		t.Run(c.Name, func(t *testing.T) {
			var value resource.Policy
			var e error
			if c.Input != "constructor" {
				e = json.Unmarshal([]byte(c.Input), &value)
			}
			if c.Observation.Failure != nil {
				if !errors.Is(e, client.ErrDecode) {
					t.Fatal("source setter failure accepted")
				}
				return
			}
			if e != nil {
				t.Fatal("source-compatible model rejected")
			}
			type projection resource.Policy
			encoded, e := json.Marshal(projection(value))
			var got, want map[string]any
			if e != nil || json.Unmarshal(encoded, &got) != nil || json.Unmarshal(c.Observation.Value, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("source typed projection differs")
			}
		})
	}
}

func TestUnitFeedbackPolicySnapshotsOwnFutureData(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"Global","AllowLogCollection":false,"AllowEmailCollection":null,"UserInitiatedMode":"","Future":9007199254740993}`)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.AllowLogCollection == nil || *value.AllowLogCollection || value.AllowEmailCollection != nil || value.UserInitiatedMode == nil || *value.UserInitiatedMode != "" {
		t.Fatal("nullable false/empty projection lost")
	}
	clear(raw)
	if string(value.RawFields["Future"]) != "9007199254740993" {
		t.Fatal("future snapshot lost")
	}
	value.RawFields["Future"][0] = 'x'
	encoded, e := json.Marshal(value)
	if e != nil || !json.Valid(encoded) {
		t.Fatal("original snapshot borrowed projected map storage")
	}
}
