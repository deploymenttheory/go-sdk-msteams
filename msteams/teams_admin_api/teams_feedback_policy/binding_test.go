package teamsfeedbackpolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

func TestUnitFeedbackPolicyCapturedCompiledBindings(t *testing.T) {
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
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 36 {
		t.Fatal("incomplete compiled feedback binding fixture")
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
		if json.Unmarshal(bound["Identity"], &identity) != nil {
			identity = ""
		}
		if json.Unmarshal(bound["Filter"], &pattern) != nil {
			pattern = ""
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
			case "New-CsTeamsFeedbackPolicy", setCommand, "Remove-CsTeamsFeedbackPolicy":
				method, path, reply, status = "POST", collection, "", 204
				fields := map[string]json.RawMessage{}
				if f.Command != "Remove-CsTeamsFeedbackPolicy" {
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
				if f.Command == "Remove-CsTeamsFeedbackPolicy" {
					method, path = "DELETE", collection+"/configuration/Fixture"
				}
			default:
				t.Fatal("unknown source command")
			}
			c.Steps = []contracttest.Exchange{compiledWire(method, path, f.Command, body, status, reply)}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsFeedbackPolicies
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
			case "New-CsTeamsFeedbackPolicy":
				v, r, e := s.CreateByIdentity(ctx, identity, resource.CreateRequest{Settings: input, Controls: controls})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case setCommand:
				v, r, e := s.PatchByIdentity(ctx, identity, resource.PatchRequest{Settings: input, Controls: controls})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case "Remove-CsTeamsFeedbackPolicy":
				v, r, e := s.DeleteByIdentity(ctx, identity, controls)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			return contracttest.Observation{Err: errors.New("unknown source command")}
		}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}
