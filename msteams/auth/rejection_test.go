package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRejectedTokenRefreshCoalescesAndKeepsOtherAudiences(t *testing.T) {
	binding := Binding{TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture-app", Mode: Application}
	var grants atomic.Int32
	provider, err := NewRefreshProvider(binding, func(ctx context.Context, r TokenRequest) (Token, error) {
		n := grants.Add(1)
		return Token{AccessToken: fmt.Sprintf("synthetic-%s-%d", r.Audience, n), ExpiresAt: time.Now().Add(time.Hour)}, ctx.Err()
	}, RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { provider.Close(); provider.Wait() }()
	graph := binding.Request(GraphAudience)
	teams := binding.Request(TeamsAdminAudience)
	old, err := provider.AcquireToken(context.Background(), graph)
	if err != nil {
		t.Fatal(err)
	}
	other, err := provider.AcquireToken(context.Background(), teams)
	if err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for range 20 {
		workers.Go(func() {
			token, err := provider.RefreshAfterRejection(context.Background(), graph, old)
			if err != nil || token.AccessToken == old.AccessToken {
				t.Error("rejected token reused")
			}
		})
	}
	workers.Wait()
	if grants.Load() != 3 {
		t.Fatalf("grants=%d", grants.Load())
	}
	late, err := provider.RefreshAfterRejection(context.Background(), graph, old)
	if err != nil || late.AccessToken == old.AccessToken || grants.Load() != 3 {
		t.Fatal("late rejection evicted replacement")
	}
	same, err := provider.AcquireToken(context.Background(), teams)
	if err != nil || same.AccessToken != other.AccessToken {
		t.Fatal("other audience cache changed")
	}
	wrong := graph
	wrong.TenantID = "other"
	if _, err = provider.RefreshAfterRejection(context.Background(), wrong, old); !errors.Is(err, ErrBindingMismatch) {
		t.Fatal(err)
	}
	provider.Close()
	if _, err = provider.RefreshAfterRejection(context.Background(), graph, old); !errors.Is(err, ErrProviderClosed) {
		t.Fatal(err)
	}
}
