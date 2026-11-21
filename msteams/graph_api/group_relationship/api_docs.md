# Group member and owner references

`sdk.GraphAPI.GroupMembers` and `sdk.GraphAPI.GroupOwners` are distinct services
for directory relationships. They do not use native conversation membership IDs
or silently call `TeamMembers` CRUD. The Add/Get/Remove-TeamUser compositions,
including preflights, last-owner checks and partial multi-stage outcomes, remain
separate CORE-26 work.

| Service | AddReference | List | RemoveReference |
| --- | --- | --- | --- |
| GroupMembers | `POST /v1.0/groups/{g}/members/$ref` | `GET /v1.0/groups/{g}/members` | `DELETE /v1.0/groups/{g}/members/{id}/$ref` |
| GroupOwners | `POST /v1.0/groups/{g}/owners/$ref` | `GET /v1.0/groups/{g}/owners` | `DELETE /v1.0/groups/{g}/owners/{id}/$ref` |

Adds/removals require 204; reads require 200. Mutations never replay, preflight,
infer duplicate success from a 400, or infer absence from authorization failures.
Group replication can cause 400 errors, so that status alone cannot establish a
duplicate relationship. Every delete fixes the literal `/$ref` suffix in the
resource operation; callers cannot omit it and accidentally address the directory
object instead. Owner removal does not remove membership. Graph's actual
last-user-owner rejection is returned; this primitive does not invent a 412.

`Reference` supplies a UUID directory ID and a known collection kind, with no
arbitrary reference URL. With no kind, member adds use the documented
`directoryObjects/{id}` binding and owner adds use `users/{id}`. Explicit
`DirectoryObjects` retains the pinned owner binding variant; explicit `Users`
and `ServicePrincipals` are also selectable. Bindings always use the configured
Graph cloud. Graph enforces allowed object types and group/principal restrictions;
Microsoft 365 group membership supports users, while security groups support
additional member types. No directory object is created or resolved here.

Read models preserve the discriminator, object ID, nullable user/name/UPN/mail
fields, optional app ID and complete unknown data. A limited `id/@odata.type`
object with null fields is not classified as an ordinary user. Native v1.0 has
documented service-principal omissions in these relationship endpoints; an
endpoint-complete traversal does not establish that omitted object types are
absent. No membership ID is synthesized from the directory ID.

`List` returns one direct relationship page, not a transitive collection.
Filter/search/select/expand/top/count and the eventual-consistency header are
explicit options; expressions are not interpreted or rewritten. Member Top is
bounded by the documented maximum of 999; owners Top is passed through without
inventing an undocumented maximum. Defaults remain server-selected. Advanced
queries can use an index with delayed visibility, and expand has separate query
restrictions. The server enforces those combinations.

`NewPager`/`ListAll` explicitly traverse opaque next links using the shared Graph
origin check, HTTP 200 expectation and traversal limits. Count/query/header
values are frozen at construction. MaxPages defaults to 1,000; Collect/ListAll
aggregation defaults to 100,000 items. Error/limit outcomes retain prior items
and the last response and are incomplete. No hidden role assignment, directory
lookup, wait, rollback or identity fallback occurs.

Documented user-member writes require GroupMember.ReadWrite.All; owner writes
require Group.ReadWrite.All, with additional delegated principal requirements.
Read scopes, hidden-member access and object-type read permission determine
visibility. Owner changes can take up to two hours to synchronize with Teams;
an owner also needs membership for relevant team capabilities. These are
documented boundaries, not observed lab eligibility or timing qualifications.

Primary sources reviewed on 2026-11-20:
[member add](https://learn.microsoft.com/en-us/graph/api/group-post-members?view=graph-rest-1.0),
[member remove](https://learn.microsoft.com/en-us/graph/api/group-delete-members?view=graph-rest-1.0),
[member list](https://learn.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0),
[owner add](https://learn.microsoft.com/en-us/graph/api/group-post-owners?view=graph-rest-1.0),
[owner remove](https://learn.microsoft.com/en-us/graph/api/group-delete-owners?view=graph-rest-1.0),
[owner list](https://learn.microsoft.com/en-us/graph/api/group-list-owners?view=graph-rest-1.0).
No relationship tenant calls or mutations were made in this implementation work.
