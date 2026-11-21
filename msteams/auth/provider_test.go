package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestUnitSuppliedTokenAudienceAndIdentityIsolation(t *testing.T) {
	t.Parallel()
	binding := Binding{TenantID: "tenant", Cloud: "commercial", IdentityID: "identity", Mode: Application}
	now := time.Unix(1000, 0)
	tokens := map[string]Token{
		"graph": {AccessToken: "graph-token", ExpiresAt: now.Add(time.Minute)},
		"teams": {AccessToken: "teams-token", ExpiresAt: now.Add(time.Minute)},
	}
	provider, err := NewSuppliedTokenProvider(binding, tokens)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(provider.Close)
	provider.now = func() time.Time { return now }
	tokens["graph"] = Token{}
	for _, audience := range []string{"graph", "teams"} {
		token, err := provider.AcquireToken(context.Background(), binding.Request(audience))
		if err != nil || token.AccessToken != audience+"-token" {
			t.Fatalf("wrong audience token: %v", err)
		}
	}
	for _, change := range []func(*TokenRequest){
		func(r *TokenRequest) { r.TenantID = "other" }, func(r *TokenRequest) { r.Cloud = "other" },
		func(r *TokenRequest) { r.IdentityID = "other" }, func(r *TokenRequest) { r.Mode = Delegated },
	} {
		request := binding.Request("graph")
		change(&request)
		if _, err := provider.AcquireToken(context.Background(), request); !errors.Is(err, ErrBindingMismatch) {
			t.Fatalf("wrong binding accepted: %v", err)
		}
	}
	if _, err := provider.AcquireToken(context.Background(), binding.Request("missing")); !errors.Is(err, ErrTokenMissing) {
		t.Fatal("missing audience was substituted")
	}
}

func TestUnitSuppliedTokenExpiryCancellationAndClose(t *testing.T) {
	t.Parallel()
	binding := Binding{TenantID: "tenant", Cloud: "commercial", IdentityID: "identity", Mode: Delegated}
	now := time.Unix(1000, 0)
	provider, err := NewSuppliedTokenProvider(binding, map[string]Token{"graph": {AccessToken: "secret", ExpiresAt: now.Add(time.Second)}})
	if err != nil {
		t.Fatal(err)
	}
	provider.now = func() time.Time { return now }
	if _, err := provider.AcquireToken(context.Background(), binding.Request("graph")); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	if _, err := provider.AcquireToken(context.Background(), binding.Request("graph")); !errors.Is(err, ErrTokenExpired) {
		t.Fatal("expired static token was refreshed or returned")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.AcquireToken(ctx, binding.Request("graph")); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation was lost")
	}
	provider.Close()
	provider.Close()
	if _, err := provider.AcquireToken(context.Background(), binding.Request("graph")); !errors.Is(err, ErrProviderClosed) {
		t.Fatal("closed provider returned token")
	}
}

func TestUnitTokenFormattingIsRedacted(t *testing.T) {
	t.Parallel()
	token := Token{AccessToken: "sensitive-access-token", ExpiresAt: time.Now()}
	for _, value := range []string{fmt.Sprintf("%v", token), fmt.Sprintf("%+v", token), fmt.Sprintf("%#v", token)} {
		if strings.Contains(value, token.AccessToken) {
			t.Fatal("formatted token exposed credentials")
		}
	}
}
