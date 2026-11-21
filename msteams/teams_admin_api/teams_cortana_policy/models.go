// Package teamscortanapolicy manages Cortana policies and their assignments.
package teamscortanapolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid Cortana policy input")
var ErrResponse = errors.New("unexpected Cortana policy response")

// Policy retains arbitrary modern configuration fields and exact original JSON.
// Known fields project independently; unprojectable future shapes remain raw.
type Policy struct {
	Identity                         string                     `json:"Identity"`
	Description                      *string                    `json:"Description"`
	CortanaVoiceInvocationMode       *string                    `json:"CortanaVoiceInvocationMode"`
	AllowCortanaVoiceInvocation      *bool                      `json:"AllowCortanaVoiceInvocation"`
	AllowCortanaAmbientListening     *bool                      `json:"AllowCortanaAmbientListening"`
	AllowCortanaInContextSuggestions *bool                      `json:"AllowCortanaInContextSuggestions"`
	RawFields                        map[string]json.RawMessage `json:"-"`
	RawJSON                          json.RawMessage            `json:"-"`
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
	value.CortanaVoiceInvocationMode = project[string](fields["CortanaVoiceInvocationMode"])
	value.AllowCortanaVoiceInvocation = project[bool](fields["AllowCortanaVoiceInvocation"])
	value.AllowCortanaAmbientListening = project[bool](fields["AllowCortanaAmbientListening"])
	value.AllowCortanaInContextSuggestions = project[bool](fields["AllowCortanaInContextSuggestions"])
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
func (Policy) String() string     { return "Cortana policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Settings uses optional typed values: nil omits while false and empty stay
// explicit. Source public strings bind null as empty; Boolean null is rejected.
type Settings struct {
	Description                      *string
	CortanaVoiceInvocationMode       *string
	AllowCortanaVoiceInvocation      *bool
	AllowCortanaAmbientListening     *bool
	AllowCortanaInContextSuggestions *bool
}
type CreateRequest struct {
	Identity string
	Settings
}
type UpdateRequest = Settings

func writeBody(identity string, input Settings) map[string]any {
	out := map[string]any{"Identity": identity}
	for _, field := range []struct {
		name  string
		value *string
	}{{"Description", input.Description}, {"CortanaVoiceInvocationMode", input.CortanaVoiceInvocationMode}} {
		if field.value != nil {
			out[field.name] = *field.value
		}
	}
	for _, field := range []struct {
		name  string
		value *bool
	}{{"AllowCortanaVoiceInvocation", input.AllowCortanaVoiceInvocation}, {"AllowCortanaAmbientListening", input.AllowCortanaAmbientListening}, {"AllowCortanaInContextSuggestions", input.AllowCortanaInContextSuggestions}} {
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
