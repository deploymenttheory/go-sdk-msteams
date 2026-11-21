package teamschannelspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_channels_policy"
)

func TestUnitChannelsPolicyPinnedCompiledCRUD(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, f := range captures(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Capture
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: x.CmdletName, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsChannelsPolicies
				controls := policy.Controls{}
				if f.ForceMode >= 0 {
					controls.Force = new(f.ForceMode == 1)
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
				}
				var settings resource.Settings
				if e := json.Unmarshal([]byte(f.Input), &settings); e != nil {
					return contracttest.Observation{Err: e}
				}
				if f.Operation == "create" {
					v, r, e := s.CreateByIdentity(ctx, scoped, resource.CreateRequest{Settings: settings, Controls: controls})
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

func TestUnitChannelsPolicyPinnedModernGrant(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../" + grantSource)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Cases []struct {
			Name, Operation, Method string
			URL                     string `json:"generated_url"`
			Status                  int    `json:"response_status"`
			Reply                   string `json:"response_body"`
			Body                    json.RawMessage
			Sends                   int
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 7 {
		t.Fatal("incomplete grant fixture")
	}
	var results []contracttest.Result
	for _, f := range fixture.Cases {
		if f.Sends != 1 {
			t.Fatal("unexpected source send count")
		}
		u, e := url.Parse(f.URL)
		if e != nil {
			t.Fatal(e)
		}
		parts := strings.Split(u.EscapedPath(), "/")
		for i, p := range parts {
			v, e := url.PathUnescape(p)
			if e != nil {
				t.Fatal(e)
			}
			parts[i] = url.PathEscape(v)
		}
		set := "Identity"
		if f.Operation == "TenantGrantPolicy" {
			set = "GrantToTenant"
		}
		if f.Operation == "GroupGrantPolicyAssignment" {
			set = "GrantToGroup"
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			results = append(results, run(t, contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Source: grantSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(), modernWire(f.Method, strings.Join(parts, "/"), f.Body, f.Status, f.Reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsChannelsPolicies
				if f.Operation == "UserGrantPolicy" {
					name := "Fixture"
					if f.Name == "user-clear-empty-policy" {
						name = ""
					}
					r, e := s.AssignToUser(ctx, "fixture/user", name)
					return contracttest.Observation{Response: r, Err: e}
				}
				if f.Operation == "TenantGrantPolicy" {
					r, e := s.AssignTenantDefault(ctx, "Fixture", false)
					return contracttest.Observation{Response: r, Err: e}
				}
				name, priority := "Fixture", new(int32(2))
				if f.Name == "group-assignment-zero-priority" {
					priority = new(int32(0))
				}
				if f.Name == "group-clear-without-priority" {
					name, priority = "", nil
				}
				v, r, e := s.AssignToGroup(ctx, "fixture/group", name, priority)
				if e == nil && (v == nil || v.Accepted != (f.Status == 202)) {
					e = errors.New("modern group acceptance changed")
				}
				return contracttest.Observation{Response: r, Err: e}
			}}))
		}
	}
	evidence(t, results)
}

func TestUnitChannelsPolicyCompiledFilterDoesNotUseModernWhitespaceRules(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	const reply = `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"},{"Identity":" "}]`
	for _, pattern := range []string{"", "Fixture*", " "} {
		name, set, want := "list", "Identity", "Global,global,Tag:Fixture, "
		if pattern == "Fixture*" {
			name, set, want = "filter-tag", "Filter", "Tag:Fixture"
		}
		if pattern == " " {
			name, set, want = "filter-whitespace", "Filter", " "
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, getCommand, nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsChannelsPolicies
			var v []*resource.Policy
			var o contracttest.Observation
			if pattern == "" {
				x, r, e := s.List(ctx)
				v = x
				o = contracttest.Observation{Response: r, Err: e}
			} else {
				x, r, e := s.ListFiltered(ctx, pattern)
				v = x
				o = contracttest.Observation{Response: r, Err: e}
			}
			var ids []string
			for _, x := range v {
				if x == nil || x.Identity == nil {
					ids = append(ids, "missing")
				} else {
					ids = append(ids, *x.Identity)
				}
			}
			if o.Err == nil && strings.Join(ids, ",") != want {
				o.Err = errors.New("compiled filter/global ordering changed")
			}
			return o
		}}))
	}
	evidence(t, results)
}
