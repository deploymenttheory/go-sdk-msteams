# CORE-01 membership resource and composition contract

Reviewed **2026-11-20** against the Microsoft Graph v1.0 documentation linked below
and the pinned MicrosoftTeams 8.1.0 implementation. This settles the resource boundary
for `Add-TeamUser`, `Get-TeamUser` and `Remove-TeamUser`; it supplies fixture cases,
not live tenant qualification. No tenant calls were made.

## Canonical service decision

`GraphAPI.TeamMembers` owns the Graph **conversationMember resource**, with genuine
`Create`, `GetByID`, `UpdateByID`, `DeleteByID` and `List` operations under
`/v1.0/teams/{teamID}/members`. Its `PromoteOwner` and `DemoteOwner` conveniences
call `UpdateByID`; promotion and demotion do not delete the membership. Microsoft
documents all five operations on this resource. This supersedes the backing-group
interpretation of `GraphAPI.TeamMembers` in the earlier
[tests](../../../tests/integration/lab).
[Resource](https://learn.microsoft.com/en-us/graph/api/resources/conversationmember?view=graph-rest-1.0)

Keep `GraphAPI.GroupMembers` and `GraphAPI.GroupOwners` as separate **relationship**
services, each exposing `AddReference`, `List` and `RemoveReference`. The explicit
`TeamUserManagement.Add`, `.List` and `.Remove` compositions reproduce the pinned
commands through these group primitives and optional directory-user resolution.
They must not silently call native `TeamMembers` CRUD instead: the pinned commands
change/read backing-group relationships. The public PowerShell descriptions also
describe those group effects.
[Add-TeamUser](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/add-teamuser?view=teams-ps),
[Remove-TeamUser](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/remove-teamuser?view=teams-ps)

| Identity/model | Contract |
| --- | --- |
| `TeamID` / `GroupID` | Separately named inputs; these compositions use the team's backing group ID. |
| `MembershipID` | Opaque `conversationMember.id`, obtained from native create/list/get. Never decode, synthesize from a user GUID, or substitute a directory ID. |
| `DirectoryObjectID` / `DirectoryUserID` | Entra object identity used in group references and `aadUserConversationMember.userId`; not the conversation membership identity. |
| `TeamMember` | Read model retains `@odata.type`, membership `id`, `roles`, `displayName` and unknown fields. The AAD subtype adds `userId`, `tenantId`, `email`; chat-only `visibleHistoryStartDateTime` is not a team write field. |
| `CreateTeamMemberRequest` / `UpdateTeamMemberRequest` | Separate native write models; do not serialize a read model or group-reference DTO as either request. |
| `DirectoryObjectReference` / `DirectoryUser` | Group relationship request and directory read models; selected user properties may be null because of limited permissions. |
| `TeamUserSummary` | Composition output: `UserID`, `User` (UPN), `Name`, `Role`. No membership ID is invented. |

The opaque-ID rule and possible cross-tenant members come from
[list team members](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0).
The separate AAD `userId`/`tenantId` fields and chat-only history property come from
[aadUserConversationMember](https://learn.microsoft.com/en-us/graph/api/resources/aaduserconversationmember?view=graph-rest-1.0).
The summary projection is pinned `GetTeamUser.Transform` (source below).

## Native team-member CRUD/list

All routes below are relative to the selected Graph origin. Use its Graph audience,
Bearer authorization, escaped path segments and JSON content type for writes.
`{m}` means `MembershipID`, `{t}` means `TeamID`.

| Operation | Exact v1.0 request | Success/body | Evidence |
| --- | --- | --- | --- |
| Create | `POST /teams/{t}/members` | `201 Created`; one `conversationMember` | [Add](https://learn.microsoft.com/en-us/graph/api/team-post-members?view=graph-rest-1.0) |
| GetByID | `GET /teams/{t}/members/{m}`, no body | `200 OK`; one `conversationMember`; no OData query customization | [Get](https://learn.microsoft.com/en-us/graph/api/team-get-members?view=graph-rest-1.0) |
| UpdateByID | `PATCH /teams/{t}/members/{m}` | `200 OK`; updated `conversationMember`, **not** an empty response | [Update](https://learn.microsoft.com/en-us/graph/api/team-update-members?view=graph-rest-1.0) |
| DeleteByID | `DELETE /teams/{t}/members/{m}`, no body | `204 No Content`; no returned member | [Delete](https://learn.microsoft.com/en-us/graph/api/team-delete-members?view=graph-rest-1.0) |
| List | `GET /teams/{t}/members`, no body | `200 OK`; `value` collection; `$filter`, `$select`, `$top`; default 100, maximum 999 per page | [List](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0) |

The documented create body for an AAD owner is:

```json
{
  "@odata.type": "#microsoft.graph.aadUserConversationMember",
  "roles": ["owner"],
  "user@odata.bind": "https://graph.microsoft.com/v1.0/users('11111111-1111-4111-8111-111111111111')"
}
```

Use `roles: []` for an ordinary member; absence defaults to an ordinary member.
The bind URL must follow the selected Graph origin. Create can return `404` for a
disabled/blocked user. Application-permission guest addition is unsupported.
The documented two-second staggering guidance and delayed membership visibility
do not turn the `201` into an async operation or justify replaying an ambiguous POST.
[Create contract](https://learn.microsoft.com/en-us/graph/api/team-post-members?view=graph-rest-1.0)

Update writes only `roles` (`["owner"]` to promote, **explicit `[]`** to demote),
optionally with the documented AAD `@odata.type`. Omitted fields retain their
existing values; omission is not demotion. Do not encode an empty roles slice as an
omitted property, JSON null, or `["member"]`. `guest` is assigned by the server,
cannot be changed through this property, and a guest cannot become an owner.
[Update contract](https://learn.microsoft.com/en-us/graph/api/team-update-members?view=graph-rest-1.0)

The SDK list model must retain `@odata.nextLink`; `List` returns a page, while an
explicit iterator/all-pages helper follows continuation with shared origin checks.
No ID lookup helper may decode the opaque membership ID. A directory-user lookup
helper must search native members by their AAD `userId` and retain the actual
returned membership ID; it must report ambiguity rather than choose arbitrarily.
These are SDK rules derived from the resource distinction, not extra endpoints.

## Authorization boundary

These are **documented permissions**, not observed grants. Personal Microsoft
accounts are unsupported for the membership endpoints. Native team-member docs
do not prescribe the group-directory roles below; do not transfer those role lists
to a different endpoint family.

| Native team operation | Delegated work/school | Application |
| --- | --- | --- |
| Get/List | `TeamMember.Read.All`; higher `TeamMember.ReadWrite.All` | Least `TeamMember.Read.Group` (resource-specific consent); higher `TeamMember.Read.All`, `TeamMember.ReadWrite.All` |
| Create/Update/Delete | Least `TeamMember.ReadWriteNonOwnerRole.All`; higher `TeamMember.ReadWrite.All` | Same permissions |

Sources: [Get permissions](https://learn.microsoft.com/en-us/graph/api/team-get-members?view=graph-rest-1.0),
[List permissions](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0),
[Create permissions](https://learn.microsoft.com/en-us/graph/api/team-post-members?view=graph-rest-1.0),
[Update permissions](https://learn.microsoft.com/en-us/graph/api/team-update-members?view=graph-rest-1.0),
[Delete permissions](https://learn.microsoft.com/en-us/graph/api/team-delete-members?view=graph-rest-1.0).

The non-owner permission does **not** permit adding/removing owners or elevating
members to owner. Owner management therefore uses `TeamMember.ReadWrite.All`;
the operation table's least-privileged entry is not evidence of owner authority.
The permission reference explicitly covers role changes under ReadWrite.All.
It does not explicitly settle demotion using only NonOwnerRole; make no narrower
qualification claim. [Permission definitions](https://learn.microsoft.com/en-us/graph/permissions-reference#teammemberreadwritenonownerroleall)

## Supporting group-reference primitives

`{g}` is `GroupID`; `{u}` is the directory object/user ID, never `MembershipID`.
Both reference-delete methods must retain the literal `/$ref` suffix. Omitting it
from member deletion can delete the directory object when the app has sufficient
object permissions. [Member deletion](https://learn.microsoft.com/en-us/graph/api/group-delete-members?view=graph-rest-1.0)

| Service/method | Exact v1.0 request | Success | Least permission, delegated/application | Evidence |
| --- | --- | --- | --- | --- |
| GroupMembers.AddReference | `POST /groups/{g}/members/$ref`, `{"@odata.id":"{Graph}/v1.0/directoryObjects/{u}"}` | `204`, empty | `GroupMember.ReadWrite.All` / same (user member) | [Add](https://learn.microsoft.com/en-us/graph/api/group-post-members?view=graph-rest-1.0) |
| GroupMembers.List | `GET /groups/{g}/members` | `200`, direct `directoryObject` collection; not transitive | `GroupMember.ReadBasic.All` / same | [List](https://learn.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0) |
| GroupMembers.RemoveReference | `DELETE /groups/{g}/members/{u}/$ref`, no body | `204`, empty | `GroupMember.ReadWrite.All` / same | [Remove](https://learn.microsoft.com/en-us/graph/api/group-delete-members?view=graph-rest-1.0) |
| GroupOwners.AddReference | `POST /groups/{g}/owners/$ref`, object containing `@odata.id` of the user | `204`, empty | `Group.ReadWrite.All` / same | [Add](https://learn.microsoft.com/en-us/graph/api/group-post-owners?view=graph-rest-1.0) |
| GroupOwners.List | `GET /groups/{g}/owners` | `200`, owners collection; preserve discriminator/limited fields | `GroupMember.Read.All` / same | [List](https://learn.microsoft.com/en-us/graph/api/group-list-owners?view=graph-rest-1.0) |
| GroupOwners.RemoveReference | `DELETE /groups/{g}/owners/{u}/$ref`, no body | `204`, empty | `Group.ReadWrite.All` / same | [Remove](https://learn.microsoft.com/en-us/graph/api/group-delete-owners?view=graph-rest-1.0) |

Group member add accepts only users for Microsoft 365 groups. Duplicate/unsupported
member addition returns `400`; nonexistent referenced objects return `404`.
Not every `400` is a duplicate: documented replication delay can also cause `400`.
[Group member add](https://learn.microsoft.com/en-us/graph/api/group-post-members?view=graph-rest-1.0)
Group owner changes can take up to two hours to reach Teams; an owner also needs
membership for team capabilities. The owner-add documentation demonstrates
`@odata.id: {Graph}/v1.0/users/{u}`; the pinned cmdlet instead uses
`{Graph}/v1.0/directoryObjects/{u}`. Preserve that exact pinned payload in its
composition evidence; no live acceptance of either form was tested here.
[Owner add](https://learn.microsoft.com/en-us/graph/api/group-post-owners?view=graph-rest-1.0)

Delegated group writes require a supported directory role in addition to OAuth
permission. Group owners, Directory Writers, Groups Administrator and User
Administrator qualify for user member/owner changes; Identity Governance
Administrator is additionally listed for member changes. Exchange, SharePoint,
Teams and Yammer Administrators qualify for Microsoft 365 groups. Custom roles
need `microsoft.directory/groups/members/update` or the applicable owners-update
permission. Role-assignable member changes additionally require
`RoleManagement.ReadWrite.Directory` and a supported user role (least Privileged
Role Administrator). Dynamic membership cannot be removed through the member
delete endpoint. Sources: [Member add roles](https://learn.microsoft.com/en-us/graph/api/group-post-members?view=graph-rest-1.0),
[Member remove roles](https://learn.microsoft.com/en-us/graph/api/group-delete-members?view=graph-rest-1.0),
[Owner add roles](https://learn.microsoft.com/en-us/graph/api/group-post-owners?view=graph-rest-1.0),
[Owner remove roles](https://learn.microsoft.com/en-us/graph/api/group-delete-owners?view=graph-rest-1.0).

For ordinary group reads, supported delegated principals include group owners,
member users, guests with limited rights, Directory Readers/Writers and Groups
Administrator. Custom permissions are members read/limitedRead or owners read.
Hidden-member reads additionally need `Member.Read.Hidden` and the appropriate
user role. Insufficient resource-type permission can leave only `id` and
`@odata.type` populated; null `userType` is not proof that the user is a member.
[Members](https://learn.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0),
[Owners](https://learn.microsoft.com/en-us/graph/api/group-list-owners?view=graph-rest-1.0)

Both group lists support `$filter/$count/$select/$search/$top/$expand`; advanced
query restrictions apply, and `$expand` cannot combine with advanced queries.
Member-list default/max page sizes are 100/999. Its advanced filter/search/order/cast
requests require `ConsistencyLevel: eventual` and `$count=true`. The pinned Add
preflight omits those advanced-query additions; do not mistake its status-only
check for a verified existence query. Follow all returned continuation links in
composition reads. [Members query contract](https://learn.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0),
[Owners query contract](https://learn.microsoft.com/en-us/graph/api/group-list-owners?view=graph-rest-1.0)

## Pinned PowerShell composition mapping

Pinned source: package SHA256
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`;
`netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll` SHA256
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`.
ILSpy output reviewed at `/private/tmp/msteams-core-decompiled`, under namespaces
`Microsoft.Teams.PowerShell.TeamsCmdlets` and `.Utils`. The reproducible package
source and inspection provenance are in [baseline inspection](baseline-inspection.md).
Methods named below are the primary implementation evidence, not inferred routes.

| Command/composition | Exact observed sequence and output |
| --- | --- |
| `Add-TeamUser` / TeamUserManagement.Add | `AddTeamUser.ProcessRecord` resolves non-GUID `User`, calls `AddUserToUnifiedGroup` for members, then owners only for case-insensitive `Role=Owner`. Omitted/Member role adds membership only. Each helper GETs `/v1.0/groups/{g}/{members\|owners}?$filter=Id eq '{u}'&$select=id`; **any 2xx skips POST**, even `value: []`. Any non-2xx proceeds to reference POST, using `@odata.id={Graph}/v1.0/directoryObjects/{u}`. No resource output. |
| `Get-TeamUser` / TeamUserManagement.List | `GetTeamUser.ProcessRecord` reads all owners unless Guest-only; all members unless Owner-only. Both use `$select=id,displayName,userPrincipalName,userType`. Owners take precedence over duplicate members; remaining `userType=guest` (case-insensitive) becomes guest; others become member. No-role order: owners, members, guests. Role Member excludes owners and guests; runtime also accepts Guest although current help omits it. `Transform` emits `UserId=id`, `User=userPrincipalName`, `Name=displayName`, lowercase Role. |
| `Remove-TeamUser` / TeamUserManagement.Remove | `RemoveTeamUser.ProcessRecord` resolves user ID, reads **all** `/v1.0/groups/{g}/owners?$select=id`, and if target is owner and owner count is one throws local synthetic `ApiException(412)` before DELETE. Existing owner reference is deleted first. Role Owner stops there; omitted/Member then deletes member reference. An Owner request for a non-owner makes no deletion. No resource output. |

The GroupOwners primitive must also preserve the server's last-user-owner rejection;
the documentation does not assign it the pinned synthetic `412` status. The local
precheck is not a concurrency guarantee. [Remove owner](https://learn.microsoft.com/en-us/graph/api/group-delete-owners?view=graph-rest-1.0)
`Remove-TeamUser -Role Owner` is demotion retaining membership; Member or omission
removes both relationships. [PowerShell role semantics](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/remove-teamuser?view=teams-ps)

**Explicit compatibility difference:** native reference primitives always report
the actual result. A semantic Add composition may inspect the returned collection
to distinguish absent from present and must propagate an unauthorized preflight.
That corrects the pinned status-only preflight and its mutate-on-any-error branch;
it must be labeled/tested as a correction, not claimed as byte-for-byte PowerShell
behavior. Preserve the pinned `200 + [] -> no POST` branch in oracle fixtures.

`GraphUtilities.GetUserId` directly GETs `/v1.0/users/{User}`; only `404` triggers
the pinned `/v1.0/myorganization/users?$filter=mail eq '{User}'` fallback, taking
the first result. GUID inputs skip these reads. Keep this alias and first-result
behavior in pinned evidence; a canonical resolver uses documented `/users` queries,
escapes literals and reports multiple matches. Reading other users needs delegated
`User.ReadBasic.All` (or greater); application `User.Read.All`. User list has the
same least privileges and cannot be called by guests. These optional resolution
permissions are additional to group permissions. [Get user](https://learn.microsoft.com/en-us/graph/api/user-get?view=graph-rest-1.0),
[List users](https://learn.microsoft.com/en-us/graph/api/user-list?view=graph-rest-1.0)

Compositions expose successful stages/IDs when a later stage fails (membership
added but ownership failed; owner removed but membership removal failed). They
provide no cross-request transaction, automatic rollback or blind mutation replay.
Group changes and client visibility are not immediate; completion of a reference
write is not proof that all Teams clients have synchronized.
[PowerShell propagation note](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/add-teamuser?view=teams-ps)

## Finite fixture acceptance set

These are required downstream cases, not fixtures claimed to have been captured.
Use synthetic IDs and Graph-shaped errors; label documentation-derived examples
separately from pinned serializer/control-flow observations.

| Cases | Assertions |
| --- | --- |
| Native C/R/U/D/L | Exact five methods/routes/statuses above; get/update decode one member; list decodes a page; delete decodes no resource. |
| Native role writes | Owner create/promotion uses `["owner"]`; member create/demotion retains `[]`; omitted roles and null are not substituted for explicit demotion; reject guest promotion. |
| Identity separation | Opaque member ID containing punctuation survives path encoding; `id != userId`; directory IDs cannot address native Update/Delete. |
| Native paging | Two pages with `@odata.nextLink`, an empty terminal page, unknown discriminator/role, nullable AAD fields and another tenant ID. |
| Native authorization/errors | Non-owner vs owner permission contract, guest app-add unsupported, disabled-user `404`; retain status and Graph error code. |
| Reference writes | Member and owner exact `$ref` routes; member JSON binding; documented owner `/users` binding and separately pinned `/directoryObjects` binding; `204` empty bodies. |
| Pinned Add preflight | `200 + []` skips POST; nonempty 200 skips POST; non-2xx takes POST branch. Canonical correction separately checks collection and preserves authorization errors. |
| Pinned Add owner sequence | Membership stage then ownership; omitted/Member skips owners; failure at ownership retains membership-stage result. |
| Pinned Get projections | Owner/member duplicate removal; guest classification; all three filters; Owner-only skips members, Guest-only skips owners; exact output fields. |
| Limited directory data | `id/@odata.type` with null user fields stays unknown/limited in native models; pinned null-userType-to-member projection remains separately evidenced. |
| Pinned Remove stages | Last-owner local `412` causes zero deletes; Role Owner deletes only owner; Member/omitted deletes owner then member; non-owner Owner request is a no-op; second-stage failure remains partial. |
| Optional user resolution | GUID skips lookup; direct lookup success; only 404 fallback; zero/multiple mail results; pinned first-result behavior versus canonical ambiguity error. |

No live, application-authentication, cloud or propagation timing qualification is
inferred from this note. All current Graph endpoint claims above are documentation
evidence reviewed on the stated date; the three command request sequences are
pinned implementation evidence.
