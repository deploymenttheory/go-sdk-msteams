package teamschannelspolicy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestUnitChannelsPolicyDiscoveryCannotEstablishAssignmentAcceptance(t *testing.T) {
	t.Parallel()

	for _, reply := range []string{`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid"}}`, "not-json"} {
		step := discovery(t)
		step.Reply.Status = 202
		step.Reply.Body = []byte(reply)
		name := "discovery-202"
		if reply == "not-json" {
			name += "-malformed"
		}
		run(t, testutil.Case{Name: name, Steps: []testutil.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsChannelsPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
			if e == nil || v != nil && v.Accepted {
				return testutil.Observation{Err: errors.New("discovery response falsely reported assignment acceptance")}
			}
			return testutil.Observation{Response: r}
		}})
	}
}
