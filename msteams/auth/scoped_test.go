package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

const regionAudience = "https://migrationsvc.teams.microsoft.com"
const regionScope = regionAudience + "/Region.Read.All"

func TestScopedRefreshCacheCoalescingAndRejectionAreIsolated(t *testing.T) {
	t.Parallel()
	binding := Binding{TenantID: "fixture-tenant", Cloud: "commercial", IdentityID: "fixture-identity", Mode: Delegated}
	var mu sync.Mutex
	calls := map[TokenRequest]int{}
	provider, err := NewRefreshProvider(binding, func(ctx context.Context, request TokenRequest) (Token, error) {
		if err := ctx.Err(); err != nil {
			return Token{}, err
		}
		mu.Lock()
		calls[request]++
		count := calls[request]
		mu.Unlock()
		return Token{AccessToken: fmt.Sprintf("opaque-%s-%d", request.Scope, count), ExpiresAt: time.Now().Add(time.Hour)}, nil
	}, RefreshOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { provider.Close(); provider.Wait() })
	first := binding.RequestScope("https://resource.invalid", "https://resource.invalid/Read")
	second := binding.RequestScope(first.Audience, "https://resource.invalid/Write")
	var group sync.WaitGroup
	for _, request := range []TokenRequest{first, second} {
		for range 8 {
			group.Go(func() {
				if _, err := provider.AcquireToken(t.Context(), request); err != nil {
					t.Error(err)
				}
			})
		}
	}
	group.Wait()
	mu.Lock()
	a, b := calls[first], calls[second]
	mu.Unlock()
	if a != 1 || b != 1 {
		t.Fatal("scope-specific refreshes did not coalesce independently")
	}
	rejected, err := provider.AcquireToken(t.Context(), first)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.RefreshAfterRejection(t.Context(), first, rejected); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.AcquireToken(t.Context(), second); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	a, b = calls[first], calls[second]
	mu.Unlock()
	if a != 2 || b != 1 {
		t.Fatal("rejection invalidated another scope's token")
	}
}

func TestSuppliedScopeTokensDoNotReuseAudienceDefault(t *testing.T) {
	t.Parallel()
	binding := Binding{TenantID: "fixture", Cloud: "commercial", IdentityID: "fixture", Mode: Delegated}
	request := binding.RequestScope(regionAudience, regionScope)
	provider, err := NewSuppliedTokenProvider(binding, map[string]Token{
		regionAudience:     {AccessToken: "default-only", ExpiresAt: time.Now().Add(time.Hour)},
		request.TokenKey(): {AccessToken: "explicit-region", ExpiresAt: time.Now().Add(time.Hour)},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(provider.Close)
	token, err := provider.AcquireToken(t.Context(), request)
	if err != nil || token.AccessToken != "explicit-region" {
		t.Fatal("explicit scope reused audience default")
	}
	other := binding.RequestScope(regionAudience, regionAudience+"/Other")
	if _, err := provider.AcquireToken(t.Context(), other); !errors.Is(err, ErrTokenMissing) {
		t.Fatal("unconfigured scope reused supplied token")
	}
}

func TestNativeRegionRequiresExactDelegatedScopeBeforeOAuth(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc"} {
		for _, scope := range []string{"", regionScope, regionAudience + "/.default"} {
			cfg := appConfig()
			cfg.Cloud = cloud
			service := newTokenService(t)
			provider := application(t, cfg, service)
			request := provider.Binding().RequestScope(regionAudience, scope)
			_, err := provider.AcquireToken(t.Context(), request)
			if err == nil || service.requests != 0 {
				t.Fatal("application region request reached OAuth")
			}
		}
	}
	service := &delegatedService{t: t, extraScopes: map[string]bool{regionScope: true}}
	provider := newTestDelegated(t, delegatedConfig(Password), service)
	for _, scope := range []string{"", regionAudience + "/.default", regionAudience + "/Other"} {
		if _, err := provider.AcquireToken(t.Context(), provider.Binding().RequestScope(regionAudience, scope)); !errors.Is(err, ErrUnsupportedScope) || service.requests != 0 {
			t.Fatal("wrong delegated scope reached OAuth")
		}
	}
	for range 2 {
		token, err := provider.AcquireToken(t.Context(), provider.Binding().RequestScope(regionAudience, regionScope))
		if err != nil || token.AccessToken != "fixture-access-"+regionScope {
			t.Fatal("native MSAL region scope changed or failed")
		}
	}
	if len(service.grants) != 1 || !stringsContainScope(service.grants[0].Get("scope"), regionScope) {
		t.Fatal("exact delegated scope was not exchanged/cached once")
	}
}

func stringsContainScope(value, scope string) bool {
	for _, part := range strings.Fields(value) {
		if part == scope {
			return true
		}
	}
	return false
}
