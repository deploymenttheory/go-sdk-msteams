package auth_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

type azureCredentialFunc func(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error)

func (f azureCredentialFunc) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return f(ctx, options)
}

func azureBinding() auth.Binding {
	return auth.Binding{TenantID: "fixture.onmicrosoft.com", Cloud: "commercial", IdentityID: "fixture-user", Mode: auth.Delegated}
}

func TestAzureCredentialPreservesRegionScopeAndMigrationResource(t *testing.T) {
	t.Parallel()
	const audience = "https://migrationsvc.teams.microsoft.com"
	const scope = audience + "/Region.Read.All"
	for _, cloud := range []string{"commercial", "gcc"} {
		t.Run(cloud, func(t *testing.T) {
			binding := azureBinding()
			binding.Cloud = cloud
			var calls []policy.TokenRequestOptions
			provider, err := auth.NewAzureTokenProvider(binding, azureCredentialFunc(func(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
				calls = append(calls, options)
				return azcore.AccessToken{Token: "fixture", ExpiresOn: time.Now().Add(time.Hour)}, nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			for _, rejected := range []auth.TokenRequest{binding.Request(audience), binding.RequestScope(audience, audience+"/.default"), binding.RequestScope(audience, audience+"/Region.ReadWrite.All")} {
				if _, err := provider.AcquireToken(t.Context(), rejected); !errors.Is(err, auth.ErrUnsupportedScope) {
					t.Fatal("region scope widening reached credential")
				}
			}
			if len(calls) != 0 {
				t.Fatal("rejected scope acquired a token")
			}
			for _, request := range []auth.TokenRequest{binding.RequestScope(audience, scope), binding.Request("https://templates.teams.microsoft.com")} {
				if _, err := provider.AcquireToken(t.Context(), request); err != nil {
					t.Fatal(err)
				}
			}
			if len(calls) != 2 || calls[0].TenantID != binding.TenantID || !reflect.DeepEqual(calls[0].Scopes, []string{scope}) || !reflect.DeepEqual(calls[1].Scopes, []string{"https://templates.teams.microsoft.com/.default"}) {
				t.Fatal("borrowed credential received a changed tenant/resource/scope")
			}
		})
	}
}

func TestAzureCredentialCustomEnvironmentBinding(t *testing.T) {
	t.Parallel()
	endpoints := &auth.Endpoints{Authority: "https://identity.example.test", Graph: "https://graph.example.test", TeamsAdmin: "https://teams.example.test", TeamsAdminResource: "api://teams.example.test"}
	binding := azureBinding()
	binding.Cloud = "custom"
	calls := 0
	credential := azureCredentialFunc(func(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
		calls++
		if len(options.Scopes) != 1 || options.Scopes[0] != "https://graph.example.test/.default" {
			t.Error("custom resource scope changed")
		}
		return azcore.AccessToken{Token: "fixture", ExpiresOn: time.Now().Add(time.Hour)}, nil
	})
	provider, err := auth.NewAzureTokenProviderWithEndpoints(binding, credential, endpoints)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &auth.Config{Cloud: "custom", Endpoints: endpoints, Method: auth.ExternalProvider, TenantID: binding.TenantID, IdentityID: binding.IdentityID, TokenMode: binding.Mode, Provider: provider}
	request := cfg.TokenBinding().Request(endpoints.Graph)
	composed, err := auth.NewAzureTokenProviderWithEndpoints(cfg.TokenBinding(), credential, endpoints)
	if err != nil {
		t.Fatalf("adapter rejected Config.TokenBinding: %v", err)
	}
	if _, err := composed.AcquireToken(t.Context(), request); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.AcquireToken(t.Context(), request); err != nil {
		t.Fatal(err)
	}
	endpoints.Authority = "https://other.example.test"
	if _, err := auth.NewAzureTokenProviderWithEndpoints(auth.Binding{TenantID: request.TenantID, Cloud: request.Cloud, IdentityID: request.IdentityID, Mode: request.Mode}, credential, endpoints); !errors.Is(err, auth.ErrBindingMismatch) {
		t.Fatal("adapter accepted a fingerprint from different endpoints")
	}
	if _, err := provider.AcquireToken(t.Context(), cfg.TokenBinding().Request(request.Audience)); !errors.Is(err, auth.ErrBindingMismatch) {
		t.Fatalf("different authority shared adapter: %v", err)
	}
	if calls != 2 {
		t.Fatal("unbound request reached credential")
	}
}

func TestAzureCredentialAdapterBindsTenantAndResourceScopes(t *testing.T) {
	t.Parallel()
	binding := azureBinding()
	var calls []policy.TokenRequestOptions
	expiry := time.Now().Add(time.Hour)
	credential := azureCredentialFunc(func(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
		calls = append(calls, options)
		return azcore.AccessToken{Token: "fixture-token", ExpiresOn: expiry}, nil
	})
	provider, err := auth.NewAzureTokenProvider(binding, credential)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatal("constructor acquired a token")
	}
	for _, audience := range []string{auth.GraphAudience, auth.TeamsAdminAudience} {
		token, err := provider.AcquireToken(t.Context(), binding.Request(audience))
		if err != nil || token.AccessToken != "fixture-token" || !token.ExpiresAt.Equal(expiry) {
			t.Fatalf("adapter lost token result: %v", err)
		}
		request := calls[len(calls)-1]
		if request.TenantID != binding.TenantID || !reflect.DeepEqual(request.Scopes, []string{audience + "/.default"}) {
			t.Fatal("adapter changed tenant or resource scope")
		}
	}
	if _, err := provider.AcquireToken(t.Context(), binding.Request(auth.GraphAudience)); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 {
		t.Fatal("adapter unexpectedly owns a second token cache")
	}
}

func TestAzureCredentialAdapterRejectsUnboundAcquisition(t *testing.T) {
	t.Parallel()
	binding := azureBinding()
	credential := azureCredentialFunc(func(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
		t.Error("invalid request reached borrowed credential")
		return azcore.AccessToken{}, errors.New("unexpected")
	})
	provider, err := auth.NewAzureTokenProvider(binding, credential)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*auth.TokenRequest){
		func(r *auth.TokenRequest) { r.TenantID = "other.onmicrosoft.com" },
		func(r *auth.TokenRequest) { r.Cloud = "government" },
		func(r *auth.TokenRequest) { r.IdentityID = "other-user" },
		func(r *auth.TokenRequest) { r.Mode = auth.Application },
	} {
		request := binding.Request(auth.GraphAudience)
		change(&request)
		if _, err := provider.AcquireToken(t.Context(), request); !errors.Is(err, auth.ErrBindingMismatch) {
			t.Fatal("binding mismatch accepted")
		}
	}
	if _, err := provider.AcquireToken(t.Context(), binding.Request("https://attacker.invalid")); !errors.Is(err, auth.ErrUnsupportedAudience) {
		t.Fatal("unknown audience accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := provider.AcquireToken(ctx, binding.Request(auth.GraphAudience)); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled acquisition accepted")
	}
	var missing azureCredentialFunc
	if _, err := auth.NewAzureTokenProvider(binding, missing); !errors.Is(err, auth.ErrInvalidProvider) {
		t.Fatal("typed-nil credential accepted")
	}
	binding.Cloud = "government"
	if _, err := auth.NewAzureTokenProvider(binding, credential); !errors.Is(err, auth.ErrUnsupportedCloud) {
		t.Fatal("unsupported cloud accepted")
	}
}

func TestAzureCredentialAdapterValidatesTokensAndRedactsErrors(t *testing.T) {
	t.Parallel()
	binding := azureBinding()
	sourceError := errors.New("fixture-secret response body")
	for _, tc := range []struct {
		name            string
		token           azcore.AccessToken
		sourceErr, want error
	}{
		{"empty", azcore.AccessToken{}, nil, auth.ErrInvalidToken},
		{"expired", azcore.AccessToken{Token: "fixture-token", ExpiresOn: time.Now().Add(-time.Minute)}, nil, auth.ErrTokenExpired},
		{"foreign-error", azcore.AccessToken{}, sourceError, sourceError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := auth.NewAzureTokenProvider(binding, azureCredentialFunc(func(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
				return tc.token, tc.sourceErr
			}))
			if err != nil {
				t.Fatal(err)
			}
			_, err = provider.AcquireToken(t.Context(), binding.Request(auth.GraphAudience))
			if !errors.Is(err, tc.want) {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, format := range []string{"%v", "%+v", "%#v"} {
				if strings.Contains(fmt.Sprintf(format, err), "fixture-secret") {
					t.Fatal("credential diagnostics leaked through formatting")
				}
			}
		})
	}
}

func TestAzureCredentialAdapterPreservesCancellationContext(t *testing.T) {
	t.Parallel()
	binding := azureBinding()
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	provider, err := auth.NewAzureTokenProvider(binding, azureCredentialFunc(func(received context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
		if received != ctx {
			t.Error("caller context was replaced")
		}
		cancel()
		return azcore.AccessToken{Token: "fixture-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.AcquireToken(ctx, binding.Request(auth.GraphAudience)); !errors.Is(err, context.Canceled) {
		t.Fatal("late result escaped cancellation")
	}
}
