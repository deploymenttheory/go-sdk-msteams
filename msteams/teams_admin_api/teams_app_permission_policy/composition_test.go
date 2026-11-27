package teamsapppermissionpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

func TestUnitAppPermissionUpdatePreparationFailures(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	var expectedBefore map[string]json.RawMessage
	if json.Unmarshal(before, &expectedBefore) != nil {
		t.Fatal("invalid real preparation fixture")
	}
	cases := []struct {
		name        string
		read, patch int
		body        string
	}{
		{"read-denied", 403, 0, ""},
		{"read-empty", 204, 0, ""},
		{"read-malformed", 200, 0, "not-json"},
	}
	cases = append(cases,
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-denied", 200, 403, string(before)},
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-throttled", 200, 429, string(before)},
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-unavailable", 200, 503, string(before)},
	)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []testutil.Exchange{wire("GET", item, "Set-CsTeamsAppPermissionPolicy", nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, wire("PATCH", item, "Set-CsTeamsAppPermissionPolicy", json.RawMessage(`{"Identity":"Global","Description":null}`), tc.patch, ""))
			}
			run(t, testutil.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Null[string]()}})
				if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted {
					t.Error("failed update lost preparation state or invented acceptance")
				}
				if tc.patch != 0 && (v.Before == nil || !reflect.DeepEqual(v.Before.RawFields, expectedBefore)) {
					t.Error("failed update changed or lost the actual preparation snapshot")
				}
				if tc.read == 403 || tc.patch != 0 {
					var apiError *client.APIError
					expected := tc.read
					if tc.patch != 0 {
						expected = tc.patch
					}
					if !errors.As(e, &apiError) || apiError.StatusCode != expected {
						t.Error("failure lost its HTTP status")
					}
				}
				return testutil.Observation{Response: r}
			}})
		})
	}
}

func TestUnitAppPermissionCatalogReplacementsUseRealPreparationOnDeniedPatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	var expected map[string]json.RawMessage
	if json.Unmarshal(before, &expected) != nil {
		t.Fatal("invalid real catalog preparation")
	}
	for _, field := range []string{"DefaultCatalogApps", "GlobalCatalogApps", "PrivateCatalogApps"} {
		t.Run(field, func(t *testing.T) {
			body := json.RawMessage(`{"Identity":"Global","` + field + `":[{"Identity":null,"Id":"unit-app"}]}`)
			run(t, testutil.Case{Name: "catalog-replacement", Steps: []testutil.Exchange{wire("GET", item, "Set-CsTeamsAppPermissionPolicy", nil, 200, string(before)), wire("PATCH", item, "Set-CsTeamsAppPermissionPolicy", body, 403, "")}, Want: testutil.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				edit := resource.ReplaceCatalogApps(resource.NewCatalogApp("unit-app"))
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
				if v == nil || v.Before == nil || !v.PatchInvoked || v.HTTPAccepted || !reflect.DeepEqual(v.Before.RawFields, expected) {
					t.Error("denied catalog update changed real preparation or invented acceptance")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitAppPermissionCatalogBuildersHaveNoCredentialsOrHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-catalog-builders", Invoke: func(_ context.Context, _ *msteams.Client) testutil.Observation {
		for _, id := range []string{"", " ", "future-app"} {
			v := resource.NewCatalogApp(id)
			out, e := resource.ReplaceCatalogApps(v, v).Apply(nil)
			if e != nil || len(out) != 2 || out[0].ID == nil || *out[0].ID != id {
				return testutil.Observation{Err: errors.New("local catalog construction changed")}
			}
		}
		return testutil.Observation{}
	}})
}
