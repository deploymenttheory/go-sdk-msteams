package directoryuser

import (
	"bytes"
	"encoding/json"

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

func decodeUser(raw []byte, id *string, displayName, principalName, userType, mail **string, fields *map[string]json.RawMessage) error {
	var wire struct {
		ID            aadString `json:"id"`
		DisplayName   aadString `json:"displayName"`
		PrincipalName aadString `json:"userPrincipalName"`
		UserType      aadString `json:"userType"`
		Mail          aadString `json:"mail"`
	}
	if json.Unmarshal(raw, &wire) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	if wire.ID.value != nil {
		*id = *wire.ID.value
	}
	*displayName, *principalName, *userType, *mail = wire.DisplayName.value, wire.PrincipalName.value, wire.UserType.value, wire.Mail.value
	return nil
}
