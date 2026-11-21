package teamsapppermissionpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

func TestAppPermissionUpdateRetainsPreparationAndAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name            string
		read, patch     int
		readBody, reply string
		accepted        bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","Description":"before","GlobalCatalogApps":[{"Id":"keep"}],"Future":false}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Identity":"read-model-identity"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Identity":"read-model-identity"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Identity":"read-model-identity"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsTeamsAppPermissionPolicy", nil, c.read, c.readBody)}
		if c.patch != 0 {
			steps = append(steps, wire("PATCH", item, "Set-CsTeamsAppPermissionPolicy", json.RawMessage(`{"Identity":"read-model-identity","Description":null}`), c.patch, c.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "update-" + c.name, Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsAppPermissionPolicy", ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Null[string]()}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (c.patch != 0) || v.HTTPAccepted != c.accepted {
				return contracttest.Observation{Err: errors.New("app permission preparation/acceptance lost")}
			}
			if c.name == "success" {
				if v.Before == nil || v.Before.Description == nil || *v.Before.Description != "before" || len(v.Before.GlobalCatalogApps) != 1 || v.After != nil {
					return contracttest.Observation{Err: errors.New("app permission read model changed")}
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			if e == nil || c.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("app permission partial failure lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestAppPermissionUpdatesEachReadonlyCatalogWithoutRawReplay(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, field := range []string{"DefaultCatalogApps", "GlobalCatalogApps", "PrivateCatalogApps"} {
		read := `{"Identity":"read-model-identity","Future":false,"` + field + `":[{"Identity":"first","Id":"a"}],"` + field + `":[{"Identity":"second","Id":"a","Future":false},{"Id":"A"}]}`
		body := `{"Identity":"read-model-identity","` + field + `":[{"Identity":"second","Id":"a"},{"Identity":null,"Id":"A"},{"Identity":null,"Id":"b"}]}`
		results = append(results, run(t, contracttest.Case{Name: "edit-" + strings.ToLower(field), Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsAppPermissionPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire("GET", item, "Set-CsTeamsAppPermissionPolicy", nil, 200, read), wire("PATCH", item, "Set-CsTeamsAppPermissionPolicy", json.RawMessage(body), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			edit := resource.ChangeCatalogApps([]*resource.CatalogApp{resource.NewCatalogApp("b")}, []*resource.CatalogApp{resource.NewCatalogApp("a")})
			input := resource.UpdateRequest{}
			switch field {
			case "DefaultCatalogApps":
				input.DefaultCatalogApps = edit
			case "GlobalCatalogApps":
				input.GlobalCatalogApps = edit
			case "PrivateCatalogApps":
				input.PrivateCatalogApps = edit
			}
			v, r, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.UpdateByIdentity(ctx, "Global", input)
			if e == nil && (v == nil || v.Before == nil || !v.HTTPAccepted || v.Before.RawFields["Future"] == nil) {
				e = errors.New("catalog preparation lost")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestAppPermissionCatalogBuildersHaveNoCredentialsOrHTTP(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "local-catalog-builders", Source: sourceModels, Invoke: func(_ context.Context, _ *msteams.Client) contracttest.Observation {
		for _, id := range []string{"", " ", "future-app"} {
			v := resource.NewCatalogApp(id)
			out, e := resource.ReplaceCatalogApps(v, v).Apply(nil)
			if e != nil || len(out) != 2 || out[0].ID == nil || *out[0].ID != id {
				return contracttest.Observation{Err: errors.New("local catalog construction changed")}
			}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{result})
}
