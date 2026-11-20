# AUTH-02 client credential evidence and implementation notes

Status: implementation in progress. Certificate parsing helpers are local work; token
acquisition and root-session integration are not yet delivered by AUTH-02. No tenant
credentials have been supplied and no app-only live result is claimed.

## Evidence boundary

The accepted command inventory remains MicrosoftTeams 8.1.0. Microsoft documents
separate Graph and Teams Admin resource tokens for app-only sessions, with client-secret
examples using the client credentials flow and certificate-based Connect examples.
Graph uses `https://graph.microsoft.com/.default`; the commercial Teams Admin resource
uses `48ac35b8-9aa8-4d74-927d-1f4a14a0b239/.default`. These scopes identify resources, not
management API routes. The native provider must acquire each independently and retain
AUTH-01's tenant/cloud/application identity binding.

App registration, consent for the selected Graph operations and suitable Teams directory
roles are separate from successfully obtaining a token. Microsoft's onboarding guidance
calls out `Organization.Read.All`, additional Graph permissions for non-Cs operations,
and directory roles for Teams administration. It warns against adding an API permission
for the Skype and Teams Tenant Admin resource. The documented app-only exclusions vary
by cmdlet/cloud; a successful token cannot mark every management operation supported.
Operation research tickets must retain their own permission and role evidence.

Source: [Microsoft Teams application authentication](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication),
reviewed 2026-11-20. This is current documentation, not proof that an 8.1.0 tenant call
succeeded or that later documentation changes alter the pinned baseline.

The [Microsoft identity client credentials protocol](https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-client-creds-grant-flow)
uses one resource's `.default` scope per token request. Microsoft authentication libraries
will construct/sign assertions and perform exchanges; the SDK will not implement its own
token protocol or silently fall back to delegated login.

## Pinned library findings

Inspected source already pinned in `go.mod`:

- MSAL Go v1.10.1 `apps/confidential`: `CertFromPEM` accepts an unencrypted PEM certificate
  chain with PKCS#1/PKCS#8 private key material. It rejects legacy encrypted PEM; its
  historical password argument is ignored. `NewCredFromCert` requires an RSA key and a
  matching certificate. `NewCredFromSecret` and `AcquireTokenByCredential` provide the
  application flow; `WithHTTPClient` permits isolated token-service tests.
- Azure Identity v1.15.0 `ParseCertificates`: supports PEM and a limited PKCS#12/PFX
  decoder. The pinned decoder cannot handle PKCS#12 SHA256 MAC or encrypted PEM keys.
  PFX support must be documented and tested as that explicit subset; modern unsupported
  bundles must fail clearly. Equivalent unencrypted PEM remains a portable input.

Library references: [MSAL confidential source](https://github.com/AzureAD/microsoft-authentication-library-for-go/blob/v1.10.1/apps/confidential/confidential.go),
[Azure Identity certificate source](https://github.com/Azure/azure-sdk-for-go/blob/sdk/azidentity/v1.15.0/sdk/azidentity/client_certificate_credential.go).

## Remaining AUTH-02 work

Integrate explicit PEM/PFX configuration and lazy Microsoft-library client-credentials
providers with root sessions, preserving construction without authentication. Validate
certificate matching and validity before signing. Keep the commercial authority fixed
until trusted cloud profiles are implemented; do not accept arbitrary token authorities
from environment strings. Preserve context, redacted errors, resource-specific caches
and owned-provider cleanup. Ensure the Microsoft library's own token cache does not
undermine the outer refresh window or leak identity between instances.

PEM ownership, PKCS#1/PKCS#8 matching, validity boundaries, strict format selection,
supported PFX success/password failure and unsupported SHA256 MAC rejection already
have synthetic tests. Next test both credential modes through isolated Microsoft-library
HTTP fixtures, independent audience scopes, invalid/expired credentials,
service errors and cancellation. Extend compiled examples and configuration documentation;
run race/lint/platform CI before opening the implementation PR. Live qualification remains
pending the lab tenant and its app roles/consent.

MSAL `AcquireTokenByCredential` checks its internal cache before exchanging credentials.
The provider must account for this when layering AUTH-01 refresh skew; simply reusing a
long-lived MSAL client can return the same near-expiry token. A fresh MSAL client per
coordinated refresh is a possible implementation because client credentials have no
refresh token. Validate this choice with token-service request counts before adopting it.
