# Application meeting configuration

`Client.TeamsAdminAPI.ApplicationMeetingConfigurations` exposes the modern
ConfigAPI configuration used by MicrosoftTeams 8.1.0. It manages the tenant
allow-list of applications permitted to remove meeting participants.

| Operation | Native behavior | PowerShell mapping |
|---|---|---|
| GetByIdentity | GET one escaped configuration identity, HTTP 200 | Get / Identity |
| List | GET bare array, stable Global-first order | Get / omitted Identity |
| ListFiltered | List plus local Identity/Tag:Identity wildcard matching | Get / Filter |
| UpdateByIdentity | Sparse PUT including Identity, HTTP 204 | Set / Identity |
| UpdateGlobal | Explicit Global identity default | Set / omitted Identity |

The configuration type is `ApplicationMeetingConfiguration`. Create and Delete
have no established contract for this resource. The service adds no pagination,
PATCH substitution, automatic read-back, mutation retry or PowerShell dependency.
Global is the documented instance; identity strings remain open to the service
rather than being treated as a locally proven enum.

`UpdateRequest.AppIDs == nil` omits `AllowRemoveParticipantAppIds`.
`ReplaceAppIDs` copies a string list; nil/empty emits `[]` to clear it.
`ChangeAppIDs` emits capitalized Add/Remove arrays and omits empty operators;
both empty emits `{}`. `LegacyAppIDs`/`ChangeLegacyAppIDs` preserve the pinned
helper's object-item compatibility after removing serialization metadata. This
does not establish backend acceptance of objects as application IDs. The zero
change is invalid and fails before discovery. No app ID defaults or UUID-only
restriction are invented from mutable documentation.

Read snapshots retain null/omitted/future fields in `RawFields`; `AppIDs` projects
string arrays. Marshaling a snapshot preserves its received fields, and snapshots
are separate from write requests. Native errors retain raw response metadata;
ordinary formatting redacts payloads. Context cancellation propagates throughout.

The existing selected-cloud ConfigAPI origin, Teams resource token and per-client
forest discovery/header adapter apply. Source evidence does not prove modern
flight eligibility, RBAC, app/delegated authorization or deployment in every
cloud. [unit tests](service_test.go).
