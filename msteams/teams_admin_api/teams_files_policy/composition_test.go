package teamsfilespolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
)

func TestFilesPolicySetRetainsPreparationAndAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name        string
		read, patch int
		body, reply string
		accepted    bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","NativeFileEntryPoints":"before","SPChannelFilesTab":"retained","Future":false}`, "", true},
		{"success-null-identity", 200, 204, `{"Identity":null}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Identity":"read-model-identity"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Identity":"read-model-identity"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Identity":"read-model-identity"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{compiledWire("GET", item, setCommand, nil, c.read, c.body)}
		if c.patch != 0 {
			body := json.RawMessage(`{"Identity":"read-model-identity","NativeFileEntryPoints":null}`)
			if c.name == "success-null-identity" {
				body = json.RawMessage(`{"Identity":null,"NativeFileEntryPoints":null}`)
			}
			steps = append(steps, compiledWire("PATCH", item, setCommand, body, c.patch, c.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "set-" + c.name, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsFilesPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{NativeFileEntryPoints: policy.Null[string]()}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (c.patch != 0) || v.HTTPAccepted != c.accepted {
				return contracttest.Observation{Err: errors.New("compiled Set preparation/acceptance changed")}
			}
			if c.name == "success" {
				if v.Before == nil || v.Before.NativeFileEntryPoints == nil || *v.Before.NativeFileEntryPoints != "before" || v.Before.SPChannelFilesTab == nil || *v.Before.SPChannelFilesTab != "retained" || v.Before.RawFields["Future"] == nil {
					return contracttest.Observation{Err: errors.New("preparation snapshot changed")}
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			if c.name == "success-null-identity" {
				if v.Before == nil || v.Before.Identity != nil {
					return contracttest.Observation{Err: errors.New("nullable preparation identity changed")}
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

func TestFilesPolicyCompiledGrantRetainsExplicitZeroAndSignedRank(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, rank := range []int32{0, -2147483648, 2147483647} {
		body, e := json.Marshal(map[string]any{"policyName": "Fixture", "priority": rank})
		if e != nil {
			t.Fatal(e)
		}
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("group-rank-%d", rank), Mappings: []contracttest.Mapping{{Command: "Grant-CsTeamsFilesPolicy", ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{compiledWire("POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsFilesPolicy", "Grant-CsTeamsFilesPolicy", body, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsFilesPolicies.AssignGroup(ctx, "group/fixture", new("Fixture"), rank)
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
