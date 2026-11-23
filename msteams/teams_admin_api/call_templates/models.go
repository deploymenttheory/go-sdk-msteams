package calltemplates

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("call template input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected call template response")

func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// AgentViewPermission values for auto recording templates (wire integers).
const (
	AgentViewPermissionNone int32 = 0
	AgentViewPermissionAll  int32 = 1
)

// AutoRecordingTemplate is a call queue auto recording and transcription
// template. Nil fields are absent or null on read and omitted on write.
// AutoRecordingAnnouncementAudioFileName is returned by the service only.
type AutoRecordingTemplate struct {
	ID                                          *string                    `json:"Id"`
	Name                                        *string                    `json:"Name"`
	Description                                 *string                    `json:"Description"`
	TranscriptionEnabled                        *bool                      `json:"TranscriptionEnabled"`
	RecordingEnabled                            *bool                      `json:"RecordingEnabled"`
	AgentViewPermission                         *int32                     `json:"AgentViewPermission"`
	SharePointHostName                          *string                    `json:"SharePointHostName"`
	SharePointSiteName                          *string                    `json:"SharePointSiteName"`
	RecordingDocumentOwner                      *string                    `json:"RecordingDocumentOwner"`
	AutoRecordingAnnouncementAudioFileID        *string                    `json:"AutoRecordingAnnouncementAudioFileId"`
	AutoRecordingAnnouncementAudioFileName      *string                    `json:"AutoRecordingAnnouncementAudioFileName"`
	AutoRecordingAnnouncementTextToSpeechPrompt *string                    `json:"AutoRecordingAnnouncementTextToSpeechPrompt"`
	RawFields                                   map[string]json.RawMessage `json:"-"`
}

func (v *AutoRecordingTemplate) UnmarshalJSON(raw []byte) error {
	type plain AutoRecordingTemplate
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = AutoRecordingTemplate(value)
	return nil
}

func (v *AutoRecordingTemplate) identity() *string { return v.ID }

func (v *AutoRecordingTemplate) body() map[string]any {
	return fields(map[string]any{"Name": v.Name, "Description": v.Description, "TranscriptionEnabled": v.TranscriptionEnabled,
		"RecordingEnabled": v.RecordingEnabled, "AgentViewPermission": v.AgentViewPermission, "SharePointHostName": v.SharePointHostName,
		"SharePointSiteName": v.SharePointSiteName, "RecordingDocumentOwner": v.RecordingDocumentOwner,
		"AutoRecordingAnnouncementAudioFileId": v.AutoRecordingAnnouncementAudioFileID, "AutoRecordingAnnouncementTextToSpeechPrompt": v.AutoRecordingAnnouncementTextToSpeechPrompt})
}

// ComplianceRecordingTemplate is a compliance recording for call queue
// template. BotID is the recording bot's resource account object ID. Detail
// reads return the Bot-prefixed service names (BotRequiredDuringCall, …);
// they are decoded into the same fields.
type ComplianceRecordingTemplate struct {
	ID                        *string                    `json:"Id"`
	Name                      *string                    `json:"Name"`
	Description               *string                    `json:"Description"`
	BotID                     *string                    `json:"BotId"`
	RequiredDuringCall        *bool                      `json:"RequiredDuringCall"`
	RequiredBeforeCall        *bool                      `json:"RequiredBeforeCall"`
	ConcurrentInvitationCount *int32                     `json:"ConcurrentInvitationCount"`
	PairedApplication         *string                    `json:"PairedApplication"`
	RawFields                 map[string]json.RawMessage `json:"-"`
}

func (v *ComplianceRecordingTemplate) UnmarshalJSON(raw []byte) error {
	type plain ComplianceRecordingTemplate
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	var domain struct {
		RequiredDuringCall        *bool  `json:"BotRequiredDuringCall"`
		RequiredBeforeCall        *bool  `json:"BotRequiredBeforeCall"`
		ConcurrentInvitationCount *int32 `json:"BotConcurrentInvitationCount"`
	}
	if json.Unmarshal(raw, &domain) != nil {
		return client.ErrDecode
	}
	if value.RequiredDuringCall == nil {
		value.RequiredDuringCall = domain.RequiredDuringCall
	}
	if value.RequiredBeforeCall == nil {
		value.RequiredBeforeCall = domain.RequiredBeforeCall
	}
	if value.ConcurrentInvitationCount == nil {
		value.ConcurrentInvitationCount = domain.ConcurrentInvitationCount
	}
	*v = ComplianceRecordingTemplate(value)
	return nil
}

func (v *ComplianceRecordingTemplate) identity() *string { return v.ID }

func (v *ComplianceRecordingTemplate) body() map[string]any {
	return fields(map[string]any{"Name": v.Name, "Description": v.Description, "BotId": v.BotID, "RequiredDuringCall": v.RequiredDuringCall,
		"RequiredBeforeCall": v.RequiredBeforeCall, "ConcurrentInvitationCount": v.ConcurrentInvitationCount, "PairedApplication": v.PairedApplication})
}

// Shared call history visibility values (wire integers).
const (
	HistoryNone                           int32 = 0
	HistoryAuthorizedUsersOnly            int32 = 1
	HistoryAuthorizedUsersAndAgents       int32 = 2 // IncomingMissedCalls, AnsweredAndOutboundCalls
	HistoryAuthorizedUsersAndGroupMembers int32 = 2 // IncomingRedirectedCalls
)

// SharedCallHistoryTemplate controls who sees a call queue's shared call
// history. Get/New/Set/Remove-CsSharedCallQueueHistoryTemplate are deprecated
// aliases of the same template.
type SharedCallHistoryTemplate struct {
	ID                       *string                    `json:"Id"`
	Name                     *string                    `json:"Name"`
	Description              *string                    `json:"Description"`
	IncomingMissedCalls      *int32                     `json:"IncomingMissedCalls"`
	AnsweredAndOutboundCalls *int32                     `json:"AnsweredAndOutboundCalls"`
	IncomingRedirectedCalls  *int32                     `json:"IncomingRedirectedCalls"`
	RawFields                map[string]json.RawMessage `json:"-"`
}

func (v *SharedCallHistoryTemplate) UnmarshalJSON(raw []byte) error {
	type plain SharedCallHistoryTemplate
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = SharedCallHistoryTemplate(value)
	return nil
}

func (v *SharedCallHistoryTemplate) identity() *string { return v.ID }

func (v *SharedCallHistoryTemplate) body() map[string]any {
	return fields(map[string]any{"Name": v.Name, "Description": v.Description, "IncomingMissedCalls": v.IncomingMissedCalls,
		"AnsweredAndOutboundCalls": v.AnsweredAndOutboundCalls, "IncomingRedirectedCalls": v.IncomingRedirectedCalls})
}

// SharedVoicemailTriageSettingsTemplate selects the AI triage features for
// shared voicemail. On create the service enables omitted features; an update
// replaces the template, so omitted features become false.
type SharedVoicemailTriageSettingsTemplate struct {
	ID                          *string                    `json:"Id"`
	Name                        *string                    `json:"Name"`
	Description                 *string                    `json:"Description"`
	EnableUrgencyDetection      *bool                      `json:"EnableUrgencyDetection"`
	EnableCategoryDetection     *bool                      `json:"EnableCategoryDetection"`
	EnableCallToActionDetection *bool                      `json:"EnableCallToActionDetection"`
	EnableVoiceToTextSummary    *bool                      `json:"EnableVoiceToTextSummary"`
	RawFields                   map[string]json.RawMessage `json:"-"`
}

func (v *SharedVoicemailTriageSettingsTemplate) UnmarshalJSON(raw []byte) error {
	type plain SharedVoicemailTriageSettingsTemplate
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = SharedVoicemailTriageSettingsTemplate(value)
	return nil
}

func (v *SharedVoicemailTriageSettingsTemplate) identity() *string { return v.ID }

func (v *SharedVoicemailTriageSettingsTemplate) body() map[string]any {
	return fields(map[string]any{"Name": v.Name, "Description": v.Description, "EnableUrgencyDetection": v.EnableUrgencyDetection,
		"EnableCategoryDetection": v.EnableCategoryDetection, "EnableCallToActionDetection": v.EnableCallToActionDetection,
		"EnableVoiceToTextSummary": v.EnableVoiceToTextSummary})
}

// TagTarget is the callable entity a tag routes to (the module's
// New-CsAutoAttendantCallableEntity value). Type is the service's entity type
// name, for example "User", "ApplicationEndpoint" or "ExternalPstn".
type TagTarget struct {
	ID                                           *string                    `json:"Id"`
	Type                                         *string                    `json:"Type"`
	EnableTranscription                          *bool                      `json:"EnableTranscription"`
	EnableSharedVoicemailSystemPromptSuppression *bool                      `json:"EnableSharedVoicemailSystemPromptSuppression"`
	CallPriority                                 *int32                     `json:"CallPriority"`
	SharedVoicemailHistoryTemplateID             *string                    `json:"SharedVoicemailHistoryTemplateId"`
	RawFields                                    map[string]json.RawMessage `json:"-"`
}

func (v *TagTarget) UnmarshalJSON(raw []byte) error {
	type plain TagTarget
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = TagTarget(value)
	return nil
}

func (v *TagTarget) body() map[string]any {
	return fields(map[string]any{"Id": v.ID, "Type": v.Type, "EnableTranscription": v.EnableTranscription,
		"EnableSharedVoicemailSystemPromptSuppression": v.EnableSharedVoicemailSystemPromptSuppression,
		"CallPriority": v.CallPriority, "SharedVoicemailHistoryTemplateId": v.SharedVoicemailHistoryTemplateID})
}

// Tag is one IVR tag: a name and the target it routes to.
type Tag struct {
	TagName    *string    `json:"TagName"`
	TagDetails *TagTarget `json:"TagDetails"`
}

// TagsTemplate is an IVR tags template.
type TagsTemplate struct {
	ID          *string                    `json:"Id"`
	Name        *string                    `json:"Name"`
	Description *string                    `json:"Description"`
	Tags        []Tag                      `json:"Tags"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *TagsTemplate) UnmarshalJSON(raw []byte) error {
	type plain TagsTemplate
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = TagsTemplate(value)
	return nil
}

func (v *TagsTemplate) identity() *string { return v.ID }

func (v *TagsTemplate) body() map[string]any {
	body := fields(map[string]any{"Name": v.Name, "Description": v.Description})
	if v.Tags != nil {
		tags := make([]map[string]any, 0, len(v.Tags))
		for _, tag := range v.Tags {
			item := fields(map[string]any{"TagName": tag.TagName})
			if tag.TagDetails != nil {
				item["TagDetails"] = tag.TagDetails.body()
			}
			tags = append(tags, item)
		}
		body["Tags"] = tags
	}
	return body
}

// fields drops nil pointer values so omitted request fields stay off the wire.
func fields(values map[string]any) map[string]any {
	out := map[string]any{}
	for name, value := range values {
		switch typed := value.(type) {
		case *string:
			if typed != nil {
				out[name] = *typed
			}
		case *bool:
			if typed != nil {
				out[name] = *typed
			}
		case *int32:
			if typed != nil {
				out[name] = *typed
			}
		}
	}
	return out
}
