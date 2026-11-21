# HTTP-02 bounded request execution

Status: implemented and qualified against synthetic offline transports. No tenant
requests were sent and no resource command coverage is promoted by this foundation.

## Evidence and policy

[Microsoft Graph throttling guidance](https://learn.microsoft.com/en-us/graph/throttling)
requires respecting `Retry-After` and recommends exponential backoff when it is absent.
[Graph error guidance](https://learn.microsoft.com/en-us/graph/errors) distinguishes
invalid authentication from authorization failures and describes delayed retries for
service unavailability. These sources inform transport policy; they do not establish
replay safety for an individual Teams command or prove undocumented service routes.

`client.Request.Replay` defaults to `ReplayNever`, even for GET. Resource implementations
must set `ReplaySafe` based on their operation contract, including ambiguous completion.
This permits evidenced read-only POST queries without duplicating unsafe mutations.
HTTP 408, 429, 502, 503 and 504, and ambiguous transport failures can then be retried.
Other statuses (including generic 500, conflicts and permission failures) return directly;
resource-specific recovery belongs to its evidenced contract. Result decoding failures
and response-size violations are terminal.

`msteams.WithRequestPolicy(client.RequestPolicy{...})` copies these controls into the
shared session. Zero fields select defaults: three total attempts, two-minute logical
operation timeout, 500 ms initial backoff, 30-second backoff cap and 16 concurrent API
attempts across Graph and Teams. One maximum attempt disables retries. The attempt limit
includes any authentication replay. Timeout covers token acquisition, queueing, requests,
refresh and waits; an earlier caller deadline wins. The existing HTTP timeout remains a
per-attempt limit. An independently timed-out attempt can retry while the operation lives.
Applications using interactive authentication should configure a total request timeout
that accommodates user interaction, or authenticate through their provider beforehand.

`Retry-After` supports nonnegative integer seconds and HTTP dates. Malformed values fall
back to capped exponential backoff with equal jitter. Past dates mean no extra delay.
Overflowing positive seconds saturate and exhaust the budget. A valid service delay is
never shortened by the backoff cap: if it cannot fit, return `ErrRetryBudget` together
with the last normalized service error and buffered response. Attempt exhaustion likewise
preserves the service/transport error. Actual deadline expiry preserves context errors.
Clock, wait and jitter functions are internal test seams, not production endpoint options.

Concurrency permits cover only API request/response buffering, not authentication or
sleep. Waiters observe caller and session cancellation. Each attempt releases its permit
before sleeping or refreshing. Expired queued tokens are reacquired before sending.
JSON is serialized once and reused byte-for-byte; per-attempt headers/Resty requests are
independent. Streaming `io.Reader` bodies are rejected before authentication. Uploads
remain a separate future contract and cannot accidentally enter this retry loop. Resty
retries stay disabled. HTTP request rewinding is disabled at the final transport boundary;
Go's connection recovery for bodyless replayable HTTP requests remains an underlying
transport behavior. Injected transports must not add their own mutation retry policy.

## Rejected-token replacement

An operation must explicitly set both `ReplaySafe` and `RefreshOnUnauthorized`. Only a
401 can then ask the provider to replace its token, at most once within the same attempt
and time budgets. This is not a claims-challenge processor and never selects a new login
method. 403 and other failures do not trigger refresh. The optional
`auth.RejectionRefresher` capability promises noninteractive replacement under the same
binding. Supplied tokens and providers without this capability return the original 401.
Returning the same token also ends the operation instead of replaying it.

`RefreshProvider` conditionally invalidates only the rejected cached value. Concurrent
rejections coalesce through its existing worker, and a late rejection cannot evict a
newer token or another audience's cache. `ApplicationProvider` supports this through a
fresh MSAL credential exchange. The bound-provider wrapper preserves validation and
error redaction for caller implementations. Delegated MSAL and borrowed Azure Identity
providers do not currently expose forced replacement through this capability; their
normal expiry/silent-refresh behavior is unchanged. A 401 never forces an interactive
prompt as a side effect of request retries.

## Verification

- `TestUnitRetryAfterParsingAndBoundedJitter`: seconds, dates, malformed/overflow values,
  bounded exponential growth, jitter and service-delay precedence.
- `TestUnitRetrySafeOperationReplaysFrozenJSONAndPreservesMetadata`: POST query retry,
  serialized body stability, conditional headers, final decode and token-free metadata.
- `TestUnitRetryRequiresOperationSafetyAndNeverReplaysStreams`: GET/POST/PUT/DELETE
  ambiguous failures remain single attempts by default; streams fail before auth.
- `TestUnitRetryAttemptsAndTotalDelayBudgets`, `TestUnitRetryStatusClassification`: bounded
  429/transient handling with preserved errors and terminal status classification.
- `TestUnitRequestCancellationReleasesQueuedInflightAndBackoff`,
  `TestUnitConcurrencySharedAcrossBackendsAndReleasedOnClose`: cancellation, shared limits
  and permit release, including session close.
- `TestUnitRequestBudgetIncludesAuthenticationAndQueue`,
  `TestUnitPerAttemptTimeoutRetriesWithinOperationDeadline`: total versus attempt deadlines.
- `TestUnitExpiredQueuedTokenIsReacquiredBeforeSending`: queue delay cannot send an expired
  cached token.
- `TestUnitUnauthorizedRefreshIsOptInOnceAndWithinAttemptBudget`: success, repeated 401,
  unsafe/disabled/unsupported refresh and attempt exhaustion.
- `TestUnitRejectedTokenRefreshCoalescesAndKeepsOtherAudiences`: concurrent/late rejection,
  binding isolation, unrelated audience retention and closed-provider behavior.

Live throttling and tenant conditional-access behavior remain unqualified. Resource
services must document their own replay decisions before opting in.
