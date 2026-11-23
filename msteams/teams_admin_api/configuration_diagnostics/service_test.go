package configurationdiagnostics_test

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
	cd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/configuration_diagnostics"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"
	source       = "msteams/teams_admin_api/configuration_diagnostics/service_test.go"
	jsonType     = "application/json; charset=utf-8"
	feedback     = "/Skype.Policy/configurations/TeamsFeedbackPolicy"
	fixture      = "msteams-sdk-fixture-configuration"
	cacheClear   = "/Teams.InternalSupport/CacheClear"
	probeKey     = "msteams-sdk-probe-nonexistent-key"
	probeRegion  = "msteams-sdk-probe-invalid-region"
)

type invocation func(context.Context, *cd.Service) (any, *resty.Response, error)

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
	if len(reply) == 0 {
		contentType = ""
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "configuration_diagnostics"
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
		t.Fatalf("configuration diagnostics contract %s differed: %+v", c.Name, result.Differences())
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
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	_ = json.Unmarshal(body, &envelope)
	result := &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}
	if envelope.Code != "" {
		result.Codes = []string{envelope.Code}
	}
	if envelope.ErrorCode != 0 {
		result.ServiceErrorCodes = []int{envelope.ErrorCode}
	}
	return result
}

func ptr[T any](v T) *T { return &v }

func text(value *cd.Configuration, key string) string {
	var out string
	if value == nil || json.Unmarshal(value.RawFields[key], &out) != nil {
		return ""
	}
	return out
}

func TestUnitConfigurationReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	list, global := readFile(t, "testdata/list.json"), readFile(t, "testdata/get_global.json")
	run(t, contracttest.Case{Name: "actual-reads", Source: "msteams/teams_admin_api/configuration_diagnostics/testdata/list.json",
		Mappings: mapping("Get-CsConfiguration"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", feedback, "Get-CsConfiguration", nil, 200, list),
			wire("GET", feedback+"/configuration/Global", "Get-CsConfiguration", nil, 200, global),
			wire("GET", feedback+"/configuration/Global?apiVersion=2&schemaVersion=1", "Get-CsConfiguration", nil, 200, global)},
		Want: contracttest.Expected{Status: 200, Body: global},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.ConfigurationDiagnostics
			values, _, err := s.ListConfigurations(ctx, "TeamsFeedbackPolicy", nil)
			if err != nil || len(values) != 4 || *values[0].Identity != "Global" || len(values[0].RawFields) != 10 {
				t.Fatal("actual list was changed")
			}
			if _, _, err := s.GetConfiguration(ctx, "TeamsFeedbackPolicy", "Global", nil); err != nil {
				return contracttest.Observation{Err: err}
			}
			value, response, err := s.GetConfiguration(ctx, "TeamsFeedbackPolicy", "Global", &cd.ReadOptions{APIVersion: ptr("2"), SchemaVersion: ptr("1")})
			if err == nil && (*value.Identity != "Global" || *value.ConfigID != "Global" || text(value, "UserInitiatedMode") == "") {
				t.Fatal("actual Global instance was changed")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	for name, c := range map[string]struct {
		path, fixture string
		status        int
		invoke        invocation
	}{
		"unknown-type": {feedback[:len(feedback)-len("TeamsFeedbackPolicy")] + "MsTeamsSdkUnknownType", "testdata/list_unknown_type_400.json", 400,
			func(ctx context.Context, s *cd.Service) (any, *resty.Response, error) {
				return s.ListConfigurations(ctx, "MsTeamsSdkUnknownType", nil)
			}},
		"missing-instance": {feedback + "/configuration/msteams-sdk-missing", "testdata/get_missing_404.json", 404,
			func(ctx context.Context, s *cd.Service) (any, *resty.Response, error) {
				return s.GetConfiguration(ctx, "TeamsFeedbackPolicy", "msteams-sdk-missing", nil)
			}},
	} {
		body := readFile(t, c.fixture)
		run(t, contracttest.Case{Name: name, Source: "msteams/teams_admin_api/configuration_diagnostics/" + c.fixture, Mappings: mapping("Get-CsConfiguration"),
			Steps: []contracttest.Exchange{discovery(t), wire("GET", c.path, "Get-CsConfiguration", nil, c.status, body)},
			Want:  contracttest.Expected{Status: c.status, Body: body, APIError: apiError(c.status, body)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.ConfigurationDiagnostics)
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
	}
}

func TestUnitConfigurationLifecycleUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	created, replaced := readFile(t, "testdata/get_created.json"), readFile(t, "testdata/get_replaced.json")
	item := feedback + "/configuration/" + fixture
	createBody := json.RawMessage(`{"Identity":"` + fixture + `","UserInitiatedMode":"Disabled"}`)
	setBody := json.RawMessage(`{"Identity":"` + fixture + `","UserInitiatedMode":"Enabled","ReceiveSurveysMode":"Disabled"}`)
	run(t, contracttest.Case{Name: "actual-lifecycle", Source: "msteams/teams_admin_api/configuration_diagnostics/testdata/get_created.json",
		Mappings: mapping("New-CsConfiguration", "Get-CsConfiguration", "Set-CsConfiguration", "Remove-CsConfiguration"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", feedback, "New-CsConfiguration", createBody, 201, nil),
			wire("GET", item, "Get-CsConfiguration", nil, 200, created),
			wire("PUT", item+"?apiVersion=2", "Set-CsConfiguration", setBody, 204, nil),
			wire("GET", item, "Get-CsConfiguration", nil, 200, replaced),
			wire("DELETE", item, "Remove-CsConfiguration", nil, 204, nil)},
		Want: contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.ConfigurationDiagnostics
			if response, err := s.CreateConfiguration(ctx, "TeamsFeedbackPolicy", cd.ConfigurationRequest{Identity: fixture,
				Settings: map[string]any{"UserInitiatedMode": "Disabled"}}, nil); err != nil || response.StatusCode() != 201 {
				t.Fatal("create was not accepted")
			}
			value, _, err := s.GetConfiguration(ctx, "TeamsFeedbackPolicy", fixture, nil)
			if err != nil || text(value, "UserInitiatedMode") != "Disabled" {
				t.Fatal("created read-back changed")
			}
			if _, err := s.SetConfiguration(ctx, "TeamsFeedbackPolicy", fixture, cd.ConfigurationRequest{Identity: fixture,
				Settings: map[string]any{"UserInitiatedMode": "Enabled", "ReceiveSurveysMode": "Disabled"}}, &cd.WriteOptions{APIVersion: ptr("2")}); err != nil {
				t.Fatal("set was not accepted")
			}
			value, _, err = s.GetConfiguration(ctx, "TeamsFeedbackPolicy", fixture, nil)
			if err != nil || text(value, "UserInitiatedMode") != "Enabled" || text(value, "ReceiveSurveysMode") != "Disabled" {
				t.Fatal("replaced read-back changed")
			}
			response, err := s.RemoveConfiguration(ctx, "TeamsFeedbackPolicy", fixture)
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitConfigurationSupportCommandsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	denied := readFile(t, "testdata/cache_clear_403.json")
	run(t, contracttest.Case{Name: "cache-clear-denied", Source: "msteams/teams_admin_api/configuration_diagnostics/testdata/cache_clear_403.json",
		Mappings: mapping("Clear-CsCacheOperation"),
		Steps: []contracttest.Exchange{discovery(t), wire("POST", cacheClear, "Clear-CsCacheOperation",
			json.RawMessage(`{"KeysToDelete":["`+probeKey+`"],"Region":"`+probeRegion+`"}`), 403, denied)},
		Want: contracttest.Expected{Status: 403, Body: denied, APIError: apiError(403, denied)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.ClearCache(ctx, cd.ClearCacheRequest{KeysToDelete: []string{probeKey}, Region: ptr(probeRegion)})
			if value != nil || !errors.Is(err, auth.ErrForbidden) {
				t.Fatal("cache clear denial was not a forbidden error")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	schema := readFile(t, "testdata/schema_defaults_403.json")
	run(t, contracttest.Case{Name: "schema-defaults-denied", Source: "msteams/teams_admin_api/configuration_diagnostics/testdata/schema_defaults_403.json",
		Mappings: mapping("Set-CsPublishPolicySchemaDefaults"),
		Steps: []contracttest.Exchange{{Request: contracttest.WireRequest{Method: "POST", URL: configOrigin + "/Skype.Policy/schemadefaults/MsTeamsSdkUnknownType",
			Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-CmdletName": {"Set-CsPublishPolicySchemaDefaults"}}},
			Reply: contracttest.Reply{Status: 403, Headers: http.Header{"Content-Type": {"application/json"}}, Body: schema}}},
		Want: contracttest.Expected{Status: 403, Body: schema, APIError: apiError(403, schema)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.PublishPolicySchemaDefaults(ctx, "MsTeamsSdkUnknownType")
			if value != "" {
				t.Fatal("denied publish returned text")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitConfigurationSuccessShapesPreserveFields(t *testing.T) {
	t.Parallel()
	full := []byte(`[{"FailedKeys":[],"Region":"emea","Extra":1}]`)
	partial := []byte(`[{"FailedKeys":["k2"],"Region":"amer"},{"FailedKeys":null,"Region":null}]`)
	run(t, contracttest.Case{Name: "cache-clear-results", Mappings: mapping("Clear-CsCacheOperation"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("POST", cacheClear, "Clear-CsCacheOperation", json.RawMessage(`{"KeysToDelete":[]}`), 200, full),
			wire("POST", cacheClear, "Clear-CsCacheOperation", json.RawMessage(`{"KeysToDelete":["k1","k2"],"Region":"amer"}`), 206, partial)},
		Want: contracttest.Expected{Status: 206, Body: partial},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.ConfigurationDiagnostics
			values, _, err := s.ClearCache(ctx, cd.ClearCacheRequest{})
			if err != nil || len(values) != 1 || values[0].FailedKeys == nil || len(values[0].FailedKeys) != 0 || *values[0].Region != "emea" || len(values[0].RawFields) != 3 {
				t.Fatal("complete cache clear result changed")
			}
			values, response, err := s.ClearCache(ctx, cd.ClearCacheRequest{KeysToDelete: []string{"k1", "k2"}, Region: ptr("amer")})
			if err == nil && (len(values) != 2 || values[0].FailedKeys[0] != "k2" || values[1].FailedKeys != nil || values[1].Region != nil) {
				t.Fatal("partial cache clear result changed")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	published := []byte(`"published"`)
	run(t, contracttest.Case{Name: "schema-defaults-published", Mappings: mapping("Set-CsPublishPolicySchemaDefaults"),
		Steps: []contracttest.Exchange{{Request: contracttest.WireRequest{Method: "POST", URL: configOrigin + "/Skype.Policy/schemadefaults/TeamsMeetingPolicy",
			Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-CmdletName": {"Set-CsPublishPolicySchemaDefaults"}}},
			Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: published}}},
		Want: contracttest.Expected{Status: 200, Body: published},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.PublishPolicySchemaDefaults(ctx, "TeamsMeetingPolicy")
			if err == nil && value != `"published"` {
				t.Fatal("publish response text changed")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	empty := []byte(`[]`)
	run(t, contracttest.Case{Name: "empty-list", Mappings: mapping("Get-CsConfiguration"),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", feedback, "Get-CsConfiguration", nil, 200, empty)},
		Want:  contracttest.Expected{Status: 200, Body: empty},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, response, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.ListConfigurations(ctx, "TeamsFeedbackPolicy", nil)
			if err == nil && (values == nil || len(values) != 0) {
				t.Fatal("empty list was not a non-nil empty slice")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitConfigurationMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	list := func(ctx context.Context, s *cd.Service) (any, *resty.Response, error) {
		return s.ListConfigurations(ctx, "TeamsFeedbackPolicy", nil)
	}
	get := func(ctx context.Context, s *cd.Service) (any, *resty.Response, error) {
		return s.GetConfiguration(ctx, "TeamsFeedbackPolicy", "Global", nil)
	}
	clearCache := func(ctx context.Context, s *cd.Service) (any, *resty.Response, error) {
		return s.ClearCache(ctx, cd.ClearCacheRequest{})
	}
	for name, c := range map[string]struct {
		method, path, command, body string
		request                     json.RawMessage
		invoke                      invocation
	}{
		"list-object":    {"GET", feedback, "Get-CsConfiguration", `{}`, nil, list},
		"list-null-item": {"GET", feedback, "Get-CsConfiguration", `[null]`, nil, list},
		"list-null":      {"GET", feedback, "Get-CsConfiguration", `null`, nil, list},
		"get-array":      {"GET", feedback + "/configuration/Global", "Get-CsConfiguration", `[]`, nil, get},
		"get-null":       {"GET", feedback + "/configuration/Global", "Get-CsConfiguration", `null`, nil, get},
		"get-bad-type":   {"GET", feedback + "/configuration/Global", "Get-CsConfiguration", `{"Identity":7}`, nil, get},
		"cache-object":   {"POST", cacheClear, "Clear-CsCacheOperation", `{}`, json.RawMessage(`{"KeysToDelete":[]}`), clearCache},
		"cache-null":     {"POST", cacheClear, "Clear-CsCacheOperation", `[null]`, json.RawMessage(`{"KeysToDelete":[]}`), clearCache},
		"cache-bad-keys": {"POST", cacheClear, "Clear-CsCacheOperation", `[{"FailedKeys":"k"}]`, json.RawMessage(`{"KeysToDelete":[]}`), clearCache},
	} {
		run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{discovery(t), wire(c.method, c.path, c.command, c.request, 200, []byte(c.body))},
			Want: contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.ConfigurationDiagnostics)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	run(t, contracttest.Case{Name: "undeclared-create-status", Steps: []contracttest.Exchange{discovery(t),
		wire("POST", feedback, "New-CsConfiguration", json.RawMessage(`{"Identity":"x"}`), 200, nil)},
		Want: contracttest.Expected{Status: 200, Error: cd.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.CreateConfiguration(ctx, "TeamsFeedbackPolicy", cd.ConfigurationRequest{Identity: "x"}, nil)
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "undeclared-remove-status", Steps: []contracttest.Exchange{discovery(t),
		wire("DELETE", feedback+"/configuration/x", "Remove-CsConfiguration", nil, 200, nil)},
		Want: contracttest.Expected{Status: 200, Error: cd.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.RemoveConfiguration(ctx, "TeamsFeedbackPolicy", "x")
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitConfigurationInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	valid := cd.ConfigurationRequest{Identity: "x"}
	for name, call := range map[string]func(context.Context, *cd.Service) error{
		"list-blank-type": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.ListConfigurations(ctx, " ", nil)
			return e
		},
		"list-slash-type": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.ListConfigurations(ctx, "a/b", nil)
			return e
		},
		"list-empty-version": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.ListConfigurations(ctx, "T", &cd.ReadOptions{APIVersion: ptr("")})
			return e
		},
		"get-dot-name": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.GetConfiguration(ctx, "T", "..", nil)
			return e
		},
		"get-control-schema": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.GetConfiguration(ctx, "T", "Global", &cd.ReadOptions{SchemaVersion: ptr("a\nb")})
			return e
		},
		"create-blank-identity": func(ctx context.Context, s *cd.Service) error {
			_, e := s.CreateConfiguration(ctx, "T", cd.ConfigurationRequest{}, nil)
			return e
		},
		"create-duplicate-identity": func(ctx context.Context, s *cd.Service) error {
			_, e := s.CreateConfiguration(ctx, "T", cd.ConfigurationRequest{Identity: "x", Settings: map[string]any{"Identity": "y"}}, nil)
			return e
		},
		"create-backslash-type": func(ctx context.Context, s *cd.Service) error {
			_, e := s.CreateConfiguration(ctx, `a\b`, valid, nil)
			return e
		},
		"set-blank-name": func(ctx context.Context, s *cd.Service) error {
			_, e := s.SetConfiguration(ctx, "T", "", valid, nil)
			return e
		},
		"set-slash-name": func(ctx context.Context, s *cd.Service) error {
			_, e := s.SetConfiguration(ctx, "T", "a/b", valid, nil)
			return e
		},
		"remove-blank":    func(ctx context.Context, s *cd.Service) error { _, e := s.RemoveConfiguration(ctx, "T", ""); return e },
		"remove-dot-type": func(ctx context.Context, s *cd.Service) error { _, e := s.RemoveConfiguration(ctx, ".", "x"); return e },
		"cache-blank-key": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.ClearCache(ctx, cd.ClearCacheRequest{KeysToDelete: []string{""}})
			return e
		},
		"cache-empty-region": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.ClearCache(ctx, cd.ClearCacheRequest{Region: ptr("")})
			return e
		},
		"publish-blank": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.PublishPolicySchemaDefaults(ctx, "")
			return e
		},
		"publish-slash": func(ctx context.Context, s *cd.Service) error {
			_, _, e := s.PublishPolicySchemaDefaults(ctx, "a/b")
			return e
		},
	} {
		run(t, contracttest.Case{Name: name, Scope: "local", Want: contracttest.Expected{Error: cd.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: call(ctx, sdk.TeamsAdminAPI.ConfigurationDiagnostics)}
			}})
	}
}

func TestUnitConfigurationNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *cd.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.ListConfigurations(nil, "T", nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, _, err := s.PublishPolicySchemaDefaults(t.Context(), "T"); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := cd.NewService(nil, nil).RemoveConfiguration(ctx, "T", "x"); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}

func TestUnitConfigurationRequestBodies(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		value any
		want  string
	}{
		"configuration-identity-only": {cd.ConfigurationRequest{Identity: "Global"}, `{"Identity":"Global"}`},
		"configuration-settings":      {cd.ConfigurationRequest{Identity: "Tag:x", Settings: map[string]any{"A": nil, "B": false}}, `{"A":null,"B":false,"Identity":"Tag:x"}`},
		"cache-nil-keys":              {cd.ClearCacheRequest{}, `{"KeysToDelete":[]}`},
		"cache-region":                {cd.ClearCacheRequest{KeysToDelete: []string{"k"}, Region: ptr("")}, `{"KeysToDelete":["k"],"Region":""}`},
	} {
		raw, err := json.Marshal(c.value)
		if err != nil || string(raw) != c.want {
			t.Errorf("%s body = %s, want %s", name, raw, c.want)
		}
	}
}
