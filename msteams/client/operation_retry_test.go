package client

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestUnitOperationMinimumAppliesInsideTransportRetries(t *testing.T) {
	for _, header := range []string{"", "1", "40"} {
		t.Run(header, func(t *testing.T) {
			calls := 0
			scope, controls := retryScope(t, RequestPolicy{}, nil, func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return retryResponse(r, 503, header), nil
				}
				return retryResponse(r, 200, ""), nil
			})
			waits := fakeRetryClock(controls)
			op, err := NewOperation(scope, Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: ReplaySafe}, "id", func([]byte) (OperationStatus, error) { return OperationStatus{State: OperationSucceeded}, nil }, 31*time.Second, nil)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := op.Poll(context.Background()); err != nil {
				t.Fatal(err)
			}
			want := 31 * time.Second
			if header == "40" {
				want = 40 * time.Second
			}
			if calls != 2 || !reflect.DeepEqual(*waits, []time.Duration{want}) {
				t.Fatalf("backend poll floor lost inside retry: %v", *waits)
			}
		})
	}
}

func TestUnitOperationMinimumAppliesToRejectedTokenReplacement(t *testing.T) {
	for _, header := range []string{"", "40"} {
		t.Run(header, func(t *testing.T) {
			binding := auth.Binding{TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture-client", Mode: auth.Application}
			acquisitions := 0
			provider, err := auth.NewRefreshProvider(binding, func(ctx context.Context, _ auth.TokenRequest) (auth.Token, error) {
				acquisitions++
				return auth.Token{AccessToken: fmt.Sprintf("fixture-%d", acquisitions), ExpiresAt: time.Now().Add(time.Hour)}, ctx.Err()
			}, auth.RefreshOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { provider.Close(); provider.Wait() }()
			calls := 0
			scope, controls := retryScope(t, RequestPolicy{}, provider, func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return retryResponse(r, 401, header), nil
				}
				return retryResponse(r, 200, ""), nil
			})
			waits := fakeRetryClock(controls)
			_, err = scope.Do(context.Background(), Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: ReplaySafe, RefreshOnUnauthorized: true, MinimumRetryDelay: 31 * time.Second}, nil)
			if err != nil {
				t.Fatal(err)
			}
			want := 31 * time.Second
			if header == "40" {
				want = 40 * time.Second
			}
			if calls != 2 || acquisitions != 2 || !reflect.DeepEqual(*waits, []time.Duration{want}) {
				t.Fatalf("replacement skipped polling floor: %v", *waits)
			}
		})
	}
}
