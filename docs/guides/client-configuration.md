# Client construction and ownership

`NewClient(authConfig, options...)` consumes explicit configuration. It never reads
`MSTEAMS_*` variables. `NewClientFromEnv(options...)` reads the documented variables,
applies options in order, and then parses and validates the resulting configuration.
`WithAuthConfig` replaces the complete authentication configuration; it also discards
an environment token payload. `WithHTTPTimeout` can replace an invalid environment timeout.

Construction validates authentication selection, required credential presence, incompatible
credential combinations and HTTP settings. It does not authenticate, call assertion
callbacks, read assertion files, parse certificate/key material, validate token signatures
or check token expiry. Supplied-token and injected-provider acquisition now check expiry
when an operation requests a token; certificate/client-secret acquisition is implemented; delegated, managed-identity and workload
identity adapters remain separate AUTH tickets.
A successfully constructed client is not proof that credentials work or an operation is
supported. Resource methods are added with their research and implementation tickets.

```go
sdk, err := msteams.NewClient(&msteams.AuthConfig{
    Method: auth.ManagedIdentity,
    TenantID: "your-tenant-id",
}, msteams.WithHTTPTimeout(30*time.Second))
// Handle err, use implemented services, and call sdk.Close() when finished.
```

The [compiled example](../../examples/client-construction/main.go) constructs and closes
a session without connecting to Microsoft. It runs on Linux, macOS and Windows.

## Authentication selection

The `auth.Method` constants select browser, device-code, username/password, certificate,
client-secret, supplied-token, managed-identity or workload-identity configuration.
`WithTokenProvider(binding, provider)` selects an injected provider and replaces the
complete authentication configuration, including environment token material.
There is no implicit credential chain or fallback to browser login. Availability and
operation-specific permissions remain provider/contract qualification concerns.

Every method requires a tenant ID. Browser/device-code require a client ID and an explicit callback; password
also requires username/password; certificate requires supplied certificate/private-key
bytes; client-secret requires a secret. Managed identity accepts an optional client ID
for a user-assigned identity. Workload identity requires a client ID and exactly one
assertion file or callback. Supplied tokens require an audience-keyed map with explicit
expirations, a delegated/application mode, and an identity ID for cache isolation.
Incompatible credential groups are rejected without including their values in errors.

The root copies credential byte slices, token maps and HTTP client values. Callers must
not mutate input concurrently with construction. Assertion callbacks, injected round
trippers, cookie jars, logger and tracer provider remain caller-owned; they must be safe
for their intended concurrent use. The SDK does not shut them down.

## Environment variables

| Variable | Meaning |
|---|---|
| `MSTEAMS_AUTH_METHOD` | Explicit method string from `auth.Method` |
| `MSTEAMS_TENANT_ID` | Tenant binding |
| `MSTEAMS_CLOUD` | Token isolation profile, defaults to `commercial`; does not enable or discover cloud endpoints |
| `MSTEAMS_CLIENT_ID` | Application or user-assigned managed identity client ID |
| `MSTEAMS_CLIENT_SECRET` | Client-secret material |
| `MSTEAMS_USERNAME`, `MSTEAMS_PASSWORD` | Password-flow material |
| `MSTEAMS_CERTIFICATE_PEM`, `MSTEAMS_PRIVATE_KEY_PEM` | Supplied certificate/key bytes |
| `MSTEAMS_ASSERTION_FILE` | Workload assertion file path (read by the future provider) |
| `MSTEAMS_ACCESS_TOKENS` | JSON object keyed by audience; values have `access_token` and RFC3339 `expires_at` |
| `MSTEAMS_TOKEN_MODE` | `delegated` or `application`, for supplied tokens |
| `MSTEAMS_IDENTITY_ID` | Supplied-token identity binding |
| `MSTEAMS_HTTP_TIMEOUT` | Positive Go duration such as `30s` |

Callbacks are configured in Go, never loaded from the environment. Environment values
are not printed by constructor errors. Missing/malformed settings produce errors that
can be matched with `errors.Is(err, auth.ErrInvalidConfig)` or `msteams.ErrInvalidOption`.
Treat configuration objects and token snapshots as sensitive, even though ordinary
formatted authentication diagnostics are redacted.

## HTTP, backend groups and lifecycle

`GraphAPI` and `TeamsAdminAPI` are distinct typed service collections, each bound to the
same per-client session and its own backend identity. They currently expose the dependency
scope for future resource constructors; no placeholder REST management methods exist.
Scopes are immutable. Different root clients receive separate sessions and default
transports. Each scope implements `client.Client.Do(ctx, request, result)` for authenticated
JSON requests. The commercial profile maps Graph to `graph.microsoft.com` and Teams
ConfigAPI to `api.interfaces.records.teams.microsoft.com`, selecting distinct token
audiences. Construction still performs no authentication, discovery or management request.

Use `client.Request.Path` for unescaped segments (including Graph's explicit `v1.0` or
`beta` prefix) and `Query` for query values. `URL` instead accepts an absolute approved
continuation/discovery destination and cannot be combined with Path/Query. Other hosts,
cloud profiles and ambiguous destinations fail before credentials are acquired. Resource
services remain responsible for evidenced API paths and payloads.

`Do` sends one attempt and never follows redirects. It returns buffered Resty response
metadata with closed bodies, caps responses at `client.MaxResponseBytes` (8 MiB), and
decodes nonempty successful JSON into the supplied result. Nil result preserves raw
response access. Non-2xx responses return `*client.APIError` with status, service codes
and diagnostic IDs; standard error formatting omits untrusted server values. Raw response
headers/bodies, explicit diagnostic fields and request payloads may contain tenant data.
Transport/decode failures use safe sentinel errors and context cancellation is preserved.
Automatic retries, paging, binary uploads and downloads are separate HTTP tickets.
See [HTTP-01 evidence and tests](../implementation/http/routing-evidence.md).

HTTP timeout defaults to 30 seconds. An injected client's positive timeout is retained
unless explicitly overridden. Negative injected timeouts are rejected unless replaced
by an explicit timeout. Typed-nil injected transports are rejected. If its transport is nil, the SDK supplies and owns one;
an explicitly supplied transport remains borrowed. The SDK copies the client value and
never mutates the caller's HTTP settings. Zap defaults to no-op, while tracing is disabled
unless a non-nil provider is supplied. Neither uses a process-global SDK client.

`Session.Begin(ctx)` preserves caller cancellation and additionally ties the operation
to session closure. Call its returned cancellation function on every completion path.
`Client.Close()` is idempotent, cancels local work, clears owned credential references
and mutable copies, and closes owned idle connections. It does not revoke Entra access
tokens, erase caller snapshots or clear caller-owned caches. Go string storage cannot
promise cryptographic memory erasure. Closed sessions reject subsequent operations and
credential snapshots.

Verification covers explicit/environment precedence, malformed overrides, zero-request
construction, borrowed-resource ownership, credential cloning, independent concurrent
clients, cancellation and redacted diagnostics. Provider token acquisition and tenant
conformance are deliberately separate qualification states in the capability ledger.

## Supplied tokens and refresh callbacks

`auth.NewSuppliedTokenProvider(binding, tokens)` copies an audience-keyed token inventory.
`auth.NewRefreshProvider(binding, callback, options)` lazily acquires and caches tokens.
Both implement `auth.Provider`. Each request carries the tenant, cloud, identity, mode
and exact audience. A mismatch fails before refresh; missing static audiences fail with
`auth.ErrTokenMissing`. Static expired tokens return `auth.ErrTokenExpired` without refresh.
Opaque token metadata is the caller/provider's assertion; the SDK does not prove issuer,
audience, delegated permissions or server acceptance by decoding a token.

Inject a refresh provider using `WithTokenProvider(binding, provider)`. The callback receives
the fully bound `auth.TokenRequest` and returns `auth.Token` with an explicit expiry. It
must honor cancellation and be safe for concurrent requests for different audiences.
The constructor makes no callback. A positive `RefreshSkew` refreshes ahead of expiration;
zero accepts any unexpired token. Invalid or already expired callback results are rejected;
tokens inside the configured refresh window return `auth.ErrTokenExpiresSoon`.

Concurrent callers for the same complete key share one refresh. Cancelling one waiter
leaves work needed by other waiters running. When all waiters leave, the callback receives
cancellation; replacement work waits for the old callback to exit. Shared work retains
the first caller's context values, with cancellation controlled by the waiter group.
It has no independent deadline: use operation deadlines and ensure callbacks honor their
context. Callback failures are not cached. Their printable errors are redacted, while
`errors.Is`/`errors.As` preserve deliberate inspection of caller-owned causes; do not log
unwrapped source diagnostics if they contain credentials.

`Session.AcquireToken(ctx, audience)` is the internal resource-adapter boundary. It supplies
the session's immutable binding, validates borrowed provider results and ties work to
session closure. Endpoint-to-audience routing belongs to the HTTP implementation; this
ticket adds no endpoint guesses or management methods. Unimplemented built-in acquisition
adapters return `client.ErrProviderUnavailable`.

Client closure clears its copied static tokens and cancels pending acquisitions. Injected
providers and their caches remain caller-owned. Owners of a refresh provider call `Close()`
then `Wait()` to cancel work, clear its cache and await callback exit; a callback that
ignores cancellation cannot be forcibly stopped by Go. All clients sharing a provider
must use the same binding. Use separate providers for different identities or modes.

AUTH-01 evidence: `msteams/auth/provider_test.go`, `refresh_test.go`, `bound_test.go` and
`msteams/auth_client_test.go` cover audience/binding isolation, expiry, refresh races, cancellation,
redaction and ownership using synthetic tokens. These support the connection/session
foundation (`Connect-MicrosoftTeams`/`Disconnect-MicrosoftTeams`); they do not mark those
commands fully implemented or live-tested. No lab tenant was contacted.

## Application authentication

`auth.ClientSecret` and `auth.Certificate` now acquire commercial-cloud tokens through MSAL
Go. Set a specific tenant GUID or DNS domain and the application/client ID. Application
identity/mode are derived from that client ID; supplying delegated-mode or unrelated credential
fields is rejected. Constructors remain local and defer certificate parsing until acquisition.
The provider accepts `auth.GraphAudience` and `auth.TeamsAdminAudience` independently.

For certificate configuration, choose either `Certificate` plus `PrivateKey` (unencrypted
certificate-chain PEM and one RSA PKCS#1/PKCS#8 key) or `PFX` plus optional `PFXPassword` bytes.
The choices are exclusive. Callers can load files with `os.ReadFile` and pass the bytes; the
SDK never accesses Windows stores or discovers credentials. PFX supports the pinned Microsoft
parser's SHA1-MAC subset; SHA256-MAC PFX and encrypted PEM keys are rejected. These limitations
are documented/tested, not a recommendation to downgrade production encryption. Callers can
supply equivalent unencrypted PEM material through their existing secret-management process.
Mutable input is copied. Certificate/key mismatches, ambiguous signing certificates and validity
failures are rejected before contacting the token service. Certificate validity is checked on
credential exchange; already issued cached access tokens retain their own expiration.

```go
cfg := &auth.Config{
    Method: auth.Certificate,
    TenantID: tenantID,
    ClientID: applicationID,
    PFX: pfxBytes,
    PFXPassword: passwordBytes,
}
sdk, err := msteams.NewClient(cfg)
// Handle err, use implemented service operations, and close sdk when finished.
```

The session owns its application provider and clears its copied credential/cache references on
close. It does not revoke tokens or close an injected transport. The default refresh window is
two minutes. Token-service errors match `auth.ErrTokenService`; library response bodies and
credential diagnostics are not exposed, including through unwrapping. Certificate errors match
`ErrInvalidCertificate`, `ErrCertificateExpired` or `ErrCertificateNotYetValid` in package auth.
The standalone [provider example](../../msteams/auth/application_example_test.go) demonstrates
ownership without authenticating. See [AUTH-02 evidence and prerequisites](../implementation/auth/client-credentials-evidence.md)
for app registration/consent, role categories, tested contracts and remaining live qualification.

## Reusing an Azure Identity credential

`auth.NewAzureTokenProvider` accepts the `azcore.TokenCredential` interface used by
Microsoft Graph's Go authentication provider. It enables reuse of an explicitly chosen
Azure Identity credential through `msteams.WithTokenProvider(binding, provider)`:

```go
binding := auth.Binding{
    TenantID: tenantID,
    Cloud: "commercial",
    IdentityID: accountIdentity,
    Mode: auth.Delegated,
}
provider, err := auth.NewAzureTokenProvider(binding, credential)
if err != nil {
    return err
}
sdk, err := msteams.NewClient(nil, msteams.WithTokenProvider(binding, provider))
if err != nil {
    return err
}
defer sdk.Close()
```

The caller configures `credential` for the stated tenant, identity, cloud and mode;
application credentials use `auth.Application`. The adapter requests explicit Graph or
Teams `/.default` scopes and forwards the tenant and context to `GetToken`. It leaves
caching and credential lifetime with the caller. Closing the SDK does not close or clear
the borrowed credential. The SDK does not add an implicit default credential chain or
change the credential's interactive behavior. Configure that behavior explicitly when
constructing Azure Identity browser or device credentials.


## Built-in delegated sign-in

The `Browser`, `DeviceCode` and `Password` methods now create client-owned MSAL Go
providers. Use a tenant GUID or DNS name and your public application's client ID.
Configure delegated permissions and consent for both Graph and Teams resources.

| Method | Explicit configuration |
|---|---|
| `auth.Browser` | `OpenBrowser(ctx, authorizationURL)` callback; optional `LoginHint`; register `http://localhost` as the application's redirect URI |
| `auth.DeviceCode` | `DevicePrompt(ctx, message)` callback presenting `VerificationURL` and `UserCode` to the signing-in user |
| `auth.Password` | `Username` and `Password`; Entra must permit this flow for that account and policy |

Callbacks must respect their context and must not log callback URLs, codes or credentials.
`LoginTimeout` bounds acquisition (15 minutes by default); caller cancellation and SDK
close also cancel pending sign-in. Construction never invokes callbacks or authenticates.
Browser reception validates state and S256 PKCE and closes active local connections on
cancellation. Device polling honors interval, slow-down and code expiry. The SDK pins the
first selected account across Graph and Teams and attempts silent refresh for later tokens.
A failed password grant never falls back to browser/device or enters a retry loop.

Use `errors.As(err, &delegatedError)` with `*auth.DelegatedError` to inspect the recognized
OAuth code and numeric Entra diagnostic codes. `errors.Is` distinguishes
`auth.ErrAuthenticationRequired`, `auth.ErrAuthenticationDenied`, `auth.ErrDeviceCodeExpired`
and context cancellation/deadline errors. Numeric Entra codes help investigate MFA and
conditional-access denials; they do not trigger automatic authentication changes. Raw
server descriptions and callback errors are excluded from formatted errors.

For a federated password flow, `FederationEndpoints` must list the exact trusted HTTPS
MEX and password endpoints. A discovered MEX document cannot grant trust to another URL.
The library's WS-Trust 1.3 exchange is covered by synthetic tests; MSAL Go does not support
WS-Trust 2005. Browser/device methods and a caller-configured Azure credential remain
explicit alternatives. Windows WAM is not required by these portable browser/device flows.

The provider owns its account/cache and copied configuration. SDK close releases these
references; it does not revoke server tokens. [AUTH-03 evidence](../implementation/auth/delegated-evidence.md)
distinguishes offline verification from pending tenant qualification.
