// Package testutil provides in-memory HTTP assertions for SDK unit tests.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"strconv"
)

// Difference identifies a structural mismatch without retaining either value.
// Object indexes follow sorted keys; no service-provided key is printed.
type Difference struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

func compareJSON(expected, actual []byte, path string) []Difference {
	want, ok := parseJSON(expected)
	if !ok {
		return []Difference{{Path: path, Kind: "invalid-expected-json"}}
	}
	got, ok := parseJSON(actual)
	if !ok {
		return []Difference{{Path: path, Kind: "invalid-actual-json"}}
	}
	return compareValue(want, got, path)
}

func parseJSON(raw []byte) (any, bool) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	value, ok := readJSON(decoder, 0)
	if !ok {
		return nil, false
	}
	_, err := decoder.Token()
	return value, err == io.EOF
}

func readJSON(decoder *json.Decoder, depth int) (any, bool) {
	if depth > 128 {
		return nil, false
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, false
	}
	switch token {
	case json.Delim('{'):
		object := map[string]any{}
		for decoder.More() {
			key, err := decoder.Token()
			name, ok := key.(string)
			if err != nil || !ok {
				return nil, false
			}
			if _, exists := object[name]; exists {
				return nil, false
			}
			value, ok := readJSON(decoder, depth+1)
			if !ok {
				return nil, false
			}
			object[name] = value
		}
		end, err := decoder.Token()
		return object, err == nil && end == json.Delim('}')
	case json.Delim('['):
		array := []any{}
		for decoder.More() {
			value, ok := readJSON(decoder, depth+1)
			if !ok {
				return nil, false
			}
			array = append(array, value)
		}
		end, err := decoder.Token()
		return array, err == nil && end == json.Delim(']')
	default:
		_, delimiter := token.(json.Delim)
		return token, !delimiter
	}
}

func compareValue(want, got any, path string) []Difference {
	differences := []Difference{}
	switch expected := want.(type) {
	case map[string]any:
		actual, ok := got.(map[string]any)
		if !ok {
			return []Difference{{Path: path, Kind: "different-type"}}
		}
		keys := make([]string, 0, len(expected)+len(actual))
		for key := range expected {
			keys = append(keys, key)
		}
		for key := range actual {
			if _, exists := expected[key]; !exists {
				keys = append(keys, key)
			}
		}
		sort.Strings(keys)
		for index, key := range keys {
			field := path + ".object[" + strconv.Itoa(index) + "]"
			value, exists := actual[key]
			if !exists {
				differences = append(differences, Difference{Path: field, Kind: "missing-field"})
				continue
			}
			wanted, exists := expected[key]
			if !exists {
				differences = append(differences, Difference{Path: field, Kind: "unexpected-field"})
				continue
			}
			differences = append(differences, compareValue(wanted, value, field)...)
		}
	case []any:
		actual, ok := got.([]any)
		if !ok {
			return []Difference{{Path: path, Kind: "different-type"}}
		}
		if len(expected) != len(actual) {
			differences = append(differences, Difference{Path: path, Kind: "different-length"})
		}
		for index := range min(len(expected), len(actual)) {
			differences = append(differences, compareValue(expected[index], actual[index], path+"["+strconv.Itoa(index)+"]")...)
		}
	default:
		// Only nil, bool, string and json.Number reach this branch. Number
		// spelling stays exact; no float conversion can round large integers.
		if want != got {
			differences = append(differences, Difference{Path: path, Kind: "different-value"})
		}
	}
	return differences
}
