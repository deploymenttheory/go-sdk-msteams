package policy_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitNullableListsPreserveFirstMatchAndOwnPointedValues(t *testing.T) {
	t.Parallel()
	entry, upper, empty := "entry", "ENTRY", ""
	values := []*string{nil, &entry, &empty}
	list := policy.NullableStrings(values...)
	replace := policy.ReplaceNullableStrings(values...)
	delta := policy.ChangeNullableStrings([]*string{nil, &upper}, []*string{nil, &entry})
	entry = "changed"
	upper = "changed"
	empty = "changed"
	values[0] = &entry
	for _, value := range []*policy.StringList{list, mustApplyNullable(t, replace, nil)} {
		if raw, err := json.Marshal(value); err != nil || string(raw) != `[null,"entry",""]` {
			t.Fatal("nullable constructor borrowed input pointers or lost null/empty distinction")
		}
	}
	current := []*string{new("entry"), nil, new("entry"), nil}
	result := mustApplyNullable(t, delta, current)
	if current[0] == nil || *current[0] != "entry" || current[1] != nil || current[2] == nil || *current[2] != "entry" || current[3] != nil {
		t.Fatal("delta mutated the prepared snapshot")
	}
	*current[2] = "changed"
	current[3] = new("changed")
	if raw, err := json.Marshal(result); err != nil || string(raw) != `["entry",null,null,"ENTRY"]` {
		t.Fatal("delta lost first null/ordinal removal, duplicates, addition order or owned strings")
	}
	for _, edit := range []*policy.ListEdit{policy.ReplaceNullableStrings(), policy.ChangeNullableStrings(nil, nil)} {
		if raw, err := json.Marshal(mustApplyNullable(t, edit, current)); err != nil || string(raw) != `[]` {
			t.Fatal("empty nullable edit did not clear")
		}
	}
	for _, edit := range []*policy.ListEdit{nil, {}} {
		if _, err := edit.ApplyNullable(nil); !errors.Is(err, policy.ErrInput) {
			t.Fatal("omitted or zero nullable edit became an implicit clear")
		}
	}
}

func mustApplyNullable(t *testing.T, edit *policy.ListEdit, current []*string) *policy.StringList {
	t.Helper()
	value, err := edit.ApplyNullable(current)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
