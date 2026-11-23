// Package calltemplates manages the reusable templates that Teams call queues
// and auto attendants reference: auto recording, compliance recording for
// call queues, shared call history, shared voicemail triage settings and IVR
// tags templates.
//
// Every template type has the same five operations on its own route under
// Teams.VoiceApps, so the package exposes one generic Templates value per
// type on Service. The deprecated *-CsSharedCallQueueHistoryTemplate cmdlets
// are aliases of the shared call history template and map to
// Service.SharedCallHistory.
package calltemplates

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const voiceApps = "Teams.VoiceApps"

// Service groups the template collections. Each field shares one bound
// ConfigAPI backend with lazy forest discovery.
type Service struct {
	// AutoRecording manages *-CsAutoRecordingTemplate.
	AutoRecording *Templates[AutoRecordingTemplate]
	// ComplianceRecording manages *-CsComplianceRecordingForCallQueueTemplate.
	ComplianceRecording *Templates[ComplianceRecordingTemplate]
	// SharedCallHistory manages *-CsSharedCallHistoryTemplate and its
	// deprecated *-CsSharedCallQueueHistoryTemplate aliases.
	SharedCallHistory *Templates[SharedCallHistoryTemplate]
	// SharedVoicemailTriageSettings manages
	// *-CsSharedVoicemailTriageSettingsTemplate.
	SharedVoicemailTriageSettings *Templates[SharedVoicemailTriageSettingsTemplate]
	// Tags manages *-CsTagsTemplate.
	Tags *Templates[TagsTemplate]
}

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
func NewService(base client.Client) *Service {
	backend := configapi.New(base)
	return &Service{
		AutoRecording: &Templates[AutoRecordingTemplate]{client: backend, spec: spec[AutoRecordingTemplate]{
			segment: "auto-recording", noun: "CsAutoRecordingTemplate", list: "AutoRecordingTemplates",
			created: "AutoRecordingTemplate", item: "AutoRecordingTemplate",
			id: (*AutoRecordingTemplate).identity, body: (*AutoRecordingTemplate).body, create: validAutoRecording}},
		ComplianceRecording: &Templates[ComplianceRecordingTemplate]{client: backend, spec: spec[ComplianceRecordingTemplate]{
			segment: "compliance-recording", noun: "CsComplianceRecordingForCallQueueTemplate", list: "ComplianceRecordings",
			created: "ComplianceRecordingForCallQueue", item: "ComplianceRecording",
			id: (*ComplianceRecordingTemplate).identity, body: (*ComplianceRecordingTemplate).body, create: validCompliance, update: validComplianceCount}},
		SharedCallHistory: &Templates[SharedCallHistoryTemplate]{client: backend, spec: spec[SharedCallHistoryTemplate]{
			segment: "shared-call-history-template", noun: "CsSharedCallHistoryTemplate", list: "AllSharedCallHistoryTemplates",
			created: "SharedCallHistoryTemplate", item: "SharedCallHistoryTemplate",
			id: (*SharedCallHistoryTemplate).identity, body: (*SharedCallHistoryTemplate).body, create: validHistory, update: validHistoryValues}},
		SharedVoicemailTriageSettings: &Templates[SharedVoicemailTriageSettingsTemplate]{client: backend, spec: spec[SharedVoicemailTriageSettingsTemplate]{
			segment: "shared-voicemail-triage-settings", noun: "CsSharedVoicemailTriageSettingsTemplate", list: "AllSharedVoicemailTriageSettings",
			created: "SharedVoicemailTriageSettings", item: "SharedVoicemailTriageSettings",
			id: (*SharedVoicemailTriageSettingsTemplate).identity, body: (*SharedVoicemailTriageSettingsTemplate).body, create: validTriage}},
		Tags: &Templates[TagsTemplate]{client: backend, spec: spec[TagsTemplate]{
			segment: "ivr-tags-template", noun: "CsTagsTemplate", list: "IvrTagsTemplates",
			created: "IvrTagsTemplate", item: "IvrTagsTemplate",
			id: (*TagsTemplate).identity, body: (*TagsTemplate).body, create: validTags, update: validTagValues}},
	}
}

// spec describes one template route and its response envelopes.
type spec[T any] struct {
	segment string // route under Teams.VoiceApps
	noun    string // cmdlet noun for X-MS-CmdletName
	list    string // list response array property
	created string // create response object property
	item    string // get and update response object property
	id      func(*T) *string
	body    func(*T) map[string]any
	create  func(*T) bool // module mandatory parameters for New-*
	update  func(*T) bool // value checks shared by New-* and Set-*
}

// Templates is one template collection. Its zero value is unusable; obtain it
// from Service.
type Templates[T any] struct {
	client client.Client
	spec   spec[T]
}

func (t *Templates[T]) check(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if t == nil || t.client == nil {
		return client.ErrClosed
	}
	return nil
}

func text(value *string) bool {
	return value != nil && strings.TrimSpace(*value) != "" && !strings.ContainsFunc(*value, unicode.IsControl)
}

func validPart(value string) bool {
	return strings.TrimSpace(value) != "" && value != "." && value != ".." &&
		!strings.ContainsAny(value, `/\`) && !strings.ContainsFunc(value, unicode.IsControl)
}

func headers(verb, noun string) http.Header {
	return http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {verb + "-" + noun}}
}

func (t *Templates[T]) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := t.client.Do(ctx, request, nil)
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

// envelope decodes a required, non-null property of a voice apps response.
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

func (t *Templates[T]) decode(response *resty.Response, name string) (*T, error) {
	var value T
	if envelope(response.Body(), name, &value) != nil {
		return nil, client.ErrDecode
	}
	return &value, nil
}

// List reads every template of this type (Get-* without Id).
func (t *Templates[T]) List(ctx context.Context) ([]*T, *resty.Response, error) {
	if err := t.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := t.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, t.spec.segment},
		Header: headers("Get", t.spec.noun), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	var values []*T
	if envelope(response.Body(), t.spec.list, &values) != nil || values == nil {
		return nil, response, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	return values, response, nil
}

// GetByID reads one template (Get-* -Id). A missing template is an HTTP 404
// API error with service code 100004.
func (t *Templates[T]) GetByID(ctx context.Context, id string) (*T, *resty.Response, error) {
	if err := t.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := t.do(ctx, client.Request{Method: http.MethodGet, Path: []string{voiceApps, t.spec.segment, id},
		Header: headers("Get", t.spec.noun), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := t.decode(response, t.spec.item)
	return value, response, err
}

// Create adds a template (New-*). ID must be nil; the service assigns it.
// Nil optional fields are omitted and take service defaults. Create is never
// replayed automatically; reconcile a lost response by listing for the name.
func (t *Templates[T]) Create(ctx context.Context, value *T) (*T, *resty.Response, error) {
	if err := t.check(ctx); err != nil {
		return nil, nil, err
	}
	if value == nil || t.spec.id(value) != nil || !t.spec.create(value) || (t.spec.update != nil && !t.spec.update(value)) {
		return nil, nil, ErrInput
	}
	response, err := t.do(ctx, client.Request{Method: http.MethodPost, Path: []string{voiceApps, t.spec.segment},
		Header: headers("New", t.spec.noun), Body: t.spec.body(value)}, http.StatusCreated, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	created, err := t.decode(response, t.spec.created)
	return created, response, err
}

// UpdateByID replaces a template (Set-*). The service stores the body as
// sent, so pass the full template — typically one read with GetByID and then
// edited. A non-nil value ID must equal id.
func (t *Templates[T]) UpdateByID(ctx context.Context, id string, value *T) (*T, *resty.Response, error) {
	if err := t.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) || value == nil || (t.spec.id(value) != nil && *t.spec.id(value) != id) || !text(namePtr(t.spec.body(value))) ||
		(t.spec.update != nil && !t.spec.update(value)) {
		return nil, nil, ErrInput
	}
	response, err := t.do(ctx, client.Request{Method: http.MethodPut, Path: []string{voiceApps, t.spec.segment, id},
		Header: headers("Set", t.spec.noun), Body: t.spec.body(value), Replay: client.ReplaySafe}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	updated, err := t.decode(response, t.spec.item)
	return updated, response, err
}

// DeleteByID removes a template (Remove-*). Deleting a missing template is an
// HTTP 404 API error with service code 100004.
func (t *Templates[T]) DeleteByID(ctx context.Context, id string) (*resty.Response, error) {
	if err := t.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(id) {
		return nil, ErrInput
	}
	return t.do(ctx, client.Request{Method: http.MethodDelete, Path: []string{voiceApps, t.spec.segment, id},
		Header: headers("Remove", t.spec.noun)}, http.StatusOK, http.StatusNoContent)
}

func namePtr(body map[string]any) *string {
	name, ok := body["Name"].(string)
	if !ok {
		return nil
	}
	return &name
}

func validAutoRecording(v *AutoRecordingTemplate) bool {
	return text(v.Name) && v.Description != nil && v.TranscriptionEnabled != nil && v.RecordingEnabled != nil &&
		v.AgentViewPermission != nil && text(v.SharePointHostName) && text(v.SharePointSiteName) && text(v.RecordingDocumentOwner)
}

func validCompliance(v *ComplianceRecordingTemplate) bool {
	return text(v.Name) && v.Description != nil && text(v.BotID)
}

// validComplianceCount mirrors the module: ConcurrentInvitationCount is 1 or 2.
func validComplianceCount(v *ComplianceRecordingTemplate) bool {
	return v.ConcurrentInvitationCount == nil || *v.ConcurrentInvitationCount == 1 || *v.ConcurrentInvitationCount == 2
}

func validHistory(v *SharedCallHistoryTemplate) bool {
	return text(v.Name) && v.Description != nil
}

func historyValue(value *int32) bool { return value == nil || (*value >= HistoryNone && *value <= 2) }

func validHistoryValues(v *SharedCallHistoryTemplate) bool {
	return historyValue(v.IncomingMissedCalls) && historyValue(v.AnsweredAndOutboundCalls) && historyValue(v.IncomingRedirectedCalls)
}

func validTriage(v *SharedVoicemailTriageSettingsTemplate) bool {
	return text(v.Name) && v.Description != nil
}

func validTags(v *TagsTemplate) bool {
	return text(v.Name) && v.Description != nil && len(v.Tags) > 0
}

func validTagValues(v *TagsTemplate) bool {
	for _, tag := range v.Tags {
		if !text(tag.TagName) || tag.TagDetails == nil || !text(tag.TagDetails.ID) || !text(tag.TagDetails.Type) {
			return false
		}
	}
	return true
}
