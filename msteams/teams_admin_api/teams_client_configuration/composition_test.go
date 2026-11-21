package teamsclientconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

func TestUnitClientConfigurationSetRetainsPreparationAndAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name        string
		read, patch int
		body, reply string
		accepted    bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","RestrictedSenderList":"before","ContentPin":"0000","AllowEmailIntoChannel":false,"Future":false}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Identity":"read-model-identity"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Identity":"read-model-identity"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Identity":"read-model-identity"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{compiledWire("GET", item, setCommand, nil, c.read, c.body)}
		if c.patch != 0 {
			steps = append(steps, compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"read-model-identity","RestrictedSenderList":null}`), c.patch, c.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "set-" + c.name, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsClientConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{RestrictedSenderList: policy.Null[string]()}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (c.patch != 0) || v.HTTPAccepted != c.accepted {
				return contracttest.Observation{Err: errors.New("compiled Set preparation/acceptance changed")}
			}
			if c.name == "success" {
				if v.Before == nil || v.Before.RestrictedSenderList == nil || *v.Before.RestrictedSenderList != "before" || v.Before.ContentPin == nil || *v.Before.ContentPin != "0000" || v.Before.AllowEmailIntoChannel == nil || *v.Before.AllowEmailIntoChannel || v.Before.RawFields["Future"] == nil {
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
