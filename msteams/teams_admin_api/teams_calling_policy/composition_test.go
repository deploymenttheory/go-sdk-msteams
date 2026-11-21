package teamscallingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_calling_policy"
)

func TestCallingPolicySetRetainsPreparationAndAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name        string
		read, patch int
		body, reply string
		accepted    bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","Description":"before","CallingSpendUserLimit":9007199254740993,"Future":false}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Identity":"read-model-identity"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Identity":"read-model-identity"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Identity":"read-model-identity"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{compiledWire("GET", item, setCommand, nil, c.read, c.body)}
		if c.patch != 0 {
			steps = append(steps, compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"read-model-identity","Description":null}`), c.patch, c.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "set-" + c.name, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsCallingPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (c.patch != 0) || v.HTTPAccepted != c.accepted {
				return contracttest.Observation{Err: errors.New("compiled Set preparation/acceptance changed")}
			}
			if c.name == "success" {
				if v.Before == nil || v.Before.Description == nil || *v.Before.Description != "before" || v.Before.CallingSpendUserLimit == nil || *v.Before.CallingSpendUserLimit != 9007199254740993 || v.Before.RawFields["Future"] == nil {
					return contracttest.Observation{Err: errors.New("preparation snapshot changed")}
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			if e == nil || c.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("partial Set failure lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestCallingPolicyCompiledGrantRetainsExplicitZeroAndSignedRank(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, rank := range []int32{0, -2147483648, 2147483647} {
		body, e := json.Marshal(map[string]any{"policyName": "Fixture", "priority": rank})
		if e != nil {
			t.Fatal(e)
		}
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("group-rank-%d", rank), Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsCallingPolicy", ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{compiledWire("POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsCallingPolicy", "Grant-CsTeamsCallingPolicy", body, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsCallingPolicies.AssignGroup(ctx, "group/fixture", new("Fixture"), rank)
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
