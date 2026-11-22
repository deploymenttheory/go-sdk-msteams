package labkit

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// Compare complete JSON values. RawMessage byte equality treats harmless nested
// object ordering and whitespace as configuration drift.
func SameResponseFields(left, right map[string]json.RawMessage) bool {
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

// Compare management properties across separate Get/List observations. Key is
// the server's XML backing object, not a property of the pinned BasePolicy DTO.
// Its schema/XName metadata can differ across calls. Every other property,
// including unknown extensions, ConfigId and ConfigMetadata, remains exact.
// Each response's complete Key still belongs to RawFields/RawJSON and is covered
// by the service's actual-response unit tests.
func SameConfigurationFields(left, right map[string]json.RawMessage) bool {
	if left == nil || right == nil {
		return false
	}
	properties := func(fields map[string]json.RawMessage) map[string]json.RawMessage {
		result := make(map[string]json.RawMessage, len(fields))
		for name, value := range fields {
			if name != "Key" {
				result[name] = value
			}
		}
		return result
	}
	return SameResponseFields(properties(left), properties(right))
}

// Return fixed diagnostics only; response values and unknown property names stay private.
func ResponseComparisonFailure(left, right map[string]json.RawMessage) string {
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
		if SameResponseFields(map[string]json.RawMessage{key: value}, map[string]json.RawMessage{key: other}) {
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
		if SameResponseFields(map[string]json.RawMessage{"value": a}, map[string]json.RawMessage{"value": b}) {
			continue
		}
		reasons = append(reasons, "read/list metadata: "+path+" differs")
	}
	return reasons
}

func LogMetadataMismatches(t *testing.T, left, right map[string]json.RawMessage) {
	t.Helper()
	for _, reason := range metadataMismatchReasons(left, right) {
		t.Log(reason)
	}
}

func SameProjection(left, right any) bool {
	a, e := json.Marshal(left)
	if e != nil {
		return false
	}
	b, e := json.Marshal(right)
	if e != nil {
		return false
	}
	var x, y any
	leftDecoder := json.NewDecoder(bytes.NewReader(a))
	rightDecoder := json.NewDecoder(bytes.NewReader(b))
	leftDecoder.UseNumber()
	rightDecoder.UseNumber()
	return leftDecoder.Decode(&x) == nil && rightDecoder.Decode(&y) == nil && reflect.DeepEqual(x, y)
}

func SamePolicyName(identity, fixture string) bool {
	if len(identity) > 4 && strings.EqualFold(identity[:4], "tag:") {
		identity = identity[4:]
	}
	if len(fixture) > 4 && strings.EqualFold(fixture[:4], "tag:") {
		fixture = fixture[4:]
	}
	return strings.EqualFold(identity, fixture)
}
