package auth_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestExplicitMethodsValidateWithoutAuthentication(t *testing.T) {
	t.Parallel()
	cases := []*auth.Config{
		{Method: auth.Browser, ClientID: "client", OpenBrowser: func(context.Context, string) error { return nil }},
		{Method: auth.DeviceCode, ClientID: "client", DevicePrompt: func(context.Context, auth.DeviceCodeMessage) error { return nil }},
		{Method: auth.Password, ClientID: "client", Username: "user", Password: "secret"},
		{Method: auth.Certificate, ClientID: "client", Certificate: []byte("certificate"), PrivateKey: []byte("key")},
		{Method: auth.ClientSecret, ClientID: "client", ClientSecret: "secret"},
		{Method: auth.ManagedIdentity},
		{Method: auth.WorkloadIdentity, ClientID: "client", AssertionFile: "/does/not/exist"},
		{Method: auth.SuppliedTokens, IdentityID: "identity", TokenMode: auth.Application,
			AccessTokens: map[string]auth.Token{"audience": {AccessToken: "token", ExpiresAt: time.Unix(1, 0)}}},
	}
	for _, cfg := range cases {
		t.Run(string(cfg.Method), func(t *testing.T) {
			t.Parallel()
			cfg.TenantID = "tenant"
			if err := cfg.Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
	// Presence validation deliberately leaves file access, certificate parsing and
	// expired-token handling to the operation-scoped authentication provider.
}

func TestRejectMissingAndIncompatibleCredentials(t *testing.T) {
	t.Parallel()
	cases := []*auth.Config{
		nil,
		{Method: auth.ManagedIdentity},
		{TenantID: "tenant"},
		{TenantID: "tenant", Method: auth.ClientSecret, ClientID: "client"},
		{TenantID: "tenant", Method: auth.DeviceCode, ClientID: "client", Password: "secret"},
		{TenantID: "tenant", Method: auth.WorkloadIdentity, ClientID: "client"},
		{TenantID: "tenant", Method: auth.SuppliedTokens, TokenMode: auth.Mode("mixed")},
		{TenantID: "tenant", Method: auth.SuppliedTokens, IdentityID: "id", TokenMode: auth.Application,
			AccessTokens: map[string]auth.Token{"audience": {AccessToken: "secret"}}},
	}
	for i, cfg := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			t.Parallel()
			err := cfg.Validate()
			if !errors.Is(err, auth.ErrInvalidConfig) {
				t.Fatalf("expected configuration error, got %v", err)
			}
			if strings.Contains(err.Error(), "secret") && !strings.Contains(err.Error(), "client_secret") {
				t.Fatal("credential value appeared in error")
			}
		})
	}
}

func TestAssertionCallbackIsNotInvokedByValidation(t *testing.T) {
	t.Parallel()
	called := false
	cfg := &auth.Config{
		TenantID: "tenant", ClientID: "client", Method: auth.WorkloadIdentity,
		Assertion: func(context.Context) (string, error) {
			called = true
			return "assertion", nil
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("configuration validation invoked the assertion callback")
	}
	cfg.AssertionFile = "file"
	if !errors.Is(cfg.Validate(), auth.ErrInvalidConfig) {
		t.Fatal("simultaneous assertion sources must fail")
	}
}

func TestCloneIsolatesMutableCredentials(t *testing.T) {
	t.Parallel()
	cfg := &auth.Config{
		Certificate: []byte("cert"), PrivateKey: []byte("private"),
		AccessTokens: map[string]auth.Token{"audience": {AccessToken: "token"}},
	}
	cloned := cfg.Clone()
	cfg.Certificate[0] = 'X'
	cfg.PrivateKey[0] = 'X'
	cfg.AccessTokens["audience"] = auth.Token{AccessToken: "different"}
	if string(cloned.Certificate) != "cert" || string(cloned.PrivateKey) != "private" {
		t.Fatal("credential byte slices were shared")
	}
	if cloned.AccessTokens["audience"].AccessToken != "token" {
		t.Fatal("token map was shared")
	}
	var absent *auth.Config
	if absent.Clone() != nil {
		t.Fatal("nil clone must remain nil")
	}
}

func TestConfigurationDiagnosticsRedactSecrets(t *testing.T) {
	t.Parallel()
	cfg := &auth.Config{TenantID: "tenant", Method: auth.DeviceCode, ClientID: "client", ClientSecret: "sensitive-value"}
	for _, output := range []string{fmt.Sprintf("%v", cfg), fmt.Sprintf("%+v", cfg), fmt.Sprintf("%#v", cfg), fmt.Sprintf("%+v", *cfg), cfg.Validate().Error()} {
		if strings.Contains(output, "sensitive-value") {
			t.Fatal("configuration diagnostic exposed a secret")
		}
	}
}

func TestPFXConfigurationIsExclusiveAndCopied(t *testing.T) {
	t.Parallel()
	cfg := &auth.Config{Method: auth.Certificate, TenantID: "tenant", ClientID: "client", PFX: []byte("pfx"), PFXPassword: []byte("password")}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	copied := cfg.Clone()
	cfg.PFX[0] = 'X'
	cfg.PFXPassword[0] = 'X'
	if string(copied.PFX) != "pfx" || string(copied.PFXPassword) != "password" {
		t.Fatal("PFX material shared with caller")
	}
	cfg.Certificate = []byte("pem")
	if !errors.Is(cfg.Validate(), auth.ErrInvalidConfig) {
		t.Fatal("mixed PEM/PFX accepted")
	}
	cfg.Certificate = nil
	cfg.Method = auth.ClientSecret
	cfg.ClientSecret = "secret"
	if !errors.Is(cfg.Validate(), auth.ErrInvalidConfig) {
		t.Fatal("PFX mixed with client-secret accepted")
	}
}
