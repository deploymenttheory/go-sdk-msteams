# AUTH-02 client credential evidence and implementation notes

Status: implemented with isolated token-service contract tests. Subsequent
[lab qualification](live-application-validation.md) verifies secret and PEM certificate
authentication with both resource audiences. Management resource coverage and
Connect/Disconnect completeness remain separate capability-ledger states.

## Evidence boundary

The accepted command inventory remains MicrosoftTeams 8.1.0. Microsoft documents
separate Graph and Teams Admin resource tokens for app-only sessions, with client-secret
examples using the client credentials flow and certificate-based Connect examples.
Graph uses `https://graph.microsoft.com/.default`; the commercial Teams Admin resource
uses `48ac35b8-9aa8-4d74-927d-1f4a14a0b239/.default`. These scopes identify resources, not
management API routes. The native provider acquires each independently and retains
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
construct/sign assertions and perform exchanges; the SDK does not implement its own
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

## Implementation and verification

`ApplicationProvider` integrates with root sessions for explicit certificate or client-secret
configuration. It accepts a tenant GUID or DNS domain and currently implements the commercial
cloud profile. GUIDs remain bound to that GUID. Domain discovery may resolve a canonical GUID
only through the exact configured domain's same-authority OIDC response, with matching issuer
and token endpoint. Redirects, alternate hosts, encoded paths and tenant changes are rejected.
Metadata/token responses are capped at 1 MiB and bodies are closed. The injected transport is
borrowed; cookies and custom redirect callbacks are not used for credential exchanges.

MSAL `AcquireTokenByCredential` checks its internal cache before exchanging credentials.
The SDK constructs a fresh MSAL client for each coordinated refresh, while AUTH-01 caches
and serializes tokens per full identity/audience key. Request-count tests prove that entering
the two-minute refresh window results in another credential exchange. MSAL's implicit region
environment setting is explicitly disabled. Library errors are replaced with a stable sentinel;
raw token-service request/response diagnostics are not retained as unwrap causes.

`application_test.go` runs the pinned MSAL library against an isolated HTTP transport. It checks
client secrets, PEM and supported PFX, PS256 assertion signatures, one resource per request,
concurrent Graph/Teams caching, refresh, expired certificates, service failures, domain discovery,
unsafe destinations, redirects, response limits and cancellation. MSAL appends its standard OIDC
scopes to the requested resource scope; this is observed library behavior, not delegated fallback.
`certificate_test.go` and `certificate_pfx_test.go` cover matching, validity, strict formats,
PKCS#1/PKCS#8, copying, PFX passwords and unsupported SHA256 MAC. Root tests verify lazy parsing,
owned provider wiring and borrowed transport lifecycle. The godoc example is executable without
credentials. Race tests, vet/build and lint are required before merge.

## Operation prerequisite examples

These are documentation-based prerequisites for planning, not tenant qualification or a complete
least-privilege permission matrix. Register the app/service principal, attach its certificate or
secret, and grant admin consent for the Graph application permissions selected for its operations.
Token acquisition itself does not establish Teams role authorization.

| Operation category | Teams directory-role prerequisite from published role guidance |
|---|---|
| Meeting policy/configuration administration | Teams Administrator or Teams Communications Administrator |
| Calling policy/phone-number administration | Teams Administrator, Teams Communications Administrator or Teams Telephony Administrator, according to the operation |
| Messaging policy administration | Teams Administrator |
| Resource-account creation/management | Applicable Teams administrator role plus User Administrator; app-only cmdlet/cloud exclusions still apply |

Source: [Teams administrator role capabilities](https://learn.microsoft.com/en-us/microsoftteams/using-admin-roles).
The application-authentication source above supplies the Graph onboarding requirements and
app-only exclusions. Each resource implementation must attach its exact endpoint permission,
role and cloud evidence; these broad role categories do not override cmdlet exclusions or imply
that every listed role is sufficient for every method. Lab conformance will record actual grants
and outcomes separately. No tenant role assignments or consent were changed by this implementation.
