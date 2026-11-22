// Package teamsshiftsapppolicy manages the generated Shifts app policy API.
package teamsshiftsapppolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid Shifts app policy input")
var ErrResponse = errors.New("unexpected Shifts app policy HTTP status")

type Policy struct {
	Identity                        *string                    `json:"Identity"`
	AllowTimeClockLocationDetection *bool                      `json:"AllowTimeClockLocationDetection"`
	RawFields                       map[string]json.RawMessage `json:"-"`
	RawJSON                         json.RawMessage            `json:"-"`
}

func (v *Policy) UnmarshalJSON(raw []byte) error {
	type plain Policy
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	value.RawJSON = append(json.RawMessage(nil), raw...)
	*v = Policy(value)
	return nil
}
func (v Policy) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain Policy
	return json.Marshal(plain(v))
}
func (Policy) String() string     { return "Shifts app policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// UpdateRequest preserves an omitted Boolean versus explicit false/true.
// Public 8.1.0 binds a non-nullable Boolean; there is no null request setting.
type UpdateRequest struct{ AllowTimeClockLocationDetection *bool }
