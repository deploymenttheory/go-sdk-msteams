// Package teamsroomvideoteleconferencingpolicy manages room video teleconferencing policies and their assignments.
package teamsroomvideoteleconferencingpolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid room video teleconferencing policy input")
var ErrResponse = errors.New("unexpected room video teleconferencing policy response")

// Policy retains arbitrary modern configuration fields and exact original JSON.
// Known fields project independently; unprojectable future shapes remain raw.
type Policy struct {
	Identity             string                     `json:"Identity"`
	Enabled              *bool                      `json:"Enabled"`
	Description          *string                    `json:"Description"`
	AreaCode             *string                    `json:"AreaCode"`
	ReceiveExternalCalls *string                    `json:"ReceiveExternalCalls"`
	ReceiveInternalCalls *string                    `json:"ReceiveInternalCalls"`
	PlaceExternalCalls   *string                    `json:"PlaceExternalCalls"`
	PlaceInternalCalls   *string                    `json:"PlaceInternalCalls"`
	RawFields            map[string]json.RawMessage `json:"-"`
	RawJSON              json.RawMessage            `json:"-"`
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

	value.Enabled = project[bool](fields["Enabled"])
	value.Description = project[string](fields["Description"])
	value.AreaCode = project[string](fields["AreaCode"])
	value.ReceiveExternalCalls = project[string](fields["ReceiveExternalCalls"])
	value.ReceiveInternalCalls = project[string](fields["ReceiveInternalCalls"])
	value.PlaceExternalCalls = project[string](fields["PlaceExternalCalls"])
	value.PlaceInternalCalls = project[string](fields["PlaceInternalCalls"])

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
func (Policy) String() string     { return "room video teleconferencing policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Settings uses optional typed values: nil omits while false and empty stay
// explicit. Source public strings bind null as empty; Boolean null is rejected.
type Settings struct {
	Enabled              *bool
	Description          *string
	AreaCode             *string
	ReceiveExternalCalls *string
	ReceiveInternalCalls *string
	PlaceExternalCalls   *string
	PlaceInternalCalls   *string
}
type CreateRequest struct {
	Identity string
	Settings
}
type UpdateRequest = Settings

func writeBody(id string, input Settings) map[string]any {
	out := map[string]any{"Identity": id}
	if input.Enabled != nil {
		out["Enabled"] = *input.Enabled
	}
	for _, field := range []struct {
		name  string
		value *string
	}{{"Description", input.Description}, {"AreaCode", input.AreaCode}, {"ReceiveExternalCalls", input.ReceiveExternalCalls}, {"ReceiveInternalCalls", input.ReceiveInternalCalls}, {"PlaceExternalCalls", input.PlaceExternalCalls}, {"PlaceInternalCalls", input.PlaceInternalCalls}} {
		if field.value != nil {
			out[field.name] = *field.value
		}
	}
	return out
}

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
func (GroupAssignment) String() string     { return "Cortana group assignment (redacted)" }
func (v GroupAssignment) GoString() string { return v.String() }

type UserAssignment struct{ Identity, PolicyName string }
type TenantAssignment struct{ PolicyName string }
type GroupGrant struct {
	ID, PolicyName string
	Rank           *int32
}
type Assignment struct {
	User   *UserAssignment
	Tenant *TenantAssignment
	Group  *GroupGrant
}
