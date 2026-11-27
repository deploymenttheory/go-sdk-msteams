package calltemplates_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ct "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"

	missing  = "00000000-0000-4000-8000-000000000001"
	botID    = "00000000-0000-4000-8000-0000000000b0"
	jsonType = "application/json; charset=utf-8"
	name     = "fixture-template"
	initial  = "SDK lab non-default initial description"
	updated  = "SDK lab non-default updated description"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply []byte) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: configOrigin + "/Teams.VoiceApps/" + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {jsonType}}, Body: reply}}
}

func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("call template contract %s differed: %+v", c.Name, result.Differences())
	}
}

func apiError(status int, body []byte) *client.APIError {
	var envelope struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(body, &envelope)
	result := &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}
	if envelope.Code != "" {
		result.Codes = []string{envelope.Code}
	}
	return result
}

func ptr[T any](v T) *T { return &v }

func is[T comparable](value *T, want T) bool { return value != nil && *value == want }

// fixtureID returns the redacted template ID from a create fixture.
func fixtureID(t *testing.T, raw []byte, property string) string {
	t.Helper()
	var envelope map[string]json.RawMessage
	var value struct {
		ID string `json:"Id"`
	}
	if json.Unmarshal(raw, &envelope) != nil || json.Unmarshal(envelope[property], &value) != nil || value.ID == "" {
		t.Fatal("fixture template ID unavailable")
	}
	return value.ID
}

// lifecycle replays one captured create/get/update/list/delete run.
type lifecycle[T any] struct {
	label, segment, noun, created string
	createBody, updateBody        string
	templates                     func(*msteams.Client) *ct.Templates[T]
	build                         func() *T
	edit                          func(*T)
	createdOK, updatedOK          func(*T) bool
	id                            func(*T) *string
}

func replay[T any](t *testing.T, l lifecycle[T]) {
	t.Helper()
	file := func(suffix string) []byte { return readFile(t, "testdata/"+l.label+"_"+suffix+".json") }
	create, get, update, list, deleted, gone := file("create_201"), file("get"), file("update"), file("list"), file("delete"), file("delete_missing_404")
	id := fixtureID(t, create, l.created)
	run(t, testutil.Case{Name: "actual-" + l.label + "-lifecycle",

		Steps: []testutil.Exchange{discovery(t),
			wire("POST", l.segment, "New-"+l.noun, json.RawMessage(l.createBody), 201, create),
			wire("GET", l.segment+"/"+id, "Get-"+l.noun, nil, 200, get),
			wire("PUT", l.segment+"/"+id, "Set-"+l.noun, json.RawMessage(l.updateBody), 200, update),
			wire("GET", l.segment, "Get-"+l.noun, nil, 200, list),
			wire("DELETE", l.segment+"/"+id, "Remove-"+l.noun, nil, 200, deleted),
			wire("DELETE", l.segment+"/"+id, "Remove-"+l.noun, nil, 404, gone)},
		Want: testutil.Expected{Status: 404, Body: gone, APIError: apiError(404, gone)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := l.templates(sdk)
			value, response, err := s.Create(ctx, l.build())
			if err != nil || response.StatusCode() != 201 || !is(l.id(value), id) || !l.createdOK(value) {
				t.Fatal("captured create was not decoded")
			}
			value, _, err = s.GetByID(ctx, id)
			if err != nil || !is(l.id(value), id) || !l.createdOK(value) {
				t.Fatal("captured read-back was not decoded")
			}
			l.edit(value)
			value, _, err = s.UpdateByID(ctx, id, value)
			if err != nil || !is(l.id(value), id) || !l.updatedOK(value) {
				t.Fatal("captured update was not decoded")
			}
			values, _, err := s.List(ctx)
			if err != nil || len(values) != 1 || !is(l.id(values[0]), id) || !l.updatedOK(values[0]) {
				t.Fatal("captured list was not decoded")
			}
			if response, err := s.DeleteByID(ctx, id); err != nil || response.StatusCode() != 200 {
				t.Fatal("captured delete was not accepted")
			}
			response, err = s.DeleteByID(ctx, id)
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitComplianceRecordingLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	replay(t, lifecycle[ct.ComplianceRecordingTemplate]{label: "compliance", segment: "compliance-recording", noun: "CsComplianceRecordingForCallQueueTemplate",
		created:    "ComplianceRecordingForCallQueue",
		createBody: `{"Name":"` + name + `","Description":"` + initial + `","BotId":"` + botID + `","RequiredDuringCall":true,"RequiredBeforeCall":false,"ConcurrentInvitationCount":1}`,
		updateBody: `{"Name":"` + name + `","Description":"` + updated + `","BotId":"` + botID + `","RequiredDuringCall":false,"RequiredBeforeCall":true,"ConcurrentInvitationCount":2}`,
		templates: func(sdk *msteams.Client) *ct.Templates[ct.ComplianceRecordingTemplate] {
			return sdk.TeamsAdminAPI.CallTemplates.ComplianceRecording
		},
		build: func() *ct.ComplianceRecordingTemplate {
			return &ct.ComplianceRecordingTemplate{Name: ptr(name), Description: ptr(initial), BotID: ptr(botID),
				RequiredDuringCall: ptr(true), RequiredBeforeCall: ptr(false), ConcurrentInvitationCount: ptr(int32(1))}
		},
		edit: func(v *ct.ComplianceRecordingTemplate) {
			v.Description, v.RequiredDuringCall, v.RequiredBeforeCall, v.ConcurrentInvitationCount = ptr(updated), ptr(false), ptr(true), ptr(int32(2))
		},
		// The detail read uses Bot-prefixed names; they decode into the same fields.
		createdOK: func(v *ct.ComplianceRecordingTemplate) bool {
			return is(v.Name, name) && is(v.BotID, botID) && is(v.RequiredDuringCall, true) && is(v.RequiredBeforeCall, false) &&
				is(v.ConcurrentInvitationCount, int32(1)) && v.PairedApplication == nil
		},
		updatedOK: func(v *ct.ComplianceRecordingTemplate) bool {
			return is(v.Description, updated) && is(v.RequiredDuringCall, false) && is(v.RequiredBeforeCall, true) && is(v.ConcurrentInvitationCount, int32(2))
		},
		id: func(v *ct.ComplianceRecordingTemplate) *string { return v.ID },
	})
}

func TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	replay(t, lifecycle[ct.SharedCallHistoryTemplate]{label: "history", segment: "shared-call-history-template", noun: "CsSharedCallHistoryTemplate",
		created:    "SharedCallHistoryTemplate",
		createBody: `{"Name":"` + name + `","Description":"` + initial + `","IncomingMissedCalls":1,"AnsweredAndOutboundCalls":2}`,
		updateBody: `{"Name":"` + name + `","Description":"` + updated + `","IncomingMissedCalls":2,"AnsweredAndOutboundCalls":0,"IncomingRedirectedCalls":2}`,
		templates: func(sdk *msteams.Client) *ct.Templates[ct.SharedCallHistoryTemplate] {
			return sdk.TeamsAdminAPI.CallTemplates.SharedCallHistory
		},
		build: func() *ct.SharedCallHistoryTemplate {
			return &ct.SharedCallHistoryTemplate{Name: ptr(name), Description: ptr(initial),
				IncomingMissedCalls: ptr(ct.HistoryAuthorizedUsersOnly), AnsweredAndOutboundCalls: ptr(ct.HistoryAuthorizedUsersAndAgents)}
		},
		edit: func(v *ct.SharedCallHistoryTemplate) {
			v.Description, v.IncomingMissedCalls, v.AnsweredAndOutboundCalls = ptr(updated), ptr(ct.HistoryAuthorizedUsersAndAgents), ptr(ct.HistoryNone)
			v.IncomingRedirectedCalls = ptr(ct.HistoryAuthorizedUsersAndGroupMembers)
		},
		createdOK: func(v *ct.SharedCallHistoryTemplate) bool {
			return len(v.RawFields) == 6 && is(v.IncomingMissedCalls, int32(1)) && is(v.AnsweredAndOutboundCalls, int32(2)) && is(v.IncomingRedirectedCalls, int32(0))
		},
		updatedOK: func(v *ct.SharedCallHistoryTemplate) bool {
			return is(v.Description, updated) && is(v.IncomingMissedCalls, int32(2)) && is(v.AnsweredAndOutboundCalls, int32(0)) && is(v.IncomingRedirectedCalls, int32(2))
		},
		id: func(v *ct.SharedCallHistoryTemplate) *string { return v.ID },
	})
}

func TestUnitVoicemailTriageLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	replay(t, lifecycle[ct.SharedVoicemailTriageSettingsTemplate]{label: "triage", segment: "shared-voicemail-triage-settings", noun: "CsSharedVoicemailTriageSettingsTemplate",
		created:    "SharedVoicemailTriageSettings",
		createBody: `{"Name":"` + name + `","Description":"` + initial + `","EnableUrgencyDetection":false,"EnableCategoryDetection":true}`,
		updateBody: `{"Name":"` + name + `","Description":"` + updated + `","EnableUrgencyDetection":true,"EnableCategoryDetection":true,"EnableCallToActionDetection":true,"EnableVoiceToTextSummary":false}`,
		templates: func(sdk *msteams.Client) *ct.Templates[ct.SharedVoicemailTriageSettingsTemplate] {
			return sdk.TeamsAdminAPI.CallTemplates.SharedVoicemailTriageSettings
		},
		build: func() *ct.SharedVoicemailTriageSettingsTemplate {
			return &ct.SharedVoicemailTriageSettingsTemplate{Name: ptr(name), Description: ptr(initial), EnableUrgencyDetection: ptr(false), EnableCategoryDetection: ptr(true)}
		},
		edit: func(v *ct.SharedVoicemailTriageSettingsTemplate) {
			v.Description, v.EnableUrgencyDetection, v.EnableVoiceToTextSummary = ptr(updated), ptr(true), ptr(false)
		},
		// Omitted features are enabled by the service on create.
		createdOK: func(v *ct.SharedVoicemailTriageSettingsTemplate) bool {
			return is(v.EnableUrgencyDetection, false) && is(v.EnableCategoryDetection, true) && is(v.EnableCallToActionDetection, true) && is(v.EnableVoiceToTextSummary, true)
		},
		updatedOK: func(v *ct.SharedVoicemailTriageSettingsTemplate) bool {
			return is(v.Description, updated) && is(v.EnableUrgencyDetection, true) && is(v.EnableVoiceToTextSummary, false)
		},
		id: func(v *ct.SharedVoicemailTriageSettingsTemplate) *string { return v.ID },
	})
}

func TestUnitTagsLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	// The fixture writer replaces the fictional tel: targets with stable
	// fixture-record values; requests use those values to match.
	update := readFile(t, "testdata/tags_update.json")
	var stored struct {
		IvrTagsTemplate struct {
			Tags []struct {
				TagDetails struct {
					ID string `json:"Id"`
				} `json:"TagDetails"`
			} `json:"Tags"`
		} `json:"IvrTagsTemplate"`
	}
	if json.Unmarshal(update, &stored) != nil || len(stored.IvrTagsTemplate.Tags) != 2 {
		t.Fatal("tags fixture unavailable")
	}
	create := readFile(t, "testdata/tags_create_201.json")
	var first struct {
		IvrTagsTemplate struct {
			Tags []struct {
				TagDetails struct {
					ID string `json:"Id"`
				} `json:"TagDetails"`
			} `json:"Tags"`
		} `json:"IvrTagsTemplate"`
	}
	if json.Unmarshal(create, &first) != nil || len(first.IvrTagsTemplate.Tags) != 1 {
		t.Fatal("tags fixture unavailable")
	}
	sales, support, billing := first.IvrTagsTemplate.Tags[0].TagDetails.ID, stored.IvrTagsTemplate.Tags[0].TagDetails.ID, stored.IvrTagsTemplate.Tags[1].TagDetails.ID
	tag := func(name, target string) ct.Tag {
		return ct.Tag{TagName: ptr(name), TagDetails: &ct.TagTarget{ID: ptr(target), Type: ptr("ExternalPstn")}}
	}
	replay(t, lifecycle[ct.TagsTemplate]{label: "tags", segment: "ivr-tags-template", noun: "CsTagsTemplate", created: "IvrTagsTemplate",
		createBody: `{"Name":"` + name + `","Description":"` + initial + `","Tags":[{"TagName":"Sales","TagDetails":{"Id":"` + sales + `","Type":"ExternalPstn"}}]}`,
		updateBody: `{"Name":"` + name + `","Description":"` + updated + `","Tags":[` +
			`{"TagName":"Support","TagDetails":{"Id":"` + support + `","Type":"ExternalPstn"}},` +
			`{"TagName":"Billing","TagDetails":{"Id":"` + billing + `","Type":"ExternalPstn"}}]}`,
		templates: func(sdk *msteams.Client) *ct.Templates[ct.TagsTemplate] { return sdk.TeamsAdminAPI.CallTemplates.Tags },
		build: func() *ct.TagsTemplate {
			return &ct.TagsTemplate{Name: ptr(name), Description: ptr(initial), Tags: []ct.Tag{tag("Sales", sales)}}
		},
		edit: func(v *ct.TagsTemplate) {
			v.Description, v.Tags = ptr(updated), []ct.Tag{tag("Support", support), tag("Billing", billing)}
		},
		createdOK: func(v *ct.TagsTemplate) bool {
			return len(v.Tags) == 1 && is(v.Tags[0].TagName, "Sales") && is(v.Tags[0].TagDetails.ID, sales) &&
				is(v.Tags[0].TagDetails.EnableTranscription, false) && len(v.Tags[0].TagDetails.RawFields) == 7
		},
		updatedOK: func(v *ct.TagsTemplate) bool {
			return is(v.Description, updated) && len(v.Tags) == 2 && is(v.Tags[1].TagName, "Billing") && is(v.Tags[1].TagDetails.ID, billing)
		},
		id: func(v *ct.TagsTemplate) *string { return v.ID },
	})
}

func TestUnitAutoRecordingRejectionsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	denied, list, invalid := readFile(t, "testdata/recording_create_401.json"), readFile(t, "testdata/recording_list.json"), readFile(t, "testdata/recording_update_invalid_400.json")
	const noun = "CsAutoRecordingTemplate"
	run(t, testutil.Case{Name: "actual-auto-recording-denial",

		Steps: []testutil.Exchange{discovery(t), wire("POST", "auto-recording", "New-"+noun,
			json.RawMessage(`{"Name":"`+name+`","Description":"`+initial+`","TranscriptionEnabled":false,"RecordingEnabled":true,"AgentViewPermission":0,`+
				`"SharePointHostName":"example.sharepoint.com","SharePointSiteName":"`+name+`","RecordingDocumentOwner":"`+missing+`"}`), 401, denied)},
		Want: testutil.Expected{Status: 401, Body: denied, APIError: apiError(401, denied)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.CallTemplates.AutoRecording.Create(ctx, &ct.AutoRecordingTemplate{Name: ptr(name), Description: ptr(initial),
				TranscriptionEnabled: ptr(false), RecordingEnabled: ptr(true), AgentViewPermission: ptr(ct.AgentViewPermissionNone),
				SharePointHostName: ptr("example.sharepoint.com"), SharePointSiteName: ptr(name), RecordingDocumentOwner: ptr(missing)})
			if value != nil {
				t.Fatal("denied create returned a template")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
	run(t, testutil.Case{Name: "actual-auto-recording-list-and-invalid-update",

		Steps: []testutil.Exchange{discovery(t), wire("GET", "auto-recording", "Get-"+noun, nil, 200, list),
			wire("PUT", "auto-recording/"+missing, "Set-"+noun, json.RawMessage(`{"Name":"`+name+`","TranscriptionEnabled":false,"RecordingEnabled":false}`), 400, invalid)},
		Want: testutil.Expected{Status: 400, Body: invalid, APIError: apiError(400, invalid)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.CallTemplates.AutoRecording
			if values, _, err := s.List(ctx); err != nil || values == nil {
				t.Fatal("empty auto recording list was not a non-nil slice")
			}
			_, response, err := s.UpdateByID(ctx, missing, &ct.AutoRecordingTemplate{Name: ptr(name), TranscriptionEnabled: ptr(false), RecordingEnabled: ptr(false)})
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitMissingTemplateIsNotFound(t *testing.T) {
	t.Parallel()
	body := readFile(t, "testdata/get_missing_404.json")
	run(t, testutil.Case{Name: "actual-missing-template",

		Steps: []testutil.Exchange{discovery(t), wire("GET", "ivr-tags-template/"+missing, "Get-CsTagsTemplate", nil, 404, body)},
		Want:  testutil.Expected{Status: 404, Body: body, APIError: apiError(404, body)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.CallTemplates.Tags.GetByID(ctx, missing)
			if value != nil {
				t.Fatal("missing template returned a value")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
	// The lab observed the same 100004 body for every type's missing reads and deletes.
	run(t, testutil.Case{Name: "actual-missing-auto-recording",

		Steps: []testutil.Exchange{discovery(t), wire("GET", "auto-recording/"+missing, "Get-CsAutoRecordingTemplate", nil, 404, body),
			wire("DELETE", "auto-recording/"+missing, "Remove-CsAutoRecordingTemplate", nil, 404, body)},
		Want: testutil.Expected{Status: 404, Body: body, APIError: apiError(404, body)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.CallTemplates.AutoRecording
			if _, _, err := s.GetByID(ctx, missing); err == nil {
				t.Fatal("missing auto recording template was read")
			}
			response, err := s.DeleteByID(ctx, missing)
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for name, reply := range map[string]string{
		"get-missing-envelope":  `{"Diagnostics":[]}`,
		"get-null-envelope":     `{"IvrTagsTemplate":null}`,
		"get-array":             `[]`,
		"get-wrong-field-type":  `{"IvrTagsTemplate":{"Name":1}}`,
		"get-wrong-tags-type":   `{"IvrTagsTemplate":{"Tags":{}}}`,
		"get-wrong-target-type": `{"IvrTagsTemplate":{"Tags":[{"TagDetails":[]}]}}`,
	} {
		run(t, testutil.Case{Name: name,
			Steps: []testutil.Exchange{discovery(t), wire("GET", "ivr-tags-template/"+missing, "Get-CsTagsTemplate", nil, 200, []byte(reply))},
			Want:  testutil.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, response, err := sdk.TeamsAdminAPI.CallTemplates.Tags.GetByID(ctx, missing)
				return testutil.Observation{Response: response, Err: err}
			}})
	}
	for name, reply := range map[string]string{
		"list-missing-array": `{"Diagnostics":[]}`,
		"list-null-array":    `{"ComplianceRecordings":null}`,
		"list-null-item":     `{"ComplianceRecordings":[null]}`,
		"list-bad-domain":    `{"ComplianceRecordings":[{"BotRequiredDuringCall":"yes"}]}`,
	} {
		run(t, testutil.Case{Name: name,
			Steps: []testutil.Exchange{discovery(t), wire("GET", "compliance-recording", "Get-CsComplianceRecordingForCallQueueTemplate", nil, 200, []byte(reply))},
			Want:  testutil.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, response, err := sdk.TeamsAdminAPI.CallTemplates.ComplianceRecording.List(ctx)
				return testutil.Observation{Response: response, Err: err}
			}})
	}
	run(t, testutil.Case{Name: "undeclared-status",
		Steps: []testutil.Exchange{discovery(t), wire("DELETE", "shared-call-history-template/"+missing, "Remove-CsSharedCallHistoryTemplate", nil, 202, []byte(`{}`))},
		Want:  testutil.Expected{Status: 202, Error: ct.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.TeamsAdminAPI.CallTemplates.SharedCallHistory.DeleteByID(ctx, missing)
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	recording := func() *ct.AutoRecordingTemplate {
		return &ct.AutoRecordingTemplate{Name: ptr("n"), Description: ptr(""), TranscriptionEnabled: ptr(true), RecordingEnabled: ptr(true),
			AgentViewPermission: ptr(ct.AgentViewPermissionAll), SharePointHostName: ptr("h"), SharePointSiteName: ptr("s"), RecordingDocumentOwner: ptr(missing)}
	}
	for name, call := range map[string]func(context.Context, *ct.Service) error{
		"get-blank": func(ctx context.Context, s *ct.Service) error { _, _, e := s.Tags.GetByID(ctx, " "); return e },
		"get-slash": func(ctx context.Context, s *ct.Service) error { _, _, e := s.Tags.GetByID(ctx, "a/b"); return e },
		"get-traversal": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.AutoRecording.GetByID(ctx, "..")
			return e
		},
		"delete-control": func(ctx context.Context, s *ct.Service) error {
			_, e := s.SharedCallHistory.DeleteByID(ctx, "a\nb")
			return e
		},
		"create-nil": func(ctx context.Context, s *ct.Service) error { _, _, e := s.Tags.Create(ctx, nil); return e },
		"create-with-id": func(ctx context.Context, s *ct.Service) error {
			v := recording()
			v.ID = ptr(missing)
			_, _, e := s.AutoRecording.Create(ctx, v)
			return e
		},
		"create-recording-no-owner": func(ctx context.Context, s *ct.Service) error {
			v := recording()
			v.RecordingDocumentOwner = nil
			_, _, e := s.AutoRecording.Create(ctx, v)
			return e
		},
		"create-recording-no-permission": func(ctx context.Context, s *ct.Service) error {
			v := recording()
			v.AgentViewPermission = nil
			_, _, e := s.AutoRecording.Create(ctx, v)
			return e
		},
		"create-compliance-no-bot": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.ComplianceRecording.Create(ctx, &ct.ComplianceRecordingTemplate{Name: ptr("n"), Description: ptr("")})
			return e
		},
		"create-compliance-count": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.ComplianceRecording.Create(ctx, &ct.ComplianceRecordingTemplate{Name: ptr("n"), Description: ptr(""), BotID: ptr(botID), ConcurrentInvitationCount: ptr(int32(3))})
			return e
		},
		"create-history-no-description": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.SharedCallHistory.Create(ctx, &ct.SharedCallHistoryTemplate{Name: ptr("n")})
			return e
		},
		"create-history-value": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.SharedCallHistory.Create(ctx, &ct.SharedCallHistoryTemplate{Name: ptr("n"), Description: ptr(""), IncomingRedirectedCalls: ptr(int32(3))})
			return e
		},
		"create-triage-blank-name": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.SharedVoicemailTriageSettings.Create(ctx, &ct.SharedVoicemailTriageSettingsTemplate{Name: ptr(" "), Description: ptr("")})
			return e
		},
		"create-tags-empty": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.Tags.Create(ctx, &ct.TagsTemplate{Name: ptr("n"), Description: ptr("")})
			return e
		},
		"create-tags-no-target": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.Tags.Create(ctx, &ct.TagsTemplate{Name: ptr("n"), Description: ptr(""), Tags: []ct.Tag{{TagName: ptr("t")}}})
			return e
		},
		"update-nil": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.Tags.UpdateByID(ctx, missing, nil)
			return e
		},
		"update-no-name": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.SharedVoicemailTriageSettings.UpdateByID(ctx, missing, &ct.SharedVoicemailTriageSettingsTemplate{})
			return e
		},
		"update-other-id": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.SharedCallHistory.UpdateByID(ctx, missing, &ct.SharedCallHistoryTemplate{ID: ptr(botID), Name: ptr("n")})
			return e
		},
		"update-compliance-count": func(ctx context.Context, s *ct.Service) error {
			_, _, e := s.ComplianceRecording.UpdateByID(ctx, missing, &ct.ComplianceRecordingTemplate{Name: ptr("n"), ConcurrentInvitationCount: ptr(int32(0))})
			return e
		},
	} {
		run(t, testutil.Case{Name: name, Want: testutil.Expected{Error: ct.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return testutil.Observation{Err: call(ctx, sdk.TeamsAdminAPI.CallTemplates)}
			}})
	}
}

func TestUnitNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var tags *ct.Templates[ct.TagsTemplate]
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := tags.List(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := tags.DeleteByID(t.Context(), missing); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil templates accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := ct.NewService(nil).Tags.GetByID(ctx, missing); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
