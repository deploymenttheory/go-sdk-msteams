package teamsfeedbackpolicy_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

const bindingSource = "contracts/endpoints/8.1.0/fixtures/feedback-policy-binding.json"

func TestFeedbackPolicyFiltersMatchCapturedSource(t *testing.T) {
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
		t.Fatal("incomplete feedback filter fixture")
	}
	var results []contracttest.Result
	for i, f := range fixture.Filters {
		if f.Observed.Sends != 0 || f.Observed.Tokens != 0 {
			t.Fatal("source filter crossed offline boundary")
		}
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
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("source-filter-%d", i), Source: bindingSource, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, getCommand, nil, 200, `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"},{"Identity":" "},{"Identity":"Global/fixture"}]`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsFeedbackPolicies
			var o contracttest.Observation
			var got []string
			if f.Pattern == "" {
				v, r, e := s.List(ctx)
				o = contracttest.Observation{Response: r, Err: e}
				for _, p := range v {
					if p == nil || p.Identity == nil {
						t.Fatal("missing filter identity")
					}
					got = append(got, *p.Identity)
				}
			} else {
				v, r, e := s.ListFiltered(ctx, f.Pattern)
				o = contracttest.Observation{Response: r, Err: e}
				for _, p := range v {
					if p == nil || p.Identity == nil {
						t.Fatal("missing filter identity")
					}
					got = append(got, *p.Identity)
				}
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("source filter/order differs")
			}
			return o
		}}))
	}
	evidence(t, results)
}
