package teamsmeetingconfiguration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
)

// Unbound requests, zero or invalid list edits and Ids the pinned
// PublishedEntraAuthenticationContext.Id setter refuses fail before HTTP.
func TestUnitWriteInputRejectsBeforeHTTP(t *testing.T) {
	t.Parallel()
	bad := func(id string) *resource.PublishedEntraAuthenticationContext {
		return &resource.PublishedEntraAuthenticationContext{ID: &id}
	}
	run(t, contracttest.Case{Name: "local-write-input", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		service := sdk.TeamsAdminAPI.TeamsMeetingConfigurations
		if _, _, err := service.PatchByIdentity(ctx, "Global", resource.PatchRequest{}); !errors.Is(err, resource.ErrInput) {
			t.Fatal("unbound patch accepted")
		}
		if _, _, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{}); !errors.Is(err, resource.ErrInput) {
			t.Fatal("unbound update accepted")
		}
		if _, _, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{PublishedEntraAuthenticationContexts: &resource.AuthenticationContextEdit{}}); !errors.Is(err, resource.ErrInput) {
			t.Fatal("zero list edit accepted")
		}
		for _, id := range []string{"", "c0", "c100", "C1", "c01", " c1", "x"} {
			if _, _, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{PublishedEntraAuthenticationContexts: resource.AddAuthenticationContexts(bad(id))}); !errors.Is(err, resource.ErrInput) {
				t.Fatal("invalid context Id accepted for update")
			}
			if _, _, err := service.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{PublishedEntraAuthenticationContexts: policy.Set([]*resource.PublishedEntraAuthenticationContext{bad(id)})}}); !errors.Is(err, resource.ErrInput) {
				t.Fatal("invalid context Id accepted for patch")
			}
		}
		if _, _, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{PublishedEntraAuthenticationContexts: resource.ReplaceAuthenticationContexts(nil)}); !errors.Is(err, resource.ErrInput) {
			t.Fatal("nil context element accepted")
		}
		if _, _, err := service.PatchByIdentity(ctx, "", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{EnableQoS: policy.Set(true)}}}); !errors.Is(err, resource.ErrInput) {
			t.Fatal("empty Set identity accepted")
		}
		return contracttest.Observation{}
	}})
}

func TestUnitContextIdMatchesPinnedSetterPattern(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"c1", "c9", "c10", "c55", "c99"} {
		value := id
		if (&resource.PublishedEntraAuthenticationContext{ID: &value}).Validate() != nil {
			t.Fatalf("valid context Id %s rejected", id)
		}
	}
	if (&resource.PublishedEntraAuthenticationContext{}).Validate() != nil {
		t.Fatal("unassigned constructor-default Id rejected")
	}
}
