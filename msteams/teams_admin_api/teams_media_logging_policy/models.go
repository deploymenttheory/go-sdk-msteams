// Package teamsmedialoggingpolicy reads media logging policies and manages their assignments.
package teamsmedialoggingpolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid media logging policy input")
var ErrResponse = errors.New("unexpected media logging policy response")

// Policy retains arbitrary modern configuration fields and exact original JSON.
// Known fields project independently; unprojectable future shapes remain raw.
type Policy struct {
	Identity          string                     `json:"Identity"`
	Description       *string                    `json:"Description"`
	AllowMediaLogging *bool                      `json:"AllowMediaLogging"`
	RawFields         map[string]json.RawMessage `json:"-"`
	RawJSON           json.RawMessage            `json:"-"`
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
	value.AllowMediaLogging = project[bool](fields["AllowMediaLogging"])
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
func (Policy) String() string     { return "media logging policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// GroupAssignment is the accepted group grant receipt. A 204 leaves it empty.
type GroupAssignment struct {
	OperationID string          `json:"OperationId"`
	Accepted    bool            `json:"-"`
	RawJSON     json.RawMessage `json:"-"`
}

func (v GroupAssignment) MarshalJSON() ([]byte, error) {
	if v.RawJSON != nil {
		return append([]byte(nil), v.RawJSON...), nil
	}
	type plain GroupAssignment
	return json.Marshal(plain(v))
}
func (GroupAssignment) String() string     { return "media logging group assignment (redacted)" }
func (v GroupAssignment) GoString() string { return v.String() }

// UserAssignment maps the Grant Identity parameter set; empty PolicyName clears.
type UserAssignment struct{ Identity, PolicyName string }

// TenantAssignment maps the GrantToTenant parameter set (-Global).
type TenantAssignment struct{ PolicyName string }

// GroupGrant maps the GrantToGroup parameter set with optional signed Rank.
type GroupGrant struct {
	ID, PolicyName string
	Rank           *int32
}

// Assignment selects exactly one public Grant parameter set.
type Assignment struct {
	User   *UserAssignment
	Tenant *TenantAssignment
	Group  *GroupGrant
}
