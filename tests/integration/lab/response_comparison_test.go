package lab_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// Compare complete JSON values. RawMessage byte equality treats harmless nested
// object ordering and whitespace as configuration drift.
func sameResponseFields(left, right map[string]json.RawMessage) bool {
	if left == nil || right == nil {
		return false
	}
	decode := func(fields map[string]json.RawMessage) (any, error) {
		raw, err := json.Marshal(fields)
		if err != nil {
			return nil, err
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var value any
		err = decoder.Decode(&value)
		return value, err
	}
	a, err := decode(left)
	if err != nil {
		return false
	}
	b, err := decode(right)
	return err == nil && reflect.DeepEqual(a, b)
}

func TestUnitResponseComparisonUsesAcceptanceJSONValues(t *testing.T) {
	t.Parallel()
	for _, service := range []string{"acs_federation_configuration", "application_meeting_configuration"} {
		t.Run(service, func(t *testing.T) {
			directory, err := os.OpenRoot("../../../msteams/teams_admin_api/" + service + "/testdata")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := directory.Close(); err != nil {
					t.Error("response fixture directory close failed")
				}
			})
			raw, err := directory.ReadFile("get_global.json")
			if err != nil {
				t.Fatal(err)
			}
			var original map[string]json.RawMessage
			if json.Unmarshal(raw, &original) != nil || original == nil {
				t.Fatal("invalid acceptance fixture")
			}
			reformatted := make(map[string]json.RawMessage, len(original))
			for key, value := range original {
				var out bytes.Buffer
				if json.Indent(&out, value, "", "  ") != nil {
					t.Fatal("invalid response property")
				}
				reformatted[key] = append([]byte(" \n"), out.Bytes()...)
			}
			if !sameResponseFields(original, reformatted) {
				t.Fatal("JSON formatting changed comparison of identical response values")
			}
			for key := range original {
				removed := make(map[string]json.RawMessage, len(original))
				for k, v := range original {
					removed[k] = v
				}
				delete(removed, key)
				if sameResponseFields(original, removed) {
					t.Fatal("missing response property was ignored")
				}
				changed := make(map[string]json.RawMessage, len(original))
				for k, v := range original {
					changed[k] = v
				}
				if bytes.Equal(bytes.TrimSpace(original[key]), []byte("null")) {
					changed[key] = json.RawMessage(`"changed"`)
				} else {
					changed[key] = json.RawMessage("null")
				}
				if sameResponseFields(original, changed) {
					t.Fatal("changed response property was ignored")
				}
			}
			if sameResponseFields(original, map[string]json.RawMessage{"malformed": json.RawMessage("{")}) {
				t.Fatal("malformed JSON accepted")
			}
		})
	}
}
