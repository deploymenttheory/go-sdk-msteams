package sharedchannelmanagement_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	associated "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/associated_team"
	share "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/shared_channel"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/shared_channel_management"
)

func TestUnitResourceBoundariesAreLocal(t *testing.T) {
	t.Parallel()
	for _, resource := range []string{"sharing", "associated", "composition"} {
		for _, state := range []string{"nil-service", "wrong-backend", "closed", "nil-context"} {
			t.Run(resource+"-"+state, func(t *testing.T) {
				want := client.ErrClosed
				if state == "wrong-backend" {
					want = client.ErrDestination
				}
				if state == "nil-context" {
					want = client.ErrNilContext
				}
				//nolint:contextcheck // The nil-context case deliberately replaces the inherited context to verify local rejection.
				run(t, contracttest.Case{Name: resource + "-" + state, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					if state == "closed" {
						if err := sdk.Close(); err != nil {
							return contracttest.Observation{Err: err}
						}
					}
					if state == "nil-context" {
						ctx = nil
					}
					switch resource {
					case "sharing":
						s := sdk.GraphAPI.ChannelShares
						if state == "nil-service" {
							s = nil
						}
						if state == "wrong-backend" {
							s = share.NewService(sdk.TeamsAdminAPI.Scope())
						}
						v, r, err := s.GetByID(ctx, hostID, channelID, targetID)
						return contracttest.Observation{Value: v, Response: r, Err: err}
					case "associated":
						s := sdk.GraphAPI.AssociatedTeams
						if state == "nil-service" {
							s = nil
						}
						if state == "wrong-backend" {
							s = associated.NewService(sdk.TeamsAdminAPI.Scope())
						}
						v, err := s.ListForUser(ctx, userID, client.PagingOptions{})
						return contracttest.Observation{Response: v.Response, Err: err}
					default:
						s := sdk.GraphAPI.SharedChannelManagement
						if state == "nil-service" {
							s = nil
						}
						if state == "wrong-backend" {
							s = management.NewService(sdk.TeamsAdminAPI.Scope())
						}
						r, err := s.UnshareWithTeam(ctx, hostID, channelID, targetID)
						return contracttest.Observation{Response: r, Err: err}
					}
				}, Want: contracttest.Expected{Error: want}})
			})
		}
	}
}
