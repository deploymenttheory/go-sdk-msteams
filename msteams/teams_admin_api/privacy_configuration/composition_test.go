package privacyconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
)

func TestPrivacyComposedUpdateRetainsPreparationAndOmission(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, tc := range []struct {
		name            string
		read, patch     int
		readBody, reply string
		accepted        bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","EnablePrivacyMode":true,"AutoInitiateContacts":true,"PublishLocationDataDefault":false,"Future":false}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Identity":"read-model-identity"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Identity":"read-model-identity"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Identity":"read-model-identity"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsPrivacyConfiguration", nil, tc.read, tc.readBody)}
		if tc.patch != 0 {
			steps = append(steps, wire("PATCH", item, "Set-CsPrivacyConfiguration", json.RawMessage(`{"Identity":"read-model-identity","EnablePrivacyMode":false}`), tc.patch, tc.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "update-" + tc.name, Mappings: []contracttest.Mapping{{Command: "Set-CsPrivacyConfiguration", ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{EnablePrivacyMode: policy.Set(false)}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted != tc.accepted {
				return contracttest.Observation{Err: errors.New("privacy preparation/acceptance lost")}
			}
			if tc.name == "success" {
				if v.Before == nil || v.Before.AutoInitiateContacts == nil || !*v.Before.AutoInitiateContacts || v.After != nil {
					return contracttest.Observation{Err: errors.New("update preparation or empty result lost")}
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

func TestPrivacyListFilteringAndGlobalOrder(t *testing.T) {
	t.Parallel()
	const reply = `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"},{"Identity":"fixture-two"}]`
	var results []contracttest.Result
	for _, filter := range []bool{false, true} {
		set, name := "Identity", "global-order"
		if filter {
			set, name = "Filter", "filter-local"
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: "Get-CsPrivacyConfiguration", ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsPrivacyConfiguration", nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			if filter {
				v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.ListFiltered(ctx, "Fixture*")
				if e == nil && (len(v) != 2 || v[0].Identity == nil || *v[0].Identity != "Tag:Fixture" || *v[1].Identity != "fixture-two") {
					e = errors.New("local filter changed")
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.List(ctx)
			if e == nil && (len(v) != 4 || v[0].Identity == nil || *v[0].Identity != "Global" || *v[1].Identity != "global") {
				e = errors.New("exact Global order changed")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestPrivacyWhitespaceFilterRemainsLiteral(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "literal-whitespace-filter", Mappings: []contracttest.Mapping{{Command: "Get-CsPrivacyConfiguration", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsPrivacyConfiguration", nil, 200, `[{"Identity":"Global"},{"Identity":" "},{"Identity":"Tag: "}]`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.ListFiltered(ctx, " ")
		if e == nil && (len(v) != 2 || v[0].Identity == nil || *v[0].Identity != " " || *v[1].Identity != "Tag: ") {
			e = errors.New("bound whitespace filter became an unfiltered list")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
	evidence(t, []contracttest.Result{result})
}

func TestPrivacyUpdateKeepsMissingBodyIdentity(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, read := range []struct{ name, body string }{{"null", `{"Identity":null,"AutoInitiateContacts":true}`}, {"absent", `{"AutoInitiateContacts":true}`}} {
		results = append(results, run(t, contracttest.Case{Name: "update-identity-" + read.name, Mappings: []contracttest.Mapping{{Command: "Set-CsPrivacyConfiguration", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
			wire("GET", item, "Set-CsPrivacyConfiguration", nil, 200, read.body),
			wire("PATCH", item, "Set-CsPrivacyConfiguration", json.RawMessage(`{"Identity":null,"AutoInitiateContacts":false}`), 204, ""),
		}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{AutoInitiateContacts: policy.Set(false)}})
			if e == nil && (v == nil || v.Before == nil || v.Before.Identity != nil || !v.HTTPAccepted || !v.PatchInvoked) {
				e = errors.New("source body identity or acceptance changed")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
