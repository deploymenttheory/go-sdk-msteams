// Package onlinevoicemailpolicy models the compiled OnlineVoicemailPolicy
// resource. Its selected generated assignment protocol is a separate adapter.
package onlinevoicemailpolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Policy retains received fields and future data rather than synthesizing
// client defaults. Audio-file fields are strings, not local file uploads.
type Policy struct {
	Identity                            *string                    `json:"Identity"`
	Description                         *string                    `json:"Description"`
	ShareData                           *string                    `json:"ShareData"`
	PrimarySystemPromptLanguage         *string                    `json:"PrimarySystemPromptLanguage"`
	SecondarySystemPromptLanguage       *string                    `json:"SecondarySystemPromptLanguage"`
	PreambleAudioFile                   *string                    `json:"PreambleAudioFile"`
	PostambleAudioFile                  *string                    `json:"PostambleAudioFile"`
	EnableTranscription                 *bool                      `json:"EnableTranscription"`
	EnableTranscriptionProfanityMasking *bool                      `json:"EnableTranscriptionProfanityMasking"`
	EnableEditingCallAnswerRulesSetting *bool                      `json:"EnableEditingCallAnswerRulesSetting"`
	EnableTranscriptionTranslation      *bool                      `json:"EnableTranscriptionTranslation"`
	PreamblePostambleMandatory          *bool                      `json:"PreamblePostambleMandatory"`
	EnableVoicemailTriage               *bool                      `json:"EnableVoicemailTriage"`
	MaximumRecordingLength              *TimeSpan                  `json:"MaximumRecordingLength"`
	RawFields                           map[string]json.RawMessage `json:"-"`
}

func (v *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var value snapshot
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Policy(value)
	return nil
}
func (v Policy) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot Policy
	return json.Marshal(snapshot(v))
}
func (Policy) String() string     { return "online voicemail policy (redacted)" }
func (v Policy) GoString() string { return v.String() }
