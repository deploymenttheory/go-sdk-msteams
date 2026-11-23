package callqueues

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Overflow actions (OverflowAction), sent as integers on the wire.
const (
	OverflowDisconnectWithBusy int32 = 0
	OverflowForward            int32 = 1
	OverflowVoicemail          int32 = 2
	OverflowSharedVoicemail    int32 = 3
)

// Timeout actions (TimeoutAction), sent as integers on the wire.
const (
	TimeoutDisconnect      int32 = 0
	TimeoutForward         int32 = 1
	TimeoutVoicemail       int32 = 2
	TimeoutSharedVoicemail int32 = 3
)

// Routing methods (RoutingMethod), sent as integers on the wire.
const (
	RoutingAttendant   int32 = 0
	RoutingSerial      int32 = 1
	RoutingRoundRobin  int32 = 2
	RoutingLongestIdle int32 = 3
)

// No-agent actions (NoAgentAction), sent as integers on the wire.
const (
	NoAgentQueue           int32 = 0
	NoAgentDisconnect      int32 = 1
	NoAgentForward         int32 = 2
	NoAgentVoicemail       int32 = 3
	NoAgentSharedVoicemail int32 = 4
)

// No-agent scopes (NoAgentApplyTo), sent as integers on the wire.
const (
	NoAgentApplyToNewCalls int32 = 0
	NoAgentApplyToAllCalls int32 = 1
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("call queue input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected call queue response")

// Settings is the call queue create (POST) and update (PUT) body. Nil fields
// are omitted; a non-nil empty list is sent as []. The service replaces the
// queue on update, so pass a complete value (see CallQueue.Settings) rather
// than only the fields to change. Actions and RoutingMethod use the integer
// constants in this package.
type Settings struct {
	Name                                                 *string   `json:"Name,omitempty"`
	DistributionLists                                    *[]string `json:"DistributionLists,omitempty"`
	Users                                                *[]string `json:"Users,omitempty"`
	WelcomeMusicAudioFileID                              *string   `json:"WelcomeMusicAudioFileId,omitempty"`
	WelcomeTextToSpeechPrompt                            *string   `json:"WelcomeTextToSpeechPrompt,omitempty"`
	MusicOnHoldAudioFileID                               *string   `json:"MusicOnHoldAudioFileId,omitempty"`
	UseDefaultMusicOnHold                                *bool     `json:"UseDefaultMusicOnHold,omitempty"`
	AgentAlertTime                                       *int32    `json:"AgentAlertTime,omitempty"`
	AllowOptOut                                          *bool     `json:"AllowOptOut,omitempty"`
	ConferenceMode                                       *bool     `json:"ConferenceMode,omitempty"`
	PresenceAwareRouting                                 *bool     `json:"PresenceAwareRouting,omitempty"`
	RoutingMethod                                        *int32    `json:"RoutingMethod,omitempty"`
	LanguageID                                           *string   `json:"LanguageId,omitempty"`
	OverflowThreshold                                    *int32    `json:"OverflowThreshold,omitempty"`
	OverflowAction                                       *int32    `json:"OverflowAction,omitempty"`
	OverflowActionTarget                                 *string   `json:"OverflowActionTarget,omitempty"`
	OverflowSharedVoicemailTextToSpeechPrompt            *string   `json:"OverflowSharedVoicemailTextToSpeechPrompt,omitempty"`
	OverflowSharedVoicemailAudioFilePrompt               *string   `json:"OverflowSharedVoicemailAudioFilePrompt,omitempty"`
	EnableOverflowSharedVoicemailTranscription           *bool     `json:"EnableOverflowSharedVoicemailTranscription,omitempty"`
	EnableOverflowSharedVoicemailSystemPromptSuppression *bool     `json:"EnableOverflowSharedVoicemailSystemPromptSuppression,omitempty"`
	OverflowDisconnectAudioFilePrompt                    *string   `json:"OverflowDisconnectAudioFilePrompt,omitempty"`
	OverflowDisconnectTextToSpeechPrompt                 *string   `json:"OverflowDisconnectTextToSpeechPrompt,omitempty"`
	OverflowRedirectPersonAudioFilePrompt                *string   `json:"OverflowRedirectPersonAudioFilePrompt,omitempty"`
	OverflowRedirectPersonTextToSpeechPrompt             *string   `json:"OverflowRedirectPersonTextToSpeechPrompt,omitempty"`
	OverflowRedirectVoiceAppAudioFilePrompt              *string   `json:"OverflowRedirectVoiceAppAudioFilePrompt,omitempty"`
	OverflowRedirectVoiceAppTextToSpeechPrompt           *string   `json:"OverflowRedirectVoiceAppTextToSpeechPrompt,omitempty"`
	OverflowRedirectPhoneNumberAudioFilePrompt           *string   `json:"OverflowRedirectPhoneNumberAudioFilePrompt,omitempty"`
	OverflowRedirectPhoneNumberTextToSpeechPrompt        *string   `json:"OverflowRedirectPhoneNumberTextToSpeechPrompt,omitempty"`
	OverflowRedirectVoicemailAudioFilePrompt             *string   `json:"OverflowRedirectVoicemailAudioFilePrompt,omitempty"`
	OverflowRedirectVoicemailTextToSpeechPrompt          *string   `json:"OverflowRedirectVoicemailTextToSpeechPrompt,omitempty"`
	TimeoutThreshold                                     *int32    `json:"TimeoutThreshold,omitempty"`
	TimeoutAction                                        *int32    `json:"TimeoutAction,omitempty"`
	TimeoutActionTarget                                  *string   `json:"TimeoutActionTarget,omitempty"`
	TimeoutSharedVoicemailTextToSpeechPrompt             *string   `json:"TimeoutSharedVoicemailTextToSpeechPrompt,omitempty"`
	TimeoutSharedVoicemailAudioFilePrompt                *string   `json:"TimeoutSharedVoicemailAudioFilePrompt,omitempty"`
	EnableTimeoutSharedVoicemailTranscription            *bool     `json:"EnableTimeoutSharedVoicemailTranscription,omitempty"`
	EnableTimeoutSharedVoicemailSystemPromptSuppression  *bool     `json:"EnableTimeoutSharedVoicemailSystemPromptSuppression,omitempty"`
	TimeoutDisconnectAudioFilePrompt                     *string   `json:"TimeoutDisconnectAudioFilePrompt,omitempty"`
	TimeoutDisconnectTextToSpeechPrompt                  *string   `json:"TimeoutDisconnectTextToSpeechPrompt,omitempty"`
	TimeoutRedirectPersonAudioFilePrompt                 *string   `json:"TimeoutRedirectPersonAudioFilePrompt,omitempty"`
	TimeoutRedirectPersonTextToSpeechPrompt              *string   `json:"TimeoutRedirectPersonTextToSpeechPrompt,omitempty"`
	TimeoutRedirectVoiceAppAudioFilePrompt               *string   `json:"TimeoutRedirectVoiceAppAudioFilePrompt,omitempty"`
	TimeoutRedirectVoiceAppTextToSpeechPrompt            *string   `json:"TimeoutRedirectVoiceAppTextToSpeechPrompt,omitempty"`
	TimeoutRedirectPhoneNumberAudioFilePrompt            *string   `json:"TimeoutRedirectPhoneNumberAudioFilePrompt,omitempty"`
	TimeoutRedirectPhoneNumberTextToSpeechPrompt         *string   `json:"TimeoutRedirectPhoneNumberTextToSpeechPrompt,omitempty"`
	TimeoutRedirectVoicemailAudioFilePrompt              *string   `json:"TimeoutRedirectVoicemailAudioFilePrompt,omitempty"`
	TimeoutRedirectVoicemailTextToSpeechPrompt           *string   `json:"TimeoutRedirectVoicemailTextToSpeechPrompt,omitempty"`
	NoAgentAction                                        *int32    `json:"NoAgentAction,omitempty"`
	NoAgentActionTarget                                  *string   `json:"NoAgentActionTarget,omitempty"`
	NoAgentSharedVoicemailTextToSpeechPrompt             *string   `json:"NoAgentSharedVoicemailTextToSpeechPrompt,omitempty"`
	NoAgentSharedVoicemailAudioFilePrompt                *string   `json:"NoAgentSharedVoicemailAudioFilePrompt,omitempty"`
	EnableNoAgentSharedVoicemailTranscription            *bool     `json:"EnableNoAgentSharedVoicemailTranscription,omitempty"`
	EnableNoAgentSharedVoicemailSystemPromptSuppression  *bool     `json:"EnableNoAgentSharedVoicemailSystemPromptSuppression,omitempty"`
	NoAgentApplyTo                                       *int32    `json:"NoAgentApplyTo,omitempty"`
	NoAgentDisconnectAudioFilePrompt                     *string   `json:"NoAgentDisconnectAudioFilePrompt,omitempty"`
	NoAgentDisconnectTextToSpeechPrompt                  *string   `json:"NoAgentDisconnectTextToSpeechPrompt,omitempty"`
	NoAgentRedirectPersonAudioFilePrompt                 *string   `json:"NoAgentRedirectPersonAudioFilePrompt,omitempty"`
	NoAgentRedirectPersonTextToSpeechPrompt              *string   `json:"NoAgentRedirectPersonTextToSpeechPrompt,omitempty"`
	NoAgentRedirectVoiceAppAudioFilePrompt               *string   `json:"NoAgentRedirectVoiceAppAudioFilePrompt,omitempty"`
	NoAgentRedirectVoiceAppTextToSpeechPrompt            *string   `json:"NoAgentRedirectVoiceAppTextToSpeechPrompt,omitempty"`
	NoAgentRedirectPhoneNumberAudioFilePrompt            *string   `json:"NoAgentRedirectPhoneNumberAudioFilePrompt,omitempty"`
	NoAgentRedirectPhoneNumberTextToSpeechPrompt         *string   `json:"NoAgentRedirectPhoneNumberTextToSpeechPrompt,omitempty"`
	NoAgentRedirectVoicemailAudioFilePrompt              *string   `json:"NoAgentRedirectVoicemailAudioFilePrompt,omitempty"`
	NoAgentRedirectVoicemailTextToSpeechPrompt           *string   `json:"NoAgentRedirectVoicemailTextToSpeechPrompt,omitempty"`
	ThreadID                                             *string   `json:"ThreadId,omitempty"`
	ShiftsTeamID                                         *string   `json:"ShiftsTeamId,omitempty"`
	ShiftsSchedulingGroupID                              *string   `json:"ShiftsSchedulingGroupId,omitempty"`
	EnableResourceAccountsForOBO                         *bool     `json:"EnableResourceAccountsForOBO,omitempty"`
	ShouldOverwriteCallableChannelProperty               *bool     `json:"ShouldOverwriteCallableChannelProperty,omitempty"`
	OboResourceAccountIDs                                *[]string `json:"OboResourceAccountIds,omitempty"`
	AuthorizedUsers                                      *[]string `json:"AuthorizedUsers,omitempty"`
	HideAuthorizedUsers                                  *[]string `json:"HideAuthorizedUsers,omitempty"`
	OverflowActionCallPriority                           *int32    `json:"OverflowActionCallPriority,omitempty"`
	TimeoutActionCallPriority                            *int32    `json:"TimeoutActionCallPriority,omitempty"`
	NoAgentActionCallPriority                            *int32    `json:"NoAgentActionCallPriority,omitempty"`
	IsCallbackEnabled                                    *bool     `json:"IsCallbackEnabled,omitempty"`
	CallbackRequestDTMF                                  *string   `json:"CallbackRequestDtmf,omitempty"`
	WaitTimeBeforeOfferingCallbackInSecond               *int32    `json:"WaitTimeBeforeOfferingCallbackInSecond,omitempty"`
	NumberOfCallsInQueueBeforeOfferingCallback           *int32    `json:"NumberOfCallsInQueueBeforeOfferingCallback,omitempty"`
	CallToAgentRatioThresholdBeforeOfferingCallback      *int32    `json:"CallToAgentRatioThresholdBeforeOfferingCallback,omitempty"`
	CallbackOfferAudioFilePromptResourceID               *string   `json:"CallbackOfferAudioFilePromptResourceId,omitempty"`
	CallbackOfferTextToSpeechPrompt                      *string   `json:"CallbackOfferTextToSpeechPrompt,omitempty"`
	CallbackEmailNotificationTarget                      *string   `json:"CallbackEmailNotificationTarget,omitempty"`
	ServiceLevelThresholdResponseTimeInSecond            *int32    `json:"ServiceLevelThresholdResponseTimeInSecond,omitempty"`
	ComplianceRecordingForCallQueueIDs                   *[]string `json:"ComplianceRecordingForCallQueueIds,omitempty"`
	TextAnnouncementForCR                                *string   `json:"TextAnnouncementForCR,omitempty"`
	CustomAudioFileAnnouncementForCR                     *string   `json:"CustomAudioFileAnnouncementForCR,omitempty"`
	TextAnnouncementForCRFailure                         *string   `json:"TextAnnouncementForCRFailure,omitempty"`
	CustomAudioFileAnnouncementForCRFailure              *string   `json:"CustomAudioFileAnnouncementForCRFailure,omitempty"`
	SharedCallQueueHistoryID                             *string   `json:"SharedCallQueueHistoryId,omitempty"`
	AutoRecordingTemplateID                              *string   `json:"AutoRecordingTemplateId,omitempty"`
	SharedVoicemailTriageSettingsID                      *string   `json:"SharedVoicemailTriageSettingsId,omitempty"`
}

// CallQueue is the service's call queue object. Typed fields cover common
// properties; RawFields keeps every property, including null and unknown ones.
type CallQueue struct {
	Identity              *string                    `json:"Identity"`
	Name                  *string                    `json:"Name"`
	TenantID              *string                    `json:"TenantId"`
	Description           *string                    `json:"Description"`
	RoutingMethod         *int32                     `json:"RoutingMethod"`
	AgentAlertTime        *int32                     `json:"AgentAlertTime"`
	AllowOptOut           *bool                      `json:"AllowOptOut"`
	ConferenceMode        *bool                      `json:"ConferenceMode"`
	PresenceAwareRouting  *bool                      `json:"PresenceAwareRouting"`
	LanguageID            *string                    `json:"LanguageId"`
	UseDefaultMusicOnHold *bool                      `json:"UseDefaultMusicOnHold"`
	OverflowThreshold     *int32                     `json:"OverflowThreshold"`
	OverflowAction        *int32                     `json:"OverflowAction"`
	TimeoutThreshold      *int32                     `json:"TimeoutThreshold"`
	TimeoutAction         *int32                     `json:"TimeoutAction"`
	NoAgentAction         *int32                     `json:"NoAgentAction"`
	DistributionLists     []string                   `json:"DistributionLists"`
	Users                 []string                   `json:"Users"`
	ApplicationInstances  []string                   `json:"ApplicationInstances"`
	RawFields             map[string]json.RawMessage `json:"-"`
}

func (v *CallQueue) UnmarshalJSON(raw []byte) error {
	type plain CallQueue
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = CallQueue(value)
	return nil
}

// settingsSources maps a Settings property to the read property it is copied
// from when the names differ, mirroring Set-CsCallQueue's backfill.
var settingsSources = map[string]string{
	"WelcomeMusicAudioFileId": "WelcomeMusicResourceId",
	"MusicOnHoldAudioFileId":  "MusicOnHoldResourceId",
	"OboResourceAccountIds":   "OboResourceAccounts",
}

// targetID reduces an action target to its identifier: the service may return
// a plain string or an object carrying Id.
func targetID(raw json.RawMessage) (json.RawMessage, bool) {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return raw, strings.TrimSpace(text) != ""
	}
	var object struct {
		ID *string `json:"Id"`
	}
	if json.Unmarshal(raw, &object) != nil || object.ID == nil || strings.TrimSpace(*object.ID) == "" {
		return nil, false
	}
	value, err := json.Marshal(*object.ID)
	return value, err == nil
}

// idList reduces a non-empty list of strings, or of objects carrying ObjectId
// or Id, to a list of strings.
func idList(raw json.RawMessage) (json.RawMessage, bool) {
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil || len(items) == 0 {
		return nil, false
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		var text string
		if json.Unmarshal(item, &text) == nil {
			ids = append(ids, text)
			continue
		}
		var object struct {
			ObjectID *string `json:"ObjectId"`
			ID       *string `json:"Id"`
		}
		switch {
		case json.Unmarshal(item, &object) != nil:
			return nil, false
		case object.ObjectID != nil:
			ids = append(ids, *object.ObjectID)
		case object.ID != nil:
			ids = append(ids, *object.ID)
		default:
			return nil, false
		}
	}
	value, err := json.Marshal(ids)
	return value, err == nil
}

// Settings returns the update body Set-CsCallQueue sends for this queue when
// no parameter is changed: every writable property present on the read queue,
// skipping null values, blank strings and empty lists. Resource IDs and action
// targets are reduced to their identifiers, and a property whose read value
// does not fit the write type is skipped. Change the result and pass it to
// UpdateByID.
func (v *CallQueue) Settings() Settings {
	var settings Settings
	if v == nil {
		return settings
	}
	target := reflect.ValueOf(&settings).Elem()
	for i := range target.NumField() {
		name, _, _ := strings.Cut(target.Type().Field(i).Tag.Get("json"), ",")
		source := name
		if alias, ok := settingsSources[name]; ok {
			source = alias
		}
		raw, ok := v.RawFields[source]
		trimmed := bytes.TrimSpace(raw)
		if !ok || bytes.Equal(trimmed, []byte("null")) {
			continue
		}
		switch {
		case strings.HasSuffix(name, "ActionTarget") || name == "CallbackEmailNotificationTarget":
			raw, ok = targetID(raw)
		case bytes.HasPrefix(trimmed, []byte("[")):
			raw, ok = idList(raw)
		case bytes.HasPrefix(trimmed, []byte(`"`)):
			var text string
			ok = json.Unmarshal(raw, &text) == nil && strings.TrimSpace(text) != ""
		}
		if !ok {
			continue
		}
		field := reflect.New(target.Field(i).Type())
		if json.Unmarshal(raw, field.Interface()) == nil {
			target.Field(i).Set(field.Elem())
		}
	}
	return settings
}

// ListOptions selects Get-CsCallQueue list parameters. Nil fields are omitted;
// non-nil booleans are sent as True/False.
type ListOptions struct {
	First          *int32
	Skip           *int32
	ExcludeContent *bool
	NameFilter     *string
	SortBy         *string
	Descending     *bool
}

// WriteOptions carries the optional channelUserObjectId query used with a
// channel-connected queue (ThreadId). Nil options send no query.
type WriteOptions struct {
	ChannelUserObjectID *string
}
