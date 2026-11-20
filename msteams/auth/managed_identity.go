package auth

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/managedidentity"
)

var (
	ErrManagedIdentityUnavailable = errors.New("managed identity is unavailable in this environment")
	ErrManagedIdentityRejected    = errors.New("managed identity endpoint rejected the selected identity or resource")
	ErrIdentityEnvironmentChanged = errors.New("managed identity environment changed; create a new session")
)

// ManagedIdentityProvider uses Azure Identity for explicit managed identity only.
// An empty Config.ClientID selects the hosting environment's default identity.
// TenantID binds SDK requests; the hosting environment determines the actual home
// tenant. Configure it correctly: managed identity cannot switch to another tenant.
// Construction snapshots environment configuration without network or file reads.
type ManagedIdentityProvider struct {
	mu          sync.RWMutex
	config      *Config
	binding     Binding
	http        http.Client
	environment map[string]string
	source      managedidentity.Source
	refresh     *RefreshProvider
}

var managedEnvironmentKeys = []string{"IDENTITY_ENDPOINT", "IDENTITY_HEADER", "IDENTITY_SERVER_THUMBPRINT", "MSI_ENDPOINT", "MSI_SECRET", "IMDS_ENDPOINT", "DEFAULT_IDENTITY_CLIENT_ID", "ProgramData"}

// NewManagedIdentityProvider copies configuration and the HTTP client. Platform
// discovery and Azure credential construction are deferred to acquisition because
// the underlying library may inspect Azure Arc installation files. No fallback to
// a developer login, client secret or workload credential is attempted.
func NewManagedIdentityProvider(cfg *Config, httpClient *http.Client) (*ManagedIdentityProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.Method != ManagedIdentity {
		return nil, &ConfigError{Field: "method", Reason: "requires managed-identity"}
	}
	binding := cfg.TokenBinding()
	if binding.Cloud != "commercial" {
		return nil, ErrUnsupportedCloud
	}
	if !validApplicationTenant(cfg.TenantID) {
		return nil, &ConfigError{Field: "tenant_id", Reason: "requires a tenant GUID or DNS domain"}
	}
	if cfg.ClientID != "" && !tenantGUID.MatchString(cfg.ClientID) {
		return nil, &ConfigError{Field: "client_id", Reason: "managed identity client ID must be a GUID"}
	}
	if httpClient == nil || httpClient.Transport != nil && nilDependency(httpClient.Transport) {
		return nil, &ConfigError{Field: "http_client", Reason: "requires a non-nil client and transport value"}
	}
	env := make(map[string]string, len(managedEnvironmentKeys))
	for _, key := range managedEnvironmentKeys {
		env[key] = os.Getenv(key)
	}
	p := &ManagedIdentityProvider{config: cfg.Clone(), binding: binding, http: *httpClient, environment: env}
	p.http.Jar = nil
	p.http.CheckRedirect = func(*http.Request, []*http.Request) error { return ErrAuthorityDestination }
	var err error
	p.refresh, err = NewRefreshProvider(binding, p.acquire, RefreshOptions{RefreshSkew: 2 * time.Minute})
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (p *ManagedIdentityProvider) Binding() Binding { return p.binding }
func (*ManagedIdentityProvider) String() string {
	return "msteams managed identity provider (redacted)"
}
func (p *ManagedIdentityProvider) GoString() string { return p.String() }
func (p *ManagedIdentityProvider) AcquireToken(ctx context.Context, request TokenRequest) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	if p.refresh == nil {
		return Token{}, ErrProviderClosed
	}
	if request.Audience != GraphAudience && request.Audience != TeamsAdminAudience {
		return Token{}, ErrUnsupportedAudience
	}
	// Check even cache hits: a changed host identity must not receive a prior token.
	p.mu.RLock()
	closed := p.config == nil
	valid := sameManagedEnvironment(p.environment)
	p.mu.RUnlock()
	if closed {
		return Token{}, ErrProviderClosed
	}
	if !valid {
		return Token{}, ErrIdentityEnvironmentChanged
	}
	return p.refresh.AcquireToken(ctx, request)
}
func (p *ManagedIdentityProvider) acquire(ctx context.Context, request TokenRequest) (Token, error) {
	p.mu.RLock()
	cfg := p.config.Clone()
	env := maps.Clone(p.environment)
	p.mu.RUnlock()
	if cfg == nil {
		return Token{}, ErrProviderClosed
	}
	if !sameManagedEnvironment(env) {
		return Token{}, ErrIdentityEnvironmentChanged
	}
	// Bound a missing IMDS endpoint even for standalone provider callers. A shorter
	// caller deadline or HTTP timeout still wins; Azure's internal retries are off.
	operation, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	source, err := managedidentity.GetSource()
	if err != nil {
		return Token{}, ErrManagedIdentityUnavailable
	}
	p.mu.Lock()
	if p.source == "" {
		p.source = source
	}
	changed := p.source != source
	p.mu.Unlock()
	if changed {
		return Token{}, ErrIdentityEnvironmentChanged
	}
	guard, err := newManagedGuard(source, env, cfg.ClientID, request.Audience)
	if err != nil {
		return Token{}, err
	}
	httpClient := p.http
	options := &azidentity.ManagedIdentityCredentialOptions{ClientOptions: azcore.ClientOptions{
		Transport: &httpClient, Retry: policy.RetryOptions{MaxRetries: -1, TryTimeout: 30 * time.Second}, PerRetryPolicies: []policy.Policy{guard},
	}}
	if cfg.ClientID != "" {
		options.ID = azidentity.ClientID(cfg.ClientID)
	}
	credential, err := azidentity.NewManagedIdentityCredential(options)
	if err != nil {
		return Token{}, ErrManagedIdentityUnavailable
	}
	// Pinned MSAL v1.10.1 uses a process-global managed-identity cache which omits
	// the hosting endpoint and tenant. Empty JSON claims bypass its cache lookup
	// without requesting additional claims. Azure Identity forwards this option;
	// the pinned MSAL MI implementation doesn't send it to the endpoint. Our outer
	// cache provides binding isolation/coalescing. Regression tests protect this
	// version-sensitive boundary; do not remove it during dependency upgrades.
	result, err := credential.GetToken(operation, policy.TokenRequestOptions{TenantID: request.TenantID, Scopes: []string{request.Audience + "/.default"}, Claims: "{}"})
	if ctx.Err() != nil {
		return Token{}, ctx.Err()
	}
	if !sameManagedEnvironment(env) {
		return Token{}, ErrIdentityEnvironmentChanged
	}
	if guard.failure != nil {
		return Token{}, guard.failure
	}
	if err != nil {
		if guard.status == 400 || guard.status == 401 || guard.status == 403 {
			return Token{}, ErrManagedIdentityRejected
		}
		return Token{}, ErrManagedIdentityUnavailable
	}
	return Token{AccessToken: result.Token, ExpiresAt: result.ExpiresOn}, nil
}
func sameManagedEnvironment(env map[string]string) bool {
	for key, value := range env {
		if os.Getenv(key) != value {
			return false
		}
	}
	return true
}

// Close cancels owned work and drops configuration/cache references. Azure's
// internal process cache is library-owned and is never read by this provider.
// The borrowed HTTP transport is not closed. Wait awaits outstanding acquisitions.
func (p *ManagedIdentityProvider) Close() {
	if p.refresh != nil {
		p.refresh.Close()
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = nil
	clear(p.environment)
}
func (p *ManagedIdentityProvider) Wait() {
	if p.refresh != nil {
		p.refresh.Wait()
	}
}
