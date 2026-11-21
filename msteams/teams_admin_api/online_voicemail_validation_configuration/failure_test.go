package onlinevoicemailvalidationconfiguration_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitValidationInputAndContextRejectBeforeTokens(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "invalid-input-before-tokens", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity", "scope/..", "/empty-scope", "scope/"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid configuration identity accepted")}
			}
		}
		for _, settings := range []resource.Settings{{}, {AudioFileValidationURI: policy.Set("")}, {AudioFileValidationURI: policy.Set("http://example.invalid/a")}} {
			if _, _, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: settings}); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("empty/invalid native patch acquired tokens")}
			}
			if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: settings}); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid composition performed a read")}
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid local wildcard acquired tokens")}
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
				return contracttest.Observation{Err: errors.New("missing service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.UpdateByIdentity(canceled, "Global", resource.UpdateRequest{Settings: resource.Settings{AudioFileValidationEnabled: policy.Set(false)}}); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		return contracttest.Observation{}
	}})
}

func TestUnitValidationPatchErrorsKeepStatusAndNeverReplay(t *testing.T) {
	t.Parallel()

	for _, status := range []int{400, 401, 403, 429, 503} {
		run(t, contracttest.Case{Name: fmt.Sprintf("patch-error-%d", status), Mappings: []contracttest.Mapping{{Command: "Set-CsOnlineVoicemailValidationConfiguration", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire("PATCH", item, "Set-CsOnlineVoicemailValidationConfiguration", json.RawMessage(`{"Identity":"Global","AudioFileValidationUri":null}`), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{AudioFileValidationURI: policy.Null[string]()}})
			var api *client.APIError
			if !errors.As(e, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(e), "fixture-private") {
				return contracttest.Observation{Err: errors.New("HTTP status or error redaction changed")}
			}
			return contracttest.Observation{Response: r}
		}})
	}
}
