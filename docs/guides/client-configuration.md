# Client construction and ownership

`NewClient(authConfig, options...)` consumes explicit configuration. It never reads
`MSTEAMS_*` variables. `NewClientFromEnv(options...)` reads the documented variables,
applies options in order, and then parses and validates the resulting configuration.
`WithAuthConfig` replaces the complete authentication configuration; it also discards
an environment token payload. `WithHTTPTimeout` can replace an invalid environment timeout.

Construction validates authentication selection, required credential presence, incompatible
credential combinations and HTTP settings. It does not authenticate, call assertion
callbacks, read assertion files, parse certificate/key material, validate token signatures
or check token expiry. Those operation-scoped checks belong to AUTH provider tickets.
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
There is no implicit credential chain or fallback to browser login. Availability and
operation-specific permissions remain provider/contract qualification concerns.

Every method requires a tenant ID. Browser/device-code require a client ID; password
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
transports. Route discovery, cloud profiles, audience routing, retries and authenticated
HTTP execution belong to HTTP/AUTH tickets; construction invents no endpoint routes.

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
