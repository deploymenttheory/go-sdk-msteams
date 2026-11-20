package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type workloadService struct {
	t                  *testing.T
	base               *tokenService
	mu                 sync.Mutex
	assertions, scopes []string
	reject             bool
}

func (s *workloadService) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Method != http.MethodPost || r.URL.String() != tokenURL {
		return s.base.RoundTrip(r)
	}
	defer func() { _ = r.Body.Close() }()
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	if r.Form.Get("client_secret") != "" || r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("client_id") != appID || r.Form.Get("client_assertion_type") != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
		s.t.Error("wrong workload grant")
	}
	assertion := r.Form.Get("client_assertion")
	if assertion == "" {
		s.t.Error("missing assertion")
	}
	scope := ""
	for _, v := range strings.Fields(r.Form.Get("scope")) {
		if strings.HasSuffix(v, "/.default") {
			scope = v
		}
	}
	s.mu.Lock()
	s.assertions = append(s.assertions, assertion)
	s.scopes = append(s.scopes, scope)
	s.mu.Unlock()
	status := 200
	payload := map[string]any{"access_token": "synthetic-workload-" + assertion, "expires_in": 3600, "token_type": "Bearer", "scope": scope}
	if s.reject {
		status = 400
		payload = map[string]any{"error": "invalid_client", "error_codes": []int{700024}, "error_description": "sensitive expired assertion " + assertion}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(string(raw))), Request: r}, nil
}
func workloadFixture(t *testing.T, cfg *Config) (*ApplicationProvider, *workloadService) {
	t.Helper()
	s := &workloadService{t: t, base: &tokenService{t: t}}
	p, err := NewApplicationProvider(cfg, &http.Client{Transport: s})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Close(); p.Wait() })
	return p, s
}
func TestWorkloadRotatedFileAndAudienceIsolation(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "assertion.jwt")
	if err := os.WriteFile(path, []byte("first-assertion\n"), 0600); err != nil {
		t.Fatal(err)
	}
	p, s := workloadFixture(t, &Config{Method: WorkloadIdentity, TenantID: appTenant, ClientID: appID, AssertionFile: path})
	if s.base.requests != 0 {
		t.Fatal("constructor contacted authority")
	}
	graph := p.Binding().Request(GraphAudience)
	teams := p.Binding().Request(TeamsAdminAudience)
	first, err := p.AcquireToken(t.Context(), graph)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("second-assertion"), 0600); err != nil {
		t.Fatal(err)
	}
	cached, err := p.AcquireToken(t.Context(), graph)
	if err != nil || cached.AccessToken != first.AccessToken {
		t.Fatal("cache not preserved")
	}
	if _, err := p.AcquireToken(t.Context(), teams); err != nil {
		t.Fatal(err)
	}
	p.refresh.mu.Lock()
	token := p.refresh.tokens[graph]
	token.ExpiresAt = time.Now().Add(-time.Second)
	p.refresh.tokens[graph] = token
	p.refresh.mu.Unlock()
	if _, err := p.AcquireToken(t.Context(), graph); err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.Join(s.assertions, ",") != "first-assertion,second-assertion,second-assertion" {
		t.Fatal("rotation not observed on refresh")
	}
	if strings.Join(s.scopes, ",") != GraphAudience+"/.default,"+TeamsAdminAudience+"/.default,"+GraphAudience+"/.default" {
		t.Fatal("audiences crossed")
	}
}
func TestWorkloadCallbackCoalescesAndRespectsBinding(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	p, s := workloadFixture(t, &Config{Method: WorkloadIdentity, TenantID: appTenant, ClientID: appID, Assertion: func(ctx context.Context) (string, error) { calls.Add(1); return "callback-assertion", ctx.Err() }})
	if calls.Load() != 0 {
		t.Fatal("constructor invoked callback")
	}
	request := p.Binding().Request(GraphAudience)
	wrong := request
	wrong.TenantID = "different"
	if _, err := p.AcquireToken(t.Context(), wrong); !errors.Is(err, ErrBindingMismatch) || calls.Load() != 0 {
		t.Fatal("binding not enforced")
	}
	var workers sync.WaitGroup
	for range 20 {
		workers.Go(func() {
			if _, err := p.AcquireToken(t.Context(), request); err != nil {
				t.Error(err)
			}
		})
	}
	workers.Wait()
	if calls.Load() != 1 {
		t.Fatalf("callbacks=%d", calls.Load())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.assertions) != 1 {
		t.Fatal("exchange not coalesced")
	}
}
func TestWorkloadInvalidAndExpiredAssertions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		source AssertionSource
		file   string
		reject bool
		want   error
	}{
		{name: "missing-file", file: filepath.Join(t.TempDir(), "absent"), want: ErrAssertionUnavailable},
		{name: "empty", source: func(context.Context) (string, error) { return " \n", nil }, want: ErrAssertionUnavailable},
		{name: "callback-error", source: func(context.Context) (string, error) { return "", errors.New("private callback diagnostic") }, want: ErrAssertionUnavailable},
		{name: "oversize", source: func(context.Context) (string, error) { return strings.Repeat("x", MaxAssertionBytes+1), nil }, want: ErrAssertionUnavailable},
		{name: "expired-at-issuer", source: func(context.Context) (string, error) { return "expired-private-assertion", nil }, reject: true, want: ErrTokenService},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, s := workloadFixture(t, &Config{Method: WorkloadIdentity, TenantID: appTenant, ClientID: appID, Assertion: tc.source, AssertionFile: tc.file})
			s.reject = tc.reject
			token, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience))
			if !errors.Is(err, tc.want) || token.AccessToken != "" {
				t.Fatalf("got %v", err)
			}
			for current := err; current != nil; current = errors.Unwrap(current) {
				if strings.Contains(current.Error(), "private") || strings.Contains(current.Error(), "sensitive") {
					t.Fatal("assertion diagnostic leaked")
				}
			}
		})
	}
}
func TestWorkloadCancellationAndOwnedClose(t *testing.T) {
	t.Parallel()
	entered := make(chan struct{})
	p, _ := workloadFixture(t, &Config{Method: WorkloadIdentity, TenantID: appTenant, ClientID: appID, Assertion: func(ctx context.Context) (string, error) { close(entered); <-ctx.Done(); return "", ctx.Err() }})
	done := make(chan error, 1)
	go func() {
		_, err := p.AcquireToken(context.Background(), p.Binding().Request(GraphAudience))
		done <- err
	}()
	<-entered
	p.Close()
	p.Wait()
	if err := <-done; !errors.Is(err, ErrProviderClosed) {
		t.Fatal(err)
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.config != nil {
		t.Fatal("owned callback reference retained")
	}
}

func TestWorkloadFileBoundsAndNonregularSources(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"empty", "oversize", "directory"} {
		t.Run(kind, func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "assertion")
			switch kind {
			case "empty":
				if err := os.WriteFile(path, nil, 0600); err != nil {
					t.Fatal(err)
				}
			case "oversize":
				if err := os.WriteFile(path, []byte(strings.Repeat("x", MaxAssertionBytes+1)), 0600); err != nil {
					t.Fatal(err)
				}
			case "directory":
				path = directory
			}
			p, s := workloadFixture(t, &Config{Method: WorkloadIdentity, TenantID: appTenant, ClientID: appID, AssertionFile: path})
			if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); !errors.Is(err, ErrAssertionUnavailable) {
				t.Fatal(err)
			}
			if s.base.requests != 0 {
				t.Fatal("invalid source contacted authority")
			}
		})
	}
	p, _ := workloadFixture(t, &Config{Method: WorkloadIdentity, TenantID: appTenant, ClientID: appID, Assertion: func(context.Context) (string, error) { return strings.Repeat(" ", MaxAssertionBytes) + "x", nil }})
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); !errors.Is(err, ErrAssertionUnavailable) {
		t.Fatal("raw callback limit bypassed by whitespace")
	}
}
