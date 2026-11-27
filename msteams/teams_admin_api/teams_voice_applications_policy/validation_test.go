package teamsvoiceapplicationspolicy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_voice_applications_policy"
)

func TestUnitConfigurationInputAndContextRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-input-context", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		service := sdk.TeamsAdminAPI.TeamsVoiceApplicationsPolicies
		for _, id := range []string{"", " ", ".", "..", "scope/", "/empty", "bad\\id", "bad\nidentity"} {
			if _, _, err := service.GetByIdentity(ctx, id); !errors.Is(err, resource.ErrInput) {
				t.Fatal("invalid configuration identity acquired credentials")
			}
		}
		for _, filter := range []string{"", "[broken"} {
			if _, _, err := service.ListFiltered(ctx, filter); !errors.Is(err, resource.ErrInput) {
				t.Fatal("invalid filter acquired credentials")
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, err := service.List(canceled); !errors.Is(err, context.Canceled) {
			t.Fatal("cancellation lost")
		}
		//nolint:contextcheck,staticcheck // Deliberately verify SDK rejection of nil caller contexts (SA1012).
		if _, _, err := service.GetByIdentity(nil, "Global"); !errors.Is(err, client.ErrNilContext) {
			t.Fatal("nil context accepted")
		}
		for _, value := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, err := value.List(ctx); !errors.Is(err, client.ErrClosed) {
				t.Fatal("missing service accepted")
			}
		}
		graph, err := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if err != nil {
			t.Fatal(err)
		}
		if _, _, err := resource.NewService(graph).List(ctx); !errors.Is(err, client.ErrDestination) {
			t.Fatal("wrong backend acquired credentials")
		}
		return testutil.Observation{}
	}})
}
