package recordingmessages_test

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

func sameJSON(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}

type modelFixture struct {
	Models []struct {
		Name        string
		Observation struct {
			Failure *string
			Value   json.RawMessage
		}
	} `json:"models"`
	Lists []struct {
		Name        string
		Observation struct {
			Failure *string
			Values  json.RawMessage
		}
	} `json:"lists"`
}

func models(t *testing.T) modelFixture {
	t.Helper()
	raw, e := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/recording-message-models.json")
	if e != nil {
		t.Fatal(e)
	}
	var v modelFixture
	if json.Unmarshal(raw, &v) != nil || len(v.Models) != 13 || len(v.Lists) != 9 {
		t.Fatal("invalid recording model fixture")
	}
	return v
}
func TestRecordingPinnedModelSetterAndConstructorRules(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Models {
		var value any
		var err error
		switch c.Name {
		case "language-null", "localization-null-language":
			var v resource.Localization
			err = json.Unmarshal([]byte(`{"Language":null}`), &v)
			value = &v
		case "language-empty":
			value = &resource.Localization{Language: new("")}
		case "language-space":
			value = &resource.Localization{Language: new(" ")}
		case "language-future":
			value = &resource.Localization{Language: new("Future-Language")}
		case "localization-absent":
			value = &resource.Localization{}
		case "message-absent", "message-null-id":
			value = &resource.Message{}
		case "message-empty-id", "message-invalid-id", "message-null-collection":
			body := `{"Id":""}`
			if c.Name == "message-invalid-id" {
				body = `{"Id":"fixture-invalid"}`
			}
			if c.Name == "message-null-collection" {
				body = `{"RecordingAndTranscriptionLocalizationCustomMessage":null}`
			}
			var v resource.Message
			err = json.Unmarshal([]byte(body), &v)
			value = &v
		case "message-id-final-lf":
			value = &resource.Message{ID: new(relative + "\n")}
		case "message-long-description":
			value = &resource.Message{Description: new(strings.Repeat("x", 2000))}
		default:
			t.Fatal("unmapped pinned model case")
		}
		if (err != nil) != (c.Observation.Failure != nil) {
			t.Fatal("pinned model acceptance changed: " + c.Name)
		}
		if err != nil {
			continue
		}
		raw, e := json.Marshal(value)
		if e != nil || !sameJSON(raw, c.Observation.Value) {
			t.Fatal("pinned constructor projection changed: " + c.Name)
		}
	}
}
func TestRecordingSnapshotsRetainPresenceAndRejectEarlierBadSetters(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{}`, `{"Identity":"fixture-private","Id":null,"Future":{"a":[],"b":null}}`, `{"Id":"` + relative + `\n","RecordingAndTranscriptionLocalizationCustomMessage":[{"Language":"","InitiatorImplicit":"","Future":false},null]}`} {
		var v resource.Message
		if json.Unmarshal([]byte(raw), &v) != nil {
			t.Fatal("valid snapshot rejected")
		}
		after, e := json.Marshal(v)
		if e != nil || !sameJSON([]byte(raw), after) {
			t.Fatal("recording snapshot presence changed")
		}
		if strings.Contains(fmt.Sprintf("%v %#v", v, v), "fixture-private") {
			t.Fatal("recording formatting exposed tenant data")
		}
	}
	for _, raw := range []string{`{"Id":"bad","Id":"` + relative + `"}`, `{"RecordingAndTranscriptionLocalizationCustomMessage":null,"RecordingAndTranscriptionLocalizationCustomMessage":[]}`, `{"RecordingAndTranscriptionLocalizationCustomMessage":[{"Language":null,"Language":"en-US"}]}`, `[]`} {
		var v resource.Message
		if json.Unmarshal([]byte(raw), &v) == nil {
			t.Fatal("earlier setter failure bypassed")
		}
	}
}

func seed() []*resource.Localization {
	makeItem := func(language, text, id string) *resource.Localization {
		return &resource.Localization{Identity: new(id), Language: new(language), InitiatorImplicit: new(text)}
	}
	return []*resource.Localization{makeItem("en-US", "a", "fixture-first"), makeItem("en-US", "a", "fixture-second"), makeItem("en-us", "a", "fixture-lower"), makeItem("en-US", "b", "fixture-different")}
}
func TestRecordingPinnedSevenFieldListEdits(t *testing.T) {
	t.Parallel()
	for _, c := range models(t).Lists {
		items := seed()
		duplicate, lower := items[1], items[2]
		var edit *resource.LocalizationEdit
		switch c.Name {
		case "null":
			edit = resource.ClearLocalizations()
		case "empty-operators":
			edit = resource.ChangeLocalizations(nil, nil)
		case "replace-empty":
			edit = resource.ReplaceLocalizations()
		case "replace-duplicates":
			edit = resource.ReplaceLocalizations(duplicate, duplicate, lower)
		case "replace-null":
			edit = resource.ReplaceLocalizations(nil)
		case "remove-one":
			edit = resource.ChangeLocalizations(nil, []*resource.Localization{duplicate})
		case "remove-two":
			edit = resource.ChangeLocalizations(nil, []*resource.Localization{duplicate, duplicate})
		case "add-existing":
			edit = resource.ChangeLocalizations([]*resource.Localization{duplicate, duplicate}, nil)
		case "remove-and-add":
			edit = resource.ChangeLocalizations([]*resource.Localization{duplicate}, []*resource.Localization{duplicate})
		default:
			t.Fatal("unmapped pinned list case")
		}
		v, e := edit.Apply(items)
		raw, encodeErr := json.Marshal(v)
		if e != nil || encodeErr != nil || c.Observation.Failure != nil || !sameJSON(raw, c.Observation.Values) {
			t.Fatal("pinned list edit changed: " + c.Name)
		}
		if len(items) != 4 || *items[0].Identity != "fixture-first" {
			t.Fatal("list edit mutated caller source")
		}
	}
}

func TestRecordingBuildersOwnInputsAndPreserveExactKeys(t *testing.T) {
	t.Parallel()
	text := "fixture-original"
	v, e := resource.NewLocalization(resource.LocalizationInput{Language: "Future-Language", InitiatorImplicit: &text})
	if e != nil {
		t.Fatal(e)
	}
	text = "changed"
	if *v.InitiatorImplicit != "fixture-original" {
		t.Fatal("builder aliased input")
	}
	edit := resource.ReplaceLocalizations(v)
	*v.InitiatorImplicit = "changed again"
	out, e := edit.Apply(nil)
	if e != nil || *out[0].InitiatorImplicit != "fixture-original" {
		t.Fatal("edit aliased input")
	}
	if _, e := resource.NewLocalization(resource.LocalizationInput{}); e == nil {
		t.Fatal("empty mandatory language accepted")
	}
	if _, e := resource.ChangeLocalizations(nil, []*resource.Localization{v}).Apply([]*resource.Localization{nil}); e == nil {
		t.Fatal("null comparator element silently skipped")
	}
	for _, edit := range []*resource.LocalizationEdit{nil, {}} {
		if _, e := edit.Apply(nil); e == nil {
			t.Fatal("unbound/zero edit accepted")
		}
	}
}

func TestRecordingRemovalComparesAllSevenOrdinalFields(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"Language", "InitiatorImplicit", "ParticipantImplicit", "InitiatorExplicit", "ParticipantExplicitRequested", "ParticipantExplicitProvided", "AgreementDialogue"} {
		remove := &resource.Localization{Language: new("en-US"), InitiatorImplicit: new("a")}
		reflect.ValueOf(remove).Elem().FieldByName(field).Set(reflect.ValueOf(new("different")))
		v, e := resource.ChangeLocalizations(nil, []*resource.Localization{remove}).Apply(seed())
		if e != nil || len(v) != 4 {
			t.Fatal("localization key field was ignored")
		}
	}
}
