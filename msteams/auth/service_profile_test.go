package auth

import (
	"errors"
	"testing"
)

func TestUnitClosedServiceAudiencesRespectCloud(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		cloud, audience string
		allowed         bool
	}{
		{"commercial", "https://retailservices.teams.microsoft.com", true},
		{"commercial", "https://chatsvcagg.teams.microsoft.com", true},
		{"gcc", "https://retailservices.teams.microsoft.com", true},
		{"gcc", "https://chatsvcagg.teams.microsoft.com", false},
		{"gcch", "https://retailservices.teams.microsoft.com", false},
		{"dod", "https://chatsvcagg.teams.microsoft.com", false},
		{"china", "https://retailservices.teams.microsoft.com", false},
		{"commercial", "https://retailservices.gcc.teams.microsoft.com", false},
		{"commercial", "https://hierarchy.teams.microsoft.com", false},
		{"commercial", "https://retailservices.teams.microsoft.com/Region.Read.All", false},
		{"commercial", "https://foreign.invalid", false},
	} {
		t.Run(tc.cloud+"-"+tc.audience, func(t *testing.T) {
			env, err := ResolveEnvironment(tc.cloud, nil)
			if err != nil {
				t.Fatal(err)
			}
			if env.supportsAudience(tc.audience) != tc.allowed {
				t.Fatal("closed audience/cloud selection differs")
			}
			if tc.allowed {
				return
			}
			cfg := appConfig()
			cfg.Cloud = tc.cloud
			service := newTokenService(t)
			p := application(t, cfg, service)
			_, err = p.AcquireToken(t.Context(), p.Binding().Request(tc.audience))
			if !errors.Is(err, ErrUnsupportedAudience) || service.requests != 0 {
				t.Fatal("unsupported audience acquired credentials or reached OAuth")
			}
		})
	}
}

func TestUnitDelegatedHierarchyUsesSeparateResourceTokensAndOneAccount(t *testing.T) {
	t.Parallel()
	targeting, hierarchy := "https://retailservices.teams.microsoft.com", "https://chatsvcagg.teams.microsoft.com"
	service := &delegatedService{t: t, extraScopes: map[string]bool{targeting + "/.default": true, hierarchy + "/.default": true}}
	p := newTestDelegated(t, delegatedConfig(Password), service)
	for _, audience := range []string{targeting, hierarchy, GraphAudience, targeting} {
		token, err := p.AcquireToken(t.Context(), p.Binding().Request(audience))
		if err != nil || token.AccessToken != "fixture-access-"+audience+"/.default" {
			t.Fatal("exact resource token acquisition failed")
		}
	}
	if len(service.grants) != 3 || service.grants[0].Get("grant_type") != "password" || service.grants[1].Get("grant_type") != "refresh_token" || service.grants[2].Get("grant_type") != "refresh_token" {
		t.Fatal("resource changes prompted/replayed login or failed to retain selected account")
	}
}
