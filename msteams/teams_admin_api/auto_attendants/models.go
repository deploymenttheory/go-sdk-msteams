package autoattendants

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Prompt types (Prompt.ActiveType).
const (
	PromptNone         = "None"
	PromptTextToSpeech = "TextToSpeech"
	PromptAudioFile    = "AudioFile"
)

// Menu option actions (MenuOption.Action).
const (
	ActionTransferCallToOperator = "TransferCallToOperator"
	ActionDisconnectCall         = "DisconnectCall"
	ActionTransferCallToTarget   = "TransferCallToTarget"
	ActionAnnouncement           = "Announcement"
	ActionMainlineAttendantFlow  = "MainlineAttendantFlow"
	ActionAgentsAndQueues        = "AgentsAndQueues"
)

// DTMF responses (MenuOption.DtmfResponse). Automatic is the menu's only
// option and runs without caller input.
const (
	DtmfAutomatic = "Automatic"
	DtmfTone0     = "Tone0"
	DtmfTone1     = "Tone1"
	DtmfTone2     = "Tone2"
	DtmfTone3     = "Tone3"
	DtmfTone4     = "Tone4"
	DtmfTone5     = "Tone5"
	DtmfTone6     = "Tone6"
	DtmfTone7     = "Tone7"
	DtmfTone8     = "Tone8"
	DtmfTone9     = "Tone9"
	DtmfToneStar  = "ToneStar"
	DtmfTonePound = "TonePound"
)

// Callable entity types (CallableEntity.Type).
const (
	EntityUser                        = "User"
	EntityOrganizationalAutoAttendant = "OrganizationalAutoAttendant"
	EntityHuntGroup                   = "HuntGroup"
	EntityApplicationEndpoint         = "ApplicationEndpoint"
	EntityExternalPstn                = "ExternalPstn"
	EntitySharedVoicemail             = "SharedVoicemail"
	EntityConfigurationEndpoint       = "ConfigurationEndpoint"
	EntityVoicemail                   = "Voicemail"
)

// Call handling association types (CallHandlingAssociation.Type).
const (
	AssociationAfterHours = "AfterHours"
	AssociationHoliday    = "Holiday"
)

// Directory search methods (Menu.DirectorySearchMethod).
const (
	SearchNone        = "None"
	SearchByName      = "ByName"
	SearchByExtension = "ByExtension"
)

// Dial scope types (DialScope.Type). The service reports an unset scope as
// "Default"; New-CsAutoAttendantDialScope creates Group scopes.
const (
	DialScopeDefault = "Default"
	DialScopeGroup   = "Group"
)

// Timeout disconnect and AI disclaimer prompt types (CallFlow).
const (
	PromptKindDefault = "Default"
	PromptKindCustom  = "Custom"
	PromptKindNone    = "None"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("auto attendant input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected auto attendant response")

// AudioFile references an uploaded audio file (see the audio prompts service).
type AudioFile struct {
	ID          *string `json:"Id,omitempty"`
	FileName    *string `json:"FileName,omitempty"`
	DownloadURI *string `json:"DownloadUri,omitempty"`
}

// Prompt is a greeting or menu prompt (New-CsAutoAttendantPrompt). Set
// ActiveType and the matching text or audio file.
type Prompt struct {
	ActiveType         *string    `json:"ActiveType,omitempty"`
	TextToSpeechPrompt *string    `json:"TextToSpeechPrompt,omitempty"`
	AudioFilePrompt    *AudioFile `json:"AudioFilePrompt,omitempty"`
}

// CallableEntity is a transfer or operator target
// (New-CsAutoAttendantCallableEntity). DisplayName and
// ConfigurationEndpointType are read-only; the service ignores them on write.
type CallableEntity struct {
	ID                                           *string `json:"Id,omitempty"`
	Type                                         *string `json:"Type,omitempty"`
	DisplayName                                  *string `json:"DisplayName,omitempty"`
	EnableTranscription                          *bool   `json:"EnableTranscription,omitempty"`
	EnableSharedVoicemailSystemPromptSuppression *bool   `json:"EnableSharedVoicemailSystemPromptSuppression,omitempty"`
	CallPriority                                 *int32  `json:"CallPriority,omitempty"`
	SharedVoicemailHistoryTemplateID             *string `json:"SharedVoicemailHistoryTemplateId,omitempty"`
	ConfigurationEndpointType                    *string `json:"ConfigurationEndpointType,omitempty"`
}

// MenuOption is one menu choice (New-CsAutoAttendantMenuOption).
type MenuOption struct {
	Action                   *string         `json:"Action,omitempty"`
	DtmfResponse             *string         `json:"DtmfResponse,omitempty"`
	VoiceResponses           []string        `json:"VoiceResponses,omitempty"`
	CallTarget               *CallableEntity `json:"CallTarget,omitempty"`
	Prompt                   *Prompt         `json:"Prompt,omitempty"`
	Description              *string         `json:"Description,omitempty"`
	MainlineAttendantTarget  *string         `json:"MainlineAttendantTarget,omitempty"`
	AgentTargetType          *int32          `json:"AgentTargetType,omitempty"`
	AgentTarget              *string         `json:"AgentTarget,omitempty"`
	AgentTargetTagTemplateID *string         `json:"AgentTargetTagTemplateId,omitempty"`
}

// Menu is a call flow menu (New-CsAutoAttendantMenu).
type Menu struct {
	Name                  *string      `json:"Name,omitempty"`
	Prompts               []Prompt     `json:"Prompts,omitempty"`
	MenuOptions           []MenuOption `json:"MenuOptions,omitempty"`
	DialByNameEnabled     *bool        `json:"DialByNameEnabled,omitempty"`
	DirectorySearchMethod *string      `json:"DirectorySearchMethod,omitempty"`
}

// CallFlow is a greeting plus menu (New-CsAutoAttendantCallFlow). The ID
// links a call handling association to the flow; DraftCallFlow assigns one,
// and any unique GUID chosen by the caller works the same way.
type CallFlow struct {
	ID                                *string  `json:"Id,omitempty"`
	Name                              *string  `json:"Name,omitempty"`
	Greetings                         []Prompt `json:"Greetings,omitempty"`
	Menu                              *Menu    `json:"Menu,omitempty"`
	ForceListenMenuEnabled            *bool    `json:"ForceListenMenuEnabled,omitempty"`
	RingResourceAccountDelegates      *bool    `json:"RingResourceAccountDelegates,omitempty"`
	TimeoutThreshold                  *int32   `json:"TimeoutThreshold,omitempty"`
	TimeoutDisconnectPromptType       *string  `json:"TimeoutDisconnectPromptType,omitempty"`
	TimeoutDisconnectPromptCustomText *string  `json:"TimeoutDisconnectPromptCustomText,omitempty"`
	AiDisclaimerType                  *string  `json:"AiDisclaimerType,omitempty"`
	AiDisclaimerCustomText            *string  `json:"AiDisclaimerCustomText,omitempty"`
}

// CallHandlingAssociation routes a schedule to a call flow
// (New-CsAutoAttendantCallHandlingAssociation). Holiday associations cannot
// be disabled.
type CallHandlingAssociation struct {
	Type       *string `json:"Type,omitempty"`
	ScheduleID *string `json:"ScheduleId,omitempty"`
	CallFlowID *string `json:"CallFlowId,omitempty"`
	Priority   *int32  `json:"Priority,omitempty"`
	Enabled    *bool   `json:"Enabled,omitempty"`
}

// GroupDialScope lists the groups of a group dial scope.
type GroupDialScope struct {
	GroupIDs []string `json:"GroupIds,omitempty"`
}

// DialScope limits dial-by-name (New-CsAutoAttendantDialScope).
type DialScope struct {
	Type           *string         `json:"Type,omitempty"`
	GroupDialScope *GroupDialScope `json:"GroupDialScope,omitempty"`
}

// DirectoryLookupScope holds the dial-by-name inclusion and exclusion scopes.
type DirectoryLookupScope struct {
	InclusionScope *DialScope `json:"InclusionScope,omitempty"`
	ExclusionScope *DialScope `json:"ExclusionScope,omitempty"`
}

// StatusRecord is one provisioning status entry. Status and Type are kept as
// the service sent them; RawFields keeps every property.
type StatusRecord struct {
	ID              *string                    `json:"Id"`
	ErrorCode       *string                    `json:"ErrorCode"`
	Message         *string                    `json:"Message"`
	StatusTimestamp *string                    `json:"StatusTimestamp"`
	AuxiliaryData   []string                   `json:"AuxiliaryData"`
	Status          json.RawMessage            `json:"Status"`
	Type            json.RawMessage            `json:"Type"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *StatusRecord) UnmarshalJSON(raw []byte) error {
	type plain StatusRecord
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = StatusRecord(value)
	return nil
}

// AutoAttendant is the service's auto attendant object. Schedules are kept
// as sent (see the schedules service for their shape); RawFields keeps every
// property, including null and unknown ones.
type AutoAttendant struct {
	ID                              *string                    `json:"Id"`
	Name                            *string                    `json:"Name"`
	TenantID                        *string                    `json:"TenantId"`
	LanguageID                      *string                    `json:"LanguageId"`
	VoiceID                         *string                    `json:"VoiceId"`
	TimeZoneID                      *string                    `json:"TimeZoneId"`
	DefaultCallFlow                 *CallFlow                  `json:"DefaultCallFlow"`
	CallFlows                       []CallFlow                 `json:"CallFlows"`
	Schedules                       []json.RawMessage          `json:"Schedules"`
	CallHandlingAssociations        []CallHandlingAssociation  `json:"CallHandlingAssociations"`
	VoiceResponseEnabled            *bool                      `json:"VoiceResponseEnabled"`
	DialByNameResourceID            *string                    `json:"DialByNameResourceId"`
	DirectoryLookupScope            *DirectoryLookupScope      `json:"DirectoryLookupScope"`
	Operator                        *CallableEntity            `json:"Operator"`
	OperatorSharedVoicemail         *CallableEntity            `json:"OperatorSharedVoicemail"`
	ApplicationInstances            []string                   `json:"ApplicationInstances"`
	AuthorizedUsers                 []string                   `json:"AuthorizedUsers"`
	HideAuthorizedUsers             []string                   `json:"HideAuthorizedUsers"`
	Status                          []*StatusRecord            `json:"Status"`
	UserNameExtension               *string                    `json:"UserNameExtension"`
	MainlineAttendantEnabled        *bool                      `json:"MainlineAttendantEnabled"`
	MainlineAttendantAgentVoiceID   *string                    `json:"MainlineAttendantAgentVoiceId"`
	AutoRecordingTemplateID         *string                    `json:"AutoRecordingTemplateId"`
	SharedVoicemailTriageSettingsID *string                    `json:"SharedVoicemailTriageSettingsId"`
	SpamDetectionTemplateID         *string                    `json:"SpamDetectionTemplateId"`
	RawFields                       map[string]json.RawMessage `json:"-"`
}

func (v *AutoAttendant) UnmarshalJSON(raw []byte) error {
	type plain AutoAttendant
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = AutoAttendant(value)
	return nil
}

// Settings is the auto attendant create (POST) and update (PUT) body. Nil
// fields are omitted; a non-nil empty list is sent as []. The service
// replaces the attendant on update, so pass a complete value (see
// AutoAttendant.Settings) rather than only the fields to change. Resource
// account associations are managed with the resource accounts service, not
// through ApplicationInstances.
type Settings struct {
	ID                              *string                    `json:"Id,omitempty"`
	Name                            *string                    `json:"Name,omitempty"`
	TenantID                        *string                    `json:"TenantId,omitempty"`
	LanguageID                      *string                    `json:"LanguageId,omitempty"`
	VoiceID                         *string                    `json:"VoiceId,omitempty"`
	TimeZoneID                      *string                    `json:"TimeZoneId,omitempty"`
	DefaultCallFlow                 *CallFlow                  `json:"DefaultCallFlow,omitempty"`
	CallFlows                       *[]CallFlow                `json:"CallFlows,omitempty"`
	Schedules                       *[]json.RawMessage         `json:"Schedules,omitempty"`
	CallHandlingAssociations        *[]CallHandlingAssociation `json:"CallHandlingAssociations,omitempty"`
	VoiceResponseEnabled            *bool                      `json:"VoiceResponseEnabled,omitempty"`
	DialByNameResourceID            *string                    `json:"DialByNameResourceId,omitempty"`
	DirectoryLookupScope            *DirectoryLookupScope      `json:"DirectoryLookupScope,omitempty"`
	Operator                        *CallableEntity            `json:"Operator,omitempty"`
	OperatorSharedVoicemail         *CallableEntity            `json:"OperatorSharedVoicemail,omitempty"`
	ApplicationInstances            *[]string                  `json:"ApplicationInstances,omitempty"`
	AuthorizedUsers                 *[]string                  `json:"AuthorizedUsers,omitempty"`
	HideAuthorizedUsers             *[]string                  `json:"HideAuthorizedUsers,omitempty"`
	UserNameExtension               *string                    `json:"UserNameExtension,omitempty"`
	MainlineAttendantEnabled        *bool                      `json:"MainlineAttendantEnabled,omitempty"`
	MainlineAttendantAgentVoiceID   *string                    `json:"MainlineAttendantAgentVoiceId,omitempty"`
	AutoRecordingTemplateID         *string                    `json:"AutoRecordingTemplateId,omitempty"`
	SharedVoicemailTriageSettingsID *string                    `json:"SharedVoicemailTriageSettingsId,omitempty"`
	SpamDetectionTemplateID         *string                    `json:"SpamDetectionTemplateId,omitempty"`
}

// listOf returns a pointer to a copy of a non-empty list, else nil.
func listOf[T any](values []T) *[]T {
	if len(values) == 0 {
		return nil
	}
	out := append([]T(nil), values...)
	return &out
}

// Settings returns the update body Set-CsAutoAttendant sends for this
// attendant when nothing is changed: every writable property present on the
// read value, skipping nulls and empty lists. ApplicationInstances is copied
// as read, as the module does. Change the result and pass it to UpdateByID.
func (v *AutoAttendant) Settings() Settings {
	if v == nil {
		return Settings{}
	}
	return Settings{
		ID: v.ID, Name: v.Name, TenantID: v.TenantID, LanguageID: v.LanguageID, VoiceID: v.VoiceID, TimeZoneID: v.TimeZoneID,
		DefaultCallFlow: v.DefaultCallFlow, CallFlows: listOf(v.CallFlows), Schedules: listOf(v.Schedules),
		CallHandlingAssociations: listOf(v.CallHandlingAssociations), VoiceResponseEnabled: v.VoiceResponseEnabled,
		DialByNameResourceID: v.DialByNameResourceID, DirectoryLookupScope: v.DirectoryLookupScope, Operator: v.Operator,
		OperatorSharedVoicemail: v.OperatorSharedVoicemail, ApplicationInstances: listOf(v.ApplicationInstances),
		AuthorizedUsers: listOf(v.AuthorizedUsers), HideAuthorizedUsers: listOf(v.HideAuthorizedUsers),
		UserNameExtension: v.UserNameExtension, MainlineAttendantEnabled: v.MainlineAttendantEnabled,
		MainlineAttendantAgentVoiceID: v.MainlineAttendantAgentVoiceID, AutoRecordingTemplateID: v.AutoRecordingTemplateID,
		SharedVoicemailTriageSettingsID: v.SharedVoicemailTriageSettingsID, SpamDetectionTemplateID: v.SpamDetectionTemplateID,
	}
}

// ListOptions selects Get-CsAutoAttendant list parameters. Nil fields are
// omitted; non-nil booleans are sent as True/False.
type ListOptions struct {
	First          *int32
	Skip           *int32
	ExcludeContent *bool
	NameFilter     *string
	SortBy         *string
	Descending     *bool
	IncludeStatus  *bool
}

// StatusOptions selects Get-CsAutoAttendantStatus resources (for example
// AudioFile or DialByNameVoiceResponses), sent comma-joined.
type StatusOptions struct {
	IncludeResources []string
}

// Voice is a text-to-speech voice of a language.
type Voice struct {
	ID   *string `json:"Id"`
	Name *string `json:"Name"`
}

// Language is a supported auto attendant language.
type Language struct {
	ID                     *string `json:"Id"`
	DisplayName            *string `json:"DisplayName"`
	Voices                 []Voice `json:"Voices"`
	VoiceResponseSupported *bool   `json:"VoiceResponseSupported"`
}

// TimeZone is a supported auto attendant time zone.
type TimeZone struct {
	ID          *string `json:"Id"`
	DisplayName *string `json:"DisplayName"`
}

// TenantInformation is the tenant's auto attendant defaults and supported
// languages and time zones. Get-CsAutoAttendantTenantInformation shows
// DefaultLanguage, DefaultTimeZone and EnabledFeatures as DefaultLanguageId,
// DefaultTimeZoneId and FlightedFeatures. RawFields keeps every property.
type TenantInformation struct {
	DefaultLanguage    *string                    `json:"DefaultLanguage"`
	DefaultTimeZone    *string                    `json:"DefaultTimeZone"`
	SupportedLanguages []*Language                `json:"SupportedLanguages"`
	SupportedTimeZones []*TimeZone                `json:"SupportedTimeZones"`
	EnabledFeatures    []string                   `json:"EnabledFeatures"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *TenantInformation) UnmarshalJSON(raw []byte) error {
	type plain TenantInformation
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = TenantInformation(value)
	return nil
}
