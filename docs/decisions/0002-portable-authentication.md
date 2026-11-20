# Decision 0002 Portable resource-aware authentication

Status: accepted architecture. Implements the authentication boundary of [ARCH-01](https://github.com/deploymenttheory/go-sdk-msteams/issues/21); AUTH tickets deliver the providers and qualification.

## Decision

Use Microsoft MSAL Go and Azure Identity implementations behind an SDK-owned provider interface. Support explicit browser, device-code, username/password where Entra permits it, certificate, client-secret, supplied-token, managed-identity and workload-federation providers. Browser and device callbacks are caller-controlled; headless providers never silently open a browser or switch accounts.

Microsoft documents separate Graph and Teams administration tokens and operation-specific app-only exceptions. Those distinctions are retained in the operation matrix, rather than treating successful sign-in as authorization for every command. Sources: [Connect-MicrosoftTeams](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/connect-microsoftteams?view=teams-ps) and [application authentication](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication). The pinned module and observations settle 8.1.0-specific behaviour when evolving documentation differs.

The provider boundary uses the following conceptual Go types:

```go
type TokenRequest struct {
    TenantID   string
    Cloud      Cloud
    Audience   string
    IdentityID string
    Mode       Mode // delegated or application; zero is invalid
}

type Token struct {
    AccessToken string
    ExpiresAt   time.Time
}

type Provider interface {
    AcquireToken(context.Context, TokenRequest) (Token, error)
}
```

Provider construction binds credentials and identity. The transport supplies the required backend audience and the same bound tenant/cloud/identity/mode for every request in that session. A provider must reject incompatible requests rather than reinterpret them. Typed identifiers and field details can be refined within AUTH-01, but all five isolation dimensions are required.

The cache key includes tenant, cloud, identity, mode and audience. Refresh is coordinated per key and respects cancellation. An in-flight request cannot receive another tenant's token after a client reconfiguration; configuration is immutable after construction. Callers construct a new client for a new identity. A failed refresh is not cached as success, and no process-global credential cache is introduced implicitly.

Supplied tokens are explicitly mapped to audiences and expirations. Static tokens return expiry errors; automatic refresh requires a callback/provider. The SDK does not assume every token is a locally parseable JWT or claim that decoding JWT claims authenticates the issuer. Backend authorization remains authoritative. Delegated/application tokens must not be mixed in one configured session.

Certificates use portable file or supplied key/certificate material, with errors for invalid pairs. Windows certificate-store thumbprints and WAM are replaced with documented portable equivalents. Managed identity is explicitly selected and reports unavailable environments clearly. Federation obtains assertions from a caller callback or file and exchanges them with Entra; the SDK does not discover CI credentials through an unrestricted fallback chain.

## Lifecycle, errors and confidentiality

Construction validates local configuration only. Acquisition happens under the operation's context; discovery and token calls share the caller's cancellation boundary. Unsupported auth modes, unavailable identity environments, token acquisition failures and service 401/403 responses remain distinguishable.

Client session cleanup clears owned cached material and closes owned idle transport resources. It does not claim remote revocation of an Entra access token and does not clear caches owned by caller-supplied providers. No token, assertion, password or private key appears in error text, Zap logs, debug payloads or OTel attributes. The logger is injected and defaults to no-op; tracing is opt-in with caller-owned lifecycle.

## Verification requirements

AUTH tests cover audience/tenant isolation, expiry, simultaneous refresh, cancellation, revoked credentials, MFA/conditional-access failures, supplied-token refresh limitations and unavailable managed identity. They must run without tenant credentials on Linux, macOS and Windows. Lab evidence is recorded separately by auth mode and operation. A skipped live environment is not an implemented-provider failure or a live-test pass.
