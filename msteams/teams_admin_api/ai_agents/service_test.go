package aiagents_test

import (
	"context"
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
	aiagents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin  = "https://api.interfaces.records.teams.microsoft.com"
	targetForest  = "https://admin.example.invalid/"
	source        = "msteams/teams_admin_api/ai_agents/service_test.go"
	configuration = "00000000-0000-4000-8000-0000000000c1"
	missing       = "00000000-0000-4000-8000-000000000001"
	provider      = "00000000-0000-4000-8000-000000000003"
	agentID       = "00000000-0000-4000-8000-0000000000a1"
	tagTemplate   = "00000000-0000-4000-8000-000000000009"
	jsonType      = "application/json; charset=utf-8"
	agents        = "/Teams.VoiceApps/aiagents"
	catalog       = "/Teams.PlatformService/AiAgents"
	mcpsOrigin    = "https://teams.microsoft.com/api/mcps/df"
	mcpsAudience  = "6bc3b958-689b-49f5-9006-36d165f30e00"
	thread        = "19:msteams-sdk-nonexistent-0000@thread.v2"
	threadQuery   = "threadId=19%3Amsteams-sdk-nonexistent-0000%40thread.v2"
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

// Captured bytes remain response-only evidence. Private request metadata and
// Set's bound target are source-authored, not claims about capture-time requests.
func wire(method, path, command string, body json.RawMessage, status int, reply []byte) contracttest.Exchange {
	switch command {
	case "Get-CsAgent":
		if strings.HasPrefix(path, agents+"/") {
			command = "Get-CsAgent_Get1"
		} else {
			command = "Get-CsAgent_Get"
		}
	case "New-CsAgent":
		command = "New-CsAgent_New"
	case "Set-CsAgent":
		command = "Set-CsAgent_SetExpanded"
		if body != nil {
			var fields map[string]json.RawMessage
			if json.Unmarshal(body, &fields) == nil {
				if _, ok := fields["AgentTargetTagTemplateId"]; !ok {
					fields["AgentTargetTagTemplateId"] = json.RawMessage(`""`)
					body, _ = json.Marshal(fields)
				}
			}
		}
	case "Remove-CsAgent":
		command = "Remove-CsAgent_Remove"
	case "Get-CsAiAgents":
		command = "Get-CsAiAgents_Get"
	}
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

func knowledgeWire(method, path string, status int, reply []byte) contracttest.Exchange {
	contentType := "application/json"
	if reply == nil {
		contentType = ""
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: mcpsOrigin + path, Audience: mcpsAudience,
		Headers: http.Header{"Accept": {"application/json"}, "X-MS-Target-Uri": nil, "X-MS-CmdletName": {map[string]string{"GET": "Get-AIGeneratedKnowledgeContainer", "DELETE": "Remove-AIGeneratedKnowledge"}[method]}}},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "ai_agents"
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
		t.Fatalf("AI agent contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(set string, commands ...string) []contracttest.Mapping {
	out := make([]contracttest.Mapping, 0, len(commands))
	for _, command := range commands {
		out = append(out, contracttest.Mapping{Command: command, ParameterSet: set})
	}
	return out
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

func agentBody(name string, tag *string) json.RawMessage {
	body := map[string]string{"Name": name, "AgentId": agentID, "AgentType": aiagents.AgentTypeMicrosoftCopilotStudio}
	if tag != nil {
		body["AgentTargetTagTemplateId"] = *tag
	}
	raw, _ := json.Marshal(body)
	return raw
}

func checkAgent(t *testing.T, value *aiagents.Agent, name string) {
	t.Helper()
	if value == nil || *value.ConfigurationID != configuration || *value.Name != name || *value.AgentID != agentID ||
		*value.AgentType != aiagents.AgentTypeMicrosoftCopilotStudio || value.AgentTargetTagTemplateID != nil ||
		string(value.RawFields["AgentTargetTagTemplateId"]) != "null" || len(value.RawFields) != 5 {
		t.Fatal("acceptance agent was changed")
	}
}

func TestUnitAIAgentLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	created, get, list := readFile(t, "testdata/create_201.json"), readFile(t, "testdata/get.json"), readFile(t, "testdata/list.json")
	updated, deleted := readFile(t, "testdata/update.json"), readFile(t, "testdata/delete.json")
	// Fixture redaction replaces agent names with this placeholder.
	name := "Fixture name"
	run(t, contracttest.Case{Name: "actual-lifecycle", Source: "msteams/teams_admin_api/ai_agents/testdata/create_201.json",
		Mappings: mapping("__AllParameterSets", "New-CsAgent", "Get-CsAgent", "Set-CsAgent", "Remove-CsAgent"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", agents, "New-CsAgent", agentBody(name, nil), 201, created),
			wire("GET", agents+"/"+configuration, "Get-CsAgent", nil, 200, get),
			wire("GET", agents, "Get-CsAgent", nil, 200, list),
			wire("PUT", agents+"/"+configuration, "Set-CsAgent", agentBody(name, nil), 200, updated),
			wire("DELETE", agents+"/"+configuration, "Remove-CsAgent", nil, 200, deleted)},
		Want: contracttest.Expected{Status: 200, Body: deleted},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.AIAgents
			request := aiagents.AgentRequest{Name: name, AgentID: agentID, AgentType: aiagents.AgentTypeMicrosoftCopilotStudio}
			value, _, err := s.Create(ctx, request)
			if err != nil {
				t.Fatal(err)
			}
			checkAgent(t, value, name)
			if value, _, err = s.GetByID(ctx, configuration); err != nil {
				t.Fatal(err)
			}
			checkAgent(t, value, name)
			listed, _, err := s.List(ctx, nil)
			if err != nil || len(listed.Agents) != 1 || *listed.AgentCount != 1 || string(listed.RawFields["Diagnostics"]) != "[]" {
				t.Fatal("acceptance list was changed")
			}
			checkAgent(t, listed.Agents[0], name)
			if value, _, err = s.UpdateByID(ctx, configuration, request); err != nil {
				t.Fatal(err)
			}
			checkAgent(t, value, name)
			response, err := s.DeleteByID(ctx, configuration)
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitAIAgentReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	empty, notGUID, missingBody := readFile(t, "testdata/query_empty.json"), readFile(t, "testdata/query_provider_not_guid_400.json"), readFile(t, "testdata/get_missing_404.json")
	run(t, contracttest.Case{Name: "catalog-query", Source: "msteams/teams_admin_api/ai_agents/testdata/query_empty.json",
		Mappings: mapping("Get", "Get-CsAiAgents"),
		Steps:    []contracttest.Exchange{discovery(t), wire("GET", catalog+"?maxResults=5&providerId="+provider+"&showCount=True", "Get-CsAiAgents", nil, 200, empty)},
		Want:     contracttest.Expected{Status: 200, Body: empty},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.QueryPlatformAgents(ctx, aiagents.QueryRequest{ProviderID: provider, MaxResults: ptr(int32(5)), ShowCount: ptr(true)})
			if err == nil && (value.Agents == nil || len(value.Agents) != 0 || *value.AgentCount != 0 || value.ContinuationToken != nil) {
				t.Fatal("empty catalog page was changed")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "catalog-provider-not-guid", Source: "msteams/teams_admin_api/ai_agents/testdata/query_provider_not_guid_400.json",
		Mappings: mapping("Get", "Get-CsAiAgents"),
		Steps:    []contracttest.Exchange{discovery(t), wire("GET", catalog+"?providerId=msteams-sdk-probe", "Get-CsAiAgents", nil, 400, notGUID)},
		Want:     contracttest.Expected{Status: 400, Body: notGUID, APIError: apiError(400, notGUID)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.QueryPlatformAgents(ctx, aiagents.QueryRequest{ProviderID: "msteams-sdk-probe"})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "get-missing", Source: "msteams/teams_admin_api/ai_agents/testdata/get_missing_404.json",
		Mappings: mapping("__AllParameterSets", "Get-CsAgent"),
		Steps:    []contracttest.Exchange{discovery(t), wire("GET", agents+"/"+missing, "Get-CsAgent", nil, 404, missingBody)},
		Want:     contracttest.Expected{Status: 404, Body: missingBody, APIError: apiError(404, missingBody)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.GetByID(ctx, missing)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
}

func TestUnitAIAgentListOptionsAndCatalogFilters(t *testing.T) {
	t.Parallel()
	list := readFile(t, "testdata/list.json")
	run(t, contracttest.Case{Name: "list-options", Mappings: mapping("__AllParameterSets", "Get-CsAgent"),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", agents+"?Descending=False&ExcludeContent=True&First=1&NameFilter=msteams-sdk&Skip=2&SortBy=Name&TypeFilter=MicrosoftCopilotStudio", "Get-CsAgent", nil, 200, list)},
		Want:  contracttest.Expected{Status: 200, Body: list},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.List(ctx, &aiagents.ListOptions{First: ptr(int32(1)), Skip: ptr(int32(2)), ExcludeContent: ptr(true),
				NameFilter: ptr("msteams-sdk"), TypeFilter: ptr(aiagents.AgentTypeMicrosoftCopilotStudio), SortBy: ptr("Name"), Descending: ptr(false)})
			if err == nil && len(value.Agents) != 1 {
				t.Fatal("list page lost its agent")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	page := []byte(`{"agents":[{"msaAppId":"a","isTeamsIvrEnabled":true,"cid":9007199254740993,"channels":["msteamsphone"],"future":1,` +
		`"description":{"displayName":"Bot","agentType":"MicrosoftCopilotStudio","starRating":4.5,"capabilities":["voice"],"userTileSmallUrl":"u"}}],"continuationToken":"next","agentCount":7}`)
	run(t, contracttest.Case{Name: "catalog-filters", Mappings: mapping("Get", "Get-CsAiAgents"),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", catalog+"?agentId=a&agentIds=a%2Cb&channel=msteamsphone&continuationToken=tok&displayNameContains=Bo&displayNamePrefix=B&isTeamsIvrEnabled=True&providerId="+provider, "Get-CsAiAgents", nil, 200, page)},
		Want:  contracttest.Expected{Status: 200, Body: page},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.QueryPlatformAgents(ctx, aiagents.QueryRequest{ProviderID: provider, AgentID: ptr("a"), AgentIDs: ptr("a,b"),
				Channel: ptr("msteamsphone"), ContinuationToken: ptr("tok"), DisplayNameContains: ptr("Bo"), DisplayNamePrefix: ptr("B"), IsTeamsIvrEnabled: ptr(true)})
			if err == nil {
				agent := value.Agents[0]
				if *value.ContinuationToken != "next" || *value.AgentCount != 7 || *agent.MsaAppID != "a" || !*agent.IsTeamsIvrEnabled || *agent.CID != 9007199254740993 ||
					agent.Channels[0] == nil || *agent.Channels[0] != "msteamsphone" || string(agent.RawFields["future"]) != "1" || *agent.Description.DisplayName != "Bot" ||
					*agent.Description.StarRating != 4.5 || agent.Description.Capabilities[0] == nil || *agent.Description.Capabilities[0] != "voice" || string(agent.Description.RawFields["userTileSmallUrl"]) != `"u"` {
					t.Fatal("catalog page was changed")
				}
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitAIAgentRejectedWritesUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	request := aiagents.AgentRequest{Name: "msteams-sdk-probe", AgentID: agentID, AgentType: "msteams-sdk-unknown"}
	body, _ := json.Marshal(map[string]string{"Name": request.Name, "AgentId": agentID, "AgentType": request.AgentType})
	for _, tc := range []struct {
		name, method, path, command, fixture string
		body                                 json.RawMessage
		status                               int
		call                                 func(context.Context, *aiagents.Service) (*resty.Response, error)
	}{
		{"create-unknown-type", "POST", agents, "New-CsAgent", "create_unknown_type_400.json", body, 400, func(ctx context.Context, s *aiagents.Service) (*resty.Response, error) {
			_, r, err := s.Create(ctx, request)
			return r, err
		}},
		{"update-unknown-type", "PUT", agents + "/" + missing, "Set-CsAgent", "update_unknown_type_400.json", body, 400, func(ctx context.Context, s *aiagents.Service) (*resty.Response, error) {
			_, r, err := s.UpdateByID(ctx, missing, request)
			return r, err
		}},
		{"update-unknown-tag-template", "PUT", agents + "/" + configuration, "Set-CsAgent", "update_unknown_tag_template_400.json", agentBody("Fixture name", ptr(tagTemplate)), 400, func(ctx context.Context, s *aiagents.Service) (*resty.Response, error) {
			_, r, err := s.UpdateByID(ctx, configuration, aiagents.AgentRequest{Name: "Fixture name", AgentID: agentID, AgentType: aiagents.AgentTypeMicrosoftCopilotStudio, AgentTargetTagTemplateID: ptr(tagTemplate)})
			return r, err
		}},
		{"delete-missing", "DELETE", agents + "/" + missing, "Remove-CsAgent", "delete_missing_404.json", nil, 404, func(ctx context.Context, s *aiagents.Service) (*resty.Response, error) {
			return s.DeleteByID(ctx, missing)
		}},
	} {
		reply := readFile(t, "testdata/"+tc.fixture)
		run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/ai_agents/testdata/" + tc.fixture, Mappings: mapping("__AllParameterSets", tc.command),
			Steps: []contracttest.Exchange{discovery(t), wire(tc.method, tc.path, tc.command, tc.body, tc.status, reply)},
			Want:  contracttest.Expected{Status: tc.status, Body: reply, APIError: apiError(tc.status, reply)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := tc.call(ctx, sdk.TeamsAdminAPI.AIAgents)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitAIGeneratedKnowledgeUsesMCPSProfile(t *testing.T) {
	t.Parallel()
	rejected := readFile(t, "testdata/knowledge_containers_app_only_401.body")
	// The acceptance reply is an empty 401 body; contract sources must be JSON
	// or Go, so this case cites the test file.
	run(t, contracttest.Case{Name: "containers-app-only",
		Mappings: mapping("__AllParameterSets", "Get-AIGeneratedKnowledgeContainer"),
		Steps:    []contracttest.Exchange{knowledgeWire("GET", "/mkb/crux/container-urls?"+threadQuery, 401, rejected)},
		Want:     contracttest.Expected{Status: 401, Error: auth.ErrUnauthenticated},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.ListKnowledgeContainers(ctx, aiagents.KnowledgeRequest{ThreadID: thread})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		urls := []byte(`["https://a.example.invalid/c1","https://a.example.invalid/c2"]`)
		run(t, contracttest.Case{Name: "containers-" + string(mode), Mode: mode, Mappings: mapping("__AllParameterSets", "Get-AIGeneratedKnowledgeContainer"),
			Steps: []contracttest.Exchange{knowledgeWire("GET", "/mkb/crux/container-urls?callId=call%2F1&"+threadQuery, 200, urls)},
			Want:  contracttest.Expected{Status: 200, Body: urls},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.AIAgents.ListKnowledgeContainers(ctx, aiagents.KnowledgeRequest{ThreadID: thread, CallID: ptr("call/1")})
				if err == nil && (len(value) != 2 || value[1] == nil || *value[1] != "https://a.example.invalid/c2") {
					t.Fatal("container URLs were changed")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	run(t, contracttest.Case{Name: "remove", Mappings: mapping("__AllParameterSets", "Remove-AIGeneratedKnowledge"),
		Steps: []contracttest.Exchange{knowledgeWire("DELETE", "/mkb/crux?"+threadQuery, 204, nil)},
		Want:  contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.AIAgents.DeleteKnowledge(ctx, aiagents.KnowledgeRequest{ThreadID: thread})
			return contracttest.Observation{Response: response, Err: err}
		}})
	for _, body := range []string{`null`, `{}`, `[{}]`} {
		run(t, contracttest.Case{Name: "malformed-containers", Mappings: mapping("__AllParameterSets", "Get-AIGeneratedKnowledgeContainer"),
			Steps: []contracttest.Exchange{knowledgeWire("GET", "/mkb/crux/container-urls?"+threadQuery, 200, []byte(body))},
			Want:  contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.AIAgents.ListKnowledgeContainers(ctx, aiagents.KnowledgeRequest{ThreadID: thread})
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
	}
}

func TestUnitAIGeneratedKnowledgeFailsLocallyOffProfile(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"gcc", "gcch", "dod", "china"} {
		run(t, contracttest.Case{Name: "no-profile-" + cloud, Cloud: cloud, Scope: "local", Want: contracttest.Expected{Error: client.ErrDestination},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := sdk.TeamsAdminAPI.AIAgents.DeleteKnowledge(ctx, aiagents.KnowledgeRequest{ThreadID: thread})
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	run(t, contracttest.Case{Name: "wrong-backend", Scope: "local", Want: contracttest.Expected{Error: client.ErrDestination},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := aiagents.NewService(sdk.TeamsAdminAPI.Scope(), sdk.TeamsAdminAPI.Scope())
			value, response, err := s.ListKnowledgeContainers(ctx, aiagents.KnowledgeRequest{ThreadID: thread})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "no-knowledge-scope", Scope: "local", Want: contracttest.Expected{Error: client.ErrClosed},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := aiagents.NewService(sdk.TeamsAdminAPI.Scope(), nil).DeleteKnowledge(ctx, aiagents.KnowledgeRequest{ThreadID: thread})
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitAIAgentMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`null`, `1`, `{"AIAgent":{"Name":"a","Name":"b"}}`} {
		run(t, contracttest.Case{Name: "malformed-agent", Mappings: mapping("__AllParameterSets", "Get-CsAgent"), Steps: []contracttest.Exchange{discovery(t), wire("GET", agents+"/"+configuration, "Get-CsAgent", nil, 200, []byte(body))}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.GetByID(ctx, configuration)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
	}
	for _, body := range []string{`{"AIAgents":[null]}`, `{"AgentCount":1.5}`, `{"AgentCount":9223372036854775808}`} {
		run(t, contracttest.Case{Name: "malformed-list", Mappings: mapping("__AllParameterSets", "Get-CsAgent"), Steps: []contracttest.Exchange{discovery(t), wire("GET", agents, "Get-CsAgent", nil, 200, []byte(body))}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.List(ctx, nil)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
	}
	for _, body := range []string{`{"agentCount":2147483648}`, `{"agents":[],"agents":[]}`} {
		run(t, contracttest.Case{Name: "malformed-catalog", Mappings: mapping("Get", "Get-CsAiAgents"), Steps: []contracttest.Exchange{discovery(t), wire("GET", catalog+"?providerId="+provider, "Get-CsAiAgents", nil, 200, []byte(body))}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.AIAgents.QueryPlatformAgents(ctx, aiagents.QueryRequest{ProviderID: provider})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
	}
	for _, status := range []int{202, 204} {
		run(t, contracttest.Case{Name: "undeclared-status", Mappings: mapping("__AllParameterSets", "Remove-CsAgent"), Steps: []contracttest.Exchange{discovery(t), wire("DELETE", agents+"/"+configuration, "Remove-CsAgent", nil, status, nil)}, Want: contracttest.Expected{Status: status, Error: aiagents.ErrResponse}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.AIAgents.DeleteByID(ctx, configuration)
			return contracttest.Observation{Response: response, Err: err}
		}})
	}
}

func TestUnitAIAgentInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	valid := aiagents.AgentRequest{Name: "n", AgentID: "a", AgentType: aiagents.AgentTypeMicrosoftCopilotStudio}
	for name, call := range map[string]func(context.Context, *aiagents.Service) error{
		"get-blank":  func(ctx context.Context, s *aiagents.Service) error { _, _, e := s.GetByID(ctx, " "); return e },
		"get-slash":  func(ctx context.Context, s *aiagents.Service) error { _, _, e := s.GetByID(ctx, "a/b"); return e },
		"delete-dot": func(ctx context.Context, s *aiagents.Service) error { _, e := s.DeleteByID(ctx, ".."); return e },
		"update-blank": func(ctx context.Context, s *aiagents.Service) error {
			_, _, e := s.UpdateByID(ctx, "", valid)
			return e
		},
		"create-empty": func(ctx context.Context, s *aiagents.Service) error {
			_, _, e := s.Create(ctx, aiagents.AgentRequest{})
			return e
		},
		"create-no-type": func(ctx context.Context, s *aiagents.Service) error {
			_, _, e := s.Create(ctx, aiagents.AgentRequest{Name: "n", AgentID: "a"})
			return e
		},

		"query-no-provider": func(ctx context.Context, s *aiagents.Service) error {
			_, _, e := s.QueryPlatformAgents(ctx, aiagents.QueryRequest{})
			return e
		},

		"knowledge-no-thread": func(ctx context.Context, s *aiagents.Service) error {
			_, _, e := s.ListKnowledgeContainers(ctx, aiagents.KnowledgeRequest{})
			return e
		},
	} {
		run(t, contracttest.Case{Name: name, Scope: "local", Want: contracttest.Expected{Error: aiagents.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: call(ctx, sdk.TeamsAdminAPI.AIAgents)}
			}})
	}
}

func TestUnitAIAgentNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *aiagents.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.List(nil, nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.DeleteKnowledge(t.Context(), aiagents.KnowledgeRequest{ThreadID: thread}); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := aiagents.NewService(nil, nil).DeleteByID(ctx, configuration); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}

// Reply bodies below are authored vectors; request metadata follows the pinned
// source packet, independently of the historical live response fixtures.
func TestUnitAIAgentSourcePresenceAndQueryConstruction(t *testing.T) {
	for _, item := range []struct {
		name, method, path, command string
		body                        json.RawMessage
		status                      int
		invoke                      func(context.Context, *aiagents.Service) (*resty.Response, error)
	}{
		{"create-omits-target", "POST", agents, "New-CsAgent", json.RawMessage(`{"Name":" ","AgentId":"a","AgentType":"custom"}`), 201, func(ctx context.Context, s *aiagents.Service) (*resty.Response, error) {
			_, r, e := s.Create(ctx, aiagents.AgentRequest{Name: " ", AgentID: "a", AgentType: "custom"})
			return r, e
		}},
		{"replacement-binds-empty-strings", "PUT", agents + "/" + configuration, "Set-CsAgent", json.RawMessage(`{"Name":"","AgentId":"","AgentType":"","AgentTargetTagTemplateId":""}`), 200, func(ctx context.Context, s *aiagents.Service) (*resty.Response, error) {
			_, r, e := s.UpdateByID(ctx, configuration, aiagents.AgentRequest{})
			return r, e
		}},
		{"list-negative-and-escaped-query", "GET", agents + "?Descending=False&First=-1&NameFilter=+%01x", "Get-CsAgent", nil, 200, func(ctx context.Context, s *aiagents.Service) (*resty.Response, error) {
			_, r, e := s.List(ctx, &aiagents.ListOptions{First: ptr(int32(-1)), Descending: ptr(false), NameFilter: ptr(" \x01x"), TypeFilter: ptr("")})
			return r, e
		}},
		{"catalog-negative-and-escaped-query", "GET", catalog + "?displayNameContains=+%01x&maxResults=-1&providerId=" + provider + "&showCount=False", "Get-CsAiAgents", nil, 200, func(ctx context.Context, s *aiagents.Service) (*resty.Response, error) {
			_, r, e := s.QueryPlatformAgents(ctx, aiagents.QueryRequest{ProviderID: provider, DisplayNameContains: ptr(" \x01x"), Channel: ptr(""), MaxResults: ptr(int32(-1)), ShowCount: ptr(false)})
			return r, e
		}},
	} {
		t.Run(item.name, func(t *testing.T) {
			set := "__AllParameterSets"
			if item.command == "Get-CsAiAgents" {
				set = "Get"
			}
			reply := []byte(`{}`)
			run(t, contracttest.Case{Name: item.name, Mappings: mapping(set, item.command), Steps: []contracttest.Exchange{discovery(t), wire(item.method, item.path, item.command, item.body, item.status, reply)}, Want: contracttest.Expected{Status: item.status, Body: reply}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, e := item.invoke(ctx, sdk.TeamsAdminAPI.AIAgents)
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitAIAgentSourceSuccessStatusesAreExact(t *testing.T) {
	for _, item := range []struct {
		name, method, command string
		status                int
	}{
		{"create-rejects-200", "POST", "New-CsAgent", 200},
		{"create-rejects-204", "POST", "New-CsAgent", 204},
		{"remove-rejects-204", "DELETE", "Remove-CsAgent", 204},
	} {
		t.Run(item.name, func(t *testing.T) {
			path := agents
			var body json.RawMessage
			if item.method == "POST" {
				body = agentBody("name", nil)
			} else {
				path += "/" + configuration
			}
			run(t, contracttest.Case{Name: item.name, Mappings: mapping("__AllParameterSets", item.command), Steps: []contracttest.Exchange{discovery(t), wire(item.method, path, item.command, body, item.status, nil)}, Want: contracttest.Expected{Status: item.status, Error: aiagents.ErrResponse}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				var r *resty.Response
				var e error
				if item.method == "POST" {
					_, r, e = sdk.TeamsAdminAPI.AIAgents.Create(ctx, aiagents.AgentRequest{Name: "name", AgentID: agentID, AgentType: aiagents.AgentTypeMicrosoftCopilotStudio})
				} else {
					r, e = sdk.TeamsAdminAPI.AIAgents.DeleteByID(ctx, configuration)
				}
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}
