package contracttest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func graphCase() Case {
	request := client.Request{Method: http.MethodPatch,
		Path:   []string{"v1.0", "teams", "fixture /?#%雪"},
		Query:  url.Values{"$select": {"id", "displayName"}, "$filter": {"displayName eq 'A & B'"}},
		Header: http.Header{"If-Match": {`W/"fixture"`}},
		Body:   json.RawMessage(`{"description":null,"settings":{"enabled":false},"members":[]}`)}
	return Case{Name: "sparse-update", Family: "teams", Scope: "primitive",
		Source:   "internal/contracttest/runner_test.go",
		Mappings: []Mapping{{Command: "Set-Team", ParameterSet: "__AllParameterSets"}},
		Steps: []Exchange{{Request: WireRequest{Method: http.MethodPatch,
			URL:      "https://graph.microsoft.com/v1.0/teams/fixture%20%2F%3F%23%25%E9%9B%AA?%24select=id&%24select=displayName&%24filter=displayName+eq+%27A+%26+B%27",
			Audience: auth.GraphAudience, Headers: request.Header.Clone(), JSON: request.Body.(json.RawMessage)},
			Reply: Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}, "Request-Id": {"fixture-response"}},
				Body: []byte(`{"id":"fixture","unknown":{"enabled":false}}`)}}},
		Want: Expected{Status: 200, Headers: http.Header{"Request-Id": {"fixture-response"}},
			JSON: json.RawMessage(`{"id":"fixture","unknown":{"enabled":false}}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) Observation {
			value := json.RawMessage{}
			response, err := sdk.GraphAPI.Scope().Do(ctx, request, &value)
			return Observation{Value: value, Response: response, Err: err}
		}}
}

func requirePass(t *testing.T, c Case) Result {
	t.Helper()
	result, err := Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("offline contract differed: %+v", result.Differences())
	}
	return result
}

func TestUnitSuiteRegisterAndRun(t *testing.T) {
	t.Parallel()
	// These are runner registration checks, not nine implemented resources.
	families := []string{"teams", "team_templates", "team_members", "targeting", "shared_channels", "operations", "deployment_migration", "channels", "channel_members"}
	suite := &Suite{}
	for _, family := range families {
		c := graphCase()
		c.Family = family
		if err := suite.Register(c); err != nil {
			t.Fatal(err)
		}
	}
	if err := suite.Register(graphCase()); !errors.Is(err, ErrFixture) {
		t.Fatal("duplicate case accepted")
	}
	results, err := suite.Run(t.Context())
	if err != nil || len(results) != len(families) {
		t.Fatal("suite did not run all registered families")
	}
	for index, result := range results {
		if !result.Passed() {
			t.Fatalf("registered family failed: %+v", result.Differences())
		}
		if index > 0 && results[index-1].family >= result.family {
			t.Fatal("suite order depends on registration order")
		}
	}
	var zero Result
	if zero.Passed() {
		t.Fatal("unexecuted result reported success")
	}
}

func TestUnitReplyBodyCloseEventRunsOnceThroughRealSDK(t *testing.T) {
	t.Parallel()
	var closes atomic.Int32
	c := graphCase()
	c.Name = "body-close-event"
	c.Steps[0].Reply.OnBodyClose = func() { closes.Add(1) }
	invoke := c.Invoke
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) Observation {
		observed := invoke(ctx, sdk)
		if observed.Err != nil {
			return observed
		}
		if closes.Load() != 1 || observed.Response == nil || observed.Response.RawResponse == nil || observed.Response.RawResponse.Body == nil || !bytes.Equal(observed.Response.Body(), c.Steps[0].Reply.Body) {
			t.Fatal("real SDK did not drain/close the reply body exactly once")
		}
		for range 2 {
			if err := observed.Response.RawResponse.Body.Close(); err != nil {
				t.Fatal(err)
			}
		}
		if closes.Load() != 1 {
			t.Fatal("repeated body close replayed the callback")
		}
		return observed
	}
	requirePass(t, c)
	raw, err := json.Marshal(c.Steps[0].Reply)
	if err != nil || bytes.Contains(raw, []byte("OnBodyClose")) {
		t.Fatal("body-close callback leaked into serializable fixture metadata")
	}
}

func TestUnitRunDetectsWireAndResultRegressions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, path string
		change     func(*Case)
	}{
		{name: "method", path: "request[0].method", change: func(c *Case) { c.Steps[0].Request.Method = http.MethodPut }},
		{name: "route", path: "request[0].path", change: func(c *Case) {
			c.Steps[0].Request.URL = strings.Replace(c.Steps[0].Request.URL, "/v1.0/teams/", "/v1.0/wrong/", 1)
		}},
		{name: "origin", path: "request[0].origin", change: func(c *Case) {
			c.Steps[0].Request.URL = strings.Replace(c.Steps[0].Request.URL, "graph.microsoft.com", "wrong.invalid", 1)
		}},
		{name: "audience", path: "request[0].audience", change: func(c *Case) {
			// Keep Graph allowed by a later step so this checks the bearer sent
			// at this exact step, not just a provider-wide audience allowlist.
			c.Steps = append(c.Steps, c.Steps[0])
			c.Steps[0].Request.Audience = auth.TeamsAdminAudience
		}},
		{name: "query ordering", path: "request[0].query", change: func(c *Case) {
			c.Steps[0].Request.URL = strings.Replace(c.Steps[0].Request.URL, "%24select=id&%24select=displayName", "%24select=displayName&%24select=id", 1)
		}},
		{name: "header", path: "request[0].headers[0]", change: func(c *Case) { c.Steps[0].Request.Headers.Set("If-Match", "wrong") }},
		{name: "omission", path: "request[0].body.object[0]", change: func(c *Case) {
			c.Steps[0].Request.JSON = json.RawMessage(`{"members":[],"settings":{"enabled":false}}`)
		}},
		{name: "no body", path: "request[0].body", change: func(c *Case) { c.Steps[0].Request.JSON = nil }},
		{name: "unknown response", path: "result.value.object[1]", change: func(c *Case) {
			c.Want.JSON = json.RawMessage(`{"id":"fixture"}`)
		}},
		{name: "status", path: "result.status", change: func(c *Case) { c.Want.Status = 201 }},
		{name: "response header", path: "result.headers[0]", change: func(c *Case) { c.Want.Headers.Set("Request-Id", "different") }},
		{name: "missing attempt", path: "requests", change: func(c *Case) { c.Steps = append(c.Steps, c.Steps[0]) }},
		{name: "unexpected attempt", path: "requests", change: func(c *Case) {
			invoke := c.Invoke
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) Observation {
				_ = invoke(ctx, sdk)
				return invoke(ctx, sdk)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := graphCase()
			tc.change(&c)
			result, err := Run(t.Context(), c)
			if err != nil || result.Passed() {
				t.Fatal("deliberate regression was not reported")
			}
			for _, difference := range result.Differences() {
				if difference.Path == tc.path {
					return
				}
			}
			t.Fatalf("expected %s mismatch, got %+v", tc.path, result.Differences())
		})
	}
}

func TestUnitRunErrorAndCancellationContracts(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		edit func(*Case)
	}{
		{name: "forbidden", edit: func(c *Case) {
			c.Steps[0].Reply.Status = 403
			c.Steps[0].Reply.Body = []byte(`{"error":{"code":"FixtureDenied","innerError":{"request-id":"fixture-inner"}}}`)
			c.Want = Expected{Status: 403, Error: auth.ErrForbidden,
				APIError: &client.APIError{Backend: client.Graph, StatusCode: 403,
					Codes: []string{"FixtureDenied"}, RequestIDs: []string{"fixture-response", "fixture-inner"}}}
		}},
		{name: "malformed success", edit: func(c *Case) {
			c.Steps[0].Reply.Body = []byte("not-json")
			c.Want = Expected{Status: 200, Error: client.ErrDecode, Body: []byte("not-json")}
		}},
		{name: "canceled", edit: func(c *Case) {
			c.Steps = []Exchange{}
			invoke := c.Invoke
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) Observation {
				canceled, cancel := context.WithCancel(ctx)
				cancel()
				return invoke(canceled, sdk)
			}
			c.Want = Expected{Error: context.Canceled}
		}},
		{name: "local constructor", edit: func(c *Case) {
			c.Scope, c.Steps = "local", []Exchange{}
			c.Invoke = func(context.Context, *msteams.Client) Observation {
				return Observation{Value: struct{ ID string }{ID: "fixture"}}
			}
			c.Want = Expected{JSON: json.RawMessage(`{"ID":"fixture"}`)}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := graphCase()
			tc.edit(&c)
			requirePass(t, c)
		})
	}
	t.Run("wrong classification", func(t *testing.T) {
		c := graphCase()
		c.Want.Error = auth.ErrForbidden
		result, err := Run(t.Context(), c)
		if err != nil || result.Passed() || result.Differences()[0].Path != "result.error" {
			t.Fatal("error mismatch was hidden")
		}
	})
	t.Run("wrong diagnostics", func(t *testing.T) {
		c := graphCase()
		c.Want.APIError = &client.APIError{Backend: client.Graph, StatusCode: 403}
		result, err := Run(t.Context(), c)
		if err != nil || result.Passed() || result.Differences()[0].Path != "result.api_error" {
			t.Fatal("typed error mismatch was hidden")
		}
	})
}

func TestUnitRunCloudAndModeBinding(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, cloud string
		mode        auth.Mode
	}{
		{name: "default application"},
		{name: "commercial delegated", cloud: "commercial", mode: auth.Delegated},
		{name: "gcc high application", cloud: "gcch", mode: auth.Application},
		{name: "china delegated", cloud: "china", mode: auth.Delegated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := graphCase()
			c.Cloud, c.Mode = tc.cloud, tc.mode
			environment, err := auth.ResolveEnvironment(c.Cloud, nil)
			if err != nil {
				t.Fatal(err)
			}
			c.Steps[0].Request.URL = strings.Replace(c.Steps[0].Request.URL, auth.GraphAudience, environment.Endpoints.Graph, 1)
			c.Steps[0].Request.Audience = environment.Endpoints.Graph
			result := requirePass(t, c)
			mode := tc.mode
			if mode == "" {
				mode = auth.Application
			}
			if result.cloud != environment.Name || result.mode != string(mode) {
				t.Fatal("cloud/mode evidence does not identify the executed binding")
			}
		})
	}
}

func TestUnitRunInvalidFixtureNeverInvokes(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		edit func(*Case)
	}{
		{name: "missing name", edit: func(c *Case) { c.Name = "" }},
		{name: "foreign source", edit: func(c *Case) { c.Source = "https://example.invalid/secret.json" }},
		{name: "source traversal", edit: func(c *Case) { c.Source = "contracts/../../secret.json" }},
		{name: "absolute testdata source", edit: func(c *Case) { c.Source = "/msteams/graph_api/channel/testdata/channel.json" }},
		{name: "windows testdata source", edit: func(c *Case) { c.Source = `C:\msteams\graph_api\channel\testdata\channel.json` }},
		{name: "service source outside testdata", edit: func(c *Case) { c.Source = "msteams/graph_api/channel/channel.json" }},
		{name: "testdata source traversal", edit: func(c *Case) { c.Source = "msteams/graph_api/channel/testdata/../../secret.json" }},
		{name: "testdata source backslash", edit: func(c *Case) { c.Source = `msteams/graph_api/channel/testdata/..\secret.json` }},
		{name: "testdata source control", edit: func(c *Case) { c.Source = "msteams/graph_api/channel/testdata/secret\n.json" }},
		{name: "testdata source bad pointer", edit: func(c *Case) { c.Source = "msteams/graph_api/channel/testdata/channel.json#secret" }},
		{name: "test vector source traversal", edit: func(c *Case) { c.Source = "msteams/graph_api/../../secret_test.go" }},
		{name: "test vector source production file", edit: func(c *Case) { c.Source = "msteams/graph_api/channel/service.go" }},
		{name: "test vector source absolute", edit: func(c *Case) { c.Source = "/msteams/graph_api/channel/service_test.go" }},
		{name: "recovery source traversal", edit: func(c *Case) { c.Source = "tests/integration/lab/../../../secret_test.go" }},
		{name: "unknown scope", edit: func(c *Case) { c.Scope = "live" }},
		{name: "unknown mode", edit: func(c *Case) { c.Mode = "mixed" }},
		{name: "local http", edit: func(c *Case) { c.Scope = "local" }},
		{name: "invalid query", edit: func(c *Case) { c.Steps[0].Request.URL += "&bad=%zz" }},
		{name: "invalid json", edit: func(c *Case) { c.Steps[0].Request.JSON = json.RawMessage(`{`) }},
		{name: "ambiguous body", edit: func(c *Case) { c.Steps[0].Request.Binary = []byte{} }},
		{name: "invalid result json", edit: func(c *Case) { c.Want.JSON = json.RawMessage(`{`) }},
		{name: "invalid command", edit: func(c *Case) { c.Mappings[0].Command = "not a command" }},
		{name: "invalid response status", edit: func(c *Case) { c.Steps[0].Reply.Status = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := graphCase()
			c.Invoke = func(context.Context, *msteams.Client) Observation {
				t.Fatal("invalid fixture invoked the SDK")
				return Observation{}
			}
			tc.edit(&c)
			result, err := Run(t.Context(), c)
			if !errors.Is(err, ErrFixture) || result.Passed() {
				t.Fatal("invalid fixture accepted")
			}
		})
	}
}

func TestUnitRunPermanentServiceTestdataSources(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		source string
	}{
		{name: "graph service", source: "msteams/graph_api/channel/testdata/channel.json"},
		{name: "administration service pointer", source: "msteams/teams_admin_api/application_access_policy/testdata/policy.json#/body"},
		{name: "service Go test vectors", source: "msteams/graph_api/channel/service_test.go"},
		{name: "adversarial recovery vectors", source: "tests/integration/lab/graph/team_lifecycle_cleanup_test.go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := graphCase()
			c.Source = tc.source
			requirePass(t, c)
		})
	}
}

func TestUnitRunUnexpectedRequestCannotEscapeToNetwork(t *testing.T) {
	t.Parallel()
	c := graphCase()
	// The callback deliberately calls the supplied HTTP client directly. Even
	// code bypassing resource Scope still meets the in-memory rejecting transport.
	c.Steps = []Exchange{}
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) Observation {
		httpClient, err := sdk.GraphAPI.Scope().Session().HTTPClient()
		if err != nil {
			return Observation{Err: err}
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://must-never-resolve.invalid/private", nil)
		if err != nil {
			return Observation{Err: err}
		}
		response, err := httpClient.Do(request)
		if response != nil {
			_ = response.Body.Close()
		}
		return Observation{Err: err}
	}
	c.Want = Expected{Error: ErrWire}
	result, err := Run(t.Context(), c)
	if err != nil || result.Passed() || result.httpCalls != 1 {
		t.Fatal("unexpected destination was not rejected by offline transport")
	}
}

func TestUnitCompareRequestPreservesNoBodyVersusJSONNull(t *testing.T) {
	t.Parallel()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://graph.microsoft.com/v1.0/teams", io.NopCloser(strings.NewReader("null")))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+fixtureToken(auth.GraphAudience))
	want := WireRequest{Method: http.MethodGet, URL: request.URL.String(), Audience: auth.GraphAudience}
	if differences := compareRequest(want, request, "request[0]"); len(differences) != 1 || differences[0].Kind != "unexpected-body" {
		t.Fatalf("JSON null collapsed into absent body: %+v", differences)
	}
}

func binaryCase(payload []byte) Case {
	return Case{Name: "binary-upload", Family: "transfer", Scope: "primitive",
		Source: "internal/contracttest/runner_test.go",
		Steps: []Exchange{{Request: WireRequest{Method: http.MethodPut,
			URL:      "https://graph.microsoft.com/v1.0/groups/fixture/photo/$value",
			Audience: auth.GraphAudience, Headers: http.Header{"Content-Type": {"image/jpeg"}},
			Binary: bytes.Clone(payload)}, Reply: Reply{Status: http.StatusNoContent}}},
		Want: Expected{Status: http.StatusNoContent},
		Invoke: func(ctx context.Context, sdk *msteams.Client) Observation {
			response, err := sdk.GraphAPI.Scope().Upload(ctx, client.Request{
				Method: http.MethodPut, Path: []string{"v1.0", "groups", "fixture", "photo", "$value"}},
				io.NopCloser(bytes.NewReader(payload)), "image/jpeg", 1024, nil)
			return Observation{Response: response, Err: err}
		}}
}

func TestUnitRunBinaryComparesExactBytesWithoutNormalization(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		actual  []byte
		want    []byte
		matches bool
	}{
		{name: "opaque bytes", actual: []byte{0xff, 0x00, 0x80, 0x0a}, want: []byte{0xff, 0x00, 0x80, 0x0a}, matches: true},
		{name: "empty bytes", actual: []byte{}, want: []byte{}, matches: true},
		{name: "changed byte", actual: []byte{0xff, 0x00}, want: []byte{0xff, 0x01}},
		{name: "truncated bytes", actual: []byte{0xff}, want: []byte{0xff, 0x00}},
		{name: "unexpected payload", actual: []byte{0xff}, want: []byte{}},
		{name: "json whitespace remains significant", actual: []byte(`{"a":1}`), want: []byte(`{ "a":1 }`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := binaryCase(tc.actual)
			c.Steps[0].Request.Binary = tc.want
			result, err := Run(t.Context(), c)
			if err != nil || result.Passed() != tc.matches || result.httpCalls != 1 {
				t.Fatal("binary comparison changed bytes, skipped or replayed upload")
			}
			if !tc.matches {
				differences := result.Differences()
				if len(differences) != 1 || differences[0] != (Difference{Path: "request[0].body", Kind: "different-value"}) {
					t.Fatal("binary mismatch exposed payload or lost structural difference")
				}
			}
		})
	}
}
