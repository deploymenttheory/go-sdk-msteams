package meetingbranding_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitBrandingInvalidInputRejectsBeforeCredentials(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "invalid-input-and-scope", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.MeetingBrandingPolicies
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity", "scope/..", "/empty-scope", "scope/"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid branding identity accepted")}
			}
		}
		if _, _, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("empty patch accepted")}
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("empty update accepted")}
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{MeetingBackgroundImages: &resource.AssetEdit[resource.BackgroundImage]{}}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid edit accepted")}
		}
		bad := resource.Settings{MeetingBrandingThemes: &resource.AssetList[resource.BrandingTheme]{}}
		if _, _, e := s.CreateWithGlobalComparison(ctx, resource.CreateRequest{Identity: "Fixture", Settings: bad}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid list acquired credentials for comparison")}
		}
		if _, _, e := s.Create(ctx, resource.CreateRequest{}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("empty create identity accepted")}
		}
		if _, e := s.DeleteByIdentity(ctx, "", policy.Controls{}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("empty delete identity accepted")}
		}
		for _, input := range []policy.Assignment{{}, {User: &policy.UserAssignment{}, Tenant: &policy.TenantAssignment{}}, {User: &policy.UserAssignment{}}, {Group: &policy.GroupAssignment{}}} {
			if _, _, e := s.Assign(ctx, input); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid assignment scope accepted")}
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid filter accepted")}
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
				return testutil.Observation{Err: errors.New("missing service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.Create(canceled, resource.CreateRequest{Identity: "Fixture"}); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("cancellation lost")}
		}
		return testutil.Observation{}
	}})
}

func TestUnitBrandingMutationsNeverReplayAndErrorsKeepStatus(t *testing.T) {
	t.Parallel()

	for _, method := range []string{"POST", "DELETE"} {
		for _, status := range []int{401, 403, 409, 429, 503} {
			command, path := "New-CsTeamsMeetingBrandingPolicy", collection
			if method == "DELETE" {
				command, path = "Remove-CsTeamsMeetingBrandingPolicy", item
			}
			run(t, testutil.Case{Name: fmt.Sprintf("%s-error-%d", strings.ToLower(method), status), Steps: []testutil.Exchange{wire(method, path, command, json.RawMessage(`{"Identity":"Global"}`), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				var observation testutil.Observation
				if method == "POST" {
					_, observation.Response, observation.Err = sdk.TeamsAdminAPI.MeetingBrandingPolicies.Create(ctx, resource.CreateRequest{Identity: "Global"})
				} else {
					observation.Response, observation.Err = sdk.TeamsAdminAPI.MeetingBrandingPolicies.DeleteByIdentity(ctx, "Global", policy.Controls{})
				}
				var api *client.APIError
				if !errors.As(observation.Err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(observation.Err), "fixture-private") {
					return testutil.Observation{Err: errors.New("normalized branding status/error redaction changed")}
				}
				return testutil.Observation{Response: observation.Response}
			}})
		}
	}
}
