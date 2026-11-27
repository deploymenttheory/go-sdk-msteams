package directoryuser

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// aadString opts only the five verified AadUser string properties into the
// pinned JsonConvert string-reader behavior. Unknown fields remain raw JSON.
type aadString struct{ value *string }

func (s *aadString) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		s.value = nil
		return nil
	}
	var value string
	if len(raw) > 0 && raw[0] == '"' {
		if json.Unmarshal(raw, &value) != nil {
			return client.ErrDecode
		}
	} else if bytes.Equal(raw, []byte("true")) || bytes.Equal(raw, []byte("false")) {
		value = string(raw)
	} else {
		var number json.Number
		if json.Unmarshal(raw, &number) != nil {
			return client.ErrDecode
		}
		value = string(number)
	}
	s.value = &value
	return nil
}

// asciiProperty matches these source ASCII property names without Go's
// Unicode SimpleFold aliases (for example long-s). Values are read in wire
// order: repeated exact names and ASCII aliases assign again, including null.
func asciiProperty(name string) string {
	raw := []byte(name)
	for i, value := range raw {
		if value >= 128 {
			return ""
		}
		if value >= 'A' && value <= 'Z' {
			raw[i] += 'a' - 'A'
		}
	}
	return string(raw)
}

func decodeUser(raw []byte, id *string, displayName, principalName, userType, mail **string, fields *map[string]json.RawMessage) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return client.ErrDecode
	}
	projected := map[string]*string{}
	captured := map[string]json.RawMessage{}
	for decoder.More() {
		token, err = decoder.Token()
		name, ok := token.(string)
		if err != nil || !ok {
			return client.ErrDecode
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return client.ErrDecode
		}
		captured[name] = value
		key := asciiProperty(name)
		switch key {
		case "id", "displayname", "userprincipalname", "usertype", "mail":
			var text aadString
			if text.UnmarshalJSON(value) != nil {
				return client.ErrDecode
			}
			projected[key] = text.value
		}
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		return client.ErrDecode
	}
	if _, err = decoder.Token(); err != io.EOF {
		return client.ErrDecode
	}
	*id = ""
	if projected["id"] != nil {
		*id = *projected["id"]
	}
	*displayName, *principalName, *userType, *mail = projected["displayname"], projected["userprincipalname"], projected["usertype"], projected["mail"]
	*fields = captured
	return nil
}
