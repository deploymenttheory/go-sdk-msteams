package onlinevoicemailpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

const bindingSource = "contracts/endpoints/8.1.0/fixtures/voicemail-wrapper-binding.json"

func TestVoicemailBindingCompositionMatchesPinnedHelper(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../" + bindingSource)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name    string  `json:"name"`
			Failure *string `json:"failure_type"`
			Calls   []struct {
				Operation  string                     `json:"operation"`
				Parameters map[string]json.RawMessage `json:"parameters"`
			} `json:"calls"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 11 {
		t.Fatal("invalid pinned voicemail binding observations")
	}
	var results []contracttest.Result
	for _, f := range fixture.Cases {
		// The native API always returns typed results/metadata. It exposes no
		// PassThru display switch; preserve the source bridge gap as evidence.
		if strings.Contains(f.Name, "passthru") {
			if f.Failure == nil || len(f.Calls) != 0 {
				t.Fatal("pinned PassThru binding limitation changed")
			}
			continue
		}
		set := "Identity"
		input := resource.Assignment{PolicyName: policy.Set("Fixture")}
		steps := []contracttest.Exchange{}
		want := contracttest.Expected{}
		if f.Failure != nil {
			want.Error = resource.ErrInput
			if f.Name == "grant-missing-policy-name" {
				input.PolicyName = policy.Field[string]{}
				input.Tenant = &resource.TenantAssignment{}
				set = "GrantToTenant"
			}
		} else {
			if len(f.Calls) != 1 {
				t.Fatal("unexpected pinned helper call count")
			}
			call := f.Calls[0]
			var policyType, name string
			if json.Unmarshal(call.Parameters["PolicyType"], &policyType) != nil || policyType != "OnlineVoicemailPolicy" || json.Unmarshal(call.Parameters["PolicyName"], &name) != nil {
				t.Fatal("invalid pinned helper parameters")
			}
			body := map[string]any{"policyName": name}
			path, method, status := "/Skype.Policy/users/fixture%2Fuser/policies/OnlineVoicemailPolicy", "PATCH", 200
			switch call.Operation {
			case "assign-user":
				input.User = &resource.UserAssignment{Identity: "fixture/user"}
				if f.Name == "clear-user" {
					input.PolicyName = policy.Null[string]()
				}
			case "assign-tenant":
				set, path = "GrantToTenant", "/Skype.Policy/tenants/policies/OnlineVoicemailPolicy"
				input.Tenant = &resource.TenantAssignment{}
				body["forceSwitchPresent"] = false
			case "assign-group":
				set, path, method, status = "GrantToGroup", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy", "POST", 204
				input.Group = &resource.GroupAssignment{ID: "fixture/group"}
				if rank, ok := call.Parameters["Rank"]; ok {
					var n int32
					if json.Unmarshal(rank, &n) != nil {
						t.Fatal("invalid pinned rank")
					}
					body["priority"], input.Group.Priority = n, &n
				}
				if f.Name == "grant-group-rank-zero" {
					input.Group.Priority = new(int32)
				}
			}
			raw, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			steps = append(steps, discovery(), generatedWire(method, origin+path, "Grant-CsOnlineVoicemailPolicy", raw, status, ""))
		}
		results = append(results, run(t, contracttest.Case{Name: f.Name, Source: bindingSource, Mappings: []contracttest.Mapping{{Command: "Grant-CsOnlineVoicemailPolicy", ParameterSet: set}}, Steps: steps, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.AssignWithPowerShellBinding(ctx, input)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}}))
	}
	if len(results) != 9 {
		t.Fatal("pinned helper composition coverage incomplete")
	}
	evidence(t, results)
}

func TestVoicemailGeneratedIdentityUsesUnreservedEscaping(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "generated-identity-unreserved", Source: generatedSource, Mappings: []contracttest.Mapping{{Command: "Grant-CsOnlineVoicemailPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(), generatedWire("PATCH", origin+"/Skype.Policy/users/sip%3Afixture%2Buser%40example.invalid/policies/OnlineVoicemailPolicy", "Grant-CsOnlineVoicemailPolicy", json.RawMessage(`{"policyName":"Fixture"}`), 200, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.AssignUser(ctx, "sip:fixture+user@example.invalid", policy.Set("Fixture"))
		if e == nil && v == nil {
			e = errors.New("missing native assignment outcome")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
	evidence(t, []contracttest.Result{result})
}
