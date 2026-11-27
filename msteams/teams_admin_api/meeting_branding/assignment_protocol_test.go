package meetingbranding_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// Primary-authored blank successful Grant receipt, not a tenant capture or
// effective-assignment proof. Pinned PolicyRp.dll a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51:
// HttpRequestHelper<T>.GrantToUser/Tenant/Group call SendAsync<T> at IL255/228/279;
// generic SendAsync d__7<T> checks IsSuccessStatusCode IL212/217 and skips
// deserialization for empty body at IL425. BaseTpmGrantCmdlet emits no
// WriteObject of PolicyResult. No group job or effective policy is inferred.
func TestUnitGrantAuthoredEmptyAcceptanceAndValidScopeUnion(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, method, path, body, set string }{
		{"user", "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/TeamsMeetingBrandingPolicy", `{"policyName":null,"additionalParameters":null}`, "GrantToUser"},
		{"explicit-user-union", "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/TeamsMeetingBrandingPolicy", `{"policyName":null,"additionalParameters":null}`, "GrantToUser"},
		{"tenant", "PATCH", "/Skype.Policy/tenants/policies/TeamsMeetingBrandingPolicy", `{"PolicyName":"","additionalParameters":null,"forceSwitchPresent":false}`, "GrantToTenant"},
		{"group", "POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMeetingBrandingPolicy", `{"policyName":null,"priority":0}`, "GrantToGroup"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
				result, err := testutil.Run(t.Context(), testutil.Case{Name: "grant-" + tc.name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{wire(tc.method, tc.path, "Grant-CsTeamsMeetingBrandingPolicy", json.RawMessage(tc.body), 204, "")}, Want: testutil.Expected{Status: 204, Body: []byte{}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					s := sdk.TeamsAdminAPI.MeetingBrandingPolicies
					switch tc.name {
					case "user":
						value, response, err := s.AssignUser(ctx, "user/fixture", nil)
						if value != nil {
							t.Fatal("blank user receipt fabricated policy data")
						}
						return testutil.Observation{Response: response, Err: err}
					case "explicit-user-union":
						value, response, err := s.Assign(ctx, policy.Assignment{User: &policy.UserAssignment{Identity: "user/fixture"}})
						if value != nil {
							t.Fatal("valid native user scope union fabricated policy data")
						}
						return testutil.Observation{Response: response, Err: err}
					case "tenant":
						value, response, err := s.AssignTenantDefault(ctx, new(""), false)
						if value != nil {
							t.Fatal("blank tenant receipt fabricated policy data")
						}
						return testutil.Observation{Response: response, Err: err}
					default:
						value, response, err := s.AssignGroup(ctx, "group/fixture", nil, 0)
						if value != nil {
							t.Fatal("blank group receipt fabricated a policy or job")
						}
						return testutil.Observation{Response: response, Err: err}
					}
				}})
				if err != nil || !result.Passed() {
					t.Fatalf("authored Grant receipt differs: %v %+v", err, result.Differences())
				}
			}
		})
	}
}

// An adversarial non-policy string after a successful HTTP status must reach
// the pinned typed decoder. This is not a captured service job/handle response.
func TestUnitGrantRejectsOpaqueNonPolicySuccessBody(t *testing.T) {
	t.Parallel()
	result, err := testutil.Run(t.Context(), testutil.Case{Name: "grant-opaque-nonpolicy", Steps: []testutil.Exchange{wire("POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMeetingBrandingPolicy", "Grant-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"policyName":null,"priority":0}`), 202, `"opaque-job"`)}, Want: testutil.Expected{Status: 202, Body: []byte(`"opaque-job"`), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.AssignGroup(ctx, "group/fixture", nil, 0)
		if value != nil {
			t.Fatal("opaque string became a fabricated policy")
		}
		return testutil.Observation{Value: value, Response: response, Err: err}
	}})
	if err != nil || !result.Passed() {
		t.Fatalf("opaque Grant protocol failure differs: %v %+v", err, result.Differences())
	}
}

// Pinned default Identity is a parameterless fallback, not a fourth assignment
// operation. ExecuteGrant recognizes only GrantToUser/Group/Tenant and reports
// Cannot determine which grant to execute otherwise. Native Go rejects that
// absent scope before credentials; PowerShell error presentation stays separate.
func TestUnitGrantDefaultIdentitySetRejectsAbsentScope(t *testing.T) {
	t.Parallel()
	result, err := testutil.Run(t.Context(), testutil.Case{Name: "grant-default-no-scope", Want: testutil.Expected{Error: policy.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.Assign(ctx, policy.Assignment{})
		if value != nil || response != nil {
			t.Fatal("absent assignment scope fabricated a policy or resource request")
		}
		return testutil.Observation{Response: response, Err: err}
	}})
	if err != nil || !result.Passed() {
		t.Fatalf("default Grant scope rejection differs: %v %+v", err, result.Differences())
	}
}
