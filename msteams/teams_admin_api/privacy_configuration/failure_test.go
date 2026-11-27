package privacyconfiguration_test

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
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
)

func TestUnitPrivacyInputAndContextRejectBeforeTokens(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "invalid-input-before-tokens", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.PrivacyConfigurations
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity", "scope/..", "/empty-scope", "scope/"} {
			if _, r, e := s.GetByIdentity(ctx, id); r != nil || !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid privacy identity accepted")}
			}
		}
		for _, force := range []*bool{nil, new(false), new(true)} {
			controls := policy.Controls{Force: force}
			if _, r, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Controls: controls}); r != nil || !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("setting-free patch acquired tokens")}
			}
			if _, r, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Controls: controls}); r != nil || !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("setting-free update performed read")}
			}
		}
		for _, pattern := range []string{"", "[broken"} {
			if _, _, e := s.ListFiltered(ctx, pattern); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid local wildcard acquired tokens")}
			}
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

		//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before credentials or HTTP.
		if _, _, e := s.List(nil); !errors.Is(e, client.ErrNilContext) {
			return testutil.Observation{Err: errors.New("nil context accepted")}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.UpdateByIdentity(canceled, "Global", resource.UpdateRequest{Settings: resource.Settings{EnablePrivacyMode: policy.Set(false)}}); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("cancellation lost")}
		}
		return testutil.Observation{}
	}})
}

func TestUnitPrivacyPatchErrorsKeepStatusAndNeverReplay(t *testing.T) {
	t.Parallel()

	for _, status := range []int{400, 401, 403, 429, 503} {
		run(t, testutil.Case{Name: fmt.Sprintf("patch-error-%d", status), Steps: []testutil.Exchange{wire("PATCH", item, "Set-CsPrivacyConfiguration", json.RawMessage(`{"Identity":"Global","PublishLocationDataDefault":null}`), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{PublishLocationDataDefault: policy.Null[bool]()}})
			var api *client.APIError
			if !errors.As(e, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(e), "fixture-private") {
				return testutil.Observation{Err: errors.New("HTTP status or error redaction changed")}
			}
			return testutil.Observation{Response: r}
		}})
	}
}

func TestUnitPrivacyEmptyAndNullReadPresentation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, body string
		list       bool
	}{{"null-read", "null", false}, {"null-list", "null", true}, {"empty-list", "[]", true}, {"no-body-list", "", true}} {
		path := item
		if tc.list {
			path = collection
		}
		run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{wire("GET", path, "Get-CsPrivacyConfiguration", nil, 200, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			if tc.list {
				v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.List(ctx)
				if e == nil && (len(v) != 0 || (v == nil) != (tc.body != "[]")) {
					e = errors.New("empty collection presence changed")
				}
				return testutil.Observation{Response: r, Err: e}
			}
			v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.GetByIdentity(ctx, "Global")
			if e == nil && v != nil {
				e = errors.New("null response fabricated a configuration")
			}
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}
