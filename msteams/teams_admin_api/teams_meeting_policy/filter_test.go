package teamsmeetingpolicy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestUnitFiltersUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	body := readResponseFixture(t, "owned_list.json")
	for index, tc := range []struct {
		pattern string
		count   int
	}{{"Global", 1}, {"fixture-polic?", 1}, {"Tag:fixture-polic[y]", 1}, {"missing-policy", 0}} {
		run(t, contracttest.Case{Name: fmt.Sprintf("filter-%d", index), Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsMeetingPolicy", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{compiledWire("GET", "/Skype.Policy/configurations/TeamsMeetingPolicy", "Get-CsTeamsMeetingPolicy", nil, 200, string(body))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, r, e := sdk.TeamsAdminAPI.TeamsMeetingPolicies.ListFiltered(ctx, tc.pattern)
			if e == nil && len(values) != tc.count {
				t.Fatalf("filter case %d selected %d identities; want %d", index, len(values), tc.count)
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
