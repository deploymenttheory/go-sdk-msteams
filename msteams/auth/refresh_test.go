package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeClock struct{ nanos atomic.Int64 }

func (c *fakeClock) now() time.Time          { return time.Unix(0, c.nanos.Load()) }
func (c *fakeClock) advance(d time.Duration) { c.nanos.Add(int64(d)) }
func testBinding() Binding {
	return Binding{TenantID: "tenant", Cloud: "commercial", IdentityID: "identity", Mode: Application}
}
func await(t *testing.T, ready func() bool) {
	t.Helper()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for !ready() {
		select {
		case <-deadline.C:
			t.Fatal("condition did not become ready")
		case <-tick.C:
		}
	}
}
func waiters(p *RefreshProvider, request TokenRequest) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	if f := p.flights[request]; f != nil {
		return f.waiters
	}
	return 0
}
func cleanupProvider(t *testing.T, p *RefreshProvider) {
	t.Helper()
	t.Cleanup(func() { p.Close(); p.Wait() })
}

func TestZeroRefreshProviderIsClosed(t *testing.T) {
	t.Parallel()
	var provider RefreshProvider
	if _, err := provider.AcquireToken(context.Background(), TokenRequest{Audience: "graph"}); !errors.Is(err, ErrProviderClosed) {
		t.Fatalf("uninitialized provider should reject acquisition: %v", err)
	}
	provider.Close()
	provider.Close()
	provider.Wait()
}

func TestRefreshSingleFlightPerAudienceAndExpiry(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{}
	var graphCalls, teamsCalls atomic.Int32
	gate := make(chan struct{})
	binding := testBinding()
	provider, err := NewRefreshProvider(binding, func(ctx context.Context, request TokenRequest) (Token, error) {
		count := &graphCalls
		if request.Audience == "teams" {
			count = &teamsCalls
		}
		n := count.Add(1)
		select {
		case <-ctx.Done():
			return Token{}, ctx.Err()
		case <-gate:
		}
		return Token{AccessToken: fmt.Sprintf("%s-%d", request.Audience, n), ExpiresAt: clock.now().Add(time.Minute)}, nil
	}, RefreshOptions{RefreshSkew: 10 * time.Second, Clock: clock.now})
	if err != nil {
		t.Fatal(err)
	}
	cleanupProvider(t, provider)
	var group sync.WaitGroup
	for _, audience := range []string{"graph", "teams"} {
		for range 16 {
			group.Go(func() {
				token, err := provider.AcquireToken(context.Background(), binding.Request(audience))
				if err != nil || token.AccessToken != audience+"-1" {
					t.Errorf("shared token acquisition failed: %v", err)
				}
			})
		}
	}
	await(t, func() bool {
		return waiters(provider, binding.Request("graph")) == 16 && waiters(provider, binding.Request("teams")) == 16
	})
	close(gate)
	group.Wait()
	if graphCalls.Load() != 1 || teamsCalls.Load() != 1 {
		t.Fatal("refresh was not serialized by audience")
	}
	if _, err := provider.AcquireToken(context.Background(), binding.Request("graph")); err != nil {
		t.Fatal(err)
	}
	if graphCalls.Load() != 1 {
		t.Fatal("unexpired cache entry was refreshed")
	}
	clock.advance(51 * time.Second)
	token, err := provider.AcquireToken(context.Background(), binding.Request("graph"))
	if err != nil || token.AccessToken != "graph-2" || teamsCalls.Load() != 1 {
		t.Fatal("near-expiry refresh crossed audience boundary")
	}
}

func TestCancelledWaiterDoesNotCancelAnotherWaiter(t *testing.T) {
	t.Parallel()
	binding := testBinding()
	started := make(chan context.Context, 1)
	gate := make(chan struct{})
	provider, err := NewRefreshProvider(binding, func(ctx context.Context, _ TokenRequest) (Token, error) {
		started <- ctx
		select {
		case <-ctx.Done():
			return Token{}, ctx.Err()
		case <-gate:
			return Token{AccessToken: "token", ExpiresAt: time.Now().Add(time.Hour)}, nil
		}
	}, RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cleanupProvider(t, provider)
	first, cancel := context.WithCancel(context.Background())
	defer cancel()
	firstResult := make(chan error, 1)
	secondResult := make(chan error, 1)
	go func() { _, err := provider.AcquireToken(first, binding.Request("graph")); firstResult <- err }()
	shared := <-started
	go func() {
		_, err := provider.AcquireToken(context.Background(), binding.Request("graph"))
		secondResult <- err
	}()
	await(t, func() bool { return waiters(provider, binding.Request("graph")) == 2 })
	cancel()
	if err := <-firstResult; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled waiter: %v", err)
	}
	if shared.Err() != nil {
		t.Fatal("one cancelled waiter cancelled shared work")
	}
	close(gate)
	if err := <-secondResult; err != nil {
		t.Fatal(err)
	}
}

func TestAbandonedRefreshMustExitBeforeReplacement(t *testing.T) {
	t.Parallel()
	binding := testBinding()
	cancelled := make(chan struct{})
	finish := make(chan struct{})
	var finishOnce sync.Once
	var calls atomic.Int32
	provider, err := NewRefreshProvider(binding, func(ctx context.Context, _ TokenRequest) (Token, error) {
		n := calls.Add(1)
		if n == 1 {
			<-ctx.Done()
			close(cancelled)
			<-finish
		}
		return Token{AccessToken: fmt.Sprint(n), ExpiresAt: time.Now().Add(time.Hour)}, nil
	}, RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { provider.Close(); finishOnce.Do(func() { close(finish) }); provider.Wait() })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first := make(chan error, 1)
	go func() { _, err := provider.AcquireToken(ctx, binding.Request("graph")); first <- err }()
	await(t, func() bool { return calls.Load() == 1 })
	cancel()
	if err := <-first; !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	<-cancelled
	result := make(chan Token, 1)
	failed := make(chan error, 1)
	go func() {
		token, err := provider.AcquireToken(context.Background(), binding.Request("graph"))
		result <- token
		failed <- err
	}()
	provider.mu.Lock()
	active := provider.flights[binding.Request("graph")]
	abandoned := active != nil && active.abandoned
	provider.mu.Unlock()
	if !abandoned || calls.Load() != 1 {
		t.Fatal("abandoned refresh was removed before its callback exited")
	}
	finishOnce.Do(func() { close(finish) })
	if token := <-result; token.AccessToken != "2" {
		t.Fatal("abandoned token was reused")
	}
	if err := <-failed; err != nil {
		t.Fatal(err)
	}
}

func TestRefreshErrorsAreRedactedAndNotCached(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("sensitive-refresh-diagnostic")
	var calls atomic.Int32
	provider, err := NewRefreshProvider(testBinding(), func(context.Context, TokenRequest) (Token, error) { calls.Add(1); return Token{}, sentinel }, RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cleanupProvider(t, provider)
	for range 2 {
		token, err := provider.AcquireToken(context.Background(), testBinding().Request("graph"))
		if !errors.Is(err, sentinel) || token.AccessToken != "" {
			t.Fatal("callback error identity or zero-token contract was lost")
		}
		for _, message := range []string{fmt.Sprintf("%v", err), fmt.Sprintf("%+v", err), fmt.Sprintf("%#v", err)} {
			if strings.Contains(message, sentinel.Error()) {
				t.Fatal("callback diagnostic leaked through SDK error")
			}
		}
	}
	if calls.Load() != 2 {
		t.Fatal("refresh failure was cached")
	}
}

func TestRefreshRejectsWrongBindingAndInvalidTokens(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{}
	var calls atomic.Int32
	provider, err := NewRefreshProvider(testBinding(), func(context.Context, TokenRequest) (Token, error) {
		calls.Add(1)
		return Token{AccessToken: "token", ExpiresAt: clock.now().Add(time.Second)}, nil
	}, RefreshOptions{Clock: clock.now, RefreshSkew: 2 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	cleanupProvider(t, provider)
	request := testBinding().Request("graph")
	request.Mode = Delegated
	if _, err := provider.AcquireToken(context.Background(), request); !errors.Is(err, ErrBindingMismatch) {
		t.Fatal("wrong mode accepted")
	}
	if calls.Load() != 0 {
		t.Fatal("wrong binding reached callback")
	}
	for range 2 {
		if _, err := provider.AcquireToken(context.Background(), testBinding().Request("graph")); !errors.Is(err, ErrTokenExpiresSoon) {
			t.Fatalf("short-lived token accepted: %v", err)
		}
	}
	if calls.Load() != 2 {
		t.Fatal("invalid refresh response was cached")
	}
}

func TestCloseCancelsRefreshAndClearsCache(t *testing.T) {
	t.Parallel()
	started := make(chan struct{})
	exited := make(chan struct{})
	provider, err := NewRefreshProvider(testBinding(), func(ctx context.Context, _ TokenRequest) (Token, error) {
		close(started)
		defer close(exited)
		<-ctx.Done()
		return Token{}, ctx.Err()
	}, RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cleanupProvider(t, provider)
	result := make(chan error, 1)
	go func() {
		_, err := provider.AcquireToken(context.Background(), testBinding().Request("graph"))
		result <- err
	}()
	<-started
	provider.Close()
	provider.Wait()
	<-exited
	if err := <-result; !errors.Is(err, ErrProviderClosed) {
		t.Fatalf("closed refresh returned: %v", err)
	}
	if _, err := provider.AcquireToken(context.Background(), testBinding().Request("graph")); !errors.Is(err, ErrProviderClosed) {
		t.Fatal("closed provider accepted work")
	}
}

func TestRefreshRejectsInvalidCallbackResults(t *testing.T) {
	t.Parallel()
	clock := &fakeClock{}
	cases := []struct {
		name  string
		token Token
		want  error
	}{
		{name: "empty value", token: Token{ExpiresAt: clock.now().Add(time.Hour)}, want: ErrInvalidToken},
		{name: "zero expiry", token: Token{AccessToken: "token"}, want: ErrInvalidToken},
		{name: "expired", token: Token{AccessToken: "token", ExpiresAt: clock.now().Add(-time.Second)}, want: ErrTokenExpired},
		{name: "expires now", token: Token{AccessToken: "token", ExpiresAt: clock.now()}, want: ErrTokenExpired},
		{name: "expires soon", token: Token{AccessToken: "token", ExpiresAt: clock.now().Add(time.Second)}, want: ErrTokenExpiresSoon},
		{name: "window boundary", token: Token{AccessToken: "token", ExpiresAt: clock.now().Add(2 * time.Second)}, want: ErrTokenExpiresSoon},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			provider, err := NewRefreshProvider(testBinding(), func(context.Context, TokenRequest) (Token, error) {
				calls.Add(1)
				return tc.token, nil
			}, RefreshOptions{Clock: clock.now, RefreshSkew: 2 * time.Second})
			if err != nil {
				t.Fatal(err)
			}
			cleanupProvider(t, provider)
			for range 2 {
				token, err := provider.AcquireToken(context.Background(), testBinding().Request("graph"))
				if !errors.Is(err, tc.want) {
					t.Fatalf("expected %v, got %v", tc.want, err)
				}
				if token.AccessToken != "" || !token.ExpiresAt.IsZero() {
					t.Fatal("invalid callback result returned token material")
				}
			}
			if calls.Load() != 2 {
				t.Fatal("invalid callback result was cached")
			}
		})
	}
}

func TestRefreshRejectsInvalidConstructionOptions(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	callback := func(context.Context, TokenRequest) (Token, error) {
		calls.Add(1)
		return Token{}, nil
	}
	cases := []struct {
		name     string
		binding  Binding
		callback RefreshFunc
		options  RefreshOptions
		want     error
	}{
		{name: "invalid binding", callback: callback, want: ErrInvalidBinding},
		{name: "nil callback", binding: testBinding(), want: ErrInvalidRefresh},
		{name: "negative skew", binding: testBinding(), callback: callback,
			options: RefreshOptions{RefreshSkew: -time.Second}, want: ErrInvalidRefresh},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := NewRefreshProvider(tc.binding, tc.callback, tc.options)
			if provider != nil || !errors.Is(err, tc.want) {
				t.Fatalf("invalid provider options: expected %v, got %v", tc.want, err)
			}
		})
	}
	if calls.Load() != 0 {
		t.Fatal("construction invoked the refresh callback")
	}
}

func TestRefreshProviderDiagnosticsRedactPopulatedCache(t *testing.T) {
	t.Parallel()
	const secret = "sensitive-cached-access-token"
	provider, err := NewRefreshProvider(testBinding(), func(context.Context, TokenRequest) (Token, error) {
		return Token{AccessToken: secret, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}, RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cleanupProvider(t, provider)
	if _, err := provider.AcquireToken(context.Background(), testBinding().Request("graph")); err != nil {
		t.Fatal(err)
	}
	provider.mu.Lock()
	cached := len(provider.tokens)
	provider.mu.Unlock()
	if cached != 1 {
		t.Fatal("test must exercise a populated token cache")
	}
	for _, message := range []string{fmt.Sprintf("%v", provider), fmt.Sprintf("%+v", provider), fmt.Sprintf("%#v", provider)} {
		if strings.Contains(message, secret) || !strings.Contains(message, "redacted") {
			t.Fatal("refresh-provider diagnostics exposed cached state")
		}
	}
}
