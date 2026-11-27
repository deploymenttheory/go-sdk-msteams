package teamsshiftspolicy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestUnitFiltersUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	body := readResponseFixture(t, "owned_list.json")
	for index, tc := range []struct {
		pattern string
		count   int
	}{{"Global", 1}, {"fixture-polic?", 1}, {"Tag:fixture-polic[y]", 1}, {"missing-policy", 0}} {
		run(t, testutil.Case{Name: fmt.Sprintf("filter-%d", index), Steps: []testutil.Exchange{compiledWire("GET", "/Skype.Policy/configurations/TeamsShiftsPolicy", "Get-CsTeamsShiftsPolicy", nil, 200, string(body))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			values, r, e := sdk.TeamsAdminAPI.TeamsShiftsPolicies.ListFiltered(ctx, tc.pattern)
			if e == nil && len(values) != tc.count {
				t.Fatalf("filter case %d selected %d identities; want %d", index, len(values), tc.count)
			}
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}
