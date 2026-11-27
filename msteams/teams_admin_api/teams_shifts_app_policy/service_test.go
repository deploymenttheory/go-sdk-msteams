package teamsshiftsapppolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shifts_app_policy"
)

const collection = "/Skype.Policy/configurations/TeamsShiftsAppPolicy"
const selected = collection + "/configuration/Global"
const get = "Get-CsTeamsShiftsAppPolicy"
const set = "Set-CsTeamsShiftsAppPolicy"

func TestUnitReadAndFiltersUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global.json")
	run(t, testutil.Case{Name: "read", Steps: []testutil.Exchange{compiledWire("GET", selected, get, nil, 200, string(raw))}, Want: testutil.Expected{JSON: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.GetByIdentity(ctx, "Global")
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
	list := readResponseFixture(t, "list.json")
	var actual []json.RawMessage
	if json.Unmarshal(list, &actual) != nil {
		t.Fatal("invalid acceptance list fixture")
	}
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"all", " ", len(actual)}, {"identity", "Glo?al", 1}, {"case", "[gG]lobal", 1}, {"missing", "missing", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", collection, get, nil, 200, string(list))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.ListFiltered(ctx, tc.pattern)
				if e == nil && len(v) != tc.count {
					t.Fatal("local filter selection differs")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitUpdateUsesSparsePutAndActualEmptyReceipt(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		flag *bool
		body string
	}{{"unbound", nil, `{"Identity":"Global"}`}, {"false", new(false), `{"Identity":"Global","AllowTimeClockLocationDetection":false}`}, {"true", new(true), `{"Identity":"Global","AllowTimeClockLocationDetection":true}`}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("PUT", selected, set, json.RawMessage(tc.body), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.UpdateGlobal(ctx, resource.UpdateRequest{AllowTimeClockLocationDetection: tc.flag})
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitIdentityIsOneEscapedSegment(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "scoped-identity", Steps: []testutil.Exchange{compiledWire("GET", collection+"/configuration/Scope%2Fa%20b%3F%23%25", get, nil, 200, string(readResponseFixture(t, "get_global.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.GetByIdentity(ctx, "Scope/a b?#%")
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitMalformedAndUnexpectedSuccessKeepMetadata(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{{"null", "null", 200, client.ErrDecode}, {"array", "[]", 200, client.ErrDecode}, {"bad-boolean", `{"AllowTimeClockLocationDetection":"false"}`, 200, client.ErrDecode}, {"unexpected-status", "", 204, resource.ErrResponse}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", selected, get, nil, tc.status, tc.body)}, Want: testutil.Expected{Error: tc.err}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies.GetByIdentity(ctx, "Global")
				if v != nil || r == nil || r.StatusCode() != tc.status {
					t.Fatal("failed read lost HTTP response or returned a configuration")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitValidationAndCancellationMakeNoRequests(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsShiftsAppPolicies
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid identity reached transport")
			}
			if _, e := s.UpdateByIdentity(ctx, id, resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid update identity reached transport")
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[bad"); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid filter reached transport")
		}
		canceled, stop := context.WithCancel(ctx)
		stop()
		if _, e := s.UpdateGlobal(canceled, resource.UpdateRequest{}); !errors.Is(e, context.Canceled) {
			t.Fatal("cancellation lost")
		}
		for _, empty := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := empty.List(ctx); !errors.Is(e, client.ErrClosed) {
				t.Fatal("uninitialized service accepted")
			}
		}
		return testutil.Observation{}
	}})
}
