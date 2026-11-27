package teamstemplatepermissionpolicy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Empty failures and deliberately malformed replies are fault injection. Successful
// bodies elsewhere in this package are the permanent acceptance responses.
func TestUnitReadsRetainAuthorizationAndDecodeFailures(t *testing.T) {
	t.Parallel()
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, status := range []int{401, 403, 404} {
			for _, list := range []bool{false, true} {
				path := "/Skype.Policy/configurations/TeamsTemplatePermissionPolicy"

				if !list {
					path += "/configuration/Global"

				}
				run(t, testutil.Case{Name: fmt.Sprintf("read-error-%s-%d-%t", mode, status, list), Mode: mode, Steps: []testutil.Exchange{compiledWire("GET", path, "Get-CsTeamsTemplatePermissionPolicy", nil, status, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					if list {
						_, r, e := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies.ListFiltered(ctx, "Global")
						return testutil.Observation{Response: r, Err: e}
					}
					_, r, e := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies.GetByIdentity(ctx, "Global")
					return testutil.Observation{Response: r, Err: e}
				}})
			}
		}
		for index, body := range []string{`{"Description":false}`, `[]`, `{`} {
			run(t, testutil.Case{Name: fmt.Sprintf("read-decode-%s-%d", mode, index), Mode: mode, Steps: []testutil.Exchange{compiledWire("GET", "/Skype.Policy/configurations/TeamsTemplatePermissionPolicy/configuration/Global", "Get-CsTeamsTemplatePermissionPolicy", nil, 200, body)}, Want: testutil.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies.GetByIdentity(ctx, "Global")
				return testutil.Observation{Response: r, Err: e}
			}})
		}
	}
}
