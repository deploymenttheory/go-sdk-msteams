package mainlineattendants

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("mainline attendant input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected mainline attendant response")

// Flow types reported in Flow.Type and accepted by FlowListOptions.Type.
const (
	FlowTypeAppointmentBooking = "AppointmentBooking"
	FlowTypeQuestionAnswer     = "QuestionAnswer"
)

// Appointment booking platforms. With MicrosoftBookings the service ignores
// the caller authentication method, API authentication type and API definitions.
const (
	BookingPlatformCustom            = "Custom"
	BookingPlatformMicrosoftBookings = "MicrosoftBookings"
)

// Spam detection actions. They travel as strings.
const (
	SpamActionDisconnectCall         = "DisconnectCall"
	SpamActionTransferCallToOperator = "TransferCallToOperator"
	SpamActionTransferCallToTarget   = "TransferCallToTarget"
)

// Call target types for which the module sends CallTarget.CallPriority.
const (
	CallTargetApplicationEndpoint   = "ApplicationEndpoint"
	CallTargetConfigurationEndpoint = "ConfigurationEndpoint"
)

// CallerAuthenticationMethod is the module enum; it travels as its integer value.
type CallerAuthenticationMethod int32

const (
	CallerAuthenticationSms CallerAuthenticationMethod = iota
	CallerAuthenticationEmail
	CallerAuthenticationVerificationLink
	CallerAuthenticationVoiceprint
	CallerAuthenticationUserDetails
)

// APIAuthenticationType is the module enum; it travels as its integer value.
type APIAuthenticationType int32

const (
	APIAuthenticationBasic APIAuthenticationType = iota
	APIAuthenticationAPIKey
	APIAuthenticationBearerTokenStatic
	APIAuthenticationBearerTokenDynamic
)

// decodeFields unmarshals raw into value (a pointer to a plain alias type) and
// keeps every service field in fields, so unknown properties survive.
func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// RelatedConfiguration names a configuration (such as an auto attendant) that
// uses a flow or template. The service sets these; the module sends them back
// unchanged on Set.
type RelatedConfiguration struct {
	ID        *string                    `json:"Id"`
	Name      *string                    `json:"Name"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *RelatedConfiguration) UnmarshalJSON(raw []byte) error {
	type plain RelatedConfiguration
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = RelatedConfiguration(value)
	return nil
}

func (v RelatedConfiguration) wire() map[string]any {
	out := map[string]any{}
	if v.ID != nil {
		out["Id"] = *v.ID
	}
	if v.Name != nil {
		out["Name"] = *v.Name
	}
	return out
}

// Flow is a mainline attendant flow: an appointment booking or question and
// answer flow (Type tells which). Fields the flow type does not use are null.
// KnowledgeBase is the service's indexing status text, not the uploaded
// content; parse it with KnowledgeBaseDocuments. APIDefinitions is JSON text.
type Flow struct {
	Name                       *string                     `json:"Name"`
	Identity                   *string                     `json:"Identity"`
	Type                       *string                     `json:"Type"`
	ConfigurationID            *string                     `json:"ConfigurationId"`
	RelatedConfigurationIDs    []*RelatedConfiguration     `json:"RelatedConfigurationIds"`
	Description                *string                     `json:"Description"`
	CallerAuthenticationMethod *CallerAuthenticationMethod `json:"CallerAuthenticationMethod"`
	APIAuthenticationType      *APIAuthenticationType      `json:"ApiAuthenticationType"`
	APIDefinitions             *string                     `json:"ApiDefinitions"`
	KnowledgeBase              *string                     `json:"KnowledgeBase"`
	AppointmentBookingPlatform *string                     `json:"AppointmentBookingPlatform"`
	BookingLink                *string                     `json:"BookingLink"`
	RawFields                  map[string]json.RawMessage  `json:"-"`
}

func (v *Flow) UnmarshalJSON(raw []byte) error {
	type plain Flow
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	for _, item := range value.RelatedConfigurationIDs {
		if item == nil {
			return client.ErrDecode
		}
	}
	*v = Flow(value)
	return nil
}

// KnowledgeBaseDocument is one entry of a question and answer flow's
// knowledge base status: the document (file name or URL) and its indexing state.
type KnowledgeBaseDocument struct {
	DocumentName *string                `json:"document_name"`
	Description  *string                `json:"description"`
	Status       *string                `json:"status"`
	Error        *KnowledgeBaseDocError `json:"error"`
}

// KnowledgeBaseDocError is the indexing failure the service reports for a document.
type KnowledgeBaseDocError struct {
	Code    *int64  `json:"code"`
	Message *string `json:"message"`
}

// Knowledge base document indexing states observed from the service.
const (
	KnowledgeBaseStatusInProgress = "InProgress"
	KnowledgeBaseStatusFailed     = "Failed"
)

// KnowledgeBaseDocuments parses the flow's KnowledgeBase status text. A flow
// without a knowledge base returns nil.
func (v *Flow) KnowledgeBaseDocuments() ([]KnowledgeBaseDocument, error) {
	if v == nil || v.KnowledgeBase == nil || *v.KnowledgeBase == "" {
		return nil, nil
	}
	var values []KnowledgeBaseDocument
	if json.Unmarshal([]byte(*v.KnowledgeBase), &values) != nil {
		return nil, client.ErrDecode
	}
	return values, nil
}

// Indexing reports whether any knowledge base document is still being
// indexed. The service rejects question and answer flow updates until it is not.
func (v *Flow) Indexing() (bool, error) {
	documents, err := v.KnowledgeBaseDocuments()
	if err != nil {
		return false, err
	}
	for _, document := range documents {
		if document.Status != nil && *document.Status == KnowledgeBaseStatusInProgress {
			return true, nil
		}
	}
	return false, nil
}

// CallTarget is the spam detection template's transfer target.
type CallTarget struct {
	ID                                           *string                    `json:"Id"`
	Type                                         *string                    `json:"Type"`
	EnableTranscription                          *bool                      `json:"EnableTranscription"`
	EnableSharedVoicemailSystemPromptSuppression *bool                      `json:"EnableSharedVoicemailSystemPromptSuppression"`
	CallPriority                                 *int32                     `json:"CallPriority"`
	SharedVoicemailHistoryTemplateID             *string                    `json:"SharedVoicemailHistoryTemplateId"`
	RawFields                                    map[string]json.RawMessage `json:"-"`
}

func (v *CallTarget) UnmarshalJSON(raw []byte) error {
	type plain CallTarget
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = CallTarget(value)
	return nil
}

// wire applies the module's CallTarget expansion: the two switches are sent
// only when true and CallPriority only for application or configuration
// endpoint targets.
func (v CallTarget) wire() map[string]any {
	out := map[string]any{}
	if v.ID != nil {
		out["Id"] = *v.ID
	}
	if v.Type != nil {
		out["Type"] = *v.Type
	}
	if v.EnableTranscription != nil && *v.EnableTranscription {
		out["EnableTranscription"] = true
	}
	if v.EnableSharedVoicemailSystemPromptSuppression != nil && *v.EnableSharedVoicemailSystemPromptSuppression {
		out["EnableSharedVoicemailSystemPromptSuppression"] = true
	}
	if v.CallPriority != nil && v.Type != nil && (*v.Type == CallTargetApplicationEndpoint || *v.Type == CallTargetConfigurationEndpoint) {
		out["CallPriority"] = *v.CallPriority
	}
	if v.SharedVoicemailHistoryTemplateID != nil {
		out["SharedVoicemailHistoryTemplateId"] = *v.SharedVoicemailHistoryTemplateID
	}
	return out
}

// SpamDetectionTemplate is a mainline attendant spam detection template.
type SpamDetectionTemplate struct {
	Name                    *string                    `json:"Name"`
	Identity                *string                    `json:"Identity"`
	Description             *string                    `json:"Description"`
	EnableSpamDetection     *bool                      `json:"EnableSpamDetection"`
	Action                  *string                    `json:"Action"`
	CallTarget              *CallTarget                `json:"CallTarget"`
	RelatedConfigurationIDs []*RelatedConfiguration    `json:"RelatedConfigurationIds"`
	InclusionScope          []string                   `json:"InclusionScope"`
	ExclusionScope          []string                   `json:"ExclusionScope"`
	RawFields               map[string]json.RawMessage `json:"-"`
}

func (v *SpamDetectionTemplate) UnmarshalJSON(raw []byte) error {
	type plain SpamDetectionTemplate
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	for _, item := range value.RelatedConfigurationIDs {
		if item == nil {
			return client.ErrDecode
		}
	}
	*v = SpamDetectionTemplate(value)
	return nil
}

// Language is a language the mainline attendant supports.
type Language struct {
	ID          *string                    `json:"Id"`
	DisplayName *string                    `json:"DisplayName"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *Language) UnmarshalJSON(raw []byte) error {
	type plain Language
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Language(value)
	return nil
}

// Voice is a voice the mainline attendant supports.
type Voice struct {
	ID        *string                    `json:"Id"`
	Name      *string                    `json:"Name"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *Voice) UnmarshalJSON(raw []byte) error {
	type plain Voice
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Voice(value)
	return nil
}

// TenantInformation is the mainline attendant part of the voice apps tenant
// information (Get-CsMainlineAttendantTenantInformation). SupportedLanguages
// and SupportedVoices are what Get-CsMainlineAttendantSupportedLanguages and
// Get-CsMainlineAttendantSupportedVoices return.
type TenantInformation struct {
	DefaultLanguage    *string                    `json:"DefaultLanguage"`
	DefaultVoice       *string                    `json:"DefaultVoice"`
	DefaultTimeZone    *string                    `json:"DefaultTimeZone"`
	SupportedLanguages []*Language                `json:"SupportedLanguages"`
	SupportedVoices    []*Voice                   `json:"SupportedVoices"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *TenantInformation) UnmarshalJSON(raw []byte) error {
	type plain TenantInformation
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	for _, item := range value.SupportedLanguages {
		if item == nil {
			return client.ErrDecode
		}
	}
	for _, item := range value.SupportedVoices {
		if item == nil {
			return client.ErrDecode
		}
	}
	*v = TenantInformation(value)
	return nil
}

// ListOptions selects the module's list parameters. Nil fields and a false
// Descending are omitted.
type ListOptions struct {
	First      *int32
	Skip       *int32
	SortBy     *string
	Descending bool
	NameFilter *string
}

// FlowListOptions adds Get-CsMainlineAttendantFlow's filters.
type FlowListOptions struct {
	ListOptions
	Type                   *string
	RelatedConfigurationID *string
}

// KnowledgeBase is the content of a question and answer flow's knowledge
// base: uploaded files and web pages. The module builds it from a JSON file
// that lists local files and URLs; here the caller supplies the bytes.
type KnowledgeBase struct {
	Files []KnowledgeBaseFile
	URLs  []KnowledgeBaseURL
}

// KnowledgeBaseFile is one uploaded document. Name is the file name (its
// extension is sent separately, as the module does); Content is base64-encoded
// on the wire.
type KnowledgeBaseFile struct {
	Name        string
	Content     []byte
	Description string
}

// KnowledgeBaseURL is one web page the service indexes.
type KnowledgeBaseURL struct {
	URL         string
	Description string
}

func (KnowledgeBaseFile) String() string     { return "knowledge base file (redacted)" }
func (v KnowledgeBaseFile) GoString() string { return v.String() }

// CreateAppointmentBookingFlowRequest is New-CsMainlineAttendantAppointmentBookingFlow.
// APIDefinitions is JSON text; it must contain the operations the service
// requires (for example AuthenticateCaller) when the platform is Custom.
type CreateAppointmentBookingFlowRequest struct {
	Name                       string
	Description                string
	AppointmentBookingPlatform *string
	BookingLink                *string
	CallerAuthenticationMethod *CallerAuthenticationMethod
	APIAuthenticationType      *APIAuthenticationType
	APIDefinitions             *string
}

// UpdateAppointmentBookingFlowRequest is Set-CsMainlineAttendantAppointmentBookingFlow.
// The module sends the whole instance; nil fields are omitted here, so pass
// the values read back for anything that should be kept.
type UpdateAppointmentBookingFlowRequest struct {
	Name                       *string
	Description                *string
	RelatedConfigurationIDs    []RelatedConfiguration
	AppointmentBookingPlatform *string
	BookingLink                *string
	CallerAuthenticationMethod *CallerAuthenticationMethod
	APIAuthenticationType      *APIAuthenticationType
	APIDefinitions             *string
}

// CreateQuestionAnswerFlowRequest is New-CsMainlineAttendantQuestionAnswerFlow.
type CreateQuestionAnswerFlowRequest struct {
	Name                  string
	Description           string
	KnowledgeBase         KnowledgeBase
	APIAuthenticationType *APIAuthenticationType
}

// UpdateQuestionAnswerFlowRequest is Set-CsMainlineAttendantQuestionAnswerFlow.
// The service requires KnowledgeBase: pass the text read back (Flow.KnowledgeBase,
// as the module does) or new content from EncodeKnowledgeBase. The service
// rejects updates while documents are indexing (see Flow.Indexing).
type UpdateQuestionAnswerFlowRequest struct {
	Name                    *string
	Description             *string
	RelatedConfigurationIDs []RelatedConfiguration
	APIAuthenticationType   *APIAuthenticationType
	KnowledgeBase           *string
}

// CreateSpamDetectionTemplateRequest is New-CsMainlineAttendantSpamDetectionTemplate.
// Nil EnableSpamDetection and Action take the module defaults (true and
// DisconnectCall); both are always sent.
type CreateSpamDetectionTemplateRequest struct {
	Name                string
	Description         string
	EnableSpamDetection *bool
	Action              *string
	CallTarget          *CallTarget
	InclusionScope      []string
	ExclusionScope      []string
}

// UpdateSpamDetectionTemplateRequest is Set-CsMainlineAttendantSpamDetectionTemplate.
// Nil fields are omitted, so pass the values read back for anything that
// should be kept.
type UpdateSpamDetectionTemplateRequest struct {
	Name                    *string
	Description             *string
	EnableSpamDetection     *bool
	Action                  *string
	CallTarget              *CallTarget
	RelatedConfigurationIDs []RelatedConfiguration
	InclusionScope          []string
	ExclusionScope          []string
}
