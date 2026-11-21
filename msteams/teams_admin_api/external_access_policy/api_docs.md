# External access policies

`Client.TeamsAdminAPI.ExternalAccessPolicies` uses the compiled provider's direct ConfigAPI
protocol for MicrosoftTeams 8.1.0. It shares the selected Teams audience and
client lifecycle. It adds MPACmdlet and per-request correlation/command metadata;
it performs no generated forest discovery or Graph identity lookup.

| Operation | Behavior |
|---|---|
| GetByIdentity | GET one identity; preserve the first scope/name slash and escape each side |
| List | One GET of a bare array; stable exact-Global-first ordering |
| ListFiltered | List plus local Identity/Tag:Identity wildcard matching |
| PatchByIdentity | One sparse PATCH with materialized full arrays |
| UpdateByIdentity | GET current, apply bound scalar/list changes, then one sparse PATCH |
| Create | One sparse POST; no implicit comparison/read-back |
| CreateWithGlobalComparison | Explicit comparison GET for non-Global names, then Create; comparison failure does not stop creation |
| DeleteByIdentity | One DELETE with Identity and optional Force in its JSON body |
| Assign / AssignUser / AssignTenantDefault / AssignGroup | Explicit user, tenant or group scope; no polling or generated operation-handle decoding |

`policy.Set(value)` binds false, empty strings and numeric values. `policy.Null[T]()`
binds JSON null; a zero `policy.Field[T]` omits the property. Native list fields
use `policy.Strings(...)`: nil field omits, an empty value sends `[]`. Scalars and
full arrays are independent; no service defaults are inserted.

Composed updates use `policy.ReplaceStrings` or `policy.ChangeStrings` for
AllowedExternalDomains / BlockedExternalDomains. They copy current data, remove the first ordinal match per Remove item,
then append Add values without deduplication. A nonnil empty edit clears the list,
matching the compiled cmdlet's IsEmpty branch. These are local read/modify/write
edits; no server Add/Remove operator object, atomicity or ETag guarantee is inferred.
The PATCH body uses the read model's Identity while the route keeps the caller's
identity. PatchByIdentity exposes BodyIdentity explicitly for native callers.

At least one family property must be bound for Update/Patch; Force alone is not
an update. Bound CRUD Force becomes string `"True"`/`"False"`; unbound Force is
absent. Caller code owns confirmation/planning; no implicit PowerShell WhatIf or
interactive prompt is implemented.

Any 2xx follows the pinned provider's optional-object dispatch. Empty/null bodies
remain nil, with raw status/body/headers available. A 202 is not propagation or
workflow completion. `Policy.RawFields` retains omitted/null/future data; native
reads preserve wire presence rather than materializing CLR constructor defaults.
String/enum values are not closed from mutable documentation. Error metadata,
warnings and malformed bodies remain available; standard model/error formatting
redacts tenant data. Mutations never replay, and updates retain preparation and
HTTP acceptance if a later step fails. No rollback is automatic.

CommunicationWithExternalOrgs uses the evidenced numeric Int32 enum; unknown
numeric values remain representable. Null collection fields rejected by the
pinned model are decode errors, with the raw response retained. Assignment
PolicyName nil versus empty and tenant Force Boolean versus CRUD Force string
remain distinct. Group Rank is Int32 with no invented backend range. Assign's
zero/mixed scope rejects locally; it is not a fourth assignment type. Clear,
default/effective-policy and group propagation semantics remain unqualified.

Exact RBAC/consent, cloud feature deployment and successful app/delegated service
behavior remain unqualified. [unit tests](service_test.go).
