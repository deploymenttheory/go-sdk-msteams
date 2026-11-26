// Package teamsguestmeetingconfiguration manages compiled Teams guest meeting configuration.
package teamsguestmeetingconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
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
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	var wire struct {
		Identity                                   policyrp.ReadString `json:"Identity"`
		AllowIPVideo                               policyrp.ReadBool   `json:"AllowIPVideo"`
		ScreenSharingMode                          policyrp.ReadString `json:"ScreenSharingMode"`
		AllowMultipleScreenshare                   policyrp.ReadBool   `json:"AllowMultipleScreenshare"`
		AllowMeetNow                               policyrp.ReadBool   `json:"AllowMeetNow"`
		LiveCaptionsEnabledType                    policyrp.ReadString `json:"LiveCaptionsEnabledType"`
		AllowTranscription                         policyrp.ReadBool   `json:"AllowTranscription"`
		AllowParticipantGiveRequestControl         policyrp.ReadBool   `json:"AllowParticipantGiveRequestControl"`
		AllowExternalParticipantGiveRequestControl policyrp.ReadBool   `json:"AllowExternalParticipantGiveRequestControl"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return client.ErrDecode
	}
	*p = Configuration{
		Identity:                           wire.Identity.Value,
		AllowIPVideo:                       wire.AllowIPVideo.Value,
		ScreenSharingMode:                  wire.ScreenSharingMode.Value,
		AllowMultipleScreenshare:           wire.AllowMultipleScreenshare.Value,
		AllowMeetNow:                       wire.AllowMeetNow.Value,
		LiveCaptionsEnabledType:            wire.LiveCaptionsEnabledType.Value,
		AllowTranscription:                 wire.AllowTranscription.Value,
		AllowParticipantGiveRequestControl: wire.AllowParticipantGiveRequestControl.Value,
		AllowExternalParticipantGiveRequestControl: wire.AllowExternalParticipantGiveRequestControl.Value,
		RawFields: fields,
		RawJSON:   append(json.RawMessage(nil), raw...),
	}
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
