package group_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
)

func listObservation(result client.ListResult[group.Group], err error) contracttest.Observation {
	ids := []string{}
	for _, item := range result.Items {
		ids = append(ids, item.ID)
	}
	return contracttest.Observation{Value: map[string]any{"ids": ids, "pages": result.Pages}, Response: result.Response, Err: err}
}

func TestUnitGroupListAdvancedBetaQueryAndOpaquePaging(t *testing.T) {
	t.Parallel()
	query := url.Values{"$filter": {"resourceProvisioningOptions/Any(x:x eq 'Team')"}, "$search": {`"displayName:fixture"`}, "$select": {"id,displayName"}, "$orderby": {"displayName"}, "$expand": {"owners($select=id)"}, "$top": {"999"}, "$count": {"true"}}

	first := wire("GET", auth.GraphAudience, "/beta/groups?"+query.Encode(), "", 200, "")
	second := wire("GET", auth.GraphAudience, "/beta/groups?$skiptoken=opaque%2B%2F%3D", "", 200, "")
	first.Request.Headers.Set("ConsistencyLevel", "eventual")
	second.Request.Headers.Set("ConsistencyLevel", "eventual")
	count := true
	runCase(t, contracttest.Case{Name: "advanced-beta-pages", Steps: []contracttest.Exchange{first, second}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		fields := []string{"id", "displayName"}
		pager, err := sdk.GraphAPI.Groups.NewPager(group.ListOptions{Version: group.Beta, Filter: query.Get("$filter"), Search: query.Get("$search"), Select: fields, Expand: []string{"owners($select=id)"}, OrderBy: []string{"displayName"}, Top: 999, Count: &count, EventualConsistency: true}, client.PagingOptions{MaxPages: 2})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		fields[0] = "changed"
		count = false
		return listObservation(pager.Collect(ctx))
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"ids":["` + groupID + `"],"pages":2}`)}})
}

func TestUnitGroupListReturnsOneNativePage(t *testing.T) {
	t.Parallel()
	count := false
	body := `{"value":[{"id":"` + groupID + `"}],"@odata.nextLink":"https://foreign.invalid/opaque","future":7}`
	runCase(t, contracttest.Case{Name: "single-native-page", Steps: []contracttest.Exchange{wire("GET", auth.GraphAudience, "/v1.0/groups?$count=false", "", 200, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		page, response, err := sdk.GraphAPI.Groups.List(ctx, group.ListOptions{Count: &count})
		if err != nil && isRequestBoundary(ctx) {
			return contracttest.Observation{Response: response, Err: err}
		}
		var value any
		if page != nil {
			value = page.RawFields
		}
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(body)}})
}

func TestUnitGroupListPartialResultsAndStatusGuard(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, next string
		limit      int
		second     bool
		status     int
		want       error
	}{
		{"page-limit", auth.GraphAudience + "/v1.0/groups?$skiptoken=next", 1, false, 200, client.ErrPageLimit},
		{"foreign-next", "https://foreign.invalid/v1.0/groups", 0, false, 200, client.ErrDestination},
		{"forbidden", auth.GraphAudience + "/v1.0/groups?$skiptoken=next", 0, true, 403, auth.ErrForbidden},
		{"unexpected-partial", auth.GraphAudience + "/v1.0/groups?$skiptoken=next", 0, true, 206, client.ErrPageStatus},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []contracttest.Exchange{wire("GET", auth.GraphAudience, "/v1.0/groups", "", 200, `{"value":[{"id":"`+groupID+`"}],"@odata.nextLink":"`+tc.next+`"}`)}
			if tc.second {
				steps = append(steps, wire("GET", auth.GraphAudience, "/v1.0/groups?$skiptoken=next", "", tc.status, `{"value":[],"error":{"code":"Forbidden"}}`))
			}
			runCase(t, contracttest.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return listObservation(sdk.GraphAPI.Groups.ListAll(ctx, group.ListOptions{}, client.PagingOptions{MaxPages: tc.limit}))
			},
				Want: contracttest.Expected{Status: tc.status, Error: tc.want, JSON: json.RawMessage(`{"ids":["` + groupID + `"],"pages":1}`)}})
		})
	}
	for _, body := range []string{`{}`, `{"value":null}`, `{"value":[null]}`, `{"value":[],"@odata.nextLink":7}`} {
		runCase(t, contracttest.Case{Steps: []contracttest.Exchange{wire("GET", auth.GraphAudience, "/v1.0/groups", "", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, response, err := sdk.GraphAPI.Groups.List(ctx, group.ListOptions{})
			return contracttest.Observation{Response: response, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}})
	}
}

func TestUnitGroupPagerConfigurationIsLocal(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		options group.ListOptions
		limits  client.PagingOptions
		want    error
	}{
		{options: group.ListOptions{Version: "private"}, want: client.ErrRequest},
		{options: group.ListOptions{Top: 1000}, want: client.ErrRequest},
		{options: group.ListOptions{Top: -1}, want: client.ErrRequest},
		{options: group.ListOptions{Select: []string{""}}, want: client.ErrRequest},
		{options: group.ListOptions{Expand: []string{""}}, want: client.ErrRequest},
		{options: group.ListOptions{OrderBy: []string{""}}, want: client.ErrRequest},
		{limits: client.PagingOptions{ExpectedStatus: http.StatusPartialContent}, want: client.ErrPagination},
		{limits: client.PagingOptions{Style: client.TokenPaging}, want: client.ErrPagination},
		{limits: client.PagingOptions{MaxItems: -1}, want: client.ErrPagination},
	} {
		runCase(t, contracttest.Case{Scope: "local", Invoke: func(_ context.Context, sdk *msteams.Client) contracttest.Observation {
			_, err := sdk.GraphAPI.Groups.NewPager(tc.options, tc.limits)
			return contracttest.Observation{Err: err}
		}, Want: contracttest.Expected{Error: tc.want}})
	}
}
