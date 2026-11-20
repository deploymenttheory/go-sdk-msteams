# CORE-01 pinned package inspection

Status: research in progress for [CORE-01](https://github.com/deploymenttheory/go-sdk-msteams/issues/27). No endpoint contract is accepted by this note and no coverage status changes.

## Detailed investigation

Subsequent method/serializer inspection covers all 40 assignments. These notes
separate pinned command compositions from native resource CRUD/list operations,
following the [user's architecture amendment](../plan-amendments.md).

| Assigned subset | Commands | Detailed evidence |
| --- | ---: | --- |
| Teams and team members | 9 | [Wire/control-flow evidence](team-wire-contract-evidence.md), [current documentation](teams-documentation-evidence.md) |
| Channels, channel members, sharing and operations | 14 | [Channel contracts](channel-contract-evidence.md) |
| Templates, targeting policies, batch deployment and local builders | 12 | [Template/policy/deployment contracts](template-policy-deployment-evidence.md) |
| Hierarchy upload/status/delete and migration reads | 5 | [Hierarchy/migration contracts](hierarchy-migration-contract-evidence.md) |

[Versioned serializer fixtures](../../../contracts/endpoints/README.md) now capture
21 team/channel/template/policy cases with repeatable offline oracles. The
[machine-readable inventory](../../../contracts/endpoints/8.1.0/core.json) maps
all 40 commands and their exact parameter sets to 60 wire primitives while retaining
their unresolved questions. Complete accepted endpoint contracts,
remaining fixtures and the explicit closure gates in the notes are still required;
CORE-01 remains open. The initial findings below are investigation history, and the
detailed notes settle some of the questions they identify.

The [policy routing investigation](policy-routing-evidence.md) additionally traces
same-origin forest discovery, bootstrap flight parsing and target headers; source
inspection does not establish universal modern-endpoint acceptance for unflighted
tenants.

## Sources and procedure

Inspected on 2026-11-20 with the pinned MicrosoftTeams 8.1.0 package. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. The [normalized runtime](../../../contracts/powershell/8.1.0/runtime.json) remains the parameter-set source.

Importing the local module and inspecting `Get-Command` metadata resolved all 40 assigned commands: 30 compiled cmdlets and 10 exported proxy functions. Reflection inspected 418 methods on the implementing types and their nested types. IL string/member references establish investigation leads; full control-flow and serializer inspection is still needed to settle contracts. No `Connect-MicrosoftTeams` or management command was invoked; no tenant credentials or data were used.

Source paths are relative to the pinned package:

| Source | SHA256 |
| --- | --- |
| `exports/ProxyCmdletDefinitionsWithHelp.ps1` | `5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5` |
| `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll` | `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49` |
| `netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll` | `3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24` |

## Complete command assignment

Compiled implementation names below omit the `Microsoft.Teams.PowerShell.TeamsCmdlets` namespace unless specified. Proxy entries are exact command names selected by the exported function in `Microsoft.Teams.ConfigAPI.Cmdlets.private`, except entries explicitly marked `custom` (the `Microsoft.Teams.ConfigAPI.Cmdlets.custom` module). This is not a wire route. Parameter-set counts come from the checked runtime snapshot.

| Command | Sets | Implementation / private proxy targets |
| --- | ---: | --- |
| `Add-TeamChannelUser` | 1 | `AddTeamChannelUser` |
| `Get-TeamChannelUser` | 1 | `GetTeamChannelUser` |
| `Remove-TeamChannelUser` | 1 | `RemoveTeamChannelUser` |
| `Get-TeamAllChannel` | 1 | `GetTeamAllChannel` |
| `Get-TeamChannel` | 1 | `GetTeamChannel` |
| `Get-TeamIncomingChannel` | 1 | `GetTeamIncomingChannel` |
| `New-TeamChannel` | 1 | `NewTeamChannel` |
| `Remove-TeamChannel` | 1 | `RemoveTeamChannel` |
| `Set-TeamChannel` | 1 | `SetTeamChannel` |
| `Get-CsBatchTeamsDeploymentStatus` | 2 | `Get-CsBatchTeamsDeploymentStatus_Get; Get-CsBatchTeamsDeploymentStatus_GetViaIdentity` |
| `Get-MultiGeoRegion` | 1 | `GetMultiGeoRegion` |
| `Get-TenantPrivateChannelMigrationStatus` | 1 | `GetTenantPrivateChannelMigrationStatus` |
| `New-CsBatchTeamsDeployment` | 1 | `custom: New-CsBatchTeamsDeployment` |
| `Get-Operation` | 1 | `GetOperation` |
| `Get-AssociatedTeam` | 1 | `GetAssociatedTeam` |
| `Get-SharedWithTeam` | 1 | `GetSharedWithTeam` |
| `Get-SharedWithTeamUser` | 1 | `GetSharedWithTeamUser` |
| `Remove-SharedWithTeam` | 1 | `RemoveSharedWithTeam` |
| `Get-CsTeamsTargetingPolicy` | 2 | `Get-CsTeamsTargetingPolicy` |
| `Get-TeamTargetingHierarchyStatus` | 1 | `GetTeamTargetingHierarchyStatus` |
| `Remove-CsTeamsTargetingPolicy` | 1 | `Remove-CsTeamsTargetingPolicy` |
| `Remove-TeamTargetingHierarchy` | 1 | `RemoveTeamTargetingHierarchy` |
| `Set-CsTeamsTargetingPolicy` | 1 | `Set-CsTeamsTargetingPolicy` |
| `Set-TeamTargetingHierarchy` | 1 | `SetTeamTargetingHierarchy` |
| `Add-TeamUser` | 1 | `AddTeamUser` |
| `Get-TeamUser` | 1 | `GetTeamUser` |
| `Remove-TeamUser` | 1 | `RemoveTeamUser` |
| `Get-CsTeamTemplate` | 2 | `Get-CsTeamTemplate_Get; Get-CsTeamTemplate_GetViaIdentity` |
| `Get-CsTeamTemplateList` | 1 | `custom: Get-CsTeamTemplateList` |
| `New-CsTeamsHiddenMeetingTemplate` | 1 | `Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewHiddenMeetingTemplateCmdlet` |
| `New-CsTeamsHiddenTemplate` | 1 | `Microsoft.Teams.Policy.Administration.Cmdlets.Core.NewHiddenTemplateCmdlet` |
| `New-CsTeamTemplate` | 4 | `custom: New-CsTeamTemplate` |
| `Remove-CsTeamTemplate` | 2 | `Remove-CsTeamTemplate_Delete; Remove-CsTeamTemplate_DeleteViaIdentity` |
| `Update-CsTeamTemplate` | 4 | `Update-CsTeamTemplate_Update; Update-CsTeamTemplate_UpdateExpanded; Update-CsTeamTemplate_UpdateViaIdentity; Update-CsTeamTemplate_UpdateViaIdentityExpanded` |
| `Get-Team` | 2 | `GetTeam` |
| `New-Team` | 2 | `NewTeam` |
| `Remove-Team` | 1 | `RemoveTeam` |
| `Set-Team` | 1 | `SetTeam` |
| `Set-TeamArchivedState` | 1 | `SetTeamArchivedState` |
| `Set-TeamPicture` | 1 | `SetTeamPicture` |

## Contract findings requiring explicit SDK handling

1. **Team commands compose group and team operations.** `GetTeam.ProcessRecord` contains a beta groups query filtered by `resourceProvisioningOptions`, optional `$search`, `ConsistencyLevel: eventual`, and a separate `/v1.0/users/{id}/joinedTeams` path. `GetTeamUser.ProcessRecord` reads group owners and members, selects `id,displayName,userPrincipalName,userType`, removes owner duplicates and classifies guests. `RemoveTeam.ProcessRecord` deletes through `/v1.0/groups/{id}`. A direct one-command/one-Graph-team-call mapping would lose behavior.

2. **Channel paths are explicitly beta in this baseline.** The channel, membership, associated-team and shared-channel cmdlets reference `/beta/teams/...` and `/beta/users/...` paths. Channel-user commands resolve a display name with a case-insensitive filter before addressing membership IDs. `AddTeamChannelUser.ProcessRecord` has member creation and owner promotion paths; `RemoveTeamChannelUser.ProcessRecord` distinguishes membership deletion from owner demotion. Do not silently substitute v1.0 routes before documenting equivalence.

3. **Targeting hierarchy has separate resource boundaries.** `GetTeamTargetingHierarchyStatus.ProcessRecord` references `TeamsHierarchyResource` with `/api/v1/hierarchies/` and `TeamsTargetingResource` with `/api/rest/v1/nodes/upload/status`. `SetTeamTargetingHierarchy` has `UploadCsvTargetingV1` and `UploadCsvTargetingV2`; the latter includes hierarchy lookup/creation and `/importCsv`. These are not evidenced as Config API routes. Resource origins, audiences, discovery, version selection, multipart fields and response models must be traced before extending the two existing backend profiles.

4. **Migration is not a Graph-only family.** `GetMultiGeoRegion.ExecuteMultiGeoRegionForEntity` references `TeamsMigrationResource` plus `/api/tms/prod/users/{id}/region` or `/api/tms/prod/groups/{id}/region`. `GetTenantPrivateChannelMigrationStatus.GetTemplatesServiceApiBaseUrl` selects templates API origins with commercial/GCC/GCCH/DoD/China branches and appends `/channels/privateChannelMigrationStatus`. The complete auth/cloud behavior remains unresolved.

5. **Hidden-template commands appear to be local object builders.** `NewHiddenTemplateCmdlet.CreateNew` constructs `HiddenTemplate(Id)`; `NewHiddenMeetingTemplateCmdlet.CreateNew` constructs `HiddenMeetingTemplate(Id)`. Both delegate `ProcessRecord` to `BaseTpmNewElementCmdlet<T>`; inspect that base implementation before settling the local-only disposition. No endpoint should be invented from these command names.

6. **Template/deployment/policy proxy definitions are dispatch evidence only.** Their parameter-set mappings identify private or custom compiled implementations. Routes, serializers and error contracts must be followed into those implementations; proxy naming alone does not settle the contract.

## Remaining work before CORE-01 can close

- Trace helper methods and private proxy implementations, including all endpoint/audience/cloud selectors; document every parameter set, default, validation and omitted/null distinction.
- Reconcile the pinned implementation with current official documentation, keeping documentation revisions and known differences explicit.
- Produce versioned endpoint contracts and representative sanitized fixtures, with request/response/error shapes, paging, asynchronous operations, replay/transfer behavior and local-builder dispositions.
- Resolve canonical Go operations and any additional backend/resource profiles before dependent implementations begin.
- Keep live qualification separate: this inspection performed no tenant operations.

