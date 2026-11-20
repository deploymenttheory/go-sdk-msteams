# Decision 0003 Backend contracts and compatibility

Status: accepted architecture. Implements backend and compatibility boundaries for [ARCH-01](https://github.com/deploymenttheory/go-sdk-msteams/issues/21).

## Decision

The shared transport owns context, credentials, timeout/retry budgets, concurrency limits, HTTP execution, redaction and observation hooks. Backend adapters own route discovery, request/response envelopes, error conversion, paging and asynchronous status semantics. Resource services own typed operation inputs/outputs and baseline-specific validation.

Resource operations select a backend identity; they do not inject arbitrary bearer headers or override auth audiences. Cloud profiles explicitly bind authority, audience and allowed service/discovery destinations. Absolute continuation/operation URLs and redirects must remain in the approved trust boundary. An override for a test server uses an explicit test transport/profile; generic URL options cannot silently send production credentials elsewhere.

Errors retain backend family, HTTP status, service code and request/correlation identifiers, with wrapping that preserves `errors.Is`/`errors.As` and context cancellation. Raw Resty response metadata remains available alongside typed errors. Decode errors, transport errors, validation errors and remote authorization failures are distinct; bounded sanitized diagnostics replace unredacted response bodies in logs.

Retry policy is operation-aware. Respect `Retry-After` and total budgets; never infer replay safety solely from a status code or HTTP verb. An ambiguous failed mutation, partially successful batch or consumed upload stream cannot be automatically replayed without an evidenced safe mechanism. Token refresh cannot create an unbounded 401 loop. Cancellation releases waiting concurrency permits and stops polling/backoff.

List adapters preserve backend continuation semantics and detect repeated cursors. Aggregate traversal must expose failure rather than return truncated success. Long-running operations distinguish acceptance from terminal success/failure and retain operation identity when a local wait is cancelled. Eventual-consistency handling is restricted to evidenced transient cases. Streaming ownership and cleanup are documented for binary/multipart operations; batch results preserve per-item failures.

## Evidence and versioning

The accepted upstream baseline is MicrosoftTeams 8.1.0, identified by package hash. The [BASE-01 workflow](../../tools/baseline/README.md) produces reviewable candidates and cannot replace existing snapshot content. Package metadata, script/configuration evidence, Microsoft documentation and sanitized observations determine contracts. Lexical manifest candidates and unauthenticated exports do not by themselves establish endpoint or public-command status.

Go SDK semantic versions are independent of PowerShell releases. A method is not renamed `GetV810` when a module release changes. Actual API-version variants are exposed only when the wire contract requires the distinction. An upstream parameter addition may be an additive Go field; a removal or changed semantic requires compatibility review rather than automatically deleting an exported symbol. Candidate 8.2.0 discovery produces a report/PR; adopting it is an explicit reviewed baseline change.

The first supported SDK release requires complete implemented and contract-tested management parity. Local builders and session equivalents count only with explicit mappings. Internal telemetry is excluded with recorded rationale. Missing lab licences/resources do not remove implementation scope; unknown wire behaviour remains a research blocker. Commercial qualification does not imply sovereign-cloud qualification.

## Verification requirements

HTTP tests use controlled servers, deterministic time/randomness where relevant, and context/race checks. They must cover both backend families, malformed responses, error identifiers, trusted routing, throttling budgets, replay safety, continuation loops, async terminal failures and partial batches. Resource acceptance suites compare equivalent SDK/PowerShell scenarios and verify cleanup separately. The coverage ledger records metadata-only, contract-tested and live-tested states without conflating them.
