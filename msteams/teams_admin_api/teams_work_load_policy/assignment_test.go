package teamsworkloadpolicy_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_load_policy"
)

func TestUnitModernGrantsUseSeparateDiscoveryAndWrapperRankRules(t *testing.T) {
	t.Parallel()
	discovery, e := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, tc := range []struct {
			name, path, body, set string
			rank                  *int32
			scope                 int
		}{
			{"user", "/Skype.Policy/users/user%2Ffixture/policies/TeamsWorkLoadPolicy", `{"policyName":""}`, "Identity", nil, 0},
			{"tenant", "/Skype.Policy/tenants/policies/TeamsWorkLoadPolicy", `{"policyName":"","forceSwitchPresent":false}`, "GrantToTenant", nil, 1},
			{"group-unbound", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsWorkLoadPolicy", `{"policyName":""}`, "GrantToGroup", nil, 2},
			{"wrapper-zero", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsWorkLoadPolicy", `{"policyName":""}`, "GrantToGroup", new(int32(0)), 2},
			{"signed-rank", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsWorkLoadPolicy", `{"policyName":"","priority":-2}`, "GrantToGroup", new(int32(-2)), 2},
		} {
			method := "PATCH"
			if tc.scope == 2 {
				method = "POST"
			}
			steps := []contracttest.Exchange{
				{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: discovery}},
				{Request: contracttest.WireRequest{Method: method, URL: origin + tc.path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {"Grant-CsTeamsWorkLoadPolicy"}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}, "MPACmdlet": nil, "Content-Type": {"application/json"}}, JSON: json.RawMessage(tc.body)}, Reply: contracttest.Reply{Status: 503}},
			}
			runExplicit(t, contracttest.Case{Name: fmt.Sprintf("%s-%s", mode, tc.name), Mode: mode, Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsWorkLoadPolicy", ParameterSet: tc.set}}, Steps: steps, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				input := resource.Assignment{}
				switch tc.scope {
				case 0:
					input.User = &resource.UserAssignment{Identity: "user/fixture"}
				case 1:
					input.Tenant = &resource.TenantAssignment{}
				case 2:
					input.Group = &resource.GroupGrant{ID: "group/fixture", Rank: tc.rank}
				}
				_, r, e := sdk.TeamsAdminAPI.TeamsWorkLoadPolicies.Assign(ctx, input)
				return contracttest.Observation{Response: r, Err: e}
			}})
		}
	}
}
