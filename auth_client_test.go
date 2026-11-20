package msteams_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	msteams "github.com/deploymenttheory/go-sdk-msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

type testProvider struct {
	acquire func(context.Context, auth.TokenRequest) (auth.Token, error)
	closes  atomic.Int32
}

func (p *testProvider) AcquireToken(ctx context.Context, r auth.TokenRequest) (auth.Token, error) {
	return p.acquire(ctx, r)
}
func (p *testProvider) Close() { p.closes.Add(1) }

func TestSessionSuppliedTokensAreCopiedAndAudienceIsolated(t *testing.T) {
	t.Parallel()
	cfg := &auth.Config{Method: auth.SuppliedTokens, TenantID: "tenant", IdentityID: "identity", TokenMode: auth.Application,
		AccessTokens: map[string]auth.Token{"graph": {AccessToken: "graph-secret", ExpiresAt: time.Now().Add(time.Hour)}}}
	sdk, err := msteams.NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	cfg.AccessTokens["graph"] = auth.Token{AccessToken: "replacement"}
	session := sdk.GraphAPI.Scope().Session()
	token, err := session.AcquireToken(t.Context(), "graph")
	if err != nil || token.AccessToken != "graph-secret" {
		t.Fatalf("copied token unavailable: %v", err)
	}
	if _, err := session.AcquireToken(t.Context(), "teams"); !errors.Is(err, auth.ErrTokenMissing) {
		t.Fatalf("unexpected audience fallback: %v", err)
	}
	if err := sdk.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := session.AcquireToken(t.Context(), "graph"); !errors.Is(err, client.ErrClosed) {
		t.Fatalf("closed session: %v", err)
	}
	if cfg.AccessTokens["graph"].AccessToken != "replacement" {
		t.Fatal("close cleared caller token map")
	}
}

func TestExternalProviderReceivesBindingAndRemainsBorrowed(t *testing.T) {
	t.Parallel()
	binding := auth.Binding{TenantID: "tenant", Cloud: "commercial", IdentityID: "identity", Mode: auth.Delegated}
	var calls atomic.Int32
	source := &testProvider{acquire: func(_ context.Context, r auth.TokenRequest) (auth.Token, error) {
		calls.Add(1)
		if r != binding.Request("teams") {
			t.Error("session changed request binding")
		}
		return auth.Token{AccessToken: "teams-secret", ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}
	sdk, err := msteams.NewClient(nil, msteams.WithTokenProvider(binding, source))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	if calls.Load() != 0 {
		t.Fatal("constructor acquired a token")
	}
	token, err := sdk.TeamsAdminAPI.Scope().Session().AcquireToken(t.Context(), "teams")
	if err != nil || token.AccessToken != "teams-secret" {
		t.Fatalf("acquire: %v", err)
	}
	if err := sdk.Close(); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 || source.closes.Load() != 0 {
		t.Fatal("SDK changed borrowed provider lifecycle")
	}
}

func TestSessionCloseCancelsBorrowedProviderOperation(t *testing.T) {
	t.Parallel()
	entered := make(chan struct{})
	exited := make(chan struct{})
	source := &testProvider{acquire: func(ctx context.Context, _ auth.TokenRequest) (auth.Token, error) {
		close(entered)
		<-ctx.Done()
		close(exited)
		return auth.Token{}, ctx.Err()
	}}
	sdk, err := msteams.NewClient(nil, msteams.WithTokenProvider(auth.Binding{TenantID: "tenant", Cloud: "commercial", IdentityID: "id", Mode: auth.Application}, source))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	result := make(chan error, 1)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	go func() { _, err := sdk.GraphAPI.Scope().Session().AcquireToken(ctx, "graph"); result <- err }()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("provider did not start")
	}
	if err := sdk.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, client.ErrClosed) {
			t.Fatalf("close: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("close did not cancel acquisition")
	}
	<-exited
	if source.closes.Load() != 0 {
		t.Fatal("closed borrowed provider")
	}
}

func TestExternalProviderRejectsInvalidConfigurationAndRedactsErrors(t *testing.T) {
	t.Parallel()
	binding := auth.Binding{TenantID: "tenant", Cloud: "commercial", IdentityID: "id", Mode: auth.Application}
	var absent *testProvider
	if _, err := msteams.NewClient(nil, msteams.WithTokenProvider(binding, absent)); !errors.Is(err, auth.ErrInvalidConfig) {
		t.Fatalf("typed nil provider: %v", err)
	}
	original := errors.New("sensitive-provider-diagnostic")
	source := &testProvider{acquire: func(context.Context, auth.TokenRequest) (auth.Token, error) {
		return auth.Token{AccessToken: "sensitive-token"}, original
	}}
	cfg := &auth.Config{Method: auth.ExternalProvider, TenantID: "tenant", IdentityID: "id", TokenMode: auth.Application, Provider: source, ClientSecret: "secret"}
	if _, err := msteams.NewClient(cfg); !errors.Is(err, auth.ErrInvalidConfig) {
		t.Fatalf("mixed credentials: %v", err)
	}
	sdk, err := msteams.NewClient(nil, msteams.WithTokenProvider(binding, source))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	token, err := sdk.GraphAPI.Scope().Session().AcquireToken(t.Context(), "graph")
	if !errors.Is(err, original) || token.AccessToken != "" {
		t.Fatal("source error identity or failed token handling")
	}
	for _, output := range []string{fmt.Sprintf("%v", err), fmt.Sprintf("%+v", err), fmt.Sprintf("%#v", err)} {
		if strings.Contains(output, "sensitive") {
			t.Fatal("provider diagnostic leaked")
		}
	}
}

func TestSessionDoesNotPretendUnimplementedProviderCanAcquire(t *testing.T) {
	t.Parallel()
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.ManagedIdentity, TenantID: "tenant"})
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	if _, err := sdk.GraphAPI.Scope().Session().AcquireToken(t.Context(), "graph"); !errors.Is(err, client.ErrProviderUnavailable) {
		t.Fatalf("adapter status: %v", err)
	}
}

func TestApplicationSessionUsesOwnedProviderWithoutConstructorAuthentication(t *testing.T) {
	t.Parallel()
	transport := &recordingTransport{}
	cfg := &auth.Config{Method: auth.ClientSecret, TenantID: "11111111-1111-4111-8111-111111111111", ClientID: "22222222-2222-4222-8222-222222222222", ClientSecret: "synthetic-secret"}
	sdk, err := msteams.NewClient(cfg, msteams.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	if transport.requests.Load() != 0 {
		t.Fatal("construction contacted token service")
	}
	session := sdk.GraphAPI.Scope().Session()
	if _, err := session.AcquireToken(t.Context(), "unknown-resource"); !errors.Is(err, auth.ErrUnsupportedAudience) {
		t.Fatalf("provider not wired: %v", err)
	}
	if err := sdk.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := session.AcquireToken(t.Context(), auth.GraphAudience); !errors.Is(err, client.ErrClosed) {
		t.Fatalf("session close: %v", err)
	}
	if transport.requests.Load() != 0 || transport.closes.Load() != 0 {
		t.Fatal("provider used or closed caller transport unexpectedly")
	}
}

func TestCertificateParsingRemainsLazyAndErrorsStayLocal(t *testing.T) {
	t.Parallel()
	transport := &recordingTransport{}
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.Certificate, TenantID: "11111111-1111-4111-8111-111111111111", ClientID: "22222222-2222-4222-8222-222222222222", PFX: []byte("invalid-pfx")}, msteams.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	closeClient(t, sdk)
	if _, err := sdk.TeamsAdminAPI.Scope().Session().AcquireToken(t.Context(), auth.TeamsAdminAudience); !errors.Is(err, auth.ErrInvalidCertificate) {
		t.Fatalf("certificate parsing: %v", err)
	}
	if transport.requests.Load() != 0 {
		t.Fatal("invalid certificate reached network")
	}
}
