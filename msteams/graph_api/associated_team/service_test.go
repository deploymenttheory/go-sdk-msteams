package associatedteam_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	associated "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/associated_team"
)

const userID = "11111111-1111-4111-8111-111111111111"

func wire(path string, body string) testutil.Exchange {
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}
func run(t *testing.T, c testutil.Case) {
	t.Helper()

	r, e := testutil.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("associated-team contract differs: setup=%v differences=%+v", e, r.Differences())
	}
}

func TestUnitAssociatedViewsKeepCurrentAndExplicitUsersSeparate(t *testing.T) {
	t.Parallel()
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		t.Run(string(mode), func(t *testing.T) {
			path := "/beta/users/" + userID + "/teamwork/associatedTeams"

			run(t, testutil.Case{Name: "explicit-user", Mode: mode, Steps: []testutil.Exchange{wire(path, `{"value":[],"@odata.nextLink":"`+auth.GraphAudience+path+`?$skiptoken=opaque%2Bvalue"}`), wire(path+"?$skiptoken=opaque%2Bvalue", `{"value":[{"id":"team","tenantId":"foreign"}]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, e := sdk.GraphAPI.AssociatedTeams.ListForUser(ctx, userID, client.PagingOptions{})

				return testutil.Observation{Value: v.Items, Response: v.Response, Err: e}
			}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`[{"id":"team","tenantId":"foreign"}]`)}})
		})
	}
	run(t, testutil.Case{Name: "delegated-me", Mode: auth.Delegated, Steps: []testutil.Exchange{wire("/beta/me/teamwork/associatedTeams", `{"value":[]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, e := sdk.GraphAPI.AssociatedTeams.ListMine(ctx, client.PagingOptions{})

		return testutil.Observation{Response: v.Response, Err: e}
	}, Want: testutil.Expected{Status: 200}})
	run(t, testutil.Case{Name: "application-me-rejected", Mode: auth.Application, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, e := sdk.GraphAPI.AssociatedTeams.ListMine(ctx, client.PagingOptions{})

		return testutil.Observation{Response: v.Response, Err: e}
	}, Want: testutil.Expected{Error: associated.ErrCurrentUser}})
	run(t, testutil.Case{Name: "empty-user-not-me", Mode: auth.Delegated, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, e := sdk.GraphAPI.AssociatedTeams.ListForUser(ctx, "", client.PagingOptions{})

		return testutil.Observation{Response: v.Response, Err: e}
	}, Want: testutil.Expected{Error: client.ErrRequest}})
}
