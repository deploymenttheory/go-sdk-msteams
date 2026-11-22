package labkit

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestUnitMetadataDiagnosticsUseFixedPathsAndRetainPrivateValues(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile(Root + "/msteams/teams_admin_api/teams_room_video_tele_conferencing_policy/testdata/get_global.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, path := range metadataDiagnosticPaths {
		var left, right map[string]json.RawMessage
		if json.Unmarshal(raw, &left) != nil || json.Unmarshal(raw, &right) != nil {
			t.Fatal("invalid acceptance fixture")
		}
		parts := strings.Split(path, ".")
		var alter func(map[string]json.RawMessage, []string)
		alter = func(fields map[string]json.RawMessage, parts []string) {
			if len(parts) == 1 {
				fields[parts[0]] = json.RawMessage(`"private-canary"`)
				return
			}
			var nested map[string]json.RawMessage
			if json.Unmarshal(fields[parts[0]], &nested) != nil || nested == nil {
				t.Fatal("missing fixed metadata path")
			}
			alter(nested, parts[1:])
			encoded, e := json.Marshal(nested)
			if e != nil {
				t.Fatal(e)
			}
			fields[parts[0]] = encoded
		}
		alter(right, parts)
		reasons := metadataMismatchReasons(left, right)
		if len(reasons) != 1 || reasons[0] != "read/list metadata: "+path+" differs" || strings.Contains(reasons[0], "private-canary") {
			t.Fatal("metadata diagnostic leaked or lost a changed fixed path")
		}
	}
}

func TestUnitResponseComparisonUsesAcceptanceJSONValues(t *testing.T) {
	t.Parallel()
	for _, service := range []string{"acs_federation_configuration", "application_meeting_configuration", "meeting_branding", "online_voicemail_policy", "online_voicemail_validation_configuration"} {
		t.Run(service, func(t *testing.T) {
			directory, err := os.OpenRoot(Root + "/msteams/teams_admin_api/" + service + "/testdata")
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
			if !SameResponseFields(original, reformatted) {
				t.Fatal("JSON formatting changed comparison of identical response values")
			}
			for key := range original {
				removed := make(map[string]json.RawMessage, len(original))
				for k, v := range original {
					removed[k] = v
				}
				delete(removed, key)
				if SameResponseFields(original, removed) {
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
				if SameResponseFields(original, changed) {
					t.Fatal("changed response property was ignored")
				}
			}
			if SameResponseFields(original, map[string]json.RawMessage{"malformed": json.RawMessage("{")}) {
				t.Fatal("malformed JSON accepted")
			}
		})
	}
}

func TestUnitResponseMismatchReasonsKeepValuesPrivate(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile(Root + "/msteams/teams_admin_api/online_voicemail_policy/testdata/get_global.json")
	if e != nil {
		t.Fatal(e)
	}
	var original map[string]json.RawMessage
	if json.Unmarshal(raw, &original) != nil {
		t.Fatal("invalid acceptance response")
	}
	for _, tc := range []struct{ name, field, value, want string }{
		{"same", "", "", ""},
		{"metadata", "Key", `"private-metadata-value"`, "read/list comparison: metadata values differ"},
		{"identity", "Identity", `"private-identity-value"`, "read/list comparison: identity values differ"},
		{"setting", "EnableTranscription", `null`, "read/list comparison: setting or extension values differ"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			next := make(map[string]json.RawMessage, len(original))
			for key, value := range original {
				next[key] = value
			}
			if tc.field != "" {
				next[tc.field] = json.RawMessage(tc.value)
			}
			if got := ResponseComparisonFailure(original, next); got != tc.want {
				t.Fatal("fixed mismatch classification differs")
			}
		})
	}
	if ResponseComparisonFailure(nil, original) != "read/list comparison: response fields missing" {
		t.Fatal("missing fields not classified")
	}
	if ResponseComparisonFailure(original, map[string]json.RawMessage{}) != "read/list comparison: property sets differ" {
		t.Fatal("missing property not classified")
	}
}

func TestUnitConfigurationComparisonKeepsStateAndExtensionsSeparateFromXmlKey(t *testing.T) {
	t.Parallel()
	directory, e := os.OpenRoot(Root + "/msteams/teams_admin_api")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if directory.Close() != nil {
			t.Error("acceptance fixture root close failed")
		}
	})
	for _, service := range []string{"teams_shifts_policy", "teams_room_video_tele_conferencing_policy", "online_voicemail_policy", "meeting_branding"} {
		raw, e := directory.ReadFile(service + "/testdata/get_global.json")
		if e != nil {
			t.Fatal(e)
		}
		var left, right map[string]json.RawMessage
		if json.Unmarshal(raw, &left) != nil || json.Unmarshal(raw, &right) != nil || left["Key"] == nil {
			t.Fatal("actual acceptance fixture missing its XML backing metadata")
		}
		right["Key"] = json.RawMessage(`{"SchemaId":{"XName":{"name":"alternate-server-serialization"}}}`)
		if SameResponseFields(left, right) || !SameConfigurationFields(left, right) {
			t.Fatal("raw equality or management-property comparison changed incorrectly")
		}
		for key := range left {
			if key == "Key" {
				continue
			}
			altered := make(map[string]json.RawMessage, len(right))
			for k, v := range right {
				altered[k] = v
			}
			delete(altered, key)
			if SameConfigurationFields(left, altered) {
				t.Fatal("missing management/extension property ignored")
			}
			altered[key] = json.RawMessage(`{"changed":true}`)
			if SameConfigurationFields(left, altered) {
				t.Fatal("changed management/extension property ignored")
			}
		}
		right["future-extension"] = json.RawMessage(`{"value":9007199254740993}`)
		if SameConfigurationFields(left, right) {
			t.Fatal("unknown extension was ignored")
		}
		if SameConfigurationFields(nil, right) || SameConfigurationFields(left, nil) {
			t.Fatal("missing snapshot accepted")
		}
		numericLeft := make(map[string]json.RawMessage, len(left))
		numericRight := make(map[string]json.RawMessage, len(left))
		for key, value := range left {
			numericLeft[key] = value
			numericRight[key] = value
		}
		numericLeft["future-number"] = json.RawMessage(`9007199254740993`)
		numericRight["future-number"] = json.RawMessage(`9007199254740992`)
		if SameConfigurationFields(numericLeft, numericRight) {
			t.Fatal("management comparison rounded unknown numeric extension")
		}
		var original map[string]json.RawMessage
		if json.Unmarshal(raw, &original) != nil || !SameResponseFields(left, original) {
			t.Fatal("comparison mutated the complete original snapshot")
		}
	}
}

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
			if SameProjection(map[string]any{"LowSeatLimit": left}, map[string]json.RawMessage{"LowSeatLimit": json.RawMessage(tc.right)}) != tc.equal {
				t.Fatal("getter comparison rounded an unsigned value")
			}
		})
	}
}
