package teamsipphonepolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ip_phone_policy"
	"github.com/go-resty/resty/v2"
)

const (
	newCommand    = "New-CsTeamsIPPhonePolicy"
	getCommand    = "Get-CsTeamsIPPhonePolicy"
	setCommand    = "Set-CsTeamsIPPhonePolicy"
	removeCommand = "Remove-CsTeamsIPPhonePolicy"
	grantCommand  = "Grant-CsTeamsIPPhonePolicy"
	item          = collection + "/configuration/fixture-policy"
)

func TestUnitIPPhonePolicyMutationsKeepStatusAndNeverReplay(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"create", "update", "delete", "user", "group", "tenant"} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			method, command, path, body, set := "POST", newCommand, collection, `{"Identity":"fixture-policy","AllowHotDesking":true,"HotDeskingIdleTimeoutInMinutes":45}`, "Identity"
			switch operation {
			case "update":
				method, command, path, body = "PUT", setCommand, item, `{"Identity":"fixture-policy","SearchOnCommonAreaPhoneMode":"Enabled"}`
			case "delete":
				method, command, path, body = "DELETE", removeCommand, item, ""
			case "user":
				method, command, path, body = "PATCH", grantCommand, "/Skype.Policy/users/user%2Ffixture/policies/TeamsIPPhonePolicy", `{"policyName":"fixture-policy"}`
			case "group":
				method, command, path, body, set = "POST", grantCommand, "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsIPPhonePolicy", `{"policyName":"fixture-policy","priority":2}`, "GrantToGroup"
			case "tenant":
				method, command, path, body, set = "PATCH", grantCommand, "/Skype.Policy/tenants/policies/TeamsIPPhonePolicy", `{"policyName":"fixture-policy","forceSwitchPresent":false}`, "GrantToTenant"
			}
			var raw json.RawMessage
			if body != "" {
				raw = json.RawMessage(body)
			}
			run(t, contracttest.Case{Name: fmt.Sprintf("%s-error-%d", operation, status), Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{compiledWire(method, path, command, raw, status, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsIPPhonePolicies
				switch operation {
				case "create":
					r, e := s.CreateByIdentity(ctx, "fixture-policy", resource.Settings{AllowHotDesking: new(true), HotDeskingIdleTimeoutInMinutes: new(int64(45))})
					return contracttest.Observation{Response: r, Err: e}
				case "update":
					r, e := s.UpdateByIdentity(ctx, "fixture-policy", resource.UpdateRequest{SearchOnCommonAreaPhoneMode: new("Enabled")})
					return contracttest.Observation{Response: r, Err: e}
				case "delete":
					r, e := s.DeleteByIdentity(ctx, "fixture-policy")
					return contracttest.Observation{Response: r, Err: e}
				case "user":
					v, r, e := s.Assign(ctx, resource.Assignment{User: &resource.UserAssignment{Identity: "user/fixture", PolicyName: "fixture-policy"}})
					if v != nil {
						t.Fatal("rejected user grant reported a receipt")
					}
					return contracttest.Observation{Response: r, Err: e}
				case "group":
					v, r, e := s.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "group/fixture", PolicyName: "fixture-policy", Rank: new(int32(2))}})
					if v != nil {
						t.Fatal("rejected group grant reported acceptance")
					}
					return contracttest.Observation{Response: r, Err: e}
				default:
					v, r, e := s.Assign(ctx, resource.Assignment{Tenant: &resource.TenantAssignment{PolicyName: "fixture-policy"}})
					if v != nil {
						t.Fatal("rejected tenant grant reported a receipt")
					}
					return contracttest.Observation{Response: r, Err: e}
				}
			}})
		}
	}
}

func TestUnitIPPhonePolicyGrantRequiresOneExplicitScope(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name  string
		input resource.Assignment
	}{{"empty", resource.Assignment{}}, {"user-and-tenant", resource.Assignment{User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}}}, {"all", resource.Assignment{User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}, Group: &resource.GroupGrant{ID: "fixture"}}}} {
		run(t, contracttest.Case{Name: "grant-" + c.name, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "Identity"}}, Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsIPPhonePolicies.Assign(ctx, c.input)
			if v != nil {
				t.Fatal("invalid grant scope returned a receipt")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
	run(t, contracttest.Case{Name: "grant-invalid-principal", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsIPPhonePolicies
		if _, e := s.AssignToUser(ctx, "", "fixture-policy"); !errors.Is(e, resource.ErrInput) {
			t.Fatal("empty user acquired credentials")
		}
		if _, _, e := s.AssignToGroup(ctx, "bad\nidentity", "fixture-policy", nil); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid group acquired credentials")
		}
		if _, e := s.CreateByIdentity(ctx, "..", resource.Settings{}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid create identity acquired credentials")
		}
		if _, e := s.DeleteByIdentity(ctx, " "); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid delete identity acquired credentials")
		}
		if _, _, e := s.Create(ctx, resource.CreateRequest{}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("missing create identity acquired credentials")
		}
		return contracttest.Observation{}
	}})
}

func TestUnitIPPhonePolicyGroupGrantReceipts(t *testing.T) {
	t.Parallel()
	const path = "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsIPPhonePolicy"
	for _, tc := range []struct {
		name, reply string
		status      int
		accepted    bool
		operation   string
		decodeErr   bool
	}{{"accepted", `{"OperationId":"00000000-0000-4000-8000-000000000001"}`, 202, true, "00000000-0000-4000-8000-000000000001", false}, {"accepted-malformed", `[`, 202, true, "", true}, {"no-content", ``, 204, false, "", false}} {
		run(t, contracttest.Case{Name: "group-" + tc.name, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{compiledWire("POST", path, grantCommand, json.RawMessage(`{"policyName":"fixture-policy","priority":0}`), tc.status, tc.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsIPPhonePolicies.AssignToGroup(ctx, "group/fixture", "fixture-policy", new(int32(0)))
			if v == nil || v.Accepted != tc.accepted || v.OperationID != tc.operation || errors.Is(e, client.ErrDecode) != tc.decodeErr {
				t.Fatalf("group receipt %s differs", tc.name)
			}
			return contracttest.Observation{Response: r}
		}})
	}
}

func TestUnitIPPhonePolicyMalformedReadsAndPartialCreate(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, path, reply string }{{"get-array", item, `[]`}, {"get-null", item, `null`}, {"list-object", collection, `{}`}, {"list-null", collection, `null`}} {
		run(t, contracttest.Case{Name: "malformed-" + tc.name, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("GET", tc.path, getCommand, nil, 200, tc.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsIPPhonePolicies
			var e error
			var r *resty.Response
			if tc.path == item {
				v, response, err := s.GetByIdentity(ctx, "fixture-policy")
				if v != nil {
					t.Fatal("malformed read produced a policy")
				}
				r, e = response, err
			} else {
				v, response, err := s.List(ctx)
				if v != nil {
					t.Fatal("malformed list produced policies")
				}
				r, e = response, err
			}
			if !errors.Is(e, client.ErrDecode) || r == nil {
				t.Fatal("malformed response lost ErrDecode or metadata")
			}
			return contracttest.Observation{Response: r}
		}})
	}
	run(t, contracttest.Case{Name: "create-readback-failure", Mappings: []contracttest.Mapping{{Command: newCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("POST", collection, newCommand, json.RawMessage(`{"Identity":"fixture-policy","Description":""}`), 201, ""),
		compiledWire("GET", item, getCommand, nil, 404, `{"code":"ClientError","errorCode":40407}`),
	}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, Codes: []string{"ClientError"}, ServiceErrorCodes: []int{40407}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsIPPhonePolicies.Create(ctx, resource.CreateRequest{Identity: "fixture-policy", Settings: resource.Settings{Description: new("")}})
		var partial *resource.ReadBackError
		if !errors.As(e, &partial) || partial.CreationResponse == nil || partial.CreationResponse.StatusCode() != 201 || v == nil || v.Identity != "fixture-policy" {
			t.Fatal("partial create lost its submission receipt")
		}
		if partial.Error() != "IP phone policy created but read-back failed" || fmt.Sprintf("%#v", *partial) != partial.String() {
			t.Fatal("partial create diagnostic changed")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}
