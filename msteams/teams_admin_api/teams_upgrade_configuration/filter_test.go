package teamsupgradeconfiguration_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestUnitFilterUsesActualCollection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"global", "Glo?al", 1}, {"missing", "missing", 0}, {"literal-whitespace", " ", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsUpgradeConfiguration", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, "Get-CsTeamsUpgradeConfiguration", nil, 200, string(readResponseFixture(t, "list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				values, r, e := sdk.TeamsAdminAPI.TeamsUpgradeConfigurations.ListFiltered(ctx, tc.pattern)
				if e == nil && len(values) != tc.count {
					t.Fatal("local filter result differs")
				}
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}
