package teamscallholdpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
)

func TestUnitCallHoldAllMutationErrorsAreSingleSend(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "patch", "delete", "user", "tenant", "group"} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			command, set := newCommand, "Identity"
			primarySource := source
			if operation == "user" || operation == "tenant" || operation == "group" {
				primarySource = grantSource
			}
			var steps []contracttest.Exchange
			switch operation {
			case "create":
				steps = []contracttest.Exchange{compiledWire("POST", collection, newCommand, json.RawMessage(`{"Identity":"Global"}`), status, `{"Message":"fixture-private"}`)}
			case "patch":
				command = setCommand
				steps = []contracttest.Exchange{compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","Description":null}`), status, `{"Message":"fixture-private"}`)}
			case "delete":
				command = deleteCommand
				steps = []contracttest.Exchange{compiledWire("DELETE", item, deleteCommand, json.RawMessage(`{"Identity":"Global"}`), status, `{"Message":"fixture-private"}`)}
			case "user":
				command = grantCommand
				steps = []contracttest.Exchange{discovery(t), modernWire("PATCH", "/Skype.Policy/users/fixture%2Fuser/policies/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":""}`), status, `{"Message":"fixture-private"}`)}
			case "tenant":
				command, set = grantCommand, "GrantToTenant"
				steps = []contracttest.Exchange{discovery(t), modernWire("PATCH", "/Skype.Policy/tenants/policies/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":"","forceSwitchPresent":false}`), status, `{"Message":"fixture-private"}`)}
			case "group":
				command, set = grantCommand, "GrantToGroup"
				steps = []contracttest.Exchange{discovery(t), modernWire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":""}`), status, `{"Message":"fixture-private"}`)}
			}
			run(t, contracttest.Case{Name: fmt.Sprintf("%s-error-%d", operation, status), Source: primarySource, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsCallHoldPolicies
				var o contracttest.Observation
				switch operation {
				case "create":
					_, r, e := s.CreateByIdentity(ctx, "Global", resource.CreateRequest{})
					o = contracttest.Observation{Response: r, Err: e}
				case "patch":
					_, r, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
					o = contracttest.Observation{Response: r, Err: e}
				case "delete":
					_, r, e := s.DeleteByIdentity(ctx, "Global", policy.Controls{})
					o = contracttest.Observation{Response: r, Err: e}
				case "user":
					r, e := s.AssignToUser(ctx, "fixture/user", "")
					o = contracttest.Observation{Response: r, Err: e}
				case "tenant":
					r, e := s.AssignTenantDefault(ctx, "", false)
					o = contracttest.Observation{Response: r, Err: e}
				case "group":
					v, r, e := s.AssignToGroup(ctx, "fixture/group", "", nil)
					if v != nil && v.Accepted {
						return contracttest.Observation{Err: errors.New("failed grant reported accepted")}
					}
					o = contracttest.Observation{Response: r, Err: e}
				}
				var api *client.APIError
				if !errors.As(o.Err, &api) || api.StatusCode != status || strings.Contains(o.Err.Error(), "fixture-private") {
					return contracttest.Observation{Err: errors.New("mutation status/redaction changed")}
				}
				o.Err = nil
				return o
			}})
		}
	}
}

func TestUnitCallHoldInvalidInputContextAndBackendHaveNoHTTP(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "local-input-context-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsCallHoldPolicies
		for _, id := range []string{"", " ", ".", "..", "scope/", "/empty", "scope/..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid compiled identity accepted")}
			}
		}
		for _, pattern := range []string{"", "[broken"} {
			if _, _, e := s.ListFiltered(ctx, pattern); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid compiled filter accepted")}
			}
		}
		for _, input := range []resource.PatchRequest{{}, {Controls: policy.Controls{Force: new(true)}}} {
			if _, _, e := s.PatchByIdentity(ctx, "Global", input); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("empty patch acquired credentials")}
			}
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty Set acquired credentials")}
		}
		for _, input := range []resource.Assignment{{}, {User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}}} {
			if _, _, e := s.Assign(ctx, input); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid assignment scope acquired credentials")}
			}
		}
		if _, e := s.AssignToUser(ctx, "", ""); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty user accepted")}
		}
		if _, _, e := s.AssignToGroup(ctx, "bad\nidentity", "", nil); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid group accepted")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		wrong := resource.NewService(graph)
		if _, _, e := wrong.List(ctx); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong CRUD backend accepted")}
		}
		if _, e := wrong.AssignTenantDefault(ctx, "", false); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong Grant backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return contracttest.Observation{Err: errors.New("missing service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, e := s.AssignTenantDefault(canceled, "", false); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("canceled Grant acquired credentials")}
		}
		return contracttest.Observation{}
	}})
}

func TestUnitCallHoldGrantOmitsZeroRankLikeMappedWrapper(t *testing.T) {
	t.Parallel()

	for _, rank := range []int32{0, 2, -1} {
		body := map[string]any{"policyName": "Fixture"}
		name := "rank-zero-omitted"
		if rank != 0 {
			body["priority"] = rank
			name = "rank-positive"
			if rank < 0 {
				name = "rank-negative"
			}
		}
		raw, _ := json.Marshal(body)
		run(t, contracttest.Case{Name: name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{discovery(t), modernWire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsCallHoldPolicy", raw, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsCallHoldPolicies.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "fixture/group", PolicyName: "Fixture", Rank: new(rank)}})
			return contracttest.Observation{Response: r, Err: e}
		}, Want: contracttest.Expected{APIError: &client.
			APIError{Backend: client.TeamsAdmin,
			StatusCode: 400,
		}},
		})
	}
}
