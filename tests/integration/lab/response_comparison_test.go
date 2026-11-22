package lab_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
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

// Return fixed diagnostics only; response values and unknown property names stay private.
func responseComparisonFailure(left, right map[string]json.RawMessage) string {
	if left == nil || right == nil {
		return "read/list comparison: response fields missing"
	}
	if len(left) != len(right) {
		return "read/list comparison: property sets differ"
	}
	metadata, identity, settings := false, false, false
	for key, value := range left {
		other, ok := right[key]
		if !ok {
			return "read/list comparison: property sets differ"
		}
		if sameResponseFields(map[string]json.RawMessage{key: value}, map[string]json.RawMessage{key: other}) {
			continue
		}
		switch key {
		case "Key", "ConfigMetadata", "ConfigId":
			metadata = true
		case "Identity":
			identity = true
		default:
			settings = true
		}
	}
	if identity {
		return "read/list comparison: identity values differ"
	}
	if settings {
		return "read/list comparison: setting or extension values differ"
	}
	if metadata {
		return "read/list comparison: metadata values differ"
	}
	return ""
}

// Metadata diagnostics enumerate only known public schema paths; no response
// values or server-supplied property names are emitted.
var metadataDiagnosticPaths = []string{
	"Key.ScopeClass", "Key.SchemaId", "Key.AuthorityId",
	"Key.DefaultXml.SchemaId", "Key.DefaultXml.Data", "Key.DefaultXml.ConfigObject",
	"Key.DefaultXml.Signature", "Key.DefaultXml.IsModified", "Key.XmlRoot",
	"ConfigMetadata", "ConfigId",
}

func metadataMismatchReasons(left, right map[string]json.RawMessage) []string {
	lookup := func(fields map[string]json.RawMessage, path string) json.RawMessage {
		raw, _ := json.Marshal(fields)
		for _, segment := range strings.Split(path, ".") {
			var value map[string]json.RawMessage
			if json.Unmarshal(raw, &value) != nil {
				return nil
			}
			raw = value[segment]
		}
		return raw
	}
	var reasons []string
	for _, path := range metadataDiagnosticPaths {
		a, b := lookup(left, path), lookup(right, path)
		if a == nil && b == nil {
			continue
		}
		if sameResponseFields(map[string]json.RawMessage{"value": a}, map[string]json.RawMessage{"value": b}) {
			continue
		}
		reasons = append(reasons, "read/list metadata: "+path+" differs")
	}
	return reasons
}
func logMetadataMismatches(t *testing.T, left, right map[string]json.RawMessage) {
	t.Helper()
	for _, reason := range metadataMismatchReasons(left, right) {
		t.Log(reason)
	}
}
func TestUnitMetadataDiagnosticsUseFixedPathsAndRetainPrivateValues(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../msteams/teams_admin_api/teams_room_video_tele_conferencing_policy/testdata/get_global.json")
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

func TestUnitResponseMismatchReasonsKeepValuesPrivate(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../msteams/teams_admin_api/online_voicemail_policy/testdata/get_global.json")
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
			if got := responseComparisonFailure(original, next); got != tc.want {
				t.Fatal("fixed mismatch classification differs")
			}
		})
	}
	if responseComparisonFailure(nil, original) != "read/list comparison: response fields missing" {
		t.Fatal("missing fields not classified")
	}
	if responseComparisonFailure(original, map[string]json.RawMessage{}) != "read/list comparison: property sets differ" {
		t.Fatal("missing property not classified")
	}
}
