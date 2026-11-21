package onlinevoicemailpolicy

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Settings preserves bound/unbound/null/false/empty values in sparse CRUD bodies.
// Audio-file settings are reference strings, never permission to read local files.
type Settings struct {
	Description                         policy.Field[string]   `json:"Description,omitzero"`
	ShareData                           policy.Field[string]   `json:"ShareData,omitzero"`
	PrimarySystemPromptLanguage         policy.Field[string]   `json:"PrimarySystemPromptLanguage,omitzero"`
	SecondarySystemPromptLanguage       policy.Field[string]   `json:"SecondarySystemPromptLanguage,omitzero"`
	PreambleAudioFile                   policy.Field[string]   `json:"PreambleAudioFile,omitzero"`
	PostambleAudioFile                  policy.Field[string]   `json:"PostambleAudioFile,omitzero"`
	EnableTranscription                 policy.Field[bool]     `json:"EnableTranscription,omitzero"`
	EnableTranscriptionProfanityMasking policy.Field[bool]     `json:"EnableTranscriptionProfanityMasking,omitzero"`
	EnableEditingCallAnswerRulesSetting policy.Field[bool]     `json:"EnableEditingCallAnswerRulesSetting,omitzero"`
	EnableTranscriptionTranslation      policy.Field[bool]     `json:"EnableTranscriptionTranslation,omitzero"`
	PreamblePostambleMandatory          policy.Field[bool]     `json:"PreamblePostambleMandatory,omitzero"`
	EnableVoicemailTriage               policy.Field[bool]     `json:"EnableVoicemailTriage,omitzero"`
	MaximumRecordingLength              policy.Field[TimeSpan] `json:"MaximumRecordingLength,omitzero"`
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type CreateRequest struct {
	Identity string
	Settings
	policy.Controls
}
type UpdateRequest struct {
	Settings
	policy.Controls
}
type UpdateResult struct {
	Before       *Policy
	After        *Policy
	ReadResponse *resty.Response
	PatchInvoked bool
	HTTPAccepted bool
}
type CreateResult struct {
	Comparison         *Policy
	ComparisonResponse *resty.Response
	ComparisonError    error
	Created            *Policy
	CreateInvoked      bool
	HTTPAccepted       bool
}

func (UpdateResult) String() string     { return "voicemail update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
func (CreateResult) String() string     { return "voicemail create result (redacted)" }
func (v CreateResult) GoString() string { return v.String() }
