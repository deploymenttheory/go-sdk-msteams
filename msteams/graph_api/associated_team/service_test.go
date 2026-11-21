package associatedteam_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	associated "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/associated_team"
)

const userID = "11111111-1111-4111-8111-111111111111"

func wire(path string, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}
func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family, c.Source = "associated_teams", "msteams/graph_api/associated_team/service_test.go"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, e := contracttest.Run(t.Context(), requestBoundary(t, c))
	if e != nil || !r.Passed() {
		t.Fatalf("associated-team contract differs: setup=%v differences=%+v", e, r.Differences())
	}
}

func TestUnitAssociatedViewsKeepCurrentAndExplicitUsersSeparate(t *testing.T) {
	t.Parallel()
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		t.Run(string(mode), func(t *testing.T) {
			path := "/beta/users/" + userID + "/teamwork/associatedTeams"

			run(t, contracttest.Case{Name: "explicit-user", Mode: mode, Steps: []contracttest.Exchange{wire(path, ""), wire(path+"?$skiptoken=opaque%2Bvalue", "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, e := sdk.GraphAPI.AssociatedTeams.ListForUser(ctx, userID, client.PagingOptions{})
				if e != nil && isRequestBoundary(ctx) {
					return contracttest.Observation{Response: v.Response, Err: e}
				}
				return contracttest.Observation{Value: v.Items, Response: v.Response, Err: e}
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`[{"id":"team","tenantId":"foreign"}]`)}})
		})
	}
	run(t, contracttest.Case{Name: "delegated-me", Mode: auth.Delegated, Steps: []contracttest.Exchange{wire("/beta/me/teamwork/associatedTeams", "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.AssociatedTeams.ListMine(ctx, client.PagingOptions{})
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: v.Response, Err: e}
		}
		return contracttest.Observation{Response: v.Response, Err: e}
	}, Want: contracttest.Expected{Status: 200}})
	run(t, contracttest.Case{Name: "application-me-rejected", Mode: auth.Application, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.AssociatedTeams.ListMine(ctx, client.PagingOptions{})
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: v.Response, Err: e}
		}
		return contracttest.Observation{Response: v.Response, Err: e}
	}, Want: contracttest.Expected{Error: associated.ErrCurrentUser}})
	run(t, contracttest.Case{Name: "empty-user-not-me", Mode: auth.Delegated, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, e := sdk.GraphAPI.AssociatedTeams.ListForUser(ctx, "", client.PagingOptions{})
		if e != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: v.Response, Err: e}
		}
		return contracttest.Observation{Response: v.Response, Err: e}
	}, Want: contracttest.Expected{Error: client.ErrRequest}})
}
