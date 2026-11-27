package teamsupdatemanagementpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_update_management_policy"
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
			{"user", "/Skype.Policy/users/user%2Ffixture/policies/TeamsUpdateManagementPolicy", `{"policyName":""}`, "Identity", nil, 0},
			{"tenant", "/Skype.Policy/tenants/policies/TeamsUpdateManagementPolicy", `{"policyName":"","forceSwitchPresent":false}`, "GrantToTenant", nil, 1},
			{"group-unbound", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsUpdateManagementPolicy", `{"policyName":""}`, "GrantToGroup", nil, 2},
			{"wrapper-zero", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsUpdateManagementPolicy", `{"policyName":""}`, "GrantToGroup", new(int32(0)), 2},
			{"signed-rank", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsUpdateManagementPolicy", `{"policyName":"","priority":-2}`, "GrantToGroup", new(int32(-2)), 2},
		} {
			method := "PATCH"
			if tc.scope == 2 {
				method = "POST"
			}
			steps := []testutil.Exchange{
				{Request: testutil.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: discovery}},
				{Request: testutil.WireRequest{Method: method, URL: origin + tc.path, Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {"Grant-CsTeamsUpdateManagementPolicy"}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}, "MPACmdlet": nil, "Content-Type": {"application/json"}}, JSON: json.RawMessage(tc.body)}, Reply: testutil.Reply{Status: 503}},
			}
			run(t, testutil.Case{Name: fmt.Sprintf("%s-%s", mode, tc.name), Mode: mode, Steps: steps, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				input := resource.Assignment{}
				switch tc.scope {
				case 0:
					input.User = &resource.UserAssignment{Identity: "user/fixture"}
				case 1:
					input.Tenant = &resource.TenantAssignment{}
				case 2:
					input.Group = &resource.GroupGrant{ID: "group/fixture", Rank: tc.rank}
				}
				_, r, e := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies.Assign(ctx, input)
				return testutil.Observation{Response: r, Err: e}
			}})
		}
	}
}
func TestUnitWeekdayAndMissingSettingsFailBeforeHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-source-validation", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsUpdateManagementPolicies
		for _, day := range []int64{-1, 7, 9223372036854775807} {
			settings := resource.ScalarSettings{UpdateDayOfWeek: policy.Set(day)}
			if _, _, e := s.CreateByIdentity(ctx, "fixture", resource.CreateRequest{ScalarSettings: settings}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("source weekday range reached credentials")
			}
			if _, _, e := s.PatchByIdentity(ctx, "fixture", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: settings}}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid weekday patch reached credentials")
			}
			if _, _, e := s.UpdateByIdentity(ctx, "fixture", resource.UpdateRequest{ScalarSettings: settings}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid weekday preparation reached credentials")
			}
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Controls: policy.Controls{Force: new(true)}}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("Force alone reached credentials")
		}
		for _, input := range []resource.Assignment{{}, {User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}}} {
			if _, _, e := s.Assign(ctx, input); !errors.Is(e, resource.ErrInput) {
				t.Fatal("ambiguous grant scope reached credentials")
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.Assign(canceled, resource.Assignment{Tenant: &resource.TenantAssignment{}}); !errors.Is(e, context.Canceled) {
			t.Fatal("grant cancellation lost")
		}
		return testutil.Observation{}
	}})
}
