package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type identityEndpoint struct {
	t                                                                *testing.T
	calls                                                            atomic.Int32
	label                                                            string
	expectedHost, expectedIdentity, identityKey, header, headerValue string
	status                                                           int
	wrongIdentity, wrongResource, expired, oversize, redirect        bool
	block                                                            chan struct{}
}

func (s *identityEndpoint) RoundTrip(r *http.Request) (*http.Response, error) {
	s.calls.Add(1)
	if r.URL.Host != s.expectedHost {
		s.t.Error("unexpected identity endpoint")
	}
	values := r.URL.Query()
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		_ = r.Body.Close()
		if err != nil {
			return nil, err
		}
		values, err = url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
	}
	if values.Get(s.identityKey) != s.expectedIdentity {
		s.t.Error("wrong identity selector")
	}
	if values.Get("claims") != "" {
		s.t.Error("cache isolation claims leaked onto local protocol")
	}
	if r.Header.Get(s.header) != s.headerValue {
		s.t.Error("missing platform header")
	}
	resource := values.Get("resource")
	if resource != GraphAudience && resource != TeamsAdminAudience {
		s.t.Error("incorrect managed resource scope")
	}
	if s.block != nil {
		close(s.block)
		<-r.Context().Done()
		return nil, r.Context().Err()
	}
	status := s.status
	if status == 0 {
		status = 200
	}
	expiry := time.Now().Add(time.Hour)
	if s.expired {
		expiry = time.Now().Add(-time.Hour)
	}
	payload := map[string]any{"access_token": s.label + "-" + resource, "expires_on": strconv.FormatInt(expiry.Unix(), 10), "token_type": "Bearer", "resource": resource}
	if s.expectedIdentity != "" {
		payload["client_id"] = s.expectedIdentity
	}
	if s.wrongIdentity {
		payload["client_id"] = "33333333-3333-4333-8333-333333333333"
	}
	if s.wrongResource {
		payload["resource"] = "other-resource"
	}
	if status >= 400 {
		payload = map[string]any{"error": "invalid_request", "error_description": "sensitive identity endpoint diagnostic"}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if s.oversize {
		raw = []byte(strings.Repeat("x", (1<<20)+1))
	}
	headers := http.Header{"Content-Type": {"application/json"}}
	if s.redirect {
		status = 302
		headers.Set("Location", "https://untrusted.invalid/token")
	}
	return &http.Response{StatusCode: status, Header: headers, Body: io.NopCloser(strings.NewReader(string(raw))), Request: r}, nil
}
func clearManagedEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range managedEnvironmentKeys {
		t.Setenv(key, "")
	}
}
func managedFixture(t *testing.T, cfg *Config, endpoint *identityEndpoint) *ManagedIdentityProvider {
	t.Helper()
	p, err := NewManagedIdentityProvider(cfg, &http.Client{Transport: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Close(); p.Wait() })
	return p
}
func TestManagedIdentityPlatformSelection(t *testing.T) {
	for _, platform := range []string{"imds-system", "imds-user", "app-service", "azure-ml", "cloud-shell"} {
		t.Run(platform, func(t *testing.T) {
			clearManagedEnvironment(t)
			service := &identityEndpoint{t: t, label: platform, expectedHost: "169.254.169.254", identityKey: "client_id", header: "Metadata", headerValue: "true"}
			cfg := &Config{Method: ManagedIdentity, TenantID: appTenant}
			switch platform {
			case "imds-user":
				cfg.ClientID = appID
				service.expectedIdentity = appID
			case "app-service":
				t.Setenv("IDENTITY_ENDPOINT", "http://identity.test/token")
				t.Setenv("IDENTITY_HEADER", "synthetic-host-key")
				cfg.ClientID = appID
				service.expectedIdentity = appID
				service.expectedHost = "identity.test"
				service.header = "X-IDENTITY-HEADER"
				service.headerValue = "synthetic-host-key"
			case "azure-ml":
				t.Setenv("MSI_ENDPOINT", "http://identity.test/token")
				t.Setenv("MSI_SECRET", "synthetic-host-key")
				t.Setenv("DEFAULT_IDENTITY_CLIENT_ID", appID)
				service.identityKey = "clientid"
				service.expectedIdentity = appID
				service.expectedHost = "identity.test"
				service.header = "secret"
				service.headerValue = "synthetic-host-key"
			case "cloud-shell":
				t.Setenv("MSI_ENDPOINT", "http://identity.test/token")
				service.expectedHost = "identity.test"
			}
			p := managedFixture(t, cfg, service)
			if service.calls.Load() != 0 {
				t.Fatal("constructor authenticated")
			}
			for _, audience := range []string{GraphAudience, TeamsAdminAudience, GraphAudience} {
				token, err := p.AcquireToken(t.Context(), p.Binding().Request(audience))
				if err != nil {
					t.Fatal(err)
				}
				if token.AccessToken != platform+"-"+audience {
					t.Fatal("wrong token")
				}
			}
			if service.calls.Load() != 2 {
				t.Fatalf("endpoint calls=%d", service.calls.Load())
			}
		})
	}
}
func TestManagedIdentityDoesNotReadAnotherSessionsGlobalCache(t *testing.T) {
	clearManagedEnvironment(t)
	for _, label := range []string{"first", "second"} {
		endpoint := &identityEndpoint{t: t, label: label, expectedHost: "169.254.169.254", identityKey: "client_id", header: "Metadata", headerValue: "true"}
		p := managedFixture(t, &Config{Method: ManagedIdentity, TenantID: label + ".example"}, endpoint)
		request := p.Binding().Request(GraphAudience)
		var workers sync.WaitGroup
		for range 20 {
			workers.Go(func() {
				token, err := p.AcquireToken(t.Context(), request)
				if err != nil {
					t.Error(err)
					return
				}
				if token.AccessToken != label+"-"+GraphAudience {
					t.Error("global token cache crossed sessions")
				}
			})
		}
		workers.Wait()
		if endpoint.calls.Load() != 1 {
			t.Fatalf("calls=%d", endpoint.calls.Load())
		}
		// Force the owned cache through its expiry boundary; the library cache must
		// not turn that into another cache hit at a different ownership boundary.
		p.refresh.mu.Lock()
		token := p.refresh.tokens[request]
		token.ExpiresAt = time.Now().Add(-time.Hour)
		p.refresh.tokens[request] = token
		p.refresh.mu.Unlock()
		if _, err := p.AcquireToken(t.Context(), request); err != nil {
			t.Fatal(err)
		}
		if endpoint.calls.Load() != 2 {
			t.Fatal("expired SDK token did not cause a fresh endpoint exchange")
		}
	}
}
func TestManagedIdentityEnvironmentChangeRejectsCachedAndFreshTokens(t *testing.T) {
	clearManagedEnvironment(t)
	service := &identityEndpoint{t: t, label: "original", expectedHost: "169.254.169.254", identityKey: "client_id", header: "Metadata", headerValue: "true"}
	p := managedFixture(t, &Config{Method: ManagedIdentity, TenantID: appTenant}, service)
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"IDENTITY_ENDPOINT", "IDENTITY_HEADER", "ProgramData"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "changed")
			for _, audience := range []string{GraphAudience, TeamsAdminAudience} {
				if _, err := p.AcquireToken(t.Context(), p.Binding().Request(audience)); !errors.Is(err, ErrIdentityEnvironmentChanged) {
					t.Fatal(err)
				}
			}
		})
	}
	if service.calls.Load() != 1 {
		t.Fatal("changed environment contacted endpoint")
	}
}
func TestManagedIdentityUnavailableAndRejectedResponses(t *testing.T) {
	for _, tc := range []struct {
		name                                                      string
		status                                                    int
		wrongIdentity, wrongResource, expired, oversize, redirect bool
		want                                                      error
	}{
		{name: "absent-endpoint", status: 404, want: ErrManagedIdentityUnavailable},
		{name: "identity-denied", status: 400, want: ErrManagedIdentityRejected},
		{name: "wrong-identity", wrongIdentity: true, want: ErrBindingMismatch},
		{name: "wrong-resource", wrongResource: true, want: ErrBindingMismatch},
		{name: "expired", expired: true, want: ErrTokenExpired},
		{name: "oversize", oversize: true, want: ErrManagedIdentityUnavailable},
		{name: "redirect", redirect: true, want: ErrManagedIdentityUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clearManagedEnvironment(t)
			service := &identityEndpoint{t: t, label: "sensitive-token", expectedHost: "169.254.169.254", expectedIdentity: appID, identityKey: "client_id", header: "Metadata", headerValue: "true", status: tc.status, wrongIdentity: tc.wrongIdentity, wrongResource: tc.wrongResource, expired: tc.expired, oversize: tc.oversize, redirect: tc.redirect}
			p := managedFixture(t, &Config{Method: ManagedIdentity, TenantID: appTenant, ClientID: appID}, service)
			token, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience))
			if !errors.Is(err, tc.want) || token.AccessToken != "" {
				t.Fatalf("got %v", err)
			}
			if service.calls.Load() != 1 {
				t.Fatal("identity request retried or redirected")
			}
			for current := err; current != nil; current = errors.Unwrap(current) {
				for _, text := range []string{fmt.Sprintf("%v", current), fmt.Sprintf("%#v", current)} {
					if strings.Contains(text, "sensitive") {
						t.Fatal("identity diagnostic leaked")
					}
				}
			}
		})
	}
}
func TestManagedIdentityCancellationAndClose(t *testing.T) {
	clearManagedEnvironment(t)
	entered := make(chan struct{})
	service := &identityEndpoint{t: t, label: "blocked", expectedHost: "169.254.169.254", identityKey: "client_id", header: "Metadata", headerValue: "true", block: entered}
	p := managedFixture(t, &Config{Method: ManagedIdentity, TenantID: appTenant}, service)
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
	if p.config != nil || len(p.environment) != 0 {
		t.Fatal("owned environment retained")
	}
}

func TestManagedIdentityBindingAndAudienceRejectBeforeDiscovery(t *testing.T) {
	clearManagedEnvironment(t)
	service := &identityEndpoint{t: t}
	p := managedFixture(t, &Config{Method: ManagedIdentity, TenantID: appTenant}, service)
	request := p.Binding().Request(GraphAudience)
	request.TenantID = "other"
	if _, err := p.AcquireToken(t.Context(), request); !errors.Is(err, ErrBindingMismatch) {
		t.Fatal(err)
	}
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request("other")); !errors.Is(err, ErrUnsupportedAudience) {
		t.Fatal(err)
	}
	if service.calls.Load() != 0 || p.source != "" {
		t.Fatal("invalid request performed discovery")
	}
}
