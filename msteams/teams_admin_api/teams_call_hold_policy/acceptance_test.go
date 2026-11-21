package teamscallholdpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitCallHoldGroupAcceptanceSurvivesOnlyActualPostResponses(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name   string
		status int
		reply  string
		bad    bool
	}{
		{"accepted-malformed", 202, "not-json", true},
		{"accepted-empty", 202, "", true},
		{"accepted-null", 202, "null", false},
		{"accepted-future", 202, `{"OperationId":"fixture-operation","Future":false}`, false},
		{"completed-empty", 204, "", false},
	} {
		results = append(results, run(t, contracttest.Case{Name: c.name, Source: grantSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{discovery(), modernWire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":"Fixture"}`), c.status, c.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsCallHoldPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
			if v == nil || v.Accepted != (c.status == 202) || c.bad && !errors.Is(e, client.ErrDecode) || !c.bad && e != nil {
				return contracttest.Observation{Err: errors.New("group acceptance/decoding changed")}
			}
			if c.name == "accepted-future" && (v.OperationID != "fixture-operation" || len(v.RawJSON) == 0) {
				return contracttest.Observation{Err: errors.New("group future payload lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}
