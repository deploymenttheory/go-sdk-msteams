package client_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestOperationAccessFailuresRemainDistinct(t *testing.T) {
	t.Parallel()
	for _, backend := range []client.Backend{client.Graph, client.TeamsAdmin} {
		for _, status := range []int{200, 401, 403, 404, 409, 500} {
			t.Run(fmt.Sprintf("%s/%d", backend, status), func(t *testing.T) {
				t.Parallel()
				// A missing-role fixture remains a forbidden result. Neither its text
				// nor a generic 404 can prove an unsupported mode or missing feature.
				scope, _, provider := transportScope(t, backend, &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(
						`{"error":{"code":"Authorization_RequestDenied","message":"missing role; unsupported auth mode; feature unavailable"}}`))}, nil
				})}, "commercial")
				path := []string{"fixture"}
				if backend == client.Graph {
					path = []string{"v1.0", "fixture"}
				}
				_, err := scope.Do(context.Background(), client.Request{Method: http.MethodGet, Path: path}, nil)
				if len(provider.requests) != 1 {
					t.Fatal("fixture did not authenticate")
				}
				if errors.Is(err, auth.ErrUnauthenticated) != (status == 401) || errors.Is(err, auth.ErrForbidden) != (status == 403) {
					t.Fatalf("wrong failure category: %v", err)
				}
				if errors.Is(err, auth.ErrUnsupportedOperationMode) || errors.Is(err, auth.ErrTenantCapabilityUnavailable) {
					t.Fatal("untrusted response text inferred a capability restriction")
				}
				if status == 200 && err != nil {
					t.Fatal(err)
				}
				if status != 200 {
					var apiErr *client.APIError
					if !errors.As(err, &apiErr) || apiErr.StatusCode != status {
						t.Fatal("lost HTTP diagnostics")
					}
				}
			})
		}
	}
	for _, cause := range []error{auth.ErrUnsupportedOperationMode, auth.ErrTenantCapabilityUnavailable} {
		err := fmt.Errorf("evidenced resource prerequisite: %w", cause)
		if !errors.Is(err, cause) || errors.Is(err, auth.ErrForbidden) || errors.Is(err, auth.ErrUnauthenticated) {
			t.Fatal("local prerequisite failure conflated with HTTP rejection")
		}
	}
}
