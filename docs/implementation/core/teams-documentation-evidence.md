# CORE-01 teams and team-members documentation evidence

Research date: 2026-11-20. Scope: `Get-Team`, `New-Team`, `Set-Team`,
`Remove-Team`, `Set-TeamArchivedState`, `Set-TeamPicture`, `Add-TeamUser`,
`Get-TeamUser`, and `Remove-TeamUser`. This is documentation research, not a
completed endpoint contract or a live qualification. No credentials or tenant
operations were used.

## Evidence boundary

The target is MicrosoftTeams **8.1.0**, not whatever module version the current
Microsoft Learn pages describe. All web sources below were accessed on the
research date; Graph references explicitly use `graph-rest-1.0`. Published update
dates are recorded where exposed, but are not package-version guarantees.

Local primary inputs:

- [CORE-01 scope](../backlog/issues/CORE-01.md), including package SHA256
  `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
- [Normalized commands](../../../contracts/powershell/8.1.0/commands.json), SHA256
  `823d63c11c524fccbe37a4ef653066b044fdce40d515dc9a0a5ef6aa575362fa`.
- [Runtime metadata](../../../contracts/powershell/8.1.0/runtime.json), SHA256
  `7acfc9d635da928a2cf3deb4ff6f660f239edb218581613708a27a74f48aa1d8`.
  Environment records module 8.1.0, PowerShell 7.6.6, macOS 27.0.1, and
  `authenticated: false`.

All nine commands are runtime-observed metadata-only candidates. Every command's
`output_types` is empty: this means no declared output-type evidence, **not** that
the operation necessarily returns nothing. Metadata establishes bindings and
validation attributes; it cannot establish HTTP routes, default serialization,
server authorization, or completion semantics.

Confidence labels used here: **pinned** for normalized 8.1.0 metadata;
**documented-current** for Microsoft documentation; **proposed** for SDK design
inferences; **unresolved** where those sources do not settle parity.

The companion [baseline inspection](baseline-inspection.md) records initial
pinned-package IL findings across CORE-01. Its route evidence can narrow the
documentation-only uncertainties here; neither note declares complete contracts.

## Pinned callable surface

The following summarizes non-common parameters from runtime metadata. Unless
otherwise specified, parameters are strings; `?bool` means nullable Boolean,
and optional does not imply that explicit null and omission have identical wire
behavior. PowerShell common parameters are not backend fields.

| Command / parameter set | Required | Optional |
| --- | --- | --- |
| Get-Team / Filters | None | User, Archived (?bool), Visibility, DisplayName, MailNickName, NumberOfThreads |
| Get-Team / Identity | GroupId | Same filters and NumberOfThreads |
| New-Team / CreateTeam (default) | DisplayName | Description, MailNickName, Classification, Visibility, Template, Owner, GiphyContentRating, RetainCreatedGroup (switch), all Boolean settings below |
| New-Team / MigrateGroup | GroupId | **No non-common optional parameters** |
| Set-Team / __AllParameterSets | GroupId | DisplayName, Description, MailNickName, Classification, Visibility, GiphyContentRating, all Boolean settings below |
| Remove-Team / __AllParameterSets | GroupId | None |
| Set-TeamArchivedState / __AllParameterSets | GroupId, Archived (bool) | SetSpoSiteReadOnlyForMembers (?bool) |
| Set-TeamPicture / __AllParameterSets | GroupId, ImagePath | None |
| Add-TeamUser / __AllParameterSets | GroupId, User | Role |
| Get-TeamUser / __AllParameterSets | GroupId | Role |
| Remove-TeamUser / __AllParameterSets | GroupId, User | Role |

The shared nullable Boolean settings on `New-Team/CreateTeam` and `Set-Team` are:
`AllowAddRemoveApps`, `AllowChannelMentions`, `AllowCreatePrivateChannels`,
`AllowCreateUpdateChannels`, `AllowCreateUpdateRemoveConnectors`,
`AllowCreateUpdateRemoveTabs`, `AllowCustomMemes`, `AllowDeleteChannels`,
`AllowGiphy`, `AllowGuestCreateUpdateChannels`, `AllowGuestDeleteChannels`,
`AllowOwnerDeleteMessages`, `AllowStickersAndMemes`, `AllowTeamMentions`,
`AllowUserDeleteMessages`, `AllowUserEditMessages`, and
`ShowInTeamsSearchAndSuggestions`.

Pinned validation:

- `Get-Team.NumberOfThreads` is **System.String**, with `ValidateRange(1,20)`.
- `New-Team` rejects null/empty DisplayName, Description, MailNickName,
  Classification, Owner, and GroupId when supplied in their respective sets.
  Visibility's case-insensitive set is Public, Private, HiddenMembership;
  Template's is EDU_Class, EDU_PLC.
- Add/Remove-TeamUser Role accepts Owner or Member, case-insensitively.
  Get-TeamUser additionally accepts Guest.
- Set-Team's string properties have no recorded validation attributes; absence
  of an attribute is not evidence that the backend accepts empty values.

These facts all derive from the pinned runtime input above. Do not overwrite them
with current help text while resolving the divergences below.

## Documented PowerShell behavior

### Get-Team

The command supports organization-wide listing, membership-by-UPN listing, and
exact GroupId lookup. DisplayName and MailNickName are filters that can produce
multiple matches; DisplayName filtering is documented as case-sensitive and
matching included text. Archived omission includes both states; false must remain
distinct from omission. Visibility accepts Public, Private, HiddenMembership.
The help calls NumberOfThreads an Int32 with default 20 and range 1–20; it is a
retrieval-concurrency setting, not a server page-size contract. The page does not
specify wire routes, pagination, or a concrete output schema.
[Get-Team](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-team?view=teams-ps).

### New-Team

Creation returns a group-like object exposing GroupID. The documented default
visibility is Private and Giphy rating Moderate. DisplayName/Description limits
are 256/1024 characters. MailNickName must be tenant-unique and becomes required
under group naming policies. An explicit Owner becomes both member and owner;
otherwise the creator does. The page says Owner is required for certificate
authentication. RetainCreatedGroup defaults false: failed team creation normally
cleans up the newly created group; true retains it. Template requires an EDU
license. Classification is reserved for Microsoft internal use. Current syntax
lists optional Owner/settings/RetainCreatedGroup for MigrateGroup, contrary to the
pinned binding. The docs describe client propagation taking up to two hours with
an owner’s Teams application open. Exact default serialization, compensation
boundaries, and retry behavior remain unresolved.
[New-Team](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-team?view=teams-ps).

### Set-Team

Updates team properties and settings, including the settings previously exposed
by separate settings cmdlets. Documented DisplayName/Description limits are
256/1024 characters; Giphy rating is Strict or Moderate; Visibility is Public or
Private; MailNickName must be tenant-unique. Classification is internal-only.
ShowInTeamsSearchAndSuggestions is marked deprecated, although pinned metadata
still exposes it. The help assigns AllowCreatePrivateChannels to a CreateTeam
parameter set even though Set-Team's pinned set is __AllParameterSets. Omission,
explicit empty-string clearing, transactional behavior across group/team
properties, and returned values are not settled by the page.
[Set-Team](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-team?view=teams-ps).

### Remove-Team

Deletes the identified team and its associated Microsoft 365 unified group and
associated components. It is not a request to detach Teams while preserving the
group. The page gives no exact response/error envelope or completion timing.
[Remove-Team](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/remove-team?view=teams-ps).

### Set-TeamArchivedState

Archived true freezes team activity; false unarchives. Teams administrators and
owners retain membership and role-management abilities. The optional
SetSpoSiteReadOnlyForMembers requests a read-only SharePoint site for members.
The help's generic `System.Object` output does not establish a response schema or
whether PowerShell waits for server completion. Handling of the SharePoint option
when unarchiving is not specified.
[Set-TeamArchivedState](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-teamarchivedstate?view=teams-ps).

### Set-TeamPicture

ImagePath denotes a local PNG, GIF, JPG, or JPEG. The command returns before client
propagation, which may require the Teams client to remain open for up to an hour.
The page claims the cmdlet is beta-only and unavailable in TeamsGCCH/TeamsDoD;
the beta-only wording conflicts with its observed export in the pinned release.
It does not establish media conversion, maximum byte size, upload headers, or
HTTP destination. Treat cloud availability as unqualified until package/lab
evidence resolves the wording.
[Set-TeamPicture](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-teampicture?view=teams-ps).

### Add-TeamUser, Get-TeamUser, Remove-TeamUser

Add uses a user's UPN, defaults Role to Member, and adds the user to both the team
and backing group; Owner also adds membership. It returns before Teams-client
propagation, documented as 24–48 hours.
[Add-TeamUser](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/add-teamuser?view=teams-ps).

Get returns an array with UPN, UserId, Name, Role. The help describes Owner/Member
filtering but omits the pinned Guest value. Wire property names, nullability,
pagination, and how UPN is obtained remain unproven.
[Get-TeamUser](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-teamuser?view=teams-ps).

Remove with Role Owner removes ownership while retaining membership. Role Member
removes both ownership and membership; omitted Role removes the user from the
team. The last owner cannot be removed. Client propagation may take an hour.
The page's parameter table calls Role mandatory, contradicting its own syntax,
examples, and pinned metadata. An ownership-only removal must not be implemented
as unconditional membership deletion.
[Remove-TeamUser](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/remove-teamuser?view=teams-ps).

## Authentication evidence and conflict

The current application-authentication guide explicitly excludes **New-Team**,
despite the cmdlet help's certificate-authentication Owner guidance. This is an
unresolved first-party contradiction. Do not assert that adding Owner settles
8.1.0 app-only support. For other non-Cs commands the guide lists Graph grants
including User.Read.All, Group.ReadWrite.All, AppCatalog.ReadWrite.All,
TeamSettings.ReadWrite.All, Channel.Delete.All, ChannelSettings.ReadWrite.All,
and ChannelMember.ReadWrite.All; Organization.Read.All applies generally.
This is module-level guidance, not per-command least-privilege proof.

The guide describes distinct Graph and Skype/Teams Admin tokens and directory
roles controlling RBAC. Its commercial examples use Graph `.default` and resource
`48ac35b8-9aa8-4d74-927d-1f4a14a0b239/.default`. Keep audiences separate; the
guide does not identify this subset's individual backend routes.
[Application-based authentication](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication),
updated 2025-09-17. See the SDK's [routing evidence](../http/routing-evidence.md)
for the already implemented shared boundary.

## Documented Graph operations: candidates, not proven cmdlet routes

The following contracts are independently documented Graph behavior. They are
useful candidate native-Go implementations, but their existence does not prove
that MicrosoftTeams 8.1.0 used them or that composing them reproduces the cmdlet.
Routes below are relative to commercial `https://graph.microsoft.com/v1.0`.
All listed operations require Graph bearer authorization; personal Microsoft
accounts are unsupported for these team/group operations. Permission entries
below distinguish delegated work/school (D) and application (A). Resource-specific
consent (RSC) requires a grant for that team, not tenant-wide permission.

| Capability | Documented request and success | Authorization and behavior | Source / update date |
| --- | --- | --- | --- |
| Read team | GET /teams/{team-id}; 200 team object; no body; $select/$expand | D Team.ReadBasic.All; A TeamSettings.Read.Group (RSC), with broader alternatives. Teams service admins can access teams without membership. | [Get team](https://learn.microsoft.com/en-us/graph/api/team-get?view=graph-rest-1.0), 2025-07-23 |
| List organization teams | GET /teams; 200 collection in value; $filter/$select/$top/$skiptoken/$count | D/A Team.ReadBasic.All. Only id, displayName, description, visibility populated; other team properties null. No documented default page size here. | [List teams](https://learn.microsoft.com/en-us/graph/api/teams-list?view=graph-rest-1.0), 2025-07-23 |
| List user teams | GET /users/{id-or-UPN}/joinedTeams; 200 collection; no OData customization | D/A Team.ReadBasic.All. Direct memberships only; excludes host teams accessed solely through a shared channel. Only id/displayName/description/isArchived/tenantId populated. | [List joinedTeams](https://learn.microsoft.com/en-us/graph/api/user-list-joinedteams?view=graph-rest-1.0), 2025-12-16 |
| Create team | POST /teams with JSON team and template binding; 202 with operation link | D/A Team.Create. Omitted fields can inherit template defaults. This does not establish PowerShell group creation, cleanup, or result projection. | [Create team](https://learn.microsoft.com/en-us/graph/api/team-post?view=graph-rest-1.0), 2025-07-23 |
| Teamify existing group | PUT /groups/{id}/team with JSON team; 201 team object | D/A Group.ReadWrite.All. Group needs an owner. Groups younger than 15 minutes can produce replication-related 404; docs recommend up to three retries separated by 10 seconds, then waiting until 15 minutes old. SharePoint provisioning remains background work. | [Create team from group](https://learn.microsoft.com/en-us/graph/api/team-put-teams?view=graph-rest-1.0), 2025-07-31 |
| Update team | PATCH /teams/{team-id} with JSON team settings; 204 | D TeamSettings.ReadWrite.All; A TeamSettings.ReadWrite.Group (RSC), broader alternatives available. Service-admin access supported. | [Update team](https://learn.microsoft.com/en-us/graph/api/team-update?view=graph-rest-1.0), 2024-09-16 |
| Delete team/group | DELETE /groups/{id}; no body; 204 | D/A Group.ReadWrite.All. Microsoft 365 groups are recoverable for 30 days. Extra rules apply to role-assignable groups. | [Delete group](https://learn.microsoft.com/en-us/graph/api/group-delete?view=graph-rest-1.0), 2024-04-04 |
| Archive | POST /teams/{id}/archive; optional JSON shouldSetSpoSiteReadOnlyForMembers; 202 + Location | D TeamSettings.ReadWrite.All; A TeamSettings.ReadWrite.Group (RSC), broader alternatives available. Team/group must have an owner. SharePoint option unsupported in application context; false/omission skips that step. | [Archive team](https://learn.microsoft.com/en-us/graph/api/team-archive?view=graph-rest-1.0), 2024-09-16 |
| Unarchive | POST /teams/{id}/unarchive; no body; 202 + Location | D TeamSettings.ReadWrite.All; A TeamSettings.ReadWrite.Group (RSC), broader alternatives available. Success acceptance precedes completion. | [Unarchive team](https://learn.microsoft.com/en-us/graph/api/team-unarchive?view=graph-rest-1.0), 2025-07-23 |
| Update picture | PUT /teams/{id}/photo/$value with binary image/jpeg; team example returns 204 | D/A TeamSettings.ReadWrite.All. Photo limit 4 MB. Group-photo endpoint also exists with different permission choices; do not substitute automatically. | [Update profilePhoto](https://learn.microsoft.com/en-us/graph/api/profilephoto-update?view=graph-rest-1.0), 2025-07-23 |
| List team members | GET /teams/{team-id}/members; 200 conversationMember collection; $filter/$select/$top | D TeamMember.Read.All; A TeamMember.Read.Group (RSC). Default/max page sizes 100/999. Membership IDs opaque; members may belong to different tenants. | [List members](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0), 2024-04-19 |
| Add member | POST /teams/{team-id}/members with conversationMember JSON; 201 conversationMember | D/A TeamMember.ReadWriteNonOwnerRole.All or TeamMember.ReadWrite.All. Basic member roles is empty; owner qualifier is owner. App-only guest addition unsupported. Disabled/blocked user can return 404. Docs advise two seconds between calls. | [Add member](https://learn.microsoft.com/en-us/graph/api/team-post-members?view=graph-rest-1.0), 2025-12-03 |
| Change role | PATCH /teams/{team-id}/members/{membership-id}; JSON roles; 200 updated conversationMember | D/A TeamMember.ReadWriteNonOwnerRole.All or TeamMember.ReadWrite.All. Roles accepts owner or empty; guest qualifier is service-managed and guests cannot become owners. | [Update member](https://learn.microsoft.com/en-us/graph/api/team-update-members?view=graph-rest-1.0), 2025-07-23 |
| Remove member | DELETE /teams/{team-id}/members/{membership-id}; no body; 204 | D/A TeamMember.ReadWriteNonOwnerRole.All or TeamMember.ReadWrite.All. Requires opaque membership identity, not user UPN or directory user ID. | [Remove member](https://learn.microsoft.com/en-us/graph/api/team-delete-members?view=graph-rest-1.0), 2025-07-23 |

The Graph team model explicitly equates team and backing-group IDs and links team
deletion to group deletion. Its nested memberSettings, guestSettings,
messagingSettings, and funSettings offer candidate projections for the flattened
PowerShell settings. Its documented visibility default is **Public**, whereas
New-Team documents Private: relying blindly on Graph defaults can change
behavior. Neither MailNickName nor PowerShell's local retention switch is a
documented team property.
[Team resource](https://learn.microsoft.com/en-us/graph/api/resources/team?view=graph-rest-1.0).

Graph membership response examples use `id`, `roles`, `displayName`, `userId`, and
`email`; these are not the PowerShell `UPN`, `UserId`, `Name`, `Role` projection.
In particular, email must not be assumed equal to UPN, and membership `id` must
not be confused with `userId`.
[List members](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0),
[Get-TeamUser](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-teamuser?view=teams-ps).

Graph's async resource uses status, operationType, attemptsCount, timestamps,
targetResourceId, targetResourceLocation, and error. Poll the returned Location;
the resource page says to wait **more than 30 seconds** between polls and treat
targetResourceLocation as opaque. Archive/unarchive examples return empty 202
bodies with relative Location values such as `/teams({id})/operations({opId})`.
This requires safe URL resolution and shared operation handling, not a fabricated
team object from a 202 response. It does not establish when PowerShell returns.
[teamsAsyncOperation](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-1.0),
updated 2024-08-19;
[Archive response](https://learn.microsoft.com/en-us/graph/api/team-archive?view=graph-rest-1.0#response),
[Unarchive response](https://learn.microsoft.com/en-us/graph/api/team-unarchive?view=graph-rest-1.0#response).

## Contract consequences and remaining evidence work

These are proposed design constraints inferred from the evidence above, not
claims of implemented or lab-observed behavior:

1. Separate team lookup/listing, new-group creation, existing-group teamification,
   update, deletion, archive state, picture upload, and membership operations.
   Preserve both New-Team parameter sets and all nine command mappings.
2. Use optional Boolean representations for settings and Archived filters so false
   survives serialization. Preserve omitted strings separately from supplied empty
   strings until clearing semantics are evidenced. Pinned `New-Team` rejects
   explicit null/empty values for the identified strings; Set-Team needs separate
   validation evidence. Do not expose a JSON-null clearing promise by inference.
3. Model user UPN, directory user ID, group/team ID, and conversation membership ID
   as distinct concepts. Separate ownership demotion from membership deletion.
4. Treat Get-Team's thread count as local scheduling. Determine whether 8.1.0
   retrieves groups then hydrates teams, and how filters combine, before fixing
   list response shape, pagination, exact matching, and permissions. Current Graph
   list output alone cannot provide the Archived/MailNickName surface.
5. Resolve New-Team app-only support, exact defaults, EDU template mapping,
   group creation steps, and compensation scope. Never apply cleanup to an
   existing supplied group. Do not retry non-idempotent creation merely because a
   transport response was lost; retain operation identity where available.
6. Resolve Set-Team's group-versus-team updates, internal Classification, and
   deprecated discovery setting against pinned package evidence. A matching Graph
   property name is insufficient to claim full command parity.
7. Resolve the picture route/media behavior from pinned implementation evidence:
   PNG/GIF cmdlet inputs versus Graph's documented image/jpeg requirement, cloud
   support, and any conversion or size enforcement. File access belongs in a local
   adapter; transport receives bytes/stream and evidenced content metadata.
8. Resolve user lookup, existing-member promotion, role precedence, backing-group
   consistency, last-owner behavior, and missing-user/idempotency responses. Keep
   guest support and owner-changing permissions explicit. The current permission
   tables do not imply NonOwnerRole permissions authorize owner changes.
9. Build only evidence-derived fixtures: metadata-binding validation, documented
   Graph example shapes labeled as such, explicit false/omission, empty successes,
   relative async locations, member-ID distinctions, paging, and documented
   propagation cases. Exact PowerShell request sequences, errors, paging tokens,
   and outputs remain unavailable from this documentation-only investigation.

The shared [HTTP notes](../http/routing-evidence.md),
[paging notes](../http/pagination-evidence.md),
[operation notes](../http/operation-evidence.md), and
[payload notes](../http/payload-evidence.md) are implementation foundations, not
evidence that these commands are complete. This subset remains blocked on the
identified wire-contract and parity questions until pinned-package inspection or
sanitized lab observations settle them.
