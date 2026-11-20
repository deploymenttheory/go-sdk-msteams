# HTTP-01 routing evidence

Status: HTTP-01 implemented with offline transport tests. No live tenant management
request has been qualified. This records source evidence and remaining routing scope.

The SDK entry point and all its tests live under `msteams/`. HTTP work belongs under
`msteams/client/`; no Go files belong at the repository root.

## Graph

Microsoft documents `https://graph.microsoft.com/{version}/{resource}` with explicit
`v1.0` or `beta` versions and query parameters. Resource identities must be escaped as
segments and query values encoded separately. Requests must not let an identity change
the destination host or introduce additional path segments. Source:
[Calling Microsoft Graph](https://learn.microsoft.com/en-us/graph/call-api).

Graph errors use HTTP status plus an `error` object with service code, message and optional
nested diagnostics. The transport must preserve status, codes and correlation identifiers
while keeping server messages, bodies, credential headers and URL values out of printable
errors. Non-JSON and empty failures still need a useful status-bearing error. Source:
[Graph errors](https://learn.microsoft.com/en-us/graph/errors).

Cloud endpoints and tokens are not interchangeable. The first transport profile targets
commercial service endpoints. Sovereign profiles must be explicit and evidenced, not
inferred from string replacement or a suffix wildcard. Source:
[Graph cloud deployments](https://learn.microsoft.com/en-us/graph/deployments).

## Pinned Teams module evidence

Inspected the already SHA512-verified MicrosoftTeams 8.1.0 package. Assembly:
`netcoreapp3.1/Microsoft.Teams.ConfigAPI.CmdletHostContract.dll`, SHA256
`477bb287ab3b0fddfcbdad6db0ecdd7e7df590a6fe2dec066c6b75e9ed7e69af`.

Read method-body IL using .NET reflection without invoking static initializers or
connecting to a tenant. The static constructor on
`Microsoft.Teams.ConfigAPI.CmdletHostContract.DeploymentConfiguration` contains:

| IL offset | Literal |
|---|---|
| 1112 | `https://api.interfaces.records.teams.microsoft.com` |
| 1130 | `https://msit.api.interfaces.records.teams.microsoft.com` |
| 1148 | `https://dev.api.interfaces.records.teams.microsoft.com` |
| 1258 | `https://api.interfaces.records.dod.teams.microsoft.us` |
| 1276 | `https://api.interfaces.records.gov.teams.microsoft.us` |
| 1294 | `https://api.interfaces.records.teams.microsoftonline.cn` |
| 1360 | `https://webdir.online.lync.com/Autodiscover/AutodiscoverService.svc` |

The same constructor contains additional development, legacy and sovereign destinations.
The follow-up IL inspection establishes the mapping: at offsets 1111–1122, the
constructor adds enum value 0 and the commercial URI to
`ConfigApiEnvironmentToEndpoints` (stored at offset 1348). Enum metadata identifies
`ConfigApiEnvironment.Public = 0`; the cloud mapping at 1092–1100 associates
`AzureCloudInstance.AzurePublic = 0` with it. `TryGetConfigApiEnvironmentFromEndpoint`
reads that same dictionary and compares the supplied URI with each mapped value.
This supports the exact commercial ConfigAPI authority, not a wildcard for all Teams
hosts. Development, legacy and sovereign authorities remain rejected by this profile. The `ConfigApiAutoDForestProvider` methods also contain discovery-related URL
construction, so the common ConfigAPI host does not establish every Cs cmdlet route.

## Implementation boundaries

Use the existing per-client session and AUTH-01 token boundary. Resource adapters select
backend and an evidenced route; the transport selects that backend's audience. Preserve
caller context and close bodies. Disable credential-bearing redirects. Validate every
explicit, discovered or continuation destination before obtaining/attaching a token.

Use Resty v2 response metadata and typed errors behind the constructor-injected request
interface. Unit fixtures must exercise Graph/Teams error envelopes, malformed/non-JSON
responses, empty bodies, escaped identities, wrong audiences, cancellation and rejected
destinations. Keep retry, paging and long-running-operation policies with their dependent
HTTP tickets. Resource method paths/permissions remain with their contract research.

## Teams error model evidence

The same verified package contains `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll`,
SHA256 `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`.
After importing the local module as an offline development oracle, inspected properties
on `Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models` and called each model's
`ToJsonString()` with synthetic string/numeric properties. No connection or cmdlet
management operation was invoked. Observed wire shapes:

| Generated model | JSON fields retained by transport |
|---|---|
| `ErrorResponse`, `ErrorObject` | `code` |
| `DefaultHttpErrorResponse` | `code`, `message.diagnostics.{correlationID,code,subCode}` |
| `NcsErrorResponse` | `diagnostics.{correlationID,code,subCode}` |
| `TeamMiddletierErrorResponse` | `operationId`, `errors` array of `ErrorObject` (CLR property is `Error`) |

`message`, `action`, `reason`, `debugContent` and `genevaLogsUrl` are deliberately
excluded from errors. They remain available only in the explicitly returned raw response.
Graph's error object and recursive `innerError` retain codes and request identifiers;
header identifiers are preserved separately from body identifiers when they differ.
The fixtures in `TestNormalizedErrorEnvelopes` are synthetic, derived from these models
and the Graph documentation, and do not contain tenant captures.

## Delivered contract and checks

`client.Scope` implements the constructor-injected `client.Client` interface using
`Do(ctx, Request, result) (*resty.Response, error)`. Graph and Teams scopes share their
session but select separate fixed authorities and token audiences. The commercial
Teams audience is the AUTH-02 evidenced Teams application resource, not its API host.

`Request.Path` is a sequence of unescaped segments; Graph callers include `v1.0` or
`beta` explicitly. `Query` uses `url.Values`. A continuation/discovery `URL` is mutually
exclusive with both, retains its encoded query, and passes the same destination checks.
No discovery request or continuation traversal is performed implicitly. Routing rejects
foreign hosts, explicit ports, credentials, fragments, non-HTTPS/opaque/relative URLs,
unsupported cloud profiles, dot segments and control/backslash path ambiguity before
acquiring tokens. A second guard validates the final request before attaching a bearer.
Backend headers survive except headers that can override credentials, routing, cookies,
framing or compression. The borrowed HTTP client's cookies and redirect callbacks are
not used. Caller configuration is not mutated.

The transport sends one attempt; HTTP-02 owns request budgets/retries and HTTP-03 owns
paging. JSON bodies preserve explicit zero values. Responses are buffered up to 8 MiB
and closed; successful nonempty JSON decodes into the caller's result. Empty success
bodies return metadata without inventing a resource. Raw responses contain tenant data
and are not safe for indiscriminate logging. `APIError` exposes status, codes and IDs,
but all standard pointer/value error formatting omits untrusted values. Transport and
decode errors discard potentially sensitive underlying diagnostic text. Context errors
retain `errors.Is` identity. Request metadata never contains the attached bearer.

| Test | Evidence |
|---|---|
| `TestBackendRoutingEscapingAndResponseMetadata` | Distinct audiences/hosts, segment/query encoding, headers, typed result and raw metadata |
| `TestRejectedDestinationsAndRequestsNeverAcquireCredentials` | Host/path/header/profile rejection before token acquisition |
| `TestContinuationURLPreservesOpaqueQuery` | Encoded continuation query remains intact |
| `TestRedirectsAndBorrowedCookiesAreNotForwarded` | One 307 response; no follow-up credentials/cookies or mutation of caller settings |
| `TestNormalizedErrorEnvelopes` | Graph and four Teams shapes, malformed/non-JSON/empty errors, safe formatting |
| `TestWrongAudienceHasNoFallback` | A Graph token cannot satisfy a Teams request |
| `TestConcurrentRequestsKeepHeadersAndResultsIsolated` | Concurrent calls do not share request headers or result destinations |
| `TestJSONBodiesSuccessDecodingAndEmptyResponses` | Explicit false/zero/empty values, empty success, safe decode failures |
| `TestBodiesClosedForSuccessFailureAndSizeLimit` | Success/error/oversize response ownership |
| `TestCancellationReachesBackendAndCloseCancelsRequest` | Caller/session cancellation reaches the HTTP server |
| `TestTransportGuardClosesRejectedRequestBodies` | Final destination/context/token rejection closes request bodies |
| `TestTransportErrorsDoNotExposeCredentialsOrURLs` | Foreign diagnostic text discarded; response-plus-error body closed |

These tests map shared transport behavior, not individual PowerShell management commands.
No command ledger entry is promoted to implemented, contract-tested or live-tested by
HTTP-01 alone. Additional API authorities, sovereign routing, binary/streaming operations,
and resource-specific route/envelope/permission contracts need their own evidenced adapters.
