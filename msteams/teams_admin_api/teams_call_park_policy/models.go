// Package teamscallparkpolicy manages call park policies and their assignments.
package teamscallparkpolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid call park policy input")
var ErrResponse = errors.New("unexpected call park policy response")

// Policy retains arbitrary modern configuration fields and exact original JSON.
// Known fields project independently; unprojectable future shapes remain raw.
type Policy struct {
	Identity           string                     `json:"Identity"`
	Description        *string                    `json:"Description"`
	AllowCallPark      *bool                      `json:"AllowCallPark"`
	ParkTimeoutSeconds *int64                     `json:"ParkTimeoutSeconds"`
	PickupRangeStart   *int64                     `json:"PickupRangeStart"`
	PickupRangeEnd     *int64                     `json:"PickupRangeEnd"`
	RawFields          map[string]json.RawMessage `json:"-"`
	RawJSON            json.RawMessage            `json:"-"`
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
	value.AllowCallPark = project[bool](fields["AllowCallPark"])
	value.ParkTimeoutSeconds = project[int64](fields["ParkTimeoutSeconds"])
	value.PickupRangeStart = project[int64](fields["PickupRangeStart"])
	value.PickupRangeEnd = project[int64](fields["PickupRangeEnd"])
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
func (Policy) String() string     { return "call park policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Settings uses optional typed values: nil omits, while false/empty/zero remain
// explicit. Int64 fields never pass through float64 or native-size int.
type Settings struct {
	Description        *string
	AllowCallPark      *bool
	ParkTimeoutSeconds *int64
	PickupRangeStart   *int64
	PickupRangeEnd     *int64
}
type CreateRequest struct {
	Identity string
	Settings
}
type UpdateRequest = Settings

func writeBody(identity string, input Settings) map[string]any {
	out := map[string]any{"Identity": identity}
	if input.Description != nil {
		out["Description"] = *input.Description
	}
	if input.AllowCallPark != nil {
		out["AllowCallPark"] = *input.AllowCallPark
	}
	for _, v := range []struct {
		name  string
		value *int64
	}{{"ParkTimeoutSeconds", input.ParkTimeoutSeconds}, {"PickupRangeStart", input.PickupRangeStart}, {"PickupRangeEnd", input.PickupRangeEnd}} {
		if v.value != nil {
			out[v.name] = *v.value
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
func (GroupAssignment) String() string     { return "call park group assignment (redacted)" }
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
