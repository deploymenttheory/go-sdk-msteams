// Package teamssharedcallingroutingpolicy implements compiled shared calling policy operations.
package teamssharedcallingroutingpolicy

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

type Policy struct {
	Identity         *string                    `json:"Identity"`
	Description      *string                    `json:"Description"`
	ResourceAccount  *string                    `json:"ResourceAccount"`
	EmergencyNumbers []string                   `json:"EmergencyNumbers"`
	RawFields        map[string]json.RawMessage `json:"-"`
	RawJSON          json.RawMessage            `json:"-"`
}

func (v *Policy) UnmarshalJSON(raw []byte) error {
	next := Policy{EmergencyNumbers: []string{}}
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			return json.Unmarshal(body, &next.Identity)
		case "description":
			return json.Unmarshal(body, &next.Description)
		case "resourceaccount":
			return json.Unmarshal(body, &next.ResourceAccount)
		case "emergencynumbers":
			if bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
				return client.ErrDecode
			}
			var items []string
			if json.Unmarshal(body, &items) != nil {
				return client.ErrDecode
			}
			next.EmergencyNumbers = append(next.EmergencyNumbers, items...)
		}
		return nil
	})
	if e != nil {
		return client.ErrDecode
	}
	next.RawFields = fields
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*v = next
	return nil
}
func (v Policy) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Policy
	return json.Marshal(plain(v))
}
func (Policy) String() string     { return "shared calling routing policy (redacted)" }
func (v Policy) GoString() string { return v.String() }
func object(raw []byte, visit func(string, json.RawMessage) error) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return nil, client.ErrDecode
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, e := d.Token(); e != nil {
		return nil, client.ErrDecode
	}
	for d.More() {
		key, e := d.Token()
		if e != nil {
			return nil, client.ErrDecode
		}
		name, ok := key.(string)
		if !ok {
			return nil, client.ErrDecode
		}
		var body json.RawMessage
		if d.Decode(&body) != nil || visit(name, body) != nil {
			return nil, client.ErrDecode
		}
	}
	return fields, nil
}
