package teamseducationconfiguration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

const bindingSource = "contracts/endpoints/8.1.0/fixtures/education-configuration-binding.json"

type bindingCapture struct {
	Name, Command string
	Arguments     map[string]json.RawMessage
	Observed      struct {
		Failure       *string
		Sends, Tokens int
		Output        *struct {
			ParameterSet string `json:"parameter_set"`
			Bound        map[string]json.RawMessage
			Specified    bool
		}
	}
}

func TestUnitEducationConfigurationCapturedBindingValues(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../" + bindingSource)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct{ Cases []bindingCapture }
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 17 {
		t.Fatal("incomplete compiled binding fixture")
	}
	var results []contracttest.Result
	for _, f := range fixture.Cases {
		if f.Observed.Sends != 0 || f.Observed.Tokens != 0 {
			t.Fatal("source binding crossed offline boundary")
		}
		bound := f.Arguments
		set := "Identity"
		if f.Observed.Output != nil {
			bound = f.Observed.Output.Bound
			set = f.Observed.Output.ParameterSet
		} else if _, ok := bound["Filter"]; ok {
			set = "Filter"
		}
		var identity, pattern string
		if json.Unmarshal(bound["Identity"], &identity) != nil {
			identity = ""
		}
		if json.Unmarshal(bound["Filter"], &pattern) != nil {
			pattern = ""
		}
		verb, _, _ := strings.Cut(f.Command, "-")
		c := contracttest.Case{Name: "binding-" + strings.ToLower(verb+"-"+f.Name), Source: bindingSource, Mappings: []contracttest.Mapping{{Command: f.Command, ParameterSet: set}}}
		var settings resource.Settings
		controls := policy.Controls{}
		if f.Command == setCommand {
			encoded, _ := json.Marshal(bound)
			if json.Unmarshal(encoded, &settings) != nil {
				t.Fatal("invalid typed bound setting")
			}
			if force, ok := bound["Force"]; ok {
				var value struct{ IsPresent bool }
				if json.Unmarshal(force, &value) != nil {
					t.Fatal("invalid source switch")
				}
				controls.Force = new(value.IsPresent)
			}
		}
		// Force-only and Identity-only fail the captured source predicate.
		invalid := f.Observed.Failure != nil || f.Command == setCommand && (f.Observed.Output == nil || !f.Observed.Output.Specified)
		if invalid {
			c.Want.Error = resource.ErrInput
		} else if f.Command == setCommand {
			body := map[string]json.RawMessage{"Identity": bound["Identity"]}
			for _, name := range []string{"ParentGuardianPreferredContactMethod", "EduGenerativeAIEnhancements", "UpdateParentInformation"} {
				if value, ok := bound[name]; ok {
					body[name] = value
				}
			}
			if controls.Force != nil {
				value := "False"
				if *controls.Force {
					value = "True"
				}
				body["Force"], _ = json.Marshal(value)
			}
			encoded, _ := json.Marshal(body)
			c.Steps = []contracttest.Exchange{compiledWire("PATCH", item, setCommand, encoded, 204, "")}
		} else {
			path, reply := collection, "[]"
			if identity != "" {
				path = item
				reply = `{"Identity":"Global"}`
			}
			c.Steps = []contracttest.Exchange{compiledWire("GET", path, getCommand, nil, 200, reply)}
			c.Want.JSON = json.RawMessage(reply)
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsEducationConfigurations
			if f.Command == setCommand {
				v, r, e := s.PatchByIdentity(ctx, identity, resource.PatchRequest{Settings: settings, Controls: controls})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
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
		}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}

func TestUnitEducationConfigurationCapturedFilters(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../" + bindingSource)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Filters []struct {
			Pattern  string
			Observed struct {
				Identities    []string
				Sends, Tokens int
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Filters) != 5 {
		t.Fatal("incomplete source filters")
	}
	var results []contracttest.Result
	const reply = `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"},{"Identity":" "},{"Identity":"Global/fixture"}]`
	for i, f := range fixture.Filters {
		if f.Observed.Sends != 0 || f.Observed.Tokens != 0 {
			t.Fatal("source filter crossed offline boundary")
		}
		// Original Get emits exact Global first after invoking this helper.
		var want []string
		for _, id := range f.Observed.Identities {
			if id == "Global" {
				want = append(want, id)
			}
		}
		for _, id := range f.Observed.Identities {
			if id != "Global" {
				want = append(want, id)
			}
		}
		set := "Filter"
		if f.Pattern == "" {
			set = "Identity"
		}
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("source-filter-%d", i), Source: bindingSource, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, getCommand, nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsEducationConfigurations
			var values []*resource.Configuration
			var o contracttest.Observation
			if f.Pattern == "" {
				v, r, e := s.List(ctx)
				values = v
				o = contracttest.Observation{Response: r, Err: e}
			} else {
				v, r, e := s.ListFiltered(ctx, f.Pattern)
				values = v
				o = contracttest.Observation{Response: r, Err: e}
			}
			var got []string
			for _, v := range values {
				if v == nil || v.Identity == nil {
					t.Fatal("missing filter identity")
				}
				got = append(got, *v.Identity)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("source filter/order differs")
			}
			return o
		}}))
	}
	evidence(t, results)
}
