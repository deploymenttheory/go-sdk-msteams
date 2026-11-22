// Package teamsnetworkroamingpolicy manages the public modern network roaming API.
package teamsnetworkroamingpolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var ErrInput = errors.New("invalid network roaming policy input")
var ErrResponse = errors.New("unexpected network roaming policy response")

// Policy projects known getters while retaining the complete modern response.
type Policy struct {
	Identity       string                     `json:"Identity"`
	Description    *string                    `json:"Description"`
	AllowIPVideo   *bool                      `json:"AllowIPVideo"`
	MediaBitRateKb *int64                     `json:"MediaBitRateKb"`
	RawFields      map[string]json.RawMessage `json:"-"`
	RawJSON        json.RawMessage            `json:"-"`
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
	v := Policy{RawFields: fields, RawJSON: append(json.RawMessage(nil), raw...), Description: project[string](fields["Description"]), AllowIPVideo: project[bool](fields["AllowIPVideo"]), MediaBitRateKb: project[int64](fields["MediaBitRateKb"])}
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
func (Policy) String() string     { return "network roaming policy (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Settings preserves optional public String, Boolean and signed Int64 bindings.
// The public numeric/Boolean parameters are non-nullable; nil omits a setting.
type Settings struct {
	Description    *string
	AllowIPVideo   *bool
	MediaBitRateKb *int64
}
type CreateRequest struct {
	Identity string
	Settings
}
type UpdateRequest = Settings

func writeBody(id string, input Settings) map[string]any {
	out := map[string]any{"Identity": id}
	if input.Description != nil {
		out["Description"] = *input.Description
	}
	if input.AllowIPVideo != nil {
		out["AllowIPVideo"] = *input.AllowIPVideo
	}
	if input.MediaBitRateKb != nil {
		out["MediaBitRateKb"] = *input.MediaBitRateKb
	}
	return out
}
