// Package teamsupgradepolicy manages upgrade policies and their assignments.
package teamsupgradepolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid upgrade policy input")
var ErrResponse = errors.New("unexpected upgrade policy response")

// ErrMigrationParameterBinding matches the pinned modern Grant wrapper, which
// exposes MigrateMeetingsToTeams but forwards it to a function without that parameter.
// Both explicit false and true fail binding; omitting it allows native grant.
var ErrMigrationParameterBinding = errors.New("MigrateMeetingsToTeams cannot bind to the pinned modern grant function")

// Policy retains arbitrary modern configuration fields and exact original JSON.
// Known fields project independently; unprojectable future shapes remain raw.
type Policy struct {
	Identity       string                     `json:"Identity"`
	Description    *string                    `json:"Description"`
	Mode           *string                    `json:"Mode"`
	NotifySfbUsers *bool                      `json:"NotifySfbUsers"`
	Action         *string                    `json:"Action"`
	RawFields      map[string]json.RawMessage `json:"-"`
	RawJSON        json.RawMessage            `json:"-"`
}

func project[T any](raw json.RawMessage) *T {
	var value *T
	if json.Unmarshal(raw, &value) != nil {
		return nil
	}
	return value
}
func (p *Policy) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	value := Policy{RawFields: fields, RawJSON: append(json.RawMessage(nil), raw...)}
	if v := project[string](fields["Identity"]); v != nil {
		value.Identity = *v
	}

	value.Description = project[string](fields["Description"])
	value.Mode = project[string](fields["Mode"])
	value.NotifySfbUsers = project[bool](fields["NotifySfbUsers"])
	value.Action = project[string](fields["Action"])

	*p = value
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type plain Policy
	return json.Marshal(plain(p))
}
func (Policy) String() string     { return "upgrade policy (redacted)" }
func (v Policy) GoString() string { return v.String() }
