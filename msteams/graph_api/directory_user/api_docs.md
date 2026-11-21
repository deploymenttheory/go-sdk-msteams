# Supporting directory-user resolution

`sdk.GraphAPI.UserResolver` supports team compositions; it does not expose
directory-user creation/deletion or grant access. Construction is local. Every
read uses the root Graph session and caller context.

`Get` reads `/v1.0/users/{id|UPN}` with optional select fields. Selectors remain
escaped segments; a leading-$ UPN uses the documented OData key form, with quote
escaping. `Current` reads `/v1.0/me` only in configured delegated mode. Application
callers supply an explicit identity. Reads retain nullable/future fields and the
complete object in `RawFields`. A directory UUID is required in a returned user.

`Resolve` accepts a UUID unchanged without network or existence/type/authority
verification. Non-UUID selectors first use direct lookup with `$select=id`.
Only an actual HTTP 404 permits fallback to canonical `/v1.0/users`, using an
escaped `mail eq '…'` literal, `$select=id` and `$top=2`. It does not use the
pinned `myorganization` alias or silently choose its first result. Authentication,
authorization, malformed responses and other failures do not trigger fallback.

Fallback follows every page until completion, a second distinct identity, or a
shared traversal error. Duplicate observations of the same ID are deduplicated.
The default bound is 16 pages; callers may set a positive MaxPages. No match
returns `ErrNotFound`; distinct matches return `ErrAmbiguous`. A limit, forbidden
continuation, malformed page or local cancellation cannot establish absence or
uniqueness, and exposes no usable resolved ID. Response metadata is retained.

An omitted selector means current user only with `AllowCurrent` explicitly set.
All errors and formatted resolution values omit selectors/IDs. A documented
202 accepted user read is surfaced as `ErrPending`, retaining response and an
available snapshot for inspection; it does not establish a usable completed
lookup or trigger mail fallback, polling or a mutation.

Resolving other users needs the appropriate delegated basic-profile or
application User.Read.All authority; `/me` has its signed-in-user permission
boundary. GUID passthrough avoids an unnecessary directory read but does not
confer permission for later group/team actions. These are documented boundaries,
not lab least-privilege qualification. No user-resolution tenant calls were made.

Primary sources reviewed on 2026-11-20:
[user get](https://learn.microsoft.com/en-us/graph/api/user-get?view=graph-rest-1.0),
[user list](https://learn.microsoft.com/en-us/graph/api/user-list?view=graph-rest-1.0).
Pinned/native differences are recorded in the
[membership contract](../../../docs/implementation/core/membership-resource-contract.md).

## Batched principal-name reads

LookupPrincipalNames validates/deduplicates explicit directory UUIDs, then reads
beta users in groups of at most ten with bounded opaque continuation. It returns
lowercase UUID keys: a present nil value preserves a returned null/absent UPN,
while a missing key means no directory result. Empty input performs no HTTP.
Failures retain only earlier completely validated groups and the last response;
a failing/ambiguous group is not partially published. Unrequested IDs fail
validation. Callers retain their initial UPN/email only for absent keys. The
channel-user and shared-channel projections reuse this supporting operation.
