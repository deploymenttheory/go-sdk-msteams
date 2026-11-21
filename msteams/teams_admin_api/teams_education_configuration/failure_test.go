package teamseducationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

func TestUnitEducationConfigurationPatchDoesNotReplayOrExposeKey(t *testing.T) {
	t.Parallel()

	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			run(t, contracttest.Case{Name: fmt.Sprintf("patch-%s-%d", mode, status), Mode: mode, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","UpdateParentInformation":"fixture-private"}`), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsEducationConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{UpdateParentInformation: policy.Set("fixture-private")}})
				if e == nil || strings.Contains(e.Error(), "fixture-private") {
					return contracttest.Observation{Err: errors.New("failed mutation or secret redaction lost")}
				}
				return contracttest.Observation{Response: r}
			}})
		}
	}
}

func TestUnitEducationConfigurationLocalGuardsBeforeCredentials(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "local-guards", Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: "Identity"}, {Command: getCommand, ParameterSet: "Filter"}, {Command: setCommand, ParameterSet: "Identity"}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsEducationConfigurations
		for _, identity := range []string{"", "/missing", "Global/", "bad\\identity", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, identity); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid selector accepted")}
			}
		}
		if _, _, e := s.ListFiltered(ctx, ""); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty filter accepted")}
		}
		for _, input := range []resource.Settings{{}} {
			if _, _, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: input}); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid sparse input accepted")}
			}
			if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: input}); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid composed input reached read")}
			}
		}
		canceled, stop := context.WithCancel(ctx)
		stop()
		if _, _, e := s.GetByIdentity(canceled, "Global"); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		var absent *resource.Service
		if _, _, e := absent.List(ctx); !errors.Is(e, client.ErrClosed) {
			return contracttest.Observation{Err: errors.New("nil service accepted")}
		}

		//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before credentials or HTTP.
		if _, _, e := absent.List(nil); !errors.Is(e, client.ErrNilContext) {
			return contracttest.Observation{Err: errors.New("nil context accepted")}
		}
		return contracttest.Observation{}
	}})
}
