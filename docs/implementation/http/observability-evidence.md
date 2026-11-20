# HTTP-06 observability and transport configuration investigation

Status: in progress for [HTTP-06](https://github.com/deploymenttheory/go-sdk-msteams/issues/58).
This document records the implementation boundary and remaining work. It is not a
claim that request logging/tracing or the additional transport options are delivered.

## Reference and existing implementation

The agreed [Jamf reference at 25714da5](https://github.com/deploymenttheory/go-sdk-jamfpro-v2/tree/25714da557956da9e317f725ca3d22e8818ec77e)
exposes WithLogger, WithDebug, WithUserAgent, WithGlobalHeader(s), WithProxy,
WithTLSClientConfig, WithTransport and concurrency/retry options in
[jamfpro/with_options.go](https://github.com/deploymenttheory/go-sdk-jamfpro-v2/blob/25714da557956da9e317f725ca3d22e8818ec77e/jamfpro/with_options.go).
These signatures were inspected through the GitHub API on 2026-11-20. Its WithDebug
explicitly enables Resty request/response debug logging. That behavior requires a
redaction boundary before adoption in this SDK; auth tokens and tenant bodies cannot
be passed to Resty's ordinary debug logger.

The current Teams root already accepts a borrowed HTTP client, timeout, request
policy, Zap logger and optional tracer provider. Logger/tracer accessors currently
return configured dependencies; they are not yet request instrumentation. Backend
requests use a quiet Resty logger. HTTP-05 stream transfers use a separate net/http
execution path with shared auth, route, deadline and concurrency guards. Both paths
must receive equivalent sanitized instrumentation.

SessionOption currently configures only RequestPolicy. Extending constructor wiring
must preserve later-option precedence and lazy construction, with no global logger,
tracer provider, HTTP transport or credential chain mutation.

## Implementation sequence

1. Add immutable per-session diagnostics/default-header configuration and root options.
   Keep injected logger/tracer lifecycles borrowed. Validate generic headers before
   credentials; reject Authorization, cookies, Host and other existing reserved fields.
   Scope default backend headers to management requests, not identity exchanges.
2. Add Jamf-style transport, proxy, TLS and user-agent options using copied transports
   and TLS configuration. Do not mutate a caller-owned transport. Explicitly define
   composition with an injected opaque RoundTripper and reject impossible combinations.
   Proxy addresses/credentials and TLS key material must never enter error text/logs.
3. Instrument logical requests and attempts using a small field allowlist: backend,
   validated method, attempt/status, elapsed time and safe error category. Never emit
   raw URLs/queries, headers, bodies, tenant/identity IDs, service messages or unwrapped
   provider errors. Debug mode remains sanitized. Trace opt-in must not enable global
   telemetry or attach baggage containing credentials to service calls.
4. Apply the same controls to JSON requests, streaming and batch outer requests;
   preserve auth/routing, retry and cancellation behavior. Polling remains explicit.
5. Capture Zap logs and OTel spans with secret fixtures across success, HTTP failures,
   authentication failures, retries and cancellation. Test custom transports, TLS
   trust, proxies, protected headers, copy isolation and caller lifecycle ownership.
   Then run all platform checks and attach exact evidence before closing HTTP-06.

## Qualification limits

No logging/tracing changes are implemented by this investigation. No proxy/TLS live
infrastructure or tenant writes have been exercised. The catalog package correction
is independently merged and is not a prerequisite for request instrumentation.
