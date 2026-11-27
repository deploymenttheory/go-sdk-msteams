package policies_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	assignment "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

// These authored HTTP fixtures exercise request and result behavior.
// Captured service responses are loaded separately from testdata.
func TestUnitPolicyCompositionConflictUsersFeedBatchWithPartialOutcome(t *testing.T) {
	t.Parallel()
	const conflict = `{"TotalCountUsersWithConflict":2,"UsersWithConflictDueDirectAssignmentsCount":2,"UsersWithConflictDueDirectAssignmentsDetail":[{"UserId":"fixture/a","PolicyInstance":"fixture-policy","AssignmentSource":"Direct"},{"UserId":"fixture/b","PolicyInstance":"fixture-policy","AssignmentSource":"Direct"}],"UsersWithConflictDueGroupAssignmentsCount":0,"UsersWithConflictDueGroupAssignmentsDetail":[],"Future":false}`
	const handle = "fixture/operation:@"
	const receipt = `{"OperationId":"` + handle + `","Future":false}`
	const status = `{"operationId":"` + handle + `","operationName":"fixture batch","overallStatus":"Completed","notStartedCount":0,"inProgressCount":0,"completedCount":1,"errorCount":1,"completedTime":null,"userState":[{"id":"fixture/a","state":"Completed","result":""},{"id":"fixture/b","state":"Error","result":"fixture rejection"}],"Future":false}`
	discovery := policyCompositionFixture(t, "internal/configapi", "service_discovery.json")
	steps := []testutil.Exchange{
		assignmentCompositionWire("GET", "/Skype.Policy/groupPolicyAssignments/fixture-group/policyTypes/TeamsMeetingPolicy/conflict", "Get-GroupPolicyAssignmentConflict", nil, 200, conflict, false),
		{Request: testutil.WireRequest{Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: map[string][]string{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Body: discovery}},
		assignmentCompositionWire("POST", "/Skype.Policy/assignments/operations?operationName=fixture%20batch", "New-CsBatchPolicyAssignmentOperation", json.RawMessage(`{"identities":["fixture/a","fixture/b"],"policy":{"policyType":"TeamsMeetingPolicy","policyName":"fixture-policy"},"additionalParameters":{"Flag":false}}`), 202, receipt, true),
		assignmentCompositionWire("GET", "/Skype.Policy/assignments/operations/fixture%2Foperation%3A%40", "Get-CsBatchPolicyAssignmentOperation", nil, 200, status, true),
	}
	r, err := testutil.Run(t.Context(), testutil.Case{Name: "conflict-users-batch-partial",

		Steps: steps, Want: testutil.Expected{Status: 200, Body: []byte(status), JSON: json.RawMessage(status)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			before, _, err := sdk.TeamsAdminAPI.AssignmentMigrations.GetConflicts(ctx, "fixture-group", "TeamsMeetingPolicy")
			if err != nil || before == nil || before.TotalCountUsersWithConflict == nil || *before.TotalCountUsersWithConflict != 2 || len(before.UsersWithConflictDueDirectAssignmentsDetail) != 2 || before.UsersWithConflictDueGroupAssignmentsCount == nil || *before.UsersWithConflictDueGroupAssignmentsCount != 0 {
				t.Fatal("batch preparation lost conflict count or typed user details")
			}
			ids := make([]string, 0, 2)
			for _, detail := range before.UsersWithConflictDueDirectAssignmentsDetail {
				if detail == nil || detail.UserID == nil {
					t.Fatal("conflict detail lost its user identity")
				}
				ids = append(ids, *detail.UserID)
			}
			if !reflect.DeepEqual(ids, []string{"fixture/a", "fixture/b"}) {
				t.Fatal("batch changed the source user order or identity namespace")
			}
			s := sdk.TeamsAdminAPI.PolicyAssignments
			accepted, response, err := s.CreateBatch(ctx, assignment.BatchCreateRequest{Identities: ids, PolicyType: policy.Set("TeamsMeetingPolicy"), PolicyName: policy.Set("fixture-policy"), OperationName: "fixture batch", AdditionalParameters: map[string]json.RawMessage{"Flag": json.RawMessage(`false`)}})
			if err != nil || response == nil || response.StatusCode() != 202 || accepted == nil || !accepted.HTTPAccepted || accepted.OperationID == nil || *accepted.OperationID != handle || string(accepted.RawFields["Future"]) != "false" {
				t.Fatal("batch receipt lost its distinct opaque operation identity")
			}
			// Caller explicitly reads the receipt's handle; the SDK does not poll.
			value, response, err := s.GetBatch(ctx, *accepted.OperationID)
			if err == nil && (value == nil || value.OperationID == nil || *value.OperationID != handle || value.CompletedCount == nil || *value.CompletedCount != 1 || value.ErrorCount == nil || *value.ErrorCount != 1 || value.InProgressCount == nil || *value.InProgressCount != 0 || value.CompletedTime != nil || string(value.RawFields["completedTime"]) != "null" || len(value.UserState) != 2 || value.UserState[0] == nil || value.UserState[0].ID == nil || *value.UserState[0].ID != ids[0] || value.UserState[0].Result == nil || *value.UserState[0].Result != "" || value.UserState[1] == nil || value.UserState[1].ID == nil || *value.UserState[1].ID != ids[1] || value.UserState[1].State == nil || *value.UserState[1].State != "Error" || value.UserState[1].Result == nil || *value.UserState[1].Result != "fixture rejection") {
				t.Fatal("batch detail erased the individual failure or confused acceptance with convergence")
			}
			if !accepted.HTTPAccepted || string(accepted.RawFields["Future"]) != "false" || string(before.RawFields["Future"]) != "false" || *before.TotalCountUsersWithConflict != 2 || *before.UsersWithConflictDueDirectAssignmentsDetail[1].UserID != ids[1] {
				t.Fatal("later partial outcome mutated earlier conflict or acceptance evidence")
			}
			return testutil.Observation{Value: value, Response: response, Err: err}
		}})
	if err != nil || !r.Passed() {
		t.Fatalf("conflict-to-batch composition differs: %v %+v", err, r.Differences())
	}
}
