package teamscallholdpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitCallHoldGroupGrantDenialsNeverReportAcceptance(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, contracttest.Case{Name: "group-denial", Source: "msteams/teams_admin_api/teams_call_hold_policy/acceptance_test.go", Steps: []contracttest.Exchange{discovery(t), modernWire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsCallHoldPolicy", json.RawMessage(`{"policyName":"Fixture"}`), status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsCallHoldPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
				var apiError *client.APIError
				if !errors.As(e, &apiError) || apiError.StatusCode != status || v != nil && v.Accepted {
					t.Error("denied group grant lost its HTTP failure or invented acceptance")
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}
