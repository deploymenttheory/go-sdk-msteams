package tenantidentity

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// groupNode follows this response's generated root/parser path. It deliberately
// does not change the unrelated tenant or support DTO decoders in this package.
func groupNode(raw []byte) (map[string]json.RawMessage, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[0] != '{' && raw[0] != '[' {
		return nil, client.ErrDecode
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if groupValue(decoder) != nil {
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

func groupValue(decoder *json.Decoder) error {
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
			if err := groupValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return client.ErrDecode
		}
	case '[':
		for decoder.More() {
			if err := groupValue(decoder); err != nil {
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

func groupString(raw json.RawMessage) *string {
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

// groupGUID follows the audited Guid.Parse forms used by GroupModel.ParseFrom.
// This is response projection, not a new GUID restriction on directory inputs.
func groupGUID(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if len(value) == 32 {
		if _, err := hex.DecodeString(value); err != nil {
			return "", false
		}
		return strings.ToLower(value[:8] + "-" + value[8:12] + "-" + value[12:16] + "-" + value[16:20] + "-" + value[20:]), true
	}
	if len(value) == 38 && ((value[0] == '{' && value[9] == '-' && value[37] == '}') || (value[0] == '(' && value[37] == ')')) {
		return groupGUIDDashed(value[1:37])
	}
	if strings.HasPrefix(value, "{") {
		return groupGUIDHex(value)
	}
	return groupGUIDDashed(value)
}

func groupGUIDDashed(value string) (string, bool) {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return "", false
	}
	parts := []string{value[:8], value[9:13], value[14:18], value[19:23], value[24:28], value[28:]}
	values := make([]uint64, len(parts))
	for index, part := range parts {
		var ok bool
		if index == 5 {
			// The final eight digits use the runtime's strict hex-number parser,
			// unlike the first five compatibility-parsed components.
			var err error
			if strings.ContainsAny(value, "xX+") {
				// The runtime's compatibility fallback permits trailing NULs in
				// its final hex-number parse; the fast ordinary D path does not.
				part = strings.TrimRight(part, "\x00")
			}
			values[index], err = strconv.ParseUint(part, 16, 32)
			ok = err == nil
		} else {
			values[index], ok = groupGUIDComponent(part)
		}
		if !ok {
			return "", false
		}
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x", values[0], values[1], values[2], values[3], values[4], values[5]), true
}

func groupGUIDComponent(value string) (uint64, bool) {
	value = strings.TrimPrefix(value, "+")
	if len(value) >= 2 && strings.EqualFold(value[:2], "0x") {
		value = value[2:]
	}
	value = strings.TrimLeft(value, "0")
	if value == "" {
		return 0, true
	}
	parsed, err := strconv.ParseUint(value, 16, 32)
	return parsed, err == nil
}

func groupGUIDHex(value string) (string, bool) {
	value = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, value)
	if len(value) < 2 || value[0] != '{' || value[len(value)-1] != '}' {
		return "", false
	}
	parts := strings.SplitN(value[1:len(value)-1], ",", 4)
	if len(parts) != 4 || len(parts[3]) < 2 || parts[3][0] != '{' || parts[3][len(parts[3])-1] != '}' {
		return "", false
	}
	parts = append(parts[:3], strings.Split(parts[3][1:len(parts[3])-1], ",")...)
	if len(parts) != 11 {
		return "", false
	}
	values := make([]uint64, len(parts))
	for index, part := range parts {
		if len(part) <= 2 || !strings.EqualFold(part[:2], "0x") {
			return "", false
		}
		var ok bool
		values[index], ok = groupGUIDComponent(part[2:])
		if !ok || index >= 3 && values[index] > 255 {
			return "", false
		}
	}
	// The actual runtime parses these two components as UInt32 and then
	// truncates to UInt16; masking preserves that observed source behavior.
	return fmt.Sprintf("%08x-%04x-%04x-%02x%02x-%02x%02x%02x%02x%02x%02x", values[0], values[1]&0xffff, values[2]&0xffff,
		values[3], values[4], values[5], values[6], values[7], values[8], values[9], values[10]), true
}
