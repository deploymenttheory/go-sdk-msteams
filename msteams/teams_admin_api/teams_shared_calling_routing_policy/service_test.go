package teamssharedcallingroutingpolicy_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shared_calling_routing_policy"
)

func TestUnitEmergencyEditsSendReplacementArraysWithoutInventedOperators(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		edit *policy.ListEdit
		body string
	}{{"replacement", policy.ReplaceStrings("+12025550101"), `{"Identity":"Global","EmergencyNumbers":["+12025550101"]}`}, {"add", policy.ChangeStrings([]string{"+12025550101"}, nil), `{"Identity":"Global","EmergencyNumbers":["+12025550101"]}`}, {"remove-absent", policy.ChangeStrings(nil, []string{"+12025550101"}), `{"Identity":"Global","EmergencyNumbers":[]}`}, {"clear", policy.ReplaceStrings(), `{"Identity":"Global","EmergencyNumbers":[]}`}} {
		t.Run(tc.name, func(t *testing.T) {
			// The preparation response is real. Hypothetical phone bindings end
			// with a controlled failure and do not claim tenant number acceptance.
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{
				compiledWire("GET", item, "Set-CsTeamsSharedCallingRoutingPolicy", nil, 200, string(readResponseFixture(t, "get_global.json"))),
				compiledWire("PATCH", item, "Set-CsTeamsSharedCallingRoutingPolicy", json.RawMessage(tc.body), 503, ""),
			}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				result, r, e := sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{EmergencyNumbers: tc.edit})
				if result == nil || result.Before == nil || !result.PatchInvoked || result.HTTPAccepted {
					t.Fatal("preparation, mutation invocation or failure state lost")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitScalarNullEmptyAndForceDoNotBindTheList(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "scalar-bindings", Steps: []testutil.Exchange{
		compiledWire("PATCH", item, "Set-CsTeamsSharedCallingRoutingPolicy", json.RawMessage(`{"Identity":"Global","Description":"","ResourceAccount":null,"Force":"False"}`), 503, ""),
	}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{Description: policy.Set(""), ResourceAccount: policy.Null[string]()}}, Controls: policy.Controls{Force: new(false)}})
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitSourceListDefaultsPopulationAndNullSetter(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, input string
		want        []string
		invalid     bool
	}{{"constructor", `{}`, []string{}, false}, {"populate", `{"EmergencyNumbers":["112"],"emergencyNumbers":["999"]}`, []string{"112", "999"}, false}, {"null-setter", `{"EmergencyNumbers":null}`, nil, true}, {"object", `{"EmergencyNumbers":{}}`, nil, true}} {
		t.Run(tc.name, func(t *testing.T) {
			// Source-derived decoder vectors are not successful server fixtures.
			var value resource.Policy
			err := json.Unmarshal([]byte(tc.input), &value)
			if (err != nil) != tc.invalid || !tc.invalid && (value.EmergencyNumbers == nil || !slices.Equal(value.EmergencyNumbers, tc.want)) {
				t.Fatal("source constructor/list/null semantics changed")
			}
		})
	}
}
