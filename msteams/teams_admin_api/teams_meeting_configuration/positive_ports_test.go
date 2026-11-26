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

func TestUnitPositivePortZeroRetainsPreparationWithoutPatch(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"ClientAudioPort", "ClientAudioPortRange", "ClientVideoPort", "ClientVideoPortRange", "ClientAppSharingPort", "ClientAppSharingPortRange"} {
		t.Run(name, func(t *testing.T) {
			var settings resource.ScalarSettings
			switch name {
			case "ClientAudioPort":
				settings.ClientAudioPort = policy.Set[uint32](0)
			case "ClientAudioPortRange":
				settings.ClientAudioPortRange = policy.Set[uint32](0)
			case "ClientVideoPort":
				settings.ClientVideoPort = policy.Set[uint32](0)
			case "ClientVideoPortRange":
				settings.ClientVideoPortRange = policy.Set[uint32](0)
			case "ClientAppSharingPort":
				settings.ClientAppSharingPort = policy.Set[uint32](0)
			case "ClientAppSharingPortRange":
				settings.ClientAppSharingPortRange = policy.Set[uint32](0)
			}
			run(t, contracttest.Case{Name: "primitive-positive-port", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, response, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: settings}})
				if !errors.Is(err, resource.ErrInput) || response != nil {
					t.Fatal("primitive zero acquired a receipt or passed validation")
				}
				return contracttest.Observation{}
			}})
			run(t, contracttest.Case{Name: "prepared-positive-port", Source: "msteams/teams_admin_api/teams_meeting_configuration/positive_ports_test.go", Steps: []contracttest.Exchange{compiledWire("GET", selected, set, nil, 200, string(readResponseFixture(t, "get_global.json")))}, Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: settings})
				if !errors.Is(err, resource.ErrInput) || value == nil || value.Before == nil || value.ReadResponse == nil || response == nil || response.StatusCode() != 200 || value.PatchInvoked || value.HTTPAccepted || value.After != nil {
					t.Fatal("zero setter error lost preparation or attempted a PATCH")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
}
