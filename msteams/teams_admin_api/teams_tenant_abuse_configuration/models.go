// Package teamstenantabuseconfiguration manages compiled TeamsTenantAbuseConfiguration operations.
package teamstenantabuseconfiguration

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
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
	BlockedTenants                       []*string                  `json:"BlockedTenants"`
	RawFields                            map[string]json.RawMessage `json:"-"`
	RawJSON                              json.RawMessage            `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Identity                             policyrp.ReadString `json:"Identity"`
		LowSeatLimit                         policyrp.ReadUint64 `json:"LowSeatLimit"`
		TrialTenantValidation                policyrp.ReadString `json:"TrialTenantValidation"`
		SkipValidation                       policyrp.ReadString `json:"SkipValidation"`
		CreateThreadThresholdPerSeat         policyrp.ReadUint64 `json:"CreateThreadThresholdPerSeat"`
		AddMemberThresholdPerSeat            policyrp.ReadUint64 `json:"AddMemberThresholdPerSeat"`
		SendMessageThresholdPerSeat          policyrp.ReadUint64 `json:"SendMessageThresholdPerSeat"`
		CreateThreadThresholdForTrialTenants policyrp.ReadUint64 `json:"CreateThreadThresholdForTrialTenants"`
		AddMemberThresholdForTrialTenants    policyrp.ReadUint64 `json:"AddMemberThresholdForTrialTenants"`
		SendMessageThresholdForTrialTenants  policyrp.ReadUint64 `json:"SendMessageThresholdForTrialTenants"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return client.ErrDecode
	}
	value := Configuration{
		Identity:                             wire.Identity.Value,
		LowSeatLimit:                         wire.LowSeatLimit.Value,
		TrialTenantValidation:                wire.TrialTenantValidation.Value,
		SkipValidation:                       wire.SkipValidation.Value,
		CreateThreadThresholdPerSeat:         wire.CreateThreadThresholdPerSeat.Value,
		AddMemberThresholdPerSeat:            wire.AddMemberThresholdPerSeat.Value,
		SendMessageThresholdPerSeat:          wire.SendMessageThresholdPerSeat.Value,
		CreateThreadThresholdForTrialTenants: wire.CreateThreadThresholdForTrialTenants.Value,
		AddMemberThresholdForTrialTenants:    wire.AddMemberThresholdForTrialTenants.Value,
		SendMessageThresholdForTrialTenants:  wire.SendMessageThresholdForTrialTenants.Value,
		BlockedTenants:                       []*string{},
	}
	fields, err := object(raw, func(key string, body json.RawMessage) error {
		if strings.EqualFold(key, "BlockedTenants") && !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
			var items []policyrp.ReadString
			if json.Unmarshal(body, &items) != nil {
				return client.ErrDecode
			}
			for _, item := range items {
				value.BlockedTenants = append(value.BlockedTenants, item.Value)
			}
		}
		return nil
	})
	if err != nil {
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
