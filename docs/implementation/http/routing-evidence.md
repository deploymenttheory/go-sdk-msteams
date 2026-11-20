# HTTP-01 routing evidence

Status: transport implementation in progress. This records source evidence and unresolved
routing questions; it does not claim a successful Teams management request.

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
These are static source candidates, not an approved production allowlist. Inspect the
associated environment/field assignments and call sites before enabling a destination.
In particular, a wildcard for the entire Teams domain would wrongly include development
hosts. The `ConfigApiAutoDForestProvider` methods also contain discovery-related URL
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
