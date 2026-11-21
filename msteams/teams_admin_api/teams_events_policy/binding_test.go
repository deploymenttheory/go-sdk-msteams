package teamseventspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_events_policy"
)

func TestUnitEventsPolicyCapturedCompiledBindings(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../" + bindingSource)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Cases []struct {
			Command, Name string
			Arguments     map[string]json.RawMessage
			Observed      struct {
				Failure       *string
				Sends, Tokens int
				Output        *struct {
					Set       string `json:"parameter_set"`
					Bound     map[string]json.RawMessage
					Specified *bool
					NoScope   bool `json:"no_scope_result"`
				}
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 121 {
		t.Fatal("incomplete compiled events binding fixture")
	}
	var results []contracttest.Result
	for _, f := range fixture.Cases {
		if f.Observed.Sends != 0 || f.Observed.Tokens != 0 {
			t.Fatal("source binding crossed offline boundary")
		}
		bound, set := f.Arguments, "Identity"
		if f.Observed.Output != nil {
			bound = f.Observed.Output.Bound
			set = f.Observed.Output.Set
		} else if _, ok := bound["Filter"]; ok {
			set = "Filter"
		}
		encoded, e := json.Marshal(bound)
		if e != nil {
			t.Fatal(e)
		}
		var input resource.Settings
		if json.Unmarshal(encoded, &input) != nil {
			t.Fatal("source bound values are not typed settings")
		}
		var identity, pattern string
		var name *string
		if json.Unmarshal(bound["Identity"], &identity) != nil {
			identity = ""
		}
		if json.Unmarshal(bound["Filter"], &pattern) != nil {
			pattern = ""
		}
		if value, ok := bound["PolicyName"]; ok && json.Unmarshal(value, &name) != nil {
			t.Fatal("invalid source policy name")
		}
		controls := policy.Controls{}
		if value, ok := bound["Force"]; ok {
			var v struct{ IsPresent bool }
			if json.Unmarshal(value, &v) != nil {
				t.Fatal("invalid source switch")
			}
			controls.Force = new(v.IsPresent)
		}
		c := contracttest.Case{Name: strings.ToLower(f.Command + "-" + f.Name), Source: bindingSource, Mappings: []contracttest.Mapping{{Command: f.Command, ParameterSet: set}}}
		invalid := f.Observed.Failure != nil || f.Command == setCommand && (f.Observed.Output == nil || f.Observed.Output.Specified == nil || !*f.Observed.Output.Specified)
		if invalid {
			c.Want.Error = resource.ErrInput
		} else {
			method, path, reply, status := "GET", collection, "[]", 200
			var body json.RawMessage
			switch f.Command {
			case getCommand:
				if identity != "" {
					path = collection + "/configuration/Fixture"
					reply = `{"Identity":"Fixture"}`
				}
				c.Want.JSON = json.RawMessage(reply)
			case "New-CsTeamsEventsPolicy", setCommand, "Remove-CsTeamsEventsPolicy":
				method, path, reply, status = "POST", collection, "", 204
				fields := map[string]json.RawMessage{}
				if f.Command != "Remove-CsTeamsEventsPolicy" {
					settings, _ := json.Marshal(input)
					if json.Unmarshal(settings, &fields) != nil {
						t.Fatal("invalid settings projection")
					}
				}
				fields["Identity"], _ = json.Marshal(identity)
				if controls.Force != nil {
					value := "False"
					if *controls.Force {
						value = "True"
					}
					fields["Force"], _ = json.Marshal(value)
				}
				body, _ = json.Marshal(fields)
				if f.Command == setCommand {
					method, path = "PATCH", collection+"/configuration/Fixture"
				}
				if f.Command == "Remove-CsTeamsEventsPolicy" {
					method, path = "DELETE", collection+"/configuration/Fixture"
				}
			case grantCommand:
				reply, status = "", 204
				var fields map[string]any
				switch set {
				case "GrantToUser":
					method, path = "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/TeamsEventsPolicy"
					fields = map[string]any{"policyName": name, "additionalParameters": nil}
				case "GrantToTenant":
					method, path = "PATCH", "/Skype.Policy/tenants/policies/TeamsEventsPolicy"
					fields = map[string]any{"PolicyName": name, "additionalParameters": nil, "forceSwitchPresent": controls.Force != nil && *controls.Force}
				case "GrantToGroup":
					var rank int32
					if json.Unmarshal(bound["Rank"], &rank) != nil {
						t.Fatal("required source Rank missing")
					}
					method, path = "POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsEventsPolicy"
					fields = map[string]any{"policyName": name, "priority": rank}
				default:
					t.Fatal("successful no-scope source grant")
				}
				body, _ = json.Marshal(fields)
			default:
				t.Fatal("unknown source command")
			}
			c.Steps = []contracttest.Exchange{compiledWire(method, path, f.Command, body, status, reply)}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsEventsPolicies
			switch f.Command {
			case getCommand:
				if _, ok := bound["Filter"]; ok {
					v, r, e := s.ListFiltered(ctx, pattern)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				if _, ok := bound["Identity"]; ok {
					v, r, e := s.GetByIdentity(ctx, identity)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.List(ctx)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case "New-CsTeamsEventsPolicy":
				v, r, e := s.CreateByIdentity(ctx, identity, resource.CreateRequest{Settings: input, Controls: controls})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case setCommand:
				v, r, e := s.PatchByIdentity(ctx, identity, resource.PatchRequest{Settings: input, Controls: controls})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case "Remove-CsTeamsEventsPolicy":
				v, r, e := s.DeleteByIdentity(ctx, identity, controls)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case grantCommand:
				switch set {
				case "GrantToUser":
					v, r, e := s.AssignUser(ctx, identity, name)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "GrantToTenant":
					v, r, e := s.AssignTenantDefault(ctx, name, controls.Force != nil && *controls.Force)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "GrantToGroup":
					var rank int32
					var id string
					if json.Unmarshal(bound["Rank"], &rank) != nil || json.Unmarshal(bound["Group"], &id) != nil {
						return contracttest.Observation{Err: errors.New("invalid typed source grant")}
					}
					v, r, e := s.AssignGroup(ctx, id, name, rank)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				default:
					v, r, e := s.Assign(ctx, policy.Assignment{})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
			}
			return contracttest.Observation{Err: errors.New("unknown source command")}
		}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}
