// Package tenantmigrationconfiguration manages the generated tenant (Meeting
// Migration Service) migration configuration API.
package tenantmigrationconfiguration

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid tenant migration configuration input")
var ErrResponse = errors.New("unexpected tenant migration configuration HTTP status")

// Configuration is one tenant migration configuration. MeetingMigrationEnabled
// reports whether the Meeting Migration Service (MMS) is enabled. It is
// nullable so an absent or JSON null read value stays distinct.
type Configuration struct {
	Identity                *string                    `json:"Identity"`
	MeetingMigrationEnabled *bool                      `json:"MeetingMigrationEnabled"`
	RawFields               map[string]json.RawMessage `json:"-"`
	RawJSON                 json.RawMessage            `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type plain Configuration
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Configuration(value)
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Configuration
	return json.Marshal(plain(v))
}
func (Configuration) String() string     { return "tenant migration configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }

// UpdateRequest preserves an omitted MeetingMigrationEnabled (nil) versus an
// explicit false/true. Public 8.1.0 binds a non-nullable System.Boolean, so
// there is no null request setting.
type UpdateRequest struct{ MeetingMigrationEnabled *bool }

// body returns the sparse PUT document: Identity plus the bound setting.
func (r UpdateRequest) body(identity string) map[string]any {
	body := map[string]any{"Identity": identity}
	if r.MeetingMigrationEnabled != nil {
		body["MeetingMigrationEnabled"] = *r.MeetingMigrationEnabled
	}
	return body
}
