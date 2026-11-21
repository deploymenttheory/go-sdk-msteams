package teamscallholdpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
)

func TestUnitCallHoldInterleavesProtocolsWithoutHeaderLeakageOnDeniedWrites(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		run(t, contracttest.Case{Name: "interleaved-" + string(mode), Mode: mode, Source: "msteams/teams_admin_api/teams_call_hold_policy/composition_test.go", Steps: []contracttest.Exchange{compiledWire("GET", item, getCommand, nil, 200, string(before)), discovery(t), modernWire("PATCH", "/Skype.Policy/users/fixture%2Fuser/policies/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":"Fixture"}`), 403, ""), compiledWire("GET", item, getCommand, nil, 200, string(before)), modernWire("PATCH", "/Skype.Policy/tenants/policies/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":"Fixture","forceSwitchPresent":false}`), 403, ""), compiledWire("DELETE", item, deleteCommand, json.RawMessage(`{"Identity":"Global"}`), 403, "")}, Want: contracttest.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsCallHoldPolicies
			if _, _, e := s.GetByIdentity(ctx, "Global"); e != nil {
				t.Fatal("real Global read failed before modern write")
			}
			if _, e := s.AssignToUser(ctx, "fixture/user", "Fixture"); !errors.Is(e, auth.ErrForbidden) {
				t.Fatal("user denial was not preserved")
			}
			if _, _, e := s.GetByIdentity(ctx, "Global"); e != nil {
				t.Fatal("modern denial changed compiled read headers")
			}
			if _, e := s.AssignTenantDefault(ctx, "Fixture", false); !errors.Is(e, auth.ErrForbidden) {
				t.Fatal("tenant-default denial was not preserved")
			}
			_, r, e := s.DeleteByIdentity(ctx, "Global", policy.Controls{})
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}

func TestUnitCallHoldUpdatePreparationFailures(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	var expectedBefore map[string]json.RawMessage
	if json.Unmarshal(before, &expectedBefore) != nil {
		t.Fatal("invalid real preparation fixture")
	}
	cases := []struct {
		name        string
		read, patch int
		body        string
	}{
		{"read-denied", 403, 0, ""},
		{"read-empty", 204, 0, ""},
		{"read-malformed", 200, 0, "not-json"},
	}
	cases = append(cases,
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-denied", 200, 403, string(before)},
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-throttled", 200, 429, string(before)},
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-unavailable", 200, 503, string(before)},
	)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []contracttest.Exchange{compiledWire("GET", item, setCommand, nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","Description":null}`), tc.patch, ""))
			}
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/teams_call_hold_policy/composition_test.go", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsCallHoldPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
				if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted {
					t.Error("failed update lost preparation state or invented acceptance")
				}
				if tc.patch != 0 && (v.Before == nil || !reflect.DeepEqual(v.Before.RawFields, expectedBefore)) {
					t.Error("failed update changed or lost the actual preparation snapshot")
				}
				if tc.read == 403 || tc.patch != 0 {
					var apiError *client.APIError
					expected := tc.read
					if tc.patch != 0 {
						expected = tc.patch
					}
					if !errors.As(e, &apiError) || apiError.StatusCode != expected {
						t.Error("failure lost its HTTP status")
					}
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}
