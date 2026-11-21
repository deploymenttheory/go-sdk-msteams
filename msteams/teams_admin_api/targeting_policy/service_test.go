package targetingpolicy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_policy"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsTargetingPolicy"

func discovery() contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid:444/ignored"}}`)}}
}
func wire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	headers := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://forest.fixture.invalid/"}}
	if body != nil {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "targeting_policy", "contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("targeting policy differs: setup=%v differences=%+v", err, r.Differences())
	}
	return r
}

func TestUnitPinnedModernResourceRoutes(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Cases []struct {
			Name     string          `json:"name"`
			Method   string          `json:"method"`
			URL      string          `json:"generated_url"`
			Body     json.RawMessage `json:"body"`
			Status   int             `json:"response_status"`
			Response string          `json:"response_body"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &fixtures) != nil {
		t.Fatal("invalid pinned request fixture")
	}
	seen := 0
	var results []contracttest.Result
	for _, f := range fixtures.Cases {
		if !strings.HasPrefix(f.Name, "targeting-policy-") {
			continue
		}
		seen++
		t.Run(f.Name, func(t *testing.T) {
			command := "Get-CsTeamsTargetingPolicy"
			if f.Method == "PUT" {
				command = "Set-CsTeamsTargetingPolicy"
			}
			if f.Method == "DELETE" {
				command = "Remove-CsTeamsTargetingPolicy"
			}
			u, err := url.Parse(f.URL)
			if err != nil {
				t.Fatal(err)
			}
			body := f.Body
			if string(body) == "null" {
				body = nil
			}
			want := contracttest.Expected{Status: f.Status}
			if f.Response != "" {
				want.JSON = json.RawMessage(f.Response)
			}
			results = append(results, run(t, contracttest.Case{Name: f.Name, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(), wire(f.Method, u.RequestURI(), command, body, f.Status, f.Response)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TargetingPolicies
				switch f.Name {
				case "targeting-policy-list":
					v, r, err := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: err}
				case "targeting-policy-read":
					v, r, err := s.GetByIdentity(ctx, "Global")
					return contracttest.Observation{Value: v, Response: r, Err: err}
				case "targeting-policy-update":
					enabled, empty := "Enabled", ""
					r, err := s.UpdateGlobal(ctx, policy.UpdateRequest{CustomTagsMode: &enabled, Description: &empty})
					return contracttest.Observation{Response: r, Err: err}
				default:
					r, err := s.DeleteByIdentity(ctx, "Global")
					return contracttest.Observation{Response: r, Err: err}
				}
			}, Want: want}))
		})
	}
	if seen != 4 {
		t.Fatal("pinned policy primitive count changed")
	}
	evidence(t, results)
}

func TestUnitSparseIdentityUpdateAndOpenValues(t *testing.T) {
	t.Parallel()
	identity := "tag:Fixture / Policy"
	mode, empty := "FutureServerMode", ""
	run(t, contracttest.Case{Name: "sparse-open-mode", Steps: []contracttest.Exchange{discovery(), wire("PUT", collection+"/configuration/"+url.PathEscape(identity), "Set-CsTeamsTargetingPolicy", json.RawMessage(`{"Identity":"tag:Fixture / Policy","ManageTagsPermissionMode":"FutureServerMode","SuggestedPresetTags":""}`), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, err := sdk.TeamsAdminAPI.TargetingPolicies.UpdateByIdentity(ctx, identity, policy.UpdateRequest{ManageTagsPermissionMode: &mode, SuggestedPresetTags: &empty})
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 204}})
	for _, id := range []string{"", " ", "..", "bad\nidentity"} {
		run(t, contracttest.Case{Name: "invalid-identity", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, err := sdk.TeamsAdminAPI.TargetingPolicies.UpdateByIdentity(ctx, id, policy.UpdateRequest{})
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Error: policy.ErrInput}})
	}
}

func TestUnitPolicyListSortsGlobalAndRetainsRawFields(t *testing.T) {
	t.Parallel()
	body := `[{"Identity":"tag:Later","CustomTagsMode":"Future","Description":null,"future":{"v":5}},{"Identity":"gLoBaL","SuggestedPresetTags":""},{"Identity":"tag:Last"}]`
	run(t, contracttest.Case{Name: "global-first-raw", Steps: []contracttest.Exchange{discovery(), wire("GET", collection, "Get-CsTeamsTargetingPolicy", nil, 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.TargetingPolicies.List(ctx)
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`[{"Identity":"gLoBaL","SuggestedPresetTags":""},{"Identity":"tag:Later","CustomTagsMode":"Future","Description":null,"future":{"v":5}},{"Identity":"tag:Last"}]`)}})
	run(t, contracttest.Case{Name: "denied", Steps: []contracttest.Exchange{discovery(), wire("GET", collection+"/configuration/Global", "Get-CsTeamsTargetingPolicy", nil, 403, `{"code":"Forbidden","message":"private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.TargetingPolicies.GetByIdentity(ctx, "Global")
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
}

func TestUnitCanceledPolicyDoesNotDiscoverOrAcquireCredentials(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "cancelled", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		ctx, stop := context.WithCancel(ctx)
		stop()
		r, err := sdk.TeamsAdminAPI.TargetingPolicies.UpdateGlobal(ctx, policy.UpdateRequest{})
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Error: context.Canceled}})
	var value policy.Policy
	if err := json.Unmarshal([]byte(`{"Identity":"Global","CustomTagsMode":17}`), &value); err == nil {
		t.Fatal("invalid typed setting accepted")
	}
}

func TestUnitTargetingFilterIsLocalAndTagAware(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "local-tag-filter", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsTargetingPolicy", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{discovery(), wire("GET", collection, "Get-CsTeamsTargetingPolicy", nil, 200, `[{"Identity":"Global"},{"Identity":"Tag:Fixture One"},{"Identity":"Fixture Two"},{"Identity":"Tag:Other"}]`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.TargetingPolicies.ListFiltered(ctx, "fIX*")
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`[{"Identity":"Tag:Fixture One"},{"Identity":"Fixture Two"}]`)}})
	evidence(t, []contracttest.Result{result})
	run(t, contracttest.Case{Name: "bad-filter-local", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.TargetingPolicies.ListFiltered(ctx, "[z-a]")
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Error: policy.ErrInput}})
}

func TestUnitTargetingFailuresDoNotInventSuccessOrReplayWrites(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{
		{"accepted-read", `{"Identity":"Global"}`, 202, policy.ErrResponse},
		{"null-read", `null`, 200, client.ErrDecode},
		{"missing-identity", `{"Description":"fixture"}`, 200, client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{discovery(), wire("GET", collection+"/configuration/Global", "Get-CsTeamsTargetingPolicy", nil, tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, err := sdk.TeamsAdminAPI.TargetingPolicies.GetByIdentity(ctx, "Global")
				return contracttest.Observation{Value: v, Response: r, Err: err}
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.err}})
		})
	}
	for _, status := range []int{429, 503} {
		run(t, contracttest.Case{Name: "delete-no-replay", Steps: []contracttest.Exchange{discovery(), wire("DELETE", collection+"/configuration/Global", "Remove-CsTeamsTargetingPolicy", nil, status, `{"code":"Unavailable","message":"private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, err := sdk.TeamsAdminAPI.TargetingPolicies.DeleteByIdentity(ctx, "Global")
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status, Codes: []string{"Unavailable"}}}})
	}
}
