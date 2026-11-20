# CORE-01 hierarchy and migration contracts

Research date: 2026-11-20. Scope: Get/Set/Remove-TeamTargetingHierarchy,
Get-MultiGeoRegion and Get-TenantPrivateChannelMigrationStatus. The first name is
`Get-TeamTargetingHierarchyStatus`; there is no invented hierarchy-list command.
This is package/documentation evidence, not live qualification.

## Primary sources

The pinned 8.1.0 package and TeamsCmdlets assembly hashes are recorded in
[baseline inspection](baseline-inspection.md). Method references below are in
`Microsoft.Teams.PowerShell.TeamsCmdlets` unless stated otherwise. Endpoint maps
come from `netcoreapp3.1/Microsoft.TeamsCmdlets.PowerShell.Connect.dll`, SHA256
`c1c2b2d1c4a5306fb70f373df7ed5bfca4f9db58e90be358afd8a1d48ed46aea`, types
`Microsoft.TeamsCmdlets.PowerShell.Connect.Models.AzureEnvironment`,
`TeamsEnvironment` and `ConnectMicrosoftTeams.DetermineTeamsTargetingEndPoint`.
ILSpy 11.2.0.10050 reconstructed these methods locally; no management method or
tenant authentication was invoked. [Runtime metadata](../../../contracts/powershell/8.1.0/runtime.json)
remains authoritative for exported parameters.

## Service and authentication profiles

These are Teams administration resources, with distinct backend profiles behind
the `TeamsAdminAPI` collection. They do not use the Config API origin or its
audience merely because they share a root collection.

| Profile | Commercial request origin | Audience / delegated scope evidenced in PerformAuthorization |
| --- | --- | --- |
| Targeting v1 | `https://retailservices.teams.microsoft.com` | audience `https://retailservices.teams.microsoft.com`; `/.default` scope |
| Hierarchy v2 | `https://hierarchy.teams.microsoft.com` | audience `https://chatsvcagg.teams.microsoft.com`; `/.default` scope |
| Multi-geo region | `https://teams.cloud.microsoft` | audience `https://migrationsvc.teams.microsoft.com`; `/Region.Read.All` scope |
| Private-channel migration | `https://teams.microsoft.com` | audience `https://templates.teams.microsoft.com`; `/.default` scope |

The first two are separate selectable API versions, not PowerShell-release method
names. Native providers currently support only Graph and Config API audiences;
these findings require additional exact-origin/audience profiles and a scoped
delegated-token contract before implementation. In particular, Region.Read.All
must not be silently rewritten to `.default`. Cache isolation must include any
scope distinction introduced by that extension. Application mode uses its own
evidenced eligibility; a delegated scope is not an application permission.

The pinned targeting endpoint selector supports commercial and GCC; GCC changes
the request origin to `https://retailservices.gcc.teams.microsoft.com` while the
AzureCloud targeting token resource remains the commercial retail audience. The
selector rejects GCCH, DoD, China, France and Germany. Its OIDC discovery checks
tenant_region_scope/sub_scope to select GCC; native explicit cloud selection may
provide that distinction without global mutable environment state. Never copy its
fallback into cross-cloud credential forwarding. Hierarchy v2 has commercial
constants but no equivalent sovereign map; deployment availability is unqualified.

`TeamsEnvironment.setMigrationEndPoint` supports TeamsCloud only. The
[application-authentication guide](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication)
lists Get-MultiGeoRegion as unsupported for application authentication; retain
that restriction independently from token acquisition success.

Private-channel migration uses these exact base paths before its fixed status path:

| Environment | Base URL | Token resource |
| --- | --- | --- |
| Commercial | `https://teams.microsoft.com/fabric/templates/api` | `https://templates.teams.microsoft.com` |
| GCC | `https://teams.microsoft.com/fabric/gcc/templates/api` | `https://templates.teams.microsoft.com` |
| GCCH | `https://gov.teams.microsoft.us/fabric/gov/templates/api` | `https://templates.teams.microsoft.us` |
| DoD | `https://dod.teams.microsoft.us/fabric/dod/templates/api` | `https://templates.teams.microsoft.us` |
| China | `https://teams.microsoftonline.cn/fabric/templates/api` | `https://templates.teams.microsoftonline.cn` |

Source: `GetTenantPrivateChannelMigrationStatus.GetTemplatesServiceApiBaseUrl` and
AzureEnvironment maps. Its switch has a commercial default; that does not establish
France/Germany support where the matching token resources are absent. Exact path
prefixes must remain part of the backend boundary, not a wildcard Teams host grant.

## Hierarchy CRUD/action matrix

The tenant has a singleton operational hierarchy. API evidence establishes a
create primitive for v2, upload/update actions, status reads and deletion. It does
not establish a general paginated List API. Document that limitation rather than
turning latest-operation status into a hierarchy collection.

| Version and canonical operation | Method/path | Input and output |
| --- | --- | --- |
| v1 ImportCSV (update singleton) | POST `/api/rest/v1/nodes/upload` | Multipart `csvFile`; string-to-string response map, commonly requestId |
| v1 GetOperation | GET `/api/rest/v1/nodes/upload/status` | Optional requestId query; omitted/blank means latest; RetailServiceOperation |
| v1 Delete | DELETE `/api/rest/v1/nodes/hierarchy` | No request body; success metadata only |
| v2 Get (supporting existence read) | GET `/api/v1/hierarchies/{tenantID}` | Upload composition inspects status only; full hierarchy response model remains unproven |
| v2 Create | POST `/api/v1/hierarchies` | `{"id":"<tenantID>","displayName":"Operational Hierarchy"}`; creation composition checks status only |
| v2 ImportCSV (update singleton) | POST `/api/v1/hierarchies/{tenantID}/importCsv` | Multipart `csvFile`; requires response id, projected to requestId |
| v2 GetOperation | GET `/api/v1/hierarchies/{tenantID}/operations/{id}` | Omitted/blank RequestId becomes literal latest; HierarchyServiceOperation |
| v2 Delete | DELETE `/api/v1/hierarchies/{tenantID}` | No request body; success metadata only |

Sources: `SetTeamTargetingHierarchy.ProcessRecord/UploadCsvTargetingV1/UploadCsvTargetingV2`,
`GetTeamTargetingHierarchyStatus.ProcessRecord`, `RemoveTeamTargetingHierarchy.ProcessRecord`.

All three exported commands default omitted ApiVersion to 1; explicit values are
the strings 1 or 2. FilePath is mandatory for Set. RequestId is optional for Get;
Remove exposes only ApiVersion beyond common parameters. Its ClientRequestId CLR
property is not a PowerShell parameter and is not a remote request header.

Upload sends original file bytes with part name `csvFile`, the file basename and
part Content-Type `application/json` even though the content is CSV. It does not
parse/validate CSV locally. v2 first reads the tenant hierarchy, creates it only
after 404, then imports; any other failed read stops. A successful create followed
by failed upload leaves a hierarchy, so retain that partial outcome. Do not replay
the upload or delete the singleton as automatic compensation. The module buffers
the file; the Go SDK can use the bounded multipart stream primitive.

The current [upload help](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-teamtargetinghierarchy?view=teams-ps)
describes replacement of the tenant's active hierarchy and default v1, but labels
v2 preview and contains a malformed syntax heading. The pinned exported command
settles its actual parameter names. The [hierarchy guide](https://learn.microsoft.com/en-us/microsoftteams/set-up-your-team-hierarchy)
requires a Teams service admin for upload. Role enforcement was not exercised.

## Operation/error models

RetailServiceOperation fields are `id`, `status`, `lastUpdatedAt`, `lastModifiedBy`,
`filename`, `errorDetails.lastKnownStatus` and `errorDetails.errorMessage`.
The last value is itself JSON text for an array of `{errorCode,description,args}`;
args is an array of string maps. Keep raw details available as typed data, separate
from sanitized diagnostic logging.

HierarchyServiceOperation fields are `id`, `status`, `lastActionDateTime`,
`lastModifiedBy`, `fileName`, and `error` with message and details. A detail has
code, message and string-map arguments. The PowerShell projection maps NotStarted
to Starting, Running to Ingesting, Succeeded to Successful, Failed to Failed and
UnknownFutureValue to Unknown; other unrecognized status text survives unchanged.
This is presentation normalization. Native models should preserve original status
alongside any normalized operation state. Unknown states are never inferred success.

Get is a single status read; Set returns an operation/request identity and does not
poll. These endpoints are not Graph teamsAsyncOperation routes and do not inherit
its 31-second minimum solely by sharing an operation helper. Status polling must
be explicit and bounded. Non-success responses retain HTTP status; the original
upload/delete paths expose raw text inconsistently. SDK errors use the existing
sanitized error contract. Missing v2 response id is a decode/contract error while
preserving response metadata. Source: both operation model classes and
`GetTeamTargetingHierarchyStatus.Result` constructors; supplementary
[status help](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-teamtargetinghierarchystatus?view=teams-ps).

## Migration reads

`GetMultiGeoRegion.ExecuteMultiGeoRegionForEntity` issues GET
`/api/tms/prod/users/{entityID}/region` or `/api/tms/prod/groups/{entityID}/region`.
Both EntityId and EntityType are mandatory; the enum accepts User or Group. The
caller ID is used directly without Graph lookup; the SDK must escape it as one
path segment. The response model contains Region, optional LastMovementTimeStamp
and PreferredDataLocation. PowerShell projects just Region when the latter two
are absent; the Go typed read can retain nullable fields consistently. The
[Multi-Geo guide](https://learn.microsoft.com/en-us/microsoft-365/enterprise/multi-geo-capabilities-in-teams-in-microsoft-365?view=o365-worldwide)
confirms these properties and user/group inputs. No create/update/delete/list route
is evidenced for this region observation resource; canonical methods are
`Regions.GetForUser` and `Regions.GetForGroup`.

`GetTenantPrivateChannelMigrationStatus` has no application parameters. It GETs
`/channels/privateChannelMigrationStatus` under the selected templates base path.
The pinned model has TenantId, MigrationStatus, Details and nullable LastChecked,
MigrationStartTimeStamp and MigrationCompletionTimeStamp. Details is a JSON string,
not an already-decoded object. Preserve the string and provide an optional typed
decoder for documented counts/details rather than discarding unknown additions.
Canonical `PrivateChannelMigration.Get` is a singleton status read; no mutation or
list route is evidenced. Current [help](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-tenantprivatechannelmigrationstatus?view=teams-ps)
requires tenant administrator permissions and documents status values including
RequiresAdminAttention. Exact app grants remain unqualified.

## Remaining contract work

Add endpoint manifests and representative payload/error/status fixtures, verify
hierarchy Create/Get response shapes before exposing complete CRUD models, and
extend the shared profile/token boundary with the exact audiences/scopes above.
Keep unsupported or unevidenced CRUD cells explicit. No live success, sovereign
deployment or application-mode access is inferred from the static maps.
