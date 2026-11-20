# HTTP-04 operations and eventual consistency

Operation handles separate an accepted mutation from observed completion. The shared
waiter performs only replay-safe GET status reads; it never repeats the initiating
mutation. Resource services supply evidenced status routes and envelope decoders.

## Pinned and current evidence

Offline reflection of MicrosoftTeams 8.1.0 found the following contracts. No module
connection or management command was invoked; serialization used synthetic values.

| Assembly/model | Observation |
|---|---|
| `Microsoft.Teams.PowerShell.TeamsCmdlets.Model.TeamsAsyncOperation` | JSON attributes map operation ID to `id`, state to `status`, and target to `targetResourceLocation` |
| `AsyncOperationStatus` | Pinned enum names: Invalid, NotStarted, InProgress, Succeeded, Failed |
| ConfigAPI `BatchJobStatus` and `SimpleBatchJobStatus` | Serialization emits `operationId`, `overallStatus`, `errorCount`, `notStartedCount`, `inProgressCount`, `completedCount` |
| ConfigAPI `GrantActionStatus` | Per-user output uses `id`, `state`, `result`; batch details use `userState` |
| ConfigAPI `GetOperationResponse` | Separate Shifts envelope uses `Id`, `status`, `details`; it is not treated as the batch envelope |

Assembly SHA256:

- TeamsCmdlets: `3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`
- ConfigAPI private: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

The accepted package identity remains recorded in the baseline provenance. These
observations establish field mappings, not authorization or tenant execution.

Current Microsoft documentation, reviewed 2026-11-20, supplies additional semantics:

- [Graph Teams async operations](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-1.0)
  return 202/Location and require more than 30 seconds between status checks.
  The helper enforces 31 seconds, including transport retries and token-replacement
  retries. The target location remains opaque metadata.
- [Create team responses](https://learn.microsoft.com/en-us/graph/api/team-post?view=graph-rest-1.0)
  include unversioned `/teams(...)/operations(...)` locations. The Graph adapter
  anchors these to the selected Graph v1.0 origin; absolute locations still pass the
  shared destination guard. Scheme-relative and foreign-host locations are rejected
  before authentication. `Content-Location` is not substituted for the operation.
- [Batch policy status](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-csbatchpolicyassignmentoperation?view=teams-ps)
  distinguishes NotStarted/InProgress/Complete and per-user Completed/Success.
  Complete with errors is a failed aggregate. Missing error counts or inconsistent
  incomplete members cannot establish success.

## Wait contract

`client.NewOperation` snapshots an explicitly replay-safe GET route, query and
headers. It accepts a supplied operation ID or learns a stable ID from the first
status body. A changed ID fails rather than switching jobs. It records the initial
response as Accepted even when its status is 200. Graph's constructor specifically
requires 202 and Location. A 202 polling response cannot establish success.

`Poll` respects the backend interval and Retry-After. `Wait` adds a five-minute
default total timeout, five-second interval and 1,000-poll bound. Caller deadlines,
session closure and cancellation interrupt waits and HTTP requests. Internal request
attempts retain HTTP-02's separate attempt/concurrency bounds. A valid server delay
is never shortened to fit the local deadline. Per-request `MinimumRetryDelay` can
only increase transport retry delays; it also covers opted-in 401 replacement.

Each return includes the last `OperationSnapshot`: handle, normalized state and
buffered Resty response. Terminal failures additionally preserve normalized backend
diagnostics. These objects redact ordinary formatting; raw response bodies are for
explicit inspection only. Local cancellation/timeout does not mean remote cancellation
and does not discard the handle. A new wait can resume the same object with a fresh
context. Remote terminal results are sticky and make no additional requests.

Call Poll/Wait serially per handle; separate handles are independent. The waiter
stores no caller context and creates no background polling worker. It does not
cancel remote jobs, follow target-resource URLs or auto-delete partial resources.

## Eventual-consistency boundary

By default every HTTP error exits the wait. A resource contract may opt into exact
backend code/status pairs via `PropagationRules`, together with a positive bounded
`PropagationWindow`. Only 400/404/409 are eligible; 401/403 are never propagation
success or retry hints. A matching code outside the window fails. Synthetic
`FixturePending` codes in tests are fixtures, not production defaults.

A service can use the same GET/decoder boundary to wait for an evidenced resource
readiness condition after mutation. It must define the readiness predicate and any
propagation errors from resource evidence. This implementation enables no guessed
tenant propagation exceptions. Unknown states fail explicitly rather than spinning
or silently claiming completion.

## Verification

| Contract | Fixture |
|---|---|
| Accepted/running/succeeded/failed, sticky terminal result and diagnostics | `TestOperationWaitSequencesAndTerminalDiagnostics` |
| Recoverable local cancellation, poll limit, distinct remote cancellation | `TestOperationCancellationAndPollingLimitAreResumable` |
| Delta/date Retry-After, deadline, stalled job | `TestOperationRetryAfterAndStalledWait` |
| Unknown future state, changed ID, 202 with success body | `TestOperationRejectsUnknownStateIdentityChangeAnd202Success` |
| Exact propagation code/status/window and default rejection | `TestOperationPropagationRequiresExactBoundedEvidence` |
| Missing/unsafe locations and Graph interval before authentication | `TestGraphOperationLocationValidationAndMinimumDelay` |
| Batch completion with item failures or inconsistent metadata | `TestBatchCompleteDoesNotHidePartialFailure` |
| Session closure interrupts polling sleep | `TestOperationCloseCancelsSleepWithoutLosingHandle` |
| Graph polling floor inside transient and rejected-token HTTP retries | `TestOperationMinimumAppliesInsideTransportRetries`, `TestOperationMinimumAppliesToRejectedTokenReplacement` |

Other operation families require their own evidenced decoders and routes in resource
tickets; generic cancellation support does not claim Graph Teams exposes a cancelled
state. No live asynchronous mutation has been run, and this foundation does not
advance individual resource commands to implemented or live-qualified status.
