# POL-01 pinned implementation inventory

Research started 2026-11-20 for [POL-01](https://github.com/deploymenttheory/go-sdk-msteams/issues/32). All 40 assigned commands resolved through offline Get-Command metadata from the pinned module. This inventory locates implementation evidence; it does not settle routes or mark contracts complete.

The [CRUD amendment](../plan-amendments.md) applies: resource create/read/update/delete/list are the primary surface, with explicit grant/assignment and other actions. PowerShell commands may compose them. No management cmdlet, credential read or tenant connection was invoked.

## Source identity

Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. [Runtime metadata](../../../contracts/powershell/8.1.0/runtime.json) owns the exact parameter sets.

| Package source | SHA256 |
| --- | --- |
| `custom/CmdletConfig.json` | `754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906` |
| `exports/ProxyCmdletDefinitionsWithHelp.ps1` | `5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5` |
| `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll` | `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49` |

## Command mapping

Compiled names omit `Microsoft.Teams.Policy.Administration.Cmdlets.Core.` below. Proxy entries name the module and selected helper; configuration mappings are client dispatch metadata, not endpoints. An absent CmdletConfig entry does not imply an unsupported command: compiled policy cmdlets have their own provider paths.

| Command | Sets | Implementing type or proxy helper |
| --- | ---: | --- |
| `Get-CsApplicationAccessPolicy` | 2 | `Get-CsApplicationAccessPolicy → Get-CsConfigurationModern {"ConfigType": "ApplicationAccessPolicy"}` |
| `Grant-CsApplicationAccessPolicy` | 3 | `Grant-CsApplicationAccessPolicy → Grant-CsTeamsPolicy {"PolicyType": "ApplicationAccessPolicy"}` |
| `New-CsApplicationAccessPolicy` | 1 | `New-CsApplicationAccessPolicy → New-CsConfigurationModern {"ConfigType": "ApplicationAccessPolicy"}` |
| `Remove-CsApplicationAccessPolicy` | 1 | `Remove-CsApplicationAccessPolicy → Remove-CsConfigurationModern {"ConfigType": "ApplicationAccessPolicy"}` |
| `Set-CsApplicationAccessPolicy` | 1 | `Set-CsApplicationAccessPolicy → Set-CsConfigurationModern {"ConfigType": "ApplicationAccessPolicy"}` |
| `Get-CsApplicationMeetingConfiguration` | 2 | `Get-CsApplicationMeetingConfiguration → Get-CsConfigurationModern {"ConfigType": "ApplicationMeetingConfiguration"}` |
| `Set-CsApplicationMeetingConfiguration` | 1 | `Set-CsApplicationMeetingConfiguration → Set-CsConfigurationModern {"ConfigType": "ApplicationMeetingConfiguration"}` |
| `Get-DirectToGroupAssignmentsMigrationStatus` | 1 | `Custom.GPA.GetDirectToGroupAssignmentsMigrationStatus` |
| `Get-GroupAssignmentRecommendationsPerPolicyName` | 1 | `Custom.GPA.GetGroupAssignmentRecommendationsPerPolicyName` |
| `Get-GroupAssignmentRecommendationsPerPolicyType` | 1 | `Custom.GPA.GetGroupAssignmentRecommendationsPerPolicyType` |
| `Get-GroupPolicyAssignmentConflict` | 1 | `Custom.GPA.GetGroupPolicyAssignmentConflict` |
| `Invoke-ClearDirectToGroupAssignmentMigration` | 1 | `Custom.GPA.ClearDirectToGroupAssignmentMigration` |
| `Invoke-StartDirectToGroupAssignmentMigration` | 1 | `Custom.GPA.StartDirectToGroupAssignmentMigration` |
| `Get-CsExternalAccessPolicy` | 2 | `GetExternalAccessPolicyCmdlet` |
| `Get-CsTeamsAcsFederationConfiguration` | 2 | `GetTeamsAcsFederationConfigurationCmdlet` |
| `Get-CsTeamsExternalAccessConfiguration` | 2 | `GetTeamsExternalAccessConfigurationCmdlet` |
| `Get-CsTenantFederationConfiguration` | 2 | `Get-CsTenantFederationConfiguration → Get-CsConfigurationModern {"ConfigType": "TenantFederationSettings"}` |
| `Grant-CsExternalAccessPolicy` | 4 | `GrantExternalAccessPolicyCmdlet` |
| `New-CsEdgeAllowAllKnownDomains` | 1 | `New-CsEdgeAllowAllKnownDomains → New-CsEdgeAllowAllKnownDomainsHelper {}` |
| `New-CsEdgeAllowList` | 1 | `New-CsEdgeAllowList → New-CsEdgeAllowListHelper {}` |
| `New-CsEdgeDomainPattern` | 1 | `New-CsEdgeDomainPattern → New-CsEdgeDomainPatternHelper {}` |
| `New-CsExternalAccessPolicy` | 1 | `NewExternalAccessPolicyCmdlet` |
| `Remove-CsExternalAccessPolicy` | 1 | `RemoveExternalAccessPolicyCmdlet` |
| `Set-CsExternalAccessPolicy` | 1 | `SetExternalAccessPolicyCmdlet` |
| `Set-CsTeamsAcsFederationConfiguration` | 1 | `SetTeamsAcsFederationConfigurationCmdlet` |
| `Set-CsTeamsExternalAccessConfiguration` | 1 | `SetTeamsExternalAccessConfigurationCmdlet` |
| `Set-CsTenantFederationConfiguration` | 1 | `Set-CsTenantFederationConfiguration → Set-CsConfigurationModern {"ConfigType": "TenantFederationSettings"}` |
| `Get-CsTeamsMeetingBrandingPolicy` | 2 | `GetTeamsMeetingBrandingPolicyCmdlet` |
| `Grant-CsTeamsMeetingBrandingPolicy` | 4 | `GrantTeamsMeetingBrandingPolicyCmdlet` |
| `New-CsTeamsMeetingBackgroundImage` | 1 | `NewMeetingBackgroundImageCmdlet` |
| `New-CsTeamsMeetingBrandingPolicy` | 1 | `NewTeamsMeetingBrandingPolicyCmdlet` |
| `New-CsTeamsMeetingBrandingTheme` | 1 | `NewMeetingBrandingThemeCmdlet` |
| `New-CsTeamsNdiAssuranceSlate` | 1 | `NewNdiAssuranceSlateCmdlet` |
| `Remove-CsTeamsMeetingBrandingPolicy` | 1 | `RemoveTeamsMeetingBrandingPolicyCmdlet` |
| `Set-CsTeamsMeetingBrandingPolicy` | 1 | `SetTeamsMeetingBrandingPolicyCmdlet` |
| `Get-CsOnlineVoicemailPolicy` | 2 | `GetOnlineVoicemailPolicyCmdlet` |
| `Grant-CsOnlineVoicemailPolicy` | 3 | `Grant-CsOnlineVoicemailPolicy → Grant-CsTeamsPolicy {"PolicyType": "OnlineVoicemailPolicy"}` |
| `New-CsOnlineVoicemailPolicy` | 1 | `NewOnlineVoicemailPolicyCmdlet` |
| `Remove-CsOnlineVoicemailPolicy` | 1 | `RemoveOnlineVoicemailPolicyCmdlet` |
| `Set-CsOnlineVoicemailPolicy` | 1 | `SetOnlineVoicemailPolicyCmdlet` |

## Investigation sequence

1. Application access policies: establish CRUD/list and grant/assignment separately, including AppIds array operators, user/global/group scope, and conditional modern dispatch.
2. Assignment migration: inspect all six GPA implementations and service client before assuming Graph or Config API routes; distinguish recommendation reads from migration start/clear actions.
3. Federation, application meeting configuration and voicemail: trace the mixed generated/compiled provider paths, identities and presence semantics.
4. Meeting branding and domain builders: inspect base classes to settle local construction versus remote persistence, and track their model serialization into consuming policies.
5. Produce versioned endpoint/response/error fixtures, source-backed permission restrictions and canonical Go mappings; preserve unknown or conditional contracts explicitly.

Shared [Config API discovery/dispatch evidence](../core/policy-routing-evidence.md) supplies a traced reference but does not prove every compiled policy type uses the same provider route. No command coverage state changes here.

