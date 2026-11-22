package teamsroomvideoteleconferencingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_room_video_tele_conferencing_policy"
)

const collection = "/Skype.Policy/configurations/TeamsRoomVideoTeleConferencingPolicy"

func TestUnitSparseModernSetKeepsFalseAndEmptyBindings(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "sparse-put", Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsRoomVideoTeleConferencingPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("PUT", collection+"/configuration/Global", "Set-CsTeamsRoomVideoTeleConferencingPolicy", json.RawMessage(`{"Identity":"Global","Enabled":false,"AreaCode":""}`), 503, ""),
	}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, e := sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies.UpdateGlobal(ctx, resource.UpdateRequest{Enabled: new(false), AreaCode: new("")})
		return contracttest.Observation{Response: r, Err: e}
	}})
}

func TestUnitModernGrantScopeBodiesAndRankOmission(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, method, path, body, set string
		args                          resource.Assignment
	}{{"user-clear", "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/TeamsRoomVideoTeleConferencingPolicy", `{"policyName":""}`, "Identity", resource.Assignment{User: &resource.UserAssignment{Identity: "user/fixture"}}}, {"tenant", "PATCH", "/Skype.Policy/tenants/policies/TeamsRoomVideoTeleConferencingPolicy", `{"policyName":"","forceSwitchPresent":false}`, "GrantToTenant", resource.Assignment{Tenant: &resource.TenantAssignment{}}}, {"group-zero", "POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsRoomVideoTeleConferencingPolicy", `{"policyName":""}`, "GrantToGroup", resource.Assignment{Group: &resource.GroupGrant{ID: "group/fixture", Rank: new(int32(0))}}}, {"group-negative", "POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsRoomVideoTeleConferencingPolicy", `{"policyName":"","priority":-1}`, "GrantToGroup", resource.Assignment{Group: &resource.GroupGrant{ID: "group/fixture", Rank: new(int32(-1))}}}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsRoomVideoTeleConferencingPolicy", ParameterSet: tc.set}}, Steps: []contracttest.Exchange{compiledWire(tc.method, tc.path, "Grant-CsTeamsRoomVideoTeleConferencingPolicy", json.RawMessage(tc.body), 403, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies.Assign(ctx, tc.args)
				if v != nil {
					t.Fatal("rejected grant reported acceptance")
				}
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitInvalidRoomInputsAndContextStayLocal(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid identity acquired credentials")
			}
		}
		if _, _, e := s.Assign(ctx, resource.Assignment{}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("missing grant scope acquired credentials")
		}
		if _, _, e := s.ListFiltered(ctx, "[bad"); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid filter acquired credentials")
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, e := s.UpdateGlobal(canceled, resource.UpdateRequest{}); !errors.Is(e, context.Canceled) {
			t.Fatal("canceled write reached HTTP")
		}
		for _, empty := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := empty.List(ctx); !errors.Is(e, client.ErrClosed) {
				t.Fatal("uninitialized service accepted")
			}
		}
		return contracttest.Observation{}
	}})
}
