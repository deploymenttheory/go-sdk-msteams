# HTTP-06 observability and transport configuration investigation

Status: implementation and fixture evidence for
[HTTP-06](https://github.com/deploymenttheory/go-sdk-msteams/issues/58).

## Reference and existing implementation

The agreed [Jamf reference at 25714da5](https://github.com/deploymenttheory/go-sdk-jamfpro-v2/tree/25714da557956da9e317f725ca3d22e8818ec77e)
exposes WithLogger, WithDebug, WithUserAgent, WithGlobalHeader(s), WithProxy,
WithTLSClientConfig, WithTransport and concurrency/retry options in
[jamfpro/with_options.go](https://github.com/deploymenttheory/go-sdk-jamfpro-v2/blob/25714da557956da9e317f725ca3d22e8818ec77e/jamfpro/with_options.go).
These signatures were inspected through the GitHub API on 2026-11-20. Its WithDebug
explicitly enables Resty request/response debug logging. That behavior requires a
redaction boundary before adoption in this SDK; auth tokens and tenant bodies cannot
be passed to Resty's ordinary debug logger.

## Implemented configuration

The root retains WithHTTPClient, WithHTTPTimeout, WithRequestPolicy, WithLogger and
WithTracerProvider. It adds WithTransport, WithProxy, WithTLSClientConfig,
WithGlobalHeader(s), WithUserAgent, WithDebug and WithMaxConcurrentRequests. Later
transport selections replace earlier selections; a later full request policy replaces
an earlier concurrency-only setting. Constructors validate and copy configuration
without token exchange, browser launch or network discovery.

An opaque RoundTripper remains borrowed. Proxy/TLS settings require *http.Transport
and operate on an owned clone; incompatible opaque combinations fail locally. Explicit
HTTP/HTTPS/SOCKS5 proxy URLs are parsed without exposing their values in errors. No
ambient proxy is discovered by the default SDK transport. Both identity and management
requests use the selected network transport, while generic management headers and
user agents apply only at the backend boundary.

TLS configuration is copied, including CA pools, certificate DER/metadata, parsed
leaf certificates and ECH configuration/key byte slices. TLS callbacks, private-key
signers and session caches remain borrowed and must be safe for concurrent use.
WithTLSClientConfig requires certificate verification and TLS 1.2 or newer (TLS 1.3
for ECH); key logging is rejected. Callers configure custom trust with RootCAs.
The SDK does not add an insecure-skip-verification convenience option.

SessionOption now configures private session settings; use WithRequestPolicy,
WithDiagnostics and WithDefaultHeaders rather than custom functions mutating a
RequestPolicy. This is a pre-release option-type change. Authentication and resource
interfaces are unchanged.

## Diagnostic boundary

The selected Zap logger receives one Info completion event per logical request.
WithDebug additionally enables Debug attempt events when the caller logger allows
Debug level. The SDK never enables Resty's request/response debug logger. Both JSON
and stream paths emit the same bounded fields: backend, recognized HTTP method,
status, attempt count, elapsed time and a fixed outcome category. Invalid methods
become unknown. No error object is formatted or passed to the logger.

WithTracerProvider enables one client span per logical request. The context reaches
auth acquisition, queueing and HTTP attempts; request attempts add sanitized events.
Spans end on success, API errors, local validation errors, auth errors and cancellation.
Status descriptions use the same fixed outcome category, never raw provider/service
messages. Parent context is honored, but no global provider/propagator is initialized,
no baggage or trace headers are automatically injected, and no caller provider is
shut down. The package uses the pinned OTel 1.47.0 interfaces, inspected in the local
module source. No telemetry dependency or exporter is added.

URLs, query strings, bodies, request/response headers, tokens, assertions, passwords,
certificate keys, tenant IDs, service codes and diagnostic IDs are excluded from
SDK-generated events. Explicit Resty response metadata and returned errors retain
existing inspection/error-matching behavior. A caller-supplied logger may already
have its own fields, and custom transports/providers/exporters may produce their own
output; that caller-owned behavior is outside the SDK-generated diagnostic contract.

Batch logs describe the outer HTTP request. An outer 200 is not a claim that every
item succeeded; callers continue to inspect the typed per-item results from HTTP-05.
Polling produces one observed request for each actual poll; it starts no telemetry
background worker. There is no global logging/metrics/profiling server or RUM tracking.

## Header and lifecycle behavior

Default headers are copied and validated before token acquisition. Request headers
replace defaults case-insensitively; only after merging does an absent Accept become
application/json. That ordering applies to uploads, downloads, pagers and operation
polling, which previously validated headers before execution. Credential/cookie/host/
framing reserved headers remain rejected. Caller mutation cannot redirect requests
or mutate the saved defaults through returned request metadata.

The client closes only its owned transport or transport clone. Borrowed transports,
loggers, tracing providers, TLS signers/callbacks and caches keep caller ownership.
Existing shared concurrency, total deadlines, queue cancellation, token isolation,
redirect rejection and explicit replay rules still apply.

## Verification

- TestRequestDiagnosticsExcludeSecretsAcrossOutcomes captures rendered Zap fields
  and span names/attributes/events/statuses for JSON success, forbidden responses,
  provider/transport failure, retry, upload/download, cancellation and invalid methods.
  It asserts no fixture secrets/tenant identifiers enter diagnostics and every created
  span ends exactly once. The actual trace context reaches auth and HTTP.
- TestDiagnosticsRemainOptInAndBorrowed checks Debug is not implicitly enabled and
  the caller logger remains usable after Close.
- TestDefaultHeadersPreservePrecedenceAcrossRequestKinds covers JSON, upload,
  download, pager and operation paths, lower-case explicit Accept and caller mutation.
- TestManagementHeadersNeverReachIdentityExchange checks native MSAL metadata/token
  requests do not receive management defaults while the Graph request does.
- TestTLSAndProxyOptionsReachVerifiedBackendWithoutMutatingTransport uses a real local
  CONNECT proxy and TLS server, with explicit trust and bearer/proxy header isolation.
- TestTLSOptionSnapshotsCertificateAndECHData verifies nested TLS copying.
- TestTransportOptionsRejectUnsafeOrIncompatibleConfiguration covers reserved headers,
  malformed proxy/TLS values, duplicates, typed-nil transports and opaque proxy conflicts.
- TestZeroSessionDiagnosticsReturnClosed protects zero-value session behavior on all
  request paths rather than allowing observation to panic before local validation.

Validation on 2026-11-20 passed: `go test -race -count=1 -shuffle=on ./...`,
`go build ./...`, `go vet ./...`, `golangci-lint run ./...`, formatter diff, coverage
ledger, backlog validation and capability matrix verification. The opt-in
`TestApplicationAccessLab` passed client-secret, certificate and invalid supplied-token
cases against the existing commercial lab identity. These are read-only authentication/
access regressions, not live proxy/TLS or telemetry qualification.

Fixture proxy/TLS servers are local only. No tenant configuration or credential
permissions are changed, and no resource command is promoted to release-ready parity
by this foundation.
