package teamscallholdpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
)

func TestCallHoldInterleavesProtocolsWithoutHeaderLeakage(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		results = append(results, run(t, contracttest.Case{Name: "interleaved-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: "Identity"}, {Command: grantCommand, ParameterSet: "Identity"}, {Command: grantCommand, ParameterSet: "GrantToTenant"}, {Command: deleteCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
			compiledWire("GET", item, getCommand, nil, 200, `{"Identity":"Global"}`),
			discovery(), modernWire("PATCH", "/Skype.Policy/users/fixture%2Fuser/policies/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":"Fixture"}`), 200, ""),
			compiledWire("GET", item, getCommand, nil, 200, `{"Identity":"Global"}`),
			modernWire("PATCH", "/Skype.Policy/tenants/policies/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":"Fixture","forceSwitchPresent":false}`), 200, ""),
			compiledWire("DELETE", item, deleteCommand, json.RawMessage(`{"Identity":"Global"}`), 204, ""),
		}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsCallHoldPolicies
			if _, r, e := s.GetByIdentity(ctx, "Global"); e != nil {
				return contracttest.Observation{Response: r, Err: e}
			}
			if r, e := s.AssignToUser(ctx, "fixture/user", "Fixture"); e != nil {
				return contracttest.Observation{Response: r, Err: e}
			}
			if _, r, e := s.GetByIdentity(ctx, "Global"); e != nil {
				return contracttest.Observation{Response: r, Err: e}
			}
			if r, e := s.AssignTenantDefault(ctx, "Fixture", false); e != nil {
				return contracttest.Observation{Response: r, Err: e}
			}
			_, r, e := s.DeleteByIdentity(ctx, "Global", policy.Controls{})
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestCallHoldSetRetainsPreparationAndAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name        string
		read, patch int
		body, reply string
		accepted    bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","Description":"before","AudioFileId":"retained","Future":false}`, "", true},
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
			v, r, e := sdk.TeamsAdminAPI.TeamsCallHoldPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (c.patch != 0) || v.HTTPAccepted != c.accepted {
				return contracttest.Observation{Err: errors.New("compiled Set preparation/acceptance changed")}
			}
			if c.name == "success" {
				if v.Before == nil || v.Before.Description == nil || *v.Before.Description != "before" || v.Before.AudioFileID == nil || *v.Before.AudioFileID != "retained" || v.Before.RawFields["Future"] == nil {
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
