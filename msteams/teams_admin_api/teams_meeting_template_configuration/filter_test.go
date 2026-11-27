package teamsmeetingtemplateconfiguration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestUnitConfigurationFiltersUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	for index, tc := range []struct {
		pattern string
		count   int
	}{{"Global", 1}, {"G*", 1}, {"missing*", 0}} {
		run(t, testutil.Case{Name: fmt.Sprintf("filter-%d", index), Steps: []testutil.Exchange{compiledWire("GET", "/Skype.Policy/configurations/TeamsMeetingTemplateConfiguration", "Get-CsTeamsMeetingTemplateConfiguration", nil, 200, string(raw))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations.ListFiltered(ctx, tc.pattern)
			if err == nil && len(values) != tc.count {
				t.Fatal("template configuration wildcard selection differs")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
	}
}
