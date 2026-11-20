package client_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	graphapi "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api"
	teamsadminapi "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api"
)

func paginationResponse(r *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}
}

func TestGraphPaginationUsesOpaqueLinksAndKeepsHeaders(t *testing.T) {
	t.Parallel()
	firstLink := "https://graph.microsoft.com/v1.0/fixture?$skiptoken=a%2Bb%2F%3D%26&$top=2"
	secondLink := "https://graph.microsoft.com/v1.0/fixture?$skiptoken=next"
	calls := 0
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("ConsistencyLevel") != "eventual" {
			t.Error("custom header lost")
		}
		switch calls {
		case 1:
			if r.URL.Query().Get("$filter") != "displayName eq 'A & B'" {
				t.Error("initial query changed")
			}
			return paginationResponse(r, 200, `{"value":[1],"@odata.nextLink":"`+firstLink+`"}`), nil
		case 2:
			if r.URL.String() != firstLink {
				t.Error("opaque link rebuilt or initial query appended")
			}
			return paginationResponse(r, 200, `{"value":[],"@odata.nextLink":"`+secondLink+`"}`), nil
		case 3:
			if r.URL.String() != secondLink {
				t.Error("second link changed")
			}
			return paginationResponse(r, 200, `{"value":[2]}`), nil
		default:
			t.Fatal("request after completion")
			return nil, nil
		}
	})}, "commercial")
	query := url.Values{"$filter": {"displayName eq 'A & B'"}}
	headers := http.Header{"ConsistencyLevel": {"eventual"}}
	pager, err := graphapi.NewPager[int](scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Query: query, Header: headers}, client.PagingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	query.Set("$filter", "changed")
	headers.Set("ConsistencyLevel", "changed")
	result, err := pager.Collect(t.Context())
	if err != nil || !reflect.DeepEqual(result.Items, []int{1, 2}) || result.Pages != 3 || calls != 3 {
		t.Fatalf("result=%#v err=%v", result.Items, err)
	}
	if _, ok, err := pager.Next(t.Context()); err != nil || ok {
		t.Fatal("completed pager made progress")
	}
	if strings.Contains(fmt.Sprintf("%#v", pager), "skiptoken") {
		t.Fatal("pager formatted cursor material")
	}
}

func TestTeamsTokenPaginationPreservesQueryBodyAndTreatsNextLinkAsToken(t *testing.T) {
	t.Parallel()
	cursor := "https://untrusted.invalid/path?x=1&opaque=+/% 雪"
	calls := 0
	scope, _, provider := transportScope(t, client.TeamsAdmin, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "api.interfaces.records.teams.microsoft.com" {
			t.Error("token changed authority")
		}
		if r.URL.Query().Get("filter") != "name eq 'A & B'" || r.URL.Query().Get("pageSize") != "2" {
			t.Error("base query lost")
		}
		body, err := io.ReadAll(r.Body)
		_ = r.Body.Close()
		if err != nil {
			return nil, err
		}
		if string(body) != `{"query":"original"}` {
			t.Error("initial body was not frozen")
		}
		if calls == 1 {
			return paginationResponse(r, 200, `{"users":[],"@nextLink":"https://untrusted.invalid/path?x=1&opaque=+/% 雪"}`), nil
		}
		if r.URL.Query().Get("$skiptoken") != cursor {
			t.Error("opaque token encoding changed")
		}
		return paginationResponse(r, 200, `{"users":[3],"@nextLink":null}`), nil
	})}, "commercial")
	body := map[string]string{"query": "original"}
	pager, err := teamsadminapi.NewTokenPager[int](scope, client.Request{Method: http.MethodPost, Path: []string{"fixture", "users"}, Query: url.Values{"filter": {"name eq 'A & B'"}, "pageSize": {"2"}}, Body: body, Replay: client.ReplaySafe}, teamsadminapi.TokenPageFormat{ItemsField: "users", TokenField: "@nextLink", QueryParameter: "$skiptoken"}, client.PagingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	body["query"] = "changed"
	result, err := pager.Collect(t.Context())
	if err != nil || calls != 2 || !reflect.DeepEqual(result.Items, []int{3}) {
		t.Fatal(err)
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	for _, r := range provider.requests {
		if r.Audience != auth.TeamsAdminAudience {
			t.Fatal("wrong token audience")
		}
	}
}

func TestOffsetPaginationContinuesAfterShortPages(t *testing.T) {
	t.Parallel()
	calls := 0
	scope, _, _ := transportScope(t, client.TeamsAdmin, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		want := []string{"5", "7", "8"}[calls-1]
		if r.URL.Query().Get("skip") != want || r.URL.Query().Get("top") != "10" || r.URL.Query().Get("filter") != "stable" {
			t.Error("offset or initial query changed")
		}
		return paginationResponse(r, 200, []string{`[1,2]`, `[3]`, `[]`}[calls-1]), nil
	})}, "commercial")
	pager, err := teamsadminapi.NewOffsetPager[int](scope, client.Request{Method: http.MethodGet, Path: []string{"fixture", "instances"}, Query: url.Values{"skip": {"5"}, "top": {"10"}, "filter": {"stable"}}, Replay: client.ReplaySafe}, "", "skip", client.PagingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := pager.Collect(t.Context())
	if err != nil || result.Pages != 3 || !reflect.DeepEqual(result.Items, []int{1, 2, 3}) {
		t.Fatalf("items=%v pages=%d err=%v", result.Items, result.Pages, err)
	}
}

func TestPaginationLoopAndDestinationErrorsPreservePartialResults(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, next string
		want       error
	}{
		{"same-link", "https://graph.microsoft.com/v1.0/fixture?b=2&a=1", client.ErrContinuationLoop},
		{"foreign-host", "https://untrusted.invalid/v1.0/fixture", client.ErrDestination},
		{"relative-link", "/v1.0/fixture?skip=2", client.ErrDestination},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			scope, _, provider := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return paginationResponse(r, 200, `{"value":[1],"@odata.nextLink":"`+tc.next+`"}`), nil
			})}, "commercial")
			pager, err := graphapi.NewPager[int](scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Query: url.Values{"a": {"1"}, "b": {"2"}}}, client.PagingOptions{})
			if err != nil {
				t.Fatal(err)
			}
			result, err := pager.Collect(t.Context())
			if !errors.Is(err, tc.want) || calls != 1 || !reflect.DeepEqual(result.Items, []int{1}) || result.Response == nil {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
			provider.mu.Lock()
			defer provider.mu.Unlock()
			if len(provider.requests) != 1 {
				t.Fatal("invalid continuation acquired another credential")
			}
		})
	}
}

func TestPaginationTransportAndDecodeErrorsNeverReturnTruncatedSuccess(t *testing.T) {
	t.Parallel()
	for _, decodeFailure := range []bool{false, true} {
		t.Run(strconv.FormatBool(decodeFailure), func(t *testing.T) {
			calls := 0
			scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return paginationResponse(r, 200, `{"value":[1],"@odata.nextLink":"https://graph.microsoft.com/v1.0/fixture?cursor=2"}`), nil
				}
				if decodeFailure {
					return paginationResponse(r, 200, `{"unexpected":[2]}`), nil
				}
				return paginationResponse(r, 400, `{"error":{"code":"fixtureFailure"}}`), nil
			})}, "commercial")
			pager, err := graphapi.NewPager[int](scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, client.PagingOptions{})
			if err != nil {
				t.Fatal(err)
			}
			result, err := pager.Collect(t.Context())
			if err == nil || calls != 2 || !reflect.DeepEqual(result.Items, []int{1}) || result.Pages != 1 {
				t.Fatal("partial failure lost")
			}
			if decodeFailure && !errors.Is(err, client.ErrDecode) {
				t.Fatal(err)
			}
			if _, ok, again := pager.Next(t.Context()); again == nil || ok || calls != 2 {
				t.Fatal("failed pager resumed as success")
			}
		})
	}
}

func TestPaginationLimitsAreTerminal(t *testing.T) {
	t.Parallel()
	for _, itemLimit := range []bool{false, true} {
		t.Run(strconv.FormatBool(itemLimit), func(t *testing.T) {
			calls := 0
			scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				values := `[1]`
				if itemLimit && calls == 2 {
					values = `[2,3]`
				}
				return paginationResponse(r, 200, fmt.Sprintf(`{"value":%s,"@odata.nextLink":"https://graph.microsoft.com/v1.0/fixture?cursor=%d"}`, values, calls)), nil
			})}, "commercial")
			pager, err := graphapi.NewPager[int](scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, client.PagingOptions{MaxPages: 2, MaxItems: 2})
			if err != nil {
				t.Fatal(err)
			}
			result, err := pager.Collect(t.Context())
			want := client.ErrPageLimit
			items := []int{1, 1}
			if itemLimit {
				want = client.ErrItemLimit
				items = []int{1}
			}
			if !errors.Is(err, want) || calls != 2 || !reflect.DeepEqual(result.Items, items) {
				t.Fatalf("items=%v calls=%d err=%v", result.Items, calls, err)
			}
			if _, ok, err := pager.Next(t.Context()); !errors.Is(err, want) || ok || calls != 2 {
				t.Fatal("limit failure allowed data loss on resume")
			}
		})
	}
}

func TestPaginationCancellationReturnsCollectedPrefix(t *testing.T) {
	t.Parallel()
	entered := make(chan struct{})
	calls := 0
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return paginationResponse(r, 200, `{"value":[1],"@odata.nextLink":"https://graph.microsoft.com/v1.0/fixture?cursor=2"}`), nil
		}
		close(entered)
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}, "commercial")
	pager, err := graphapi.NewPager[int](scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, client.PagingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan struct{})
	var result client.ListResult[int]
	var got error
	go func() { defer close(done); result, got = pager.Collect(ctx) }()
	<-entered
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("cancellation stalled")
	}
	if !errors.Is(got, context.Canceled) || !reflect.DeepEqual(result.Items, []int{1}) {
		t.Fatal("cancelled partial result lost")
	}
}

func TestDecodeCollectionRequiresArrayEvenForBytes(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`{"value":"c2VjcmV0"}`, `{"value":null}`, `{"value":{}}`, `{}`, `null`} {
		if _, err := client.DecodeCollection[byte]([]byte(body), "value", "@odata.nextLink"); !errors.Is(err, client.ErrDecode) {
			t.Fatal("non-array accepted")
		}
	}
	page, err := client.DecodeCollection[byte]([]byte(`{"value":[1,2],"@odata.nextLink":null}`), "value", "@odata.nextLink")
	if err != nil || !reflect.DeepEqual(page.Items, []byte{1, 2}) {
		t.Fatal(err)
	}
}

func TestIndependentPagersDoNotShareTraversalState(t *testing.T) {
	t.Parallel()
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) { return paginationResponse(r, 200, `{"value":[1]}`), nil })}, "commercial")
	var group sync.WaitGroup
	for range 10 {
		group.Go(func() {
			pager, err := graphapi.NewPager[int](scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, client.PagingOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			result, err := pager.Collect(t.Context())
			if err != nil || len(result.Items) != 1 {
				t.Error("independent traversal failed")
			}
		})
	}
	group.Wait()
}

func TestTeamsTokenLoopsStopBeforeAnotherCredentialRequest(t *testing.T) {
	t.Parallel()
	calls := 0
	scope, _, provider := transportScope(t, client.TeamsAdmin, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return paginationResponse(r, 200, `{"agents":[1],"continuationToken":"repeat"}`), nil
	})}, "commercial")
	pager, err := teamsadminapi.NewTokenPager[int](scope, client.Request{Method: http.MethodGet, Path: []string{"fixture", "agents"}, Replay: client.ReplaySafe}, teamsadminapi.TokenPageFormat{ItemsField: "agents", TokenField: "continuationToken", QueryParameter: "continuationToken"}, client.PagingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := pager.Collect(t.Context())
	if !errors.Is(err, client.ErrContinuationLoop) || calls != 2 || len(result.Items) != 2 {
		t.Fatal("token loop was not bounded")
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if len(provider.requests) != 2 {
		t.Fatal("loop acquired extra token")
	}
}

func TestPaginationConfigurationAndOffsetOverflow(t *testing.T) {
	t.Parallel()
	calls := 0
	scope, _, _ := transportScope(t, client.TeamsAdmin, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) { calls++; return paginationResponse(r, 200, `[1]`), nil })}, "commercial")
	request := client.Request{Method: http.MethodGet, Path: []string{"fixture"}, Replay: client.ReplaySafe}
	for _, query := range []url.Values{{"skip": {"-1"}}, {"skip": {"0", "1"}}, {"skip": {"not-an-offset"}}} {
		r := request
		r.Query = query
		if _, err := teamsadminapi.NewOffsetPager[int](scope, r, "", "skip", client.PagingOptions{}); !errors.Is(err, client.ErrPagination) {
			t.Fatal("ambiguous offset accepted")
		}
	}
	unsafe := request
	unsafe.Replay = client.ReplayNever
	if _, err := teamsadminapi.NewOffsetPager[int](scope, unsafe, "", "skip", client.PagingOptions{}); !errors.Is(err, client.ErrPagination) {
		t.Fatal("unsafe operation accepted")
	}
	if calls != 0 {
		t.Fatal("constructor sent request")
	}
	request.Query = url.Values{"skip": {"9223372036854775807"}}
	pager, err := teamsadminapi.NewOffsetPager[int](scope, request, "", "skip", client.PagingOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := pager.Next(t.Context()); !errors.Is(err, client.ErrPagination) || calls != 1 {
		t.Fatal("offset overflow accepted")
	}
}
