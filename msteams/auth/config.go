// Package auth defines portable authentication configuration and provider boundaries.
// Authentication providers are implemented separately from client construction.
package auth

import (
	"context"
	"errors"
	"fmt"
	"maps"
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
	Method        Method
	TenantID      string
	ClientID      string
	ClientSecret  string
	Username      string
	Password      string
	Certificate   []byte
	PrivateKey    []byte
	AccessTokens  map[string]Token // audience -> token; never positional
	TokenMode     Mode             // required only for SuppliedTokens
	IdentityID    string           // required for supplied-token cache isolation
	AssertionFile string
	Assertion     AssertionSource
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
	cloned.Certificate = slices.Clone(c.Certificate)
	cloned.PrivateKey = slices.Clone(c.PrivateKey)
	cloned.AccessTokens = maps.Clone(c.AccessTokens)
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
	allowed := map[string]bool{}
	required := map[string]bool{}
	switch c.Method {
	case Browser, DeviceCode:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
	case Password:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
		required["username"] = strings.TrimSpace(c.Username) != ""
		required["password"] = c.Password != ""
		allowed["username"], allowed["password"] = true, true
	case Certificate:
		required["client_id"] = strings.TrimSpace(c.ClientID) != ""
		required["certificate"] = len(c.Certificate) > 0
		required["private_key"] = len(c.PrivateKey) > 0
		allowed["certificate"], allowed["private_key"] = true, true
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
		"access_tokens": len(c.AccessTokens) > 0, "identity_id": c.IdentityID != "", "token_mode": c.TokenMode != "",
		"assertion_file": c.AssertionFile != "", "assertion": c.Assertion != nil,
	}
	for _, field := range slices.Sorted(maps.Keys(present)) {
		if present[field] && !allowed[field] {
			return &ConfigError{Field: field, Reason: "is incompatible with the selected method"}
		}
	}
	return nil
}
