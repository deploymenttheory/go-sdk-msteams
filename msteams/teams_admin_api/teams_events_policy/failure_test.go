package teamseventspolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_events_policy"
)

func TestEventsPolicyGrantRequiresOneExplicitScope(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name  string
		input policy.Assignment
	}{{"empty", policy.Assignment{}}, {"mixed", policy.Assignment{User: &policy.UserAssignment{Identity: "fixture"}, Tenant: &policy.TenantAssignment{}}}} {
		results = append(results, run(t, contracttest.Case{Name: "grant-" + c.name, Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsEventsPolicy", ParameterSet: "Identity"}}, Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsEventsPolicies.Assign(ctx, c.input)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestEventsPolicyInputAndContextRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "local-input-context-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsEventsPolicies
		for _, id := range []string{"", " ", ".", "..", "scope/", "/empty", "scope/..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid policy identity accepted")}
			}
		}
		for _, pattern := range []string{"", "[broken"} {
			if _, _, e := s.ListFiltered(ctx, pattern); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid policy filter accepted")}
			}
		}
		for _, input := range []resource.PatchRequest{{}, {Controls: policy.Controls{Force: new(true)}}} {
			if _, _, err := s.PatchByIdentity(ctx, "Global", input); !errors.Is(err, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("empty patch acquired credentials")}
			}
		}
		if _, _, err := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{}); !errors.Is(err, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty Set acquired credentials")}
		}
		if _, _, e := s.AssignUser(ctx, "", nil); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty user accepted")}
		}
		if _, _, e := s.AssignGroup(ctx, "bad\nidentity", nil, 0); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid group accepted")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).List(ctx); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return contracttest.Observation{Err: errors.New("missing policy service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.AssignTenantDefault(canceled, nil, false); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{result})
}

func TestEventsPolicyMutationsKeepStatusAndNeverReplay(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, operation := range []string{"create", "patch", "delete", "user", "group", "tenant"} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			method, command, path, body, set := "POST", "New-CsTeamsEventsPolicy", collection, `{"Identity":"Global"}`, "Identity"
			switch operation {
			case "patch":
				method, command, path, body = "PATCH", "Set-CsTeamsEventsPolicy", item, `{"Identity":"Global","Description":null}`
			case "delete":
				method, command, path = "DELETE", "Remove-CsTeamsEventsPolicy", item
			case "user":
				method, command, path, body, set = "PATCH", "Grant-CsTeamsEventsPolicy", "/Skype.Policy/users/user%2Ffixture/policies/TeamsEventsPolicy", `{"policyName":null,"additionalParameters":null}`, "GrantToUser"
			case "group":
				method, command, path, body, set = "POST", "Grant-CsTeamsEventsPolicy", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsEventsPolicy", `{"policyName":null,"priority":0}`, "GrantToGroup"
			case "tenant":
				method, command, path, body, set = "PATCH", "Grant-CsTeamsEventsPolicy", "/Skype.Policy/tenants/policies/TeamsEventsPolicy", `{"PolicyName":null,"additionalParameters":null,"forceSwitchPresent":false}`, "GrantToTenant"
			}
			results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("%s-error-%d", operation, status), Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{compiledWire(method, path, command, json.RawMessage(body), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsEventsPolicies
				var observation contracttest.Observation
				switch operation {
				case "create":
					_, r, e := s.CreateByIdentity(ctx, "Global", resource.CreateRequest{})
					observation = contracttest.Observation{Response: r, Err: e}
				case "patch":
					_, r, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
					observation = contracttest.Observation{Response: r, Err: e}
				case "delete":
					_, r, e := s.DeleteByIdentity(ctx, "Global", policy.Controls{})
					observation = contracttest.Observation{Response: r, Err: e}
				case "user":
					_, r, e := s.AssignUser(ctx, "user/fixture", nil)
					observation = contracttest.Observation{Response: r, Err: e}
				case "group":
					_, r, e := s.AssignGroup(ctx, "group/fixture", nil, 0)
					observation = contracttest.Observation{Response: r, Err: e}
				case "tenant":
					_, r, e := s.AssignTenantDefault(ctx, nil, false)
					observation = contracttest.Observation{Response: r, Err: e}
				}
				var api *client.APIError
				if !errors.As(observation.Err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(observation.Err), "fixture-private") {
					return contracttest.Observation{Err: errors.New("policy mutation status/redaction changed")}
				}
				observation.Err = nil
				return observation
			}}))
		}
	}
	evidence(t, results)
}
