// Package teamsupdatemanagementpolicy manages compiled TeamsUpdateManagementPolicy operations.
package teamsupdatemanagementpolicy

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Policy owns all returned fields and projects the nullable pinned getters.
type Policy struct {
	Identity                  *string                    `json:"Identity"`
	Description               *string                    `json:"Description"`
	AllowManagedUpdates       *bool                      `json:"AllowManagedUpdates"`
	AllowPreview              *bool                      `json:"AllowPreview"`
	UpdateDayOfWeek           *int64                     `json:"UpdateDayOfWeek"`
	UpdateTime                *string                    `json:"UpdateTime"`
	UpdateTimeOfDay           *DateTime                  `json:"UpdateTimeOfDay"`
	AllowPublicPreview        *string                    `json:"AllowPublicPreview"`
	UseNewTeamsClient         *string                    `json:"UseNewTeamsClient"`
	BlockLegacyAuthorization  *bool                      `json:"BlockLegacyAuthorization"`
	DisabledInProductMessages []string                   `json:"DisabledInProductMessages"`
	RawFields                 map[string]json.RawMessage `json:"-"`
	RawJSON                   json.RawMessage            `json:"-"`
}

func (v *Policy) UnmarshalJSON(raw []byte) error {
	value := Policy{DisabledInProductMessages: []string{}}
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			return json.Unmarshal(body, &value.Identity)
		case "description":
			return json.Unmarshal(body, &value.Description)
		case "allowmanagedupdates":
			return json.Unmarshal(body, &value.AllowManagedUpdates)
		case "allowpreview":
			return json.Unmarshal(body, &value.AllowPreview)
		case "updatedayofweek":
			return json.Unmarshal(body, &value.UpdateDayOfWeek)
		case "updatetime":
			return json.Unmarshal(body, &value.UpdateTime)
		case "updatetimeofday":
			return json.Unmarshal(body, &value.UpdateTimeOfDay)
		case "allowpublicpreview":
			return json.Unmarshal(body, &value.AllowPublicPreview)
		case "usenewteamsclient":
			return json.Unmarshal(body, &value.UseNewTeamsClient)
		case "blocklegacyauthorization":
			return json.Unmarshal(body, &value.BlockLegacyAuthorization)
		case "disabledinproductmessages":
			if !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
				var values []string
				if json.Unmarshal(body, &values) != nil {
					return client.ErrDecode
				}
				value.DisabledInProductMessages = append(value.DisabledInProductMessages, values...)
			}
		}
		return nil
	})
	if e != nil {
		return client.ErrDecode
	}
	if value.UpdateDayOfWeek != nil && (*value.UpdateDayOfWeek < 0 || *value.UpdateDayOfWeek > 6) {
		return client.ErrDecode
	}
	value.RawFields = fields
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = value
	return nil
}
func (v Policy) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Policy
	return json.Marshal(plain(v))
}
func (Policy) String() string     { return "TeamsUpdateManagementPolicy (redacted)" }
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
