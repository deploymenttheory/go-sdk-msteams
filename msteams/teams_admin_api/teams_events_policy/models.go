// Package teamseventspolicy manages compiled Teams events policy.
package teamseventspolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// Source strings remain opaque and both Boolean settings remain nullable.
type Policy struct {
	Identity                                *string                    `json:"Identity"`
	AllowWebinars                           *string                    `json:"AllowWebinars"`
	EventAccessType                         *string                    `json:"EventAccessType"`
	AllowTownhalls                          *string                    `json:"AllowTownhalls"`
	TownhallEventAttendeeAccess             *string                    `json:"TownhallEventAttendeeAccess"`
	ExternalPresenterJoinVerification       *string                    `json:"ExternalPresenterJoinVerification"`
	AllowEmailEditing                       *string                    `json:"AllowEmailEditing"`
	AllowedQuestionTypesInRegistrationForm  *string                    `json:"AllowedQuestionTypesInRegistrationForm"`
	AllowEventIntegrations                  *bool                      `json:"AllowEventIntegrations"`
	AllowedWebinarTypesForRecordingPublish  *string                    `json:"AllowedWebinarTypesForRecordingPublish"`
	AllowedTownhallTypesForRecordingPublish *string                    `json:"AllowedTownhallTypesForRecordingPublish"`
	RecordingForTownhall                    *string                    `json:"RecordingForTownhall"`
	RecordingForWebinar                     *string                    `json:"RecordingForWebinar"`
	TranscriptionForTownhall                *string                    `json:"TranscriptionForTownhall"`
	TranscriptionForWebinar                 *string                    `json:"TranscriptionForWebinar"`
	TownhallChatExperience                  *string                    `json:"TownhallChatExperience"`
	BroadcastPremiumApps                    *string                    `json:"BroadcastPremiumApps"`
	UseMicrosoftECDN                        *bool                      `json:"UseMicrosoftECDN"`
	ImmersiveEvents                         *string                    `json:"ImmersiveEvents"`
	TownhallMaxResolution                   *string                    `json:"TownhallMaxResolution"`
	HighBitrateForTownhall                  *string                    `json:"HighBitrateForTownhall"`
	BackroomChat                            *string                    `json:"BackroomChat"`
	Registration                            *string                    `json:"Registration"`
	AllowEngagementReport                   *string                    `json:"AllowEngagementReport"`
	InfoShownInReportMode                   *string                    `json:"InfoShownInReportMode"`
	Description                             *string                    `json:"Description"`
	RawFields                               map[string]json.RawMessage `json:"-"`
	RawJSON                                 json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Policy(next)
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}
func (Policy) String() string     { return "events policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose enums, list parsing or cross-field rules.
type Settings struct {
	AllowWebinars                           policy.Field[string] `json:"AllowWebinars,omitzero"`
	EventAccessType                         policy.Field[string] `json:"EventAccessType,omitzero"`
	AllowTownhalls                          policy.Field[string] `json:"AllowTownhalls,omitzero"`
	TownhallEventAttendeeAccess             policy.Field[string] `json:"TownhallEventAttendeeAccess,omitzero"`
	ExternalPresenterJoinVerification       policy.Field[string] `json:"ExternalPresenterJoinVerification,omitzero"`
	AllowEmailEditing                       policy.Field[string] `json:"AllowEmailEditing,omitzero"`
	AllowedQuestionTypesInRegistrationForm  policy.Field[string] `json:"AllowedQuestionTypesInRegistrationForm,omitzero"`
	AllowEventIntegrations                  policy.Field[bool]   `json:"AllowEventIntegrations,omitzero"`
	AllowedWebinarTypesForRecordingPublish  policy.Field[string] `json:"AllowedWebinarTypesForRecordingPublish,omitzero"`
	AllowedTownhallTypesForRecordingPublish policy.Field[string] `json:"AllowedTownhallTypesForRecordingPublish,omitzero"`
	RecordingForTownhall                    policy.Field[string] `json:"RecordingForTownhall,omitzero"`
	RecordingForWebinar                     policy.Field[string] `json:"RecordingForWebinar,omitzero"`
	TranscriptionForTownhall                policy.Field[string] `json:"TranscriptionForTownhall,omitzero"`
	TranscriptionForWebinar                 policy.Field[string] `json:"TranscriptionForWebinar,omitzero"`
	TownhallChatExperience                  policy.Field[string] `json:"TownhallChatExperience,omitzero"`
	BroadcastPremiumApps                    policy.Field[string] `json:"BroadcastPremiumApps,omitzero"`
	UseMicrosoftECDN                        policy.Field[bool]   `json:"UseMicrosoftECDN,omitzero"`
	ImmersiveEvents                         policy.Field[string] `json:"ImmersiveEvents,omitzero"`
	TownhallMaxResolution                   policy.Field[string] `json:"TownhallMaxResolution,omitzero"`
	HighBitrateForTownhall                  policy.Field[string] `json:"HighBitrateForTownhall,omitzero"`
	BackroomChat                            policy.Field[string] `json:"BackroomChat,omitzero"`
	Registration                            policy.Field[string] `json:"Registration,omitzero"`
	AllowEngagementReport                   policy.Field[string] `json:"AllowEngagementReport,omitzero"`
	InfoShownInReportMode                   policy.Field[string] `json:"InfoShownInReportMode,omitzero"`
	Description                             policy.Field[string] `json:"Description,omitzero"`
}

// CreateRequest permits no settings, matching the compiled New processing.
type CreateRequest struct {
	Settings
	policy.Controls
}

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
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "events policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "events policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
