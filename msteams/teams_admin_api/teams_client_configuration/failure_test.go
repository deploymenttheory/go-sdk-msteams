package teamsclientconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

func TestUnitClientConfigurationPatchFailuresNeverReplay(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, status := range []int{400, 401, 403, 429, 503} {
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("patch-error-%d", status), Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","RestrictedSenderList":null}`), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, response, err := sdk.TeamsAdminAPI.TeamsClientConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{RestrictedSenderList: policy.Null[string]()}})
			var api *client.APIError
			if !errors.As(err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(err), "fixture-private") {
				return contracttest.Observation{Err: errors.New("patch failure status/redaction changed")}
			}
			return contracttest.Observation{Response: response}
		}}))
	}
	evidence(t, results)
}

func TestUnitClientConfigurationInputContextAndBackendHaveNoHTTP(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "local-input-context-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsClientConfigurations
		for _, id := range []string{"", " ", ".", "..", "scope/", "/empty", "scope/..", "bad\\id", "bad\nidentity"} {
			if _, _, err := s.GetByIdentity(ctx, id); !errors.Is(err, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid identity accepted")}
			}
		}
		for _, pattern := range []string{"", "[broken"} {
			if _, _, err := s.ListFiltered(ctx, pattern); !errors.Is(err, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid filter accepted")}
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
		graph, err := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if _, _, err := resource.NewService(graph).List(ctx); !errors.Is(err, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, err := missing.List(ctx); !errors.Is(err, client.ErrClosed) {
				return contracttest.Observation{Err: errors.New("missing service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		//nolint:contextcheck // Deliberately tests rejected nil/canceled contexts before credentials or HTTP.
		for _, c := range []struct {
			ctx  context.Context
			want error
		}{{nil, client.ErrNilContext}, {canceled, context.Canceled}} {
			if _, _, err := s.GetByIdentity(c.ctx, "Global"); !errors.Is(err, c.want) {
				return contracttest.Observation{Err: errors.New("read context guard failed")}
			}
			if _, _, err := s.List(c.ctx); !errors.Is(err, c.want) {
				return contracttest.Observation{Err: errors.New("list context guard failed")}
			}
			if _, _, err := s.ListFiltered(c.ctx, "Global"); !errors.Is(err, c.want) {
				return contracttest.Observation{Err: errors.New("filter context guard failed")}
			}
			settings := resource.Settings{AllowEmailIntoChannel: policy.Set(false)}
			if _, _, err := s.PatchByIdentity(c.ctx, "Global", resource.PatchRequest{Settings: settings}); !errors.Is(err, c.want) {
				return contracttest.Observation{Err: errors.New("patch context guard failed")}
			}
			if _, _, err := s.UpdateByIdentity(c.ctx, "Global", resource.UpdateRequest{Settings: settings}); !errors.Is(err, c.want) {
				return contracttest.Observation{Err: errors.New("Set context guard failed")}
			}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{result})
}
