package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	graphapi "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api"
	graphbatch "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/batch"
	teamsadminapi "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api"
	"github.com/go-resty/resty/v2"
)

func TestTeamsBatchDetailsSurviveAggregateFailure(t *testing.T) {
	t.Parallel()
	body := []byte(`{"operationId":"fixture","overallStatus":"Complete","errorCount":1,"completedCount":2,"notStartedCount":0,"inProgressCount":0,"userState":[{"id":"one","state":"Completed","result":"Success"},{"id":"two","state":"Completed","result":"FutureFailureValue"}]}`)
	result, err := teamsadminapi.DecodeBatchPolicyResult(body)
	if !errors.Is(err, client.ErrOperationFailed) || len(result.Users) != 2 || result.Users[0].Result != "Success" || result.Users[1].Result != "FutureFailureValue" {
		t.Fatalf("partial Teams outcomes lost: %v", err)
	}
	result, err = teamsadminapi.DecodeBatchPolicyResult([]byte(`{"operationId":"fixture","overallStatus":"Complete","errorCount":0,"userState":[{"id":"one","state":"Completed","result":"Success"},{"id":"ONE","state":"Completed","result":"Success"}]}`))
	if !errors.Is(err, client.ErrDecode) || len(result.Users) != 2 {
		t.Fatal("duplicate user result falsely established success")
	}
}

func TestGraphBatchRetainsMixedOutcomesWithoutReplay(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		execute func(context.Context, *client.Scope, string, []graphapi.BatchRequest) ([]graphapi.BatchResponse, *resty.Response, error)
	}{
		{"compatibility-api", graphapi.ExecuteBatch},
		{"leaf-api", graphbatch.Execute},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != http.MethodPost || r.URL.Path != "/v1.0/$batch" {
					t.Fatal("wrong batch route")
				}
				var envelope struct {
					Requests []graphapi.BatchRequest `json:"requests"`
				}
				if json.NewDecoder(r.Body).Decode(&envelope) != nil || len(envelope.Requests) != 3 || string(envelope.Requests[0].Body) != `{"displayName":"fixture"}` {
					t.Fatal("batch request changed")
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"responses":[{"id":"B","status":429,"headers":{"Retry-After":"30"},"body":{"error":{"code":"TooManyRequests"}}},{"id":"a","status":201,"body":{"id":"created"}},{"id":"c","status":424,"body":{"error":{"code":"FailedDependency"}}}]}`)), Request: r}, nil
			})}, "")
			requests := []graphapi.BatchRequest{
				{ID: "a", Method: http.MethodPost, URL: "/teams", Headers: map[string]string{"Content-Type": "application/json"}, Body: json.RawMessage(`{"displayName":"fixture"}`)},
				{ID: "b", Method: http.MethodGet, URL: "/teams/fixture"},
				{ID: "c", Method: http.MethodGet, URL: "/teams/fixture/channels", DependsOn: []string{"b"}},
			}
			items, response, err := tc.execute(t.Context(), scope, "v1.0", requests)
			if !errors.Is(err, graphapi.ErrBatchItems) || response.StatusCode() != 200 || calls != 1 || len(items) != 3 {
				t.Fatalf("mixed outcomes lost: %v", err)
			}
			if items[0].Headers["Retry-After"] != "30" || items[1].Status != 201 || string(items[1].Body) != `{"id":"created"}` || items[2].Status != 424 {
				t.Fatal("item data/order/correlation lost")
			}
		})
	}
}

func TestGraphBatchValidatesBeforeAuthentication(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		version  string
		requests []graphapi.BatchRequest
	}{
		{"empty", "v1.0", nil},
		{"version", "v9", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "/teams"}}},
		{"duplicate", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "/teams"}, {ID: "A", Method: "GET", URL: "/teams"}}},
		{"foreign", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "https://foreign.test/teams"}}},
		{"authority", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "//foreign.test/teams"}}},
		{"traversal", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "/teams/%2e%2e/beta/users"}}},
		{"nested-version", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "/beta/users"}}},
		{"nested-batch", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "POST", URL: "/%24batch"}}},
		{"credentials", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "/teams", Headers: map[string]string{"Authorization": "fixture-secret"}}}},
		{"content-type", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "POST", URL: "/teams", Body: json.RawMessage(`{}`)}}},
		{"json", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "POST", URL: "/teams", Headers: map[string]string{"Content-Type": "application/json"}, Body: json.RawMessage(`{`)}}},
		{"dependency", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "/teams", DependsOn: []string{"missing"}}}},
		{"cycle", "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "/teams", DependsOn: []string{"b"}}, {ID: "b", Method: "GET", URL: "/teams", DependsOn: []string{"a"}}}},
		{"oversize", "v1.0", make([]graphapi.BatchRequest, 21)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scope, _, provider := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
				t.Fatal("invalid batch reached transport")
				return nil, nil
			})}, "")
			if _, _, err := graphapi.ExecuteBatch(t.Context(), scope, tc.version, tc.requests); !errors.Is(err, graphapi.ErrBatchRequest) {
				t.Fatalf("invalid batch accepted: %v", err)
			}
			if len(provider.requests) != 0 {
				t.Fatal("invalid batch acquired credentials")
			}
		})
	}
}

func TestGraphBatchMalformedResponsesPreserveKnownResults(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		`{"responses":[{"id":"a","status":200}]}`,
		`{"responses":[{"id":"a","status":200},{"id":"A","status":201}]}`,
		`{"responses":[{"id":"a","status":200},{"id":"unknown","status":200}]}`,
		`{"responses":[{"id":"a","status":200},{"id":"b","status":0}]}`,
		`{"responses":[{"id":"a","status":200},{"id":"b","status":"bad"}]}`,
	} {
		scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
		})}, "")
		items, response, err := graphapi.ExecuteBatch(t.Context(), scope, "beta", []graphapi.BatchRequest{{ID: "a", Method: "GET", URL: "/teams"}, {ID: "b", Method: "GET", URL: "/teams"}})
		if !errors.Is(err, graphapi.ErrBatchResponse) || len(items) != 1 || items[0].ID != "a" || response == nil {
			t.Fatalf("malformed batch lost known outcome: %v", err)
		}
	}
}

func TestGraphBatchOuterThrottleIsNotRetried(t *testing.T) {
	t.Parallel()
	calls := 0
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {"0"}}, Body: io.NopCloser(strings.NewReader(`{"error":{"code":"TooManyRequests"}}`)), Request: r}, nil
	})}, "")
	_, response, err := graphapi.ExecuteBatch(t.Context(), scope, "v1.0", []graphapi.BatchRequest{{ID: "a", Method: "POST", URL: "/teams"}})
	if err == nil || response.StatusCode() != 429 || calls != 1 {
		t.Fatal("outer batch replayed or throttle lost")
	}
}
