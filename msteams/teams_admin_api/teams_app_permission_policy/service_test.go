package teamsapppermissionpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

func TestAppPermissionPinnedProviderRequests(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, f := range captures(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Capture
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			set := "Identity"
			switch f.Operation {
			case "user":
				set = "GrantToUser"
			case "tenant":
				set = "GrantToTenant"
			case "group":
				set = "GrantToGroup"
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: x.CmdletName, ParameterSet: set}}, Steps: []contracttest.Exchange{wire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies
				controls := policy.Controls{}
				if f.Force {
					controls.Force = new(false)
				}
				switch f.Operation {
				case "read":
					v, r, e := s.GetByIdentity(ctx, scoped)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "list":
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "delete":
					v, r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "user":
					v, r, e := s.AssignUser(ctx, "user/fixture", f.PolicyName)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "tenant":
					v, r, e := s.AssignTenantDefault(ctx, f.PolicyName, false)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "group":
					v, r, e := s.AssignGroup(ctx, "group/fixture", f.PolicyName, -1)
					return contracttest.Observation{Value: v, Response: r, Err: e}
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
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}

func TestAppPermissionFilteringKeepsGlobalAndLiteralWhitespace(t *testing.T) {
	t.Parallel()
	const reply = `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"},{"Identity":" "}]`
	var results []contracttest.Result
	for _, pattern := range []string{"", "Fixture*", " "} {
		name, set, want := "list", "Identity", "Global,global,Tag:Fixture, "
		if pattern == "Fixture*" {
			name, set, want = "filter-tag", "Filter", "Tag:Fixture"
		}
		if pattern == " " {
			name, set, want = "filter-whitespace", "Filter", " "
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsAppPermissionPolicy", ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsTeamsAppPermissionPolicy", nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			var observation contracttest.Observation
			if pattern == "" {
				v, r, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.List(ctx)
				observation = contracttest.Observation{Value: v, Response: r, Err: e}
			} else {
				v, r, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.ListFiltered(ctx, pattern)
				observation = contracttest.Observation{Value: v, Response: r, Err: e}
			}
			values, _ := observation.Value.([]*resource.Policy)
			var ids []string
			for _, v := range values {
				if v == nil || v.Identity == nil {
					ids = append(ids, "missing")
				} else {
					ids = append(ids, *v.Identity)
				}
			}
			if observation.Err == nil && strings.Join(ids, ",") != want {
				observation.Err = errors.New("app permission filter/order changed")
			}
			return observation
		}}))
	}
	evidence(t, results)
}
