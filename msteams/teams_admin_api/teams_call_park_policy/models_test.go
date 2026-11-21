package teamscallparkpolicy_test

import (
	"encoding/json"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
)

func TestUnitCallParkInt64ProjectionsAndRawPresence(t *testing.T) {
	raw := []byte(`{"Identity":"Global","Description":"","AllowCallPark":false,"ParkTimeoutSeconds":-9223372036854775808,"PickupRangeStart":9007199254740993,"PickupRangeEnd":9223372036854775807,"Future":false}`)
	original := append([]byte(nil), raw...)
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil || value.Description == nil || *value.Description != "" || value.AllowCallPark == nil || *value.AllowCallPark || value.ParkTimeoutSeconds == nil || *value.ParkTimeoutSeconds != -9223372036854775808 || value.PickupRangeStart == nil || *value.PickupRangeStart != 9007199254740993 || value.PickupRangeEnd == nil || *value.PickupRangeEnd != 9223372036854775807 {
		t.Fatal("signed Int64 or presence projection changed")
	}
	raw[0] = ' '
	*value.PickupRangeStart = 0
	got, e := json.Marshal(value)
	if e != nil || string(got) != string(original) {
		t.Fatal("raw snapshot borrowed storage, rounded numbers or changed with projections")
	}
	for _, raw := range []string{`{"PickupRangeStart":9223372036854775808}`, `{"PickupRangeStart":"future"}`, `{"PickupRangeStart":0.5}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil || value.PickupRangeStart != nil || len(value.RawFields["PickupRangeStart"]) == 0 {
			t.Fatal("future/overflow shape lost or projected as zero")
		}
	}
	var value2 resource.Policy
	if json.Unmarshal([]byte(`{"ParkTimeoutSeconds":0,"PickupRangeEnd":null}`), &value2) != nil || value2.ParkTimeoutSeconds == nil || *value2.ParkTimeoutSeconds != 0 || value2.PickupRangeEnd != nil || value2.PickupRangeStart != nil {
		t.Fatal("absent/null/zero collapsed")
	}
}
