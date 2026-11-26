package teamsmeetingconfiguration_test

import (
	"context"
	"encoding/json"
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
		for _, id := range []string{"", "c0", "c100", "C1", "c01", " c1", "x", "c1\r\n", "c1\n\n"} {
			if _, _, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{PublishedEntraAuthenticationContexts: resource.AddAuthenticationContexts(bad(id))}); !errors.Is(err, resource.ErrInput) {
				t.Fatal("invalid context Id accepted for update")
			}
			if _, _, err := service.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{PublishedEntraAuthenticationContexts: policy.Set([]*resource.PublishedEntraAuthenticationContext{bad(id)})}}); !errors.Is(err, resource.ErrInput) {
				t.Fatal("invalid context Id accepted for patch")
			}
		}
		var missing *resource.PublishedEntraAuthenticationContext
		if !errors.Is(missing.Validate(), resource.ErrInput) {
			t.Fatal("standalone nil context accepted")
		}
		if _, _, err := service.PatchByIdentity(ctx, "", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{EnableQoS: policy.Set(true)}}}); !errors.Is(err, resource.ErrInput) {
			t.Fatal("empty Set identity accepted")
		}
		return contracttest.Observation{}
	}})
}

func TestUnitContextIdMatchesPinnedSetterPattern(t *testing.T) {
	t.Parallel()
	for _, id := range []string{"c1", "c9", "c10", "c55", "c99", "c1\n", "c99\n"} {
		value := id
		item := &resource.PublishedEntraAuthenticationContext{ID: &value}
		if item.Validate() != nil {
			t.Fatalf("valid context Id %s rejected", id)
		}
		if item.ID == nil || *item.ID != id {
			t.Fatal("validation changed the assigned Id")
		}
	}
	if (&resource.PublishedEntraAuthenticationContext{}).Validate() != nil {
		t.Fatal("unassigned constructor-default Id rejected")
	}
}

// Authored from the pinned setter/compiled PATCH contract, not a tenant capture.
func TestUnitContextIdFinalLFIsPreservedInRequest(t *testing.T) {
	t.Parallel()
	id := "c1\n"
	item := &resource.PublishedEntraAuthenticationContext{ID: &id}
	body := json.RawMessage(`{"Identity":"Global","PublishedEntraAuthenticationContexts":[{"Identity":null,"Id":"c1\n","PublishedName":null,"PublishedDescription":null,"EntraName":null,"EntraDescription":null,"Published":null}]}`)
	run(t, contracttest.Case{Name: "context-id-final-lf", Source: "msteams/teams_admin_api/teams_meeting_configuration/write_validation_test.go", Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("PATCH", selected, set, body, 204, ""),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, response, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{
			PublishedEntraAuthenticationContexts: policy.Set([]*resource.PublishedEntraAuthenticationContext{item}),
		}})
		if item.ID == nil || *item.ID != "c1\n" {
			t.Fatal("request normalized the caller's Id")
		}
		return contracttest.Observation{Response: response, Err: err}
	}})
}
