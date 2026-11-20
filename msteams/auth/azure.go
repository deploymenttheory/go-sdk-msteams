package auth

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// NewAzureTokenProvider adapts the same TokenCredential interface used by Microsoft
// Graph's Go authentication provider. The credential and its token cache remain
// caller-owned; the adapter neither constructs a credential chain nor closes it.
//
// Requests are bound to one tenant, cloud, identity and mode. The selected Graph and
// Teams audiences map to their own /.default scopes, never a scope inferred from
// the management API host. The caller must configure the credential for the stated
// identity, mode and cloud; opaque access tokens are not locally decoded to prove
// those assertions. Backend services still validate destination authority separately.
//
// Returned provider errors use the borrowed-provider redaction boundary. Explicit
// unwrapping may expose diagnostics from the caller's credential implementation.
func NewAzureTokenProvider(binding Binding, credential azcore.TokenCredential) (Provider, error) {
	return NewAzureTokenProviderWithEndpoints(binding, credential, nil)
}

// NewAzureTokenProviderWithEndpoints adapts a borrowed credential with explicit
// trusted endpoint overrides. Pass the same endpoints as Config and either its
// TokenBinding or a binding with the raw cloud label. A supplied fingerprint must
// match the resolved tuple exactly.
// The caller must configure the borrowed credential's authority independently.
func NewAzureTokenProviderWithEndpoints(binding Binding, credential azcore.TokenCredential, endpoints *Endpoints) (Provider, error) {
	if err := binding.Validate(); err != nil {
		return nil, err
	}
	if nilDependency(credential) {
		return nil, ErrInvalidProvider
	}
	cloud, _, fingerprinted := strings.Cut(binding.Cloud, "@")
	environment, err := ResolveEnvironment(cloud, endpoints)
	if err != nil {
		return nil, err
	}
	if fingerprinted && binding.Cloud != environment.BindingCloud {
		return nil, ErrBindingMismatch
	}
	// Use canonical cloud labels in binding keys, just as Config.TokenBinding does.
	binding.Cloud = environment.BindingCloud
	if !validApplicationTenant(binding.TenantID) {
		return nil, &ConfigError{Field: "tenant_id", Reason: "requires a tenant GUID or DNS domain"}
	}
	return NewBoundProvider(binding, azureTokenSource{credential: credential, environment: environment})
}

type azureTokenSource struct {
	credential  azcore.TokenCredential
	environment Environment
}

func (s azureTokenSource) AcquireToken(ctx context.Context, request TokenRequest) (Token, error) {
	if !s.environment.supportsAudience(request.Audience) {
		return Token{}, ErrUnsupportedAudience
	}
	result, err := s.credential.GetToken(ctx, policy.TokenRequestOptions{
		TenantID: request.TenantID,
		Scopes:   []string{request.Audience + "/.default"},
	})
	if err != nil {
		return Token{}, err
	}
	return Token{AccessToken: result.Token, ExpiresAt: result.ExpiresOn}, nil
}
