package callqueues_test

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
	cq "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_queues"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"

	probe       = "00000000-0000-4000-8000-000000000001"
	queueID     = "00000000-0000-4000-8000-000000000003"
	fixtureName = "msteams-sdk-fixture-queue"
	jsonType    = "application/json; charset=utf-8"
	queues      = "/Teams.VoiceApps/callqueues"
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
	contentType := jsonType
	if reply == nil {
		contentType = ""
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("call queue contract %s differed: %+v", c.Name, result.Differences())
	}
}

// apiError is the normalized error for a fixture: ConfigAPI reports the code
// at the top level or inside an "error" object.
func apiError(status int, body []byte) *client.APIError {
	var envelope struct {
		Code  string `json:"code"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &envelope)
	result := &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}
	for _, code := range []string{envelope.Code, envelope.Error.Code} {
		if code != "" {
			result.Codes = append(result.Codes, code)
		}
	}
	return result
}

func ptr[T any](v T) *T { return &v }

func TestUnitCallQueueLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	created, got, listed := readFile(t, "testdata/create_201.json"), readFile(t, "testdata/get.json"), readFile(t, "testdata/list.json")
	updated, deleted := readFile(t, "testdata/update.json"), readFile(t, "testdata/delete.json")
	createBody := json.RawMessage(`{"Name":"` + fixtureName + `","UseDefaultMusicOnHold":true,"AgentAlertTime":30,"PresenceAwareRouting":false,"RoutingMethod":3,` +
		`"LanguageId":"en-US","OverflowThreshold":50,"TimeoutThreshold":1200}`)
	// The update body is the read-back's settings with three changes.
	updateBody := json.RawMessage(`{"Name":"Fixture name","UseDefaultMusicOnHold":true,"AgentAlertTime":45,"AllowOptOut":true,"ConferenceMode":true,` +
		`"PresenceAwareRouting":false,"RoutingMethod":2,"LanguageId":"en-US","OverflowThreshold":50,"OverflowAction":0,"TimeoutThreshold":1200,` +
		`"TimeoutAction":0,"NoAgentAction":0,"NoAgentApplyTo":0}`)
	run(t, testutil.Case{Name: "actual-lifecycle",

		Steps: []testutil.Exchange{discovery(t),
			wire("POST", queues, "New-CsCallQueue", createBody, 201, created),
			wire("GET", queues+"/"+queueID+"?FilterInvalidObos=False", "Get-CsCallQueue", nil, 200, got),
			wire("GET", queues+"?FilterInvalidObos=False&First=10&NameFilter="+fixtureName, "Get-CsCallQueue", nil, 200, listed),
			wire("PUT", queues+"/"+queueID, "Set-CsCallQueue", updateBody, 200, updated),
			wire("DELETE", queues+"/"+queueID, "Remove-CsCallQueue", nil, 200, deleted)},
		Want: testutil.Expected{Status: 200, Body: deleted},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.CallQueues
			// LongestIdle forces PresenceAwareRouting off, as New-CsCallQueue does.
			value, response, err := s.Create(ctx, cq.Settings{Name: ptr(fixtureName), RoutingMethod: ptr(cq.RoutingLongestIdle), AgentAlertTime: ptr(int32(30)),
				UseDefaultMusicOnHold: ptr(true), OverflowThreshold: ptr(int32(50)), TimeoutThreshold: ptr(int32(1200)), LanguageID: ptr("en-US"),
				PresenceAwareRouting: ptr(true)}, nil)
			if err != nil || response.StatusCode() != 201 || *value.Identity != queueID || *value.RoutingMethod != cq.RoutingLongestIdle || *value.PresenceAwareRouting {
				t.Fatal("create receipt was not decoded")
			}
			value, _, err = s.GetByID(ctx, queueID)
			if err != nil || *value.TimeoutThreshold != 1200 || *value.LanguageID != "en-US" || len(value.RawFields) != 135 || value.RawFields["ETag"] == nil {
				t.Fatal("get lost fields")
			}
			values, _, err := s.List(ctx, &cq.ListOptions{NameFilter: ptr(fixtureName), First: ptr(int32(10))})
			if err != nil || len(values) != 1 || *values[0].Identity != queueID {
				t.Fatal("list was not decoded")
			}
			settings := value.Settings()
			settings.AgentAlertTime, settings.AllowOptOut, settings.RoutingMethod = ptr(int32(45)), ptr(true), ptr(cq.RoutingRoundRobin)
			// The service answers an update with diagnostics only.
			if result, response, err := s.UpdateByID(ctx, queueID, settings, nil); err != nil || result != nil || response.StatusCode() != 200 {
				t.Fatal("diagnostics-only update reply was not a nil value")
			}
			response, err = s.DeleteByID(ctx, queueID)
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitCallQueueMissingIdentityIsHTTP400(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		method, path, command, fixture string
		invoke                         func(context.Context, *cq.Service) (*resty.Response, error)
	}{
		"get-missing": {"GET", queues + "/" + probe + "?FilterInvalidObos=False", "Get-CsCallQueue", "get_missing_400.json",
			func(ctx context.Context, s *cq.Service) (*resty.Response, error) {
				value, response, err := s.GetByID(ctx, probe)
				if value != nil {
					t.Fatal("missing queue returned a value")
				}
				return response, err
			}},
		"delete-missing": {"DELETE", queues + "/" + queueID, "Remove-CsCallQueue", "delete_missing_400.json",
			func(ctx context.Context, s *cq.Service) (*resty.Response, error) { return s.DeleteByID(ctx, queueID) }},
	} {
		reply := readFile(t, "testdata/"+c.fixture)
		run(t, testutil.Case{Name: "actual-" + name,
			Steps: []testutil.Exchange{discovery(t), wire(c.method, c.path, c.command, nil, 400, reply)},
			Want:  testutil.Expected{Status: 400, Body: reply, APIError: apiError(400, reply)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				response, err := c.invoke(ctx, sdk.TeamsAdminAPI.CallQueues)
				return testutil.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitCallQueueOptionalQueriesAndNormalization(t *testing.T) {
	t.Parallel()
	empty := []byte(`{"CallQueues":[],"Diagnostics":[]}`)
	queue := []byte(`{"CallQueue":{"Identity":"` + queueID + `","Name":"q"}}`)
	createBody := json.RawMessage(`{"Name":"q","ThreadId":"19:thread","OverflowAction":1,"OverflowActionCallPriority":5,"TimeoutAction":0,"TimeoutActionCallPriority":0,` +
		`"ConferenceMode":true,"AutoRecordingTemplateId":"tpl","DistributionLists":[]}`)
	updateBody := json.RawMessage(`{"Name":"q","NoAgentAction":2,"NoAgentActionCallPriority":1}`)
	run(t, testutil.Case{Name: "queries-and-normalization",
		Steps: []testutil.Exchange{discovery(t),
			wire("GET", queues+"?Descending=True&ExcludeContent=True&FilterInvalidObos=False&First=1&NameFilter=a+b&Skip=2&SortBy=Name", "Get-CsCallQueue", nil, 200, empty),
			wire("GET", queues+"?FilterInvalidObos=False", "Get-CsCallQueue", nil, 200, empty),
			wire("POST", queues+"?channelUserObjectId="+probe, "New-CsCallQueue", createBody, 200, queue),
			wire("PUT", queues+"/"+queueID+"?channelUserObjectId="+probe, "Set-CsCallQueue", updateBody, 204, nil)},
		Want: testutil.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.CallQueues
			values, _, err := s.List(ctx, &cq.ListOptions{First: ptr(int32(1)), Skip: ptr(int32(2)), ExcludeContent: ptr(true), NameFilter: ptr("a b"),
				SortBy: ptr("Name"), Descending: ptr(true)})
			if err != nil || values == nil || len(values) != 0 {
				t.Fatal("empty list was not a non-nil empty slice")
			}
			if _, _, err := s.List(ctx, nil); err != nil {
				t.Fatal("default list failed")
			}
			owner := &cq.WriteOptions{ChannelUserObjectID: ptr(probe)}
			// Blank strings are omitted as the module removes them; an empty
			// list is sent.
			value, response, err := s.Create(ctx, cq.Settings{Name: ptr("q"), ThreadID: ptr("19:thread"), OverflowAction: ptr(cq.OverflowForward),
				OverflowActionCallPriority: ptr(int32(5)), TimeoutAction: ptr(cq.TimeoutDisconnect), TimeoutActionCallPriority: ptr(int32(0)),
				ConferenceMode: ptr(true), AutoRecordingTemplateID: ptr("tpl"), DistributionLists: &[]string{}, LanguageID: ptr("  ")}, owner)
			if err != nil || response.StatusCode() != 200 || *value.Name != "q" {
				t.Fatal("create with channel owner failed")
			}
			value, response, err = s.UpdateByID(ctx, queueID, cq.Settings{Name: ptr("q"), NoAgentAction: ptr(cq.NoAgentForward), NoAgentActionCallPriority: ptr(int32(1))}, owner)
			return testutil.Observation{Value: value, Response: response, Err: err}
		}})
}

func TestUnitCallQueueMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	type invocation func(context.Context, *cq.Service) (*resty.Response, error)
	list := func(ctx context.Context, s *cq.Service) (*resty.Response, error) {
		_, r, e := s.List(ctx, nil)
		return r, e
	}
	get := func(ctx context.Context, s *cq.Service) (*resty.Response, error) {
		_, r, e := s.GetByID(ctx, "x")
		return r, e
	}
	for name, c := range map[string]struct {
		method, path, command, body string
		status                      int
		invoke                      invocation
	}{
		"list-missing":   {"GET", queues + "?FilterInvalidObos=False", "Get-CsCallQueue", `{"Diagnostics":[]}`, 200, list},
		"list-null-item": {"GET", queues + "?FilterInvalidObos=False", "Get-CsCallQueue", `{"CallQueues":[null]}`, 200, list},
		"list-bad-item":  {"GET", queues + "?FilterInvalidObos=False", "Get-CsCallQueue", `{"CallQueues":[7]}`, 200, list},
		"get-null":       {"GET", queues + "/x?FilterInvalidObos=False", "Get-CsCallQueue", `{"CallQueue":null}`, 200, get},
		"get-bad-type":   {"GET", queues + "/x?FilterInvalidObos=False", "Get-CsCallQueue", `{"CallQueue":{"AgentAlertTime":"x"}}`, 200, get},
		"create-missing": {"POST", queues, "New-CsCallQueue", `{"Diagnostics":[]}`, 201, func(ctx context.Context, s *cq.Service) (*resty.Response, error) {
			_, r, e := s.Create(ctx, cq.Settings{Name: ptr("q")}, nil)
			return r, e
		}},
		"update-bad": {"PUT", queues + "/x", "Set-CsCallQueue", `{"CallQueue":7}`, 200, func(ctx context.Context, s *cq.Service) (*resty.Response, error) {
			_, r, e := s.UpdateByID(ctx, "x", cq.Settings{Name: ptr("q")}, nil)
			return r, e
		}},
	} {
		var body json.RawMessage
		if c.method == "POST" || c.method == "PUT" {
			body = json.RawMessage(`{"Name":"q"}`)
		}
		run(t, testutil.Case{Name: name, Steps: []testutil.Exchange{discovery(t), wire(c.method, c.path, c.command, body, c.status, []byte(c.body))},
			Want: testutil.Expected{Status: c.status, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				response, err := c.invoke(ctx, sdk.TeamsAdminAPI.CallQueues)
				return testutil.Observation{Response: response, Err: err}
			}})
	}
	run(t, testutil.Case{Name: "undeclared-status", Steps: []testutil.Exchange{discovery(t), wire("DELETE", queues+"/x", "Remove-CsCallQueue", nil, 202, nil)},
		Want: testutil.Expected{Status: 202, Error: cq.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.TeamsAdminAPI.CallQueues.DeleteByID(ctx, "x")
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitCallQueueInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	name := ptr("q")
	create := func(settings cq.Settings) func(context.Context, *cq.Service) error {
		return func(ctx context.Context, s *cq.Service) error { _, _, e := s.Create(ctx, settings, nil); return e }
	}
	for label, call := range map[string]func(context.Context, *cq.Service) error{
		"get-blank":  func(ctx context.Context, s *cq.Service) error { _, _, e := s.GetByID(ctx, " "); return e },
		"get-slash":  func(ctx context.Context, s *cq.Service) error { _, _, e := s.GetByID(ctx, "a/b"); return e },
		"delete-dot": func(ctx context.Context, s *cq.Service) error { _, e := s.DeleteByID(ctx, ".."); return e },
		"list-negative": func(ctx context.Context, s *cq.Service) error {
			_, _, e := s.List(ctx, &cq.ListOptions{Skip: ptr(int32(-1))})
			return e
		},
		"create-no-name":               create(cq.Settings{}),
		"create-blank-name":            create(cq.Settings{Name: ptr(" ")}),
		"create-control":               create(cq.Settings{Name: ptr("a\nb")}),
		"priority-not-forward":         create(cq.Settings{Name: name, OverflowAction: ptr(cq.OverflowDisconnectWithBusy), OverflowActionCallPriority: ptr(int32(2))}),
		"priority-unset-action":        create(cq.Settings{Name: name, TimeoutActionCallPriority: ptr(int32(1))}),
		"priority-range":               create(cq.Settings{Name: name, NoAgentAction: ptr(cq.NoAgentForward), NoAgentActionCallPriority: ptr(int32(6))}),
		"recording-without-conference": create(cq.Settings{Name: name, ConferenceMode: ptr(false), AutoRecordingTemplateID: ptr("t")}),
		"channel-owner-slash": func(ctx context.Context, s *cq.Service) error {
			_, _, e := s.Create(ctx, cq.Settings{Name: name}, &cq.WriteOptions{ChannelUserObjectID: ptr("a/b")})
			return e
		},
		"update-no-name": func(ctx context.Context, s *cq.Service) error {
			_, _, e := s.UpdateByID(ctx, "x", cq.Settings{}, nil)
			return e
		},
		"update-blank-id": func(ctx context.Context, s *cq.Service) error {
			_, _, e := s.UpdateByID(ctx, "", cq.Settings{Name: name}, nil)
			return e
		},
	} {
		run(t, testutil.Case{Name: label, Want: testutil.Expected{Error: cq.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return testutil.Observation{Err: call(ctx, sdk.TeamsAdminAPI.CallQueues)}
			}})
	}
}

func TestUnitCallQueueSettingsBackfillsFromRead(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Identity":"` + queueID + `","Name":"q","Description":"ignored","AgentAlertTime":20,"AllowOptOut":null,` +
		`"WelcomeMusicResourceId":"welcome","MusicOnHoldResourceId":" ","WelcomeTextToSpeechPrompt":"","LanguageId":"en-GB",` +
		`"OverflowAction":1,"OverflowActionTarget":{"Id":"target","Type":"User"},"TimeoutActionTarget":"sip:x","NoAgentActionTarget":{"Id":null},` +
		`"OboResourceAccounts":[{"ObjectId":"obo1"},{"Id":"obo2"}],"Users":["u1"],"DistributionLists":[],"AuthorizedUsers":[{"Type":"x"}],` +
		`"CallbackEmailNotificationTarget":{"Id":"mail"},"EnableResourceAccountsForOBO":true,"ShiftsTeamId":7,"Future":1}`)
	var value cq.CallQueue
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(value.Settings())
	if err != nil {
		t.Fatal(err)
	}
	want := `{"Name":"q","Users":["u1"],"WelcomeMusicAudioFileId":"welcome","AgentAlertTime":20,"LanguageId":"en-GB","OverflowAction":1,` +
		`"OverflowActionTarget":"target","TimeoutActionTarget":"sip:x","EnableResourceAccountsForOBO":true,"OboResourceAccountIds":["obo1","obo2"],` +
		`"CallbackEmailNotificationTarget":"mail"}`
	if string(got) != want {
		t.Fatalf("settings backfill differed:\n got %s\nwant %s", got, want)
	}
	var none *cq.CallQueue
	if settings := none.Settings(); settings.Name != nil {
		t.Fatal("nil queue produced settings")
	}
	if json.Unmarshal([]byte(`[]`), &value) == nil {
		t.Fatal("non-object queue was accepted")
	}
}

func TestUnitCallQueueNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *cq.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.List(nil, nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.DeleteByID(t.Context(), "x"); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := cq.NewService(nil).GetByID(ctx, "x"); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
