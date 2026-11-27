package teamsmeetingbroadcastconfiguration

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// readValue exposes a typed convenience view without changing the dynamic
// configuration value retained by the modern helper's BeforeJson pipeline.
func readValue[T any](raw json.RawMessage) *T {
	var value *T
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return value
}

// validReadKeys checks before map materialization: the generated parser rejects
// exact duplicate keys, and the later ConvertFrom-Json rejects case collisions.
// Both stages inspect nested objects, including unknown fields.
func validReadKeys(raw []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if !readKeys(decoder) {
		return false
	}
	_, err := decoder.Token()
	return err == io.EOF
}

func readKeys(decoder *json.Decoder) bool {
	token, err := decoder.Token()
	if err != nil {
		return false
	}
	switch token {
	case json.Delim('{'):
		var names []string
		for decoder.More() {
			key, err := decoder.Token()
			name, ok := key.(string)
			if err != nil || !ok {
				return false
			}
			for _, previous := range names {
				if strings.EqualFold(previous, name) {
					return false
				}
			}
			names = append(names, name)
			if !readKeys(decoder) {
				return false
			}
		}
		end, err := decoder.Token()
		return err == nil && end == json.Delim('}')
	case json.Delim('['):
		for decoder.More() {
			if !readKeys(decoder) {
				return false
			}
		}
		end, err := decoder.Token()
		return err == nil && end == json.Delim(']')
	default:
		return true
	}
}
