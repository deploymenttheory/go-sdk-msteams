package teamsaipolicy_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ai_policy"
)

func TestUnitEveryTypedSettingUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"get_global.json", "update_readback.json", "create_readback.json"} {
		t.Run(file, func(t *testing.T) {
			raw := readResponseFixture(t, file)
			var value resource.Policy
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			fields := reflect.ValueOf(value)
			typ := fields.Type()
			checked := 0
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				name := field.Tag.Get("json")
				if name == "-" {
					continue
				}
				actual, ok := value.RawFields[name]
				if !ok {
					t.Fatalf("tenant response missing typed field %s", name)
				}
				pointer := reflect.New(field.Type)
				if json.Unmarshal(actual, pointer.Interface()) != nil || !reflect.DeepEqual(fields.Field(i).Interface(), pointer.Elem().Interface()) {
					t.Fatalf("typed field %s differs from acceptance response", name)
				}
				checked++
			}
			if checked != 6 {
				t.Fatal("typed setting count changed")
			}
			snapshot := append([]byte(nil), raw...)
			clear(raw)
			marshaled, err := json.Marshal(value)
			if err != nil || !jsonEqual(snapshot, marshaled) || !bytes.Equal(bytes.TrimSpace(snapshot), value.RawJSON) {
				t.Fatal("model lost owned response snapshot")
			}
		})
	}
}
func jsonEqual(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
