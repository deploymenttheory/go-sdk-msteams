package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type delegatedService struct {
	t                  *testing.T
	mu                 sync.Mutex
	grants             []url.Values
	requests           int
	failure            string
	errorCodes         []int
	user               string
	federated          bool
	federationMetadata string
	omitRefreshID      bool
	transportFail      bool
	authorize          string
	verification       string
	challenge          string
	deviceExpires      int
	pending            int
	extraScopes        map[string]bool
}

func (s *delegatedService) RoundTrip(r *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests++
	if s.transportFail {
		return nil, errors.New("fixture-secret transport failure")
	}
	if r.URL.Host != "login.microsoftonline.com" {
		s.t.Error("request escaped tenant authority")
	}
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}
	status := http.StatusOK
	var payload any
	switch {
	case strings.HasSuffix(r.URL.Path, "/.well-known/openid-configuration"):
		authorize := s.authorize
		if authorize == "" {
			authorize = strings.TrimSuffix(tokenURL, "token") + "authorize"
		}
		payload = map[string]any{"issuer": "https://login.microsoftonline.com/" + appTenant + "/v2.0", "token_endpoint": tokenURL, "authorization_endpoint": authorize}
	case r.URL.Path == "/common/discovery/instance":
		payload = map[string]any{"tenant_discovery_endpoint": "https://login.microsoftonline.com/" + appTenant + "/v2.0/.well-known/openid-configuration", "metadata": []any{map[string]any{"preferred_network": "login.microsoftonline.com", "preferred_cache": "login.microsoftonline.com", "aliases": []string{"login.microsoftonline.com"}}}}
	case strings.HasPrefix(r.URL.Path, "/common/UserRealm/"):
		accountType := "Managed"
		if s.federated {
			accountType = "Federated"
		}
		metadata := s.federationMetadata
		if metadata == "" {
			metadata = "https://untrusted.invalid/mex"
		}
		payload = map[string]any{"account_type": accountType, "federation_protocol": "WSTrust", "domain_name": "example.test", "cloud_instance_name": "login.microsoftonline.com", "cloud_audience_urn": "urn:federation:MicrosoftOnline", "federation_metadata_url": metadata, "federation_active_auth_url": "https://untrusted.invalid/password"}
	case strings.HasSuffix(r.URL.Path, "/devicecode"):
		verification := s.verification
		if verification == "" {
			verification = "https://microsoft.com/devicelogin"
		}
		expires := s.deviceExpires
		if expires == 0 {
			expires = 900
		}
		payload = map[string]any{"device_code": "fixture-private-code", "user_code": "FIXTURE", "verification_uri": verification, "expires_in": expires, "interval": 1, "message": "fixture instructions"}
	case r.URL.String() == tokenURL:
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		s.grants = append(s.grants, r.Form)
		if r.Form.Get("client_id") != appID {
			s.t.Error("wrong client ID")
		}
		scope := ""
		for _, part := range strings.Fields(r.Form.Get("scope")) {
			if strings.HasSuffix(part, "/.default") || s.extraScopes[part] {
				scope = part
			}
		}
		if scope != GraphAudience+"/.default" && scope != TeamsAdminAudience+"/.default" && !s.extraScopes[scope] {
			s.t.Error("wrong resource scope")
		}
		switch r.Form.Get("grant_type") {
		case "password":
			if r.Form.Get("username") != "fixture@example.test" || r.Form.Get("password") != "fixture-password" {
				s.t.Error("wrong user credentials")
			}
		case "authorization_code":
			sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
			if base64.RawURLEncoding.EncodeToString(sum[:]) != s.challenge || r.Form.Get("code") != "fixture-code" {
				s.t.Error("PKCE/code redemption mismatch")
			}
		case "urn:ietf:params:oauth:grant-type:saml2-bearer":
			assertion, err := base64.StdEncoding.DecodeString(r.Form.Get("assertion"))
			if err != nil || !strings.Contains(string(assertion), "fixture-assertion") {
				s.t.Error("federation assertion lost")
			}
		case "refresh_token":
			if r.Form.Get("refresh_token") != "fixture-refresh" {
				s.t.Error("refresh token lost")
			}
		case "device_code":
			if r.Form.Get("device_code") != "fixture-private-code" {
				s.t.Error("wrong device code")
			}
		default:
			s.t.Errorf("unexpected grant %q", r.Form.Get("grant_type"))
		}
		if s.pending > 0 {
			s.pending--
			status = 400
			payload = map[string]any{"error": "authorization_pending"}
			break
		}
		if s.failure != "" {
			status = 400
			payload = map[string]any{"error": s.failure, "error_description": "fixture-secret diagnostic", "error_codes": s.errorCodes}
			break
		}
		user := s.user
		if user == "" {
			user = "fixture-user"
		}
		info, _ := json.Marshal(map[string]string{"uid": user, "utid": appTenant})
		claims, _ := json.Marshal(map[string]any{"sub": user, "oid": user, "tid": appTenant, "iss": "https://login.microsoftonline.com/" + appTenant + "/v2.0", "aud": appID, "preferred_username": "fixture@example.test", "exp": time.Now().Add(time.Hour).Unix()})
		idToken := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." + base64.RawURLEncoding.EncodeToString(claims) + ".fixture"

		payload = map[string]any{"access_token": "fixture-access-" + scope, "expires_in": 3600, "token_type": "Bearer", "scope": r.Form.Get("scope"), "refresh_token": "fixture-refresh", "client_info": base64.RawURLEncoding.EncodeToString(info), "id_token": idToken}
		if s.omitRefreshID && r.Form.Get("grant_type") == "refresh_token" {
			delete(payload.(map[string]any), "id_token")
		}
	default:
		s.t.Errorf("unexpected identity route %s", r.URL.Path)
		return nil, errors.New("unexpected route")
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(string(content))), Request: r}, nil
}

func delegatedConfig(method Method) *Config {
	cfg := &Config{Method: method, TenantID: appTenant, ClientID: appID}
	switch method {
	case Password:
		cfg.Username = "fixture@example.test"
		cfg.Password = "fixture-password"
	case Browser:
		cfg.OpenBrowser = func(context.Context, string) error { return errors.New("unexpected browser callback") }
	case DeviceCode:
		cfg.DevicePrompt = func(context.Context, DeviceCodeMessage) error { return nil }
	}
	return cfg
}

func newTestDelegated(t *testing.T, cfg *Config, service *delegatedService) *DelegatedProvider {
	t.Helper()
	provider, err := NewDelegatedProvider(cfg, &http.Client{Transport: service, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { provider.Close(); provider.Wait() })
	if service.requests != 0 {
		t.Fatal("constructor performed network I/O")
	}
	return provider
}

func TestUnitDelegatedPasswordAndSilentSecondAudience(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t}
	cfg := delegatedConfig(Password)
	provider := newTestDelegated(t, cfg, service)
	cfg.Password = "mutated"
	for _, audience := range []string{GraphAudience, TeamsAdminAudience, GraphAudience} {
		token, err := provider.AcquireToken(t.Context(), provider.Binding().Request(audience))
		if err != nil || token.AccessToken != "fixture-access-"+audience+"/.default" {
			t.Fatalf("delegated acquisition failed: %v; cause: %v; calls: %d grants: %d", err, errors.Unwrap(err), service.requests, len(service.grants))
		}
	}
	if len(service.grants) != 2 || service.grants[0].Get("grant_type") != "password" || service.grants[1].Get("grant_type") != "refresh_token" {
		t.Fatal("second audience did not reuse the selected account and refresh token")
	}
}

func TestUnitDelegatedOAuthFailuresPreserveDiagnosticsWithoutFallback(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		code  string
		codes []int
		want  error
	}{
		{"invalid_grant", []int{50076}, ErrAuthenticationDenied},
		{"interaction_required", []int{50079}, ErrAuthenticationRequired},
		{"access_denied", []int{53003}, ErrAuthenticationDenied},
		{"consent_required", []int{65001}, ErrAuthenticationRequired},
		{"expired_token", nil, ErrDeviceCodeExpired},
	} {
		t.Run(tc.code, func(t *testing.T) {
			service := &delegatedService{t: t, failure: tc.code, errorCodes: tc.codes}
			provider := newTestDelegated(t, delegatedConfig(Password), service)
			_, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience))
			var failure *DelegatedError
			if !errors.Is(err, tc.want) || !errors.As(err, &failure) || failure.OAuthCode != tc.code {
				t.Fatalf("lost error classification: %v; cause: %v; calls: %d grants: %d", err, errors.Unwrap(err), service.requests, len(service.grants))
			}
			if len(failure.ErrorCodes) != len(tc.codes) {
				t.Fatal("Entra diagnostic codes lost")
			}
			if len(service.grants) != 1 {
				t.Fatal("denied password grant was repeated")
			}
			for _, format := range []string{"%v", "%+v", "%#v"} {
				if strings.Contains(fmt.Sprintf(format, err), "fixture-secret") {
					t.Fatal("raw OAuth diagnostics leaked")
				}
			}
		})
	}
}

func TestUnitDelegatedDeviceFlowAndCallback(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t, pending: 1}
	cfg := delegatedConfig(DeviceCode)
	var prompts atomic.Int32
	cfg.DevicePrompt = func(ctx context.Context, message DeviceCodeMessage) error {
		prompts.Add(1)
		if message.UserCode != "FIXTURE" || message.VerificationURL != "https://microsoft.com/devicelogin" {
			t.Error("prompt details lost")
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Error("device callback lacks expiration")
		}
		if strings.Contains(fmt.Sprintf("%#v", message), "FIXTURE") {
			t.Error("device code leaked in formatted prompt")
		}
		return nil
	}
	provider := newTestDelegated(t, cfg, service)
	// Real MSAL polling with an injected delay recorder keeps this offline test fast.
	var delays []time.Duration
	provider.authority.wait = func(ctx context.Context, d time.Duration) error { delays = append(delays, d); return ctx.Err() }
	if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience)); err != nil {
		t.Fatal(err)
	}
	if prompts.Load() != 1 || len(service.grants) != 2 || len(delays) != 2 || delays[0] <= 0 {
		t.Fatal("device pending/prompt/pacing contract lost")
	}
}

func TestUnitDelegatedDeviceRejections(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, verification, failure string
		expires                     int
		want                        error
	}{
		{"wrong-verification", "https://attacker.invalid/", "", 0, ErrAuthorityDestination},
		{"expired", "", "", -1, ErrDeviceCodeExpired},
		{"denied", "", "authorization_declined", 0, ErrAuthenticationDenied},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &delegatedService{t: t, verification: tc.verification, failure: tc.failure, deviceExpires: tc.expires}
			provider := newTestDelegated(t, delegatedConfig(DeviceCode), service)
			provider.authority.wait = func(ctx context.Context, _ time.Duration) error { return ctx.Err() }
			if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience)); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestUnitDelegatedAccountCannotChangeBetweenAudiences(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t}
	provider := newTestDelegated(t, delegatedConfig(Password), service)
	if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience)); err != nil {
		t.Fatal(err)
	}
	service.mu.Lock()
	service.user = "different-user"
	service.mu.Unlock()
	if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(TeamsAdminAudience)); !errors.Is(err, ErrBindingMismatch) {
		t.Fatalf("different delegated account accepted: %v", err)
	}
}

func TestUnitDelegatedCancellationWhileWaitingAndClose(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t}
	cfg := delegatedConfig(DeviceCode)
	entered := make(chan struct{})
	cfg.DevicePrompt = func(ctx context.Context, _ DeviceCodeMessage) error { close(entered); <-ctx.Done(); return ctx.Err() }
	provider := newTestDelegated(t, cfg, service)
	first := make(chan error, 1)
	go func() {
		_, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience))
		first <- err
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("device callback never started")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()
	if _, err := provider.AcquireToken(ctx, provider.Binding().Request(TeamsAdminAudience)); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting audience did not cancel: %v", err)
	}
	provider.Close()
	select {
	case err := <-first:
		if !errors.Is(err, ErrProviderClosed) {
			t.Fatalf("close did not reject waiting acquisition: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("close blocked on device callback")
	}
	provider.Wait()
	if provider.config != nil || provider.msal != nil || provider.account.HomeAccountID != "" {
		t.Fatal("owned credential/account references retained after close")
	}
}

func TestUnitDelegatedFederationDoesNotForwardPassword(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t, federated: true}
	provider := newTestDelegated(t, delegatedConfig(Password), service)
	if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience)); !errors.Is(err, ErrFederatedPassword) {
		t.Fatalf("unexpected federation failure: %v", err)
	}
	if len(service.grants) != 0 {
		t.Fatal("password sent after unapproved federation discovery")
	}
}

func TestUnitDelegatedSilentRefreshWithoutIDTokenKeepsSelectedAccount(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t, omitRefreshID: true}
	provider := newTestDelegated(t, delegatedConfig(Password), service)
	for _, audience := range []string{GraphAudience, TeamsAdminAudience} {
		if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(audience)); err != nil {
			t.Fatalf("refresh without a new ID token failed: %v (%v)", err, errors.Unwrap(err))
		}
	}
	if len(service.grants) != 2 || service.grants[1].Get("grant_type") != "refresh_token" {
		t.Fatal("missing account record forced a credential login")
	}
}

func TestUnitDelegatedSilentTransportFailureDoesNotPrompt(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t}
	cfg := delegatedConfig(DeviceCode)
	var prompts atomic.Int32
	cfg.DevicePrompt = func(context.Context, DeviceCodeMessage) error { prompts.Add(1); return nil }
	provider := newTestDelegated(t, cfg, service)
	provider.authority.wait = func(ctx context.Context, _ time.Duration) error { return ctx.Err() }
	if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience)); err != nil {
		t.Fatal(err)
	}
	service.mu.Lock()
	service.transportFail = true
	service.mu.Unlock()
	if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(TeamsAdminAudience)); !errors.Is(err, ErrDelegatedAuthentication) {
		t.Fatalf("unexpected silent transport error: %v", err)
	}
	if prompts.Load() != 1 {
		t.Fatal("transport failure unexpectedly launched a second login")
	}
}

func TestUnitDevicePollingHonorsIntervalAndSlowDown(t *testing.T) {
	t.Parallel()
	now := time.Unix(1000, 0)
	var delays []time.Duration
	var calls int
	transport := tokenRoundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Body != nil {
			_ = r.Body.Close()
		}
		calls++
		code := "authorization_pending"
		if calls == 1 {
			code = "slow_down"
		}
		return &http.Response{StatusCode: 400, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"error":"` + code + `"}`)), Request: r}, nil
	})
	authority := newDelegatedHTTP(&http.Client{Transport: transport}, appTenant, "")
	authority.base.canonicalTokenPath = "/" + appTenant + "/oauth2/v2.0/token"
	authority.now = func() time.Time { return now }
	authority.wait = func(ctx context.Context, d time.Duration) error {
		delays = append(delays, d)
		now = now.Add(d)
		return ctx.Err()
	}
	authority.startPolling(7)
	for range 2 {
		request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, tokenURL, strings.NewReader("fixture"))
		if err != nil {
			t.Fatal(err)
		}
		response, err := authority.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
	}
	if len(delays) != 2 || delays[0] != 7*time.Second || delays[1] != 12*time.Second {
		t.Fatalf("service interval or slow_down ignored: %v", delays)
	}
	if authority.failure() != nil {
		t.Fatal("pending response classified as terminal failure")
	}
	authority.stopPolling()
}

func TestUnitDelegatedConfigurationRequiresExplicitCallbacks(t *testing.T) {
	t.Parallel()
	for _, cfg := range []*Config{
		{Method: Browser, TenantID: appTenant, ClientID: appID},
		{Method: DeviceCode, TenantID: appTenant, ClientID: appID},
		{Method: Password, TenantID: appTenant, ClientID: appID, Username: "fixture", Password: "fixture", LoginTimeout: -time.Second},
		{Method: ClientSecret, TenantID: appTenant, ClientID: appID, ClientSecret: "fixture", OpenBrowser: func(context.Context, string) error { return nil }},
	} {
		if _, err := NewDelegatedProvider(cfg, &http.Client{}); !errors.Is(err, ErrInvalidConfig) {
			t.Fatal("invalid delegated configuration accepted")
		}
	}
}

func TestUnitDeviceCodeExpiresWhilePromptWaits(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t, deviceExpires: 1}
	cfg := delegatedConfig(DeviceCode)
	cfg.DevicePrompt = func(ctx context.Context, _ DeviceCodeMessage) error { <-ctx.Done(); return ctx.Err() }
	provider := newTestDelegated(t, cfg, service)
	if _, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience)); !errors.Is(err, ErrDeviceCodeExpired) {
		t.Fatalf("expired prompt did not classify expiry: %v (%v)", err, errors.Unwrap(err))
	}
	if len(service.grants) != 0 {
		t.Fatal("expired prompt started polling")
	}
}

func TestUnitDevicePromptCanCancelWithoutLeakingCallbackError(t *testing.T) {
	t.Parallel()
	service := &delegatedService{t: t}
	cfg := delegatedConfig(DeviceCode)
	cfg.DevicePrompt = func(context.Context, DeviceCodeMessage) error {
		return fmt.Errorf("fixture-secret: %w", context.Canceled)
	}
	provider := newTestDelegated(t, cfg, service)
	_, err := provider.AcquireToken(t.Context(), provider.Binding().Request(GraphAudience))
	if !errors.Is(err, context.Canceled) || strings.Contains(fmt.Sprintf("%#v", err), "fixture-secret") {
		t.Fatal("callback cancellation lost identity or leaked diagnostics")
	}
}
