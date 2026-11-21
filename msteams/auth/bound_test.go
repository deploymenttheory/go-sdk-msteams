package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type providerFunc func(context.Context, TokenRequest) (Token, error)

func (f providerFunc) AcquireToken(ctx context.Context, r TokenRequest) (Token, error) {
	return f(ctx, r)
}

func TestUnitBoundProviderRejectsCrossBindingBeforeCallingSource(t *testing.T) {
	t.Parallel()
	binding := testBinding()
	called := false
	p, err := NewBoundProvider(binding, providerFunc(func(context.Context, TokenRequest) (Token, error) { called = true; return Token{}, nil }))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"tenant", "cloud", "identity", "mode"} {
		r := binding.Request("graph")
		switch field {
		case "tenant":
			r.TenantID = "other"
		case "cloud":
			r.Cloud = "other"
		case "identity":
			r.IdentityID = "other"
		case "mode":
			r.Mode = Delegated
		}
		if _, err := p.AcquireToken(t.Context(), r); !errors.Is(err, ErrBindingMismatch) {
			t.Fatalf("%s: %v", field, err)
		}
	}
	if called {
		t.Fatal("wrong binding reached source")
	}
}

func TestUnitBoundProviderValidatesResultsAndCancellation(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		token Token
		want  error
	}{
		{"empty", Token{}, ErrInvalidToken},
		{"no expiry", Token{AccessToken: "token"}, ErrInvalidToken},
		{"expired", Token{AccessToken: "token", ExpiresAt: time.Now().Add(-time.Hour)}, ErrTokenExpired},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, err := NewBoundProvider(testBinding(), providerFunc(func(context.Context, TokenRequest) (Token, error) { return test.token, nil }))
			if err != nil {
				t.Fatal(err)
			}
			if token, err := p.AcquireToken(t.Context(), testBinding().Request("graph")); !errors.Is(err, test.want) || token.AccessToken != "" {
				t.Fatalf("result: %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	p, err := NewBoundProvider(testBinding(), providerFunc(func(context.Context, TokenRequest) (Token, error) {
		cancel()
		return Token{AccessToken: "token", ExpiresAt: time.Now().Add(time.Hour)}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if token, err := p.AcquireToken(ctx, testBinding().Request("graph")); !errors.Is(err, context.Canceled) || token.AccessToken != "" {
		t.Fatalf("cancellation: %v", err)
	}
}
