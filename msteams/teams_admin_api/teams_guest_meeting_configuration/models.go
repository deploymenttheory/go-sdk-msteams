// Package teamsguestmeetingconfiguration manages compiled Teams guest meeting configuration.
package teamsguestmeetingconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Configuration preserves the eight nullable compiled getters and an owned
// complete snapshot. Source strings remain opaque; no local enum is invented.
type Configuration struct {
	Identity                                   *string                    `json:"Identity"`
	AllowIPVideo                               *bool                      `json:"AllowIPVideo"`
	ScreenSharingMode                          *string                    `json:"ScreenSharingMode"`
	AllowMultipleScreenshare                   *bool                      `json:"AllowMultipleScreenshare"`
	AllowMeetNow                               *bool                      `json:"AllowMeetNow"`
	LiveCaptionsEnabledType                    *string                    `json:"LiveCaptionsEnabledType"`
	AllowTranscription                         *bool                      `json:"AllowTranscription"`
	AllowParticipantGiveRequestControl         *bool                      `json:"AllowParticipantGiveRequestControl"`
	AllowExternalParticipantGiveRequestControl *bool                      `json:"AllowExternalParticipantGiveRequestControl"`
	RawFields                                  map[string]json.RawMessage `json:"-"`
	RawJSON                                    json.RawMessage            `json:"-"`
}

func (p *Configuration) UnmarshalJSON(raw []byte) error {
	type snapshot Configuration
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Configuration(next)
	return nil
}
func (p Configuration) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Configuration
	return json.Marshal(snapshot(p))
}
func (Configuration) String() string     { return "guest meeting configuration (redacted)" }
func (p Configuration) GoString() string { return p.String() }
