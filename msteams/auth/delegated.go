package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

// DeviceCodeMessage contains instructions intended only for the signing-in user.
// The private device code is never passed to application callbacks.
type DeviceCodeMessage struct {
	UserCode        string
	VerificationURL string
	Message         string
	ExpiresAt       time.Time
}

func (DeviceCodeMessage) String() string     { return "msteams device instructions (redacted)" }
func (m DeviceCodeMessage) GoString() string { return m.String() }

// DelegatedProvider owns one MSAL account/cache and the selected explicit login
// method. It never switches between browser, device and password modes on failure.
type DelegatedProvider struct {
	mu          sync.RWMutex
	config      *Config
	binding     Binding
	environment Environment
	msal        *public.Client
	authority   *delegatedHTTP
	account     public.Account
	gate        chan struct{}
	refresh     *RefreshProvider
}

// NewDelegatedProvider validates and wires an explicitly configured session without
// network calls, browser launch or callback execution. Callbacks must honor their
// context. LoginTimeout bounds shared acquisition (default 15 minutes); individual
// callers retain their own cancellation while waiting for that acquisition.
func NewDelegatedProvider(cfg *Config, httpClient *http.Client) (*DelegatedProvider, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.Method != Browser && cfg.Method != DeviceCode && cfg.Method != Password {
		return nil, &ConfigError{Field: "method", Reason: "requires browser, device-code or password"}
	}
	binding := cfg.TokenBinding()
	environment, err := ResolveEnvironment(cfg.Cloud, cfg.Endpoints)
	if err != nil {
		return nil, err
	}
	if !validApplicationTenant(cfg.TenantID) {
		return nil, &ConfigError{Field: "tenant_id", Reason: "requires a tenant GUID or DNS domain"}
	}
	if httpClient == nil || httpClient.Transport != nil && nilDependency(httpClient.Transport) {
		return nil, &ConfigError{Field: "http_client", Reason: "requires a non-nil client and transport value"}
	}
	clonedHTTP := *httpClient
	authority := newDelegatedHTTP(&clonedHTTP, strings.ToLower(cfg.TenantID), cfg.Username)
	authority.base.authority = environment.Endpoints.Authority
	if err := authority.trustFederation(cfg.FederationEndpoints); err != nil {
		return nil, err
	}
	msal, err := public.New(cfg.ClientID, public.WithAuthority(environment.Endpoints.Authority+"/"+strings.ToLower(cfg.TenantID)), public.WithHTTPClient(authority), public.WithInstanceDiscovery(false))
	if err != nil {
		return nil, ErrDelegatedAuthentication
	}
	p := &DelegatedProvider{config: cfg.Clone(), binding: binding, environment: environment, msal: &msal, authority: authority, gate: make(chan struct{}, 1)}
	// MSAL owns refresh tokens and its silent-acquisition cache. This outer cache
	// coalesces exact-audience requests without repeatedly returning a near-expiry
	// MSAL access token into a conflicting proactive refresh window.
	p.refresh, err = NewRefreshProvider(binding, p.acquire, RefreshOptions{})
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *DelegatedProvider) Binding() Binding { return p.binding }
func (*DelegatedProvider) String() string     { return "msteams delegated provider (redacted)" }
func (p *DelegatedProvider) GoString() string { return p.String() }

func (p *DelegatedProvider) AcquireToken(ctx context.Context, request TokenRequest) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	if p.refresh == nil {
		return Token{}, ErrProviderClosed
	}
	if !p.environment.supportsAudience(request.Audience) {
		return Token{}, ErrUnsupportedAudience
	}
	token, err := p.refresh.AcquireToken(ctx, request)
	var failure *DelegatedError
	if errors.As(err, &failure) {
		return Token{}, failure
	}
	return token, err
}

func (p *DelegatedProvider) acquire(ctx context.Context, request TokenRequest) (Token, error) {
	p.mu.RLock()
	cfg := p.config.Clone()
	p.mu.RUnlock()
	if cfg == nil {
		return Token{}, ErrProviderClosed
	}
	defer func() { cfg.Password = ""; cfg.OpenBrowser = nil; cfg.DevicePrompt = nil }()
	timeout := cfg.LoginTimeout
	if timeout == 0 {
		timeout = 15 * time.Minute
	}
	operation, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	select {
	case p.gate <- struct{}{}:
	case <-operation.Done():
		return Token{}, operation.Err()
	}
	defer func() { <-p.gate }()
	p.mu.RLock()
	msal, account := p.msal, p.account
	p.mu.RUnlock()
	if msal == nil {
		return Token{}, ErrProviderClosed
	}
	if err := operation.Err(); err != nil {
		return Token{}, err
	}
	p.authority.resetFailure()
	scopes := []string{request.Audience + "/.default"}
	var result public.AuthResult
	var err error
	if account.HomeAccountID != "" {
		result, err = msal.AcquireTokenSilent(operation, scopes, public.WithSilentAccount(account))
		if err == nil {
			// A refresh response may omit its ID token. MSAL then returns no new
			// account record; keep the account explicitly used for this refresh.
			if result.Account.HomeAccountID == "" {
				result.Account = account
			}
			return p.accept(operation, result)
		}
		if operation.Err() != nil {
			return Token{}, operation.Err()
		}
		if failure := p.authority.failure(); failure != nil && failure.OAuthCode != "invalid_grant" && failure.OAuthCode != "interaction_required" && failure.OAuthCode != "consent_required" {
			return Token{}, failure
		}
	}
	p.authority.resetFailure()
	switch cfg.Method {
	case Browser:
		result, err = browserToken(operation, *msal, p.authority, cfg, scopes)
	case DeviceCode:
		result, err = p.deviceToken(operation, *msal, cfg, scopes)
	case Password:
		//nolint:staticcheck // SA1019: explicit legacy password mode preserves the PowerShell baseline; never selected as a fallback.
		result, err = msal.AcquireTokenByUsernamePassword(operation, scopes, cfg.Username, cfg.Password)
	}
	if operation.Err() != nil {
		return Token{}, operation.Err()
	}
	if err != nil {
		if failure := p.authority.failure(); failure != nil {
			return Token{}, failure
		}
		for _, known := range []error{context.Canceled, context.DeadlineExceeded, ErrAuthorityDestination, ErrDeviceCodeExpired, ErrLoginCallback, ErrAuthenticationDenied, ErrDelegatedAuthentication, ErrFederatedPassword} {
			if errors.Is(err, known) {
				return Token{}, known
			}
		}
		return Token{}, ErrDelegatedAuthentication
	}
	return p.accept(operation, result)
}

func (p *DelegatedProvider) accept(ctx context.Context, result public.AuthResult) (Token, error) {
	if err := ctx.Err(); err != nil {
		return Token{}, err
	}
	if result.AccessToken == "" || result.ExpiresOn.IsZero() || result.Account.HomeAccountID == "" {
		return Token{}, ErrInvalidToken
	}
	if !result.ExpiresOn.After(time.Now()) {
		return Token{}, ErrTokenExpired
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.config == nil {
		return Token{}, ErrProviderClosed
	}
	if p.account.HomeAccountID != "" && p.account.HomeAccountID != result.Account.HomeAccountID {
		return Token{}, ErrBindingMismatch
	}
	p.account = result.Account
	return Token{AccessToken: result.AccessToken, ExpiresAt: result.ExpiresOn}, nil
}

func (p *DelegatedProvider) deviceToken(ctx context.Context, msal public.Client, cfg *Config, scopes []string) (public.AuthResult, error) {
	code, err := msal.AcquireTokenByDeviceCode(ctx, scopes)
	if err != nil {
		return public.AuthResult{}, err
	}
	if !p.authority.validDeviceVerificationURL(code.Result.VerificationURL) {
		return public.AuthResult{}, ErrAuthorityDestination
	}
	if code.Result.UserCode == "" || !code.Result.ExpiresOn.After(time.Now()) {
		return public.AuthResult{}, ErrDeviceCodeExpired
	}
	operation, cancel := context.WithDeadline(ctx, code.Result.ExpiresOn)
	defer cancel()
	if err := cfg.DevicePrompt(operation, DeviceCodeMessage{UserCode: code.Result.UserCode, VerificationURL: code.Result.VerificationURL, Message: code.Result.Message, ExpiresAt: code.Result.ExpiresOn}); err != nil {
		if ctx.Err() != nil {
			return public.AuthResult{}, ctx.Err()
		}
		if !code.Result.ExpiresOn.After(time.Now()) {
			return public.AuthResult{}, ErrDeviceCodeExpired
		}
		return public.AuthResult{}, loginCallbackError(operation, err)
	}
	p.authority.startPolling(code.Result.Interval)
	defer p.authority.stopPolling()
	result, err := code.AuthenticationResult(operation)
	if ctx.Err() == nil && !code.Result.ExpiresOn.After(time.Now()) {
		return public.AuthResult{}, ErrDeviceCodeExpired
	}
	return result, err
}

// Close cancels owned work and releases account/cache references without revoking
// server tokens or closing the borrowed HTTP transport. Wait awaits callback exit.
func (p *DelegatedProvider) Close() {
	if p.refresh != nil {
		p.refresh.Close()
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.config != nil {
		p.config.Password = ""
		p.config.OpenBrowser = nil
		p.config.DevicePrompt = nil
	}
	p.config = nil
	p.msal = nil
	p.account = public.Account{}
}

func (p *DelegatedProvider) Wait() {
	if p.refresh != nil {
		p.refresh.Wait()
	}
}
