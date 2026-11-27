# Channels

`sdk.GraphAPI.Channels` owns native Graph channel CRUD. Routes explicitly use
`/beta`, matching MicrosoftTeams 8.1.0. Construction is local; all requests use
the root client's tenant, selected Graph cloud and Graph audience.

| Method | Behavior |
| --- | --- |
| `Create` | One POST; standard/private return a channel at 201; shared returns an accepted operation at 202 |
| `GetByID` | One 200 resource read, with optional select/unknown-enum preference |
| `UpdateByID` | One PATCH of display name/description; requires 204 |
| `DeleteByID` | One channel DELETE; requires 204 |
| `List` | One hosted/all/incoming page with an opaque next link |
| `NewPager`, `ListComplete` | Lazy or collected bounded paging; failures retain partial results |
| `ListAll`, `ListIncoming` | Complete distinct allChannels/incomingChannels collections |
| `CreateWithOwner` | Explicit directory-user resolution followed by creation |
| `ResolveByDisplayName`, `UpdateByDisplayName`, `DeleteByDisplayName` | Complete unique hosted-name lookup and optional ID-based mutation |
| `ResumeOperation` | Local reopening of a validated beta team-operation location |

`CreateWithOwner` retains `OwnerResolution` and its last `OwnerResponse` alongside
the creation result. If a later local owner guard or cancellation prevents the
POST, `Response` retains that preparation response. When POST returns an HTTP
response, `Response` describes creation while `OwnerResponse` remains independently
available, including on denial or accepted malformed output. A directory read
does not establish creation eligibility. Primitive `Create` has no resolution
metadata; no retry or compensation is added by the composition.
If creation produces no HTTP response, the fallback is a lookup receipt; it does
not establish the creation outcome.

IDs remain opaque channel strings and are escaped as individual segments. Team
and owner directory IDs must be UUIDs. Creation defaults to `standard`; only
standard/private/shared are writable membership types. Owner input is allowed
for private/shared. Nil description/name pointers omit properties; pointers to
empty strings retain explicit empty values. An empty patch stays `{}`. Typed
read fields and RawFields preserve null/absence/future fields; RawFields never
becomes a write payload. Moderation settings and summaries remain raw JSON.

An accepted shared create is not a channel. Its Location must match the selected
Graph origin, beta team-operation family and initiating team. Versionless Graph
locations resolve to beta explicitly. No query, fragment, alternate origin/team,
v1.0 status route or arbitrary same-origin URL is accepted. Polling uses the
shared waiter with a 31-second minimum; caller cancellation preserves the handle.
No create/update/delete mutation is automatically replayed. Operation target URLs
remain metadata. `HostTeamID` passively parses the tenant-prefixed OData channel
identity for enhanced projections and rejects unknown/inconsistent shapes.

Current documented tenant-wide application permissions are Channel.ReadBasic.All
for reads, Channel.Create for creation, ChannelSettings.ReadWrite.All for update,
and Channel.Delete.All for deletion. Resource-specific `.Group` alternatives
require consent on the team. Owner lookup additionally needs suitable user-read
access. The role/consent, private/shared eligibility, external access and cloud
restrictions remain server decisions. A token or Teams role alone proves none
of these operations works. Never turn a 403 into an unsupported-feature result.

The canonical payload uses documented lowercase membership values and configured
cloud bindings. The pinned serializer used uppercase values and hard-coded the
commercial bind origin. The SDK rejects an owner on standard creation even when
type was omitted; pinned PowerShell bypassed that guard for omitted type. Name
composition traverses all bounded pages and rejects ambiguity; pinned Set/Remove
searched the first page. These are explicit SDK decisions, not byte-equivalence
claims. The historical myorganization user fallback is not used.

See [unit tests](service_test.go)
and [unit tests](service_test.go).
Live tenant CRUD and PowerShell conformance remain unqualified.

Shared-channel 202 receipts accept Location and, when it is absent, the
[documented Content-Location](https://learn.microsoft.com/en-us/graph/api/channel-post?view=graph-rest-beta).
Both pass the same Graph origin and addressed-team operation-route validation.
A present malformed Location remains an error; the fallback does not hide it.
Acceptance retains an operation handle rather than a completed channel, and local
wait cancellation does not cancel or resubmit the remote create.

`CreateRequest.SharedWithTeams` accepts typed `InitialSharedTeam` references for
initial sharing while creating a shared channel, as documented in beta Example8.
Nil omits the property; a pointer to an empty collection sends `[]`; a pointer to
a nil collection preserves explicit JSON null. Null or incompatible membership
inputs remain service-validated, with no acceptance claim from serialization. References
carry team UUIDs. The [compiled example](example_test.go) shows sharing the new
channel with its host team. The service does not infer relationship POST/PATCH
operations or claim that arbitrary cross-tenant recipients are authorized.
