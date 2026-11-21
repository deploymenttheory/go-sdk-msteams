package applicationaccesspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

type wireCase struct {
	Name     string          `json:"name"`
	Method   string          `json:"method"`
	URL      string          `json:"generated_url"`
	Body     json.RawMessage `json:"body"`
	Status   int             `json:"response_status"`
	Response string          `json:"response_body"`
}

// requestCases defines request inputs independently of tenant response fixtures.
func requestCases(t *testing.T) map[string]wireCase {
	t.Helper()
	return map[string]wireCase{
		"policy-list":                 {Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationAccessPolicy", Status: 200, Response: applicationResponseFixture(t, "list.json")},
		"policy-read-escaped":         {Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/tag%3AFixture%20%2F%20Policy", Status: 400},
		"policy-create":               {Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationAccessPolicy", Status: 400, Body: json.RawMessage(`{"AppIds":["fixture-app"],"Description":"","Identity":"tag:Fixture / Policy"}`)},
		"policy-update-put":           {Method: "PUT", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/tag%3AFixture%20%2F%20Policy", Status: 400, Body: json.RawMessage(`{"AppIds":["fixture-app"],"Description":"","Identity":"tag:Fixture / Policy"}`)},
		"policy-delete":               {Method: "DELETE", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/tag%3AFixture%20%2F%20Policy", Status: 400},
		"policy-forbidden":            {Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/Fixture", Status: 403},
		"user-assignment":             {Method: "PATCH", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/ApplicationAccessPolicy", Status: 400, Body: json.RawMessage(`{"policyName":"Fixture"}`)},
		"user-clear-empty-policy":     {Method: "PATCH", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/users/fixture%2Fuser/policies/ApplicationAccessPolicy", Status: 400, Body: json.RawMessage(`{"policyName":""}`)},
		"tenant-assignment":           {Method: "PATCH", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/tenants/policies/ApplicationAccessPolicy", Status: 400, Body: json.RawMessage(`{"forceSwitchPresent":false,"policyName":"Fixture"}`)},
		"group-assignment-accepted":   {Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/ApplicationAccessPolicy", Status: 400, Body: json.RawMessage(`{"policyName":"Fixture","priority":2}`)},
		"group-assignment-no-content": {Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/ApplicationAccessPolicy", Status: 400, Body: json.RawMessage(`{"policyName":"Fixture","priority":2}`)},
	}
}

func scriptedClient(t *testing.T, steps []wireCase) (*msteams.Client, *int) {
	t.Helper()
	var mu sync.Mutex
	calls := 0
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		if request.Header.Get("Authorization") != "Bearer fixture-admin-token" {
			t.Error("wrong token audience")
			return nil, errors.New("fixture audience mismatch")
		}
		if request.URL.Host != "api.interfaces.records.teams.microsoft.com" {
			t.Error("credential destination changed")
			return nil, errors.New("fixture host mismatch")
		}
		if calls == 1 {
			if request.Method != http.MethodGet || request.URL.Path != "/Teams.Tenant/serviceDiscovery" {
				t.Error("discovery must precede management")
			}
			if request.Header.Get("X-MS-Target-Uri") != "" {
				t.Error("discovery unexpectedly inherited target header")
			}
			return httpResponse(request, 200, discoveryResponseFixture(t)), nil
		}
		index := calls - 2
		if index >= len(steps) {
			t.Error("unexpected request/replay")
			return nil, errors.New("unexpected fixture request")
		}
		step := steps[index]
		expected, err := url.Parse(step.URL)
		if err != nil {
			return nil, err
		}
		if request.Method != step.Method || !sameSegments(request.URL, expected) || request.URL.RawQuery != expected.RawQuery {
			t.Errorf("route mismatch for %s", step.Name)
		}
		if request.Header.Get("X-MS-Target-Uri") != "https://admin.example.invalid/" {
			t.Error("forest root header missing/changed")
		}
		body := []byte{}
		if request.Body != nil {
			body, err = io.ReadAll(request.Body)
			if err != nil {
				return nil, err
			}
		}
		if string(step.Body) == "null" || len(step.Body) == 0 {
			if len(body) != 0 {
				t.Error("unexpected body")
			}
		} else {
			equalJSON(t, body, string(step.Body))
		}
		return httpResponse(request, step.Status, step.Response), nil
	})
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{
		auth.TeamsAdminAudience: {AccessToken: "fixture-admin-token", ExpiresAt: time.Now().Add(time.Hour)},
		auth.GraphAudience:      {AccessToken: "fixture-graph-token", ExpiresAt: time.Now().Add(time.Hour)},
	}}, msteams.WithTransport(transport), msteams.WithRequestPolicy(client.RequestPolicy{MaxAttempts: 1}))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sdk.Close(); err != nil {
			t.Error(err)
		}
	})
	return sdk, &calls
}

func httpResponse(request *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}
}

// Percent-encoding spellings may differ for colons; encoded slashes must remain
// inside one segment. Compare each segment after exactly one decode.
func sameSegments(actual, expected *url.URL) bool {
	left, right := strings.Split(actual.EscapedPath(), "/"), strings.Split(expected.EscapedPath(), "/")
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		a, err := url.PathUnescape(left[index])
		if err != nil {
			return false
		}
		b, err := url.PathUnescape(right[index])
		if err != nil || a != b {
			return false
		}
	}
	return true
}

func TestUnitCRUDAndAssignmentRequests(t *testing.T) {
	t.Parallel()
	cases := requestCases(t)
	for _, test := range []struct {
		name   string
		steps  []string
		invoke func(context.Context, *policy.Service) error
	}{
		{name: "get", steps: []string{"policy-read-escaped"}, invoke: func(ctx context.Context, s *policy.Service) error {
			_, _, err := s.GetByIdentity(ctx, "tag:Fixture / Policy")
			return err
		}},
		{name: "list", steps: []string{"policy-list"}, invoke: func(ctx context.Context, s *policy.Service) error {
			result, _, err := s.List(ctx)
			if err == nil && result == nil {
				return errors.New("empty array became nil")
			}
			return err
		}},
		{name: "create request", steps: []string{"policy-create"}, invoke: func(ctx context.Context, s *policy.Service) error {
			description := ""
			_, _, err := s.Create(ctx, policy.CreateRequest{Identity: "tag:Fixture / Policy", AppIDs: policy.ReplaceAppIDs([]string{"fixture-app"}), Description: &description})
			return err
		}},
		{name: "update sparse PUT", steps: []string{"policy-update-put"}, invoke: func(ctx context.Context, s *policy.Service) error {
			description := ""
			_, err := s.UpdateByIdentity(ctx, "tag:Fixture / Policy", policy.UpdateRequest{AppIDs: policy.ReplaceAppIDs([]string{"fixture-app"}), Description: &description})
			return err
		}},
		{name: "delete", steps: []string{"policy-delete"}, invoke: func(ctx context.Context, s *policy.Service) error {
			_, err := s.DeleteByIdentity(ctx, "tag:Fixture / Policy")
			return err
		}},
		{name: "assign user", steps: []string{"user-assignment"}, invoke: func(ctx context.Context, s *policy.Service) error {
			_, err := s.AssignToUser(ctx, "fixture/user", "Fixture")
			return err
		}},
		{name: "clear user", steps: []string{"user-clear-empty-policy"}, invoke: func(ctx context.Context, s *policy.Service) error {
			_, err := s.AssignToUser(ctx, "fixture/user", "")
			return err
		}},
		{name: "assign tenant", steps: []string{"tenant-assignment"}, invoke: func(ctx context.Context, s *policy.Service) error {
			_, err := s.AssignTenantDefault(ctx, "Fixture", false)
			return err
		}},
		{name: "group request failure", steps: []string{"group-assignment-accepted"}, invoke: func(ctx context.Context, s *policy.Service) error {
			rank := 2
			_, _, err := s.AssignToGroup(ctx, "fixture/group", "Fixture", &rank)
			return err
		}},
		{name: "group alternate request failure", steps: []string{"group-assignment-no-content"}, invoke: func(ctx context.Context, s *policy.Service) error {
			rank := 2
			_, _, err := s.AssignToGroup(ctx, "fixture/group", "Fixture", &rank)
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			steps := make([]wireCase, 0, len(test.steps))
			for _, name := range test.steps {
				steps = append(steps, cases[name])
			}
			sdk, calls := scriptedClient(t, steps)
			err := test.invoke(t.Context(), sdk.TeamsAdminAPI.ApplicationAccessPolicies)
			if steps[len(steps)-1].Status >= 400 {
				var api *client.APIError
				if !errors.As(err, &api) || api.StatusCode != 400 {
					t.Fatalf("request failure lost: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if *calls != len(steps)+1 {
				t.Fatal("request sequence incomplete")
			}
		})
	}
}

func TestUnitCreateReadbackFailureRetainsCreationWithoutCleanup(t *testing.T) {
	t.Parallel()
	initial := "SDK lab non-default initial description"
	sdk, calls := scriptedClient(t, []wireCase{
		{Method: "POST", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy", Status: 201,
			Body: json.RawMessage(`{"Identity":"fixture-policy","AppIds":["` + responseApplicationID + `"],"Description":"` + initial + `"}`)},
		{Method: "GET", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/fixture-policy", Status: 403},
	})
	result, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.Create(t.Context(), policy.CreateRequest{Identity: responsePolicyID, AppIDs: policy.ReplaceAppIDs([]string{responseApplicationID}), Description: &initial})
	var partial *policy.ReadBackError
	if !errors.As(err, &partial) || !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("partial failure lost: %v", err)
	}
	if result.Identity != responsePolicyID || response.StatusCode() != 403 || partial.CreationResponse.StatusCode() != 201 || *calls != 3 {
		t.Fatal("creation metadata lost or cleanup/replay occurred")
	}
}

func TestUnitListFilteredUsesAcceptanceResponse(t *testing.T) {
	t.Parallel()
	cases := requestCases(t)
	step := cases["policy-list"]

	sdk, calls := scriptedClient(t, []wireCase{step})
	items, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.ListFiltered(t.Context(), "fixture*")
	if err != nil || response.StatusCode() != 200 || len(items) != 1 || !responsePolicyIdentity(items[0].Identity) {
		t.Fatal("local filter lost identity/tag behavior")
	}
	if *calls != 2 {
		t.Fatal("filter made extra HTTP calls")
	}
	if items[0].RawFields["ConfigMetadata"] == nil {
		t.Fatal("filter lost unknown response properties")
	}
	if _, _, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.ListFiltered(t.Context(), "["); !errors.Is(err, policy.ErrInput) {
		t.Fatal("invalid pattern accepted")
	}
	if *calls != 2 {
		t.Fatal("invalid filter performed HTTP")
	}
}

func TestUnitValidationAndCancellationNeverDiscoverOrSend(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		invoke func(context.Context, *policy.Service) error
		want   error
	}{
		{name: "blank identity", invoke: func(ctx context.Context, s *policy.Service) error { _, _, err := s.GetByIdentity(ctx, " "); return err }, want: policy.ErrInput},
		{name: "invalid list update", invoke: func(ctx context.Context, s *policy.Service) error {
			_, err := s.UpdateByIdentity(ctx, "Fixture", policy.UpdateRequest{AppIDs: &policy.ListChange{}})
			return err
		}, want: policy.ErrInput},
		{name: "canceled", invoke: func(ctx context.Context, s *policy.Service) error {
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			_, _, err := s.List(canceled)
			return err
		}, want: context.Canceled},
	} {
		t.Run(test.name, func(t *testing.T) {
			sdk, calls := scriptedClient(t, nil)
			if err := test.invoke(t.Context(), sdk.TeamsAdminAPI.ApplicationAccessPolicies); !errors.Is(err, test.want) {
				t.Fatalf("wrong local failure: %v", err)
			}
			if *calls != 0 {
				t.Fatal("local validation performed I/O")
			}
		})
	}
}

func TestUnitMalformedOrEmptySuccessfulReadsAreFailures(t *testing.T) {
	t.Parallel()
	cases := requestCases(t)
	for _, body := range []string{"", `null`, `{}`, `{"value":[]}`, `not-json`} {
		t.Run("body-"+strings.ReplaceAll(body, "/", "_"), func(t *testing.T) {
			step := cases["policy-list"]
			step.Response = body
			sdk, _ := scriptedClient(t, []wireCase{step})
			_, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.List(t.Context())
			if !errors.Is(err, client.ErrDecode) || response == nil || response.StatusCode() != 200 {
				t.Fatal("malformed success became an empty list or lost HTTP metadata")
			}
		})
	}
}
