# CORE-01 team and membership wire evidence

Inspected 2026-11-20. Covers the nine team/team-member commands assigned to
CORE-28 and CORE-26. This supplements [current documentation evidence](teams-documentation-evidence.md)
with the pinned implementation; it does not claim live qualification or close CORE-01.

## Sources and method

The [baseline inspection](baseline-inspection.md) records package and assembly hashes.
All type names below are under `Microsoft.Teams.PowerShell.TeamsCmdlets` in
`netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll`, SHA256
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`.
ILSpyCmd 11.2.0.10050 reconstructed control flow into temporary files. Findings were
compared with reflected IL and the checked runtime parameter sets. Decompiled
implementation code is not included in the SDK.

Eight [serialization fixtures](../../../contracts/endpoints/8.1.0/fixtures/core-team-serialization.json)
were obtained offline from model constructors, `TeamSettings.ToTeam` and the actual
private `HttpUtilities.JsonSerializerSettings`. No management cmdlet was invoked.
These fixtures prove serialization, not that a live server accepts every value.
The [capture script](../../../tools/contracts/capture-team-serialization.ps1)
reproduced all eight parsed-JSON cases exactly. Its existing-output and wrong-assembly
guards were checked: existing content was preserved and a changed assembly was
rejected before import.

## CRUD and composition boundary

The SDK's primary methods operate on resources. The following canonical interface
follows the user's CRUD/list clarification; request/response models remain separate.
The direct Graph additions are documented API operations, distinct from the exact
PowerShell request sequences below. `GetByID` and `List` never imply the same result
shape or completeness.

| Service | CRUD/list operation | Wire contract and evidence |
| --- | --- | --- |
| GraphAPI.Teams | Create | POST `/v1.0/teams`; JSON team/template binding; 202 and operation location. [Graph create](https://learn.microsoft.com/en-us/graph/api/team-post?view=graph-rest-1.0); supporting native SDK operation, not the pinned New-Team sequence. |
| GraphAPI.Teams | GetByID | GET `/v1.0/teams/{id}`; team response. Pinned `TeamsCmdlet.GetTeamWithGroupAttributes` and [Graph read](https://learn.microsoft.com/en-us/graph/api/team-get?view=graph-rest-1.0). |
| GraphAPI.Teams | UpdateByID | PATCH `/v1.0/teams/{id}` with team settings; no returned resource required. Pinned `SetTeam.ProcessRecord` and [Graph update](https://learn.microsoft.com/en-us/graph/api/team-update?view=graph-rest-1.0). |
| GraphAPI.Teams | DeleteByID | DELETE `/v1.0/groups/{id}` deletes the backing group and team. Pinned `RemoveTeam.ProcessRecord`; document this full resource lifecycle. |
| GraphAPI.Teams | List | GET `/v1.0/teams`; paged basic team properties. [Graph list](https://learn.microsoft.com/en-us/graph/api/teams-list?view=graph-rest-1.0); supporting native SDK operation, not the enriched Get-Team sequence. |
| GraphAPI.Groups | Create/GetByID/UpdateByID/DeleteByID/List | POST/GET/PATCH/DELETE `/v1.0/groups[/{id}]`; evidenced by `NewTeam`, `SetTeam`, `RemoveTeam`, `TeamsCmdlet` and group-list helpers. This supporting service provides the group operations required by compositions; group creation and team creation are distinct. |
| GraphAPI.TeamMembers | Create/List/DeleteByID | Membership/ownership reference operations described below. Directory user ID identifies the backing-group membership; do not replace it with a Graph conversation-member ID. |
| GraphAPI.TeamMembers | GetByID/UpdateByID | Canonical read and role-change operations must be composed from the evidenced list/reference operations or a separately verified direct endpoint. Updating a role is not deleting the member. |

Additional team operations are `CreateFromGroup`, `Archive`, `Unarchive` and
`UploadPicture`; member operations include `PromoteOwner` and `DemoteOwner`.
Explicit compositions/recipes provide `New-Team` group-then-team creation,
enriched team reads/lists, combined group/team updates and UPN-to-ID resolution.
They expose partial results and created IDs, and retain the existing shared replay,
cancellation and compensation-ownership rules. A failed team update after a successful
group update is not represented as an atomic rollback.

## Shared Graph contracts

`TeamsCmdlet.PerformAuthorization` defaults to the selected environment's Graph
resource and Graph scopes. None of these nine commands explicitly selects the
Config API token. Native SDK token acquisition remains separate from permission
qualification. Current application-auth documentation excludes New-Team while its
help describes an Owner with certificate auth; the pinned code can resolve an explicit
Owner but calls `/me` if omitted. Keep that documented conflict visible rather than
claiming app-only qualification from route inspection.

`HttpUtilities` serializes JSON using camel-case property names and null omission.
Explicit false and empty strings survive. Successful empty bodies decode to no value;
Graph failure envelopes retain status and error code. Its legacy retry helpers are not
permission to replay SDK mutations: SDK requests use the accepted operation-aware
policy and preserve partial failures.

`GraphUtilities.GetUserId` uses GET `/v1.0/me?$select=id` for an implicit current user.
For a supplied non-GUID user it uses GET `/v1.0/users/{user}`; only a 404 triggers
GET `/v1.0/myorganization/users?$filter=mail eq '{user}'`, selecting the first result.
The SDK must escape paths and OData literals and report ambiguity rather than allow
identities to change destinations. A GUID input skips lookup in the callers that
explicitly accept it. Direct user reads and mail lookup are separate supporting
operations, not a reason to mix directory objects into team models.

## Get-Team

Source: `GetTeam.ProcessRecord`, `GetTeamProperties`, `GetNextTeamBatch`, and
`TeamsCmdlet.GetTeamWithGroupAttributes`.

- Identity lookup reads both `/v1.0/groups/{id}` and `/v1.0/teams/{id}`. Group fields
  overwrite displayName, description, visibility, mailNickname and classification
  in the team model; nested settings/internalId/archive state come from the team.
  This branch ignores other filters, despite their presence in the Identity set.
- Tenant listing starts at `/beta/groups` with
  `$filter=resourceProvisioningOptions/Any(x:x eq 'Team')`, `$top=999`, `$count=true`
  and `ConsistencyLevel: eventual`. DisplayName and MailNickName append `$search`
  clauses. Supplying both produces repeated `$search` keys in the pinned code;
  combined search semantics need an explicit SDK decision/fixture.
- It follows `value`/`@odata.nextLink`, then hydrates teams through
  POST `/v1.0/$batch` with GET `/teams/{id}` items. NumberOfThreads is actually a
  batch-width setting, parsed from string, default 20, validated 1..20. Group
  properties are merged into each successful team. Subrequest 429s are retried
  up to three times, using a positive Retry-After or 10 seconds; successes and
  failures remain distinct. One auth retry exists in the pinned batch path.
- User listing resolves the user, reads all `/v1.0/users/{id}/joinedTeams` pages and
  enriches each team with the paired group/team reads. DisplayName uses exact
  case-sensitive matching here; MailNickName and Visibility are case-insensitive;
  Archived is optional and preserves false. Tenant listing uses server-side search
  for the first two filters, then local Visibility/Archived filtering.
- The pinned command emits error strings/warnings beside partial results. SDK
  aggregate APIs must return typed partial results plus errors, preserving failed
  identities without logging tenant data. Do not silently turn hydration failures
  into a complete list.

## New-Team

Source: `NewTeam.CreateGroup`, `CreateTeam`, `IsGroupAvailable`,
`RetainOrRemoveGroup`, `SetGroupPropertiesBasedOnTemplate`, `TeamUtilities`.

CreateTeam and MigrateGroup are different input types. The pinned MigrateGroup binding
accepts only GroupId and common parameters; it does not accept the creation settings
advertised by some current help. New group creation resolves Owner (GUID/user lookup),
or current `/me` when absent, and POSTs `/v1.0/groups` with Unified group type,
mailEnabled true, securityEnabled false, owner and member bindings to that user,
display name and optional description/classification. Visibility defaults Private.

An explicit MailNickName is used directly. Omission generates `msteams_` plus six
GUID hex characters and checks a Unified-group mailNickname query. The helper's
general alias-normalization/suffix retry logic is not a license to mutate explicitly
supplied nicknames. Ordinary groups use creationOptions `ExchangeProvisioningFlags:3552`.
EDU_Class rejects explicitly supplied nonblank visibility, forces HiddenMembership,
uses creationOptions `ExchangeProvisioningFlags:461` and `classAssignments`, the
education extension value `Section`, and five evidenced resourceBehaviorOptions.
EDU_PLC uses creationOptions `PLC` and two resourceBehaviorOptions. Those full
template constants still need their own contract fixture before implementation.

Both paths probe GET `/v1.0/groups/{id}` then PUT `/v1.0/groups/{id}/team` with nested
settings. Each stage attempts at most five times for 404 propagation, separated by
20 seconds. Other failures do not establish propagation. The SDK must bound this
sequence by context and retain the group ID if local waiting ends.

Creation returns enriched group/team settings. The pinned new-group path deletes its
new group when teamification returns no team unless RetainCreatedGroup is true; the
existing-group path never deletes the supplied group. The original helper fails to
verify cleanup's HTTP status. SDK compensation must report its result and must not
delete on an ambiguous remote mutation outcome without establishing ownership/state.

## Set-Team and settings serialization

Source: `SetTeam.ProcessRecord`, `Model.TeamSettings.ToTeam`, `Model.Team`,
`Model.TeamFunSettings`, and `HttpUtilities`.

Set-Team first PATCHes group displayName, description, classification, visibility and
mailNickname, then PATCHes the team settings and reads both resources back. There is
no cross-request transaction. Even omitted settings generate five empty containers:
memberSettings, guestSettings, messagingSettings, funSettings and discoverySettings.
The fixture records explicit false and empty string behavior; nulls are omitted,
so an explicit JSON-null clearing contract is not established by these cmdlets.

The flattened setting-to-container mapping is complete in `TeamSettings.ToTeam`:
six member flags, two guest flags, five messaging flags, three fun flags plus
GiphyContentRating, and ShowInTeamsSearchAndSuggestions in discoverySettings.
The runtime metadata and companion note list every flag. Response projection is
the inverse plus GroupId, InternalId, DisplayName, Description, Visibility,
MailNickName, Classification and Archived. Preserve unknown response fields without
panicking when optional nested settings are missing.

**Observed discrepancy:** the actual offline serializer emits numeric `1` for
GiphyContentRating moderate (`strict=0`, `unknownFutureValue=2`). The property uses
ItemConverterType rather than a scalar converter. Current [Graph documentation](https://learn.microsoft.com/en-us/graph/api/resources/teamfunsettings?view=graph-rest-1.0)
specifies a string enum. `unknownFutureValue` input is omitted by the cmdlet; other
unrecognized strings fail Enum.Parse. Preserve the numeric observation as evidence;
do not silently rewrite the fixture to match documentation. The native CRUD model
should follow the documented Graph enum, with the difference explicitly tested and
recorded in the PowerShell composition mapping. Live acceptance remains unobserved.

## Delete, archive and picture

`RemoveTeam.ProcessRecord` calls DELETE `/v1.0/groups/{id}` and emits no resource.
`SetTeamArchivedState` validates that the SharePoint flag is absent when unarchiving,
reads current enriched state and rejects an already-matching state. Archive POSTs
`/v1.0/teams/{id}/archive`; the helper serializes JSON null when the SharePoint flag
is omitted and an object when supplied. Unarchive POSTs `/unarchive` with a JSON
empty string. It immediately reads state back instead of polling the returned
operation. These body literals are captured in fixtures. Native Archive/Unarchive
use the documented body contracts and return an operation handle; a composition can
wait and read back without reporting acceptance as completion. See the existing
[operation evidence](../http/operation-evidence.md).

`SetTeamPicture` reads original file bytes without image conversion and PUTs
`/v1.0/groups/{id}/photo/$value` with `Content-Type: image/jpeg`. It emits no result
resource and contains no explicit size/extension check. Native UploadPicture accepts
a bounded stream with its documented ownership, and a local file helper may wrap
it. PNG/GIF help examples do not prove that the bytes are converted. Group-photo
permission/cloud qualification is separate from team-photo routes.

## Membership operations

Source: `AddTeamUser.AddUserToUnifiedGroup`, `GetTeamUser.ProcessRecord`/`Transform`,
`RemoveTeamUser.ProcessRecord`, and `GraphUtilities`.

- Create membership POSTs `/v1.0/groups/{group}/members/$ref` with `@odata.id` set to
  the Graph `/v1.0/directoryObjects/{user}` URL. Owner creation additionally POSTs
  `/owners/$ref`. The pinned preflight GET filters members/owners by ID but treats
  any 2xx response as existence without inspecting `value`. This does not prove a
  nonempty membership; native CRUD must inspect the typed collection. The [Graph
  collection contract](https://learn.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0)
  and [add-member contract](https://learn.microsoft.com/en-us/graph/api/group-post-members?view=graph-rest-1.0)
  govern those supporting operations. Do not suppress unrelated 4xx errors by
  treating every failed preflight as permission to mutate.
- List reads all group owners and members with
  `$select=id,displayName,userPrincipalName,userType`. Guest-only skips owners;
  Owner-only skips members. Otherwise owners are removed from the member list;
  remaining guest userTypes become guests, all others members. The flattened output
  is UserId=id, User=userPrincipalName, Name=displayName, Role=owner/member/guest.
  User is the CLR property; presentation UPN is not a new wire field.
- Remove/role change lists owners first. Removing the last owner fails locally
  (pinned synthetic status 412). Existing owners are deleted through
  `/owners/{user}/$ref`. Role Owner stops there, leaving membership; Member or
  omission then deletes `/members/{user}/$ref`. No resource is returned. Demotion
  and full deletion remain distinct public operations with partial-failure evidence.

## Closure gates

Before CORE-01 closes, integrate these findings into versioned CRUD/action contracts,
complete EDU fixtures and supporting Graph group/membership permissions, and resolve
the documented search/serialization compatibility decisions. Record each field's
presence semantics and each command composition in the ledger mapping. No lab
permission, mutation, app-only or sovereign-cloud support is inferred from these
offline observations.
