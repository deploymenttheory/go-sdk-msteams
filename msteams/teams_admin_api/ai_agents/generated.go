package aiagents

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// generatedNode follows the generated root parser: empty text and array roots
// produce no response object; scalar roots fail. Exact duplicate keys anywhere
// in the document fail before materialization, including ignored properties.
func generatedNode(raw []byte) (map[string]json.RawMessage, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[0] != '{' && raw[0] != '[' {
		return nil, client.ErrDecode
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if validateGeneratedValue(decoder) != nil {
		return nil, client.ErrDecode
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, client.ErrDecode
	}
	if raw[0] == '[' {
		return nil, nil
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return nil, client.ErrDecode
	}
	return fields, nil
}

func validateGeneratedValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return client.ErrDecode
			}
			seen[name] = true
			if err := validateGeneratedValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return client.ErrDecode
		}
	case '[':
		for decoder.More() {
			if err := validateGeneratedValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return client.ErrDecode
		}
	default:
		return client.ErrDecode
	}
	return nil
}

func generatedString(raw json.RawMessage) *string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '"' {
		return nil
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return &value
}
func generatedBool(raw json.RawMessage) *bool {
	raw = bytes.TrimSpace(raw)
	if !bytes.Equal(raw, []byte("true")) && !bytes.Equal(raw, []byte("false")) {
		return nil
	}
	var value bool
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return &value
}
func isGeneratedNumber(raw []byte) bool {
	return len(raw) > 0 && (raw[0] == '-' || raw[0] >= '0' && raw[0] <= '9')
}
func generatedInt64(raw json.RawMessage) (*int64, error) {
	raw = bytes.TrimSpace(raw)
	if !isGeneratedNumber(raw) {
		return nil, nil
	}
	value, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil {
		return nil, client.ErrDecode
	}
	return &value, nil
}
func generatedInt32(raw json.RawMessage) (*int32, error) {
	raw = bytes.TrimSpace(raw)
	if !isGeneratedNumber(raw) {
		return nil, nil
	}
	value, err := strconv.ParseInt(string(raw), 10, 32)
	if err != nil {
		return nil, client.ErrDecode
	}
	result := int32(value)
	return &result, nil
}
func generatedDouble(raw json.RawMessage) (*float64, error) {
	raw = bytes.TrimSpace(raw)
	if !isGeneratedNumber(raw) {
		return nil, nil
	}
	value, err := strconv.ParseFloat(string(raw), 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return nil, client.ErrDecode
	}
	return &value, nil
}
func generatedArray(raw json.RawMessage) ([]json.RawMessage, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '[' {
		return nil, nil
	}
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil {
		return nil, client.ErrDecode
	}
	return values, nil
}
func generatedStrings(raw json.RawMessage) ([]*string, error) {
	values, err := generatedArray(raw)
	if err != nil || values == nil {
		return nil, err
	}
	result := make([]*string, len(values))
	for index, value := range values {
		result[index] = generatedString(value)
	}
	return result, nil
}

func generatedObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	fields, err := generatedNode(raw)
	if err != nil || fields == nil {
		return nil, client.ErrDecode
	}
	return fields, nil
}

// knowledgeStrings is intentionally separate: the legacy JsonConvert string
// reader coerces numeric lexemes and booleans, and preserves null entries.
func knowledgeStrings(raw []byte) ([]*string, error) {
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil || items == nil {
		return nil, client.ErrDecode
	}
	result := make([]*string, len(items))
	for index, item := range items {
		item = bytes.TrimSpace(item)
		if bytes.Equal(item, []byte("null")) {
			continue
		}
		if len(item) > 0 && item[0] == '"' {
			if json.Unmarshal(item, &result[index]) != nil {
				return nil, client.ErrDecode
			}
			continue
		}
		if bytes.Equal(item, []byte("true")) || bytes.Equal(item, []byte("false")) {
			text := string(item)
			result[index] = &text
			continue
		}
		var number json.Number
		if json.Unmarshal(item, &number) != nil {
			return nil, client.ErrDecode
		}
		text := number.String()
		result[index] = &text
	}
	return result, nil
}

func firstGeneratedByte(raw []byte) byte {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return 0
	}
	return raw[0]
}
