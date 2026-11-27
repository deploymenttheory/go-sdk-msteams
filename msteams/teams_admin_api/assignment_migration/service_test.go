package assignmentmigration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

const fixture = "msteams/teams_admin_api/assignment_migration/testdata/list_400.json"
const base = "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/"
const body = `{"authority":"Tenant","policytype":"FixturePolicy","policyname":"Fixture","groupid":"00000000-0000-4000-8000-000000000001"}`

func request(allow bool) resource.CreateRequest {
	return resource.CreateRequest{Authority: "Tenant", PolicyType: "FixturePolicy", PolicyName: "Fixture", GroupID: "00000000-0000-4000-8000-000000000001", AllowNonEmptyGroup: allow}
}
func command(operation string) string {
	switch operation {
	case "types":
		return "Get-GroupAssignmentRecommendationsPerPolicyType"
	case "names":
		return "Get-GroupAssignmentRecommendationsPerPolicyName"
	case "conflicts":
		return "Get-GroupPolicyAssignmentConflict"
	case "create":
		return "Invoke-StartDirectToGroupAssignmentMigration"
	case "clear":
		return "Invoke-ClearDirectToGroupAssignmentMigration"
	default:
		return "Get-DirectToGroupAssignmentsMigrationStatus"
	}
}
func mapping(operation string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: command(operation), ParameterSet: "__AllParameterSets"}}
}
func wire(operation, method, suffix string, input json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command(operation))
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if input != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: base + suffix, Audience: auth.TeamsAdminAudience, Headers: h, JSON: input}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "assignment_migration", fixture
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := contracttest.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("GPA contract differs: %v %+v", e, r.Differences())
	}
	return r
}

func invoke(ctx context.Context, sdk *msteams.Client, operation string, allow bool) contracttest.Observation {
	s := sdk.TeamsAdminAPI.AssignmentMigrations
	switch operation {
	case "list":
		v, r, e := s.List(ctx)
		return contracttest.Observation{Value: v, Response: r, Err: e}
	case "types":
		v, r, e := s.RecommendationsByPolicyType(ctx, "groups", resource.RecommendationOptions{})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	case "names":
		v, r, e := s.RecommendationsByPolicyName(ctx, "groups", "FixturePolicy", resource.RecommendationOptions{})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	case "conflicts":
		v, r, e := s.GetConflicts(ctx, "fixture-group", "FixturePolicy")
		return contracttest.Observation{Value: v, Response: r, Err: e}
	case "create":
		v, r, e := s.Create(ctx, request(allow))
		return contracttest.Observation{Value: v, Response: r, Err: e}
	case "clear":
		v, r, e := s.ClearDirectAssignments(ctx, request(allow))
		return contracttest.Observation{Value: v, Response: r, Err: e}
	default:
		v, r, e := s.GetByID(ctx, "fixture-job")
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}
}

func TestUnitTypedGPARequestHelper(t *testing.T) {
	t.Parallel()

	requestInputs := struct {
		Cases []struct {
			Name, Operation, Response string
			Status                    int
			Allow                     bool
			Request                   struct {
				Method, Url, Body, CmdletName string

				FailureType *string
			}
		}
	}{
		Cases: []struct {
			Name, Operation, Response string
			Status                    int
			Allow                     bool
			Request                   struct {
				Method, Url, Body, CmdletName string

				FailureType *string
			}
		}{

			{
				Name:      "list",
				Operation: "list",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration",
					CmdletName: "Get-DirectToGroupAssignmentsMigrationStatus",
				},
			},

			{
				Name:      "read",
				Operation: "read",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/fixture-job",
					CmdletName: "Get-DirectToGroupAssignmentsMigrationStatus",
				},
			},

			{
				Name:      "types",
				Operation: "types",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/groups/recommendations?groupThreshold=500",
					CmdletName: "Get-GroupAssignmentRecommendationsPerPolicyType",
				},
			},

			{
				Name:      "names",
				Operation: "names",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/groups/policyTypes/FixturePolicy/recommendations?groupThreshold=500",
					CmdletName: "Get-GroupAssignmentRecommendationsPerPolicyName",
				},
			},

			{
				Name:      "conflicts",
				Operation: "conflicts",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture-group/policyTypes/FixturePolicy/conflict",
					CmdletName: "Get-GroupPolicyAssignmentConflict",
				},
			},

			{
				Name:      "create",
				Operation: "create",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "POST",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/migrate?allowNonEmptyGroup=False",
					Body:       "{\"authority\":\"Tenant\",\"policytype\":\"FixturePolicy\",\"policyname\":\"Fixture\",\"groupid\":\"00000000-0000-4000-8000-000000000001\"}",
					CmdletName: "Invoke-StartDirectToGroupAssignmentMigration",
				},
			},

			{
				Name:      "clear",
				Operation: "clear",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "POST",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/cleanup?allowNonEmptyGroup=False",
					Body:       "{\"authority\":\"Tenant\",\"policytype\":\"FixturePolicy\",\"policyname\":\"Fixture\",\"groupid\":\"00000000-0000-4000-8000-000000000001\"}",
					CmdletName: "Invoke-ClearDirectToGroupAssignmentMigration",
				},
			},

			{
				Name:      "create-allow",
				Operation: "create",
				Status:    400,
				Allow:     true,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "POST",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/migrate?allowNonEmptyGroup=True",
					Body:       "{\"authority\":\"Tenant\",\"policytype\":\"FixturePolicy\",\"policyname\":\"Fixture\",\"groupid\":\"00000000-0000-4000-8000-000000000001\"}",
					CmdletName: "Invoke-StartDirectToGroupAssignmentMigration",
				},
			},

			{
				Name:      "clear-allow",
				Operation: "clear",
				Status:    400,
				Allow:     true,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "POST",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/cleanup?allowNonEmptyGroup=True",
					Body:       "{\"authority\":\"Tenant\",\"policytype\":\"FixturePolicy\",\"policyname\":\"Fixture\",\"groupid\":\"00000000-0000-4000-8000-000000000001\"}",
					CmdletName: "Invoke-ClearDirectToGroupAssignmentMigration",
				},
			},

			{
				Name:      "read-denied",
				Operation: "read",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/fixture-job",
					CmdletName: "Get-DirectToGroupAssignmentsMigrationStatus",
				},
			},

			{
				Name:      "create-denied",
				Operation: "create",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "POST",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/migrate?allowNonEmptyGroup=False",
					Body:       "{\"authority\":\"Tenant\",\"policytype\":\"FixturePolicy\",\"policyname\":\"Fixture\",\"groupid\":\"00000000-0000-4000-8000-000000000001\"}",
					CmdletName: "Invoke-StartDirectToGroupAssignmentMigration",
				},
			},

			{
				Name:      "read-malformed",
				Operation: "read",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/fixture-job",
					CmdletName: "Get-DirectToGroupAssignmentsMigrationStatus",
				},
			},

			{
				Name:      "read-empty",
				Operation: "read",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration/fixture-job",
					CmdletName: "Get-DirectToGroupAssignmentsMigrationStatus",
				},
			},

			{
				Name:      "list-null",
				Operation: "list",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration",
					CmdletName: "Get-DirectToGroupAssignmentsMigrationStatus",
				},
			},

			{
				Name:      "list-envelope",
				Operation: "list",
				Status:    400,
				Allow:     false,
				Request: struct {
					Method, Url, Body, CmdletName string

					FailureType *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/directAssignmentsMigration",
					CmdletName: "Get-DirectToGroupAssignmentsMigrationStatus",
				},
			},
		},
	}

	for _, f := range requestInputs.Cases {
		c := f.Request

		var input json.RawMessage
		if c.Body != "" {
			input = json.RawMessage(c.Body)
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			test := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: mapping(f.Operation), Steps: []contracttest.Exchange{wire(f.Operation, c.Method, strings.TrimPrefix(c.Url, base), input, f.Status, f.Response)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return invoke(ctx, sdk, f.Operation, f.Allow)
			}}
			// Success/decode contracts have separate source-authored vectors.
			if f.Status != 400 || f.Response != "" {
				t.Fatal("rejected-wire vector unexpectedly contains a response contract")
			}
			test.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			run(t, test)
		}
	}
}

func TestUnitGPAEncodesSegmentsAndThresholdSeparately(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, value, suffix string }{
		{"empty", "", ""}, {"opaque", "-1 +&/?#", "-1+%2B%26%2F%3F%23"},
	} {
		run(t, contracttest.Case{Name: "threshold-" + tc.name, Mappings: mapping("types"), Steps: []contracttest.Exchange{wire("types", "GET", "group%2Ffixture%3F%23/recommendations?groupThreshold="+tc.suffix, nil, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.AssignmentMigrations.RecommendationsByPolicyType(ctx, "group/fixture?#", resource.RecommendationOptions{GroupThreshold: &tc.value})
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}, Want: contracttest.Expected{APIError: &client.
			APIError{Backend: client.TeamsAdmin,
			StatusCode: 400,
		}},
		})
	}
	run(t, contracttest.Case{Name: "opaque-job-id", Mappings: mapping("read"), Steps: []contracttest.Exchange{wire("read", "GET", "directAssignmentsMigration/tag%3AFixture%20%2F%20Job%3F%23", nil, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.AssignmentMigrations.GetByID(ctx, "tag:Fixture / Job?#")
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{APIError: &client.
		APIError{Backend: client.TeamsAdmin,
		StatusCode: 400,
	}},
	})
}

func TestUnitGPAErrorsRetainHTTPStatusAndDoNotReplaySubmissions(t *testing.T) {
	t.Parallel()

	for _, op := range []string{"create", "clear"} {
		for _, status := range []int{400, 401, 403, 404, 409, 429, 502, 503} {
			suffix := "directAssignmentsMigration/migrate?allowNonEmptyGroup=False"
			if op == "clear" {
				suffix = "directAssignmentsMigration/cleanup?allowNonEmptyGroup=False"
			}
			run(t, contracttest.Case{Name: fmt.Sprintf("%s-error-%d", op, status), Mappings: mapping(op), Steps: []contracttest.Exchange{wire(op, "POST", suffix, json.RawMessage(body), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				observation := invoke(ctx, sdk, op, false)
				var api *client.APIError
				if observation.Err == nil || !errors.As(observation.Err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(observation.Err), "fixture-private") {
					return contracttest.Observation{Err: errors.New("normalized GPA error/status lost")}
				}
				return contracttest.Observation{Response: observation.Response}
			}})
		}
	}
}
