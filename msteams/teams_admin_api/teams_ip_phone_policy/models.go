// Package teamsipphonepolicy manages Teams IP phone policies (sign-in mode,
// home screen, hot desking and Better Together) and their assignments.
package teamsipphonepolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid IP phone policy input")
var ErrResponse = errors.New("unexpected IP phone policy response")

// Policy retains arbitrary modern configuration fields and exact original JSON.
// Known fields project independently; unprojectable future shapes remain raw.
type Policy struct {
	Identity                       string                     `json:"Identity"`
	Description                    *string                    `json:"Description"`
	AllowBetterTogether            *string                    `json:"AllowBetterTogether"`
	AllowHomeScreen                *string                    `json:"AllowHomeScreen"`
	AllowHotDesking                *bool                      `json:"AllowHotDesking"`
	HotDeskingIdleTimeoutInMinutes *int64                     `json:"HotDeskingIdleTimeoutInMinutes"`
	SearchOnCommonAreaPhoneMode    *string                    `json:"SearchOnCommonAreaPhoneMode"`
	SignInMode                     *string                    `json:"SignInMode"`
	RawFields                      map[string]json.RawMessage `json:"-"`
	RawJSON                        json.RawMessage            `json:"-"`
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
	value.AllowBetterTogether = project[string](fields["AllowBetterTogether"])
	value.AllowHomeScreen = project[string](fields["AllowHomeScreen"])
	value.AllowHotDesking = project[bool](fields["AllowHotDesking"])
	value.HotDeskingIdleTimeoutInMinutes = project[int64](fields["HotDeskingIdleTimeoutInMinutes"])
	value.SearchOnCommonAreaPhoneMode = project[string](fields["SearchOnCommonAreaPhoneMode"])
	value.SignInMode = project[string](fields["SignInMode"])

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
func (Policy) String() string     { return "IP phone policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Settings uses optional typed values: nil omits while false, zero and empty
// stay explicit. The public proxy binds AllowHotDesking as a non-nullable
// Boolean, HotDeskingIdleTimeoutInMinutes as a non-nullable Int64 and the mode
// settings as opaque strings. It declares no enum or range validation, so mode
// strings and timeouts are passed through unchanged and server validation applies.
type Settings struct {
	Description                    *string
	AllowBetterTogether            *string
	AllowHomeScreen                *string
	AllowHotDesking                *bool
	HotDeskingIdleTimeoutInMinutes *int64
	SearchOnCommonAreaPhoneMode    *string
	SignInMode                     *string
}
type CreateRequest struct {
	Identity string
	Settings
}
type UpdateRequest = Settings

func writeBody(id string, input Settings) map[string]any {
	out := map[string]any{"Identity": id}
	if input.AllowHotDesking != nil {
		out["AllowHotDesking"] = *input.AllowHotDesking
	}
	if input.HotDeskingIdleTimeoutInMinutes != nil {
		out["HotDeskingIdleTimeoutInMinutes"] = *input.HotDeskingIdleTimeoutInMinutes
	}
	for _, field := range []struct {
		name  string
		value *string
	}{{"Description", input.Description}, {"AllowBetterTogether", input.AllowBetterTogether}, {"AllowHomeScreen", input.AllowHomeScreen}, {"SearchOnCommonAreaPhoneMode", input.SearchOnCommonAreaPhoneMode}, {"SignInMode", input.SignInMode}} {
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
func (GroupAssignment) String() string     { return "IP phone group assignment (redacted)" }
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
