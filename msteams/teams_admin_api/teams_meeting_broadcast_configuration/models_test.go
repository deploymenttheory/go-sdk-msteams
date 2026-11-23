package teamsmeetingbroadcastconfiguration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
)

func TestUnitEveryTypedSettingUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	// Default reads carry three typed settings; the exposed SDN read carries all ten.
	for file, want := range map[string]int{"get_global.json": 3, "update_readback.json": 3, "restored_readback.json": 3, "get_global_sdn.json": 10} {
		t.Run(file, func(t *testing.T) {
			raw := readResponseFixture(t, file)
			var value resource.Configuration
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
					if !fields.Field(i).IsNil() {
						t.Fatalf("absent field %s decoded a value", name)
					}
					continue
				}
				pointer := reflect.New(field.Type)
				if json.Unmarshal(actual, pointer.Interface()) != nil || !reflect.DeepEqual(fields.Field(i).Interface(), pointer.Elem().Interface()) {
					t.Fatalf("typed field %s differs from acceptance response", name)
				}
				checked++
			}
			if checked != want {
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

func TestUnitProviderTokenIsNeverFormatted(t *testing.T) {
	t.Parallel()
	const marker = "sdk-unit-provider-value"
	provider := marker
	request := resource.UpdateRequest{SdnAPIToken: &provider}
	value := resource.Configuration{SdnAPIToken: &provider}
	for _, text := range []string{fmt.Sprint(request), fmt.Sprintf("%+v", request), fmt.Sprintf("%#v", request), fmt.Sprint(value), fmt.Sprintf("%+v", value), fmt.Sprintf("%#v", value), fmt.Sprintf("%v", &request), fmt.Sprintf("%v", &value)} {
		if strings.Contains(text, marker) {
			t.Fatal("provider token formatted")
		}
	}
}

func jsonEqual(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
