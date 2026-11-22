// Package teamsmigrationconfiguration manages the generated migration configuration API.
package teamsmigrationconfiguration

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid migration configuration input")
var ErrResponse = errors.New("unexpected migration configuration HTTP status")

type Configuration struct {
	Identity                  *string                    `json:"Identity"`
	EnableLegacyClientInterop *bool                      `json:"EnableLegacyClientInterop"`
	RawFields                 map[string]json.RawMessage `json:"-"`
	RawJSON                   json.RawMessage            `json:"-"`
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
func (Configuration) String() string     { return "migration configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }

// UpdateRequest preserves an omitted Boolean versus explicit false/true.
// Public 8.1.0 binds a non-nullable Boolean; there is no null request setting.
type UpdateRequest struct{ EnableLegacyClientInterop *bool }
