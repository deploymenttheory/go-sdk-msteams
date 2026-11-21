// Package auth defines portable authentication configuration and provider boundaries.
// Authentication providers are implemented separately from client construction.
package auth

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"
)

// Method selects exactly one authentication flow. There is no default credential chain.
type Method string

const (
	Browser          Method = "browser"
	DeviceCode       Method = "device-code"
	Password         Method = "password"
	Certificate      Method = "certificate"
	ClientSecret     Method = "client-secret"
	SuppliedTokens   Method = "supplied-tokens"
	ExternalProvider Method = "provider"
	ManagedIdentity  Method = "managed-identity"
	WorkloadIdentity Method = "workload-identity"
)

// Mode distinguishes delegated tokens from application tokens within a session.
type Mode string

const (
	Delegated   Mode = "delegated"
	Application Mode = "application"
)

// Token is opaque access-token material with an explicit expiry.
// Decoding a JWT is not a substitute for server-side token validation.
type Token struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// AssertionSource supplies a workload assertion under the requesting operation's context.
// The caller owns its lifecycle and must make it safe for concurrent calls.
type AssertionSource func(context.Context) (string, error)

// Config selects portable authentication material. Client construction copies its
// mutable byte slices and maps. Referenced callbacks remain caller-owned.
// Provider-specific credential parsing and token exchange belong to the provider.
type Config struct {
	Method              Method
	TenantID            string
	Cloud               string     // named environment; empty selects commercial
	Endpoints           *Endpoints // optional trusted endpoint overrides, copied at construction
	ClientID            string
	ClientSecret        string
	Username            string
	Password            string
	Certificate         []byte
	PrivateKey          []byte
	PFX                 []byte // alternative to Certificate and PrivateKey PEM
	PFXPassword         []byte
	AccessTokens        map[string]Token // TokenRequest.TokenKey -> token; never positional
	TokenMode           Mode             // required for supplied tokens and external providers
	IdentityID          string           // required for token cache isolation
	Provider            Provider         // borrowed; only for ExternalProvider
	AssertionFile       string
	Assertion           AssertionSource
	OpenBrowser         func(context.Context, string) error            // required for Browser; must honor context
	DevicePrompt        func(context.Context, DeviceCodeMessage) error // required for DeviceCode
	LoginHint           string                                         // optional Browser account hint, not an identity assertion
	LoginTimeout        time.Duration                                  // delegated acquisition bound; zero selects 15 minutes
	FederationEndpoints []string                                       // exact trusted HTTPS MEX and password endpoints; Password only
}

// ErrInvalidConfig identifies local authentication configuration failures.
var ErrInvalidConfig = errors.New("invalid authentication configuration")

// ConfigError identifies a field without retaining its potentially sensitive value.
type ConfigError struct {
	Field  string
	Reason string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("authentication %s: %s", e.Field, e.Reason)
}

func (e *ConfigError) Unwrap() error { return ErrInvalidConfig }

// Clone returns an independent copy of mutable credential material.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}
	cloned := *c
	if c.Endpoints != nil {
		endpoints := *c.Endpoints
		cloned.Endpoints = &endpoints
	}
	cloned.Certificate = slices.Clone(c.Certificate)
	cloned.PrivateKey = slices.Clone(c.PrivateKey)
	cloned.PFX = slices.Clone(c.PFX)
	cloned.PFXPassword = slices.Clone(c.PFXPassword)
	cloned.AccessTokens = maps.Clone(c.AccessTokens)
	cloned.FederationEndpoints = slices.Clone(c.FederationEndpoints)
	return &cloned
}

// String prevents ordinary formatted diagnostics from printing credential material.
func (Config) String() string { return "msteams authentication configuration (redacted)" }

// GoString redacts Go-syntax formatted diagnostics as well.
func (c Config) GoString() string { return c.String() }

// Validate checks local selection and credential presence without reading files,
// invoking callbacks, discovering services or authenticating to a tenant.
func (c *Config) Validate() error {
	if c == nil {
		return &ConfigError{Field: "config", Reason: "is required"}
	}
	if strings.TrimSpace(c.TenantID) == "" {
		return &ConfigError{Field: "tenant_id", Reason: "is required"}
	}
	if _, err := ResolveEnvironment(c.Cloud, c.Endpoints); err != nil {
		return err
	}
	allowed := map[string]bool{}
	required := map[string]bool{}
	switch c.Method {
	case Browser:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
		required["open_browser"] = c.OpenBrowser != nil
		allowed["open_browser"], allowed["login_hint"], allowed["login_timeout"] = true, true, true
	case DeviceCode:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
		required["device_prompt"] = c.DevicePrompt != nil
		allowed["device_prompt"], allowed["login_timeout"] = true, true
	case Password:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
		required["username"] = strings.TrimSpace(c.Username) != ""
		required["password"] = c.Password != ""
		allowed["username"], allowed["password"], allowed["federation_endpoints"] = true, true, true
		allowed["login_timeout"] = true
	case Certificate:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
		if len(c.PFX) > 0 {
			allowed["pfx"], allowed["pfx_password"] = true, true
		} else {
			required["certificate"] = len(c.Certificate) > 0
			required["private_key"] = len(c.PrivateKey) > 0
			allowed["certificate"], allowed["private_key"] = true, true
		}
	case ClientSecret:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
		required["client_secret"] = c.ClientSecret != ""
		allowed["client_secret"] = true
	case SuppliedTokens:
		required["access_tokens"] = len(c.AccessTokens) > 0
		required["identity_id"] = strings.TrimSpace(c.IdentityID) != ""
		required["token_mode"] = c.TokenMode == Delegated || c.TokenMode == Application
		allowed["access_tokens"], allowed["identity_id"], allowed["token_mode"] = true, true, true
		for audience, token := range c.AccessTokens {
			invalidToken := strings.TrimSpace(audience) == "" || token.AccessToken == "" || token.ExpiresAt.IsZero()
			if invalidToken {
				return &ConfigError{Field: "access_tokens", Reason: "each token requires an audience, value and expiry"}
			}
		}
	case ExternalProvider:
		required["identity_id"] = strings.TrimSpace(c.IdentityID) != ""
		required["token_mode"] = c.TokenMode == Delegated || c.TokenMode == Application
		required["provider"] = !nilDependency(c.Provider)
		allowed["provider"], allowed["identity_id"], allowed["token_mode"] = true, true, true
	case ManagedIdentity:
		// An empty client ID selects the environment's system-assigned identity.
	case WorkloadIdentity:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
		if (c.AssertionFile == "") == (c.Assertion == nil) {
			return &ConfigError{Field: "assertion", Reason: "select exactly one assertion file or callback"}
		}
		allowed["assertion_file"], allowed["assertion"] = true, true
	default:
		return &ConfigError{Field: "method", Reason: "select a supported explicit authentication method"}
	}
	for _, field := range slices.Sorted(maps.Keys(required)) {
		if !required[field] {
			return &ConfigError{Field: field, Reason: "is required for the selected method"}
		}
	}
	present := map[string]bool{
		"client_secret": c.ClientSecret != "", "username": c.Username != "", "password": c.Password != "",
		"certificate": len(c.Certificate) > 0, "private_key": len(c.PrivateKey) > 0,
		"pfx": len(c.PFX) > 0, "pfx_password": len(c.PFXPassword) > 0,
		"access_tokens": len(c.AccessTokens) > 0, "identity_id": c.IdentityID != "", "token_mode": c.TokenMode != "",
		"assertion_file": c.AssertionFile != "", "assertion": c.Assertion != nil,
		"provider":     c.Provider != nil,
		"open_browser": c.OpenBrowser != nil, "device_prompt": c.DevicePrompt != nil,
		"login_hint": c.LoginHint != "", "login_timeout": c.LoginTimeout != 0,
		"federation_endpoints": len(c.FederationEndpoints) > 0,
	}
	for _, field := range slices.Sorted(maps.Keys(present)) {
		if present[field] && !allowed[field] {
			return &ConfigError{Field: field, Reason: "is incompatible with the selected method"}
		}
	}
	if c.LoginTimeout < 0 {
		return &ConfigError{Field: "login_timeout", Reason: "must be positive when specified"}
	}
	return nil
}

// TokenBinding returns the configured identity for supplied or externally acquired
// tokens. Cloud labels are isolation keys; accepting one does not enable its routes.
func (c *Config) TokenBinding() Binding {
	cloud := c.Cloud
	if environment, err := ResolveEnvironment(c.Cloud, c.Endpoints); err == nil {
		cloud = environment.BindingCloud
	}
	if c.Method == Certificate || c.Method == ClientSecret || c.Method == WorkloadIdentity {
		return Binding{TenantID: c.TenantID, Cloud: cloud, IdentityID: c.ClientID, Mode: Application}
	}
	if c.Method == ManagedIdentity {
		identity := c.ClientID
		if identity == "" {
			identity = "system-assigned"
		}
		return Binding{TenantID: c.TenantID, Cloud: cloud, IdentityID: identity, Mode: Application}
	}
	if c.Method == Browser || c.Method == DeviceCode || c.Method == Password {
		return Binding{TenantID: c.TenantID, Cloud: cloud, IdentityID: c.ClientID, Mode: Delegated}
	}
	return Binding{TenantID: c.TenantID, Cloud: cloud, IdentityID: c.IdentityID, Mode: c.TokenMode}
}

func nilDependency(provider any) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
