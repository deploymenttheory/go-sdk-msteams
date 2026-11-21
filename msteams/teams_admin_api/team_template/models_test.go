package teamtemplate_test

import (
	"encoding/json"
	"testing"

	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

func TestUnitTemplatePreservesResponseAndProjectsFlattenedDefinition(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"displayName":"Root","@odata.id":"/opaque/reference","teamDefinition":{"displayName":"Definition","channels":[],"futureNested":false},"extendedProperties":{"shortDescription":"Summary","futureValues":[]},"futureNull":null}`)
	var value template.Template
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	if value.DisplayName == nil || *value.DisplayName != "Root" || value.ShortDescription == nil || *value.ShortDescription != "Summary" || value.ODataID == nil || *value.ODataID != "/opaque/reference" {
		t.Fatal("read view lost flattened known fields or opaque identity")
	}
	if value.Channels == nil || len(value.Channels) != 0 || value.MemberSettings == nil {
		t.Fatal("read projection lost empty array or Body model defaults")
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, raw, encoded)
	// The snapshot owns its raw JSON; clearing the caller buffer cannot alter it.
	clear(raw)
	again, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, encoded, again)
	definition, err := json.Marshal(value.Definition)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(definition, &fields) != nil {
		t.Fatal("projected definition is not JSON")
	}
	if _, exists := fields["futureNull"]; exists {
		t.Fatal("response snapshot became an unknown-field write passthrough")
	}
}

func TestUnitTemplateSummaryPreservesOmissionUnknownValuesAndZero(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, raw string
	}{
		{name: "unknown and zero", raw: `{"id":"fixture","name":"","channelCount":0,"appCount":0,"future":{"flag":false},"futureNull":null}`},
		{name: "future known type", raw: `{"id":"fixture","channelCount":"future-type"}`},
		{name: "omitted fields", raw: `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var value template.TemplateSummary
			if err := json.Unmarshal([]byte(tc.raw), &value); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, []byte(tc.raw), encoded)
			if tc.name == "unknown and zero" && (value.ChannelCount == nil || *value.ChannelCount != 0 || value.Name == nil || *value.Name != "") {
				t.Fatal("known zero or empty values were omitted")
			}
		})
	}
}

func TestUnitErrorEnvelopeRetainsFutureFieldsAndEmptyErrors(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`{"operationId":"fixture-operation","errors":[{"code":"FixtureCode","message":"Fixture message","action":"Fixture action","futureDetail":false}],"futureEnvelope":[]}`,
		`{"errors":[]}`,
		`{"errors":null}`,
	} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			var value template.ErrorEnvelope
			if err := json.Unmarshal([]byte(raw), &value); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, []byte(raw), encoded)
		})
	}
}

func TestUnitReadModelsRejectNonObjects(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		new  func() any
	}{
		{name: "detail", new: func() any { return &template.Template{} }},
		{name: "summary", new: func() any { return &template.TemplateSummary{} }},
		{name: "error envelope", new: func() any { return &template.ErrorEnvelope{} }},
		{name: "error detail", new: func() any { return &template.ErrorDetail{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, raw := range []string{"null", "[]", `"text"`} {
				if json.Unmarshal([]byte(raw), tc.new()) == nil {
					t.Fatal("non-object decoded as a successful read model")
				}
			}
		})
	}
}
