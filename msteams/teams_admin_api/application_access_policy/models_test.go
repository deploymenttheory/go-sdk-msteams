package applicationaccesspolicy_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
)

func equalJSON(t *testing.T, actual []byte, expected string) {
	t.Helper()
	var left, right any
	if err := json.Unmarshal(actual, &left); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(expected), &right); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(left, right) {
		t.Errorf("JSON differs: got %s, want %s", actual, expected)
	}
}

func TestUnitPolicyRetainsUnknownAndUnprojectedAppIDFields(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, wire string }{
		{name: "known and unknown", wire: `{"Identity":"tag:Fixture","AppIds":["fixture"],"Description":"","FutureFlag":false}`},
		{name: "legacy object array", wire: `{"Identity":"Fixture","AppIds":[{"AppId":"fixture","Enabled":false}],"Future":[]}`},
		{name: "explicit null", wire: `{"Identity":"Fixture","Description":null,"AppIds":null}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var result policy.Policy
			if err := json.Unmarshal([]byte(test.wire), &result); err != nil {
				t.Fatal(err)
			}
			actual, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, actual, test.wire)
			if result.Identity == "" || result.RawFields == nil {
				t.Fatal("known identity/raw fields lost")
			}
		})
	}
}

func TestUnitListChangesPreserveReplacementDeltaAndEmpty(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		change   *policy.ListChange
		expected string
	}{
		{name: "replacement", change: policy.ReplaceAppIDs([]string{"a", "b"}), expected: `["a","b"]`},
		{name: "empty replacement", change: policy.ReplaceAppIDs(nil), expected: `[]`},
		{name: "delta", change: policy.ChangeAppIDs([]string{"a"}, []string{"b"}), expected: `{"Add":["a"],"Remove":["b"]}`},
		{name: "empty operators", change: policy.ChangeAppIDs(nil, []string{}), expected: `{}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			actual, err := json.Marshal(test.change)
			if err != nil {
				t.Fatal(err)
			}
			equalJSON(t, actual, test.expected)
		})
	}
	input := []string{"original"}
	change := policy.ReplaceAppIDs(input)
	input[0] = "changed"
	actual, err := json.Marshal(change)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, actual, `["original"]`)
	if _, err := json.Marshal(policy.ListChange{}); !errors.Is(err, policy.ErrInput) {
		t.Fatal("invalid zero change accepted")
	}
}

func TestUnitLegacyListChangesCleanOnlyMetadata(t *testing.T) {
	t.Parallel()
	input := []json.RawMessage{json.RawMessage(`{"AppId":"fixture","Enabled":false,"DataSource":"discard","Key":"discard"}`)}
	replacement, err := policy.LegacyAppIDs(input)
	if err != nil {
		t.Fatal(err)
	}
	delta, err := policy.ChangeLegacyAppIDs(input, nil)
	if err != nil {
		t.Fatal(err)
	}
	input[0][0] = 'x'
	actual, err := json.Marshal(replacement)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, actual, `[{"AppId":"fixture","Enabled":false}]`)
	actual, err = json.Marshal(delta)
	if err != nil {
		t.Fatal(err)
	}
	equalJSON(t, actual, `{"Add":[{"AppId":"fixture","Enabled":false}]}`)
	if _, err := policy.LegacyAppIDs([]json.RawMessage{json.RawMessage(`"not-object"`)}); !errors.Is(err, policy.ErrInput) {
		t.Fatal("non-object legacy item accepted")
	}
}
