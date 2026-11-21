package teamtemplate_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

type modelObservation struct {
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
	Wire  json.RawMessage `json:"wire"`
}

func modelObservations(t *testing.T) []modelObservation {
	t.Helper()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/core-model-serialization.json")
	if err != nil {
		t.Fatal(err)
	}
	var artifact struct {
		Version string             `json:"baseline_version"`
		Cases   []modelObservation `json:"cases"`
	}
	if json.Unmarshal(raw, &artifact) != nil || artifact.Version != "8.1.0" {
		t.Fatal("pinned model observations unavailable")
	}
	return artifact.Cases
}

func equalJSON(t *testing.T, expected, actual []byte) {
	t.Helper()
	var want, got any
	if json.Unmarshal(expected, &want) != nil || json.Unmarshal(actual, &got) != nil || !reflect.DeepEqual(want, got) {
		t.Fatal("JSON differs from pinned observation; no values printed")
	}
}

func TestUnitNewDefinition(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, display, short string
		invalid              bool
	}{
		{name: "missing display name", short: "Summary", invalid: true},
		{name: "missing short description", display: "Fixture", invalid: true},
		{name: "sparse expanded", display: "Fixture", short: "Summary"},
		{name: "mandatory string whitespace", display: " ", short: " "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			value, err := template.NewDefinition(tc.display, tc.short)
			if tc.invalid {
				if !errors.Is(err, template.ErrInput) {
					t.Fatal("missing expanded argument accepted")
				}
				return
			}
			if err != nil || value.DisplayName == nil || *value.DisplayName != tc.display || value.ShortDescription == nil || *value.ShortDescription != tc.short || value.MemberSettings != nil {
				t.Fatal("expanded construction changed input or initialized omitted settings")
			}
		})
	}
}

func TestUnitDefinitionMatchesPinnedModelObservations(t *testing.T) {
	t.Parallel()
	selected := 0
	for _, fixture := range modelObservations(t) {
		if !strings.HasPrefix(fixture.Name, "template-json-") && !strings.HasPrefix(fixture.Name, "template-expanded-") {
			continue
		}
		selected++
		t.Run(fixture.Name, func(t *testing.T) {
			t.Parallel()
			var value *template.Definition
			if strings.HasPrefix(fixture.Name, "template-json-") {
				var input string
				if json.Unmarshal(fixture.Input, &input) != nil {
					t.Fatal("model fixture input is not JSON text")
				}
				var err error
				value, err = template.ParseDefinition([]byte(input))
				if err != nil {
					t.Fatal("Go Body parser rejected captured model input")
				}
			} else {
				definition, err := template.NewDefinition("Fixture", "Fixture summary")
				if err != nil {
					t.Fatal(err)
				}
				value = &definition
				if fixture.Name == "template-expanded-empty-arrays-false" {
					no := false
					value.Apps, value.Categories, value.Channels = []*template.App{}, []string{}, []*template.Channel{}
					value.IsMembershipLimitedToOwners = &no
				}
			}
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, fixture.Wire, raw)
		})
	}
	if selected < 12 {
		t.Fatal("expected template observations disappeared")
	}
}

func TestUnitParseDefinitionRejectsNonObjectsAndResetsReceiver(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"", "null", "[]", "true", `"text"`, "{"} {
		t.Run("invalid "+raw, func(t *testing.T) {
			t.Parallel()
			if _, err := template.ParseDefinition([]byte(raw)); !errors.Is(err, template.ErrInput) {
				t.Fatal("invalid Body root accepted")
			}
		})
	}
	var value template.Definition
	if json.Unmarshal([]byte(`{"displayName":"First","apps":[{"id":"fixture"}]}`), &value) != nil || json.Unmarshal([]byte(`{}`), &value) != nil {
		t.Fatal("valid Body input rejected")
	}
	if value.DisplayName != nil || value.Apps != nil || value.MemberSettings == nil {
		t.Fatal("Body decoding retained prior values or lost default settings")
	}
}
