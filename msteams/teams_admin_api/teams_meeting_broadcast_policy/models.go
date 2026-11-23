// Package teamsmeetingbroadcastpolicy manages Teams meeting broadcast (live
// event) policies and their assignments.
package teamsmeetingbroadcastpolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid meeting broadcast policy input")
var ErrResponse = errors.New("unexpected meeting broadcast policy response")

// Policy retains arbitrary modern configuration fields and exact original JSON.
// Known fields project independently; unprojectable future shapes remain raw.
type Policy struct {
	Identity                        string                     `json:"Identity"`
	Description                     *string                    `json:"Description"`
	AllowBroadcastScheduling        *bool                      `json:"AllowBroadcastScheduling"`
	AllowBroadcastTranscription     *bool                      `json:"AllowBroadcastTranscription"`
	BroadcastAttendeeVisibilityMode *string                    `json:"BroadcastAttendeeVisibilityMode"`
	BroadcastRecordingMode          *string                    `json:"BroadcastRecordingMode"`
	RawFields                       map[string]json.RawMessage `json:"-"`
	RawJSON                         json.RawMessage            `json:"-"`
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
	value.AllowBroadcastScheduling = project[bool](fields["AllowBroadcastScheduling"])
	value.AllowBroadcastTranscription = project[bool](fields["AllowBroadcastTranscription"])
	value.BroadcastAttendeeVisibilityMode = project[string](fields["BroadcastAttendeeVisibilityMode"])
	value.BroadcastRecordingMode = project[string](fields["BroadcastRecordingMode"])

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
func (Policy) String() string     { return "meeting broadcast policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Settings uses optional typed values: nil omits while false and empty stay
// explicit. The public proxy binds non-nullable Booleans and opaque strings; it
// declares no enum validation, so mode strings are passed through unchanged.
type Settings struct {
	Description                     *string
	AllowBroadcastScheduling        *bool
	AllowBroadcastTranscription     *bool
	BroadcastAttendeeVisibilityMode *string
	BroadcastRecordingMode          *string
}
type CreateRequest struct {
	Identity string
	Settings
}
type UpdateRequest = Settings

func writeBody(id string, input Settings) map[string]any {
	out := map[string]any{"Identity": id}
	for _, field := range []struct {
		name  string
		value *bool
	}{{"AllowBroadcastScheduling", input.AllowBroadcastScheduling}, {"AllowBroadcastTranscription", input.AllowBroadcastTranscription}} {
		if field.value != nil {
			out[field.name] = *field.value
		}
	}
	for _, field := range []struct {
		name  string
		value *string
	}{{"Description", input.Description}, {"BroadcastAttendeeVisibilityMode", input.BroadcastAttendeeVisibilityMode}, {"BroadcastRecordingMode", input.BroadcastRecordingMode}} {
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
func (GroupAssignment) String() string     { return "meeting broadcast group assignment (redacted)" }
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
