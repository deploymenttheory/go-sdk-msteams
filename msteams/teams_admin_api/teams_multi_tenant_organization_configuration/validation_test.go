package teamsmultitenantorganizationconfiguration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_multi_tenant_organization_configuration"
)

func TestUnitConfigurationInputAndContextRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-input-context", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		service := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations
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
		// The concrete source guard counts only setting keys, including null/false,
		// and emits an error before token acquisition when only Force is bound.
		for _, force := range []*bool{nil, new(false), new(true)} {
			if value, response, err := service.PatchByIdentity(ctx, "Global", resource.PatchRequest{Controls: policy.Controls{Force: force}}); !errors.Is(err, resource.ErrInput) || value != nil || response != nil {
				t.Fatal("no-settings PATCH or Force-only acquired credentials or invented a receipt")
			}
			if value, response, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Controls: policy.Controls{Force: force}}); !errors.Is(err, resource.ErrInput) || value != nil || response != nil {
				t.Fatal("no-settings Update or Force-only acquired credentials or invented preparation")
			}
		}
		return testutil.Observation{}
	}})
}
