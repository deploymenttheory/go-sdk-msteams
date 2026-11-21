package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

type retryTransport func(*http.Request) (*http.Response, error)

func (f retryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type retryProvider struct{ calls atomic.Int32 }

func (p *retryProvider) AcquireToken(ctx context.Context, _ auth.TokenRequest) (auth.Token, error) {
	p.calls.Add(1)
	return auth.Token{AccessToken: "synthetic-access", ExpiresAt: time.Now().Add(time.Hour)}, ctx.Err()
}
func retryScope(t *testing.T, policy RequestPolicy, provider auth.Provider, transport retryTransport) (*Scope, *requestController) {
	t.Helper()
	if provider == nil {
		provider = &retryProvider{}
	}
	session, err := NewSession(&auth.Config{Method: auth.ExternalProvider, TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture-client", TokenMode: auth.Application, Provider: provider}, &http.Client{Transport: transport}, nil, WithRequestPolicy(policy))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(session.Close)
	scope, err := NewScope(Graph, session)
	if err != nil {
		t.Fatal(err)
	}
	return scope, session.requests
}
func retryResponse(r *http.Request, status int, after string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}, "Retry-After": {after}}, Body: io.NopCloser(strings.NewReader(`{"error":{"code":"fixtureError"}}`)), Request: r}
}
func fakeRetryClock(c *requestController) *[]time.Duration {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	waits := []time.Duration{}
	c.now = func() time.Time { return now }
	c.jitter = func(time.Duration) time.Duration { return 0 }
	c.sleep = func(ctx context.Context, d time.Duration) error {
		waits = append(waits, d)
		now = now.Add(d)
		return ctx.Err()
	}
	return &waits
}

func TestUnitRetryAfterParsingAndBoundedJitter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		input string
		want  time.Duration
		valid bool
	}{
		{"7", 7 * time.Second, true}, {"0", 0, true}, {" 4 ", 4 * time.Second, true},
		{now.Add(5 * time.Second).Format(http.TimeFormat), 5 * time.Second, true}, {now.Add(-time.Second).Format(http.TimeFormat), 0, true},
		{"99999999999999999999999999999", time.Duration(1<<63 - 1), true},
		{"", 0, false}, {"-1", 0, false}, {"1.5", 0, false}, {"+2", 0, false}, {"junk", 0, false},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, valid := retryAfter(tc.input, now)
			if got != tc.want || valid != tc.valid {
				t.Fatalf("got %v/%v", got, valid)
			}
		})
	}
	c, err := newRequestController([]SessionOption{WithRequestPolicy(RequestPolicy{BaseDelay: time.Second, MaxDelay: 4 * time.Second})})
	if err != nil {
		t.Fatal(err)
	}
	c.jitter = func(cap time.Duration) time.Duration { return cap - 1 }
	for i, want := range []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 4 * time.Second} {
		if got := c.delay("invalid", i+1); got != want-1 {
			t.Fatalf("delay %d: %v", i, got)
		}
	}
	if got := c.delay("120", 1); got != 2*time.Minute {
		t.Fatal("server minimum was capped")
	}
}

func TestUnitRetrySafeOperationReplaysFrozenJSONAndPreservesMetadata(t *testing.T) {
	calls := 0
	bodies := []string{}
	inputs := map[string]string{"value": "first"}
	scope, controls := retryScope(t, RequestPolicy{MaxAttempts: 4}, nil, func(r *http.Request) (*http.Response, error) {
		calls++
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		_ = r.Body.Close()
		bodies = append(bodies, string(payload))
		if r.Header.Get("If-Match") != "fixture-etag" {
			t.Error("conditional header lost")
		}
		if calls == 1 {
			inputs["value"] = "changed"
			return retryResponse(r, 429, "3"), nil
		}
		if calls == 2 {
			return retryResponse(r, 503, "invalid"), nil
		}
		return retryResponse(r, 200, ""), nil
	})
	waits := fakeRetryClock(controls)
	var result map[string]any
	response, err := scope.Do(context.Background(), Request{Method: http.MethodPost, Path: []string{"v1.0", "fixture-query"}, Body: inputs, Header: http.Header{"If-Match": {"fixture-etag"}}, Replay: ReplaySafe}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 3 || !reflect.DeepEqual(*waits, []time.Duration{3 * time.Second, 500 * time.Millisecond}) {
		t.Fatalf("calls=%d waits=%v", calls, *waits)
	}
	for _, body := range bodies {
		if body != `{"value":"first"}` {
			t.Fatal("retry remarshal changed payload")
		}
	}
	if response.Request.Header.Get("Authorization") != "" || response.RawResponse.Request.Header.Get("Authorization") != "" {
		t.Fatal("token retained in metadata")
	}
	if result == nil {
		t.Fatal("final response was not decoded")
	}
}

func TestUnitRetryRequiresOperationSafetyAndNeverReplaysStreams(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			calls := 0
			scope, _ := retryScope(t, RequestPolicy{}, nil, func(r *http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("ambiguous mutation; private diagnostic")
			})
			_, err := scope.Do(context.Background(), Request{Method: method, Path: []string{"v1.0", "fixture"}}, nil)
			if !errors.Is(err, ErrTransport) || calls != 1 || strings.Contains(err.Error(), "private") {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
		})
	}
	provider := &retryProvider{}
	scope, _ := retryScope(t, RequestPolicy{}, provider, func(*http.Request) (*http.Response, error) { t.Fatal("stream sent"); return nil, nil })
	_, err := scope.Do(context.Background(), Request{Method: http.MethodPost, Path: []string{"v1.0", "fixture"}, Body: strings.NewReader("non-rewindable"), Replay: ReplaySafe}, nil)
	if !errors.Is(err, ErrRequest) || provider.calls.Load() != 0 {
		t.Fatal("stream must fail before token acquisition")
	}
}

func TestUnitRetryAttemptsAndTotalDelayBudgets(t *testing.T) {
	for _, tc := range []struct {
		name, after  string
		budget       time.Duration
		calls, waits int
	}{
		{"attempts", "1", time.Minute, 3, 2}, {"server-delay", "120", time.Minute, 1, 0}, {"cumulative-delay", "4", 7 * time.Second, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			scope, c := retryScope(t, RequestPolicy{Timeout: tc.budget}, nil, func(r *http.Request) (*http.Response, error) { calls++; return retryResponse(r, 429, tc.after), nil })
			waits := fakeRetryClock(c)
			response, err := scope.Do(context.Background(), Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: ReplaySafe}, nil)
			var apiError *APIError
			if !errors.Is(err, ErrRetryBudget) || !errors.As(err, &apiError) || response.StatusCode() != 429 || calls != tc.calls || len(*waits) != tc.waits {
				t.Fatalf("calls=%d waits=%v err=%v", calls, *waits, err)
			}
		})
	}
}

func TestUnitRetryStatusClassification(t *testing.T) {
	for _, status := range []int{400, 401, 403, 404, 408, 409, 429, 500, 502, 503, 504} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			scope, c := retryScope(t, RequestPolicy{}, nil, func(r *http.Request) (*http.Response, error) { calls++; return retryResponse(r, status, "0"), nil })
			fakeRetryClock(c)
			_, err := scope.Do(context.Background(), Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: ReplaySafe}, nil)
			want := 1
			if status == 408 || status == 429 || status == 502 || status == 503 || status == 504 {
				want = 3
			}
			if err == nil || calls != want {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
		})
	}
}

func TestUnitRequestCancellationReleasesQueuedInflightAndBackoff(t *testing.T) {
	for _, phase := range []string{"queue", "inflight", "backoff"} {
		t.Run(phase, func(t *testing.T) {
			started := make(chan struct{})
			var once sync.Once
			scope, c := retryScope(t, RequestPolicy{MaxConcurrent: 1}, nil, func(r *http.Request) (*http.Response, error) {
				if phase == "inflight" {
					once.Do(func() { close(started) })
					<-r.Context().Done()
					return nil, r.Context().Err()
				}
				return retryResponse(r, 429, "1"), nil
			})
			if phase == "queue" {
				if err := c.acquire(context.Background()); err != nil {
					t.Fatal(err)
				}
				close(started)
			}
			if phase == "backoff" {
				c.sleep = func(ctx context.Context, _ time.Duration) error { close(started); <-ctx.Done(); return ctx.Err() }
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := scope.Do(ctx, Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: ReplaySafe}, nil)
				done <- err
			}()
			<-started
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("cancellation stalled")
			}
			if phase == "queue" {
				c.release()
			}
			if len(c.slots) != 0 {
				t.Fatal("permit leaked")
			}
		})
	}
}

func TestUnitRequestBudgetIncludesAuthenticationAndQueue(t *testing.T) {
	binding := auth.Binding{TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture-client", Mode: auth.Application}
	provider, err := auth.NewRefreshProvider(binding, func(ctx context.Context, _ auth.TokenRequest) (auth.Token, error) {
		<-ctx.Done()
		return auth.Token{}, ctx.Err()
	}, auth.RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { provider.Close(); provider.Wait() }()
	scope, _ := retryScope(t, RequestPolicy{Timeout: 20 * time.Millisecond}, provider, func(*http.Request) (*http.Response, error) { t.Fatal("backend called"); return nil, nil })
	_, err = scope.Do(context.Background(), Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	scope, c := retryScope(t, RequestPolicy{Timeout: 20 * time.Millisecond, MaxConcurrent: 1}, nil, func(*http.Request) (*http.Response, error) { t.Fatal("backend called"); return nil, nil })
	if err = c.acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer c.release()
	_, err = scope.Do(context.Background(), Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestUnitConcurrencySharedAcrossBackendsAndReleasedOnClose(t *testing.T) {
	entered := make(chan struct{}, 8)
	scope, c := retryScope(t, RequestPolicy{MaxConcurrent: 2}, nil, func(r *http.Request) (*http.Response, error) {
		entered <- struct{}{}
		<-r.Context().Done()
		return nil, r.Context().Err()
	})
	admin, err := NewScope(TeamsAdmin, scope.session)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 6)
	for i := 0; i < 6; i++ {
		target := scope
		if i%2 == 0 {
			target = admin
		}
		go func() {
			_, err := target.Do(context.Background(), Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil)
			done <- err
		}()
	}
	<-entered
	<-entered
	if len(c.slots) != 2 {
		t.Fatal("limit not occupied")
	}
	scope.session.Close()
	for i := 0; i < 6; i++ {
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) && !errors.Is(err, ErrClosed) {
				t.Fatal(err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("close stalled")
		}
	}
	if len(c.slots) != 0 || len(entered) != 0 {
		t.Fatal("concurrency limit exceeded or permit leaked")
	}
}

func TestUnitUnauthorizedRefreshIsOptInOnceAndWithinAttemptBudget(t *testing.T) {
	for _, tc := range []struct {
		name                                    string
		safe, refresh, capable                  bool
		attempts, status, wantCalls, wantGrants int
	}{
		{"success", true, true, true, 3, 200, 2, 2}, {"loop", true, true, true, 3, 401, 2, 2},
		{"unsafe", false, true, true, 3, 200, 1, 1}, {"disabled", true, false, true, 3, 200, 1, 1},
		{"unsupported", true, true, false, 3, 200, 1, 1}, {"budget", true, true, true, 1, 200, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var grants atomic.Int32
			binding := auth.Binding{TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture-client", Mode: auth.Application}
			refresh, err := auth.NewRefreshProvider(binding, func(ctx context.Context, _ auth.TokenRequest) (auth.Token, error) {
				n := grants.Add(1)
				value := "old"
				if n > 1 {
					value = "new"
				}
				return auth.Token{AccessToken: value, ExpiresAt: time.Now().Add(time.Hour)}, ctx.Err()
			}, auth.RefreshOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { refresh.Close(); refresh.Wait() }()
			var provider auth.Provider = refresh
			if !tc.capable {
				provider = acquireOnly{Provider: refresh}
			}
			calls := 0
			scope, _ := retryScope(t, RequestPolicy{MaxAttempts: tc.attempts}, provider, func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return retryResponse(r, 401, ""), nil
				}
				if r.Header.Get("Authorization") != "Bearer new" {
					t.Error("token not replaced")
				}
				return retryResponse(r, tc.status, ""), nil
			})
			replay := ReplayNever
			if tc.safe {
				replay = ReplaySafe
			}
			_, err = scope.Do(context.Background(), Request{Method: http.MethodPost, Path: []string{"v1.0", "fixture"}, Replay: replay, RefreshOnUnauthorized: tc.refresh}, nil)
			if tc.name == "success" && err != nil {
				t.Fatal(err)
			}
			if tc.name != "success" && err == nil {
				t.Fatal("expected unauthorized error")
			}
			if calls != tc.wantCalls || int(grants.Load()) != tc.wantGrants {
				t.Fatalf("calls=%d grants=%d", calls, grants.Load())
			}
		})
	}
}

type acquireOnly struct{ auth.Provider }

func TestUnitExpiredQueuedTokenIsReacquiredBeforeSending(t *testing.T) {
	var grants atomic.Int32
	acquired := make(chan struct{})
	binding := auth.Binding{TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture-client", Mode: auth.Application}
	var expires time.Time
	provider, err := auth.NewRefreshProvider(binding, func(ctx context.Context, _ auth.TokenRequest) (auth.Token, error) {
		n := grants.Add(1)
		if n == 1 {
			expires = time.Now().Add(100 * time.Millisecond)
			close(acquired)
			return auth.Token{AccessToken: "old", ExpiresAt: expires}, ctx.Err()
		}
		return auth.Token{AccessToken: "new", ExpiresAt: time.Now().Add(time.Hour)}, ctx.Err()
	}, auth.RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { provider.Close(); provider.Wait() }()
	calls := 0
	scope, c := retryScope(t, RequestPolicy{MaxConcurrent: 1}, provider, func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("Authorization") != "Bearer new" {
			t.Error("expired token sent")
		}
		return retryResponse(r, 200, ""), nil
	})
	if err = c.acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := scope.Do(context.Background(), Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil)
		done <- err
	}()
	<-acquired
	timer := time.NewTimer(time.Until(expires) + time.Millisecond)
	<-timer.C
	c.release()
	if err = <-done; err != nil {
		t.Fatal(err)
	}
	if calls != 1 || grants.Load() != 2 {
		t.Fatalf("calls=%d grants=%d", calls, grants.Load())
	}
}

func TestUnitPerAttemptTimeoutRetriesWithinOperationDeadline(t *testing.T) {
	calls := 0
	scope, c := retryScope(t, RequestPolicy{Timeout: time.Second}, nil, func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			<-r.Context().Done()
			return nil, r.Context().Err()
		}
		return retryResponse(r, 200, ""), nil
	})
	scope.session.http.Timeout = 20 * time.Millisecond
	fakeRetryClock(c)
	_, err := scope.Do(context.Background(), Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}, Replay: ReplaySafe}, nil)
	if err != nil || calls != 2 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
}

func TestUnitRequestPolicyValidation(t *testing.T) {
	for _, policy := range []RequestPolicy{{MaxAttempts: -1}, {MaxAttempts: 101}, {Timeout: -1}, {BaseDelay: -1}, {MaxDelay: time.Nanosecond}, {MaxConcurrent: -1}, {MaxConcurrent: 1025}} {
		if _, err := newRequestController([]SessionOption{WithRequestPolicy(policy)}); !errors.Is(err, ErrRequestPolicy) {
			t.Fatalf("accepted %#v", policy)
		}
	}
	if _, err := newRequestController([]SessionOption{nil}); !errors.Is(err, ErrRequestPolicy) {
		t.Fatal(err)
	}
}
