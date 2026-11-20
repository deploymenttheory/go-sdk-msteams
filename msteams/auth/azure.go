package auth

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// NewAzureTokenProvider adapts the same TokenCredential interface used by Microsoft
// Graph's Go authentication provider. The credential and its token cache remain
// caller-owned; the adapter neither constructs a credential chain nor closes it.
//
// Requests are bound to one tenant, cloud, identity and mode. Commercial Graph and
// Teams audiences map to their own /.default scopes, never a scope inferred from
// the management API host. The caller must configure the credential for the stated
// identity, mode and cloud; opaque access tokens are not locally decoded to prove
// those assertions. Backend services still validate destination authority separately.
//
// Returned provider errors use the borrowed-provider redaction boundary. Explicit
// unwrapping may expose diagnostics from the caller's credential implementation.
func NewAzureTokenProvider(binding Binding, credential azcore.TokenCredential) (Provider, error) {
	if err := binding.Validate(); err != nil {
		return nil, err
	}
	if nilDependency(credential) {
		return nil, ErrInvalidProvider
	}
	if binding.Cloud != "commercial" {
		return nil, ErrUnsupportedCloud
	}
	if !validApplicationTenant(binding.TenantID) {
		return nil, &ConfigError{Field: "tenant_id", Reason: "requires a tenant GUID or DNS domain"}
	}
	return NewBoundProvider(binding, azureTokenSource{credential: credential})
}

type azureTokenSource struct {
	credential azcore.TokenCredential
}

func (s azureTokenSource) AcquireToken(ctx context.Context, request TokenRequest) (Token, error) {
	if request.Audience != GraphAudience && request.Audience != TeamsAdminAudience {
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
