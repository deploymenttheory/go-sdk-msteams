package teamstemplatepermissionpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_template_permission_policy"
)

func TestUnitInputAndCancellationMakeNoRequests(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid policy identity reached transport")
			}
		}
		if _, _, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Controls: policy.Controls{Force: new(true)}}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("Force alone reached transport")
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("empty update reached transport")
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.CreateByIdentity(canceled, "fixture", resource.CreateRequest{}); !errors.Is(e, context.Canceled) {
			t.Fatal("mutation cancellation lost")
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				t.Fatal("uninitialized service accepted")
			}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			t.Fatal(e)
		}
		if _, _, e := resource.NewService(graph).List(ctx); !errors.Is(e, client.ErrDestination) {
			t.Fatal("wrong destination acquired credentials")
		}
		return testutil.Observation{}
	}})
}

// Controlled error statuses are fault injection; successful bodies are actual acceptance fixtures.
func TestUnitMutationsRetainStatusAndNeverReplay(t *testing.T) {
	t.Parallel()
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, operation := range []string{"create", "patch", "delete"} {
			for _, status := range []int{400, 401, 403, 429, 503} {
				method, command, path, body := "POST", "New-CsTeamsTemplatePermissionPolicy", collection, `{"Identity":"Global"}`
				if operation == "patch" {
					method, command, path, body = "PATCH", "Set-CsTeamsTemplatePermissionPolicy", collection+"/configuration/Global", `{"Identity":"Global","Description":null}`
				}
				if operation == "delete" {
					method, command, path = "DELETE", "Remove-CsTeamsTemplatePermissionPolicy", collection+"/configuration/Global"
				}
				run(t, testutil.Case{Name: fmt.Sprintf("%s-%s-%d", mode, operation, status), Mode: mode, Steps: []testutil.Exchange{compiledWire(method, path, command, json.RawMessage(body), status, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					s := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies
					if operation == "create" {
						_, r, e := s.CreateByIdentity(ctx, "Global", resource.CreateRequest{})
						return testutil.Observation{Response: r, Err: e}
					}
					if operation == "patch" {
						_, r, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{Description: policy.Null[string]()}}})
						return testutil.Observation{Response: r, Err: e}
					}
					_, r, e := s.DeleteByIdentity(ctx, "Global", policy.Controls{})
					return testutil.Observation{Response: r, Err: e}
				}})
			}
		}
	}
}
