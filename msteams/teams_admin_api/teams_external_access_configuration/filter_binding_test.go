package teamsexternalaccessconfiguration_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_external_access_configuration"
)

// The effective pinned public Filter parameter inherits ValidateNotNullOrEmpty.
// BaseTpmGetCmdlet<T>.FilterPolicies' empty branch handles an omitted filter;
// it does not authorize binding an explicitly empty Filter. Source metadata/IL
// and the distinction from List are recorded in pol90-current-test-audit.md.
func TestUnitExplicitEmptyFilterRejectsBeforeCredentials(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "bound-empty-filter", Source: "msteams/teams_admin_api/teams_external_access_configuration/filter_binding_test.go", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsExternalAccessConfiguration", ParameterSet: "Filter"}}, Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.ListFiltered(ctx, "")
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}})
}
