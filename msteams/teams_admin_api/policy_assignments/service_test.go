package policyassignments_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

const source = "msteams/teams_admin_api/policy_assignments/service_test.go"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const forest = "https://admin.example.invalid/"

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "policy_assignments"
	if c.Source == "" {
		c.Source = source
	}
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("assignment contract %s differs: %v %+v", c.Name, err, r.Differences())
	}
	return r
}

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Body: body}}
}

func wire(method, url, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}}
	h.Set("X-MS-CmdletName", command)
	h.Set("X-MS-Target-Uri", forest)
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: url, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
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

func invoke(name string, ctx context.Context, sdk *msteams.Client) contracttest.Observation {
	s := sdk.TeamsAdminAPI.PolicyAssignments
	var o contracttest.Observation
	groupID, policyType := "fixture/group:@", "Fixture/Policy:@"
	if name == "group-get-null" || name == "group-get-malformed" || name == "group-forbidden" || name == "group-create-malformed" {
		groupID, policyType = "fixture/group", "Fixture"
	}
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
	case strings.HasPrefix(name, "group-create"):
		settings := resource.GroupSettings{}
		if name == "group-create" {
			settings.PolicyName, settings.Priority = policy.Set("Fixture"), policy.Set(int32(0))
		}
		if name == "group-create-empty-name-negative-rank" || name == "group-create-malformed" {
			settings.PolicyName, settings.Priority = policy.Set(""), policy.Set(int32(-1))
		}
		o.Value, o.Response, o.Err = s.CreateGroup(ctx, groupID, policyType, settings)
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

func mapping(name string) []contracttest.Mapping {
	command, set := "Get-CsGroupPolicyAssignment", "Get"
	switch {
	case strings.HasPrefix(name, "group-update"):
		return nil // Additional operation, no exported command claim.
	case name == "group-list":
		set = "Get1"
	case name == "group-list-type" || name == "group-list-all":
		set = "Get2"
	case strings.HasPrefix(name, "user-effective"):
		command, set = "Get-CsUserPolicyAssignment", "Get"
	case strings.HasPrefix(name, "batch-list"):
		command, set = "Get-CsBatchPolicyAssignmentOperation", "Get"
	case name == "batch-get":
		command, set = "Get-CsBatchPolicyAssignmentOperation", "Get1"
	case strings.HasPrefix(name, "batch-create"):
		command, set = "New-CsBatchPolicyAssignmentOperation", "NewExpanded"
	case strings.HasPrefix(name, "group-create"):
		command, set = "New-CsGroupPolicyAssignment", "NewExpanded"
	case strings.HasPrefix(name, "group-delete"):
		command, set = "Remove-CsGroupPolicyAssignment", "Remove"
	}
	return []contracttest.Mapping{{Command: command, ParameterSet: set}}
}

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
		{name: "group-create-empty", method: "PUT", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes/Fixture%2FPolicy%3A%40", body: json.RawMessage("{}")},
		{name: "group-create-empty-name-negative-rank", method: "PUT", url: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup%3A%40/policyTypes/Fixture%2FPolicy%3A%40", body: json.RawMessage("{\"policyName\":\"\",\"priority\":-1}")},
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
			command := "Set-CsGroupPolicyAssignment"
			mappings := mapping(tc.name)
			if len(mappings) > 0 {
				command = mappings[0].Command
			}
			for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
				run(t, contracttest.Case{Name: tc.name + "-" + string(mode), Mode: mode, Mappings: mappings, Steps: []contracttest.Exchange{discovery(t), wire(tc.method, tc.url, command, tc.body, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					return invoke(tc.name, ctx, sdk)
				}, Want: contracttest.Expected{Error: auth.ErrForbidden}})
			}
		})
	}
}
