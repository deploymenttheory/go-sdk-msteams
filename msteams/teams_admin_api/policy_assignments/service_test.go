package policyassignments_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const forest = "https://admin.example.invalid/"

func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	if len(c.Steps) == 0 {

	} else {

		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, err := testutil.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("assignment contract %s differs: %v %+v", c.Name, err, r.Differences())
	}
	return r
}

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Body: body}}
}

func wire(method, url, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}}
	h.Set("X-MS-CmdletName", command)
	h.Set("X-MS-Target-Uri", forest)
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: url, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func batchInput(name string) resource.BatchCreateRequest {
	input := resource.BatchCreateRequest{Identities: []string{}}
	if name == "batch-create" {
		input.Identities = []string{"fixture/a", "fixture/a", "fixture/b"}
		input.PolicyType, input.PolicyName = policy.Set("Fixture"), policy.Set("")
		input.OperationName = "Fixture /?&:@"
		input.AdditionalParameters = map[string]json.RawMessage{"flag": []byte("false"), "n": []byte("0"), "empty": []byte(`""`), "null": []byte("null")}
	}
	if name == "batch-create-nested-additional" {
		input.AdditionalParameters = map[string]json.RawMessage{"nested": []byte(`{"flag":false,"null":null}`), "array": []byte(`[0,false,"",null,{"null":null}]`)}
	}
	return input
}

func invoke(name string, ctx context.Context, sdk *msteams.Client) testutil.Observation {
	s := sdk.TeamsAdminAPI.PolicyAssignments
	var o testutil.Observation
	groupID, policyType := "fixture/group:@", "Fixture/Policy:@"
	switch {
	case name == "batch-list" || name == "batch-list-null":
		o.Value, o.Response, o.Err = s.ListBatches(ctx, "")
	case name == "batch-list-filter":
		o.Value, o.Response, o.Err = s.ListBatches(ctx, "Future /?&:@")
	case name == "batch-get":
		o.Value, o.Response, o.Err = s.GetBatch(ctx, "fixture/operation:@")
	case name == "group-list":
		o.Value, o.Response, o.Err = s.ListGroup(ctx, groupID)
	case name == "group-list-type":
		o.Value, o.Response, o.Err = s.ListGroupsByPolicy(ctx, "Fixture /?&:@")
	case name == "group-list-all":
		o.Value, o.Response, o.Err = s.ListGroupsByPolicy(ctx, "")
	case name == "user-effective":
		o.Value, o.Response, o.Err = s.GetUser(ctx, "fixture/user:@", "Fixture /?&:@")
	case name == "user-effective-all":
		o.Value, o.Response, o.Err = s.GetUser(ctx, "fixture/user:@", "")
	case name == "group-create":
		o.Value, o.Response, o.Err = s.CreateGroup(ctx, groupID, policyType, resource.GroupSettings{PolicyName: policy.Set("Fixture"), Priority: policy.Set(int32(0))})
	case strings.HasPrefix(name, "group-update"):
		settings := resource.GroupSettings{}
		if name == "group-update" {
			settings.PolicyName, settings.Priority = policy.Set("UpdatedFixture"), policy.Set(int32(0))
		}
		o.Value, o.Response, o.Err = s.UpdateGroup(ctx, groupID, policyType, settings)
	case strings.HasPrefix(name, "group-delete"):
		o.Value, o.Response, o.Err = s.DeleteGroup(ctx, groupID, policyType)
	case strings.HasPrefix(name, "batch-create"):
		o.Value, o.Response, o.Err = s.CreateBatch(ctx, batchInput(name))
	default:
		o.Value, o.Response, o.Err = s.GetGroup(ctx, groupID, policyType)
	}
	return o
}

// Additional operation, no exported command claim.

func TestUnitAssignmentRequestVectors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, method, url string
		body              json.RawMessage
	}{
		{name: "batch-list", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations", body: nil},
		{name: "batch-list-filter", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations?status=Future%20%2F%3F%26%3A%40", body: nil},
		{name: "batch-get", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations/fixture%2Foperation%3A%40", body: nil},
		{name: "group-get", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes/Fixture%2FPolicy%3A%40", body: nil},
		{name: "group-list", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes", body: nil},
		{name: "group-list-type", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments?policyType=Fixture%20%2F%3F%26%3A%40", body: nil},
		{name: "group-list-all", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments", body: nil},
		{name: "user-effective", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.User/users/fixture%2Fuser%3A%40/policies?policyType=Fixture%20%2F%3F%26%3A%40", body: nil},
		{name: "user-effective-all", method: "GET", url: "https://api.interfaces.records.teams.microsoft.com/Skype.User/users/fixture%2Fuser%3A%40/policies", body: nil},
		{name: "group-create", method: "PUT", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes/Fixture%2FPolicy%3A%40", body: json.RawMessage("{\"policyName\":\"Fixture\",\"priority\":0}")},
		{name: "group-delete", method: "DELETE", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes/Fixture%2FPolicy%3A%40", body: nil},
		{name: "group-delete-no-content", method: "DELETE", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes/Fixture%2FPolicy%3A%40", body: nil},
		{name: "group-update", method: "PATCH", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes/Fixture%2FPolicy%3A%40", body: json.RawMessage("{\"policyName\":\"UpdatedFixture\",\"priority\":0}")},
		{name: "group-update-empty", method: "PATCH", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes/Fixture%2FPolicy%3A%40", body: json.RawMessage("{}")},
		{name: "batch-create", method: "POST", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations?operationName=Fixture%20%2F%3F%26%3A%40", body: json.RawMessage("{\"policy\":{\"policyType\":\"Fixture\",\"policyName\":\"\"},\"identities\":[\"fixture/a\",\"fixture/a\",\"fixture/b\"],\"additionalParameters\":{\"flag\":false,\"n\":0,\"empty\":\"\"}}")},
		{name: "batch-create-empty", method: "POST", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations", body: json.RawMessage("{\"identities\":[]}")},
		{name: "batch-create-nested-additional", method: "POST", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/assignments/operations", body: json.RawMessage("{\"identities\":[],\"additionalParameters\":{\"nested\":{\"flag\":false},\"array\":[0,false,\"\",\"null\",{}]}}")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			command := commandHeader(tc.name)
			for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
				run(t, testutil.Case{Name: tc.name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{discovery(t), wire(tc.method, tc.url, command, tc.body, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					return invoke(tc.name, ctx, sdk)
				}, Want: testutil.Expected{Error: auth.ErrForbidden}})
			}
		})
	}
}

func commandHeader(name string) string {
	command := "Get-CsGroupPolicyAssignment"
	switch {
	case strings.HasPrefix(name, "group-update"):
		return "Set-CsGroupPolicyAssignment"
	case name == "group-list":
	case name == "group-list-type" || name == "group-list-all":
	case strings.HasPrefix(name, "user-effective"):
		command = "Get-CsUserPolicyAssignment"
	case strings.HasPrefix(name, "batch-list"):
		command = "Get-CsBatchPolicyAssignmentOperation"
	case name == "batch-get":
		command = "Get-CsBatchPolicyAssignmentOperation"
	case strings.HasPrefix(name, "batch-create"):
		command = "New-CsBatchPolicyAssignmentOperation"
	case strings.HasPrefix(name, "group-create"):
		command = "New-CsGroupPolicyAssignment"
	case strings.HasPrefix(name, "group-delete"):
		command = "Remove-CsGroupPolicyAssignment"
	}
	return command
}
