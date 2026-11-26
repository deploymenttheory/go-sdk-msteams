package externalaccesspolicy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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
		{"user", "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/ExternalAccessPolicy", `{"policyName":null,"additionalParameters":null}`, "GrantToUser"},
		{"identity", "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/ExternalAccessPolicy", `{"policyName":null,"additionalParameters":null}`, "Identity"},
		{"tenant", "PATCH", "/Skype.Policy/tenants/policies/ExternalAccessPolicy", `{"PolicyName":"","additionalParameters":null,"forceSwitchPresent":false}`, "GrantToTenant"},
		{"group", "POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/ExternalAccessPolicy", `{"policyName":null,"priority":0}`, "GrantToGroup"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
				result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "grant-" + tc.name + "-" + string(mode), Family: "external_access_policy", Scope: "service", Mode: mode, Source: "msteams/teams_admin_api/external_access_policy/assignment_protocol_test.go", Mappings: []contracttest.Mapping{{Command: "Grant-CsExternalAccessPolicy", ParameterSet: tc.set}}, Steps: []contracttest.Exchange{wire(tc.method, tc.path, "Grant-CsExternalAccessPolicy", json.RawMessage(tc.body), 204, "")}, Want: contracttest.Expected{Status: 204, Body: []byte{}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					s := sdk.TeamsAdminAPI.ExternalAccessPolicies
					switch tc.name {
					case "user":
						value, response, err := s.AssignUser(ctx, "user/fixture", nil)
						if value != nil {
							t.Fatal("blank user receipt fabricated policy data")
						}
						return contracttest.Observation{Response: response, Err: err}
					case "identity":
						value, response, err := s.Assign(ctx, policy.Assignment{User: &policy.UserAssignment{Identity: "user/fixture"}})
						if value != nil {
							t.Fatal("valid Identity scope union fabricated policy data")
						}
						return contracttest.Observation{Response: response, Err: err}
					case "tenant":
						value, response, err := s.AssignTenantDefault(ctx, new(""), false)
						if value != nil {
							t.Fatal("blank tenant receipt fabricated policy data")
						}
						return contracttest.Observation{Response: response, Err: err}
					default:
						value, response, err := s.AssignGroup(ctx, "group/fixture", nil, 0)
						if value != nil {
							t.Fatal("blank group receipt fabricated a policy or job")
						}
						return contracttest.Observation{Response: response, Err: err}
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
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "grant-opaque-nonpolicy", Family: "external_access_policy", Scope: "service", Source: "msteams/teams_admin_api/external_access_policy/assignment_protocol_test.go", Mappings: []contracttest.Mapping{{Command: "Grant-CsExternalAccessPolicy", ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{wire("POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/ExternalAccessPolicy", "Grant-CsExternalAccessPolicy", json.RawMessage(`{"policyName":null,"priority":0}`), 202, `"opaque-job"`)}, Want: contracttest.Expected{Status: 202, Body: []byte(`"opaque-job"`), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.ExternalAccessPolicies.AssignGroup(ctx, "group/fixture", nil, 0)
		if value != nil {
			t.Fatal("opaque string became a fabricated policy")
		}
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}})
	if err != nil || !result.Passed() {
		t.Fatalf("opaque Grant protocol failure differs: %v %+v", err, result.Differences())
	}
}
