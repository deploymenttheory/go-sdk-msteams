package auth

import (
	"context"
	"errors"
)

// ErrRefreshUnsupported means a provider cannot replace a server-rejected token.
// Supplied tokens and borrowed credentials without this capability are not retried.
var ErrRefreshUnsupported = errors.New("provider cannot refresh a rejected token")

// RejectionRefresher is an optional provider capability. It replaces only the
// rejected token for the exact binding, without interactive login or changing
// identity. Concurrent stale rejections must reuse an already-replaced token.
// Implementations must honor context and must not log rejected token material.
type RejectionRefresher interface {
	RefreshAfterRejection(context.Context, TokenRequest, Token) (Token, error)
}

// RefreshAfterRejection invalidates only the matching cached access token. A late
// rejection cannot discard a replacement installed by another caller. Acquisition
// still coalesces through the ordinary per-binding refresh worker.
func (p *RefreshProvider) RefreshAfterRejection(ctx context.Context, request TokenRequest, rejected Token) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	if rejected.AccessToken == "" {
		return Token{}, ErrInvalidToken
	}
	p.mu.Lock()
	if cached, ok := p.tokens[request]; ok && cached.AccessToken == rejected.AccessToken {
		delete(p.tokens, request)
	}
	p.mu.Unlock()
	return p.AcquireToken(ctx, request)
}

// RefreshAfterRejection performs a new application credential exchange when the
// rejected token is still cached. It never changes the selected credential method.
func (p *ApplicationProvider) RefreshAfterRejection(ctx context.Context, request TokenRequest, rejected Token) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	if p.refresh == nil {
		return Token{}, ErrProviderClosed
	}
	if err := p.environment.validateResourceRequest(request); err != nil {
		return Token{}, err
	}
	return p.refresh.RefreshAfterRejection(ctx, request, rejected)
}

func (p *boundProvider) RefreshAfterRejection(ctx context.Context, request TokenRequest, rejected Token) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	source, ok := p.source.(RejectionRefresher)
	if !ok {
		return Token{}, ErrRefreshUnsupported
	}
	token, err := source.RefreshAfterRejection(ctx, request, rejected)
	return validateBorrowedToken(ctx, token, err)
}
