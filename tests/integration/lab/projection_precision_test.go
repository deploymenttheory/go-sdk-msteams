package lab_test

import (
	"encoding/json"
	"testing"
)

func TestUnitGetterComparisonPreservesAllUInt64Bits(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, left, right string
		equal             bool
	}{
		{"above-float-integer-precision", "9007199254740993", "9007199254740992", false},
		{"uint64-upper-bound", "18446744073709551615", "18446744073709551614", false},
		{"equal-upper-bound", "18446744073709551615", "18446744073709551615", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var left uint64
			if json.Unmarshal([]byte(tc.left), &left) != nil {
				t.Fatal("invalid source UInt64 vector")
			}
			if sameTemplateProjection(map[string]any{"LowSeatLimit": left}, map[string]json.RawMessage{"LowSeatLimit": json.RawMessage(tc.right)}) != tc.equal {
				t.Fatal("getter comparison rounded an unsigned value")
			}
		})
	}
}
