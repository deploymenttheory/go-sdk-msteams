# Decision 0003 Backend routing, retries and testing

Status: accepted architecture. Implements backend and compatibility boundaries for [ARCH-01](https://github.com/deploymenttheory/go-sdk-msteams/issues/21).

## Decision

The shared transport owns context, credentials, timeout/retry budgets, concurrency limits, HTTP execution, redaction and observation hooks. Backend adapters own route discovery, request/response envelopes, error conversion, paging and asynchronous status semantics. Resource services own typed operation inputs/outputs and operation-specific validation.

Resource operations select a backend identity; they do not inject arbitrary bearer headers or override auth audiences. Cloud profiles explicitly bind authority, audience and allowed service/discovery destinations. Absolute continuation/operation URLs and redirects must remain in the approved trust boundary. An override for a test server uses an explicit test transport/profile; generic URL options cannot silently send production credentials elsewhere.

Errors retain backend family, HTTP status, service code and request/correlation identifiers, with wrapping that preserves `errors.Is`/`errors.As` and context cancellation. Raw Resty response metadata remains available alongside typed errors. Decode errors, transport errors, validation errors and remote authorization failures are distinct; bounded sanitized diagnostics replace unredacted response bodies in logs.

Retry policy is operation-aware. Respect `Retry-After` and total budgets; never infer replay safety solely from a status code or HTTP verb. An ambiguous failed mutation, partially successful batch or consumed upload stream cannot be automatically replayed without an evidenced safe mechanism. Token refresh cannot create an unbounded 401 loop. Cancellation releases waiting concurrency permits and stops polling/backoff.

List adapters preserve backend continuation semantics and detect repeated cursors. Aggregate traversal must expose failure rather than return truncated success. Long-running operations distinguish acceptance from terminal success/failure and retain operation identity when a local wait is cancelled. Eventual-consistency handling is restricted to evidenced transient cases. Streaming ownership and cleanup are documented for binary/multipart operations; batch results preserve per-item failures.

## Versioning and supported behaviour

Go SDK semantic versions are independent of the PowerShell version used by an
optional tenant comparison. API methods and fields follow real SDK operations;
API-version variants are exposed only where the service requires them. Review
public API compatibility before removing or changing exported symbols.

Use existing SDK code, service documentation and actual response observations
when changing an endpoint or payload. Standalone binary/CLR research, command
inventories and contract ledgers are not development or release prerequisites.
Keep supported functionality and missing permissions, licences or tenant fixtures
explicit in API documentation and the GitHub backlog. Commercial tenant results
do not establish sovereign-cloud support.

## Verification requirements

Use ordinary unit tests with controlled HTTP responses and service testdata JSON.
Cover malformed responses, error identifiers, trusted routing, throttling/retry
budgets, cancellation, continuation loops, async failures and partial batches as
appropriate. Resource acceptance tests exercise real tenant operations and verify
read-back and cleanup independently. Preserve useful existing PowerShell
comparisons inside selected scenarios without a separate qualification framework.

Release checks assess implemented SDK functionality, unit/quality checks and
available real-tenant results. A skipped or unavailable scenario is not a passing
operation. Voice administration is deferred until the required lab licensing is
available.
