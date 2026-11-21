package teamscallparkpolicy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestUnitCallParkDiscoveryCannotEstablishAssignmentAcceptance(t *testing.T) {
	t.Parallel()

	for _, reply := range []string{`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid"}}`, "not-json"} {
		step := discovery(t)
		step.Reply.Status = 202
		step.Reply.Body = []byte(reply)
		name := "discovery-202"
		if reply == "not-json" {
			name += "-malformed"
		}
		run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsCallParkPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
			if e == nil || v != nil && v.Accepted {
				return contracttest.Observation{Err: errors.New("discovery response falsely reported assignment acceptance")}
			}
			return contracttest.Observation{Response: r}
		}})
	}
}
