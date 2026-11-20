package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func postBrowserResult(ctx context.Context, raw string, deny bool, service *delegatedService) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	query := u.Query()
	service.mu.Lock()
	service.challenge = query.Get("code_challenge")
	service.mu.Unlock()
	if query.Get("code_challenge_method") != "S256" || query.Get("response_mode") != "form_post" || query.Get("state") == "" {
		return errors.New("missing PKCE/state/form-post")
	}
	form := url.Values{"state": {query.Get("state")}}
	if deny {
		form.Set("error", "access_denied")
		form.Set("error_description", "fixture-secret")
	} else {
		form.Set("code", "fixture-code")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, query.Get("redirect_uri"), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 || strings.Contains(string(content), "fixture-secret") {
		return errors.New("unsafe callback response")
	}
	return nil
}

func TestBrowserPKCEAndConcurrentAudiencesShareAccount(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t}
	cfg := delegatedConfig(Browser)
	var prompts atomic.Int32
	cfg.OpenBrowser = func(ctx context.Context, raw string) error {
		prompts.Add(1)
		return postBrowserResult(ctx, raw, false, service)
	}
	provider := newTestDelegated(t, cfg, service)
	var group sync.WaitGroup
	for _, audience := range []string{GraphAudience, TeamsAdminAudience, GraphAudience} {
		group.Go(func() {
			if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(audience)); err != nil {
				t.Errorf("browser acquisition failed: %v (%v)", err, errors.Unwrap(err))
			}
		})
	}
	group.Wait()
	if prompts.Load() != 1 || len(service.grants) != 2 {
		t.Fatal("concurrent audiences prompted for independent accounts")
	}
}

func TestBrowserDenialCallbackErrorAndUntrustedAuthority(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name               string
		deny, callbackFail bool
		authority          string
		want               error
	}{
		{"denied", true, false, "", ErrAuthenticationDenied},
		{"callback", false, true, "", ErrLoginCallback},
		{"authority", false, false, "https://attacker.invalid/authorize", ErrAuthorityDestination},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &delegatedService{t: t, authorize: tc.authority}
			cfg := delegatedConfig(Browser)
			var prompts int
			cfg.OpenBrowser = func(ctx context.Context, raw string) error {
				prompts++
				if tc.callbackFail {
					return errors.New("fixture-secret")
				}
				return postBrowserResult(ctx, raw, tc.deny, service)
			}
			provider := newTestDelegated(t, cfg, service)
			_, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience))
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v (%v), want %v", err, errors.Unwrap(err), tc.want)
			}
			if strings.Contains(fmt.Sprintf("%#v", err), "fixture-secret") || len(service.grants) != 0 {
				t.Fatal("failure leaked diagnostics or redeemed a code")
			}
			if tc.authority != "" && prompts != 0 {
				t.Fatal("untrusted authorization URL reached browser callback")
			}
		})
	}
}

func TestBrowserCancellationClosesSlowCallbackConnection(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t}
	cfg := delegatedConfig(Browser)
	started := make(chan struct{})
	var connection net.Conn
	cfg.OpenBrowser = func(ctx context.Context, raw string) error {
		u, err := url.Parse(raw)
		if err != nil {
			return err
		}
		redirect, err := url.Parse(u.Query().Get("redirect_uri"))
		if err != nil {
			return err
		}
		var dialer net.Dialer
		connection, err = dialer.DialContext(ctx, "tcp", redirect.Host)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(connection, "POST / HTTP/1.1\r\nHost: %s\r\nContent-Type: application/x-www-form-urlencoded\r\nContent-Length: 10000\r\n\r\nstate=", redirect.Host)
		close(started)
		return err
	}
	provider := newTestDelegated(t, cfg, service)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := provider.AcquireToken(ctx, provider.Binding().Request(GraphAudience)); result <- err }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("partial browser callback never started")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation not preserved: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("browser acquisition did not cancel")
	}
	provider.Close()
	done := make(chan struct{})
	go func() { provider.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("browser worker remained blocked on partial request body")
	}
	if connection == nil {
		t.Fatal("slow callback connection was not exercised")
	}
	_ = connection.Close()
}

func TestBrowserCallbackRequiresStateBeforeErrorOrCode(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, method, host, body string
		want                     int
		publish                  bool
	}{
		{"good", http.MethodPost, "localhost:1234", "state=expected&code=fixture", 200, true},
		{"bad-state-error", http.MethodPost, "localhost:1234", "state=wrong&error=access_denied&error_description=fixture-secret", 400, false},
		{"bad-state-code", http.MethodPost, "localhost:1234", "state=wrong&code=fixture", 400, false},
		{"duplicate-state", http.MethodPost, "localhost:1234", "state=expected&state=wrong&code=fixture", 400, false},
		{"host", http.MethodPost, "attacker.invalid", "state=expected&code=fixture", 400, false},
		{"get", http.MethodGet, "localhost:1234", "state=expected&code=fixture", 400, false},
		{"large", http.MethodPost, "localhost:1234", "state=expected&code=" + strings.Repeat("x", 17<<10), 400, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			results := make(chan browserResult, 1)
			request := httptest.NewRequestWithContext(t.Context(), tc.method, "http://localhost:1234/", strings.NewReader(tc.body))
			request.Host = tc.host
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()
			browserCallback("localhost:1234", "expected", results).ServeHTTP(response, request)
			if response.Code != tc.want || (len(results) == 1) != tc.publish || strings.Contains(response.Body.String(), "fixture-secret") {
				t.Fatal("callback accepted invalid state or exposed diagnostic text")
			}
		})
	}
}
