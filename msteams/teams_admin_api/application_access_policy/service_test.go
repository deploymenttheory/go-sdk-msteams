package applicationaccesspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
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

func pinnedCases(t *testing.T) map[string]wireCase {
	t.Helper()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json")
	if err != nil {
		t.Fatal(err)
	}
	var value struct {
		Cases []wireCase `json:"cases"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	result := make(map[string]wireCase, len(value.Cases))
	for _, item := range value.Cases {
		result[item.Name] = item
	}
	return result
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
			return httpResponse(request, 200, `{"Endpoints":{"AdminServiceEndpoint":"admin.example.invalid:8443/path?ignored=true"}}`), nil
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

func TestUnitCRUDAndAssignmentsAgainstPinnedRequests(t *testing.T) {
	t.Parallel()
	cases := pinnedCases(t)
	for _, test := range []struct {
		name   string
		steps  []string
		invoke func(context.Context, *policy.Service) error
	}{
		{name: "get", steps: []string{"policy-read-escaped"}, invoke: func(ctx context.Context, s *policy.Service) error {
			result, _, err := s.GetByIdentity(ctx, "tag:Fixture / Policy")
			if err == nil && result.Identity != "tag:Fixture / Policy" {
				return errors.New("identity lost")
			}
			return err
		}},
		{name: "list", steps: []string{"policy-list"}, invoke: func(ctx context.Context, s *policy.Service) error {
			result, _, err := s.List(ctx)
			if err == nil && result == nil {
				return errors.New("empty array became nil")
			}
			return err
		}},
		{name: "create and readback", steps: []string{"policy-create", "policy-read-escaped"}, invoke: func(ctx context.Context, s *policy.Service) error {
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
		{name: "group accepted", steps: []string{"group-assignment-accepted"}, invoke: func(ctx context.Context, s *policy.Service) error {
			rank := 2
			result, _, err := s.AssignToGroup(ctx, "fixture/group", "Fixture", &rank)
			if err == nil && (!result.Accepted || result.OperationID != "fixture-operation") {
				return errors.New("accepted handle lost")
			}
			return err
		}},
		{name: "group no content", steps: []string{"group-assignment-no-content"}, invoke: func(ctx context.Context, s *policy.Service) error {
			rank := 2
			result, _, err := s.AssignToGroup(ctx, "fixture/group", "Fixture", &rank)
			if err == nil && result.Accepted {
				return errors.New("204 became async acceptance")
			}
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
			if err := test.invoke(t.Context(), sdk.TeamsAdminAPI.ApplicationAccessPolicies); err != nil {
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
	cases := pinnedCases(t)
	failed := cases["policy-forbidden"]
	failed.URL = cases["policy-read-escaped"].URL
	sdk, calls := scriptedClient(t, []wireCase{cases["policy-create"], failed})
	description := ""
	result, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.Create(t.Context(), policy.CreateRequest{Identity: "tag:Fixture / Policy", AppIDs: policy.ReplaceAppIDs([]string{"fixture-app"}), Description: &description})
	var partial *policy.ReadBackError
	if !errors.As(err, &partial) || !errors.Is(err, auth.ErrForbidden) {
		t.Fatalf("partial failure lost: %v", err)
	}
	if result.Identity != "tag:Fixture / Policy" || response.StatusCode() != 403 || partial.CreationResponse.StatusCode() != 201 || *calls != 3 {
		t.Fatal("creation metadata lost or cleanup/replay occurred")
	}
}

func TestUnitListFilteredUsesLocalTagAwareMatching(t *testing.T) {
	t.Parallel()
	cases := pinnedCases(t)
	step := cases["policy-list"]
	step.Response = `[{"Identity":"Tag:Fixture","AppIds":[],"Future":false},{"Identity":"Other","AppIds":[]}]`
	sdk, calls := scriptedClient(t, []wireCase{step})
	items, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.ListFiltered(t.Context(), "fix*")
	if err != nil || response.StatusCode() != 200 || len(items) != 1 || items[0].Identity != "Tag:Fixture" {
		t.Fatal("local filter lost identity/tag behavior")
	}
	if *calls != 2 {
		t.Fatal("filter made extra HTTP calls")
	}
	if string(items[0].RawFields["Future"]) != "false" {
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
	cases := pinnedCases(t)
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
