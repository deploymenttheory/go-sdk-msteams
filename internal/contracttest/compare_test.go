package contracttest

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUnitCompareJSON(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, expected, actual, path, kind string
	}{
		{name: "object order", expected: `{"a":1,"b":2}`, actual: `{"b":2,"a":1}`},
		{name: "omitted null", expected: `{"a":null}`, actual: `{}`, path: "body.object[0]", kind: "missing-field"},
		{name: "unknown field", expected: `{}`, actual: `{"untrusted-private-key":"never print this"}`, path: "body.object[0]", kind: "unexpected-field"},
		{name: "array order", expected: `["a","b"]`, actual: `["b","a"]`, path: "body[0]", kind: "different-value"},
		{name: "array length", expected: `[1]`, actual: `[]`, path: "body", kind: "different-length"},
		{name: "empty null", expected: `[]`, actual: `null`, path: "body", kind: "different-type"},
		{name: "false true", expected: `false`, actual: `true`, path: "body", kind: "different-value"},
		{name: "integer precision", expected: `9007199254740992`, actual: `9007199254740993`, path: "body", kind: "different-value"},
		{name: "number spelling", expected: `1`, actual: `1.0`, path: "body", kind: "different-value"},
		{name: "duplicate key", expected: `{"a":1}`, actual: `{"a":0,"a":1}`, path: "body", kind: "invalid-actual-json"},
		{name: "trailing value", expected: `{}`, actual: `{} {}`, path: "body", kind: "invalid-actual-json"},
		{name: "malformed fixture", expected: `{`, actual: `{}`, path: "body", kind: "invalid-expected-json"},
		{name: "scalar object", expected: `"fixture"`, actual: `{}`, path: "body", kind: "different-value"},
		{name: "nested malformed", expected: `{}`, actual: `{"a":[}`, path: "body", kind: "invalid-actual-json"},
		{name: "depth bound", expected: `[]`, actual: strings.Repeat("[", 130) + strings.Repeat("]", 130), path: "body", kind: "invalid-actual-json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			differences := compareJSON([]byte(tc.expected), []byte(tc.actual), "body")
			if tc.kind == "" {
				if len(differences) != 0 {
					t.Fatalf("equivalent JSON differed: %+v", differences)
				}
				return
			}
			if len(differences) == 0 || differences[0] != (Difference{Path: tc.path, Kind: tc.kind}) {
				t.Fatalf("missing expected structural difference: %+v", differences)
			}
			raw, err := json.Marshal(differences)
			if err != nil || strings.Contains(string(raw), "untrusted-private-key") || strings.Contains(string(raw), "never print this") {
				t.Fatal("structural diff leaked fixture content")
			}
		})
	}
}
