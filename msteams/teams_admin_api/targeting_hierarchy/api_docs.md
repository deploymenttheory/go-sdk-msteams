# Targeting hierarchy

`sdk.TeamsAdminAPI.TargetingHierarchy` owns the tenant's singleton operational
hierarchy. `NativeV1()` and `NativeV2()` expose independent resource primitives;
`UploadCSV`, `GetStatus` and `Delete` compose the three pinned public commands.
`Version("1")` and `Version("2")` select service protocols; zero selects v1.
These names are independent of PowerShell release numbers.

| Native resource | Available operations | Contract boundary |
|---|---|---|
| Targeting v1 | ImportCSV, GetOperation, Delete | Singleton update/import, status read and delete; no standalone hierarchy Create/Get/List established |
| Hierarchy v2 | GetHierarchy, CreateHierarchy, ImportCSV, GetOperation, Delete | Get/Create return status-bearing opaque responses; full hierarchy response schema and List remain unsettled |

V2 tenant paths and create body use the canonical GUID from the immutable client
session, with the evidenced `Operational Hierarchy` name. Callers cannot substitute
a tenant ID per operation. Configure a GUID tenant rather than a domain alias for
v2. Operation IDs remain opaque and escaped as one path segment; v1 requestId is
an encoded query value. Blank status selectors use latest. Supporting Get/Create
do not deserialize an invented hierarchy model. Successful HTTP handling follows
the pinned helpers' 2xx behavior; status/import models must still decode correctly.

`CSV` takes ownership of its `io.ReadCloser` immediately on import/upload,
including validation, cancellation, unavailable profile and failed preparation.
It closes the body exactly once; `Close` must unblock `Read`. Supply a positive
`MaxBytes` for the whole multipart wire payload including framing. This is a
caller budget, not a claimed server maximum. Filename becomes a portable basename.
Bytes are streamed unchanged in `csvFile`, with the pinned `application/json`
part type. The SDK performs no CSV parsing, conversion or full-file buffering.

Native ImportCSV sends just that import. Composed UploadCSV selects v1 directly,
or for v2 reads the bound tenant hierarchy, creates only after a received 404,
then imports. All other failed reads stop. Local input/profile failures close the
source before a possible write. A CSV source read error while streaming can occur
after preparation or partial transmission; UploadResult retains what is known.
`LookupHTTPAccepted`, `CreationHTTPAccepted`, `CreateAttempted`, preparation responses, `ImportAttempted`,
`ImportHTTPAccepted` and native import data remain available on later failure.
These flags describe HTTP acknowledgement, not completed creation/ingestion or
a verified read-back. V1 may omit requestId; v2 needs a
nonblank string id for a usable handle. Missing/invalid v2 id returns ErrDecode
with response/partial data, rather than fabricating an HTTP 400. Unknown response
map fields remain available, including shapes the pinned string dictionary rejects.

Create/import/delete are sent once. A timeout/failed response can leave remote
state and requires explicit reconciliation. The SDK never repeats a write,
deletes the singleton as compensation, restores prior contents, or polls silently.
GetStatus performs one read; Graph's 31-second poll interval does not apply here.

Native status models retain raw timestamps/status, nullable fields, nested JSON
error text and future data. DecodeErrors explicitly interprets v1 error JSON.
GetStatus also produces the pinned presentation: v2 status normalization, grouped
error descriptions, JSON argument order and platform line endings. Numeric and
comma-separated enum strings follow the observed CLR constructor quirk; RawStatus
remains separate and the SDK never infers native success from that display text.
Malformed projection details return a readable native status plus an error. Normal
formatting of status/upload results is redacted; explicit fields remain available.

Targeting v1 uses the commercial retail audience, including GCC's distinct retail
request origin. Hierarchy v2 uses the hierarchy origin and chatsvcagg audience.
Only the versioned closed cloud maps are enabled: targeting commercial/GCC,
hierarchy commercial. Fixed namespaces are enforced before token acquisition.
No ConfigAPI discovery/header, Graph/ConfigAPI token substitution, arbitrary
special-service override or commercial fallback is used. Unsupported local maps
are not evidence of remote service denial. Native providers use exact resource
`/.default` scopes; other services' explicit delegated scopes are not introduced.

The hierarchy guide requires a Teams service administrator for upload. Exact
per-operation grants, application eligibility, cloud feature deployment and role
enforcement remain unqualified. A valid token does not prove any of them. See
[pinned endpoint/profile evidence](../../../docs/implementation/core/hierarchy-migration-contract-evidence.md)
and [candidate SDK evidence](../../../docs/implementation/core/targeting-sdk-evidence.md).
Tenant import/delete require a separately reviewed singleton backup/restore
procedure; the prepared lab cases are read-only and unrun.
