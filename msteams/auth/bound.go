package auth

import (
	"context"
	"time"
)

// boundProvider enforces a session's identity at the borrowed-provider boundary.
// It owns neither the source nor its cache and has no close operation.
type boundProvider struct {
	binding Binding
	source  Provider
}

// NewBoundProvider borrows a provider under one explicit session binding. It
// validates requests and results and redacts source errors, preserving errors.Is/As.
// Token audience and issuer remain the source's assertion until server validation;
// this wrapper does not decode opaque tokens or discover tenant permissions.
func NewBoundProvider(binding Binding, source Provider) (Provider, error) {
	if err := binding.Validate(); err != nil {
		return nil, err
	}
	if nilDependency(source) {
		return nil, ErrInvalidProvider
	}
	return &boundProvider{binding: binding, source: source}, nil
}

func (p *boundProvider) AcquireToken(ctx context.Context, request TokenRequest) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	token, err := p.source.AcquireToken(ctx, request)
	if ctx.Err() != nil {
		return Token{}, ctx.Err()
	}
	if err != nil {
		return Token{}, AcquisitionError{cause: err}
	}
	if token.AccessToken == "" || token.ExpiresAt.IsZero() {
		return Token{}, ErrInvalidToken
	}
	if !token.ExpiresAt.After(time.Now()) {
		return Token{}, ErrTokenExpired
	}
	return token, nil
}

func (*boundProvider) String() string { return "msteams bound provider (redacted)" }

func (p *boundProvider) GoString() string { return p.String() }
