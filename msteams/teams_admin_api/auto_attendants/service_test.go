package autoattendants_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin  = "https://api.interfaces.records.teams.microsoft.com"
	targetForest  = "https://admin.example.invalid/"
	source        = "msteams/teams_admin_api/auto_attendants/service_test.go"
	probe         = "00000000-0000-4000-8000-000000000001"
	attendantID   = "00000000-0000-4000-8000-000000000003"
	defaultFlowID = "00000000-0000-4000-8000-000000000004"
	afterHoursID  = "00000000-0000-4000-8000-000000000005"
	fixtureName   = "msteams-sdk-fixture-attendant"
	pstn          = "tel:+14255550100"
	// redactedTarget is how the fixture writer records the PSTN target ID.
	redactedTarget = "fixture-record-d480caa6"
	jsonType       = "application/json; charset=utf-8"
	attendants     = "/Teams.VoiceApps/auto-attendants"
	information    = "/Teams.VoiceApps/information"
	languagesRoute = "/Teams.VoiceApps/supported-languages/"
	timeZonesRoute = "/Teams.VoiceApps/supported-timezones/"
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
	if reply == nil {
		contentType = ""
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "auto_attendants"
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
		t.Fatalf("auto attendant contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(commands ...string) []contracttest.Mapping {
	out := make([]contracttest.Mapping, 0, len(commands))
	for _, command := range commands {
		out = append(out, contracttest.Mapping{Command: command, ParameterSet: "__AllParameterSets"})
	}
	return out
}

// apiError is the normalized error for a fixture: ConfigAPI reports the code
// at the top level.
func apiError(status int, body []byte) *client.APIError {
	var envelope struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(body, &envelope)
	result := &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}
	if envelope.Code != "" {
		result.Codes = append(result.Codes, envelope.Code)
	}
	return result
}

func ptr[T any](v T) *T { return &v }

func marshal(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func greeting(text string) aa.Prompt {
	return aa.Prompt{ActiveType: ptr(aa.PromptTextToSpeech), TextToSpeechPrompt: &text}
}

// lifecycleSettings is the create body the acceptance test sends: two call
// flows with caller-chosen IDs, no resource account and no schedule.
func lifecycleSettings() aa.Settings {
	transfer := aa.MenuOption{Action: ptr(aa.ActionTransferCallToTarget), DtmfResponse: ptr(aa.DtmfTone1),
		CallTarget: &aa.CallableEntity{ID: ptr(pstn), Type: ptr(aa.EntityExternalPstn)}}
	disconnect := aa.MenuOption{Action: ptr(aa.ActionDisconnectCall), DtmfResponse: ptr(aa.DtmfAutomatic)}
	return aa.Settings{Name: ptr(fixtureName), LanguageID: ptr("en-US"), TimeZoneID: ptr("UTC"),
		DefaultCallFlow: &aa.CallFlow{ID: ptr(defaultFlowID), Name: ptr("Default call flow"), Greetings: []aa.Prompt{greeting("Welcome to the SDK probe")},
			Menu: &aa.Menu{Name: ptr("Default menu"), Prompts: []aa.Prompt{greeting("Press 1 for the SDK probe line")}, MenuOptions: []aa.MenuOption{transfer},
				DirectorySearchMethod: ptr(aa.SearchNone)}},
		CallFlows: &[]aa.CallFlow{{ID: ptr(afterHoursID), Name: ptr("After hours call flow"),
			Menu: &aa.Menu{Name: ptr("After hours menu"), MenuOptions: []aa.MenuOption{disconnect}}}}}
}

func TestUnitAutoAttendantLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	created, got, listed := readFile(t, "testdata/create_201.json"), readFile(t, "testdata/get.json"), readFile(t, "testdata/list.json")
	status, updated, refreshed, deleted := readFile(t, "testdata/status.json"), readFile(t, "testdata/update.json"), readFile(t, "testdata/refresh.json"), readFile(t, "testdata/delete.json")
	settings := lifecycleSettings()
	// The update body is the read-back's settings with a new time zone and greeting.
	var read struct{ AutoAttendant aa.AutoAttendant }
	if err := json.Unmarshal(got, &read); err != nil {
		t.Fatal(err)
	}
	change := func(s aa.Settings) aa.Settings {
		s.TimeZoneID = ptr("GMT Standard Time")
		flow := *s.DefaultCallFlow
		flow.Greetings = []aa.Prompt{greeting("Updated SDK probe greeting")}
		s.DefaultCallFlow = &flow
		return s
	}
	updateBody := marshal(t, change(read.AutoAttendant.Settings()))
	run(t, contracttest.Case{Name: "actual-lifecycle", Source: "msteams/teams_admin_api/auto_attendants/testdata/get.json",
		Mappings: mapping("New-CsAutoAttendant", "Get-CsAutoAttendant", "Get-CsAutoAttendantStatus", "Set-CsAutoAttendant", "Update-CsAutoAttendant", "Remove-CsAutoAttendant"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", attendants, "New-CsAutoAttendant", marshal(t, settings), 201, created),
			wire("GET", attendants+"/"+attendantID, "Get-CsAutoAttendant", nil, 200, got),
			wire("GET", attendants+"/status/"+attendantID, "Get-CsAutoAttendantStatus", nil, 200, status),
			wire("GET", attendants+"?IncludeStatus=True&NameFilter="+fixtureName, "Get-CsAutoAttendant", nil, 200, listed),
			wire("PUT", attendants+"/"+attendantID, "Set-CsAutoAttendant", updateBody, 200, updated),
			wire("POST", attendants+"/"+attendantID+"/refresh", "Update-CsAutoAttendant", nil, 200, refreshed),
			wire("DELETE", attendants+"/"+attendantID, "Remove-CsAutoAttendant", nil, 200, deleted)},
		Want: contracttest.Expected{Status: 200, Body: deleted},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.AutoAttendants
			value, response, err := s.Create(ctx, settings)
			if err != nil || response.StatusCode() != 201 || *value.ID != attendantID || *value.DefaultCallFlow.ID != defaultFlowID ||
				*value.CallFlows[0].ID != afterHoursID || *value.VoiceID != "Male" {
				t.Fatal("create receipt was not decoded")
			}
			value, _, err = s.GetByID(ctx, attendantID)
			if err != nil || *value.LanguageID != "en-US" || len(value.RawFields) != 29 || value.RawFields["AutoRecordingThread"] == nil ||
				*value.DefaultCallFlow.Menu.MenuOptions[0].CallTarget.ID != redactedTarget || *value.DirectoryLookupScope.InclusionScope.Type != aa.DialScopeDefault {
				t.Fatal("get lost fields")
			}
			records, _, err := s.GetStatusByID(ctx, attendantID, nil)
			if err != nil || records == nil || len(records) != 0 {
				t.Fatal("empty status was not a non-nil empty slice")
			}
			values, _, err := s.List(ctx, &aa.ListOptions{NameFilter: ptr(fixtureName), IncludeStatus: ptr(true)})
			if err != nil || len(values) != 1 || *values[0].ID != attendantID {
				t.Fatal("list was not decoded")
			}
			result, _, err := s.UpdateByID(ctx, attendantID, change(value.Settings()))
			if err != nil || *result.TimeZoneID != "GMT Standard Time" {
				t.Fatal("update reply was not decoded")
			}
			if result, _, err = s.RefreshByID(ctx, attendantID); err != nil || *result.ID != attendantID {
				t.Fatal("refresh reply was not decoded")
			}
			response, err = s.DeleteByID(ctx, attendantID)
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitAutoAttendantReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	info, language, zone := readFile(t, "testdata/information.json"), readFile(t, "testdata/language.json"), readFile(t, "testdata/time_zone.json")
	run(t, contracttest.Case{Name: "actual-reads", Source: "msteams/teams_admin_api/auto_attendants/testdata/information.json",
		Mappings: mapping("Get-CsAutoAttendantTenantInformation", "Get-CsAutoAttendantSupportedLanguage", "Get-CsAutoAttendantSupportedTimeZone"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", information, "Get-CsAutoAttendantTenantInformation", nil, 200, info),
			wire("GET", information, "Get-CsAutoAttendantTenantInformation", nil, 200, info),
			wire("GET", information, "Get-CsAutoAttendantTenantInformation", nil, 200, info),
			wire("GET", languagesRoute+"en-US", "Get-CsAutoAttendantSupportedLanguage", nil, 200, language),
			wire("GET", timeZonesRoute+"UTC", "Get-CsAutoAttendantSupportedTimeZone", nil, 200, zone)},
		Want: contracttest.Expected{Status: 200, Body: zone},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.AutoAttendants
			value, _, err := s.GetTenantInformation(ctx)
			if err != nil || *value.DefaultLanguage != "en-US" || *value.DefaultTimeZone != "UTC" || len(value.SupportedLanguages) != 54 ||
				len(value.SupportedTimeZones) != 108 || value.RawFields["MainlineAttendantSupportedLanguages"] == nil {
				t.Fatal("tenant information was not decoded")
			}
			if languages, _, err := s.ListSupportedLanguages(ctx); err != nil || len(languages) != 54 {
				t.Fatal("supported languages were not decoded")
			}
			if zones, _, err := s.ListSupportedTimeZones(ctx); err != nil || len(zones) != 108 {
				t.Fatal("supported time zones were not decoded")
			}
			if english, _, err := s.GetSupportedLanguageByID(ctx, "en-US"); err != nil || len(english.Voices) != 4 || english.VoiceResponseSupported == nil {
				t.Fatal("supported language was not decoded")
			}
			result, response, err := s.GetSupportedTimeZoneByID(ctx, "UTC")
			return contracttest.Observation{Value: result, Response: response, Err: err}
		}})
}

func TestUnitAutoAttendantDraftsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	prompt, entity, option := readFile(t, "testdata/draft_prompt.json"), readFile(t, "testdata/draft_callable_entity.json"), readFile(t, "testdata/draft_menu_option.json")
	menu, flow, association := readFile(t, "testdata/draft_menu.json"), readFile(t, "testdata/draft_call_flow.json"), readFile(t, "testdata/draft_call_handling_association.json")
	var decoded struct {
		Prompt         aa.Prompt
		CallableEntity aa.CallableEntity
		MenuOption     aa.MenuOption
		Menu           aa.Menu
	}
	for _, raw := range [][]byte{prompt, entity, option, menu} {
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
	}
	optionValue := aa.MenuOption{Action: ptr(aa.ActionTransferCallToTarget), DtmfResponse: ptr(aa.DtmfTone1), CallTarget: &decoded.CallableEntity}
	menuValue := aa.Menu{Name: ptr("Default menu"), Prompts: []aa.Prompt{decoded.Prompt}, MenuOptions: []aa.MenuOption{decoded.MenuOption},
		DirectorySearchMethod: ptr(aa.SearchNone)}
	flowValue := aa.CallFlow{Name: ptr("Default call flow"), Greetings: []aa.Prompt{decoded.Prompt}, Menu: &decoded.Menu}
	run(t, contracttest.Case{Name: "actual-drafts", Source: "msteams/teams_admin_api/auto_attendants/testdata/draft_call_flow.json",
		Mappings: mapping("New-CsAutoAttendantPrompt", "New-CsAutoAttendantCallableEntity", "New-CsAutoAttendantMenuOption", "New-CsAutoAttendantMenu",
			"New-CsAutoAttendantCallFlow", "New-CsAutoAttendantCallHandlingAssociation"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", attendants+"/prompts/draft", "New-CsAutoAttendantPrompt", marshal(t, greeting("Welcome to the SDK probe")), 200, prompt),
			wire("POST", attendants+"/callable-entities/draft", "New-CsAutoAttendantCallableEntity", json.RawMessage(`{"Id":"`+pstn+`","Type":"ExternalPstn"}`), 200, entity),
			wire("POST", attendants+"/menus/menu-options/draft", "New-CsAutoAttendantMenuOption", marshal(t, optionValue), 200, option),
			wire("POST", attendants+"/menus/draft", "New-CsAutoAttendantMenu", marshal(t, menuValue), 200, menu),
			wire("POST", attendants+"/call-flows/draft", "New-CsAutoAttendantCallFlow", marshal(t, flowValue), 200, flow),
			wire("POST", attendants+"/call-handling-associations/draft", "New-CsAutoAttendantCallHandlingAssociation",
				json.RawMessage(`{"Type":"AfterHours","ScheduleId":"`+probe+`","CallFlowId":"`+defaultFlowID+`","Enabled":true}`), 200, association)},
		Want: contracttest.Expected{Status: 200, Body: association},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.AutoAttendants
			if value, _, err := s.DraftPrompt(ctx, greeting("Welcome to the SDK probe")); err != nil || *value.ActiveType != aa.PromptTextToSpeech {
				t.Fatal("prompt draft was not decoded")
			}
			if value, _, err := s.DraftCallableEntity(ctx, aa.CallableEntity{ID: ptr(pstn), Type: ptr(aa.EntityExternalPstn)}); err != nil || *value.ID != redactedTarget {
				t.Fatal("callable entity draft was not decoded")
			}
			if value, _, err := s.DraftMenuOption(ctx, optionValue); err != nil || *value.DtmfResponse != aa.DtmfTone1 {
				t.Fatal("menu option draft was not decoded")
			}
			if value, _, err := s.DraftMenu(ctx, menuValue); err != nil || len(value.MenuOptions) != 1 {
				t.Fatal("menu draft was not decoded")
			}
			if value, _, err := s.DraftCallFlow(ctx, flowValue); err != nil || *value.ID != defaultFlowID {
				t.Fatal("call flow draft did not carry the assigned ID")
			}
			value, response, err := s.DraftCallHandlingAssociation(ctx, aa.CallHandlingAssociation{Type: ptr(aa.AssociationAfterHours), ScheduleID: ptr(probe),
				CallFlowID: ptr(defaultFlowID)})
			if err == nil && (*value.Priority != 1 || !*value.Enabled) {
				t.Fatal("association draft was not decoded")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		method, path, command, fixture string
		body                           json.RawMessage
		status                         int
		invoke                         func(context.Context, *aa.Service) (*resty.Response, error)
	}{
		"get-missing": {"GET", attendants + "/" + probe, "Get-CsAutoAttendant", "get_missing_404.json", nil, 404,
			func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
				value, response, err := s.GetByID(ctx, probe)
				if value != nil {
					t.Fatal("missing attendant returned a value")
				}
				return response, err
			}},
		"delete-missing": {"DELETE", attendants + "/" + attendantID, "Remove-CsAutoAttendant", "delete_missing_404.json", nil, 404,
			func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
				return s.DeleteByID(ctx, attendantID)
			}},
		"language-missing": {"GET", languagesRoute + "xx-XX", "Get-CsAutoAttendantSupportedLanguage", "language_missing_404.json", nil, 404,
			func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
				_, r, e := s.GetSupportedLanguageByID(ctx, "xx-XX")
				return r, e
			}},
		"dial-scope-missing": {"POST", attendants + "/group-dial-scopes/draft", "New-CsAutoAttendantDialScope", "draft_dial_scope_missing_404.json",
			json.RawMessage(`{"GroupIds":["` + probe + `"]}`), 404,
			func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
				_, r, e := s.DraftGroupDialScope(ctx, []string{probe})
				return r, e
			}},
		"prompt-invalid": {"POST", attendants + "/prompts/draft", "New-CsAutoAttendantPrompt", "draft_prompt_invalid_400.json",
			json.RawMessage(`{"ActiveType":"None"}`), 400,
			func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
				_, r, e := s.DraftPrompt(ctx, aa.Prompt{ActiveType: ptr(aa.PromptNone)})
				return r, e
			}},
		"holiday-disabled": {"POST", attendants + "/call-handling-associations/draft", "New-CsAutoAttendantCallHandlingAssociation",
			"draft_call_handling_association_invalid_400.json",
			json.RawMessage(`{"Type":"Holiday","ScheduleId":"` + probe + `","CallFlowId":"` + defaultFlowID + `","Enabled":false}`), 400,
			func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
				_, r, e := s.DraftCallHandlingAssociation(ctx, aa.CallHandlingAssociation{Type: ptr(aa.AssociationHoliday), ScheduleID: ptr(probe),
					CallFlowID: ptr(defaultFlowID), Enabled: ptr(false)})
				return r, e
			}},
	} {
		reply := readFile(t, "testdata/"+c.fixture)
		run(t, contracttest.Case{Name: "actual-" + name, Source: "msteams/teams_admin_api/auto_attendants/testdata/" + c.fixture, Mappings: mapping(c.command),
			Steps: []contracttest.Exchange{discovery(t), wire(c.method, c.path, c.command, c.body, c.status, reply)},
			Want:  contracttest.Expected{Status: c.status, Body: reply, APIError: apiError(c.status, reply)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := c.invoke(ctx, sdk.TeamsAdminAPI.AutoAttendants)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitAutoAttendantOptionalQueriesAndNormalization(t *testing.T) {
	t.Parallel()
	empty := []byte(`{"AutoAttendants":[],"Diagnostics":[]}`)
	attendant := []byte(`{"AutoAttendant":{"Id":"` + attendantID + `","Name":"a"}}`)
	records := []byte(`[{"Id":"r","Status":"Pending","Type":"AudioFile","ErrorCode":null,"Future":1}]`)
	flow := &aa.CallFlow{Name: ptr("f"), Menu: &aa.Menu{Name: ptr("m")}}
	// Enabling the mainline attendant turns voice response on, as New-CsAutoAttendant does.
	createBody := json.RawMessage(`{"Name":"a","LanguageId":"en-US","TimeZoneId":"UTC","DefaultCallFlow":{"Name":"f","Menu":{"Name":"m"}},` +
		`"VoiceResponseEnabled":true,"MainlineAttendantEnabled":true,"AutoRecordingTemplateId":"tpl","AuthorizedUsers":[]}`)
	run(t, contracttest.Case{Name: "queries-and-normalization", Mappings: mapping("Get-CsAutoAttendant", "Get-CsAutoAttendantStatus", "New-CsAutoAttendant", "Set-CsAutoAttendant"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", attendants+"?Descending=True&ExcludeContent=True&First=1&IncludeStatus=False&NameFilter=a+b&Skip=2&SortBy=Name", "Get-CsAutoAttendant", nil, 200, empty),
			wire("GET", attendants, "Get-CsAutoAttendant", nil, 200, empty),
			wire("GET", attendants+"/status/"+attendantID+"?includeResources=AudioFile%2CDialByNameVoiceResponses", "Get-CsAutoAttendantStatus", nil, 200, records),
			wire("POST", attendants, "New-CsAutoAttendant", createBody, 200, attendant),
			wire("PUT", attendants+"/"+attendantID, "Set-CsAutoAttendant", json.RawMessage(`{"Name":"a"}`), 204, nil)},
		Want: contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.AutoAttendants
			values, _, err := s.List(ctx, &aa.ListOptions{First: ptr(int32(1)), Skip: ptr(int32(2)), ExcludeContent: ptr(true), NameFilter: ptr("a b"),
				SortBy: ptr("Name"), Descending: ptr(true), IncludeStatus: ptr(false)})
			if err != nil || values == nil || len(values) != 0 {
				t.Fatal("empty list was not a non-nil empty slice")
			}
			if _, _, err := s.List(ctx, nil); err != nil {
				t.Fatal("default list failed")
			}
			statuses, _, err := s.GetStatusByID(ctx, attendantID, &aa.StatusOptions{IncludeResources: []string{"AudioFile", "DialByNameVoiceResponses"}})
			if err != nil || len(statuses) != 1 || string(statuses[0].Status) != `"Pending"` || statuses[0].RawFields["Future"] == nil {
				t.Fatal("status records were not decoded")
			}
			value, response, err := s.Create(ctx, aa.Settings{Name: ptr("a"), LanguageID: ptr("en-US"), TimeZoneID: ptr("UTC"), DefaultCallFlow: flow,
				VoiceResponseEnabled: ptr(false), MainlineAttendantEnabled: ptr(true), AutoRecordingTemplateID: ptr("tpl"), AuthorizedUsers: &[]string{}})
			if err != nil || response.StatusCode() != 200 || *value.Name != "a" {
				t.Fatal("create with 200 reply failed")
			}
			// An empty update reply is a nil value.
			result, response, err := s.UpdateByID(ctx, attendantID, aa.Settings{Name: ptr("a")})
			if result != nil {
				t.Fatal("empty update reply produced a value")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitAutoAttendantMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	type invocation func(context.Context, *aa.Service) (*resty.Response, error)
	list := func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
		_, r, e := s.List(ctx, nil)
		return r, e
	}
	get := func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
		_, r, e := s.GetByID(ctx, "x")
		return r, e
	}
	status := func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
		_, r, e := s.GetStatusByID(ctx, "x", nil)
		return r, e
	}
	languages := func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
		_, r, e := s.ListSupportedLanguages(ctx)
		return r, e
	}
	for name, c := range map[string]struct {
		method, path, command, body string
		status                      int
		invoke                      invocation
	}{
		"list-missing":      {"GET", attendants, "Get-CsAutoAttendant", `{"Diagnostics":[]}`, 200, list},
		"list-null-item":    {"GET", attendants, "Get-CsAutoAttendant", `{"AutoAttendants":[null]}`, 200, list},
		"list-bad-item":     {"GET", attendants, "Get-CsAutoAttendant", `{"AutoAttendants":[7]}`, 200, list},
		"get-null":          {"GET", attendants + "/x", "Get-CsAutoAttendant", `{"AutoAttendant":null}`, 200, get},
		"get-bad-type":      {"GET", attendants + "/x", "Get-CsAutoAttendant", `{"AutoAttendant":{"Name":7}}`, 200, get},
		"status-object":     {"GET", attendants + "/status/x", "Get-CsAutoAttendantStatus", `{}`, 200, status},
		"status-null":       {"GET", attendants + "/status/x", "Get-CsAutoAttendantStatus", `null`, 200, status},
		"status-null-item":  {"GET", attendants + "/status/x", "Get-CsAutoAttendantStatus", `[null]`, 200, status},
		"languages-missing": {"GET", information, "Get-CsAutoAttendantTenantInformation", `{"TenantInformation":{"DefaultLanguage":"en-US"}}`, 200, languages},
		"information-array": {"GET", information, "Get-CsAutoAttendantTenantInformation", `{"TenantInformation":[]}`, 200, languages},
		"refresh-missing": {"POST", attendants + "/x/refresh", "Update-CsAutoAttendant", `{"Diagnostics":[]}`, 200, func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
			_, r, e := s.RefreshByID(ctx, "x")
			return r, e
		}},
		"update-bad": {"PUT", attendants + "/x", "Set-CsAutoAttendant", `{"AutoAttendant":7}`, 200, func(ctx context.Context, s *aa.Service) (*resty.Response, error) {
			_, r, e := s.UpdateByID(ctx, "x", aa.Settings{Name: ptr("a")})
			return r, e
		}},
	} {
		var body json.RawMessage
		if c.method == "PUT" {
			body = json.RawMessage(`{"Name":"a"}`)
		}
		run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{discovery(t), wire(c.method, c.path, c.command, body, c.status, []byte(c.body))},
			Want: contracttest.Expected{Status: c.status, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := c.invoke(ctx, sdk.TeamsAdminAPI.AutoAttendants)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	run(t, contracttest.Case{Name: "undeclared-status", Steps: []contracttest.Exchange{discovery(t), wire("DELETE", attendants+"/x", "Remove-CsAutoAttendant", nil, 202, nil)},
		Want: contracttest.Expected{Status: 202, Error: aa.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.AutoAttendants.DeleteByID(ctx, "x")
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitAutoAttendantInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	flow := &aa.CallFlow{Name: ptr("f"), Menu: &aa.Menu{Name: ptr("m")}}
	valid := func() aa.Settings {
		return aa.Settings{Name: ptr("a"), LanguageID: ptr("en-US"), TimeZoneID: ptr("UTC"), DefaultCallFlow: flow}
	}
	create := func(edit func(*aa.Settings)) func(context.Context, *aa.Service) error {
		settings := valid()
		edit(&settings)
		return func(ctx context.Context, s *aa.Service) error { _, _, e := s.Create(ctx, settings); return e }
	}
	for label, call := range map[string]func(context.Context, *aa.Service) error{
		"get-blank":     func(ctx context.Context, s *aa.Service) error { _, _, e := s.GetByID(ctx, " "); return e },
		"get-slash":     func(ctx context.Context, s *aa.Service) error { _, _, e := s.GetByID(ctx, "a/b"); return e },
		"delete-dot":    func(ctx context.Context, s *aa.Service) error { _, e := s.DeleteByID(ctx, ".."); return e },
		"refresh-blank": func(ctx context.Context, s *aa.Service) error { _, _, e := s.RefreshByID(ctx, ""); return e },
		"status-slash":  func(ctx context.Context, s *aa.Service) error { _, _, e := s.GetStatusByID(ctx, "a/b", nil); return e },
		"status-comma": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.GetStatusByID(ctx, "x", &aa.StatusOptions{IncludeResources: []string{"AudioFile,SipProvisioning"}})
			return e
		},
		"language-slash": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.GetSupportedLanguageByID(ctx, "en/US")
			return e
		},
		"zone-blank": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.GetSupportedTimeZoneByID(ctx, "")
			return e
		},
		"list-negative": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.List(ctx, &aa.ListOptions{First: ptr(int32(-1))})
			return e
		},
		"create-no-name":     create(func(s *aa.Settings) { s.Name = nil }),
		"create-control":     create(func(s *aa.Settings) { s.Name = ptr("a\nb") }),
		"create-no-language": create(func(s *aa.Settings) { s.LanguageID = ptr(" ") }),
		"create-no-zone":     create(func(s *aa.Settings) { s.TimeZoneID = nil }),
		"create-no-flow":     create(func(s *aa.Settings) { s.DefaultCallFlow = nil }),
		"recording-without-mainline": create(func(s *aa.Settings) {
			s.AutoRecordingTemplateID = ptr("tpl")
		}),
		"update-no-name": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.UpdateByID(ctx, "x", aa.Settings{})
			return e
		},
		"update-blank-id": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.UpdateByID(ctx, "", aa.Settings{Name: ptr("a")})
			return e
		},
		"prompt-no-type": func(ctx context.Context, s *aa.Service) error { _, _, e := s.DraftPrompt(ctx, aa.Prompt{}); return e },
		"option-no-dtmf": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.DraftMenuOption(ctx, aa.MenuOption{Action: ptr(aa.ActionDisconnectCall)})
			return e
		},
		"menu-no-name": func(ctx context.Context, s *aa.Service) error { _, _, e := s.DraftMenu(ctx, aa.Menu{}); return e },
		"flow-no-menu": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.DraftCallFlow(ctx, aa.CallFlow{Name: ptr("f")})
			return e
		},
		"entity-no-type": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.DraftCallableEntity(ctx, aa.CallableEntity{ID: ptr(pstn)})
			return e
		},
		"scope-empty": func(ctx context.Context, s *aa.Service) error { _, _, e := s.DraftGroupDialScope(ctx, nil); return e },
		"scope-blank-group": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.DraftGroupDialScope(ctx, []string{" "})
			return e
		},
		"association-no-flow": func(ctx context.Context, s *aa.Service) error {
			_, _, e := s.DraftCallHandlingAssociation(ctx, aa.CallHandlingAssociation{Type: ptr(aa.AssociationAfterHours), ScheduleID: ptr(probe)})
			return e
		},
	} {
		run(t, contracttest.Case{Name: label, Scope: "local", Want: contracttest.Expected{Error: aa.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: call(ctx, sdk.TeamsAdminAPI.AutoAttendants)}
			}})
	}
}

func TestUnitAutoAttendantSettingsCopiesWritableFields(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Id":"` + attendantID + `","Name":"a","TenantId":"t","LanguageId":"en-US","VoiceId":"Male","TimeZoneId":"UTC",` +
		`"DefaultCallFlow":{"Id":"f","Name":"f"},"CallFlows":[],"Schedules":[{"Id":"s","Name":"s","Future":1}],` +
		`"CallHandlingAssociations":[{"Type":"AfterHours","ScheduleId":"s","CallFlowId":"g","Priority":1,"Enabled":true}],` +
		`"VoiceResponseEnabled":false,"DialByNameResourceId":null,"Operator":{"Id":"o","Type":"User","DisplayName":"read-only"},` +
		`"ApplicationInstances":["ra"],"AuthorizedUsers":[],"HideAuthorizedUsers":null,"Status":null,"UserNameExtension":"None",` +
		`"AutoRecordingThread":"ignored","GreetingsSettingAuthorizedUsers":[]}`)
	var value aa.AutoAttendant
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(value.Settings())
	if err != nil {
		t.Fatal(err)
	}
	want := `{"Id":"` + attendantID + `","Name":"a","TenantId":"t","LanguageId":"en-US","VoiceId":"Male","TimeZoneId":"UTC",` +
		`"DefaultCallFlow":{"Id":"f","Name":"f"},"Schedules":[{"Id":"s","Name":"s","Future":1}],` +
		`"CallHandlingAssociations":[{"Type":"AfterHours","ScheduleId":"s","CallFlowId":"g","Priority":1,"Enabled":true}],` +
		`"VoiceResponseEnabled":false,"Operator":{"Id":"o","Type":"User","DisplayName":"read-only"},"ApplicationInstances":["ra"],"UserNameExtension":"None"}`
	if string(got) != want {
		t.Fatalf("settings copy differed:\n got %s\nwant %s", got, want)
	}
	var none *aa.AutoAttendant
	if settings := none.Settings(); settings.Name != nil {
		t.Fatal("nil attendant produced settings")
	}
	for _, target := range []any{&value, &aa.StatusRecord{}, &aa.TenantInformation{}} {
		if json.Unmarshal([]byte(`[]`), target) == nil {
			t.Fatal("non-object value was accepted")
		}
	}
}

func TestUnitAutoAttendantNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *aa.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.List(nil, nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.DeleteByID(t.Context(), "x"); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	if _, _, err := s.GetTenantInformation(t.Context()); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := aa.NewService(nil).GetByID(ctx, "x"); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
