package assignmentmigration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

const fixture = "contracts/endpoints/8.1.0/fixtures/assignment-migration-dispatch.json"
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
func evidence(t *testing.T, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, err := os.OpenRoot(t.ArtifactDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	f, err := root.Create("assignment-migration-service.json")
	if err != nil {
		t.Fatal(err)
	}
	writeErr, closeErr := contracttest.WriteEvidence(f, revision, results), f.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("GPA evidence write failed")
	}
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

func TestUnitPinnedGPARequestHelper(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../" + fixture)
	if e != nil {
		t.Fatal(e)
	}
	var captured struct {
		Cases []struct {
			Name, Operation, Response string
			Status                    int
			Allow                     bool
			Capture                   struct {
				Method, Url, Body, CmdletName, MpaCmdlet string
				Sends, TokenRequests                     int
				CorrelationIsGuid, HasTargetUri          bool
				FailureType                              *string
			}
		}
	}
	if json.Unmarshal(raw, &captured) != nil || len(captured.Cases) != 15 {
		t.Fatal("invalid pinned GPA fixture")
	}
	var results []contracttest.Result
	for _, f := range captured.Cases {
		c := f.Capture
		if c.Sends != 1 || c.TokenRequests != 0 || !c.CorrelationIsGuid || c.HasTargetUri || c.MpaCmdlet != "true" || c.FailureType != nil || c.CmdletName != command(f.Operation) || !strings.HasPrefix(c.Url, base) {
			t.Fatal("unsafe/incomplete GPA observation")
		}
		var input json.RawMessage
		if c.Body != "" {
			input = json.RawMessage(c.Body)
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			test := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: mapping(f.Operation), Steps: []contracttest.Exchange{wire(f.Operation, c.Method, strings.TrimPrefix(c.Url, base), input, f.Status, f.Response)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return invoke(ctx, sdk, f.Operation, f.Allow)
			}}
			switch f.Name {
			case "read-denied", "create-denied":
				test.Want.Error = auth.ErrForbidden
			case "read-malformed", "read-empty", "list-envelope":
				test.Want.Error = client.ErrDecode
			default:
				test.Want.JSON = json.RawMessage(f.Response)
			}
			results = append(results, run(t, test))
		}
	}
	evidence(t, results)
}

func TestUnitGPAEncodesSegmentsAndThresholdSeparately(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, tc := range []struct{ name, value, suffix string }{
		{"empty", "", ""}, {"opaque", "-1 +&/?#", "-1+%2B%26%2F%3F%23"},
	} {
		results = append(results, run(t, contracttest.Case{Name: "threshold-" + tc.name, Mappings: mapping("types"), Steps: []contracttest.Exchange{wire("types", "GET", "group%2Ffixture%3F%23/recommendations?groupThreshold="+tc.suffix, nil, 200, "[]")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.AssignmentMigrations.RecommendationsByPolicyType(ctx, "group/fixture?#", resource.RecommendationOptions{GroupThreshold: &tc.value})
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}}))
	}
	results = append(results, run(t, contracttest.Case{Name: "opaque-job-id", Mappings: mapping("read"), Steps: []contracttest.Exchange{wire("read", "GET", "directAssignmentsMigration/tag%3AFixture%20%2F%20Job%3F%23", nil, 200, `{"state":"Future","Future":false}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.AssignmentMigrations.GetByID(ctx, "tag:Fixture / Job?#")
		if e == nil {
			state, ok := v.State.Text()
			if !ok || state != "Future" || string(v.RawFields["Future"]) != "false" {
				e = errors.New("open state or future data lost")
			}
		}
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}}))
	evidence(t, results)
}

func TestUnitGPAErrorsRetainHTTPStatusAndDoNotReplaySubmissions(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, op := range []string{"create", "clear"} {
		for _, status := range []int{400, 401, 403, 404, 409, 429, 502, 503} {
			suffix := "directAssignmentsMigration/migrate?allowNonEmptyGroup=False"
			if op == "clear" {
				suffix = "directAssignmentsMigration/cleanup?allowNonEmptyGroup=False"
			}
			results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("%s-error-%d", op, status), Mappings: mapping(op), Steps: []contracttest.Exchange{wire(op, "POST", suffix, json.RawMessage(body), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				observation := invoke(ctx, sdk, op, false)
				var api *client.APIError
				if observation.Err == nil || !errors.As(observation.Err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(observation.Err), "fixture-private") {
					return contracttest.Observation{Err: errors.New("normalized GPA error/status lost")}
				}
				return contracttest.Observation{Response: observation.Response}
			}}))
		}
	}
	evidence(t, results)
}
