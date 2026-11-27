package teamsapppermissionpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

func TestUnitAppPermissionGrantRequiresOneExplicitScope(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		name  string
		input policy.Assignment
	}{{"empty", policy.Assignment{}}, {"mixed", policy.Assignment{User: &policy.UserAssignment{Identity: "fixture"}, Tenant: &policy.TenantAssignment{}}}} {
		run(t, testutil.Case{Name: "grant-" + c.name, Want: testutil.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.Assign(ctx, c.input)
			return testutil.Observation{Value: v, Response: r, Err: e}
		}})
	}
}

func TestUnitAppPermissionInputAndContextRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-input-context-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies
		for _, id := range []string{"", " ", ".", "..", "scope/", "/empty", "scope/..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid policy identity accepted")}
			}
		}
		for _, pattern := range []string{"", "[broken"} {
			if _, _, e := s.ListFiltered(ctx, pattern); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid policy filter accepted")}
			}
		}
		for _, input := range []resource.PatchRequest{{}, {Controls: policy.Controls{Force: new(true)}}, {Settings: resource.Settings{DefaultCatalogApps: policy.Null[[]*resource.CatalogApp]()}}, {Settings: resource.Settings{GlobalCatalogApps: policy.Null[[]*resource.CatalogApp]()}}, {Settings: resource.Settings{PrivateCatalogApps: policy.Null[[]*resource.CatalogApp]()}}} {
			if _, _, e := s.PatchByIdentity(ctx, "Global", input); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid sparse catalog input accepted")}
			}
		}
		for _, input := range []resource.UpdateRequest{{}, {Controls: policy.Controls{Force: new(false)}}, {DefaultCatalogApps: &resource.CatalogEdit{}}} {
			if _, _, e := s.UpdateByIdentity(ctx, "Global", input); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid policy update performed a read")}
			}
		}
		if _, _, e := s.CreateByIdentity(ctx, "Global", resource.CreateRequest{PrivateCatalogApps: &resource.CatalogEdit{}}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid create edit accepted")}
		}
		if _, _, e := s.AssignUser(ctx, "", nil); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("empty user accepted")}
		}
		if _, _, e := s.AssignGroup(ctx, "bad\nidentity", nil, 0); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid group accepted")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return testutil.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).List(ctx); !errors.Is(e, client.ErrDestination) {
			return testutil.Observation{Err: errors.New("wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return testutil.Observation{Err: errors.New("missing policy service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.AssignTenantDefault(canceled, nil, false); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("cancellation lost")}
		}
		return testutil.Observation{}
	}})
}

func TestUnitAppPermissionMutationsKeepStatusAndNeverReplay(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "patch", "delete", "user", "group", "tenant"} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			method, command, path, body := "POST", "New-CsTeamsAppPermissionPolicy", collection, `{"Identity":"Global"}`
			switch operation {
			case "patch":
				method, command, path, body = "PATCH", "Set-CsTeamsAppPermissionPolicy", item, `{"Identity":"Global","Description":null}`
			case "delete":
				method, command, path = "DELETE", "Remove-CsTeamsAppPermissionPolicy", item
			case "user":
				method, command, path, body = "PATCH", "Grant-CsTeamsAppPermissionPolicy", "/Skype.Policy/users/user%2Ffixture/policies/TeamsAppPermissionPolicy", `{"policyName":null,"additionalParameters":null}`
			case "group":
				method, command, path, body = "POST", "Grant-CsTeamsAppPermissionPolicy", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsAppPermissionPolicy", `{"policyName":null,"priority":0}`
			case "tenant":
				method, command, path, body = "PATCH", "Grant-CsTeamsAppPermissionPolicy", "/Skype.Policy/tenants/policies/TeamsAppPermissionPolicy", `{"PolicyName":null,"additionalParameters":null,"forceSwitchPresent":false}`
			}
			run(t, testutil.Case{Name: fmt.Sprintf("%s-error-%d", operation, status), Steps: []testutil.Exchange{wire(method, path, command, json.RawMessage(body), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies
				var observation testutil.Observation
				switch operation {
				case "create":
					_, r, e := s.CreateByIdentity(ctx, "Global", resource.CreateRequest{})
					observation = testutil.Observation{Response: r, Err: e}
				case "patch":
					_, r, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{Description: policy.Null[string]()}}})
					observation = testutil.Observation{Response: r, Err: e}
				case "delete":
					_, r, e := s.DeleteByIdentity(ctx, "Global", policy.Controls{})
					observation = testutil.Observation{Response: r, Err: e}
				case "user":
					_, r, e := s.AssignUser(ctx, "user/fixture", nil)
					observation = testutil.Observation{Response: r, Err: e}
				case "group":
					_, r, e := s.AssignGroup(ctx, "group/fixture", nil, 0)
					observation = testutil.Observation{Response: r, Err: e}
				case "tenant":
					_, r, e := s.AssignTenantDefault(ctx, nil, false)
					observation = testutil.Observation{Response: r, Err: e}
				}
				var api *client.APIError
				if !errors.As(observation.Err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(observation.Err), "fixture-private") {
					return testutil.Observation{Err: errors.New("policy mutation status/redaction changed")}
				}
				observation.Err = nil
				return observation
			}})
		}
	}
}
