package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// Commercial resource identifiers documented for Teams application sessions. These
// identify token audiences; they are not management API endpoint URLs.
const (
	GraphAudience      = "https://graph.microsoft.com"
	TeamsAdminAudience = "48ac35b8-9aa8-4d74-927d-1f4a14a0b239"
)

var (
	ErrUnsupportedCloud     = errors.New("authentication cloud profile is not supported")
	ErrUnsupportedAudience  = errors.New("token audience is not supported")
	ErrUnsupportedScope     = errors.New("token scope is not supported for the selected resource and mode")
	ErrTokenService         = errors.New("microsoft application token acquisition failed")
	ErrAuthorityDestination = errors.New("authentication destination is not allowed")
	tenantGUID              = regexp.MustCompile(`^[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`)
	tenantLabel             = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)
)

// ApplicationProvider acquires secret, certificate or workload tokens using MSAL Go.
// It owns copied credentials and an audience-isolated refresh cache. Constructors
// perform no network calls, certificate parsing, file reads or store lookups.
type ApplicationProvider struct {
	mu          sync.RWMutex
	config      *Config
	binding     Binding
	environment Environment
	http        http.Client
	refresh     *RefreshProvider
}

// NewApplicationProvider creates an explicitly configured application session. The HTTP
// client value is copied and its transport is borrowed. No fallback login occurs.
// Call Close when finished; Wait can then await pending acquisitions.
func NewApplicationProvider(cfg *Config, httpClient *http.Client) (*ApplicationProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.Method != Certificate && cfg.Method != ClientSecret && cfg.Method != WorkloadIdentity {
		return nil, &ConfigError{Field: "method", Reason: "requires certificate, client-secret or workload-identity"}
	}
	binding := cfg.TokenBinding()
	environment, err := ResolveEnvironment(cfg.Cloud, cfg.Endpoints)
	if err != nil {
		return nil, err
	}
	if !validApplicationTenant(binding.TenantID) {
		return nil, &ConfigError{Field: "tenant_id", Reason: "requires a tenant GUID or DNS domain"}
	}
	if httpClient == nil || (httpClient.Transport != nil && nilDependency(httpClient.Transport)) {
		return nil, &ConfigError{Field: "http_client", Reason: "requires a non-nil client and transport value"}
	}
	p := &ApplicationProvider{config: cfg.Clone(), binding: binding, environment: environment, http: *httpClient}
	// Token-bearing redirects are not delegated to an injected redirect callback.
	p.http.CheckRedirect = func(*http.Request, []*http.Request) error { return ErrAuthorityDestination }
	p.http.Jar = nil
	refresh, err := NewRefreshProvider(binding, p.acquire, RefreshOptions{RefreshSkew: 2 * time.Minute})
	if err != nil {
		return nil, err
	}
	p.refresh = refresh
	return p, nil
}

// Binding returns this provider's immutable tenant/application identity.
func (p *ApplicationProvider) Binding() Binding { return p.binding }
func (*ApplicationProvider) String() string     { return "msteams application provider (redacted)" }
func (p *ApplicationProvider) GoString() string { return p.String() }

// AcquireToken returns only the requested resource token. Cached access tokens
// remain valid until their refresh window even if the signing certificate expires;
// certificate validity is checked before each new credential exchange.
func (p *ApplicationProvider) AcquireToken(ctx context.Context, request TokenRequest) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	if p.refresh == nil {
		return Token{}, ErrProviderClosed
	}
	if err := p.environment.validateResourceRequest(request); err != nil {
		return Token{}, err
	}
	return p.refresh.AcquireToken(ctx, request)
}

func (p *ApplicationProvider) acquire(ctx context.Context, request TokenRequest) (Token, error) {
	p.mu.RLock()
	cfg := p.config.Clone()
	p.mu.RUnlock()
	if cfg == nil {
		return Token{}, ErrProviderClosed
	}
	defer clearApplicationMaterial(cfg)
	if err := ctx.Err(); err != nil {
		return Token{}, err
	}
	credential, err := applicationCredential(ctx, cfg, time.Now())
	if err != nil {
		return Token{}, err
	}
	// Fresh MSAL clients avoid a second token cache returning a near-expiry result
	// after our refresh cache has intentionally requested new credentials. The outer
	// provider still serializes each audience, so concurrent callers share one exchange.
	tenant := strings.ToLower(cfg.TenantID)
	msal, err := confidential.New(p.environment.Endpoints.Authority+"/"+tenant, cfg.ClientID, credential,
		confidential.WithHTTPClient(&authorityHTTP{client: &p.http, tenant: tenant, authority: p.environment.Endpoints.Authority}),
		confidential.WithInstanceDiscovery(false),
		confidential.WithAzureRegion("DisableMsalForceRegion"))
	if err != nil {
		return Token{}, ErrTokenService
	}
	result, err := msal.AcquireTokenByCredential(ctx, []string{request.Audience + "/.default"})
	if err != nil {
		if ctx.Err() != nil {
			return Token{}, ctx.Err()
		}
		// Library errors can include request/response bodies and credential diagnostics.
		// Do not retain them as an unwrap cause across the SDK boundary.
		return Token{}, ErrTokenService
	}
	return Token{AccessToken: result.AccessToken, ExpiresAt: result.ExpiresOn}, nil
}

func applicationCredential(ctx context.Context, cfg *Config, now time.Time) (confidential.Credential, error) {
	if cfg.Method == WorkloadIdentity {
		assertion, err := workloadAssertion(ctx, cfg)
		if err != nil {
			return confidential.Credential{}, err
		}
		return confidential.NewCredFromAssertionCallback(func(ctx context.Context, _ confidential.AssertionRequestOptions) (string, error) {
			return assertion, ctx.Err()
		}), nil
	}
	if cfg.Method == ClientSecret {
		cred, err := confidential.NewCredFromSecret(cfg.ClientSecret)
		if err != nil {
			return confidential.Credential{}, ErrTokenService
		}
		return cred, nil
	}
	var parsed parsedCertificate
	var err error
	if len(cfg.PFX) > 0 {
		parsed, err = parsePFXCertificate(cfg.PFX, cfg.PFXPassword)
	} else {
		parsed, err = parsePEMCertificate(cfg.Certificate, cfg.PrivateKey)
	}
	if err != nil {
		return confidential.Credential{}, err
	}
	if err := parsed.validateTime(now); err != nil {
		return confidential.Credential{}, err
	}
	return parsed.credential, nil
}

func clearApplicationMaterial(cfg *Config) {
	clear(cfg.Certificate)
	clear(cfg.PrivateKey)
	clear(cfg.PFX)
	clear(cfg.PFXPassword)
	cfg.ClientSecret = ""
	cfg.Assertion = nil
	cfg.AssertionFile = ""
}

// Close cancels refreshes and clears owned credential/cache references. It does not
// revoke server tokens or close the borrowed HTTP transport.
func (p *ApplicationProvider) Close() {
	if p.refresh != nil {
		p.refresh.Close()
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.config != nil {
		clearApplicationMaterial(p.config)
		p.config = nil
	}
}

// Wait waits for pending callbacks after Close. The transport must honor context.
func (p *ApplicationProvider) Wait() {
	if p.refresh != nil {
		p.refresh.Wait()
	}
}

// authorityHTTP confines library metadata and token calls to the explicit configured
// authority and tenant. A malicious metadata token endpoint cannot receive a secret.
type authorityHTTP struct {
	client    *http.Client
	tenant    string
	authority string
	mu        sync.RWMutex
	// canonicalTokenPath is learned only from this tenant's same-authority OIDC response.
	canonicalTokenPath string
}

// MSAL borrows the transport; session/caller ownership controls idle cleanup.
func (*authorityHTTP) CloseIdleConnections() {}

func (a *authorityHTTP) Do(request *http.Request) (*http.Response, error) {
	if request == nil || request.URL == nil || a.client == nil {
		return rejectAuthorityRequest(request)
	}
	u := request.URL
	if !a.allowsURL(u) {
		return rejectAuthorityRequest(request)
	}
	if request.Host != "" && request.Host != a.host() {
		return rejectAuthorityRequest(request)
	}
	var path string
	metadataRequest := false
	if request.Method == http.MethodGet && u.Path == "/common/discovery/instance" {
		path = "/common/discovery/instance"
	}
	if request.Method == http.MethodGet && u.Path == "/"+a.tenant+"/v2.0/.well-known/openid-configuration" {
		path = "/" + a.tenant + "/v2.0/.well-known/openid-configuration"
		metadataRequest = true
	}
	if request.Method == http.MethodPost && u.RawQuery == "" && !u.ForceQuery {
		a.mu.RLock()
		tokenPath := a.canonicalTokenPath
		a.mu.RUnlock()
		if tokenPath == "" && tenantGUID.MatchString(a.tenant) {
			tokenPath = "/" + a.tenant + "/oauth2/v2.0/token"
		}
		if tokenPath != "" && u.Path == tokenPath {
			path = tokenPath
		}
	}
	if path == "" {
		return rejectAuthorityRequest(request)
	}
	return a.send(request, path, metadataRequest)
}

func (a *authorityHTTP) send(request *http.Request, path string, metadataRequest bool) (*http.Response, error) {
	// Send a reconstructed destination rather than forwarding any alternative URL
	// representation supplied by discovery. Preserve the library's headers/body.
	trusted := request.Clone(request.Context())
	trusted.URL = &url.URL{Scheme: "https", Host: a.host(), Path: path, RawQuery: request.URL.RawQuery}
	trusted.Host = a.host()
	httpClient := *a.client
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return ErrAuthorityDestination }
	httpClient.Jar = nil
	response, err := httpClient.Do(trusted)
	if err != nil {
		return nil, err
	}
	const responseLimit = 1024 * 1024
	content, readErr := io.ReadAll(io.LimitReader(response.Body, responseLimit+1))
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil || len(content) > responseLimit {
		return nil, ErrTokenService
	}
	if metadataRequest && response.StatusCode >= 200 && response.StatusCode < 300 {
		if err := a.bindCanonicalTenant(content); err != nil {
			return nil, err
		}
	}
	response.Body = io.NopCloser(bytes.NewReader(content))
	response.ContentLength = int64(len(content))
	return response, nil
}

func rejectAuthorityRequest(request *http.Request) (*http.Response, error) {
	if request != nil && request.Body != nil {
		if err := request.Body.Close(); err != nil {
			return nil, ErrAuthorityDestination
		}
	}
	return nil, ErrAuthorityDestination
}

func (a *authorityHTTP) host() string {
	if a.authority == "" {
		return "login.microsoftonline.com"
	}
	return strings.TrimPrefix(a.authority, "https://")
}

func (a *authorityHTTP) allowsURL(u *url.URL) bool {
	return u != nil && u.Scheme == "https" && u.Host == a.host() &&
		u.User == nil && u.Fragment == "" && u.RawPath == "" && u.Opaque == ""
}

func validApplicationTenant(tenant string) bool {
	if tenantGUID.MatchString(tenant) {
		return true
	}
	if len(tenant) > 253 || !strings.Contains(tenant, ".") {
		return false
	}
	for _, label := range strings.Split(tenant, ".") {
		if !tenantLabel.MatchString(label) {
			return false
		}
	}
	return true
}

func (a *authorityHTTP) bindCanonicalTenant(content []byte) error {
	var metadata struct {
		Issuer        string `json:"issuer"`
		TokenEndpoint string `json:"token_endpoint"`
	}
	if err := json.Unmarshal(content, &metadata); err != nil {
		return ErrAuthorityDestination
	}
	issuer, issuerErr := url.Parse(metadata.Issuer)
	endpoint, endpointErr := url.Parse(metadata.TokenEndpoint)
	if issuerErr != nil || endpointErr != nil || !a.allowsURL(issuer) || !a.allowsURL(endpoint) {
		return ErrAuthorityDestination
	}
	if issuer.RawQuery != "" || issuer.ForceQuery || endpoint.RawQuery != "" || endpoint.ForceQuery {
		return ErrAuthorityDestination
	}
	parts := strings.Split(issuer.Path, "/")
	if len(parts) != 3 || parts[0] != "" || !tenantGUID.MatchString(parts[1]) || parts[2] != "v2.0" {
		return ErrAuthorityDestination
	}
	if tenantGUID.MatchString(a.tenant) && !strings.EqualFold(parts[1], a.tenant) {
		return ErrAuthorityDestination
	}
	tokenPath := "/" + parts[1] + "/oauth2/v2.0/token"
	if endpoint.Path != tokenPath {
		return ErrAuthorityDestination
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.canonicalTokenPath != "" && a.canonicalTokenPath != tokenPath {
		return ErrAuthorityDestination
	}
	a.canonicalTokenPath = tokenPath
	return nil
}
