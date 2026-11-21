# Paired team management

`sdk.GraphAPI.TeamManagement` composes backing-group properties and native team
state on one Graph tenant/cloud/identity binding. Native CRUD/list/actions remain
available through `Groups` and `Teams`.

| Method | 8.1.0 command branch | Ordered stages |
| --- | --- | --- |
| `GetByID` | `Get-Team` / Identity | Get group, get team, project both snapshots. |
| `List` | `Get-Team` / Filters | Complete tenant/user discovery, bounded hydration, filtering and typed partial outcomes. |
| `Create` | `New-Team` / CreateTeam | Resolve owner, create group, wait for availability, teamify, paired readback. |
| `MigrateGroup` | `New-Team` / MigrateGroup | Read supplied group, teamify with default settings, paired readback; never delete the supplied group. |
| `UpdateByID` | `Set-Team` | Patch group, patch team settings, get group, get team. |
| `SetArchivedState` | `Set-TeamArchivedState` | Paired read, state check, submit once, wait, paired readback, state confirmation. |

IDs must be directory UUIDs. Both read snapshots must match the requested ID.
`ManagedTeam` takes display name, description, visibility, mail nickname and
classification from the group; settings, internal ID and archive state come
from the team. Nullable fields remain unknown. Both native snapshots retain
future fields. Paired reads have no transaction or shared-version guarantee.

`UpdateRequest` separates group properties from settings. Omitted group fields
stay omitted; explicit empty strings and false values are preserved. All five
settings containers are supplied, including empty omitted containers, matching
the actual pinned serializer capture without modifying caller inputs. Native
`Teams.UpdateByID` remains sparse. Giphy ratings use native strings; the pinned
numeric serializer fixture is unchanged.

`Result` retains the known ID, ordered stages and latest received response on
failure. `Value` requires both reads to succeed. Completed writes remain applied
when a later write/read fails; an unconfirmed mutation may have taken effect
remotely. The composition stops on error without replay, rollback or cleanup.
Context cancellation and client close cover the entire composition.

The caller needs access to both resource endpoints; see the [group notes](../group/api_docs.md)
and [team notes](../team/api_docs.md). No roles or permissions are granted.
Offline tests execute the root SDK with synthetic source-derived scenarios and
the actual empty-container capture, not live cmdlets or tenant CRUD.

`SetArchivedState` rejects an unknown or already-requested archive state after
its paired pre-read. A SharePoint option is invalid for unarchive, and any
explicit SharePoint option requires delegated mode. Invalid wait bounds or
unevidenced propagation allowances fail before any HTTP request. Zero wait
options select the shared defaults; native polling enforces its minimum cadence
and Retry-After. Action submission is not replayed. HTTP 202 is recorded as
Accepted; completion requires both a successful operation and a matching paired
readback. Missing or stale readback yields `ErrArchivedStateUnconfirmed`, retaining
the observed `After` snapshot and successful operation without inventing state.

`ArchiveStateResult` separates Before, After, remote acceptance and the last
operation Snapshot. Cancellation, timeout, failed/unknown operation states or
readback failure retain stages and recovery metadata; they do not roll back or
prove a mutation was absent. Default formatting is redacted. The retained
`Accepted.Operation` can wait again under a new context while preserving its
cadence. Persist recovery IDs privately with tenant/cloud/identity binding; after
process restart, use `Teams.ResumeOperation`, wait, then `GetByID` and explicitly
check archive state. Resumption must never submit the original action again.

The original pinned archive JSON-null and unarchive JSON-empty-string captures
remain unchanged. Native omitted options/unarchive send no body. Current primary
documentation reviewed 2026-11-21: [archive](https://learn.microsoft.com/en-us/graph/api/team-archive?view=graph-rest-1.0),
[unarchive](https://learn.microsoft.com/en-us/graph/api/team-unarchive?view=graph-rest-1.0),
[async operation](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-1.0).
Offline virtual-time tests do not qualify successful tenant archive transitions.

The current [v1.0 OpenAPI metadata](https://github.com/microsoftgraph/msgraph-metadata/blob/master/openapi/v1.0/openapi.yaml)
declares a `2XX` JSON envelope for the team-operation item GET (reviewed blob
`d82714867604667937f58d0286e5db3cf434c243`). The existing shared interpretation is
preserved; no HTTP-200-only rule is inferred from the separate beta chat-operation
documentation. HTTP acceptance cannot establish a successful terminal state.

See the [creation and migration notes](creation_api_docs.md) for pinned group
presets, generated aliases, propagation controls, ownership and compensation.
See the [listing notes](list_api_docs.md) for tenant search intersection, direct
joined teams, batch correlation/throttles and typed partial outcomes. Remaining
command evidence and successful lab acceptance are required before CORE-28 can
close. Native archive/unarchive, resumable
operation waiting and picture upload remain available separately through Teams.
