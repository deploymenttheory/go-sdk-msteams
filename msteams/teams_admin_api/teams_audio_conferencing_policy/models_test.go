package teamsaudioconferencingpolicy_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
)

func TestUnitAudioPolicyPinnedListNormalization(t *testing.T) {
	raw, e := os.ReadFile("../../../" + modelSource)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Cases []struct {
			Name     string
			Rejected bool
			Wire     json.RawMessage
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 15 {
		t.Fatal("incomplete normalization fixture")
	}
	legacy := json.RawMessage(`{"AppId":"future-phone","DataSource":"fixture-source","Key":"fixture-key","Enabled":false}`)
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var value *resource.ListChange
			var e error
			switch c.Name {
			case "phone-list-omitted":
			case "phone-list-scalar":
				value = resource.ReplacePhoneNumbers("future-phone")
			case "phone-list-replacement":
				value = resource.ReplacePhoneNumbers("+440000000001", "not-a-phone")
			case "phone-list-empty-replacement", "phone-list-null-replacement":
				value = resource.ReplacePhoneNumbers()
			case "phone-list-operators":
				value = resource.ChangePhoneNumbers([]string{"+440000000001"}, []string{"not-a-phone"})
			case "phone-list-empty-operators":
				value = resource.ChangePhoneNumbers(nil, nil)
			case "phone-list-legacy-object-replacement":
				value, e = resource.ReplacePhoneValues([]json.RawMessage{legacy})
			case "phone-list-legacy-object-operators":
				value, e = resource.ChangePhoneValues([]json.RawMessage{legacy}, nil)
			case "phone-list-null-item-rejected":
				value, e = resource.ReplacePhoneValues([]json.RawMessage{json.RawMessage(`"future-phone"`), json.RawMessage(`null`)})
			case "phone-list-numeric-item-rejected":
				value, e = resource.ReplacePhoneValues([]json.RawMessage{json.RawMessage(`123`)})
			case "phone-list-empty-string":
				value = resource.ReplacePhoneNumbers("")
			case "phone-list-whitespace-string":
				value = resource.ReplacePhoneNumbers(" ")
			case "phone-list-replace-operator-rejected", "phone-list-ps-list-modifier-rejected":
				// These PowerShell-only input containers have no native constructor.
				// A malformed native change likewise cannot produce a request payload.
				value = new(resource.ListChange)
			default:
				t.Fatal("unknown source input")
			}
			body := map[string]any{"Identity": "tag:Fixture"}
			if e == nil && value != nil {
				var encoded []byte
				encoded, e = json.Marshal(value)
				if e == nil {
					body["MeetingInvitePhoneNumbers"] = json.RawMessage(encoded)
				}
			}
			if c.Rejected {
				if !errors.Is(e, resource.ErrInput) {
					t.Fatal("source-rejected input accepted")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			actual, e := json.Marshal(body)
			if e != nil {
				t.Fatal(e)
			}
			var got, want any
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal(c.Wire, &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("normalization differs: got %s want %s", actual, c.Wire)
			}
		})
	}
}

func TestUnitAudioPolicyRawFutureShapesAndOwnedChanges(t *testing.T) {
	raw := []byte(`{"Identity":"Global","AllowTollFreeDialin":"FutureMode","MeetingInvitePhoneNumbers":["first",false],"Future":9007199254740993}`)
	original := append([]byte(nil), raw...)
	var v resource.Policy
	if json.Unmarshal(raw, &v) != nil || v.AllowTollFreeDialin != nil || v.MeetingInvitePhoneNumbers != nil {
		t.Fatal("unprojectable known field leaked a partial projection")
	}
	raw[0] = ' '
	actual, e := json.Marshal(v)
	if e != nil || string(actual) != string(original) {
		t.Fatal("raw snapshot lost future shape, number or ownership")
	}
	input := []string{"first", "first", ""}
	change := resource.ReplacePhoneNumbers(input...)
	input[0] = "changed"
	actual, e = json.Marshal(change)
	if e != nil || string(actual) != `["first","first",""]` {
		t.Fatal("replacement borrowed input")
	}
	legacy := json.RawMessage(`{"Number":"future-phone","DataSource":"source","Future":false}`)
	change, e = resource.ReplacePhoneValues([]json.RawMessage{legacy})
	if e != nil {
		t.Fatal(e)
	}
	legacy[0] = ' '
	actual, e = json.Marshal(change)
	if e != nil {
		t.Fatal(e)
	}
	var got []map[string]any
	if json.Unmarshal(actual, &got) != nil || len(got) != 1 || got[0]["Future"] != false || got[0]["Number"] != "future-phone" {
		t.Fatal("legacy value borrowed or lost future data")
	}
	if _, exists := got[0]["DataSource"]; exists {
		t.Fatal("legacy metadata was retained")
	}
}
