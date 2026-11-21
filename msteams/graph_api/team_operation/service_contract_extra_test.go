package teamoperation_test

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	operation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_operation"
)

func TestUnitConfiguredGraphCloudAndModeOwnTheBetaRequest(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcch", "china"} {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			t.Run(cloud+"/"+string(mode), func(t *testing.T) {
				env, err := auth.ResolveEnvironment(cloud, nil)
				if err != nil {
					t.Fatal(err)
				}
				step := wire(200, body("succeeded"))
				step.Request.URL = env.Endpoints.Graph + path()
				step.Request.Audience = env.Endpoints.Graph
				run(t, contracttest.Case{Name: "configured-cloud", Cloud: cloud, Mode: mode, Steps: []contracttest.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					v, r, e := sdk.GraphAPI.TeamOperations.GetByID(ctx, teamID, operationID)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}, Want: contracttest.Expected{Status: 200}})
			})
		}
	}
}

func TestUnitClosedWrongBackendAndInvalidTeamAreLocal(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"closed", "wrong-backend", "nil-service", "invalid-team"} {
		t.Run(name, func(t *testing.T) {
			expected := client.ErrClosed
			if name == "wrong-backend" {
				expected = client.ErrDestination
			}
			if name == "invalid-team" {
				expected = client.ErrRequest
			}
			run(t, contracttest.Case{Name: name, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.GraphAPI.TeamOperations
				id := teamID
				switch name {
				case "closed":
					if e := sdk.Close(); e != nil {
						return contracttest.Observation{Err: e}
					}
				case "wrong-backend":
					s = operation.NewService(sdk.TeamsAdminAPI.Scope())
				case "nil-service":
					s = nil
				case "invalid-team":
					id = "../team"
				}
				v, r, e := s.GetByID(ctx, id, operationID)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}, Want: contracttest.Expected{Error: expected}})
		})
	}
}

func TestUnitPollingRejectsIdentityChangeAndContradictoryAcceptance(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"different-id", `{"id":"other","status":"succeeded"}`, 200},
		{"accepted-success", body("succeeded"), 202},
		{"missing-status", `{"id":"` + operationID + `"}`, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				run(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{wire(tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					v, e := sdk.GraphAPI.TeamOperations.WaitByID(ctx, teamID, operationID, client.WaitOptions{MaxPolls: 1, Timeout: time.Minute})
					if v.Status.State == client.OperationSucceeded {
						t.Error("invalid response established success")
					}
					return contracttest.Observation{Response: v.Response, Err: e}
				}, Want: contracttest.Expected{Status: tc.status, Error: client.ErrOperationState}})
			})
		})
	}
}
