# Paired team management

`sdk.GraphAPI.TeamManagement` composes backing-group properties and native team
state on one Graph tenant/cloud/identity binding. Native CRUD/list/actions remain
available through `Groups` and `Teams`.

| Method | 8.1.0 command branch | Ordered stages |
| --- | --- | --- |
| `GetByID` | `Get-Team` / Identity | Get group, get team, project both snapshots. |
| `UpdateByID` | `Set-Team` | Patch group, patch team settings, get group, get team. |

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

CORE-28 still requires the tenant/user filter branches of Get-Team, group-first
New-Team, the archive-state composition and successful lab acceptance. Native
archive/unarchive, resumable operation waiting and picture upload are available
separately through `Teams`.
