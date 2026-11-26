package teamsmedialoggingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_logging_policy"
)

func grantSteps(t *testing.T, method, path, body string, status int, reply []byte) []contracttest.Exchange {
	t.Helper()
	discovery, e := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if e != nil {
		t.Fatal(e)
	}
	headers := http.Header{"Content-Type": {"application/json"}}
	var replyHeaders http.Header
	if len(reply) != 0 {
		replyHeaders = headers
	}
	return []contracttest.Exchange{
		{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: headers, Body: discovery}},
		{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {"Grant-CsTeamsMediaLoggingPolicy"}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}, "MPACmdlet": nil, "Content-Type": {"application/json"}}, JSON: json.RawMessage(body)}, Reply: contracttest.Reply{Status: status, Headers: replyHeaders, Body: reply}},
	}
}

func TestUnitModernGrantsUseSeparateDiscoveryAndWrapperRankRules(t *testing.T) {
	t.Parallel()
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, tc := range []struct {
			name, path, body, set string
			rank                  *int32
			scope                 int
		}{
			{"user", "/Skype.Policy/users/user%2Ffixture/policies/TeamsMediaLoggingPolicy", `{"policyName":"Enabled"}`, "Identity", nil, 0},
			{"user-clear", "/Skype.Policy/users/user%2Ffixture/policies/TeamsMediaLoggingPolicy", `{"policyName":""}`, "Identity", nil, 0},
			{"tenant", "/Skype.Policy/tenants/policies/TeamsMediaLoggingPolicy", `{"policyName":"Enabled","forceSwitchPresent":false}`, "GrantToTenant", nil, 1},
			{"group-unbound", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMediaLoggingPolicy", `{"policyName":"Enabled"}`, "GrantToGroup", nil, 2},
			{"wrapper-zero", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMediaLoggingPolicy", `{"policyName":"Enabled"}`, "GrantToGroup", new(int32(0)), 2},
			{"signed-rank", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMediaLoggingPolicy", `{"policyName":"Enabled","priority":-2}`, "GrantToGroup", new(int32(-2)), 2},
		} {
			method := "PATCH"
			if tc.scope == 2 {
				method = "POST"
			}
			var body map[string]any
			if json.Unmarshal([]byte(tc.body), &body) != nil {
				t.Fatal("invalid case body")
			}
			name, _ := body["policyName"].(string)
			runExplicit(t, contracttest.Case{Name: fmt.Sprintf("%s-%s", mode, tc.name), Mode: mode, Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsMediaLoggingPolicy", ParameterSet: tc.set}}, Steps: grantSteps(t, method, tc.path, tc.body, 503, nil), Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				input := resource.Assignment{}
				switch tc.scope {
				case 0:
					input.User = &resource.UserAssignment{Identity: "user/fixture", PolicyName: name}
				case 1:
					input.Tenant = &resource.TenantAssignment{PolicyName: name}
				case 2:
					input.Group = &resource.GroupGrant{ID: "group/fixture", PolicyName: name, Rank: tc.rank}
				}
				_, r, e := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, input)
				return contracttest.Observation{Response: r, Err: e}
			}})
		}
	}
}

func TestUnitTenantForceAndExplicitZeroPriorityAreLowLevelOnly(t *testing.T) {
	t.Parallel()
	runExplicit(t, contracttest.Case{Name: "tenant-force", Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsMediaLoggingPolicy", ParameterSet: "GrantToTenant"}}, Steps: grantSteps(t, "PATCH", "/Skype.Policy/tenants/policies/TeamsMediaLoggingPolicy", `{"policyName":"","forceSwitchPresent":true}`, 503, nil), Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, e := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.AssignTenantDefault(ctx, "", true)
		return contracttest.Observation{Response: r, Err: e}
	}})
	runExplicit(t, contracttest.Case{Name: "group-explicit-zero", Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsMediaLoggingPolicy", ParameterSet: "GrantToGroup"}}, Steps: grantSteps(t, "POST", "/Skype.Policy/groupPolicyAssignments/group/policyTypes/TeamsMediaLoggingPolicy", `{"policyName":"Enabled","priority":0}`, 503, nil), Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.AssignToGroup(ctx, "group", "Enabled", new(int32(0)))
		return contracttest.Observation{Response: r, Err: e}
	}})
}

func TestUnitGroupGrantReceiptsDistinguishAcceptedFromNoContent(t *testing.T) {
	t.Parallel()
	path := "/Skype.Policy/groupPolicyAssignments/group/policyTypes/TeamsMediaLoggingPolicy"
	for _, tc := range []struct {
		name     string
		status   int
		reply    []byte
		accepted bool
		op       string
		decode   bool
	}{
		{"accepted", 202, []byte(`{"OperationId":"00000000-0000-4000-8000-000000000099"}`), true, "00000000-0000-4000-8000-000000000099", false},
		{"no-content", 204, nil, false, "", false},
		{"accepted-malformed", 202, []byte(`[`), true, "", true},
		{"accepted-empty", 202, nil, true, "", true},
	} {
		want := contracttest.Expected{}
		if tc.decode {
			want.Error = client.ErrDecode
		}
		runExplicit(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsMediaLoggingPolicy", ParameterSet: "GrantToGroup"}}, Steps: grantSteps(t, "POST", path, `{"policyName":"Enabled","priority":1}`, tc.status, tc.reply), Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			out, r, e := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "group", PolicyName: "Enabled", Rank: new(int32(1))}})
			if out == nil || out.Accepted != tc.accepted || out.OperationID != tc.op {
				t.Fatalf("group receipt differs: %v", out)
			}
			if tc.decode != errors.Is(e, client.ErrDecode) {
				t.Fatal("group receipt decode outcome differs")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}

// Actual tenant receipts: user grant and clear return 200 with no body; the
// owned group grant returned 204 (no accepted operation receipt).
func TestUnitActualGrantReceiptsFromAcceptance(t *testing.T) {
	t.Parallel()
	user := "/Skype.Policy/users/00000000-0000-4000-8000-000000000010/policies/TeamsMediaLoggingPolicy"
	for _, tc := range []struct{ file, name string }{{"assign_user", "Enabled"}, {"clear_user", ""}} {
		body, _ := json.Marshal(map[string]string{"policyName": tc.name})
		runExplicit(t, contracttest.Case{Name: tc.file, Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsMediaLoggingPolicy", ParameterSet: "Identity"}}, Steps: grantSteps(t, "PATCH", user, string(body), 200, readResponseFixture(t, tc.file+".body")), Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			out, r, e := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, resource.Assignment{User: &resource.UserAssignment{Identity: "00000000-0000-4000-8000-000000000010", PolicyName: tc.name}})
			if out != nil {
				t.Fatal("user grant fabricated a group receipt")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
	group := "/Skype.Policy/groupPolicyAssignments/00000000-0000-4000-8000-000000000020/policyTypes/TeamsMediaLoggingPolicy"
	runExplicit(t, contracttest.Case{Name: "assign_group_204", Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsMediaLoggingPolicy", ParameterSet: "GrantToGroup"}}, Steps: grantSteps(t, "POST", group, `{"policyName":"Enabled","priority":1}`, 204, readResponseFixture(t, "assign_group_204.body")), Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		out, r, e := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "00000000-0000-4000-8000-000000000020", PolicyName: "Enabled", Rank: new(int32(1))}})
		if e == nil && (out == nil || out.Accepted || out.RawJSON != nil) {
			t.Fatal("204 group grant receipt differs")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}
