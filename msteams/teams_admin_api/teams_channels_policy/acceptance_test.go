package teamschannelspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitChannelsPolicyGroupGrantDenialsNeverReportAcceptance(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, testutil.Case{Name: "group-denial", Steps: []testutil.Exchange{discovery(t), modernWire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsChannelsPolicy", json.RawMessage(`{"policyName":"Fixture"}`), status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsChannelsPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
				var apiError *client.APIError
				if !errors.As(e, &apiError) || apiError.StatusCode != status || v != nil && v.Accepted {
					t.Error("denied group grant lost its HTTP failure or invented acceptance")
				}
				return testutil.Observation{Response: r}
			}})
		})
	}
}
