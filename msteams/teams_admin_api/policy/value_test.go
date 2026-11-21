package policy_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestFieldPreservesOmittedNullFalseAndEmpty(t *testing.T) {
	t.Parallel()
	input := struct {
		Omitted policy.Field[bool]   `json:"omitted,omitzero"`
		Null    policy.Field[bool]   `json:"null,omitzero"`
		False   policy.Field[bool]   `json:"false,omitzero"`
		Empty   policy.Field[string] `json:"empty,omitzero"`
	}{Null: policy.Null[bool](), False: policy.Set(false), Empty: policy.Set("")}
	raw, err := json.Marshal(input)
	if err != nil || string(raw) != `{"null":null,"false":false,"empty":""}` {
		t.Fatal("request field presence changed")
	}
	if input.Omitted.IsSet() || input.Null.IsZero() || !input.Null.IsNull() || !input.False.IsSet() {
		t.Fatal("field states changed")
	}
	if value, ok := input.False.Value(); !ok || value {
		t.Fatal("bound false lost")
	}
	if _, ok := input.Null.Value(); ok {
		t.Fatal("null became a value")
	}
	if _, err := json.Marshal(policy.Field[bool]{}); !errors.Is(err, policy.ErrUnsetField) {
		t.Fatal("direct unset field serialized")
	}
	var read struct {
		Known   policy.Field[bool]   `json:"known,omitzero"`
		Missing policy.Field[bool]   `json:"missing,omitzero"`
		Null    policy.Field[string] `json:"null,omitzero"`
	}
	if json.Unmarshal([]byte(`{"known":false,"null":null}`), &read) != nil || !read.Known.IsSet() || read.Missing.IsSet() || !read.Null.IsNull() {
		t.Fatal("field JSON binding changed")
	}
}

func TestCompiledListEditsMatchObservedHostBehavior(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-provider-dispatch.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Lists []struct {
			Name        string `json:"name"`
			Observation struct {
				Values []string `json:"values"`
			} `json:"observation"`
		} `json:"lists"`
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Lists) != 7 {
		t.Fatal("invalid pinned modifier observations")
	}
	for _, c := range fixture.Lists {
		var edit *policy.ListEdit
		switch c.Name {
		case "null":
			edit = policy.ReplaceStrings()
		case "empty-operators":
			edit = policy.ChangeStrings(nil, nil)
		case "replacement":
			edit = policy.ReplaceStrings("b", "b", "")
		case "remove-one":
			edit = policy.ChangeStrings(nil, []string{"a"})
		case "remove-missing":
			edit = policy.ChangeStrings(nil, []string{"missing"})
		case "add-existing":
			edit = policy.ChangeStrings([]string{"a", "c", "c"}, nil)
		case "remove-and-add":
			edit = policy.ChangeStrings([]string{"b", "a"}, []string{"a", "b"})
		default:
			t.Fatal("unknown modifier observation")
		}
		current := []string{"a", "a", "A", "b"}
		value, err := edit.Apply(current)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(value)
		var actual []string
		if err != nil || json.Unmarshal(encoded, &actual) != nil || !reflect.DeepEqual(actual, c.Observation.Values) || !reflect.DeepEqual(current, []string{"a", "a", "A", "b"}) {
			t.Fatal("compiled modifier differs or mutates current state")
		}
	}
}

func TestPolicyListsOwnInputsAndInvalidEditStaysInvalid(t *testing.T) {
	t.Parallel()
	values := []string{"fixture"}
	list := policy.Strings(values...)
	edit := policy.ReplaceStrings(values...)
	add, remove := []string{"added"}, []string{"fixture"}
	delta := policy.ChangeStrings(add, remove)
	values[0] = "changed"
	add[0] = "changed"
	remove[0] = "changed"
	for _, value := range []*policy.StringList{list, mustApply(t, edit, nil), mustApply(t, delta, []string{"fixture"})} {
		raw, err := json.Marshal(value)
		if err != nil || string(raw) != `["fixture"]` && string(raw) != `["added"]` {
			t.Fatal("policy value borrowed input")
		}
	}
	if _, err := (&policy.ListEdit{}).Apply(nil); !errors.Is(err, policy.ErrInput) {
		t.Fatal("zero edit accepted")
	}
	var missing *policy.ListEdit
	if _, err := missing.Apply(nil); !errors.Is(err, policy.ErrInput) {
		t.Fatal("nil edit applied")
	}
}
func mustApply(t *testing.T, edit *policy.ListEdit, current []string) *policy.StringList {
	t.Helper()
	v, e := edit.Apply(current)
	if e != nil {
		t.Fatal(e)
	}
	return v
}
