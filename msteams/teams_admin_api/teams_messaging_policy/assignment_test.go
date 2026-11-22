package teamsmessagingpolicy_test

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
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_policy"
)

// Injected empty failures exercise source-defined requests and replay safety;
// successful resource bodies come exclusively from acceptance testdata.
func TestUnitMutationsKeepFailuresAndNeverReplay(t *testing.T) {
	t.Parallel()
	discovery, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, operation := range []string{"create", "patch", "delete", "user", "tenant", "group", "group-zero", "group-negative"} {
			for _, status := range []int{401, 403, 429, 503} {
				command, method, path, body, set := "New-CsTeamsMessagingPolicy", "POST", "/Skype.Policy/configurations/TeamsMessagingPolicy", `{"Identity":"fixture-policy"}`, "Identity"
				switch operation {
				case "create":
				case "patch":
					command, method, path, body = "Set-CsTeamsMessagingPolicy", "PATCH", path+"/configuration/fixture-policy", `{"Identity":"fixture-policy","Description":null}`
				case "delete":
					command, method, path = "Remove-CsTeamsMessagingPolicy", "DELETE", path+"/configuration/fixture-policy"
				case "user":
					command, method, path, body = "Grant-CsTeamsMessagingPolicy", "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/TeamsMessagingPolicy", `{"policyName":""}`
				case "tenant":
					command, method, path, body, set = "Grant-CsTeamsMessagingPolicy", "PATCH", "/Skype.Policy/tenants/policies/TeamsMessagingPolicy", `{"policyName":"fixture-policy","forceSwitchPresent":false}`, "GrantToTenant"
				default:
					command, method, path, body, set = "Grant-CsTeamsMessagingPolicy", "POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMessagingPolicy", `{"policyName":"fixture-policy"}`, "GrantToGroup"
					if operation == "group-negative" {
						body = `{"policyName":"fixture-policy","priority":-1}`
					}
				}
				step := compiledWire(method, path, command, json.RawMessage(body), status, "")
				steps := []contracttest.Exchange{}
				if command == "Grant-CsTeamsMessagingPolicy" {
					steps = append(steps, contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: discovery}})
					step.Request.Headers = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}, http.CanonicalHeaderKey("MPACmdlet"): nil}
				}
				steps = append(steps, step)
				run(t, contracttest.Case{Name: fmt.Sprintf("%s-%s-%d", mode, operation, status), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: steps, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					service := sdk.TeamsAdminAPI.TeamsMessagingPolicies
					switch operation {
					case "create":
						_, r, e := service.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{})
						return contracttest.Observation{Response: r, Err: e}
					case "patch":
						_, r, e := service.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
						return contracttest.Observation{Response: r, Err: e}
					case "delete":
						_, r, e := service.DeleteByIdentity(ctx, "fixture-policy", policy.Controls{})
						return contracttest.Observation{Response: r, Err: e}
					case "user":
						r, e := service.AssignToUser(ctx, "user/fixture", "")
						return contracttest.Observation{Response: r, Err: e}
					case "tenant":
						_, r, e := service.Assign(ctx, resource.Assignment{Tenant: &resource.TenantAssignment{PolicyName: "fixture-policy"}})
						return contracttest.Observation{Response: r, Err: e}
					default:
						var rank *int32
						if operation == "group-zero" {
							rank = new(int32(0))
						}
						if operation == "group-negative" {
							rank = new(int32(-1))
						}
						_, r, e := service.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "group/fixture", PolicyName: "fixture-policy", Rank: rank}})
						return contracttest.Observation{Response: r, Err: e}
					}
				}})
			}
		}
	}
}
func TestUnitGrantRequiresExactlyOneScope(t *testing.T) {
	t.Parallel()
	for _, input := range []resource.Assignment{{}, {User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}}} {
		run(t, contracttest.Case{Name: "scope-validation", Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsMessagingPolicies.Assign(ctx, input)
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}

func TestUnitGroupAcceptanceKeepsMalformedReceiptStatus(t *testing.T) {
	t.Parallel()
	discovery, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	// Fault injection checks that an accepted POST remains distinguishable from
	// receipt decoding failure. These bodies are deliberately invalid, not replies
	// used as successful tenant fixtures.
	for index, body := range []string{"null", "[]", `{"OperationId":123}`, "{"} {
		steps := []contracttest.Exchange{
			{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: discovery}},
			{Request: contracttest.WireRequest{Method: "POST", URL: origin + "/Skype.Policy/groupPolicyAssignments/group/policyTypes/TeamsMessagingPolicy", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {"Grant-CsTeamsMessagingPolicy"}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}, http.CanonicalHeaderKey("MPACmdlet"): nil}, JSON: json.RawMessage(`{"policyName":"fixture-policy","priority":0}`)}, Reply: contracttest.Reply{Status: 202, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}},
		}
		run(t, contracttest.Case{Name: fmt.Sprintf("bad-receipt-%d", index), Steps: steps, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsMessagingPolicies.AssignToGroup(ctx, "group", "fixture-policy", new(int32(0)))
			if value == nil || !value.Accepted || response == nil || response.StatusCode() != 202 || !errors.Is(err, client.ErrDecode) {
				t.Fatal("accepted POST status was lost on invalid group receipt")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	}
}
