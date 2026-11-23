// Package teamstenantabuseconfiguration manages compiled TeamsTenantAbuseConfiguration operations.
package teamstenantabuseconfiguration

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Configuration owns all returned fields and projects the nullable pinned getters.
type Configuration struct {
	Identity                             *string                    `json:"Identity"`
	LowSeatLimit                         *uint64                    `json:"LowSeatLimit"`
	TrialTenantValidation                *string                    `json:"TrialTenantValidation"`
	SkipValidation                       *string                    `json:"SkipValidation"`
	CreateThreadThresholdPerSeat         *uint64                    `json:"CreateThreadThresholdPerSeat"`
	AddMemberThresholdPerSeat            *uint64                    `json:"AddMemberThresholdPerSeat"`
	SendMessageThresholdPerSeat          *uint64                    `json:"SendMessageThresholdPerSeat"`
	CreateThreadThresholdForTrialTenants *uint64                    `json:"CreateThreadThresholdForTrialTenants"`
	AddMemberThresholdForTrialTenants    *uint64                    `json:"AddMemberThresholdForTrialTenants"`
	SendMessageThresholdForTrialTenants  *uint64                    `json:"SendMessageThresholdForTrialTenants"`
	BlockedTenants                       []string                   `json:"BlockedTenants"`
	RawFields                            map[string]json.RawMessage `json:"-"`
	RawJSON                              json.RawMessage            `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	value := Configuration{BlockedTenants: []string{}}
	fields, e := object(raw, func(key string, body json.RawMessage) error {
		switch strings.ToLower(key) {
		case "identity":
			return json.Unmarshal(body, &value.Identity)
		case "lowseatlimit":
			return json.Unmarshal(body, &value.LowSeatLimit)
		case "trialtenantvalidation":
			return json.Unmarshal(body, &value.TrialTenantValidation)
		case "skipvalidation":
			return json.Unmarshal(body, &value.SkipValidation)
		case "createthreadthresholdperseat":
			return json.Unmarshal(body, &value.CreateThreadThresholdPerSeat)
		case "addmemberthresholdperseat":
			return json.Unmarshal(body, &value.AddMemberThresholdPerSeat)
		case "sendmessagethresholdperseat":
			return json.Unmarshal(body, &value.SendMessageThresholdPerSeat)
		case "createthreadthresholdfortrialtenants":
			return json.Unmarshal(body, &value.CreateThreadThresholdForTrialTenants)
		case "addmemberthresholdfortrialtenants":
			return json.Unmarshal(body, &value.AddMemberThresholdForTrialTenants)
		case "sendmessagethresholdfortrialtenants":
			return json.Unmarshal(body, &value.SendMessageThresholdForTrialTenants)
		case "blockedtenants":
			if !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
				var values []string
				if json.Unmarshal(body, &values) != nil {
					return client.ErrDecode
				}
				value.BlockedTenants = append(value.BlockedTenants, values...)
			}
		}
		return nil
	})
	if e != nil {
		return client.ErrDecode
	}

	value.RawFields = fields
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = value
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Configuration
	return json.Marshal(plain(v))
}
func (Configuration) String() string     { return "TeamsTenantAbuseConfiguration (redacted)" }
func (v Configuration) GoString() string { return v.String() }
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
