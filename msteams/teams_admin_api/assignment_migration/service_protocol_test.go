package assignmentmigration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

// Authored offline vectors, not captured service replies. Pinned Core.dll
// 68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49:
// Custom.GPA Get status d__6 MoveNext IL336/458 deserialize List<Job>/Job;
// Start migration d__22 IL277/389 and Clear d__22 IL338 deserialize Job.
// Property/JsonProperty metadata supplies the shapes below. Future fields and
// nullable presence are explicit SDK preservation extensions, not CLR defaults.
// Full owning types and reproduction: pol90-current-test-audit.md source audit.
const authoredJob = `{"MigrationEventId":"00000000-0000-4000-8000-000000000002","TenantId":"00000000-0000-4000-8000-000000000003","GroupId":"00000000-0000-4000-8000-000000000001","GroupName":"Fixture","PolicyType":"FixturePolicy","PolicyInstance":"Fixture","State":"PopulationInProgress","StartTime":"2026-01-01T12:00:00Z","CompletionTime":null,"DirectAssignmentsProcessed":0,"Future":false}`

func verifyAuthoredJob(t *testing.T, job *resource.Job) {
	t.Helper()
	if job == nil || job.MigrationEventID == nil || *job.MigrationEventID != "00000000-0000-4000-8000-000000000002" || job.GroupID == nil || *job.GroupID != request(false).GroupID || job.DirectAssignmentsProcessed == nil || *job.DirectAssignmentsProcessed != 0 || job.CompletionTime != nil || string(job.RawFields["CompletionTime"]) != "null" || string(job.RawFields["Future"]) != "false" {
		t.Fatal("typed job identity, explicit zero/null, or future snapshot lost")
	}
	if state, ok := job.State.Text(); !ok || state != resource.PopulationInProgress {
		t.Fatal("migration state changed or was interpreted as completion")
	}
}

func TestUnitGPAAuthoredTypedResponses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, operation, method, suffix, reply string
		body                                   json.RawMessage
	}{
		{"list", "list", "GET", "directAssignmentsMigration", "[" + authoredJob + "]", nil},
		{"read", "read", "GET", "directAssignmentsMigration/fixture-job", authoredJob, nil},
		{"types", "types", "GET", "groups/recommendations?groupThreshold=500", `[{"policyType":"FixturePolicy","daCount":0,"Future":false}]`, nil},
		{"names", "names", "GET", "groups/policyTypes/FixturePolicy/recommendations?groupThreshold=500", `[{"policyName":"Fixture","daCount":2}]`, nil},
		{"conflicts", "conflicts", "GET", "fixture-group/policyTypes/FixturePolicy/conflict", `{"TotalCountUsersWithConflict":1,"UsersWithConflictDueDirectAssignmentsCount":1,"UsersWithConflictDueDirectAssignmentsDetail":[{"UserId":"fixture-user","PolicyInstance":"Fixture","AssignmentSource":"Direct"}],"UsersWithConflictDueGroupAssignmentsCount":0,"UsersWithConflictDueGroupAssignmentsDetail":[],"Future":null}`, nil},
		{"create", "create", "POST", "directAssignmentsMigration/migrate?allowNonEmptyGroup=False", authoredJob, json.RawMessage(body)},
		{"clear", "clear", "POST", "directAssignmentsMigration/cleanup?allowNonEmptyGroup=False", authoredJob, json.RawMessage(body)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := contracttest.Run(t.Context(), contracttest.Case{Name: "typed-" + tc.name, Family: "assignment_migration", Scope: "service", Source: "msteams/teams_admin_api/assignment_migration/service_protocol_test.go", Mappings: mapping(tc.operation), Steps: []contracttest.Exchange{wire(tc.operation, tc.method, tc.suffix, tc.body, 200, tc.reply)}, Want: contracttest.Expected{Status: 200, Body: []byte(tc.reply), JSON: json.RawMessage(tc.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				observed := invoke(ctx, sdk, tc.operation, false)
				if observed.Err != nil {
					return observed
				}
				switch tc.operation {
				case "list":
					values, ok := observed.Value.([]*resource.Job)
					if !ok || len(values) != 1 {
						t.Fatal("list result is not a nonempty typed job slice")
					}
					verifyAuthoredJob(t, values[0])
				case "types":
					values, ok := observed.Value.([]*resource.PolicyTypeRecommendation)
					if !ok || len(values) != 1 || values[0] == nil || values[0].PolicyType == nil || *values[0].PolicyType != "FixturePolicy" || values[0].DirectAssignmentsCount == nil || *values[0].DirectAssignmentsCount != 0 || string(values[0].RawFields["Future"]) != "false" {
						t.Fatal("typed recommendation lost type, zero count, or raw extension")
					}
				case "names":
					values, ok := observed.Value.([]*resource.PolicyNameRecommendation)
					if !ok || len(values) != 1 || values[0] == nil || values[0].PolicyName == nil || *values[0].PolicyName != "Fixture" || values[0].DirectAssignmentsCount == nil || *values[0].DirectAssignmentsCount != 2 {
						t.Fatal("typed recommendation lost name or count")
					}
				case "conflicts":
					value, ok := observed.Value.(*resource.Conflict)
					if !ok || value == nil || value.TotalCountUsersWithConflict == nil || *value.TotalCountUsersWithConflict != 1 || value.UsersWithConflictDueGroupAssignmentsCount == nil || *value.UsersWithConflictDueGroupAssignmentsCount != 0 || len(value.UsersWithConflictDueGroupAssignmentsDetail) != 0 || len(value.UsersWithConflictDueDirectAssignmentsDetail) != 1 || value.UsersWithConflictDueDirectAssignmentsDetail[0] == nil || value.UsersWithConflictDueDirectAssignmentsDetail[0].UserID == nil || *value.UsersWithConflictDueDirectAssignmentsDetail[0].UserID != "fixture-user" || string(value.RawFields["Future"]) != "null" {
						t.Fatal("typed conflict lost counts, detail identity, or null extension")
					}
				default:
					job, ok := observed.Value.(*resource.Job)
					if !ok {
						t.Fatal("read/action response is not a typed job")
					}
					verifyAuthoredJob(t, job)
				}
				return observed
			}})
			if err != nil || !r.Passed() {
				t.Fatalf("authored typed GPA protocol differs: %v %+v", err, r.Differences())
			}
		})
	}
}

func TestUnitGPAAuthoredMalformedSuccessRetainsResponse(t *testing.T) {
	t.Parallel()
	// These intentionally adversarial bodies test the documented bare array,
	// object, Int32 count and enum scalar contracts; none is a service capture.
	for _, tc := range []struct{ name, operation, suffix, reply string }{
		{"malformed-object", "read", "directAssignmentsMigration/fixture-job", `{"State":`},
		{"empty-object-body", "read", "directAssignmentsMigration/fixture-job", ""},
		{"collection-envelope", "list", "directAssignmentsMigration", `{"value":[]}`},
		{"count-overflow", "read", "directAssignmentsMigration/fixture-job", `{"DirectAssignmentsProcessed":2147483648}`},
		{"invalid-state", "read", "directAssignmentsMigration/fixture-job", `{"State":{}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := contracttest.Run(t.Context(), contracttest.Case{Name: tc.name, Family: "assignment_migration", Scope: "service", Source: "msteams/teams_admin_api/assignment_migration/service_protocol_test.go", Mappings: mapping(tc.operation), Steps: []contracttest.Exchange{wire(tc.operation, "GET", tc.suffix, nil, 200, tc.reply)}, Want: contracttest.Expected{Status: 200, Body: []byte(tc.reply), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return invoke(ctx, sdk, tc.operation, false)
			}})
			if err != nil || !r.Passed() {
				t.Fatalf("malformed successful GPA response differs: %v %+v", err, r.Differences())
			}
		})
	}
}
