package teamscallingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_calling_policy"
)

func TestUnitCallingPolicyUpdatePreparationFailures(t *testing.T) {
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
			steps := []testutil.Exchange{compiledWire("GET", item, setCommand, nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","Description":null}`), tc.patch, ""))
			}
			run(t, testutil.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsCallingPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
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
				return testutil.Observation{Response: r}
			}})
		})
	}
}

func TestUnitCallingPolicyCompiledGrantRetainsExplicitZeroAndSignedRankRequestDenials(t *testing.T) {
	t.Parallel()

	for _, rank := range []int32{0, -2147483648, 2147483647} {
		body, e := json.Marshal(map[string]any{"policyName": "Fixture", "priority": rank})
		if e != nil {
			t.Fatal(e)
		}
		run(t, testutil.Case{Want: testutil.Expected{Error: auth.ErrForbidden}, Name: fmt.Sprintf("group-rank-%d", rank), Steps: []testutil.Exchange{compiledWire("POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsCallingPolicy", "Grant-CsTeamsCallingPolicy", body, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsCallingPolicies.AssignGroup(ctx, "group/fixture", new("Fixture"), rank)
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}
