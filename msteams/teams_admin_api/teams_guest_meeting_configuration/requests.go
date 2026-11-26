package teamsguestmeetingconfiguration

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Settings distinguishes omission, explicit null and bound false/empty values
// for every pinned Set parameter. The compiled setter imposes no enum or
// cross-field rule on ScreenSharingMode or LiveCaptionsEnabledType.
type Settings struct {
	AllowIPVideo                               policy.Field[bool]   `json:"AllowIPVideo,omitzero"`
	ScreenSharingMode                          policy.Field[string] `json:"ScreenSharingMode,omitzero"`
	AllowMultipleScreenshare                   policy.Field[bool]   `json:"AllowMultipleScreenshare,omitzero"`
	AllowMeetNow                               policy.Field[bool]   `json:"AllowMeetNow,omitzero"`
	LiveCaptionsEnabledType                    policy.Field[string] `json:"LiveCaptionsEnabledType,omitzero"`
	AllowTranscription                         policy.Field[bool]   `json:"AllowTranscription,omitzero"`
	AllowParticipantGiveRequestControl         policy.Field[bool]   `json:"AllowParticipantGiveRequestControl,omitzero"`
	AllowExternalParticipantGiveRequestControl policy.Field[bool]   `json:"AllowExternalParticipantGiveRequestControl,omitzero"`
}

func (Settings) String() string     { return "guest meeting configuration settings (redacted)" }
func (p Settings) GoString() string { return p.String() }

// PatchRequest selects settings for a single PATCH, with optional body Identity.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the explicit Set read-then-PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
// Acceptance does not establish persisted or effective tenant settings.
type UpdateResult struct {
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "guest meeting configuration update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }
