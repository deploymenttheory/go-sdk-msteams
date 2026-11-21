package teamsenhancedencryptionpolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_enhanced_encryption_policy"
)

func fixtureSettings() resource.Settings {
	return resource.Settings{Description: new(""), CallingEndToEndEncryptionEnabledType: new("FutureCalling"), MeetingEndToEndEncryption: new("FutureMeeting")}
}
func TestEnhancedEncryptionPinnedGeneratedRequests(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, f := range captures(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			command, set := getCommand, "Identity"
			switch f.Operation {
			case "ConfigurationCreateConfiguration":
				command = newCommand
			case "ConfigurationUpdateConfiguration":
				command = setCommand
			case "ConfigurationDeleteConfiguration":
				command = deleteCommand
			case "UserGrantPolicy":
				command = grantCommand
			case "TenantGrantPolicy":
				command, set = grantCommand, "GrantToTenant"
			case "GroupGrantPolicyAssignment":
				command, set = grantCommand, "GrantToGroup"
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}}
			if !f.Canceled {
				c.Steps = []contracttest.Exchange{discovery(), capturedWire(t, f, command)}
			} else {
				c.Want.Error = context.Canceled
			}
			if f.Method == "GET" && f.Status == 200 && !f.Canceled {
				c.Want.JSON = json.RawMessage(f.Reply)
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies
				if f.Canceled {
					canceled, cancel := context.WithCancel(ctx)
					cancel()
					v, r, e := s.List(canceled)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				var o contracttest.Observation
				targetID, settings := identity, fixtureSettings()
				if f.Name == "policy-identity-only-put" {
					targetID, settings = "Global", resource.Settings{}
				}
				if f.Name == "policy-empty-string-put" {
					targetID = "Global"
					settings = resource.Settings{Description: new(""), CallingEndToEndEncryptionEnabledType: new(""), MeetingEndToEndEncryption: new("")}
				}
				switch f.Operation {
				case "ConfigurationGetConfigurations":
					v, r, e := s.List(ctx)
					o = contracttest.Observation{Value: v, Response: r, Err: e}
				case "ConfigurationGetConfiguration":
					id := identity
					if f.Status >= 400 {
						id = "Fixture"
					}
					v, r, e := s.GetByIdentity(ctx, id)
					o = contracttest.Observation{Value: v, Response: r, Err: e}
				case "ConfigurationCreateConfiguration":
					r, e := s.CreateByIdentity(ctx, identity, fixtureSettings())
					o = contracttest.Observation{Response: r, Err: e}
				case "ConfigurationUpdateConfiguration":
					r, e := s.UpdateByIdentity(ctx, targetID, settings)
					o = contracttest.Observation{Response: r, Err: e}
				case "ConfigurationDeleteConfiguration":
					r, e := s.DeleteByIdentity(ctx, identity)
					o = contracttest.Observation{Response: r, Err: e}
				case "UserGrantPolicy":
					name := "Fixture"
					if f.Name == "user-clear-empty-policy" {
						name = ""
					}
					r, e := s.AssignToUser(ctx, "fixture/user", name)
					o = contracttest.Observation{Response: r, Err: e}
				case "TenantGrantPolicy":
					r, e := s.AssignTenantDefault(ctx, "Fixture", false)
					o = contracttest.Observation{Response: r, Err: e}
				case "GroupGrantPolicyAssignment":
					name, priority := "Fixture", new(int32(2))
					if f.Name == "group-assignment-zero-priority" {
						priority = new(int32(0))
					}
					if f.Name == "group-clear-without-priority" {
						name, priority = "", nil
					}
					v, r, e := s.AssignToGroup(ctx, "fixture/group", name, priority)
					if e == nil && (v == nil || v.Accepted != (f.Status == 202)) {
						e = errors.New("group acceptance status changed")
					}
					o = contracttest.Observation{Response: r, Err: e}
				default:
					o.Err = errors.New("unknown pinned operation")
				}
				if f.Status >= 400 {
					var api *client.APIError
					if !errors.As(o.Err, &api) || api.StatusCode != f.Status || strings.Contains(o.Err.Error(), "Fixture message") {
						return contracttest.Observation{Err: errors.New("generated error status/redaction changed")}
					}
					o.Err = nil
				}
				return o
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}

func TestEnhancedEncryptionModernFilterAndWhitespace(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	const reply = `[{"Identity":"Other"},{"Identity":"gLoBaL"},{"Identity":"Global"},{"Identity":"Tag:Global"},{"Identity":"Fixture"}]`
	for _, pattern := range []string{"", "*", "Global", " "} {
		want := "gLoBaL,Global,Other,Tag:Global,Fixture"
		name, set := "list", "Identity"
		if pattern != "" {
			set = "Filter"
			switch pattern {
			case "*":
				name = "filter-all"
			case "Global":
				name = "filter-tag"
			case " ":
				name = "filter-whitespace"
			}
		}
		if pattern == "Global" {
			want = "gLoBaL,Global,Tag:Global"
		}
		results = append(results, run(t, contracttest.Case{Name: name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(), wire("GET", collection, getCommand, nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies
			v, r, e := s.ListFiltered(ctx, pattern)
			var ids []string
			for _, x := range v {
				ids = append(ids, x.Identity)
			}
			if e == nil && strings.Join(ids, ",") != want {
				e = errors.New("modern filter/global-first behavior changed")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
