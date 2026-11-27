package teamsapppermissionpolicy_test

import (
	"context"
	"encoding/json"

	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

func TestUnitAppPermissionTypedProviderRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Request
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}

			switch f.Operation {
			case "user":

			case "tenant":

			case "group":

			}
			c := testutil.Case{Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{wire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			if len(c.Steps) != 0 {
				c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies
				controls := policy.Controls{}
				if f.Force {
					controls.Force = new(false)
				}
				switch f.Operation {
				case "read":
					v, r, e := s.GetByIdentity(ctx, scoped)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "list":
					v, r, e := s.List(ctx)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "delete":
					v, r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "user":
					v, r, e := s.AssignUser(ctx, "user/fixture", f.PolicyName)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "tenant":
					v, r, e := s.AssignTenantDefault(ctx, f.PolicyName, false)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "group":
					v, r, e := s.AssignGroup(ctx, "group/fixture", f.PolicyName, -1)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				settings := resource.Settings{}
				input := resource.CreateRequest{Controls: controls}
				if f.Fields == "all" || f.Fields == "description" {
					scalars := resource.ScalarSettings{Description: policy.Set(""), DefaultCatalogAppsType: policy.Set("FutureDefaultMode"), GlobalCatalogAppsType: policy.Set(""), PrivateCatalogAppsType: policy.Null[string]()}
					settings.ScalarSettings = scalars
					input.ScalarSettings = scalars
				}
				if f.Fields == "all" || f.Fields == "empty-lists" {
					var a, b, c []*resource.CatalogApp
					if f.Fields == "all" {
						a = []*resource.CatalogApp{resource.NewCatalogApp("fixture-default")}
						b = []*resource.CatalogApp{resource.NewCatalogApp("fixture-global")}
						c = []*resource.CatalogApp{resource.NewCatalogApp("fixture-private")}
					}
					settings.DefaultCatalogApps = policy.Set(a)
					settings.GlobalCatalogApps = policy.Set(b)
					settings.PrivateCatalogApps = policy.Set(c)
					input.DefaultCatalogApps = resource.ReplaceCatalogApps(a...)
					input.GlobalCatalogApps = resource.ReplaceCatalogApps(b...)
					input.PrivateCatalogApps = resource.ReplaceCatalogApps(c...)
				}
				if f.Operation == "create" {
					v, r, e := s.CreateByIdentity(ctx, scoped, input)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}

func TestUnitAppPermissionFiltersUseRealResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	var rows []json.RawMessage
	if json.Unmarshal(raw, &rows) != nil {
		t.Fatal("invalid real list fixture")
	}
	identities := func(row json.RawMessage) string {
		var value struct {
			Identity string `json:"Identity"`
		}
		if json.Unmarshal(row, &value) != nil {
			t.Fatal("invalid real list identity")
		}
		return value.Identity
	}
	slices.SortStableFunc(rows, func(a, b json.RawMessage) int {
		left, right := strings.EqualFold(identities(a), "Global"), strings.EqualFold(identities(b), "Global")
		if left == right {
			return 0
		}
		if left {
			return -1
		}
		return 1
	})
	cases := []struct {
		Name, Pattern, OnlyID string
		Empty                 bool
	}{
		{Name: "list"}, {Name: "all", Pattern: "*"}, {Name: "global", Pattern: "Global", OnlyID: "Global"},
		{Name: "missing", Pattern: "__sdk_missing_fixture__", Empty: true},
		{Name: "whitespace", Pattern: " ", Empty: true},
	}
	for _, row := range rows {
		id := identities(row)
		if strings.HasPrefix(id, "Tag:") && !strings.ContainsAny(id, "*?[]`") {
			cases = append(cases, struct {
				Name, Pattern, OnlyID string
				Empty                 bool
			}{Name: "tag-prefix", Pattern: strings.TrimPrefix(id, "Tag:"), OnlyID: id})
			break
		}
	}
	for _, tc := range cases {
		want := make([]json.RawMessage, 0, len(rows))
		for _, row := range rows {
			if !tc.Empty && (tc.OnlyID == "" || strings.EqualFold(identities(row), tc.OnlyID)) {
				want = append(want, row)
			}
		}
		expected, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		result, err := testutil.Run(t.Context(), testutil.Case{
			Name:  "real-filter-" + tc.Name,
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsAppPermissionPolicy", 200, raw), Want: testutil.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				service := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies
				if tc.Pattern == "" {
					value, response, err := service.List(ctx)
					return testutil.Observation{Value: value, Response: response, Err: err}
				}
				value, response, err := service.ListFiltered(ctx, tc.Pattern)
				return testutil.Observation{Value: value, Response: response, Err: err}
			},
		})
		if err != nil || !result.Passed() {
			t.Fatalf("real filtering response differs: %v %+v", err, result.Differences())
		}
	}

}
