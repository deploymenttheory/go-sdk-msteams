package policies_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

// These authored HTTP fixtures exercise request and result behavior.
// Captured service responses are loaded separately from testdata.
func TestUnitPolicyCompositionMigrationDecodeFailureKeepsIndependentReadUsable(t *testing.T) {
	t.Parallel()
	const root = "/Skype.Policy/groupPolicyAssignments"
	const group = "00000000-0000-4000-8000-000000000001"
	const jobID = "00000000-0000-4000-8000-000000000002"
	const recommendation = `[{"policyName":"fixture-policy","daCount":0}]`
	const job = `{"MigrationEventId":"` + jobID + `","GroupId":"` + group + `","PolicyType":"TeamsMeetingPolicy","PolicyInstance":"fixture-policy","State":"PopulationInProgress","DirectAssignmentsProcessed":0,"Future":false}`
	const body = `{"authority":"Tenant","policytype":"TeamsMeetingPolicy","policyname":"fixture-policy","groupid":"` + group + `"}`
	for _, tc := range []struct{ name, action, command string }{
		{"start", "migrate", "Invoke-StartDirectToGroupAssignmentMigration"},
		{"clear", "cleanup", "Invoke-ClearDirectToGroupAssignmentMigration"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			read := assignmentCompositionWire("GET", root+"/directAssignmentsMigration/"+jobID, "Get-DirectToGroupAssignmentsMigrationStatus", nil, 200, job, false)
			steps := []testutil.Exchange{
				read,
				assignmentCompositionWire("GET", root+"/groups/policyTypes/TeamsMeetingPolicy/recommendations?groupThreshold=500", "Get-GroupAssignmentRecommendationsPerPolicyName", nil, 200, recommendation, false),
				assignmentCompositionWire("POST", root+"/directAssignmentsMigration/"+tc.action+"?allowNonEmptyGroup=False", tc.command, json.RawMessage(body), 200, `{"State":`, false),
				read,
			}
			r, err := testutil.Run(t.Context(), testutil.Case{Name: "migration-decode-isolation-" + tc.name,

				Steps: steps, Want: testutil.Expected{Status: 200, Body: []byte(job), JSON: json.RawMessage(job)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					s := sdk.TeamsAdminAPI.AssignmentMigrations
					before, _, err := s.GetByID(ctx, jobID)
					if err != nil || before == nil || before.MigrationEventID == nil || *before.MigrationEventID != jobID || before.GroupID == nil || *before.GroupID != group || before.DirectAssignmentsProcessed == nil || *before.DirectAssignmentsProcessed != 0 {
						t.Fatal("independent migration read lost its typed identity or zero progress")
					}
					names, _, err := s.RecommendationsByPolicyName(ctx, "groups", "TeamsMeetingPolicy", migration.RecommendationOptions{})
					if err != nil || len(names) != 1 || names[0] == nil || names[0].PolicyName == nil || *names[0].PolicyName != "fixture-policy" || names[0].DirectAssignmentsCount == nil || *names[0].DirectAssignmentsCount != 0 {
						t.Fatal("submission lacked its typed recommendation")
					}
					input := migration.CreateRequest{Authority: "Tenant", PolicyType: "TeamsMeetingPolicy", PolicyName: *names[0].PolicyName, GroupID: group}
					submit := s.Create
					if tc.action == "cleanup" {
						submit = s.ClearDirectAssignments
					}
					accepted, response, err := submit(ctx, input)
					if accepted != nil || response == nil || response.StatusCode() != 200 || string(response.Body()) != `{"State":` || !errors.Is(err, client.ErrDecode) || string(before.RawFields["Future"]) != "false" || *names[0].PolicyName != "fixture-policy" {
						t.Fatal("malformed submission fabricated a handle or erased earlier evidence")
					}
					// Explicit read, never automatic retry, clear, deletion or polling.
					last, response, err := s.GetByID(ctx, jobID)
					if err == nil && (last == nil || !reflect.DeepEqual(last.RawFields, before.RawFields)) {
						t.Fatal("submission decode failure poisoned the independent read or snapshot")
					}
					return testutil.Observation{Value: last, Response: response, Err: err}
				}})
			if err != nil || !r.Passed() {
				t.Fatalf("migration decode isolation differs: %v %+v", err, r.Differences())
			}
		})
	}
}
