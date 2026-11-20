# HTTP-05 binary, multipart and batch contracts

The shared transfer boundary preserves bytes, stream ownership and partial results.
Backend services still own evidenced routes, content types, size limits and permissions.
No tenant upload, application publication, policy assignment or batch mutation was run.

## Pinned evidence

Offline IL inspection of MicrosoftTeams 8.1.0 (accepted package identity in baseline
provenance) found `SetTeamPicture.ProcessRecord` constructing ByteArrayContent at
IL 216, selecting `image/jpeg` at 235 and PUT to the Graph group photo value route.
`NewTeamsApp.ProcessRecord` selects `application/zip` at 341;
`SetTeamsApp.ProcessRecord` selects it at 306. These findings establish that binary
payloads cannot be JSON-marshaled byte slices (which Go would encode as base64).
TeamsCmdlets assembly SHA256:
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`.

The pinned ConfigAPI `GrantActionStatus` model has `id`, `state`, `result`; batch
`userState` and counts are already evidenced in [HTTP-04](operation-evidence.md).
`DecodeBatchPolicyResult` now exposes those per-user outcomes alongside aggregate
failure, preserving unknown failure text and rejecting duplicate IDs as inconsistent.
It does not turn every Teams batch into Graph's JSON batch envelope.
ConfigAPI private assembly SHA256:
`8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`.

A loaded-assembly literal scan found the JPEG/ZIP content types above, but no
multipart content-type literal. UploadMultipart is explicit form-data infrastructure;
no resource is silently assigned a guessed multipart route or field name. Resource
tickets must establish their own wire encoding. A PowerShell Byte[] parameter alone
is not proof that its backend takes a raw binary body.

## Current primary documentation

Reviewed 2026-11-20, separate from the pinned package:

- [Graph photo update](https://learn.microsoft.com/en-us/graph/api/profilephoto-update?view=graph-rest-1.0)
  defines binary photo updates. The resource-specific photo limit is not used as a
  universal SDK transfer limit.
- [Teams app publication](https://learn.microsoft.com/en-us/graph/api/teamsapp-publish?view=graph-rest-1.0)
  takes an application/zip manifest payload. This helper does not publish an app or
  assume the caller has permission to do so.
- [Graph JSON batching](https://learn.microsoft.com/en-us/graph/json-batching)
  permits up to 20 operations, unique case-insensitive IDs, relative URLs, explicit
  Content-Type with bodies, and optional dependsOn ordering. Responses can arrive
  out of order; outer HTTP 200 does not imply each item succeeded.
- [Graph throttling](https://learn.microsoft.com/en-us/graph/throttling)
  requires handling item-level throttling independently. The SDK exposes item status,
  Retry-After headers and bodies; it never replays a whole partially successful batch.

## Transfer ownership and bounds

`client.TransferClient` exposes Upload and Download; Scope also provides
UploadMultipart. All use the selected backend's exact destination/audience guard,
local session cancellation, total request timeout and shared concurrency permits.
Streaming uses net/http directly and returns Resty response metadata. The pinned
Resty v2.17.2 `createHTTPRequest`/`getBodyCopy` middleware buffers io.Reader bodies
for GetBody; using it for this path would defeat bounded streaming. Existing JSON
operations continue through Resty unchanged.

Upload owns its ReadCloser immediately, including validation failures, and closes it
exactly once. Its Close must unblock concurrent Read. Cancellation invokes Close;
final source errors are synchronized before return. There is no producer goroutine.
A positive byte cap is required; the reader probes one byte beyond the cap without
sending it. Oversize/source failures can leave partial remote state and are not
transactional. Content-Type is explicit. Optional successful JSON output is decoded
from a bounded response. Error bodies keep existing APIError classification.

UploadMultipart owns every supplied part, including parts following an invalid one.
Sources must be independent. Repeated comparable source instances are rejected and
closed once; aliased wrappers are caller misuse. Framing is generated with mime/multipart
and composed with io.MultiReader; file contents are not buffered. The limit includes
framing. Field names, filenames and content types reject control characters; escaped
quoted filenames round-trip. At most 100 parts are accepted.

Download borrows its Writer and never closes it. It writes at most the supplied
positive cap, preserves the written byte count on failures and closes response bodies.
Non-2xx content never reaches the destination; it is buffered for error reporting.
Truncated bodies, destination errors and close errors are explicit, redacted failures.
A Writer must return promptly: arbitrary blocked Write implementations cannot be
canceled by Go context. Partially written output is not rolled back.

Both stream APIs reject ReplaySafe and RefreshOnUnauthorized. No SDK retry is made
for a consumed source, partial destination or server error. Redirects are returned
as failures. Resource callers decide any recovery using their own idempotency evidence.
No size cap is inferred from Content-Length; actual bytes are counted.

## Batch behavior

`graphapi.ExecuteBatch` validates every nested destination/method/header before
credentials, disallows foreign authorities, path traversal, version switching and
nested batches, checks unique IDs and acyclic existing dependencies, then performs
one POST to the selected v1.0/beta batch endpoint. Bodies are explicit json.RawMessage;
base64url binary bodies can be supplied as JSON strings according to the Graph contract.

Mixed failures return valid item results plus ErrBatchItems. Missing, duplicate,
unknown or malformed item outcomes return known valid results plus ErrBatchResponse.
The buffered raw outer response remains available for explicit inspection. Item
statuses and headers are never interpreted as permission to retry unrelated mutations.
`teamsadminapi.DecodeBatchPolicyResult` instead preserves the pinned ConfigAPI job
summary and userState details; absent details remain nil rather than proving no users.

## Verification

- TestBinaryTransferPreservesBytesAndOwnership: exact binary bytes on both backends,
  content types, JSON result and bearer-free returned metadata.
- TestUploadReachesHTTPServerBeforeSourceCompletes: real local HTTP server receives
  a chunk while the caller still holds the unfinished source.
- TestUploadFailuresCloseSourceAndNeverReplay: oversize/read/429/foreign/unsafe replay.
- TestUploadCancellationClosesBlockedSource: cancellation releases a blocked reader.
- TestDownloadLimitsTruncationAndPartialWrites: byte caps, unexpected EOF, partial
  writer failure, error body isolation and close errors.
- TestMultipartFramingBytesAndFailureCleanup and
  TestMultipartRejectsRepeatedSourceWithoutDoubleClose: boundaries, bytes, filenames,
  fields and all-part cleanup.
- Graph batch tests cover mixed 201/429/424, pre-auth validation, dependency cycles,
  malformed/missing results, preserved successes and outer-throttle no-replay.
- TestTeamsBatchDetailsSurviveAggregateFailure: success plus unknown item failure
  remains inspectable; duplicate user results cannot establish success.

These foundation fixtures do not promote any resource command to implemented or
live-qualified status. Individual resource wrappers and their live qualification
remain assigned to their resource tickets.
