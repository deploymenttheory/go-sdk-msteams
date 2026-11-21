package teamsfeedbackpolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

func TestFeedbackPolicyInterleavesProtocolsWithoutHeaderLeakage(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		results = append(results, run(t, contracttest.Case{Name: "interleaved-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: "Identity"}, {Command: grantCommand, ParameterSet: "Identity"}, {Command: grantCommand, ParameterSet: "GrantToTenant"}, {Command: deleteCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
			compiledWire("GET", item, getCommand, nil, 200, `{"Identity":"Global"}`),
			discovery(), modernWire("PATCH", "/Skype.Policy/users/fixture%2Fuser/policies/TeamsFeedbackPolicy", json.RawMessage(`{"policyName":"Fixture"}`), 200, ""),
			compiledWire("GET", item, getCommand, nil, 200, `{"Identity":"Global"}`),
			modernWire("PATCH", "/Skype.Policy/tenants/policies/TeamsFeedbackPolicy", json.RawMessage(`{"policyName":"Fixture","forceSwitchPresent":false}`), 200, ""),
			compiledWire("DELETE", item, deleteCommand, json.RawMessage(`{"Identity":"Global"}`), 204, ""),
		}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsFeedbackPolicies
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

func TestFeedbackPolicySetRetainsPreparationAndAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name        string
		read, patch int
		body, reply string
		accepted    bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","UserInitiatedMode":"before","AllowLogCollection":false,"Future":false}`, "", true},
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
			body := json.RawMessage(`{"Identity":"read-model-identity","UserInitiatedMode":null}`)
			if c.name == "success-null-identity" {
				body = json.RawMessage(`{"Identity":null,"UserInitiatedMode":null}`)
			}
			steps = append(steps, compiledWire("PATCH", item, setCommand, body, c.patch, c.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "set-" + c.name, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsFeedbackPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{UserInitiatedMode: policy.Null[string]()}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (c.patch != 0) || v.HTTPAccepted != c.accepted {
				return contracttest.Observation{Err: errors.New("compiled Set preparation/acceptance changed")}
			}
			if c.name == "success" {
				if v.Before == nil || v.Before.UserInitiatedMode == nil || *v.Before.UserInitiatedMode != "before" || v.Before.AllowLogCollection == nil || *v.Before.AllowLogCollection || v.Before.RawFields["Future"] == nil {
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
