package teamsmeetingbroadcastpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_policy"
	"github.com/go-resty/resty/v2"
)

const (
	newCommand    = "New-CsTeamsMeetingBroadcastPolicy"
	getCommand    = "Get-CsTeamsMeetingBroadcastPolicy"
	setCommand    = "Set-CsTeamsMeetingBroadcastPolicy"
	removeCommand = "Remove-CsTeamsMeetingBroadcastPolicy"
	grantCommand  = "Grant-CsTeamsMeetingBroadcastPolicy"
	item          = collection + "/configuration/fixture-policy"
)

func TestUnitMeetingBroadcastPolicyMutationsKeepStatusAndNeverReplay(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"create", "update", "delete", "user", "group", "tenant"} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			method, command, path, body, set := "POST", newCommand, collection, `{"Identity":"fixture-policy","AllowBroadcastTranscription":true}`, "Identity"
			switch operation {
			case "update":
				method, command, path, body = "PUT", setCommand, item, `{"Identity":"fixture-policy","BroadcastAttendeeVisibilityMode":"EveryoneInCompany"}`
			case "delete":
				method, command, path, body = "DELETE", removeCommand, item, ""
			case "user":
				method, command, path, body = "PATCH", grantCommand, "/Skype.Policy/users/user%2Ffixture/policies/TeamsMeetingBroadcastPolicy", `{"policyName":"fixture-policy"}`
			case "group":
				method, command, path, body, set = "POST", grantCommand, "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMeetingBroadcastPolicy", `{"policyName":"fixture-policy","priority":2}`, "GrantToGroup"
			case "tenant":
				method, command, path, body, set = "PATCH", grantCommand, "/Skype.Policy/tenants/policies/TeamsMeetingBroadcastPolicy", `{"policyName":"fixture-policy","forceSwitchPresent":false}`, "GrantToTenant"
			}
			var raw json.RawMessage
			if body != "" {
				raw = json.RawMessage(body)
			}
			run(t, contracttest.Case{Name: fmt.Sprintf("%s-error-%d", operation, status), Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{compiledWire(method, path, command, raw, status, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies
				switch operation {
				case "create":
					r, e := s.CreateByIdentity(ctx, "fixture-policy", resource.Settings{AllowBroadcastTranscription: new(true)})
					return contracttest.Observation{Response: r, Err: e}
				case "update":
					r, e := s.UpdateByIdentity(ctx, "fixture-policy", resource.UpdateRequest{BroadcastAttendeeVisibilityMode: new("EveryoneInCompany")})
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

func TestUnitMeetingBroadcastPolicyGrantRequiresOneExplicitScope(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name  string
		input resource.Assignment
	}{{"empty", resource.Assignment{}}, {"user-and-tenant", resource.Assignment{User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}}}, {"all", resource.Assignment{User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}, Group: &resource.GroupGrant{ID: "fixture"}}}} {
		run(t, contracttest.Case{Name: "grant-" + c.name, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "Identity"}}, Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies.Assign(ctx, c.input)
			if v != nil {
				t.Fatal("invalid grant scope returned a receipt")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
	run(t, contracttest.Case{Name: "grant-invalid-principal", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies
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

func TestUnitMeetingBroadcastPolicyGroupGrantReceipts(t *testing.T) {
	t.Parallel()
	const path = "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsMeetingBroadcastPolicy"
	for _, tc := range []struct {
		name, reply string
		status      int
		accepted    bool
		operation   string
		decodeErr   bool
	}{{"accepted", `{"OperationId":"00000000-0000-4000-8000-000000000001"}`, 202, true, "00000000-0000-4000-8000-000000000001", false}, {"accepted-malformed", `[`, 202, true, "", true}, {"no-content", ``, 204, false, "", false}} {
		run(t, contracttest.Case{Name: "group-" + tc.name, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{compiledWire("POST", path, grantCommand, json.RawMessage(`{"policyName":"fixture-policy","priority":0}`), tc.status, tc.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies.AssignToGroup(ctx, "group/fixture", "fixture-policy", new(int32(0)))
			if v == nil || v.Accepted != tc.accepted || v.OperationID != tc.operation || errors.Is(e, client.ErrDecode) != tc.decodeErr {
				t.Fatalf("group receipt %s differs", tc.name)
			}
			return contracttest.Observation{Response: r}
		}})
	}
}

func TestUnitMeetingBroadcastPolicyMalformedReadsAndPartialCreate(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, path, reply string }{{"get-array", item, `[]`}, {"get-null", item, `null`}, {"list-object", collection, `{}`}, {"list-null", collection, `null`}} {
		run(t, contracttest.Case{Name: "malformed-" + tc.name, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("GET", tc.path, getCommand, nil, 200, tc.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies
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
		v, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies.Create(ctx, resource.CreateRequest{Identity: "fixture-policy", Settings: resource.Settings{Description: new("")}})
		var partial *resource.ReadBackError
		if !errors.As(e, &partial) || partial.CreationResponse == nil || partial.CreationResponse.StatusCode() != 201 || v == nil || v.Identity != "fixture-policy" {
			t.Fatal("partial create lost its submission receipt")
		}
		if partial.Error() != "meeting broadcast policy created but read-back failed" || fmt.Sprintf("%#v", *partial) != partial.String() {
			t.Fatal("partial create diagnostic changed")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}
