package teamscustombannertext_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
)

func TestBannerPinnedProviderAndSelectors(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, f := range captures(t) {
		if f.Name == "create-default-id" {
			continue
		} // Fresh UUID protocol checked independently.
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Observed.Request
			set := "Identity"
			if f.Selector == "relative" {
				set = "ParentAndRelativeKey"
			}
			command := x.CmdletName
			if x.Sends == 0 {
				command = setCommand
				if f.Operation == "create" {
					command = newCommand
				}
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}}
			if x.Sends != 0 {
				var body json.RawMessage
				if x.Body != "" {
					body = json.RawMessage(x.Body)
				}
				c.Steps = []contracttest.Exchange{wire(x.Method, strings.TrimPrefix(x.Url, origin), command, body, f.Status, f.Response)}
			}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if x.Sends == 0 {
					c.Want.Error = resource.ErrInput
				}
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Operation == "read" || f.Operation == "list" {
				if f.Response == "" {
					c.Want.JSON = json.RawMessage("null")
				} else {
					c.Want.JSON = json.RawMessage(f.Response)
				}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsCustomBannerTexts
				controls := policy.Controls{}
				if f.ForceMode >= 0 {
					controls.Force = new(f.ForceMode == 1)
				}
				var settings resource.Settings
				if json.Unmarshal([]byte(f.Input), &settings) != nil {
					return contracttest.Observation{Err: resource.ErrInput}
				}
				switch f.Operation {
				case "read":
					if f.Selector == "relative" {
						v, r, e := s.GetByID(ctx, f.ID)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.GetByIdentity(ctx, f.Identity)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "list":
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "create":
					text, _ := settings.Text.Value()
					input := resource.CreateRequest{Text: text, Description: settings.Description, Controls: controls}
					if f.Selector == "relative" {
						v, r, e := s.CreateByID(ctx, f.ID, input)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.CreateByIdentity(ctx, f.Identity, input)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "patch":
					input := resource.PatchRequest{Settings: settings, Controls: controls}
					if f.Selector == "relative" {
						input.BodyID = policy.Null[string]()
						if f.ReadID != "" {
							input.BodyID = policy.Set(f.ReadID)
						}
						v, r, e := s.PatchByID(ctx, f.ID, input)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					input.BodyIdentity = policy.Set("read-model-identity")
					v, r, e := s.PatchByIdentity(ctx, f.Identity, input)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "delete":
					if f.Selector == "relative" {
						v, r, e := s.DeleteByID(ctx, f.ID, controls)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.DeleteByIdentity(ctx, f.Identity, controls)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				return contracttest.Observation{Err: resource.ErrInput}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}
func TestBannerFiltersMatchCapturedSource(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct{ name, pattern, want string }{{"list", "", "Global,global,Tag:Fixture, ,Global/fixture"}, {"filter-tag", "Fixture*", "Tag:Fixture"}, {"filter-space", " ", " "}, {"filter-global", "Global", "Global,global"}, {"filter-scoped", "Global/*", "Global/fixture"}} {
		set := "Filter"
		if c.pattern == "" {
			set = "Identity"
		}
		results = append(results, run(t, contracttest.Case{Name: c.name, Source: binderSource, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", collection, getCommand, nil, 200, `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"},{"Identity":" "},{"Identity":"Global/fixture"}]`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			var values []*resource.Banner
			var o contracttest.Observation
			if c.pattern == "" {
				v, r, e := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.List(ctx)
				values = v
				o = contracttest.Observation{Response: r, Err: e}
			} else {
				v, r, e := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.ListFiltered(ctx, c.pattern)
				values = v
				o = contracttest.Observation{Response: r, Err: e}
			}
			ids := []string{}
			for _, v := range values {
				if v != nil && v.Identity != nil {
					ids = append(ids, *v.Identity)
				}
			}
			if strings.Join(ids, ",") != c.want {
				o.Err = resource.ErrInput
			}
			return o
		}}))
	}
	evidence(t, results)
}
