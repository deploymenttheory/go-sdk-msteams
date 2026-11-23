// Package mainlineattendants manages the Teams mainline attendant: its
// appointment booking and question and answer flows, spam detection
// templates, and the supported languages and voices reported in the voice
// apps tenant information.
package mainlineattendants

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	root               = "Teams.VoiceApps"
	flows              = "mainline-attendant-flow"
	appointmentBooking = "appointment-booking"
	questionAnswer     = "question-answer"
	spamTemplate       = "spam-detection-template"
	// maxDescription is the module's spam detection template description limit.
	maxDescription = 500
)

// Service composes the bound ConfigAPI backend with lazy forest discovery.
type Service struct{ client client.Client }

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func (s *Service) check(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.client == nil {
		return client.ErrClosed
	}
	return nil
}

func validText(value string) bool {
	return strings.TrimSpace(value) != "" && !strings.ContainsFunc(value, unicode.IsControl)
}

func validPart(value string) bool {
	return validText(value) && value != "." && value != ".." && !strings.ContainsAny(value, `/\`)
}

func headers(command string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
}

func (s *Service) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := s.client.Do(ctx, request, nil)
	if err != nil {
		return response, err
	}
	for _, status := range statuses {
		if response != nil && response.StatusCode() == status {
			return response, nil
		}
	}
	return response, ErrResponse
}

// envelope decodes the required, non-null property name of a voice apps response.
func envelope(raw []byte, name string, out any) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	value, ok := fields[name]
	if !ok || string(value) == "null" || json.Unmarshal(value, out) != nil {
		return client.ErrDecode
	}
	return nil
}

func decodeList[T any](raw []byte, name string) ([]*T, error) {
	var values []*T
	if err := envelope(raw, name, &values); err != nil || values == nil {
		return nil, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, client.ErrDecode
		}
	}
	return values, nil
}

func decodeOne[T any](raw []byte, name string) (*T, error) {
	var value T
	if err := envelope(raw, name, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (o *ListOptions) query(values url.Values) url.Values {
	if o == nil {
		return values
	}
	if values == nil {
		values = url.Values{}
	}
	if o.First != nil {
		values.Set("First", strconv.FormatInt(int64(*o.First), 10))
	}
	if o.Skip != nil {
		values.Set("Skip", strconv.FormatInt(int64(*o.Skip), 10))
	}
	if o.NameFilter != nil {
		values.Set("NameFilter", *o.NameFilter)
	}
	if o.SortBy != nil {
		values.Set("SortBy", *o.SortBy)
	}
	if o.Descending {
		values.Set("Descending", "True")
	}
	return values
}

func (o *ListOptions) valid() bool {
	return o == nil || (o.First == nil || *o.First >= 0) && (o.Skip == nil || *o.Skip >= 0)
}

func (s *Service) list(ctx context.Context, segments []string, query url.Values, command string) (*resty.Response, error) {
	if len(query) == 0 {
		query = nil
	}
	return s.do(ctx, client.Request{Method: http.MethodGet, Path: segments, Query: query, Header: headers(command), Replay: client.ReplaySafe}, http.StatusOK)
}

func (s *Service) get(ctx context.Context, segments []string, command string) (*resty.Response, error) {
	return s.do(ctx, client.Request{Method: http.MethodGet, Path: segments, Header: headers(command), Replay: client.ReplaySafe}, http.StatusOK)
}

// GetTenantInformation reads the mainline attendant tenant information: the
// default language, voice and time zone and the supported languages and
// voices (Get-CsMainlineAttendantTenantInformation,
// Get-CsMainlineAttendantSupportedLanguages and
// Get-CsMainlineAttendantSupportedVoices all read this one response).
func (s *Service) GetTenantInformation(ctx context.Context) (*TenantInformation, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.get(ctx, []string{root, "information"}, "Get-CsMainlineAttendantTenantInformation")
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[TenantInformation](response.Body(), "MainlineAttendantTenantInformation")
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// ListFlows reads mainline attendant flows of every type (Get-CsMainlineAttendantFlow).
func (s *Service) ListFlows(ctx context.Context, options *FlowListOptions) ([]*Flow, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	var query url.Values
	if options != nil {
		if !options.valid() || options.Type != nil && !validText(*options.Type) ||
			options.RelatedConfigurationID != nil && !validText(*options.RelatedConfigurationID) {
			return nil, nil, ErrInput
		}
		query = url.Values{}
		if options.Type != nil {
			query.Set("type", *options.Type)
		}
		if options.RelatedConfigurationID != nil {
			query.Set("relatedConfigurationId", *options.RelatedConfigurationID)
		}
		query = options.query(query)
	}
	response, err := s.list(ctx, []string{root, flows}, query, "Get-CsMainlineAttendantFlow")
	if err != nil {
		return nil, response, err
	}
	values, err := decodeList[Flow](response.Body(), "MainlineAttendantFlowResponses")
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// GetFlowByID reads one flow of any type (Get-CsMainlineAttendantFlow -Identity).
func (s *Service) GetFlowByID(ctx context.Context, id string) (*Flow, *resty.Response, error) {
	return s.getFlow(ctx, []string{root, flows, id}, id, "Get-CsMainlineAttendantFlow")
}

func (s *Service) getFlow(ctx context.Context, segments []string, id, command string) (*Flow, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := s.get(ctx, segments, command)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[Flow](response.Body(), "MainlineAttendantFlowResponse")
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

func (s *Service) listFlows(ctx context.Context, kind string, options *ListOptions, command string) ([]*Flow, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !options.valid() {
		return nil, nil, ErrInput
	}
	response, err := s.list(ctx, []string{root, flows, kind}, options.query(nil), command)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeList[Flow](response.Body(), "MainlineAttendantFlowResponses")
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

func (s *Service) writeFlow(ctx context.Context, method string, segments []string, body map[string]any, command, property string, statuses ...int) (*Flow, *resty.Response, error) {
	response, err := s.do(ctx, client.Request{Method: method, Path: segments, Header: headers(command), Body: body}, statuses...)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[Flow](response.Body(), property)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

func (s *Service) remove(ctx context.Context, kind, id, command string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(id) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: []string{root, flows, kind, id}, Header: headers(command)}, http.StatusOK, http.StatusNoContent)
}

func related(values []RelatedConfiguration) []map[string]any {
	out := make([]map[string]any, 0, len(values))
	for _, value := range values {
		out = append(out, value.wire())
	}
	return out
}

func validAPIDefinitions(value *string) bool {
	return value == nil || json.Valid([]byte(*value))
}

func setOptional(body map[string]any, name string, value any) {
	switch v := value.(type) {
	case *string:
		if v != nil {
			body[name] = *v
		}
	case *bool:
		if v != nil {
			body[name] = *v
		}
	case *CallerAuthenticationMethod:
		if v != nil {
			body[name] = int32(*v)
		}
	case *APIAuthenticationType:
		if v != nil {
			body[name] = int32(*v)
		}
	}
}

// ListAppointmentBookingFlows reads appointment booking flows
// (Get-CsMainlineAttendantAppointmentBookingFlow).
func (s *Service) ListAppointmentBookingFlows(ctx context.Context, options *ListOptions) ([]*Flow, *resty.Response, error) {
	return s.listFlows(ctx, appointmentBooking, options, "Get-CsMainlineAttendantAppointmentBookingFlow")
}

// GetAppointmentBookingFlowByID reads one appointment booking flow.
func (s *Service) GetAppointmentBookingFlowByID(ctx context.Context, id string) (*Flow, *resty.Response, error) {
	return s.getFlow(ctx, []string{root, flows, appointmentBooking, id}, id, "Get-CsMainlineAttendantAppointmentBookingFlow")
}

// CreateAppointmentBookingFlow creates an appointment booking flow
// (New-CsMainlineAttendantAppointmentBookingFlow). It is never replayed
// automatically; reconcile a lost response by listing for the name.
func (s *Service) CreateAppointmentBookingFlow(ctx context.Context, request CreateAppointmentBookingFlowRequest) (*Flow, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.Name) || !validText(request.Description) || !validAPIDefinitions(request.APIDefinitions) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"Name": request.Name, "Description": request.Description}
	setOptional(body, "CallerAuthenticationMethod", request.CallerAuthenticationMethod)
	setOptional(body, "ApiAuthenticationType", request.APIAuthenticationType)
	setOptional(body, "ApiDefinitions", request.APIDefinitions)
	setOptional(body, "AppointmentBookingPlatform", request.AppointmentBookingPlatform)
	setOptional(body, "BookingLink", request.BookingLink)
	return s.writeFlow(ctx, http.MethodPost, []string{root, flows, appointmentBooking}, body,
		"New-CsMainlineAttendantAppointmentBookingFlow", "MainlineAttendantAppointmentBookingFlow", http.StatusCreated, http.StatusOK)
}

// UpdateAppointmentBookingFlowByID replaces an appointment booking flow
// (Set-CsMainlineAttendantAppointmentBookingFlow).
func (s *Service) UpdateAppointmentBookingFlowByID(ctx context.Context, id string, request UpdateAppointmentBookingFlowRequest) (*Flow, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) || !validAPIDefinitions(request.APIDefinitions) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"Type": FlowTypeAppointmentBooking}
	setOptional(body, "Name", request.Name)
	if request.RelatedConfigurationIDs != nil {
		body["RelatedConfigurationIds"] = related(request.RelatedConfigurationIDs)
	}
	setOptional(body, "Description", request.Description)
	setOptional(body, "CallerAuthenticationMethod", request.CallerAuthenticationMethod)
	setOptional(body, "ApiAuthenticationType", request.APIAuthenticationType)
	setOptional(body, "ApiDefinitions", request.APIDefinitions)
	setOptional(body, "AppointmentBookingPlatform", request.AppointmentBookingPlatform)
	setOptional(body, "BookingLink", request.BookingLink)
	return s.writeFlow(ctx, http.MethodPut, []string{root, flows, appointmentBooking, id}, body,
		"Set-CsMainlineAttendantAppointmentBookingFlow", "MainlineAttendantAppointmentBookingFlow", http.StatusOK)
}

// DeleteAppointmentBookingFlowByID removes an appointment booking flow
// (Remove-CsMainlineAttendantAppointmentBookingFlow).
func (s *Service) DeleteAppointmentBookingFlowByID(ctx context.Context, id string) (*resty.Response, error) {
	return s.remove(ctx, appointmentBooking, id, "Remove-CsMainlineAttendantAppointmentBookingFlow")
}

// EncodeKnowledgeBase serializes knowledge base content into the text the
// service accepts, as the module's ReadKnowledgeBaseContent does. At least one
// file or URL is required; file names must not contain path separators.
func EncodeKnowledgeBase(value KnowledgeBase) (string, error) {
	if len(value.Files) == 0 && len(value.URLs) == 0 {
		return "", ErrInput
	}
	type file struct {
		Name        string `json:"file_name"`
		Content     string `json:"file_content"`
		Extension   string `json:"file_extension"`
		Description string `json:"file_description"`
	}
	type link struct {
		URL         string `json:"url_link"`
		Description string `json:"url_description"`
	}
	content := struct {
		Files []file `json:"file_descriptions"`
		URLs  []link `json:"url_descriptions"`
	}{Files: []file{}, URLs: []link{}}
	for _, item := range value.Files {
		if !validPart(item.Name) || len(item.Content) == 0 || strings.ContainsFunc(item.Description, unicode.IsControl) {
			return "", ErrInput
		}
		content.Files = append(content.Files, file{Name: item.Name, Content: base64.StdEncoding.EncodeToString(item.Content),
			Extension: path.Ext(item.Name), Description: item.Description})
	}
	for _, item := range value.URLs {
		if !validText(item.URL) || strings.ContainsFunc(item.Description, unicode.IsControl) {
			return "", ErrInput
		}
		content.URLs = append(content.URLs, link(item))
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return "", ErrInput
	}
	return string(raw), nil
}

// ListQuestionAnswerFlows reads question and answer flows
// (Get-CsMainlineAttendantQuestionAnswerFlow).
func (s *Service) ListQuestionAnswerFlows(ctx context.Context, options *ListOptions) ([]*Flow, *resty.Response, error) {
	return s.listFlows(ctx, questionAnswer, options, "Get-CsMainlineAttendantQuestionAnswerFlow")
}

// GetQuestionAnswerFlowByID reads one question and answer flow.
func (s *Service) GetQuestionAnswerFlowByID(ctx context.Context, id string) (*Flow, *resty.Response, error) {
	return s.getFlow(ctx, []string{root, flows, questionAnswer, id}, id, "Get-CsMainlineAttendantQuestionAnswerFlow")
}

// CreateQuestionAnswerFlow creates a question and answer flow
// (New-CsMainlineAttendantQuestionAnswerFlow). The service indexes the
// knowledge base asynchronously. It is never replayed automatically.
func (s *Service) CreateQuestionAnswerFlow(ctx context.Context, request CreateQuestionAnswerFlowRequest) (*Flow, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.Name) || !validText(request.Description) {
		return nil, nil, ErrInput
	}
	knowledgeBase, err := EncodeKnowledgeBase(request.KnowledgeBase)
	if err != nil {
		return nil, nil, err
	}
	body := map[string]any{"Name": request.Name, "Description": request.Description, "KnowledgeBase": knowledgeBase}
	setOptional(body, "ApiAuthenticationType", request.APIAuthenticationType)
	return s.writeFlow(ctx, http.MethodPost, []string{root, flows, questionAnswer}, body,
		"New-CsMainlineAttendantQuestionAnswerFlow", "MainlineAttendantQuestionAnswerFlow", http.StatusCreated, http.StatusOK)
}

// UpdateQuestionAnswerFlowByID replaces a question and answer flow
// (Set-CsMainlineAttendantQuestionAnswerFlow).
func (s *Service) UpdateQuestionAnswerFlowByID(ctx context.Context, id string, request UpdateQuestionAnswerFlowRequest) (*Flow, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"Type": FlowTypeQuestionAnswer}
	setOptional(body, "Name", request.Name)
	if request.RelatedConfigurationIDs != nil {
		body["RelatedConfigurationIds"] = related(request.RelatedConfigurationIDs)
	}
	setOptional(body, "Description", request.Description)
	setOptional(body, "ApiAuthenticationType", request.APIAuthenticationType)
	setOptional(body, "KnowledgeBase", request.KnowledgeBase)
	return s.writeFlow(ctx, http.MethodPut, []string{root, flows, questionAnswer, id}, body,
		"Set-CsMainlineAttendantQuestionAnswerFlow", "MainlineAttendantQuestionAnswerFlow", http.StatusOK)
}

// DeleteQuestionAnswerFlowByID removes a question and answer flow
// (Remove-CsMainlineAttendantQuestionAnswerFlow).
func (s *Service) DeleteQuestionAnswerFlowByID(ctx context.Context, id string) (*resty.Response, error) {
	return s.remove(ctx, questionAnswer, id, "Remove-CsMainlineAttendantQuestionAnswerFlow")
}

// ListSpamDetectionTemplates reads spam detection templates
// (Get-CsMainlineAttendantSpamDetectionTemplate).
func (s *Service) ListSpamDetectionTemplates(ctx context.Context, options *ListOptions) ([]*SpamDetectionTemplate, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !options.valid() {
		return nil, nil, ErrInput
	}
	response, err := s.list(ctx, []string{root, flows, spamTemplate}, options.query(nil), "Get-CsMainlineAttendantSpamDetectionTemplate")
	if err != nil {
		return nil, response, err
	}
	values, err := decodeList[SpamDetectionTemplate](response.Body(), "SpamDetectionTemplates")
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// GetSpamDetectionTemplateByID reads one spam detection template.
func (s *Service) GetSpamDetectionTemplateByID(ctx context.Context, id string) (*SpamDetectionTemplate, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := s.get(ctx, []string{root, flows, spamTemplate, id}, "Get-CsMainlineAttendantSpamDetectionTemplate")
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[SpamDetectionTemplate](response.Body(), "SpamDetectionTemplate")
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

func validDescription(value string) bool {
	return !strings.ContainsFunc(value, unicode.IsControl) && utf8.RuneCountInString(value) <= maxDescription
}

func spamBody(body map[string]any, target *CallTarget, inclusion, exclusion []string) bool {
	if target != nil {
		body["CallTarget"] = target.wire()
	}
	for name, scope := range map[string][]string{"InclusionScope": inclusion, "ExclusionScope": exclusion} {
		if scope == nil {
			continue
		}
		for _, item := range scope {
			if !validText(item) {
				return false
			}
		}
		body[name] = scope
	}
	return true
}

func (s *Service) writeTemplate(ctx context.Context, method string, segments []string, body map[string]any, command string, statuses ...int) (*SpamDetectionTemplate, *resty.Response, error) {
	response, err := s.do(ctx, client.Request{Method: method, Path: segments, Header: headers(command), Body: body}, statuses...)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[SpamDetectionTemplate](response.Body(), "SpamDetectionTemplate")
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// CreateSpamDetectionTemplate creates a spam detection template
// (New-CsMainlineAttendantSpamDetectionTemplate). The description is required
// and limited to 500 characters, as in the module. It is never replayed
// automatically; reconcile a lost response by listing for the name.
func (s *Service) CreateSpamDetectionTemplate(ctx context.Context, request CreateSpamDetectionTemplateRequest) (*SpamDetectionTemplate, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.Name) || !validText(request.Description) || !validDescription(request.Description) {
		return nil, nil, ErrInput
	}
	enabled, action := true, SpamActionDisconnectCall
	if request.EnableSpamDetection != nil {
		enabled = *request.EnableSpamDetection
	}
	if request.Action != nil {
		action = *request.Action
	}
	if !validText(action) {
		return nil, nil, ErrInput
	}
	body := map[string]any{"Name": request.Name, "Description": request.Description, "EnableSpamDetection": enabled, "Action": action}
	if !spamBody(body, request.CallTarget, request.InclusionScope, request.ExclusionScope) {
		return nil, nil, ErrInput
	}
	return s.writeTemplate(ctx, http.MethodPost, []string{root, flows, spamTemplate}, body,
		"New-CsMainlineAttendantSpamDetectionTemplate", http.StatusCreated, http.StatusOK)
}

// UpdateSpamDetectionTemplateByID replaces a spam detection template
// (Set-CsMainlineAttendantSpamDetectionTemplate).
func (s *Service) UpdateSpamDetectionTemplateByID(ctx context.Context, id string, request UpdateSpamDetectionTemplateRequest) (*SpamDetectionTemplate, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) || request.Description != nil && !validDescription(*request.Description) {
		return nil, nil, ErrInput
	}
	body := map[string]any{}
	setOptional(body, "Name", request.Name)
	setOptional(body, "Description", request.Description)
	setOptional(body, "EnableSpamDetection", request.EnableSpamDetection)
	setOptional(body, "Action", request.Action)
	if request.RelatedConfigurationIDs != nil {
		body["RelatedConfigurationIds"] = related(request.RelatedConfigurationIDs)
	}
	if !spamBody(body, request.CallTarget, request.InclusionScope, request.ExclusionScope) {
		return nil, nil, ErrInput
	}
	return s.writeTemplate(ctx, http.MethodPut, []string{root, flows, spamTemplate, id}, body,
		"Set-CsMainlineAttendantSpamDetectionTemplate", http.StatusOK)
}

// DeleteSpamDetectionTemplateByID removes a spam detection template
// (Remove-CsMainlineAttendantSpamDetectionTemplate).
func (s *Service) DeleteSpamDetectionTemplateByID(ctx context.Context, id string) (*resty.Response, error) {
	return s.remove(ctx, spamTemplate, id, "Remove-CsMainlineAttendantSpamDetectionTemplate")
}
