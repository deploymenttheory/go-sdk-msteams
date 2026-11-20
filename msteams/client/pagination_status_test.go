package client_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestPagerExpectedStatusRejectsUnexpectedPageBeforeDecode(t *testing.T) {
	t.Parallel()
	for _, status := range []int{201, 202, 204, 206} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls, decodes := 0, 0
			scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return paginationResponse(r, status, `{"value":[2]}`), nil
			})}, "commercial")
			pager, err := client.NewPager(scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: client.ReplaySafe}, func(body []byte) (client.PageContent[int], error) {
				decodes++
				return client.DecodeCollection[int](body, "value", "@odata.nextLink")
			}, client.PagingOptions{Style: client.LinkPaging, ExpectedStatus: http.StatusOK})
			if err != nil {
				t.Fatal(err)
			}
			result, err := pager.Collect(t.Context())
			if !errors.Is(err, client.ErrPageStatus) || len(result.Items) != 0 || result.Pages != 0 || result.Response == nil || result.Response.StatusCode() != status || calls != 1 || decodes != 0 {
				t.Fatal("unexpected status became a decoded or complete page")
			}
			if _, _, err := pager.Next(t.Context()); !errors.Is(err, client.ErrPageStatus) || calls != 1 {
				t.Fatal("failed traversal resumed HTTP")
			}
		})
	}
}

func TestPagerExpectedStatusRetainsEarlierItemsAndLastResponse(t *testing.T) {
	t.Parallel()
	calls := 0
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return paginationResponse(r, 200, `{"value":[1],"@odata.nextLink":"https://graph.microsoft.com/v1.0/fixture?$skiptoken=next"}`), nil
		}
		return paginationResponse(r, 206, `{"value":[2]}`), nil
	})}, "commercial")
	pager, err := client.NewPager(scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: client.ReplaySafe}, func(body []byte) (client.PageContent[int], error) {
		return client.DecodeCollection[int](body, "value", "@odata.nextLink")
	}, client.PagingOptions{Style: client.LinkPaging, ExpectedStatus: 200})
	if err != nil {
		t.Fatal(err)
	}
	result, err := pager.Collect(t.Context())
	if !errors.Is(err, client.ErrPageStatus) || len(result.Items) != 1 || result.Items[0] != 1 || result.Pages != 1 || result.Response.StatusCode() != 206 || calls != 2 {
		t.Fatal("partial prefix or unexpected response was lost")
	}
}

func TestPagerStatusOptionPreservesGenericAndHTTPFailureBehavior(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name             string
		status, expected int
		want             error
	}{
		{"default-compatible", 202, 0, nil},
		{"explicit-success", 201, 201, nil},
		{"forbidden-stays-forbidden", 403, 200, auth.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				return paginationResponse(r, tc.status, `{"value":[],"error":{"code":"Forbidden"}}`), nil
			})}, "commercial")
			pager, err := client.NewPager(scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: client.ReplaySafe}, func(body []byte) (client.PageContent[int], error) {
				return client.DecodeCollection[int](body, "value", "@odata.nextLink")
			}, client.PagingOptions{Style: client.LinkPaging, ExpectedStatus: tc.expected})
			if err != nil {
				t.Fatal(err)
			}
			result, err := pager.Collect(t.Context())
			if !errors.Is(err, tc.want) || result.Response == nil || result.Response.StatusCode() != tc.status {
				t.Fatal("generic or typed HTTP error behavior changed")
			}
		})
	}
}

func TestPagerExpectedStatusValidationIsCredentialFree(t *testing.T) {
	t.Parallel()
	scope, _, provider := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		t.Error("invalid constructor sent HTTP")
		return nil, client.ErrRequest
	})}, "commercial")
	for _, status := range []int{-1, 100, 199, 300, 403, 999} {
		if _, err := client.NewPager(scope, client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: client.ReplaySafe}, func(body []byte) (client.PageContent[int], error) {
			return client.DecodeCollection[int](body, "value", "@odata.nextLink")
		}, client.PagingOptions{Style: client.LinkPaging, ExpectedStatus: status}); !errors.Is(err, client.ErrPagination) {
			t.Fatal("non-success expected status accepted")
		}
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if len(provider.requests) != 0 {
		t.Fatal("constructor validation acquired credentials")
	}
}
