package auth

import (
	"context"
	"errors"
	"maps"
	"strings"
	"sync"
	"time"
	"unicode"
)

// Binding identifies one authentication session. Cloud is an explicit profile
// identity, not a destination URL; backend routing resolves its endpoints separately.
type Binding struct {
	TenantID   string
	Cloud      string
	IdentityID string
	Mode       Mode
}

// TokenRequest includes every isolation dimension and the exact resource audience.
// Audiences are opaque identifiers; the provider never substitutes another audience.
type TokenRequest struct {
	TenantID   string
	Cloud      string
	IdentityID string
	Mode       Mode
	Audience   string
	// Scope is one explicit delegated resource scope. Empty preserves the existing
	// audience + /.default contract. The comparable value participates in cache,
	// coalescing and rejected-token isolation; no scope is silently substituted.
	Scope string
}

// Provider acquires a resource token under the operation's cancellation context.
// Implementations must respect cancellation and must not log token material.
type Provider interface {
	AcquireToken(context.Context, TokenRequest) (Token, error)
}

var (
	ErrInvalidBinding  = errors.New("token binding requires tenant, cloud, identity and a single mode")
	ErrInvalidRequest  = errors.New("token request requires a context and audience")
	ErrBindingMismatch = errors.New("token request does not match provider binding")
	ErrTokenMissing    = errors.New("no supplied token for the requested audience")
	ErrTokenExpired    = errors.New("access token has expired")
	ErrProviderClosed  = errors.New("authentication provider is closed")
	ErrInvalidToken    = errors.New("token requires an opaque value and explicit expiration")
	ErrInvalidProvider = errors.New("token provider is required")
)

// Validate checks identity presence and rejects mixed/unspecified authentication mode.
func (b Binding) Validate() error {
	missingIdentity := strings.TrimSpace(b.TenantID) == "" || strings.TrimSpace(b.Cloud) == "" || strings.TrimSpace(b.IdentityID) == ""
	invalidMode := b.Mode != Delegated && b.Mode != Application
	if missingIdentity || invalidMode {
		return ErrInvalidBinding
	}
	return nil
}

// Request creates a fully bound request for one exact resource audience.
func (b Binding) Request(audience string) TokenRequest {
	return TokenRequest{TenantID: b.TenantID, Cloud: b.Cloud, IdentityID: b.IdentityID, Mode: b.Mode, Audience: audience}
}

// RequestScope creates a bound request with an exact delegated scope.
func (b Binding) RequestScope(audience, scope string) TokenRequest {
	request := b.Request(audience)
	request.Scope = scope
	return request
}

// TokenKey addresses explicitly supplied tokens. Existing audience-only keys
// remain unchanged; scoped tokens use audience + a space + exact scope. Callers
// are responsible for the opaque token's actual consent; no JWT is decoded.
func (r TokenRequest) TokenKey() string {
	if r.Scope == "" {
		return r.Audience
	}
	return r.Audience + " " + r.Scope
}

func (r TokenRequest) resourceScope() string {
	if r.Scope != "" {
		return r.Scope
	}
	return r.Audience + "/.default"
}

func (b Binding) matches(request TokenRequest) bool {
	return b == (Binding{TenantID: request.TenantID, Cloud: request.Cloud, IdentityID: request.IdentityID, Mode: request.Mode})
}

func validateRequest(ctx context.Context, binding Binding, request TokenRequest) error {
	if ctx == nil || strings.TrimSpace(request.Audience) == "" {
		return ErrInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !binding.matches(request) {
		return ErrBindingMismatch
	}
	if request.Scope != "" && (request.Mode != Delegated || strings.ContainsFunc(request.Scope, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) || !strings.HasPrefix(request.Scope, request.Audience+"/")) {
		return ErrInvalidRequest
	}
	return nil
}

// SuppliedTokenProvider owns copied TokenKey-addressed tokens. It never refreshes them
// or tries to determine issuer validity by decoding a JWT; the server is authoritative.
type SuppliedTokenProvider struct {
	mu      sync.RWMutex
	binding Binding
	tokens  map[string]Token
	now     func() time.Time
	closed  bool
}

// NewSuppliedTokenProvider binds a copied token inventory. Expiration is checked
// on acquisition so an expired input can be diagnosed without contacting a service.
func NewSuppliedTokenProvider(binding Binding, tokens map[string]Token) (*SuppliedTokenProvider, error) {
	if err := binding.Validate(); err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, ErrTokenMissing
	}
	for audience, token := range tokens {
		invalidToken := token.AccessToken == "" || token.ExpiresAt.IsZero()
		if strings.TrimSpace(audience) == "" || invalidToken {
			return nil, ErrInvalidToken
		}
	}
	return &SuppliedTokenProvider{binding: binding, tokens: maps.Clone(tokens), now: time.Now}, nil
}

// AcquireToken returns only the token for the exact bound identity and audience.
func (p *SuppliedTokenProvider) AcquireToken(ctx context.Context, request TokenRequest) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed || p.tokens == nil {
		return Token{}, ErrProviderClosed
	}
	token, ok := p.tokens[request.TokenKey()]
	if !ok {
		return Token{}, ErrTokenMissing
	}
	if !token.ExpiresAt.After(p.now()) {
		return Token{}, ErrTokenExpired
	}
	return token, nil
}

// Close clears owned references. It does not revoke a token or erase caller copies.
func (p *SuppliedTokenProvider) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	clear(p.tokens)
	p.closed = true
}

// String redacts owned token material from formatted diagnostics.
func (*SuppliedTokenProvider) String() string { return "msteams supplied token provider (redacted)" }

// GoString redacts Go-syntax formatted diagnostics.
func (p *SuppliedTokenProvider) GoString() string { return p.String() }

// String prevents ordinary formatted diagnostics from printing token material.
func (Token) String() string { return "msteams token (redacted)" }

// GoString also redacts Go-syntax formatted diagnostics.
func (t Token) GoString() string { return t.String() }
