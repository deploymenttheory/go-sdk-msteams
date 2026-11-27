package teamsnetworkroamingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_network_roaming_policy"
)

const collection = "/Skype.Policy/configurations/TeamsNetworkRoamingPolicy"
const selected = collection + "/configuration/Global"
const get = "Get-CsTeamsNetworkRoamingPolicy"
const set = "Set-CsTeamsNetworkRoamingPolicy"
const create = "New-CsTeamsNetworkRoamingPolicy"

func TestUnitSparsePutPreservesFalseEmptyZeroAndInt64(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		body string
		args resource.UpdateRequest
	}{{"unbound", `{"Identity":"Global"}`, resource.UpdateRequest{}}, {"zero", `{"Identity":"Global","AllowIPVideo":false,"MediaBitRateKb":0,"Description":""}`, resource.UpdateRequest{Description: new(""), AllowIPVideo: new(false), MediaBitRateKb: new(int64(0))}}, {"int64", `{"Identity":"Global","MediaBitRateKb":9223372036854775807}`, resource.UpdateRequest{MediaBitRateKb: new(int64(math.MaxInt64))}}} {
		t.Run(tc.name, func(t *testing.T) {
			// These are request-binding vectors, not live claims that the service
			// accepts zero or extreme bitrates. Terminate with a controlled failure.
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("PUT", selected, set, json.RawMessage(tc.body), 503, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				r, e := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies.UpdateGlobal(ctx, tc.args)
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitCreateCompositionKeepsTheAcceptedPostWhenReadBackFails(t *testing.T) {
	t.Parallel()
	missing := readResponseFixture(t, "get_deleted_404.json")
	run(t, testutil.Case{Name: "create-read-failure", Steps: []testutil.Exchange{
		compiledWire("POST", collection, create, json.RawMessage(`{"Identity":"fixture-policy"}`), 201, string(readResponseFixture(t, "create_201.body"))),
		compiledWire("GET", collection+"/configuration/fixture-policy", get, nil, 404, string(missing)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies.Create(ctx, resource.CreateRequest{Identity: "fixture-policy"})
		var partial *resource.ReadBackError
		var api *client.APIError
		if v == nil || v.Identity != "fixture-policy" || !errors.As(e, &partial) || partial.CreationResponse == nil || partial.CreationResponse.StatusCode() != 201 || !errors.As(e, &api) || api.StatusCode != 404 {
			t.Fatal("accepted POST or subsequent failure was lost")
		}
		return testutil.Observation{Response: r}
	}})
}

func TestUnitMutationFailuresKeepStatusWithoutReplayOrReadBack(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"create", "update", "delete"} {
		for _, status := range []int{401, 403, 429, 503} {
			method, command, path, body := "POST", create, collection, json.RawMessage(`{"Identity":"Global"}`)
			switch operation {
			case "update":
				method, command, path = "PUT", set, selected
			case "delete":
				method, command, path, body = "DELETE", "Remove-CsTeamsNetworkRoamingPolicy", selected, nil
			}
			run(t, testutil.Case{Name: operation, Steps: []testutil.Exchange{compiledWire(method, path, command, body, status, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies
				switch operation {
				case "create":
					_, r, e := s.Create(ctx, resource.CreateRequest{Identity: "Global"})
					return testutil.Observation{Response: r, Err: e}
				case "update":
					r, e := s.UpdateGlobal(ctx, resource.UpdateRequest{})
					return testutil.Observation{Response: r, Err: e}
				default:
					r, e := s.DeleteByIdentity(ctx, "Global")
					return testutil.Observation{Response: r, Err: e}
				}
			}})
		}
	}
}

func TestUnitFiltersAndEscapingUseModernRules(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"all", " ", 2}, {"global", "[gG]lobal", 1}, {"tagged", "fixture-polic?", 1}, {"missing", "missing", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", collection, get, nil, 200, string(readResponseFixture(t, "owned_list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies.ListFiltered(ctx, tc.pattern)
				if e == nil && len(v) != tc.count {
					t.Fatal("local Identity/Tag filter result differs")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
	run(t, testutil.Case{Name: "escaped", Steps: []testutil.Exchange{compiledWire("GET", collection+"/configuration/Scope%2Fa%20b%3F%23%25", get, nil, 200, string(readResponseFixture(t, "get_global.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies.GetByIdentity(ctx, "Scope/a b?#%")
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitInvalidInputContextAndClosedServicesStayLocal(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid read identity reached HTTP")
			}
			if _, e := s.CreateByIdentity(ctx, id, resource.Settings{}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid create identity reached HTTP")
			}
			if _, e := s.UpdateByIdentity(ctx, id, resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid update identity reached HTTP")
			}
			if _, e := s.DeleteByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid delete identity reached HTTP")
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[bad"); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid wildcard reached HTTP")
		}
		canceled, stop := context.WithCancel(ctx)
		stop()
		if _, e := s.UpdateGlobal(canceled, resource.UpdateRequest{}); !errors.Is(e, context.Canceled) {
			t.Fatal("canceled update reached HTTP")
		}
		for _, v := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := v.List(ctx); !errors.Is(e, client.ErrClosed) {
				t.Fatal("uninitialized service accepted")
			}
		}
		return testutil.Observation{}
	}})
}
