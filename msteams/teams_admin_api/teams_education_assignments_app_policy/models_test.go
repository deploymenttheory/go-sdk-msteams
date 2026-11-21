package teamseducationassignmentsapppolicy_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_assignments_app_policy"
)

func projection(v resource.Policy) map[string]*string {
	return map[string]*string{"Identity": v.Identity, "ParentDigestEnabledType": v.ParentDigestEnabledType, "MakeCodeEnabledType": v.MakeCodeEnabledType, "TurnItInEnabledType": v.TurnItInEnabledType, "TurnItInApiUrl": v.TurnItInAPIURL, "TurnItInApiKey": v.TurnItInAPIKey}
}

func TestUnitEducationAssignmentsModelsMatchTypedSetters(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Models []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   map[string]*string
			}
		}
	}{
		Models: []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   map[string]*string
			}
		}{

			{
				Name:  "constructor",
				Input: "constructor",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "empty-object",
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "all-settings",
				Input: "{\"ParentDigestEnabledType\":\"FutureParent\",\"MakeCodeEnabledType\":\"FutureCode\",\"TurnItInEnabledType\":\"FutureTurnItIn\",\"TurnItInApiUrl\":\"HTTPS://fixture.invalid\",\"TurnItInApiKey\":\"offline-fixture-key\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     new("FutureCode"),
						"ParentDigestEnabledType": new("FutureParent"),
						"TurnItInApiKey":          new("offline-fixture-key"),
						"TurnItInApiUrl":          new("HTTPS://fixture.invalid"),
						"TurnItInEnabledType":     new("FutureTurnItIn"),
					},
				},
			},

			{
				Name:  "ParentDigestEnabledType-null",
				Input: "{\"ParentDigestEnabledType\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "ParentDigestEnabledType-empty",
				Input: "{\"ParentDigestEnabledType\":\"\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": new(""),
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "MakeCodeEnabledType-null",
				Input: "{\"MakeCodeEnabledType\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "MakeCodeEnabledType-empty",
				Input: "{\"MakeCodeEnabledType\":\"\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     new(""),
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "TurnItInEnabledType-null",
				Input: "{\"TurnItInEnabledType\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "TurnItInEnabledType-empty",
				Input: "{\"TurnItInEnabledType\":\"\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     new(""),
					},
				},
			},

			{
				Name:  "TurnItInApiUrl-null",
				Input: "{\"TurnItInApiUrl\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "TurnItInApiUrl-empty",
				Input: "{\"TurnItInApiUrl\":\"\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Name:  "TurnItInApiKey-null",
				Input: "{\"TurnItInApiKey\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "TurnItInApiKey-empty",
				Input: "{\"TurnItInApiKey\":\"\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          new(""),
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "url-http",
				Input: "{\"TurnItInApiUrl\":\"http://fixture.invalid\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Name:  "url-upper",
				Input: "{\"TurnItInApiUrl\":\"HTTPS://fixture.invalid\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          new("HTTPS://fixture.invalid"),
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "url-opaque",
				Input: "{\"TurnItInApiUrl\":\"https://opaque\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          new("https://opaque"),
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "url-space",
				Input: "{\"TurnItInApiUrl\":\"https:// \"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          new("https:// "),
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "url-prefix-space",
				Input: "{\"TurnItInApiUrl\":\" https://fixture.invalid\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Name:  "url-final-lf",
				Input: "{\"TurnItInApiUrl\":\"https://fixture.invalid\\n\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          new("https://fixture.invalid\n"),
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "url-middle-lf",
				Input: "{\"TurnItInApiUrl\":\"https://fixture\\n.invalid\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Name:  "url-empty-host",
				Input: "{\"TurnItInApiUrl\":\"https://\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Name:  "long-api-key",
				Input: "{\"TurnItInApiKey\":\"" + strings.Repeat("x", 2000) + "\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": nil,
						"TurnItInApiKey":          new("" + strings.Repeat("x", 2000)),
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "case-duplicate",
				Input: "{\"ParentDigestEnabledType\":\"first\",\"parentdigestenabledtype\":\"second\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"Identity":                nil,
						"MakeCodeEnabledType":     nil,
						"ParentDigestEnabledType": new("second"),
						"TurnItInApiKey":          nil,
						"TurnItInApiUrl":          nil,
						"TurnItInEnabledType":     nil,
					},
				},
			},

			{
				Name:  "invalid-then-valid-url",
				Input: "{\"TurnItInApiUrl\":\"bad\",\"turnitinapiurl\":\"https://fixture.invalid\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Failure: new("Newtonsoft.Json.JsonSerializationException"),
				},
			},

			{
				Name:  "invalid-string-array",
				Input: "{\"TurnItInApiKey\":[]}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},
		},
	}

	for _, c := range fixture.Models {
		t.Run(c.Name, func(t *testing.T) {
			var v resource.Policy
			var err error
			if c.Input != "constructor" {
				err = json.Unmarshal([]byte(c.Input), &v)
			}
			if c.Observation.Failure != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source setter failure accepted")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(projection(v), c.Observation.Value) {
				t.Fatal("known model projection differs")
			}
		})
	}
}

func TestUnitEducationAssignmentsSnapshotOwnershipAndRedaction(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"TurnItInApiKey":"fixture-private","Future":9007199254740993,"TurnItInApiUrl":"https:// "}`)
	var v resource.Policy
	if json.Unmarshal(raw, &v) != nil {
		t.Fatal("source-compatible opaque URL rejected")
	}
	clear(raw)
	if !strings.Contains(string(v.RawJSON), "9007199254740993") || string(v.RawFields["Future"]) != "9007199254740993" {
		t.Fatal("owned future snapshot lost")
	}
	encoded, e := json.Marshal(v)
	if e != nil || !strings.Contains(string(encoded), "fixture-private") {
		t.Fatal("caller-owned original snapshot lost")
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if strings.Contains(fmt.Sprintf(format, v), "fixture-private") {
			t.Fatal("ordinary policy formatting exposed key")
		}
	}
	for _, bad := range []string{`{"TurnItInApiUrl":"bad","turnitinapiurl":"https://fixture.invalid"}`, `{"TurnItInApiUrl":"https://x\n\n"}`, `{"TurnItInApiUrl":"httpſ://fixture.invalid"}`} {
		if !errors.Is(json.Unmarshal([]byte(bad), &v), client.ErrDecode) {
			t.Fatal("source-invalid URL accepted")
		}
	}
}
