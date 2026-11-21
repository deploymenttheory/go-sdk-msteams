package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const appTenant = "11111111-1111-4111-8111-111111111111"
const appID = "22222222-2222-4222-8222-222222222222"
const tokenURL = "https://login.microsoftonline.com/" + appTenant + "/oauth2/v2.0/token"

type tokenService struct {
	issuer       string
	redirect     bool
	metadataPath string
	t            *testing.T
	mu           sync.Mutex
	scopes       map[string]int
	requests     int
	public       *rsa.PublicKey
	fail         bool
	endpoint     string
	ttl          int
	block        chan struct{}
	entered      chan struct{}
	extraScopes  map[string]bool
}

func (s *tokenService) RoundTrip(r *http.Request) (*http.Response, error) {
	s.mu.Lock()
	s.requests++
	s.mu.Unlock()
	if r.URL.Host != "login.microsoftonline.com" {
		s.t.Error("request escaped authority")
	}
	var payload any
	status := http.StatusOK
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/.well-known/openid-configuration"):
		if s.redirect {
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"https://other.test/token"}}, Body: io.NopCloser(strings.NewReader("redirect")), Request: r}, nil
		}
		if s.metadataPath != "" && r.URL.Path != s.metadataPath {
			s.t.Error("metadata did not stay bound to configured tenant")
		}
		issuer := s.issuer
		if issuer == "" {
			issuer = "https://login.microsoftonline.com/" + appTenant + "/v2.0"
		}
		endpoint := s.endpoint
		if endpoint == "" {
			endpoint = tokenURL
		}
		payload = map[string]any{"authorization_endpoint": "https://login.microsoftonline.com/" + appTenant + "/oauth2/v2.0/authorize",
			"token_endpoint": endpoint, "issuer": issuer, "jwks_uri": "https://login.microsoftonline.com/common/discovery/v2.0/keys"}
	case r.Method == http.MethodGet && r.URL.Path == "/common/discovery/instance":
		payload = map[string]any{"tenant_discovery_endpoint": "https://login.microsoftonline.com/" + appTenant + "/v2.0/.well-known/openid-configuration",
			"metadata": []any{map[string]any{"preferred_network": "login.microsoftonline.com", "preferred_cache": "login.windows.net", "aliases": []string{"login.microsoftonline.com", "login.windows.net"}}}}
	case r.Method == http.MethodPost && r.URL.String() == tokenURL:
		defer func() {
			if err := r.Body.Close(); err != nil {
				s.t.Error("could not close token request body")
			}
		}()
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
		scope := ""
		defaults := map[string]bool{"openid": false, "offline_access": false, "profile": false}
		for _, part := range strings.Fields(r.Form.Get("scope")) {
			if _, ok := defaults[part]; ok {
				if defaults[part] {
					s.t.Error("duplicate library default scope")
				}
				defaults[part] = true
				continue
			}
			if scope != "" {
				s.t.Error("multiple resource scopes in one exchange")
			}
			scope = part
		}
		if r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("client_id") != appID {
			s.t.Error("incorrect credential grant identity")
		}
		if scope != GraphAudience+"/.default" && scope != TeamsAdminAudience+"/.default" && !s.extraScopes[scope] {
			s.t.Errorf("incorrect resource scope: %q", scope)
		}
		if s.public == nil {
			if r.Form.Get("client_secret") != "synthetic-secret" {
				s.t.Error("wrong copied client secret")
			}
		} else {
			if r.Form.Get("client_secret") != "" || r.Form.Get("client_assertion_type") != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" {
				s.t.Error("incorrect certificate grant")
			}
			s.verifyAssertion(r.Form.Get("client_assertion"))
		}
		s.mu.Lock()
		s.scopes[scope]++
		count := s.scopes[scope]
		ttl := s.ttl
		s.mu.Unlock()
		if s.entered != nil {
			s.entered <- struct{}{}
		}
		if s.block != nil {
			select {
			case <-s.block:
			case <-r.Context().Done():
				return nil, r.Context().Err()
			}
		}
		if s.fail {
			status = 400
			payload = map[string]any{"error": "invalid_client", "error_description": "sensitive-server-diagnostic synthetic-secret"}
		} else {
			if ttl == 0 {
				ttl = 3600
			}
			payload = map[string]any{"token_type": "Bearer", "access_token": fmt.Sprintf("%s-token-%d", scope, count), "expires_in": ttl, "scope": scope}
		}
	default:
		return nil, errors.New("unexpected token-service request")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(string(raw))), Request: r}, nil
}
func (s *tokenService) verifyAssertion(assertion string) {
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		s.t.Error("missing signed assertion")
		return
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		s.t.Error("invalid assertion signature encoding")
		return
	}
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPSS(s.public, crypto.SHA256, hash[:], signature, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}); err != nil {
		s.t.Error("MSAL assertion did not verify")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		s.t.Error("invalid assertion claims")
		return
	}
	var claims map[string]any
	if err := json.Unmarshal(raw, &claims); err != nil {
		s.t.Error("invalid assertion JSON")
		return
	}
	if claims["iss"] != appID || claims["sub"] != appID || claims["aud"] != tokenURL || claims["jti"] == "" {
		s.t.Error("assertion binding mismatch")
	}
}
func (s *tokenService) count(scope string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.scopes[scope+"/.default"]
}
func newTokenService(t *testing.T) *tokenService {
	return &tokenService{t: t, scopes: map[string]int{}}
}
func appConfig() *Config {
	return &Config{Method: ClientSecret, TenantID: appTenant, ClientID: appID, ClientSecret: "synthetic-secret"}
}
func application(t *testing.T, cfg *Config, service *tokenService) *ApplicationProvider {
	t.Helper()
	p, err := NewApplicationProvider(cfg, &http.Client{Transport: service})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Close(); p.Wait() })
	return p
}

func TestApplicationMSALCredentialModesAndAudiences(t *testing.T) {
	t.Parallel()
	pfx, err := os.ReadFile("testdata/client-credentials-sha1.pfx")
	if err != nil {
		t.Fatal(err)
	}
	certs, key, err := azidentity.ParseCertificates(pfx, []byte("fixture-only"))
	if err != nil {
		t.Fatal(err)
	}
	rawKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	configs := []*Config{appConfig(),
		{Method: Certificate, TenantID: appTenant, ClientID: appID, PFX: pfx, PFXPassword: []byte("fixture-only")},
		{Method: Certificate, TenantID: appTenant, ClientID: appID, Certificate: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certs[0].Raw}), PrivateKey: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: rawKey})},
	}
	for i, cfg := range configs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			service := newTokenService(t)
			audiences := []string{GraphAudience, TeamsAdminAudience, "https://retailservices.teams.microsoft.com", "https://chatsvcagg.teams.microsoft.com"}
			service.extraScopes = map[string]bool{"https://retailservices.teams.microsoft.com/.default": true, "https://chatsvcagg.teams.microsoft.com/.default": true}
			if cfg.Method == Certificate {
				service.public = certs[0].PublicKey.(*rsa.PublicKey)
			}
			p := application(t, cfg, service)
			cfg.ClientSecret = "changed-after-construction"
			if service.requests != 0 {
				t.Fatal("constructor authenticated")
			}
			var group sync.WaitGroup
			for _, audience := range audiences {
				for range 8 {
					group.Go(func() {
						token, err := p.AcquireToken(t.Context(), p.Binding().Request(audience))
						if err != nil || token.AccessToken != audience+"/.default-token-1" {
							t.Errorf("acquisition failed: %v", err)
						}
					})
				}
			}
			group.Wait()
			for _, audience := range audiences {
				if service.count(audience) != 1 {
					t.Fatal("audiences were not independently cached and serialized")
				}
			}
		})
	}
}

func TestApplicationRefreshActuallyExchangesNewCredentials(t *testing.T) {
	t.Parallel()
	service := newTokenService(t)
	p := application(t, appConfig(), service)
	clock := &fakeClock{}
	clock.nanos.Store(time.Now().UnixNano())
	p.refresh.now = clock.now
	request := p.Binding().Request(GraphAudience)
	if _, err := p.AcquireToken(t.Context(), request); err != nil {
		t.Fatal(err)
	}
	clock.advance(59 * time.Minute)
	service.mu.Lock()
	service.ttl = 7200
	service.mu.Unlock()
	token, err := p.AcquireToken(t.Context(), request)
	if err != nil || !strings.HasSuffix(token.AccessToken, "-2") || service.count(GraphAudience) != 2 {
		t.Fatalf("refresh reused an inner MSAL token cache: %v", err)
	}
}

func TestApplicationRejectsBindingCloudAndUnsafeTenant(t *testing.T) {
	t.Parallel()
	service := newTokenService(t)
	p := application(t, appConfig(), service)
	wrong := p.Binding().Request(GraphAudience)
	wrong.Mode = Delegated
	if _, err := p.AcquireToken(t.Context(), wrong); !errors.Is(err, ErrBindingMismatch) {
		t.Fatalf("wrong mode: %v", err)
	}
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request("https://other.test")); !errors.Is(err, ErrUnsupportedAudience) {
		t.Fatalf("wrong audience: %v", err)
	}
	for _, tenant := range []string{"common", "organizations", "consumers", "../tenant", "tenant?x=y", "tenant/other", "tenant..domain"} {
		cfg := appConfig()
		cfg.TenantID = tenant
		if _, err := NewApplicationProvider(cfg, &http.Client{Transport: service}); !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("unsafe tenant accepted: %v", err)
		}
	}
	cfg := appConfig()
	cfg.Cloud = "other-cloud"
	if _, err := NewApplicationProvider(cfg, &http.Client{Transport: service}); !errors.Is(err, ErrUnsupportedCloud) {
		t.Fatalf("unsupported cloud: %v", err)
	}
	if service.requests != 0 {
		t.Fatal("invalid configuration reached token service")
	}
}

func TestApplicationErrorsDoNotRetainServiceBodies(t *testing.T) {
	t.Parallel()
	service := newTokenService(t)
	service.fail = true
	p := application(t, appConfig(), service)
	token, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience))
	if token.AccessToken != "" || !errors.Is(err, ErrTokenService) {
		t.Fatalf("token failure: %v", err)
	}
	for current := err; current != nil; current = errors.Unwrap(current) {
		for _, output := range []string{fmt.Sprintf("%v", current), fmt.Sprintf("%+v", current), fmt.Sprintf("%#v", current)} {
			if strings.Contains(output, "sensitive") || strings.Contains(output, "synthetic-secret") {
				t.Fatal("token service error exposed credentials")
			}
		}
	}
}

func TestApplicationRejectsMetadataTokenDestinations(t *testing.T) {
	t.Parallel()
	for _, endpoint := range []string{"https://other.test/token", "http://login.microsoftonline.com/" + appTenant + "/oauth2/v2.0/token", "https://login.microsoftonline.com/other-tenant/oauth2/v2.0/token"} {
		service := newTokenService(t)
		service.endpoint = endpoint
		p := application(t, appConfig(), service)
		if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); err == nil {
			t.Fatal("unsafe metadata destination accepted")
		}
		if service.count(GraphAudience) != 0 {
			t.Fatal("unsafe destination received credentials")
		}
	}
}

func TestApplicationCloseCancelsTokenExchange(t *testing.T) {
	t.Parallel()
	service := newTokenService(t)
	service.block = make(chan struct{})
	service.entered = make(chan struct{}, 1)
	p := application(t, appConfig(), service)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := p.AcquireToken(ctx, p.Binding().Request(GraphAudience)); result <- err }()
	select {
	case <-service.entered:
	case <-ctx.Done():
		t.Fatal("exchange did not start")
	}
	p.Close()
	p.Wait()
	select {
	case err := <-result:
		if !errors.Is(err, ErrProviderClosed) {
			t.Fatalf("close: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("exchange did not cancel")
	}
	if p.config != nil {
		t.Fatal("close retained credential configuration")
	}
}

func TestApplicationIgnoresImplicitMSALRegion(t *testing.T) {
	t.Setenv("MSAL_FORCE_REGION", "unrequested-region")
	service := newTokenService(t)
	p := application(t, appConfig(), service)
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); err != nil {
		t.Fatalf("explicit commercial authority changed: %v", err)
	}
}

func TestApplicationDomainCanonicalizationAndIssuerIsolation(t *testing.T) {
	t.Parallel()
	service := newTokenService(t)
	service.metadataPath = "/example.onmicrosoft.com/v2.0/.well-known/openid-configuration"
	cfg := appConfig()
	cfg.TenantID = "EXAMPLE.onmicrosoft.com"
	p := application(t, cfg, service)
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); err != nil {
		t.Fatalf("verified domain discovery: %v", err)
	}
	if service.count(GraphAudience) != 1 {
		t.Fatal("domain did not reach canonical token endpoint")
	}
	for _, issuer := range []string{"https://other.test/" + appTenant + "/v2.0", "https://login.microsoftonline.com/33333333-3333-4333-8333-333333333333/v2.0"} {
		service := newTokenService(t)
		service.issuer = issuer
		p := application(t, appConfig(), service)
		if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); err == nil {
			t.Fatal("issuer mismatch accepted")
		}
		if service.count(GraphAudience) != 0 {
			t.Fatal("issuer mismatch received credentials")
		}
	}
}

func TestApplicationRejectsRedirectsAndExpiredCertificatesBeforeExchange(t *testing.T) {
	t.Parallel()
	service := newTokenService(t)
	service.redirect = true
	p := application(t, appConfig(), service)
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); err == nil {
		t.Fatal("redirect followed")
	}
	if service.count(GraphAudience) != 0 {
		t.Fatal("redirect reached a credential exchange")
	}
	data, err := os.ReadFile("testdata/client-credentials-sha1.pfx")
	if err != nil {
		t.Fatal(err)
	}
	certs, key, err := azidentity.ParseCertificates(data, []byte("fixture-only"))
	if err != nil {
		t.Fatal(err)
	}
	template := *certs[0]
	template.SerialNumber = big.NewInt(3)
	template.NotBefore = time.Now().Add(-2 * time.Hour)
	template.NotAfter = time.Now().Add(-time.Hour)
	raw, err := x509.CreateCertificate(rand.Reader, &template, &template, certs[0].PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	rawKey, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &Config{Method: Certificate, TenantID: appTenant, ClientID: appID, Certificate: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw}), PrivateKey: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: rawKey})}
	service = newTokenService(t)
	p = application(t, cfg, service)
	if _, err := p.AcquireToken(t.Context(), p.Binding().Request(GraphAudience)); !errors.Is(err, ErrCertificateExpired) {
		t.Fatalf("expired certificate: %v", err)
	}
	if service.requests != 0 {
		t.Fatal("expired certificate reached service")
	}
}

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

type tokenRoundTrip func(*http.Request) (*http.Response, error)

func (f tokenRoundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAuthorityRejectsAlternativeDestinationsAndClosesBodies(t *testing.T) {
	t.Parallel()
	for _, alter := range []func(*http.Request){
		func(r *http.Request) { r.Host = "other.test" },
		func(r *http.Request) { r.URL.Opaque = "//other.test/token" },
		func(r *http.Request) { r.URL.RawPath = "/%2e%2e/token" },
		func(r *http.Request) { r.URL.User = url.User("identity") },
		func(r *http.Request) { r.URL.RawQuery = "forward=elsewhere" },
		func(r *http.Request) { r.URL.ForceQuery = true },
	} {
		body := &trackedBody{Reader: strings.NewReader("synthetic-secret")}
		request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, tokenURL, body)
		if err != nil {
			t.Fatal(err)
		}
		alter(request)
		called := false
		guard := &authorityHTTP{tenant: appTenant, client: &http.Client{Transport: tokenRoundTrip(func(*http.Request) (*http.Response, error) { called = true; return nil, errors.New("unexpected") })}}
		response, err := guard.Do(request)
		if response != nil {
			if closeErr := response.Body.Close(); closeErr != nil {
				t.Error("could not close unexpected response")
			}
			t.Fatal("invalid request returned a response")
		}
		if !errors.Is(err, ErrAuthorityDestination) {
			t.Fatalf("invalid destination: %v", err)
		}
		if !body.closed || called {
			t.Fatal("invalid request leaked body or reached transport")
		}
	}
}

func TestAuthorityBoundsAndClosesResponses(t *testing.T) {
	t.Parallel()
	body := &trackedBody{Reader: strings.NewReader(strings.Repeat("x", 1024*1024+1))}
	guard := &authorityHTTP{tenant: appTenant, client: &http.Client{Transport: tokenRoundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: body, Request: r}, nil
	})}}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://login.microsoftonline.com/"+appTenant+"/v2.0/.well-known/openid-configuration", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := guard.Do(request)
	if response != nil {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Error("could not close unexpected response")
		}
		t.Fatal("oversized response accepted")
	}
	if !errors.Is(err, ErrTokenService) {
		t.Fatalf("oversized response: %v", err)
	}
	if !body.closed {
		t.Fatal("oversized response body remained open")
	}
}
