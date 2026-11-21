package onlinevoicemailpolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitVoicemailInvalidInputAndContextFailBeforeTokens(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "invalid-input-and-scope", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity", "scope/..", "/empty-scope", "scope/"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid voicemail identity accepted")}
			}
		}
		if _, _, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty patch accepted")}
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty update accepted")}
		}
		if _, _, e := s.Create(ctx, resource.CreateRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty create identity accepted")}
		}
		if _, e := s.DeleteByIdentity(ctx, "", policy.Controls{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty delete identity accepted")}
		}
		for _, input := range []resource.Assignment{{}, {User: &resource.UserAssignment{Identity: "Fixture"}}, {PolicyName: policy.Set("Fixture")}, {PolicyName: policy.Null[string](), User: &resource.UserAssignment{}, Tenant: &resource.TenantAssignment{}}, {PolicyName: policy.Set("Fixture"), Group: &resource.GroupAssignment{}}} {
			if _, _, e := s.Assign(ctx, input); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid assignment acquired credentials")}
			}
		}
		if _, _, e := s.GetAssignmentOperation(ctx, ""); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty operation ID accepted")}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid wildcard acquired credentials")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).AssignTenantDefault(ctx, policy.Set("Fixture"), false); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong assignment backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return contracttest.Observation{Err: errors.New("missing service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.AssignTenantDefault(canceled, policy.Set("Fixture"), false); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("assignment cancellation lost")}
		}
		return contracttest.Observation{}
	}})
}

func TestUnitVoicemailWriteDenialsKeepStatusAndNeverReplay(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "delete", "group"} {
		for _, status := range []int{401, 403, 429, 503} {
			command, method, path := "New-CsOnlineVoicemailPolicy", "POST", collection
			steps := []contracttest.Exchange{}
			body := json.RawMessage(`{"Identity":"Global"}`)
			if operation == "delete" {
				command, method, path = "Remove-CsOnlineVoicemailPolicy", "DELETE", item
			}
			if operation == "group" {
				command, path = "Grant-CsOnlineVoicemailPolicy", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy"
				body = json.RawMessage(`{"policyName":"Fixture"}`)
				steps = append(steps, discovery(t), generatedWire(method, origin+path, command, body, status, `{"message":"fixture-private"}`))
			} else {
				steps = append(steps, wire(method, path, command, body, status, `{"Message":"fixture-private"}`))
			}
			run(t, contracttest.Case{Name: fmt.Sprintf("%s-error-%d", operation, status), Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
				var observation contracttest.Observation
				switch operation {
				case "create":
					_, observation.Response, observation.Err = s.Create(ctx, resource.CreateRequest{Identity: "Global"})
				case "delete":
					observation.Response, observation.Err = s.DeleteByIdentity(ctx, "Global", policy.Controls{})
				default:
					_, observation.Response, observation.Err = s.AssignGroup(ctx, "fixture/group", policy.Set("Fixture"), nil)
				}
				var api *client.APIError
				if !errors.As(observation.Err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(observation.Err), "fixture-private") {
					return contracttest.Observation{Err: errors.New("normalized error/status lost")}
				}
				return contracttest.Observation{Response: observation.Response}
			}})
		}
	}
}
