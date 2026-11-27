package assignmentmigration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

func TestUnitGPARejectsInvalidInputBeforeCredentialAcquisition(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "invalid-input-and-context", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.AssignmentMigrations
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByID(ctx, id); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid job ID accepted")}
			}
			if _, _, e := s.RecommendationsByPolicyType(ctx, id, resource.RecommendationOptions{}); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid entity accepted")}
			}
			if _, _, e := s.RecommendationsByPolicyName(ctx, "groups", id, resource.RecommendationOptions{}); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid recommendation policy type accepted")}
			}
			if _, _, e := s.GetConflicts(ctx, id, "FixturePolicy"); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid conflict group accepted")}
			}
			if _, _, e := s.GetConflicts(ctx, "fixture-group", id); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid conflict policy type accepted")}
			}
		}
		for _, field := range []string{"authority", "type", "name", "group"} {
			input := request(false)
			switch field {
			case "authority":
				input.Authority = ""
			case "type":
				input.PolicyType = ""
			case "name":
				input.PolicyName = ""
			case "group":
				input.GroupID = "bad-guid"
			}
			if _, _, e := s.Create(ctx, input); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid migration submission accepted")}
			}
			if _, _, e := s.ClearDirectAssignments(ctx, input); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid cleanup submission accepted")}
			}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return testutil.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).List(ctx); !errors.Is(e, client.ErrDestination) {
			return testutil.Observation{Err: errors.New("wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return testutil.Observation{Err: errors.New("missing service accepted")}
			}
		}

		//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before credentials or HTTP.
		if _, _, e := s.List(nil); !errors.Is(e, client.ErrNilContext) {
			return testutil.Observation{Err: errors.New("nil context accepted")}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.Create(canceled, request(false)); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("submission cancellation lost")}
		}
		return testutil.Observation{}
	}})
}

func TestUnitGPAMalformedAcceptedResponsesStayErrorsWithoutResubmission(t *testing.T) {
	t.Parallel()

	for _, op := range []string{"create", "clear"} {
		for i, reply := range []string{"", "not-json", `[]`, `{"state":{}}`, `{"directAssignmentsProcessed":2147483648}`} {
			suffix := "directAssignmentsMigration/migrate?allowNonEmptyGroup=False"
			if op == "clear" {
				suffix = "directAssignmentsMigration/cleanup?allowNonEmptyGroup=False"
			}
			run(t, testutil.Case{Name: fmt.Sprintf("%s-malformed-%d", op, i), Want: testutil.Expected{Error: client.ErrDecode}, Steps: []testutil.Exchange{wire(op, "POST", suffix, json.RawMessage(body), 202, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return invoke(ctx, sdk, op, false)
			}})
		}
	}
}

func TestUnitGPARequestCanonicalGUIDOnRejectedSubmission(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "canonical-guid-rejected-submission", Want: testutil.Expected{APIError: &client.
		APIError{Backend: client.TeamsAdmin,
		StatusCode: 400,
	}}, Steps: []testutil.Exchange{wire("create", "POST", "directAssignmentsMigration/migrate?allowNonEmptyGroup=False", json.RawMessage(`{"authority":"Tenant","policytype":"FixturePolicy","policyname":"Fixture","groupid":"abcdef01-0000-4000-8000-000000000001"}`), 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		input := request(false)
		input.GroupID = "ABCDEF01-0000-4000-8000-000000000001"
		v, r, e := sdk.TeamsAdminAPI.AssignmentMigrations.Create(ctx, input)
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
}
