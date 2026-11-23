package mainlineattendants_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ma "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/mainline_attendants"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"
	source       = "msteams/teams_admin_api/mainline_attendants/service_test.go"
	flows        = "/Teams.VoiceApps/mainline-attendant-flow"
	spamID       = "00000000-0000-4000-8000-0000000000b1"
	bookingID    = "00000000-0000-4000-8000-0000000000b2"
	qaID         = "00000000-0000-4000-8000-0000000000b3"
	probeID      = "00000000-0000-4000-8000-000000000001"
	jsonType     = "application/json; charset=utf-8"
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

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply []byte) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	contentType := jsonType
	if status >= 400 {
		contentType = "application/json"
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "mainline_attendants"
	if c.Source == "" {
		c.Source = source
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("mainline attendant contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(commands ...string) []contracttest.Mapping {
	out := make([]contracttest.Mapping, 0, len(commands))
	for _, command := range commands {
		out = append(out, contracttest.Mapping{Command: command, ParameterSet: "__AllParameterSets"})
	}
	return out
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

func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func TestUnitMainlineAttendantReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	information := readFile(t, "testdata/tenant_information.json")
	run(t, contracttest.Case{Name: "tenant-information", Source: "msteams/teams_admin_api/mainline_attendants/testdata/tenant_information.json",
		Mappings: mapping("Get-CsMainlineAttendantTenantInformation", "Get-CsMainlineAttendantSupportedLanguages", "Get-CsMainlineAttendantSupportedVoices"),
		Steps:    []contracttest.Exchange{discovery(t), wire("GET", "/Teams.VoiceApps/information", "Get-CsMainlineAttendantTenantInformation", nil, 200, information)},
		Want:     contracttest.Expected{Status: 200, Body: information},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.MainlineAttendants.GetTenantInformation(ctx)
			if err != nil || text(value.DefaultLanguage) != "en-US" || text(value.DefaultTimeZone) != "UTC" || value.DefaultVoice == nil ||
				len(value.SupportedLanguages) == 0 || len(value.SupportedVoices) == 0 || value.SupportedLanguages[0].ID == nil || value.SupportedVoices[0].ID == nil {
				t.Fatal("mainline attendant tenant information was not decoded")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})

	empty := readFile(t, "testdata/flows.json")
	templates := readFile(t, "testdata/spam_detection_templates.json")
	run(t, contracttest.Case{Name: "lists", Source: "msteams/teams_admin_api/mainline_attendants/testdata/flows.json",
		Mappings: mapping("Get-CsMainlineAttendantFlow", "Get-CsMainlineAttendantAppointmentBookingFlow", "Get-CsMainlineAttendantQuestionAnswerFlow", "Get-CsMainlineAttendantSpamDetectionTemplate"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", flows, "Get-CsMainlineAttendantFlow", nil, 200, empty),
			wire("GET", flows+"/appointment-booking", "Get-CsMainlineAttendantAppointmentBookingFlow", nil, 200, readFile(t, "testdata/appointment_booking_flows.json")),
			wire("GET", flows+"/question-answer?First=5&NameFilter=msteams&Skip=1", "Get-CsMainlineAttendantQuestionAnswerFlow", nil, 200, readFile(t, "testdata/question_answer_flows.json")),
			wire("GET", flows+"/spam-detection-template?Descending=True&First=50&SortBy=Name", "Get-CsMainlineAttendantSpamDetectionTemplate", nil, 200, templates)},
		Want: contracttest.Expected{Status: 200, Body: templates},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.MainlineAttendants
			if values, _, err := s.ListFlows(ctx, nil); err != nil || values == nil {
				t.Fatal("flow list was not a non-nil slice")
			}
			if values, _, err := s.ListAppointmentBookingFlows(ctx, &ma.ListOptions{}); err != nil || values == nil {
				t.Fatal("appointment booking list failed")
			}
			if _, _, err := s.ListQuestionAnswerFlows(ctx, &ma.ListOptions{First: new(int32(5)), Skip: new(int32(1)), NameFilter: new("msteams")}); err != nil {
				t.Fatal("question answer list options were not sent")
			}
			values, response, err := s.ListSpamDetectionTemplates(ctx, &ma.ListOptions{First: new(int32(50)), SortBy: new("Name"), Descending: true})
			if err != nil || values == nil {
				t.Fatal("spam detection template list failed")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})

	filtered := readFile(t, "testdata/flows_filtered.json")
	run(t, contracttest.Case{Name: "flow-filters", Source: "msteams/teams_admin_api/mainline_attendants/testdata/flows_filtered.json",
		Mappings: mapping("Get-CsMainlineAttendantFlow"),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", flows+"?NameFilter=fixture-qa&relatedConfigurationId="+probeID+"&type=QuestionAnswer",
			"Get-CsMainlineAttendantFlow", nil, 200, filtered)},
		Want: contracttest.Expected{Status: 200, Body: filtered},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, response, err := sdk.TeamsAdminAPI.MainlineAttendants.ListFlows(ctx, &ma.FlowListOptions{Type: new(ma.FlowTypeQuestionAnswer),
				RelatedConfigurationID: new(probeID), ListOptions: ma.ListOptions{NameFilter: new("fixture-qa")}})
			if err != nil || len(values) != 1 || text(values[0].Identity) != qaID || text(values[0].Type) != ma.FlowTypeQuestionAnswer {
				t.Fatal("filtered flow list was not decoded")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitMainlineAttendantMissingResourcesAreTypedErrors(t *testing.T) {
	t.Parallel()
	flowMissing := readFile(t, "testdata/flow_missing_404.json")
	templateMissing := readFile(t, "testdata/spam_detection_template_missing_404.json")
	deleted := readFile(t, "testdata/spam_detection_template_delete_404.json")
	for name, item := range map[string]struct {
		path, command string
		method        string
		body          []byte
		invoke        func(context.Context, *ma.Service) (*resty.Response, error)
	}{
		"flow": {flows + "/" + probeID, "Get-CsMainlineAttendantFlow", "GET", flowMissing, func(ctx context.Context, s *ma.Service) (*resty.Response, error) {
			_, r, e := s.GetFlowByID(ctx, probeID)
			return r, e
		}},
		"template": {flows + "/spam-detection-template/" + probeID, "Get-CsMainlineAttendantSpamDetectionTemplate", "GET", templateMissing, func(ctx context.Context, s *ma.Service) (*resty.Response, error) {
			_, r, e := s.GetSpamDetectionTemplateByID(ctx, probeID)
			return r, e
		}},
		"repeat-delete": {flows + "/spam-detection-template/" + spamID, "Remove-CsMainlineAttendantSpamDetectionTemplate", "DELETE", deleted, func(ctx context.Context, s *ma.Service) (*resty.Response, error) {
			return s.DeleteSpamDetectionTemplateByID(ctx, spamID)
		}},
		"booking-delete": {flows + "/appointment-booking/" + bookingID, "Remove-CsMainlineAttendantAppointmentBookingFlow", "DELETE", flowMissing, func(ctx context.Context, s *ma.Service) (*resty.Response, error) {
			return s.DeleteAppointmentBookingFlowByID(ctx, bookingID)
		}},
	} {
		run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{discovery(t), wire(item.method, item.path, item.command, nil, 404, item.body)},
			Want: contracttest.Expected{Status: 404, Body: item.body, APIError: apiError(404, item.body)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := item.invoke(ctx, sdk.TeamsAdminAPI.MainlineAttendants)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	rejected := []byte(`{"code":"100001","message":"Fixture message","action":"Please input correct parameter values.","errorDetails":[{"ErrorCode":"MainlineAttendantQuestionAnswerFlow_KnowledgeBaseNullOrEmpty","Message":"Fixture message","Target":"KnowledgeBase","TargetEntityId":""}]}`)
	run(t, contracttest.Case{Name: "update-without-knowledge-base", Mappings: mapping("Set-CsMainlineAttendantQuestionAnswerFlow"),
		Steps: []contracttest.Exchange{discovery(t), wire("PUT", flows+"/question-answer/"+qaID, "Set-CsMainlineAttendantQuestionAnswerFlow",
			json.RawMessage(`{"Type":"QuestionAnswer","Description":"x"}`), 400, rejected)},
		Want: contracttest.Expected{Status: 400, Body: rejected, APIError: apiError(400, rejected)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.MainlineAttendants.UpdateQuestionAnswerFlowByID(ctx, qaID, ma.UpdateQuestionAnswerFlowRequest{Description: new("x")})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
}

func TestUnitMainlineAttendantSpamDetectionTemplateLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	created, updated, got, deleted := readFile(t, "testdata/spam_detection_template_create_201.json"), readFile(t, "testdata/spam_detection_template_update.json"),
		readFile(t, "testdata/spam_detection_template_get.json"), readFile(t, "testdata/spam_detection_template_delete.json")
	item := flows + "/spam-detection-template/" + spamID
	run(t, contracttest.Case{Name: "spam-lifecycle", Source: "msteams/teams_admin_api/mainline_attendants/testdata/spam_detection_template_create_201.json",
		Mappings: mapping("New-CsMainlineAttendantSpamDetectionTemplate", "Set-CsMainlineAttendantSpamDetectionTemplate", "Get-CsMainlineAttendantSpamDetectionTemplate", "Remove-CsMainlineAttendantSpamDetectionTemplate"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", flows+"/spam-detection-template", "New-CsMainlineAttendantSpamDetectionTemplate",
				json.RawMessage(`{"Name":"Fixture name","Description":"Fixture description","EnableSpamDetection":true,"Action":"DisconnectCall"}`), 201, created),
			wire("PUT", item, "Set-CsMainlineAttendantSpamDetectionTemplate",
				json.RawMessage(`{"Name":"Fixture name","Description":"Fixture description","EnableSpamDetection":false,"Action":"DisconnectCall","RelatedConfigurationIds":[]}`), 200, updated),
			wire("GET", item, "Get-CsMainlineAttendantSpamDetectionTemplate", nil, 200, got),
			wire("DELETE", item, "Remove-CsMainlineAttendantSpamDetectionTemplate", nil, 200, deleted)},
		Want: contracttest.Expected{Status: 200, Body: deleted},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.MainlineAttendants
			value, response, err := s.CreateSpamDetectionTemplate(ctx, ma.CreateSpamDetectionTemplateRequest{Name: "Fixture name", Description: "Fixture description"})
			if err != nil || response.StatusCode() != 201 || text(value.Identity) != spamID || value.EnableSpamDetection == nil || !*value.EnableSpamDetection ||
				text(value.Action) != ma.SpamActionDisconnectCall || value.CallTarget != nil || value.InclusionScope != nil || value.RelatedConfigurationIDs == nil {
				t.Fatal("spam detection template create receipt changed")
			}
			value, _, err = s.UpdateSpamDetectionTemplateByID(ctx, spamID, ma.UpdateSpamDetectionTemplateRequest{Name: value.Name, Description: value.Description,
				EnableSpamDetection: new(false), Action: value.Action, RelatedConfigurationIDs: []ma.RelatedConfiguration{}})
			if err != nil || *value.EnableSpamDetection {
				t.Fatal("spam detection template update was not decoded")
			}
			value, _, err = s.GetSpamDetectionTemplateByID(ctx, spamID)
			if err != nil || *value.EnableSpamDetection || len(value.RawFields) != 9 {
				t.Fatal("spam detection template read-back changed")
			}
			response, err = s.DeleteSpamDetectionTemplateByID(ctx, spamID)
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitMainlineAttendantSpamDetectionTemplateSendsCallTargetAsTheModule(t *testing.T) {
	t.Parallel()
	created := readFile(t, "testdata/spam_detection_template_create_201.json")
	body := `{"Name":"n","Description":"d","EnableSpamDetection":false,"Action":"TransferCallToTarget","InclusionScope":["a"],"ExclusionScope":[],` +
		`"CallTarget":{"Id":"` + probeID + `","Type":"ApplicationEndpoint","EnableTranscription":true,"CallPriority":2}}`
	other := `{"Name":"n","CallTarget":{"Id":"x","Type":"User","SharedVoicemailHistoryTemplateId":"h"}}`
	run(t, contracttest.Case{Name: "call-target", Mappings: mapping("New-CsMainlineAttendantSpamDetectionTemplate", "Set-CsMainlineAttendantSpamDetectionTemplate"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", flows+"/spam-detection-template", "New-CsMainlineAttendantSpamDetectionTemplate", json.RawMessage(body), 201, created),
			wire("PUT", flows+"/spam-detection-template/"+spamID, "Set-CsMainlineAttendantSpamDetectionTemplate", json.RawMessage(other), 200, readFile(t, "testdata/spam_detection_template_update.json"))},
		Want: contracttest.Expected{Status: 200, Body: readFile(t, "testdata/spam_detection_template_update.json")},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.MainlineAttendants
			if _, _, err := s.CreateSpamDetectionTemplate(ctx, ma.CreateSpamDetectionTemplateRequest{Name: "n", Description: "d", EnableSpamDetection: new(false),
				Action: new(ma.SpamActionTransferCallToTarget), InclusionScope: []string{"a"}, ExclusionScope: []string{},
				CallTarget: &ma.CallTarget{ID: new(probeID), Type: new(ma.CallTargetApplicationEndpoint), EnableTranscription: new(true),
					EnableSharedVoicemailSystemPromptSuppression: new(false), CallPriority: new(int32(2))}}); err != nil {
				t.Fatal("call target create was not sent as the module sends it")
			}
			value, response, err := s.UpdateSpamDetectionTemplateByID(ctx, spamID, ma.UpdateSpamDetectionTemplateRequest{Name: new("n"),
				CallTarget: &ma.CallTarget{ID: new("x"), Type: new("User"), CallPriority: new(int32(1)), SharedVoicemailHistoryTemplateID: new("h")}})
			return contracttest.Observation{Value: value != nil, Response: response, Err: err}
		}})
}

func TestUnitMainlineAttendantAppointmentBookingFlowLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	created, updated, got, deleted := readFile(t, "testdata/appointment_booking_flow_create_201.json"), readFile(t, "testdata/appointment_booking_flow_update.json"),
		readFile(t, "testdata/appointment_booking_flow_get.json"), readFile(t, "testdata/appointment_booking_flow_delete.json")
	item := flows + "/appointment-booking/" + bookingID
	const link = "https://outlook.office.com/book/example/"
	run(t, contracttest.Case{Name: "booking-lifecycle", Source: "msteams/teams_admin_api/mainline_attendants/testdata/appointment_booking_flow_create_201.json",
		Mappings: mapping("New-CsMainlineAttendantAppointmentBookingFlow", "Set-CsMainlineAttendantAppointmentBookingFlow", "Get-CsMainlineAttendantAppointmentBookingFlow", "Remove-CsMainlineAttendantAppointmentBookingFlow"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", flows+"/appointment-booking", "New-CsMainlineAttendantAppointmentBookingFlow",
				json.RawMessage(`{"Name":"n","Description":"d","AppointmentBookingPlatform":"MicrosoftBookings","BookingLink":"`+link+`"}`), 201, created),
			wire("PUT", item, "Set-CsMainlineAttendantAppointmentBookingFlow",
				json.RawMessage(`{"Name":"n","Type":"AppointmentBooking","RelatedConfigurationIds":[{"Id":"r","Name":"AA"}],"Description":"u","AppointmentBookingPlatform":"MicrosoftBookings","BookingLink":"`+link+`"}`), 200, updated),
			wire("GET", item, "Get-CsMainlineAttendantAppointmentBookingFlow", nil, 200, got),
			wire("DELETE", item, "Remove-CsMainlineAttendantAppointmentBookingFlow", nil, 200, deleted)},
		Want: contracttest.Expected{Status: 200, Body: deleted},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.MainlineAttendants
			platform := ma.BookingPlatformMicrosoftBookings
			value, _, err := s.CreateAppointmentBookingFlow(ctx, ma.CreateAppointmentBookingFlowRequest{Name: "n", Description: "d", AppointmentBookingPlatform: &platform, BookingLink: new(link)})
			if err != nil || text(value.Identity) != bookingID || text(value.Type) != ma.FlowTypeAppointmentBooking || value.CallerAuthenticationMethod != nil ||
				text(value.APIDefinitions) != "{}" || value.KnowledgeBase != nil {
				t.Fatal("appointment booking create receipt changed")
			}
			if _, _, err = s.UpdateAppointmentBookingFlowByID(ctx, bookingID, ma.UpdateAppointmentBookingFlowRequest{Name: new("n"), Description: new("u"),
				RelatedConfigurationIDs: []ma.RelatedConfiguration{{ID: new("r"), Name: new("AA")}}, AppointmentBookingPlatform: &platform, BookingLink: new(link)}); err != nil {
				t.Fatal("appointment booking update failed")
			}
			if value, _, err = s.GetAppointmentBookingFlowByID(ctx, bookingID); err != nil || text(value.AppointmentBookingPlatform) != platform || value.RawFields["KnowledgeBase"] == nil {
				t.Fatal("appointment booking read-back changed")
			}
			response, err := s.DeleteAppointmentBookingFlowByID(ctx, bookingID)
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitMainlineAttendantCustomBookingSendsEnumsAsIntegers(t *testing.T) {
	t.Parallel()
	created := readFile(t, "testdata/appointment_booking_flow_create_201.json")
	definitions := `{"AuthenticateCaller":{"url":"https://example.invalid/a"}}`
	encoded, _ := json.Marshal(definitions)
	run(t, contracttest.Case{Name: "custom-booking", Mappings: mapping("New-CsMainlineAttendantAppointmentBookingFlow"),
		Steps: []contracttest.Exchange{discovery(t), wire("POST", flows+"/appointment-booking", "New-CsMainlineAttendantAppointmentBookingFlow",
			json.RawMessage(`{"Name":"n","Description":"d","AppointmentBookingPlatform":"Custom","CallerAuthenticationMethod":1,"ApiAuthenticationType":3,"ApiDefinitions":`+string(encoded)+`}`), 201, created)},
		Want: contracttest.Expected{Status: 201, Body: created},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.MainlineAttendants.CreateAppointmentBookingFlow(ctx, ma.CreateAppointmentBookingFlowRequest{Name: "n", Description: "d",
				AppointmentBookingPlatform: new(ma.BookingPlatformCustom), CallerAuthenticationMethod: new(ma.CallerAuthenticationEmail),
				APIAuthenticationType: new(ma.APIAuthenticationBearerTokenDynamic), APIDefinitions: new(definitions)})
			return contracttest.Observation{Value: value != nil, Response: response, Err: err}
		}})
}

func TestUnitMainlineAttendantQuestionAnswerFlowLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	created, updated, got := readFile(t, "testdata/question_answer_flow_create_201.json"), readFile(t, "testdata/question_answer_flow_update.json"), readFile(t, "testdata/question_answer_flow_get.json")
	deleted, gone := readFile(t, "testdata/question_answer_flow_delete.json"), readFile(t, "testdata/question_answer_flow_get_404.json")
	item := flows + "/question-answer/" + qaID
	knowledgeBase := `{"file_descriptions":[{"file_name":"faq.txt","file_content":"` + base64.StdEncoding.EncodeToString([]byte("Hours")) +
		`","file_extension":".txt","file_description":"FAQ file"}],"url_descriptions":[{"url_link":"https://www.example.com/faq","url_description":"FAQ page"}]}`
	createBody, _ := json.Marshal(map[string]any{"Name": "n", "Description": "d", "KnowledgeBase": knowledgeBase, "ApiAuthenticationType": 1})
	var flow struct {
		Flow struct {
			KnowledgeBase string `json:"KnowledgeBase"`
		} `json:"MainlineAttendantQuestionAnswerFlow"`
	}
	if json.Unmarshal(created, &flow) != nil {
		t.Fatal("fixture unavailable")
	}
	updateBody, _ := json.Marshal(map[string]any{"Name": "n", "Type": "QuestionAnswer", "Description": "u", "KnowledgeBase": flow.Flow.KnowledgeBase})
	run(t, contracttest.Case{Name: "qa-lifecycle", Source: "msteams/teams_admin_api/mainline_attendants/testdata/question_answer_flow_create_201.json",
		Mappings: mapping("New-CsMainlineAttendantQuestionAnswerFlow", "Set-CsMainlineAttendantQuestionAnswerFlow", "Get-CsMainlineAttendantFlow", "Remove-CsMainlineAttendantQuestionAnswerFlow"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", flows+"/question-answer", "New-CsMainlineAttendantQuestionAnswerFlow", createBody, 201, created),
			wire("PUT", item, "Set-CsMainlineAttendantQuestionAnswerFlow", updateBody, 200, updated),
			wire("GET", flows+"/"+qaID, "Get-CsMainlineAttendantFlow", nil, 200, got),
			wire("DELETE", item, "Remove-CsMainlineAttendantQuestionAnswerFlow", nil, 200, deleted),
			wire("GET", flows+"/"+qaID, "Get-CsMainlineAttendantFlow", nil, 404, gone)},
		Want: contracttest.Expected{Status: 404, Body: gone, APIError: apiError(404, gone)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.MainlineAttendants
			value, _, err := s.CreateQuestionAnswerFlow(ctx, ma.CreateQuestionAnswerFlowRequest{Name: "n", Description: "d", APIAuthenticationType: new(ma.APIAuthenticationAPIKey),
				KnowledgeBase: ma.KnowledgeBase{Files: []ma.KnowledgeBaseFile{{Name: "faq.txt", Content: []byte("Hours"), Description: "FAQ file"}},
					URLs: []ma.KnowledgeBaseURL{{URL: "https://www.example.com/faq", Description: "FAQ page"}}}})
			if err != nil || text(value.Identity) != qaID || text(value.Type) != ma.FlowTypeQuestionAnswer {
				t.Fatal("question answer create receipt changed")
			}
			if indexing, err := value.Indexing(); err != nil || !indexing {
				t.Fatal("new knowledge base was not reported as indexing")
			}
			_, _, err = s.UpdateQuestionAnswerFlowByID(ctx, qaID, ma.UpdateQuestionAnswerFlowRequest{Name: new("n"), Description: new("u"), KnowledgeBase: value.KnowledgeBase})
			if err != nil {
				t.Fatal("question answer update failed")
			}
			value, _, err = s.GetFlowByID(ctx, qaID)
			documents, derr := value.KnowledgeBaseDocuments()
			if err != nil || derr != nil || len(documents) != 1 || text(documents[0].Status) != ma.KnowledgeBaseStatusFailed || documents[0].Error == nil || documents[0].Error.Code == nil {
				t.Fatal("question answer knowledge base status changed")
			}
			if indexing, err := value.Indexing(); err != nil || indexing {
				t.Fatal("failed indexing was reported as in progress")
			}
			if _, err = s.DeleteQuestionAnswerFlowByID(ctx, qaID); err != nil {
				t.Fatal("question answer delete failed")
			}
			value, response, err := s.GetFlowByID(ctx, qaID)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
}

func TestUnitMainlineAttendantMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{
		"null":          `null`,
		"missing":       `{"Diagnostics":[]}`,
		"null-list":     `{"MainlineAttendantFlowResponses":null}`,
		"null-item":     `{"MainlineAttendantFlowResponses":[null]}`,
		"bad-type":      `{"MainlineAttendantFlowResponses":[{"Name":7}]}`,
		"bad-related":   `{"MainlineAttendantFlowResponses":[{"RelatedConfigurationIds":[null]}]}`,
		"array-element": `{"MainlineAttendantFlowResponses":[[]]}`,
	} {
		run(t, contracttest.Case{Name: "list-" + name, Steps: []contracttest.Exchange{discovery(t), wire("GET", flows, "Get-CsMainlineAttendantFlow", nil, 200, []byte(body))},
			Want: contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				values, response, err := sdk.TeamsAdminAPI.MainlineAttendants.ListFlows(ctx, nil)
				if values != nil {
					t.Fatal("malformed list returned values")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	for name, body := range map[string]string{
		"information-missing": `{"TenantInformation":{}}`,
		"information-null":    `{"MainlineAttendantTenantInformation":null}`,
		"voice-null":          `{"MainlineAttendantTenantInformation":{"SupportedVoices":[null]}}`,
	} {
		run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{discovery(t), wire("GET", "/Teams.VoiceApps/information", "Get-CsMainlineAttendantTenantInformation", nil, 200, []byte(body))},
			Want: contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.MainlineAttendants.GetTenantInformation(ctx)
				if value != nil {
					t.Fatal("malformed tenant information returned a value")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	run(t, contracttest.Case{Name: "template-wrong-envelope", Steps: []contracttest.Exchange{discovery(t),
		wire("GET", flows+"/spam-detection-template/"+spamID, "Get-CsMainlineAttendantSpamDetectionTemplate", nil, 200, []byte(`{"MainlineAttendantFlowResponse":{}}`))},
		Want: contracttest.Expected{Status: 200, Error: client.ErrDecode},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.MainlineAttendants.GetSpamDetectionTemplateByID(ctx, spamID)
			if value != nil {
				t.Fatal("malformed template returned a value")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "undeclared-status", Steps: []contracttest.Exchange{discovery(t),
		wire("DELETE", flows+"/question-answer/"+qaID, "Remove-CsMainlineAttendantQuestionAnswerFlow", nil, 202, nil)},
		Want: contracttest.Expected{Status: 202, Error: ma.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.MainlineAttendants.DeleteQuestionAnswerFlowByID(ctx, qaID)
			return contracttest.Observation{Response: response, Err: err}
		}})
	flow := ma.Flow{KnowledgeBase: new("not json")}
	if _, err := flow.KnowledgeBaseDocuments(); !errors.Is(err, client.ErrDecode) {
		t.Fatal("malformed knowledge base status was accepted")
	}
}

func TestUnitMainlineAttendantInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("é", 501)
	for name, call := range map[string]func(context.Context, *ma.Service) error{
		"get-empty": func(ctx context.Context, s *ma.Service) error { _, _, e := s.GetFlowByID(ctx, ""); return e },
		"get-slash": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.GetQuestionAnswerFlowByID(ctx, "a/b")
			return e
		},
		"get-dot": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.GetAppointmentBookingFlowByID(ctx, "..")
			return e
		},
		"get-control": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.GetSpamDetectionTemplateByID(ctx, "a\nb")
			return e
		},
		"delete-blank": func(ctx context.Context, s *ma.Service) error {
			_, e := s.DeleteQuestionAnswerFlowByID(ctx, " ")
			return e
		},
		"list-negative": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.ListSpamDetectionTemplates(ctx, &ma.ListOptions{First: new(int32(-1))})
			return e
		},
		"flows-blank-type": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.ListFlows(ctx, &ma.FlowListOptions{Type: new(" ")})
			return e
		},
		"booking-no-name": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateAppointmentBookingFlow(ctx, ma.CreateAppointmentBookingFlowRequest{Description: "d"})
			return e
		},
		"booking-bad-definitions": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateAppointmentBookingFlow(ctx, ma.CreateAppointmentBookingFlowRequest{Name: "n", Description: "d", APIDefinitions: new("{")})
			return e
		},
		"booking-update-bad-definitions": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.UpdateAppointmentBookingFlowByID(ctx, bookingID, ma.UpdateAppointmentBookingFlowRequest{APIDefinitions: new("x")})
			return e
		},
		"qa-empty-knowledge-base": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateQuestionAnswerFlow(ctx, ma.CreateQuestionAnswerFlowRequest{Name: "n", Description: "d"})
			return e
		},
		"qa-file-path": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateQuestionAnswerFlow(ctx, ma.CreateQuestionAnswerFlowRequest{Name: "n", Description: "d",
				KnowledgeBase: ma.KnowledgeBase{Files: []ma.KnowledgeBaseFile{{Name: "dir/faq.txt", Content: []byte("x")}}}})
			return e
		},
		"qa-empty-file": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateQuestionAnswerFlow(ctx, ma.CreateQuestionAnswerFlowRequest{Name: "n", Description: "d",
				KnowledgeBase: ma.KnowledgeBase{Files: []ma.KnowledgeBaseFile{{Name: "faq.txt"}}}})
			return e
		},
		"qa-blank-url": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateQuestionAnswerFlow(ctx, ma.CreateQuestionAnswerFlowRequest{Name: "n", Description: "d",
				KnowledgeBase: ma.KnowledgeBase{URLs: []ma.KnowledgeBaseURL{{URL: " "}}}})
			return e
		},
		"spam-no-description": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateSpamDetectionTemplate(ctx, ma.CreateSpamDetectionTemplateRequest{Name: "n"})
			return e
		},
		"spam-long-description": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateSpamDetectionTemplate(ctx, ma.CreateSpamDetectionTemplateRequest{Name: "n", Description: long})
			return e
		},
		"spam-update-long-description": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.UpdateSpamDetectionTemplateByID(ctx, spamID, ma.UpdateSpamDetectionTemplateRequest{Description: &long})
			return e
		},
		"spam-blank-scope": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateSpamDetectionTemplate(ctx, ma.CreateSpamDetectionTemplateRequest{Name: "n", Description: "d", ExclusionScope: []string{""}})
			return e
		},
		"spam-blank-action": func(ctx context.Context, s *ma.Service) error {
			_, _, e := s.CreateSpamDetectionTemplate(ctx, ma.CreateSpamDetectionTemplateRequest{Name: "n", Description: "d", Action: new("")})
			return e
		},
	} {
		run(t, contracttest.Case{Name: name, Scope: "local", Want: contracttest.Expected{Error: ma.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: call(ctx, sdk.TeamsAdminAPI.MainlineAttendants)}
			}})
	}
}

func TestUnitMainlineAttendantEncodeKnowledgeBaseMatchesTheModule(t *testing.T) {
	t.Parallel()
	got, err := ma.EncodeKnowledgeBase(ma.KnowledgeBase{URLs: []ma.KnowledgeBaseURL{{URL: "https://www.example.com/faq", Description: "FAQ page"}}})
	if err != nil || got != `{"file_descriptions":[],"url_descriptions":[{"url_link":"https://www.example.com/faq","url_description":"FAQ page"}]}` {
		t.Fatal("URL-only knowledge base encoding changed")
	}
	got, err = ma.EncodeKnowledgeBase(ma.KnowledgeBase{Files: []ma.KnowledgeBaseFile{{Name: "Guide.PDF", Content: []byte{0xff, 0x00}}}})
	if err != nil || got != `{"file_descriptions":[{"file_name":"Guide.PDF","file_content":"/wA=","file_extension":".PDF","file_description":""}],"url_descriptions":[]}` {
		t.Fatal("file knowledge base encoding changed")
	}
	if (ma.KnowledgeBaseFile{Content: []byte("secret")}).String() != "knowledge base file (redacted)" {
		t.Fatal("knowledge base file formatting is not redacted")
	}
}

func TestUnitMainlineAttendantNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *ma.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.GetTenantInformation(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.DeleteSpamDetectionTemplateByID(t.Context(), spamID); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	var flow *ma.Flow
	if documents, err := flow.KnowledgeBaseDocuments(); documents != nil || err != nil {
		t.Fatal("nil flow knowledge base was not empty")
	}
}
