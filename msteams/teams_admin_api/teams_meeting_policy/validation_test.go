package teamsmeetingpolicy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_policy"
)

func TestUnitMeetingPolicyInputAndContextRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "local-input-context-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsMeetingPolicies
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
		if _, e := s.AssignTenantDefault(canceled, "", false); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		return contracttest.Observation{}
	}})
}

func TestUnitMeetingSetterConstraintsRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	for _, settings := range []resource.Settings{
		{MediaBitRateKb: policy.Set(uint32(0))},
		{RecordingAndTranscriptionCustomMessageIdentifier: policy.Set("")},
		{SyntheticMediaDetectionAppId: policy.Set("bad-guid")},
		{PreMeetingConsentContentIdentifier: policy.Set("00000000-0000-4000-8000-000000000001suffix")},
	} {
		run(t, contracttest.Case{Name: "setter-validation", Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			service := sdk.TeamsAdminAPI.TeamsMeetingPolicies
			if _, _, err := service.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{Settings: settings}); !errors.Is(err, resource.ErrInput) {
				t.Fatal("invalid create setting reached transport")
			}
			if _, _, err := service.UpdateByIdentity(ctx, "fixture-policy", resource.UpdateRequest{Settings: settings}); !errors.Is(err, resource.ErrInput) {
				t.Fatal("invalid setting reached Set preparation")
			}
			_, r, e := service.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: settings})
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
