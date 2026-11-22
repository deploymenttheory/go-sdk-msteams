// Package teamsworkloadpolicy manages the public modern workload API.
package teamsworkloadpolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid workload policy input")
var ErrResponse = errors.New("unexpected workload policy response")

// Policy projects known getters while retaining the complete modern response.
type Policy struct {
	Identity             string                     `json:"Identity"`
	AllowCalling         *bool                      `json:"AllowCalling"`
	AllowCallingPinned   *bool                      `json:"AllowCallingPinned"`
	AllowMeeting         *bool                      `json:"AllowMeeting"`
	AllowMeetingPinned   *bool                      `json:"AllowMeetingPinned"`
	AllowMessaging       *bool                      `json:"AllowMessaging"`
	AllowMessagingPinned *bool                      `json:"AllowMessagingPinned"`
	Description          *string                    `json:"Description"`
	RawFields            map[string]json.RawMessage `json:"-"`
	RawJSON              json.RawMessage            `json:"-"`
}

func project[T any](raw json.RawMessage) *T {
	var v *T
	if json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return v
}
func (p *Policy) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	v := Policy{RawFields: fields, RawJSON: append(json.RawMessage(nil), raw...)}
	v.AllowCalling = project[bool](fields["AllowCalling"])
	v.AllowCallingPinned = project[bool](fields["AllowCallingPinned"])
	v.AllowMeeting = project[bool](fields["AllowMeeting"])
	v.AllowMeetingPinned = project[bool](fields["AllowMeetingPinned"])
	v.AllowMessaging = project[bool](fields["AllowMessaging"])
	v.AllowMessagingPinned = project[bool](fields["AllowMessagingPinned"])
	v.Description = project[string](fields["Description"])

	if id := project[string](fields["Identity"]); id != nil {
		v.Identity = *id
	}
	*p = v
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type plain Policy
	return json.Marshal(plain(p))
}
func (Policy) String() string     { return "workload policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Settings preserves optional public String and non-nullable Boolean bindings.
// Nil omits a setting; false and empty strings remain explicitly bound.
type Settings struct {
	AllowCalling         *bool
	AllowCallingPinned   *bool
	AllowMeeting         *bool
	AllowMeetingPinned   *bool
	AllowMessaging       *bool
	AllowMessagingPinned *bool
	Description          *string
}
type CreateRequest struct {
	Identity string
	Settings
}
type UpdateRequest = Settings

func writeBody(id string, input Settings) map[string]any {
	out := map[string]any{"Identity": id}
	if input.AllowCalling != nil {
		out["AllowCalling"] = *input.AllowCalling
	}
	if input.AllowCallingPinned != nil {
		out["AllowCallingPinned"] = *input.AllowCallingPinned
	}
	if input.AllowMeeting != nil {
		out["AllowMeeting"] = *input.AllowMeeting
	}
	if input.AllowMeetingPinned != nil {
		out["AllowMeetingPinned"] = *input.AllowMeetingPinned
	}
	if input.AllowMessaging != nil {
		out["AllowMessaging"] = *input.AllowMessaging
	}
	if input.AllowMessagingPinned != nil {
		out["AllowMessagingPinned"] = *input.AllowMessagingPinned
	}
	if input.Description != nil {
		out["Description"] = *input.Description
	}
	return out
}
