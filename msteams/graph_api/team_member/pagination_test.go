package teammember_test

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	teammember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_member"
)

func TestListReturnsOneNativePageWithoutImplicitTraversal(t *testing.T) {
	t.Parallel()
	next := auth.GraphAudience + "/v1.0/teams/" + teamID + "/members?$skiptoken=opaque%2B%2F%3D"
	query := url.Values{"$filter": {"roles/any(r:r eq 'owner')"}, "$select": {"id,roles,userId"}, "$top": {"999"}}
	body := `{"value":[{"id":"opaque","@odata.type":"#microsoft.graph.futureMember","roles":["future"],"userId":null,"tenantId":"other-tenant"}],"@odata.nextLink":"` + next + `","futurePage":7}`
	nativeCase(t, contracttest.Case{Name: "single-page", Steps: []contracttest.Exchange{wire("GET", auth.GraphAudience, "/v1.0/teams/"+teamID+"/members?"+query.Encode(), "", 200, body)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			page, response, err := sdk.GraphAPI.TeamMembers.List(ctx, teamID, teammember.ListOptions{Filter: "roles/any(r:r eq 'owner')", Select: []string{"id", "roles", "userId"}, Top: 999})
			var value any
			if page != nil {
				value = map[string]any{"count": len(page.Items), "id": page.Items[0].ID, "roles": page.Items[0].Roles, "user_id": page.Items[0].UserID, "next": page.NextLink, "raw": page.RawFields}
			}
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"count":1,"id":"opaque","roles":["future"],"user_id":null,"next":"` + next + `","raw":` + body + `}`)}})
}

func listObservation(result client.ListResult[teammember.Member], err error) contracttest.Observation {
	ids := []teammember.MembershipID{}
	for _, member := range result.Items {
		ids = append(ids, member.ID)
	}
	return contracttest.Observation{Value: map[string]any{"ids": ids, "pages": result.Pages}, Response: result.Response, Err: err}
}

func TestListAllOpaquePagingAndFrozenOptions(t *testing.T) {
	t.Parallel()
	path := "/v1.0/teams/" + teamID + "/members"
	next := auth.GraphAudience + path + "?$skiptoken=opaque%2B%2F%3D&$select=id%2CuserId"
	nativeCase(t, contracttest.Case{Name: "all-pages", Steps: []contracttest.Exchange{
		wire("GET", auth.GraphAudience, path+"?$select=id%2CuserId&$top=1", "", 200, `{"value":[{"id":"opaque-1","roles":[]}],"@odata.nextLink":"`+next+`"}`),
		wire("GET", auth.GraphAudience, path+"?$skiptoken=opaque%2B%2F%3D&$select=id%2CuserId", "", 200, `{"value":[]}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			fields := []string{"id", "userId"}
			pager, err := sdk.GraphAPI.TeamMembers.NewPager(teamID, teammember.ListOptions{Select: fields, Top: 1}, client.PagingOptions{MaxPages: 2, MaxItems: 1})
			if err != nil {
				return contracttest.Observation{Err: err}
			}
			fields[0] = "changed-after-construction"
			result, err := pager.Collect(ctx)
			return listObservation(result, err)
		}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"ids":["opaque-1"],"pages":2}`)}})
}

func TestListAllPreservesPartialResultsOnFailureAndLimits(t *testing.T) {
	t.Parallel()
	path := "/v1.0/teams/" + teamID + "/members"
	for _, tc := range []struct {
		name, next    string
		limits        client.PagingOptions
		second        bool
		want          error
		status, pages int
	}{
		{"forbidden", auth.GraphAudience + path + "?$skiptoken=next", client.PagingOptions{}, true, auth.ErrForbidden, 403, 1},
		{"page-limit", auth.GraphAudience + path + "?$skiptoken=next", client.PagingOptions{MaxPages: 1}, false, client.ErrPageLimit, 200, 1},
		{"foreign-next", "https://foreign.invalid/v1.0/teams/" + teamID + "/members", client.PagingOptions{}, false, client.ErrDestination, 200, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []contracttest.Exchange{wire("GET", auth.GraphAudience, path, "", 200, `{"value":[{"id":"opaque"}],"@odata.nextLink":"`+tc.next+`"}`)}
			if tc.second {
				steps = append(steps, wire("GET", auth.GraphAudience, path+"?$skiptoken=next", "", 403, `{"error":{"code":"Forbidden"}}`))
			}
			nativeCase(t, contracttest.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return listObservation(sdk.GraphAPI.TeamMembers.ListAll(ctx, teamID, teammember.ListOptions{}, tc.limits))
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.want, JSON: json.RawMessage(`{"ids":["opaque"],"pages":1}`)}})
		})
	}
	t.Run("aggregate-limit-retains-prefix", func(t *testing.T) {
		nativeCase(t, contracttest.Case{Name: "item-limit", Steps: []contracttest.Exchange{
			wire("GET", auth.GraphAudience, path, "", 200, `{"value":[{"id":"opaque-1"}],"@odata.nextLink":"`+auth.GraphAudience+path+`?$skiptoken=next"}`),
			wire("GET", auth.GraphAudience, path+"?$skiptoken=next", "", 200, `{"value":[{"id":"opaque-2"}]}`)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return listObservation(sdk.GraphAPI.TeamMembers.ListAll(ctx, teamID, teammember.ListOptions{}, client.PagingOptions{MaxItems: 1}))
			}, Want: contracttest.Expected{Status: 200, Error: client.ErrItemLimit, JSON: json.RawMessage(`{"ids":["opaque-1"],"pages":2}`)}})
	})
}

func TestListRejectsMalformedCollectionsAndPreservesResponse(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`{}`, `null`, `{"value":null}`, `{"value":[null]}`, `{"value":[{"userId":"directory-only"}]}`, `{"value":[],"@odata.nextLink":7}`} {
		nativeCase(t, contracttest.Case{Steps: []contracttest.Exchange{wire("GET", auth.GraphAudience, "/v1.0/teams/"+teamID+"/members", "", 200, body)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, response, err := sdk.GraphAPI.TeamMembers.List(ctx, teamID, teammember.ListOptions{})
				return contracttest.Observation{Response: response, Err: err}
			}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}})
	}
}

func TestCancellationClosedBackendAndInvalidPagerAreLocal(t *testing.T) {
	t.Parallel()
	for _, action := range []string{"get", "create", "update", "delete", "list", "all", "closed", "wrong-backend", "invalid-team", "invalid-user", "invalid-projection", "invalid-style", "invalid-status", "invalid-limit", "nil-service"} {
		t.Run(action, func(t *testing.T) {
			var want = context.Canceled
			switch action {
			case "closed", "nil-service":
				want = client.ErrClosed
			case "wrong-backend":
				want = client.ErrDestination
			case "invalid-team", "invalid-user", "invalid-projection":
				want = client.ErrRequest
			case "invalid-style", "invalid-status", "invalid-limit":
				want = client.ErrPagination
			}
			nativeCase(t, contracttest.Case{Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				var err error
				switch action {
				case "get":
					_, _, err = sdk.GraphAPI.TeamMembers.GetByID(cancelled, teamID, memberID)
				case "create":
					_, _, err = sdk.GraphAPI.TeamMembers.Create(cancelled, teamID, teammember.CreateRequest{UserID: userID})
				case "update":
					_, _, err = sdk.GraphAPI.TeamMembers.UpdateByID(cancelled, teamID, memberID, teammember.UpdateRequest{})
				case "delete":
					_, err = sdk.GraphAPI.TeamMembers.DeleteByID(cancelled, teamID, memberID)
				case "list":
					_, _, err = sdk.GraphAPI.TeamMembers.List(cancelled, teamID, teammember.ListOptions{})
				case "all":
					_, err = sdk.GraphAPI.TeamMembers.ListAll(cancelled, teamID, teammember.ListOptions{}, client.PagingOptions{})
				case "closed":
					if closeErr := sdk.Close(); closeErr != nil {
						return contracttest.Observation{Err: closeErr}
					}
					_, err = sdk.GraphAPI.TeamMembers.NewPager(teamID, teammember.ListOptions{}, client.PagingOptions{})
				case "wrong-backend":
					_, _, err = teammember.NewService(sdk.TeamsAdminAPI.Scope()).GetByID(ctx, teamID, memberID)
				case "invalid-team":
					_, _, err = sdk.GraphAPI.TeamMembers.GetByID(ctx, "../private", memberID)
				case "invalid-user":
					_, _, err = sdk.GraphAPI.TeamMembers.Create(ctx, teamID, teammember.CreateRequest{UserID: "../private"})
				case "invalid-projection":
					_, _, err = sdk.GraphAPI.TeamMembers.List(ctx, teamID, teammember.ListOptions{Select: []string{""}})
				case "invalid-style":
					_, err = sdk.GraphAPI.TeamMembers.NewPager(teamID, teammember.ListOptions{}, client.PagingOptions{Style: client.TokenPaging})
				case "invalid-status":
					_, err = sdk.GraphAPI.TeamMembers.NewPager(teamID, teammember.ListOptions{}, client.PagingOptions{ExpectedStatus: 206})
				case "invalid-limit":
					_, err = sdk.GraphAPI.TeamMembers.NewPager(teamID, teammember.ListOptions{}, client.PagingOptions{MaxPages: -1})
				case "nil-service":
					var service *teammember.Service
					_, _, err = service.GetByID(ctx, teamID, memberID)
				}
				return contracttest.Observation{Err: err}
			}, Want: contracttest.Expected{Error: want}})
		})
	}
}

func TestListAllRejectsUnexpectedSuccessfulStatus(t *testing.T) {
	t.Parallel()
	for _, status := range []int{202, 206} {
		nativeCase(t, contracttest.Case{Name: "unexpected-status", Steps: []contracttest.Exchange{wire("GET", auth.GraphAudience, "/v1.0/teams/"+teamID+"/members", "", status, `{"value":[{"id":"opaque"}]}`)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return listObservation(sdk.GraphAPI.TeamMembers.ListAll(ctx, teamID, teammember.ListOptions{}, client.PagingOptions{}))
			}, Want: contracttest.Expected{Status: status, Error: client.ErrPageStatus, JSON: json.RawMessage(`{"ids":[],"pages":0}`)}})
	}
}
