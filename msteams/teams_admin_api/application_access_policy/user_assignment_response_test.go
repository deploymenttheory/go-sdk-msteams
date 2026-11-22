package applicationaccesspolicy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestUnitUserAssignmentUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	const user = "00000000-0000-4000-8000-000000000010"
	const get = "/Skype.User/users/" + user + "/policies?policyType=ApplicationAccessPolicy"
	const grant = "/Skype.Policy/users/" + user + "/policies/ApplicationAccessPolicy"
	fixtures, err := os.OpenRoot("../policy_assignments/testdata")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixtures.Close(); err != nil {
			t.Error(err)
		}
	})
	read := func(name string) []byte {
		t.Helper()
		raw, err := fixtures.ReadFile(name + ".json")
		if err != nil || !json.Valid(raw) {
			t.Fatal("invalid actual user-assignment response fixture")
		}
		return raw
	}
	before := read("user_application_access_before")
	assigned := read("user_application_access_assigned")
	restored := read("user_application_access_restored")
	var expected []struct {
		PolicyType string `json:"policyType"`
		PolicyName string `json:"policyName"`
		Sources    []struct {
			PolicyName string `json:"policyName"`
			PolicyType string `json:"policyType"`
			Type       string `json:"assignmentType"`
			Reference  string `json:"reference"`
		} `json:"policySource"`
	}
	if json.Unmarshal(assigned, &expected) != nil || len(expected) != 1 || len(expected[0].Sources) != 1 {
		t.Fatal("actual direct-assignment fixture is incomplete")
	}
	wire := func(method, path, command string, body json.RawMessage, reply []byte) contracttest.Exchange {
		headers := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}}
		headers[http.CanonicalHeaderKey("MPACmdlet")] = nil
		if body != nil {
			headers.Set("Content-Type", "application/json")
		}
		return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: policyOrigin + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: reply}}
	}
	steps := []contracttest.Exchange{
		{Request: contracttest.WireRequest{Method: "GET", URL: policyOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(discoveryResponseFixture(t))}},
		wire("GET", get, "Get-CsUserPolicyAssignment", nil, before),
		wire("PATCH", grant, "Grant-CsApplicationAccessPolicy", json.RawMessage(`{"policyName":"fixture-policy"}`), []byte(applicationResponseFixture(t, "assign_user.body"))),
		wire("GET", get, "Get-CsUserPolicyAssignment", nil, assigned),
		wire("PATCH", grant, "Grant-CsApplicationAccessPolicy", json.RawMessage(`{"policyName":""}`), []byte(applicationResponseFixture(t, "clear_user.body"))),
		wire("GET", get, "Get-CsUserPolicyAssignment", nil, restored),
	}
	// The assignment service and application-policy service own separate discovery adapters.
	steps = slices.Insert(steps, 2, steps[0])
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "owned-user-assignment", Family: "application_access_policy", Scope: "service", Source: "msteams/teams_admin_api/policy_assignments/testdata/user_application_access_assigned.json", Steps: steps, Want: contracttest.Expected{Status: 200, Body: restored, JSON: restored}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		values, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, user, "ApplicationAccessPolicy")
		if err != nil || len(values) != 0 || response == nil || response.StatusCode() != 200 {
			t.Fatal("actual initial assignment state differs")
		}
		response, err = sdk.TeamsAdminAPI.ApplicationAccessPolicies.AssignToUser(ctx, user, "fixture-policy")
		if err != nil || response == nil || response.StatusCode() != 200 {
			t.Fatal("actual grant request or response differs")
		}
		values, response, err = sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, user, "ApplicationAccessPolicy")
		if err != nil || response == nil || response.StatusCode() != 200 || len(values) != 1 || values[0] == nil {
			t.Fatal("actual assignment was not decoded")
		}
		value := values[0]
		want := expected[0]
		if value.PolicyName == nil || *value.PolicyName != want.PolicyName || value.PolicyType == nil || *value.PolicyType != want.PolicyType || len(value.PolicySource) != 1 || value.PolicySource[0] == nil {
			t.Fatal("effective policy fields or source list changed")
		}
		source := value.PolicySource[0]
		original := want.Sources[0]
		if source.PolicyName == nil || *source.PolicyName != original.PolicyName || source.PolicyType == nil || *source.PolicyType != original.PolicyType || source.AssignmentType == nil || *source.AssignmentType != original.Type || source.Reference == nil || *source.Reference != original.Reference {
			t.Fatal("actual direct-assignment source was lost")
		}
		var raw []map[string]json.RawMessage
		if json.Unmarshal(assigned, &raw) != nil || !reflect.DeepEqual(value.RawFields, raw[0]) {
			t.Fatal("complete actual assignment snapshot changed")
		}
		response, err = sdk.TeamsAdminAPI.ApplicationAccessPolicies.AssignToUser(ctx, user, "")
		if err != nil || response == nil || response.StatusCode() != 200 {
			t.Fatal("actual clear request or response differs")
		}
		values, response, err = sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, user, "ApplicationAccessPolicy")
		if err != nil || len(values) != 0 {
			t.Fatal("actual cleared assignment state differs")
		}
		return contracttest.Observation{Value: values, Response: response, Err: err}
	}})
	if err != nil || !result.Passed() {
		t.Fatalf("actual user-assignment workflow differs: %v %+v", err, result.Differences())
	}
}
