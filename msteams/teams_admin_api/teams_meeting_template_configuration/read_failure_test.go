package teamsmeetingtemplateconfiguration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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
				path := "/Skype.Policy/configurations/TeamsMeetingTemplateConfiguration"
				set := "Filter"
				if !list {
					path += "/configuration/Global"
					set = "Identity"
				}
				run(t, contracttest.Case{Name: fmt.Sprintf("read-error-%s-%d-%t", mode, status, list), Mode: mode, Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsMeetingTemplateConfiguration", ParameterSet: set}}, Steps: []contracttest.Exchange{compiledWire("GET", path, "Get-CsTeamsMeetingTemplateConfiguration", nil, status, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					if list {
						_, r, e := sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations.ListFiltered(ctx, "Global")
						return contracttest.Observation{Response: r, Err: e}
					}
					_, r, e := sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations.GetByIdentity(ctx, "Global")
					return contracttest.Observation{Response: r, Err: e}
				}})
			}
		}
		for index, body := range []string{`{"Description":{}}`, `[]`, `{`} {
			run(t, contracttest.Case{Name: fmt.Sprintf("read-decode-%s-%d", mode, index), Mode: mode, Steps: []contracttest.Exchange{compiledWire("GET", "/Skype.Policy/configurations/TeamsMeetingTemplateConfiguration/configuration/Global", "Get-CsTeamsMeetingTemplateConfiguration", nil, 200, body)}, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations.GetByIdentity(ctx, "Global")
				return contracttest.Observation{Response: r, Err: e}
			}})
		}
	}
}
