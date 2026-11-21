package onlinevoicemailvalidationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestValidationComposedUpdateRetainsPreparationAndOmission(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, tc := range []struct {
		name        string
		read, patch int
		reply       string
		accepted    bool
	}{
		{"success", 200, 204, "", true}, {"read-denied", 403, 0, "", false}, {"patch-throttled", 200, 429, `{"Message":"fixture-private"}`, false}, {"patch-unavailable", 200, 503, "not-json", false}, {"accepted-malformed", 200, 200, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsOnlineVoicemailValidationConfiguration", nil, tc.read, `{"Identity":"read-model-identity","AudioFileValidationEnabled":true,"AudioFileValidationUri":"https://example.invalid/preserved","Future":false}`)}
		if tc.patch != 0 {
			steps = append(steps, wire("PATCH", item, "Set-CsOnlineVoicemailValidationConfiguration", json.RawMessage(`{"Identity":"read-model-identity","AudioFileValidationEnabled":false}`), tc.patch, tc.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "update-" + tc.name, Mappings: []contracttest.Mapping{{Command: "Set-CsOnlineVoicemailValidationConfiguration", ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{AudioFileValidationEnabled: policy.Set(false)}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted != tc.accepted {
				return contracttest.Observation{Err: errors.New("configuration preparation/acceptance lost")}
			}
			if tc.name == "success" {
				if v.Before == nil || v.Before.AudioFileValidationURI == nil || *v.Before.AudioFileValidationURI != "https://example.invalid/preserved" {
					return contracttest.Observation{Err: errors.New("update preparation lost omitted setting")}
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			if e == nil || tc.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("partial update error lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestValidationListFilteringAndGlobalOrder(t *testing.T) {
	t.Parallel()
	const reply = `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"}]`
	var results []contracttest.Result
	for _, filter := range []bool{false, true} {
		set, name := "Identity", "global-order"
		if filter {
			set, name = "Filter", "filter-local"
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: "Get-CsOnlineVoicemailValidationConfiguration", ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsOnlineVoicemailValidationConfiguration", nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			if filter {
				v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.ListFiltered(ctx, "Fixture*")
				if e == nil && (len(v) != 1 || v[0].Identity == nil || *v[0].Identity != "Tag:Fixture") {
					e = errors.New("local filter changed")
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.List(ctx)
			if e == nil && (len(v) != 3 || v[0].Identity == nil || *v[0].Identity != "Global" || *v[1].Identity != "global") {
				e = errors.New("exact Global order changed")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
