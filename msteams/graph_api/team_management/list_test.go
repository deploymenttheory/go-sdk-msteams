package teammanagement_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

const secondTeamID = "33333333-3333-4333-8333-333333333333"

func listCase(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Mappings = []contracttest.Mapping{{Command: "Get-Team", ParameterSet: "Filters"}}
	return runCase(t, "Get-Team", c)
}

func listingObserved(result management.ListResult, err error) contracttest.Observation {
	ids := []string{}
	for _, item := range result.Items {
		ids = append(ids, item.ID)
	}
	outcomes := []map[string]any{}
	for _, outcome := range result.Outcomes {
		outcomes = append(outcomes, map[string]any{"id": outcome.ID, "failed": outcome.Err != nil, "filtered": outcome.Filtered, "attempts": outcome.Attempts})
	}
	return contracttest.Observation{Value: map[string]any{"ids": ids, "outcomes": outcomes, "discovery_complete": result.DiscoveryComplete, "complete": result.Complete}, Response: result.Response, Err: err}
}

func discovery(search, body string) contracttest.Exchange {
	query := url.Values{"$filter": {"resourceProvisioningOptions/Any(x:x eq 'Team')"}, "$top": {"999"}, "$count": {"true"}}
	if search != "" {
		query.Set("$search", search)
	}
	step := wire("GET", "/beta/groups?"+query.Encode(), "", 200, body)
	step.Request.Headers.Set("ConsistencyLevel", "eventual")
	return step
}

func batch(ids []string, response string) contracttest.Exchange {
	requests := []map[string]string{}
	for index, id := range ids {
		requests = append(requests, map[string]string{"id": fmt.Sprint(index + 1), "method": "GET", "url": "/teams/" + id})
	}
	body, err := json.Marshal(map[string]any{"requests": requests})
	if err != nil {
		panic("synthetic batch fixture could not be encoded")
	}
	return wire("POST", "/v1.0/$batch", string(body), 200, response)
}

func oneGroup() string {
	return `{"value":[{"id":"` + groupID + `","displayName":"Group name","visibility":"Private","mailNickname":"fixture"}]}`
}

func TestListSourceDerivedTenantUserAndPartialBatch(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, name := range []string{"tenant-search-intersection", "joined-filtered", "tenant-partial-batch"} {
		options := management.ListOptions{}
		var steps []contracttest.Exchange
		want := contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"ids":["` + groupID + `"],"outcomes":[{"id":"` + groupID + `","failed":false,"filtered":false,"attempts":1}],"discovery_complete":true,"complete":true}`)}
		switch name {
		case "tenant-search-intersection":
			disabled := false
			options.DisplayName = "Fixture"
			options.MailNickname = "fixture"
			options.Archived = &disabled
			steps = fixture(t, "team-list-search-intersection")
		case "joined-filtered":
			user := ownerID
			disabled := false
			options.User = &user
			options.DisplayName = "Group name"
			options.Visibility = "private"
			options.Archived = &disabled
			steps = fixture(t, "team-list-joined-user")
		case "tenant-partial-batch":
			// The authored scenario's illustrative fixture-denied identity is
			// substituted with a synthetic directory UUID for the native route.
			steps = []contracttest.Exchange{discovery("", `{"value":[{"id":"`+groupID+`"},{"id":"`+secondTeamID+`"}]}`)}
			partial := fixture(t, "team-list-partial-batch")[0]
			partial.Request.JSON = json.RawMessage(strings.ReplaceAll(string(partial.Request.JSON), "fixture-denied", secondTeamID))
			steps = append(steps, partial)
			want.Error = auth.ErrForbidden
			want.JSON = json.RawMessage(`{"ids":["` + groupID + `"],"outcomes":[{"id":"` + groupID + `","failed":false,"filtered":false,"attempts":1},{"id":"` + secondTeamID + `","failed":true,"filtered":false,"attempts":1}],"discovery_complete":true,"complete":false}`)
		}
		results = append(results, listCase(t, contracttest.Case{Name: name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, options)
			if name == "joined-filtered" && (result.Items[0].DisplayName == nil || *result.Items[0].DisplayName != "Group name") {
				t.Fatal("paired group properties were not projected")
			}
			return listingObserved(result, err)
		}, Want: want}))
	}
	evidence(t, "team-list-service.json", results)
}

func TestIncompleteDiscoveryCannotEstablishIntersectionOrHydrate(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"page-bound", "foreign-next", "second-search-denied"} {
		first := discovery(`"displayName:Group"`, `{"value":[{"id":"`+groupID+`"}],"@odata.nextLink":"`+auth.GraphAudience+`/beta/groups?$skiptoken=next"}`)
		options := management.ListOptions{DisplayName: "Group", MailNickname: "fixture"}
		steps := []contracttest.Exchange{first}
		want := client.ErrPageLimit
		status := 200
		switch name {
		case "page-bound":
			options.Traversal.MaxPages = 1
		case "foreign-next":
			first.Reply.Body = []byte(`{"value":[{"id":"` + groupID + `"}],"@odata.nextLink":"https://foreign.invalid/private"}`)
			steps[0] = first
			want = client.ErrDestination
		case "second-search-denied":
			first.Reply.Body = []byte(oneGroup())
			steps[0] = first
			second := discovery(`"mailNickname:fixture"`, `{"error":{"code":"Forbidden"}}`)
			second.Reply.Status = 403
			steps = append(steps, second)
			want = auth.ErrForbidden
			status = 403
		}
		listCase(t, contracttest.Case{Name: name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, options)
			if result.Complete || result.DiscoveryComplete || len(result.Items) != 0 || len(result.Outcomes) != 0 || len(result.Discovery) == 0 || len(result.Discovery[0].Groups) != 1 {
				t.Fatal("partial discovery was lost or treated as a complete selection")
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: status, Error: want}})
	}
}

func TestHydrationRetriesOnlyThrottledItemsAndPreservesOrdering(t *testing.T) {
	for _, header := range []string{"2", "0"} {
		synctest.Test(t, func(t *testing.T) {
			groups := discovery("", `{"value":[{"id":"`+groupID+`"},{"id":"`+secondTeamID+`"}]}`)
			first := batch([]string{groupID, secondTeamID}, `{"responses":[{"id":"2","status":429,"headers":{"rEtRy-AfTeR":"`+header+`"},"body":{"error":{"code":"TooManyRequests"}}},{"id":"1","status":200,"body":{"id":"`+groupID+`"}}]}`)
			second := wire("POST", "/v1.0/$batch", `{"requests":[{"id":"2","method":"GET","url":"/teams/`+secondTeamID+`"}]}`, 200, `{"responses":[{"id":"2","status":200,"body":{"id":"`+secondTeamID+`"}}]}`)
			start := time.Now()
			listCase(t, contracttest.Case{Name: "throttle-" + header, Steps: []contracttest.Exchange{groups, first, second}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{})
				if len(result.Outcomes) != 2 || result.Outcomes[0].Attempts != 1 || result.Outcomes[1].Attempts != 2 || len(result.Outcomes[1].Subresponses) != 2 {
					t.Fatal("retry changed successful peers or lost attempt metadata")
				}
				return listingObserved(result, err)
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"ids":["` + groupID + `","` + secondTeamID + `"],"outcomes":[{"id":"` + groupID + `","failed":false,"filtered":false,"attempts":1},{"id":"` + secondTeamID + `","failed":false,"filtered":false,"attempts":2}],"discovery_complete":true,"complete":true}`)}})
			want := 2 * time.Second
			if header == "0" {
				want = 10 * time.Second
			}
			if time.Since(start) != want {
				t.Fatal("positive Retry-After or fallback was not honored")
			}
		})
	}
}

func TestHydrationTimeoutKeepsSuccessfulPeers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		steps := []contracttest.Exchange{discovery("", `{"value":[{"id":"`+groupID+`"},{"id":"`+secondTeamID+`"}]}`), batch([]string{groupID, secondTeamID}, `{"responses":[{"id":"1","status":200,"body":{"id":"`+groupID+`"}},{"id":"2","status":429,"headers":{"Retry-After":"30"}}]}`)}
		listCase(t, contracttest.Case{Name: "throttle-timeout", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{Timeout: time.Second})
			if len(result.Items) != 1 || result.Items[0].ID != groupID || result.Complete || result.Outcomes[1].Attempts != 1 || len(result.Outcomes[1].Subresponses) != 1 {
				t.Fatal("timeout lost successful peers or retained throttle observation")
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: context.DeadlineExceeded}})
	})
}

func TestHydrationRejectsMalformedWrongIdentityAndUnexpectedStatuses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		want       error
	}{
		{"wrong-team-id", `{"responses":[{"id":"1","status":200,"body":{"id":"` + secondTeamID + `"}}]}`, 200, client.ErrDecode},
		{"missing-team-id", `{"responses":[{"id":"1","status":200,"body":{}}]}`, 200, client.ErrDecode},
		{"sub-accepted", `{"responses":[{"id":"1","status":202,"body":{"id":"` + groupID + `"}}]}`, 200, management.ErrHydrationStatus},
		{"sub-partial", `{"responses":[{"id":"1","status":206,"body":{"id":"` + groupID + `"}}]}`, 200, management.ErrHydrationStatus},
		{"outer-partial", `{"responses":[{"id":"1","status":200,"body":{"id":"` + groupID + `"}}]}`, 206, management.ErrHydrationStatus},
	} {
		step := batch([]string{groupID}, tc.body)
		step.Reply.Status = tc.status
		listCase(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{discovery("", oneGroup()), step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{})
			if result.Complete || len(result.Items) != 0 || len(result.Outcomes) != 1 || result.Outcomes[0].ID != groupID {
				t.Fatal("invalid hydration claimed complete data or lost identity")
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: tc.status, Error: tc.want}})
	}
}

func TestListingLocalValidationAndCurrentUserMode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"batch-negative", "batch-oversized", "attempts-negative", "timeout-negative", "delay-negative", "pages-negative", "wrong-paging", "control-search", "current-application"} {
		want := client.ErrRequest
		mode := auth.Delegated
		if name == "pages-negative" || name == "wrong-paging" {
			want = client.ErrPagination
		}
		if name == "current-application" {
			mode = auth.Application
			want = directoryuser.ErrCurrentUser
		}
		listCase(t, contracttest.Case{Name: name, Scope: "local", Mode: mode, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			options := management.ListOptions{}
			switch name {
			case "batch-negative":
				options.BatchSize = -1
			case "batch-oversized":
				options.BatchSize = 21
			case "attempts-negative":
				options.MaxThrottleAttempts = -1
			case "timeout-negative":
				options.Timeout = -1
			case "delay-negative":
				options.ThrottleDelay = -1
			case "pages-negative":
				options.Traversal.MaxPages = -1
			case "wrong-paging":
				options.Traversal.Style = client.OffsetPaging
			case "control-search":
				options.DisplayName = "private\nsearch"
			case "current-application":
				user := ""
				options.User = &user
			}
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, options)
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Error: want}})
	}
}

func TestBatchWidthChunksAndResponseOrderAreIndependent(t *testing.T) {
	t.Parallel()
	ids := []string{groupID, secondTeamID, "44444444-4444-4444-8444-444444444444", "55555555-5555-4555-8555-555555555555", "66666666-6666-4666-8666-666666666666"}
	items := []map[string]string{}
	for _, id := range ids {
		items = append(items, map[string]string{"id": id})
	}
	body, err := json.Marshal(map[string]any{"value": items})
	if err != nil {
		t.Fatal(err)
	}
	steps := []contracttest.Exchange{discovery("", string(body))}
	for first := 0; first < len(ids); first += 2 {
		requests := []map[string]string{}
		responses := []map[string]any{}
		for index := first; index < min(first+2, len(ids)); index++ {
			id := fmt.Sprint(index + 1)
			requests = append(requests, map[string]string{"id": id, "method": "GET", "url": "/teams/" + ids[index]})
			responses = append([]map[string]any{{"id": id, "status": 200, "body": map[string]string{"id": ids[index]}}}, responses...)
		}
		requestBody, err := json.Marshal(map[string]any{"requests": requests})
		if err != nil {
			t.Fatal(err)
		}
		responseBody, err := json.Marshal(map[string]any{"responses": responses})
		if err != nil {
			t.Fatal(err)
		}
		steps = append(steps, wire("POST", "/v1.0/$batch", string(requestBody), 200, string(responseBody)))
	}
	listCase(t, contracttest.Case{Name: "batch-width-two", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{BatchSize: 2})
		if err != nil || !result.Complete || len(result.Items) != len(ids) {
			t.Fatal("batch width or complete outcomes changed")
		}
		for index, id := range ids {
			if result.Items[index].ID != id || result.Outcomes[index].ID != id || result.Outcomes[index].Attempts != 1 {
				t.Fatal("response order changed caller-visible order/correlation")
			}
		}
		return contracttest.Observation{Response: result.Response, Err: err}
	}, Want: contracttest.Expected{Status: 200}})
}

func TestHydrationFailureHasNormalizedSubrequestDiagnostics(t *testing.T) {
	t.Parallel()
	steps := []contracttest.Exchange{discovery("", oneGroup()), batch([]string{groupID}, `{"responses":[{"id":"1","status":403,"headers":{"request-id":"fixture-subrequest","client-request-id":"fixture-client"},"body":{"error":{"code":"Forbidden","message":"private-server-message","innerError":{"code":"FixtureInner"}}}}]}`)}
	listCase(t, contracttest.Case{Name: "subrequest-diagnostics", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{})
		var apiError *client.APIError
		var failure *management.ItemFailure
		if !errors.As(err, &apiError) || apiError.StatusCode != 403 || len(apiError.Codes) != 2 || apiError.Codes[0] != "Forbidden" || apiError.Codes[1] != "FixtureInner" || len(apiError.RequestIDs) != 1 || apiError.RequestIDs[0] != "fixture-subrequest" || len(apiError.ClientRequestIDs) != 1 || apiError.ClientRequestIDs[0] != "fixture-client" || !errors.As(err, &failure) || failure.ID != groupID {
			t.Fatal("typed failure lost identity or normalized batch diagnostics")
		}
		for _, value := range []any{result, result.Outcomes[0], result.Outcomes[0].Subresponses[0], failure, *failure, apiError, *apiError} {
			if rendered := fmt.Sprintf("%+v %#v", value, value); strings.Contains(rendered, groupID) || strings.Contains(rendered, "private-server-message") || strings.Contains(rendered, "fixture-subrequest") {
				t.Fatal("default formatting exposed listing response or identity")
			}
		}
		return contracttest.Observation{Response: result.Response, Err: err}
	}, Want: contracttest.Expected{Status: 200, Error: auth.ErrForbidden}})
}

func TestUserFiltersUsePairedStateAndKnownNonmatches(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"exact-case", "mail-case-insensitive", "archive-unknown", "known-nonmatch-with-unknown-archive"} {
		user := ownerID
		options := management.ListOptions{User: &user}
		steps := fixture(t, "team-list-joined-user")
		// Basic joined projections do not override the subsequent paired reads.
		steps[0].Reply.Body = []byte(`{"value":[{"id":"` + groupID + `","isArchived":true}]}`)
		failed, filtered := false, false
		switch name {
		case "exact-case":
			options.DisplayName = "group name"
			filtered = true
		case "mail-case-insensitive":
			options.MailNickname = "FIXTURE"
			disabled := false
			options.Archived = &disabled
		case "archive-unknown":
			steps[2].Reply.Body = []byte(`{"id":"` + groupID + `"}`)
			disabled := false
			options.Archived = &disabled
			failed = true
		case "known-nonmatch-with-unknown-archive":
			steps[2].Reply.Body = []byte(`{"id":"` + groupID + `"}`)
			disabled := false
			options.Archived = &disabled
			options.Visibility = "public"
			filtered = true
		}
		var want error
		if failed {
			want = management.ErrFilterUnavailable
		}
		listCase(t, contracttest.Case{Name: name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, options)
			if len(result.Outcomes) != 1 || result.Outcomes[0].Filtered != filtered || (result.Outcomes[0].Err != nil) != failed || result.Complete == failed || result.Outcomes[0].Value == nil {
				t.Fatal("filtering fabricated unknown state or changed comparison semantics")
			}
			if !failed && !filtered && len(result.Items) != 1 {
				t.Fatal("paired current state did not match")
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: want}})
	}
}

func TestMissingBatchResultsKeepKnownSuccessAndFailedIdentity(t *testing.T) {
	t.Parallel()
	steps := []contracttest.Exchange{discovery("", `{"value":[{"id":"`+groupID+`"},{"id":"`+secondTeamID+`"}]}`), batch([]string{groupID, secondTeamID}, `{"responses":[{"id":"1","status":200,"body":{"id":"`+groupID+`"}}]}`)}
	listCase(t, contracttest.Case{Name: "missing-batch-item", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{})
		if result.Complete || len(result.Items) != 1 || result.Items[0].ID != groupID || len(result.Outcomes) != 2 || result.Outcomes[1].Err == nil || result.Outcomes[1].ID != secondTeamID || !errors.Is(err, management.ErrListingIncomplete) {
			t.Fatal("malformed batch dropped a known success or claimed complete selection")
		}
		return contracttest.Observation{Response: result.Response, Err: err}
	}, Want: contracttest.Expected{Status: 200, Error: management.ErrListingIncomplete}})
}

func TestThrottleExhaustionIsBoundedAndPreservesFinalDiagnostics(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		throttled := batch([]string{groupID}, `{"responses":[{"id":"1","status":429,"body":{"error":{"code":"TooManyRequests"}}}]}`)
		steps := []contracttest.Exchange{discovery("", oneGroup()), throttled, throttled, throttled, throttled}
		start := time.Now()
		listCase(t, contracttest.Case{Name: "throttle-exhaustion", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{})
			var apiError *client.APIError
			if result.Complete || len(result.Items) != 0 || result.Outcomes[0].Attempts != 4 || len(result.Outcomes[0].Subresponses) != 4 || !errors.As(err, &apiError) || apiError.StatusCode != 429 {
				t.Fatal("throttle attempts were unbounded or final error lost")
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: management.ErrListingIncomplete}})
		if time.Since(start) != 30*time.Second {
			t.Fatal("throttle default retry count or delay differs")
		}
	})
}

func TestClientCloseCancelsPendingListThrottleWithoutLosingSuccess(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		steps := []contracttest.Exchange{discovery("", `{"value":[{"id":"`+groupID+`"},{"id":"`+secondTeamID+`"}]}`), batch([]string{groupID, secondTeamID}, `{"responses":[{"id":"1","status":200,"body":{"id":"`+groupID+`"}},{"id":"2","status":429,"headers":{"Retry-After":"30"}}]}`)}
		listCase(t, contracttest.Case{Name: "close-during-list", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			done := make(chan error, 1)
			go func() { time.Sleep(time.Second); done <- sdk.Close() }()
			result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{})
			if closeErr := <-done; closeErr != nil {
				t.Fatal("client close failed")
			}
			if result.Complete || len(result.Items) != 1 || result.Items[0].ID != groupID || result.Outcomes[1].Attempts != 1 {
				t.Fatal("close lost known result or resubmitted the pending item")
			}
			return contracttest.Observation{Response: result.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200, Error: context.Canceled}})
	})
}
