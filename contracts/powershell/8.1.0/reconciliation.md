# MicrosoftTeams 8.1.0 export reconciliation

This report supplements the immutable discovery seed. The package is pinned by SHA512 in
`provenance.json`; `reconciliation.json` records per-file SHA256, source locations,
original spellings, parameter-set names, dispositions and ticket ownership.

The default unauthenticated macOS observation contains 680 exports (416 functions,
264 cmdlets, zero aliases). All 730 Gallery candidates remain represented. The union
with nested manifest declarations contains **745 candidates**, including 15 additional
research candidates. No management exclusion follows from a missing export.

## Missing default-session exports

All 50 original gaps are listed below. `conditional-config-export` means the package's
`custom/CmdletConfig.json` lists only Torus and/or SelfHost, without TeamsGA.
`conditional-oce-declaration` means the root manifest groups the name under its
`#MPA OCE cmdlets` comment. These are evidence of conditional declarations, not proof
that a caller can use the command or that it is outside SDK scope.

The root loader selects an OCE assembly when `MSTeamsContextInternal=IsOCEModule`
(`MicrosoftTeams.psm1`, lines 15–40). The policy administration loader selects OCE,
preview or default manifests (`Microsoft.Teams.Policy.Administration.psm1`, lines 4–13).
PowerShell Desktop uses net472; PowerShell Core uses netcoreapp3.1 for the ordinary
assemblies. None of these conditions establishes a Windows requirement for this SDK.

| Candidate | Evidence disposition | Research / implementation owner |
|---|---|---|
| `Clear-CsCacheOperation` | conditional-config-export | EXT-01 / EXT-23 |
| `Clear-CsOCEContext` | conditional-config-export | BASE-02 / AUTH-06 |
| `Clear-CsRegionContext` | conditional-config-export | BASE-02 / AUTH-06 |
| `Get-CsAadTenant` | conditional-config-export | EXT-03 / EXT-29 |
| `Get-CsAadUser` | conditional-config-export | EXT-03 / EXT-29 |
| `Get-CsBatchOperationDefinition` | conditional-oce-declaration | POL-02 / POL-27 |
| `Get-CsBatchOperationStatus` | conditional-oce-declaration | POL-02 / POL-27 |
| `Get-CsBusinessVoiceDirectoryDiagnosticData` | conditional-config-export | EXT-02 / EXT-26 |
| `Get-CsCloudTenant` | conditional-config-export | EXT-03 / EXT-29 |
| `Get-CsCloudUser` | conditional-config-export | EXT-03 / EXT-29 |
| `Get-CsConfiguration` | conditional-oce-declaration | EXT-01 / EXT-23 |
| `Get-CsGroupPolicyAssignments` | conditional-oce-declaration | POL-02 / POL-27 |
| `Get-CsHostingProvider` | conditional-config-export | EXT-03 / EXT-29 |
| `Get-CsLoginInfo` | conditional-oce-declaration | EXT-02 / EXT-26 |
| `Get-CsMasObjectChangelog` | conditional-config-export | EXT-02 / EXT-26 |
| `Get-CsMasVersionedSchemaData` | conditional-config-export | EXT-02 / EXT-26 |
| `Get-CsMeetingMigrationTransactionHistory` | conditional-config-export | EXT-02 / EXT-25 |
| `Get-CsMoveTenantServiceInstanceTaskStatus` | conditional-config-export | EXT-02 / EXT-25 |
| `Get-CsRegionContext` | conditional-config-export | BASE-02 / AUTH-06 |
| `Get-CsUserProvHistory` | conditional-oce-declaration | EXT-02 / EXT-26 |
| `Get-CsUssUserSettings` | conditional-config-export | EXT-02 / EXT-26 |
| `Get-GPAGroupMembers` | conditional-oce-declaration | POL-02 / POL-27 |
| `Get-GPAUserMembership` | conditional-oce-declaration | POL-02 / POL-27 |
| `Get-NgtProvInstanceFailOverStatus` | conditional-oce-declaration | EXT-02 / EXT-25 |
| `Invoke-CsCustomHandlerCallBackNgtprov` | conditional-config-export | EXT-02 / EXT-26 |
| `Invoke-CsCustomHandlerNgtprov` | conditional-config-export | EXT-02 / EXT-26 |
| `Invoke-CsDirectObjectSync` | conditional-config-export | EXT-02 / EXT-26 |
| `Invoke-CsDirectoryObjectSync` | conditional-oce-declaration | EXT-02 / EXT-26 |
| `Invoke-CsGenericNgtProvCommand` | conditional-oce-declaration | EXT-02 / EXT-26 |
| `Invoke-CsMsodsSync` | conditional-config-export | EXT-02 / EXT-26 |
| `Invoke-CsRefreshGroupUsers` | conditional-oce-declaration | POL-02 / POL-27 |
| `Invoke-CsRehomeuser` | conditional-config-export | EXT-02 / EXT-26 |
| `Invoke-CsReprocessBatchOperation` | conditional-oce-declaration | POL-02 / POL-27 |
| `Invoke-CsReprocessGroupPolicyAssignment` | conditional-oce-declaration | POL-02 / POL-27 |
| `Move-CsAvsTenantPartition` | conditional-config-export | EXT-02 / EXT-25 |
| `Move-CsTenantCrossRegion` | conditional-config-export | EXT-02 / EXT-25 |
| `Move-CsTenantServiceInstance` | conditional-config-export | EXT-02 / EXT-25 |
| `Move-NgtProvInstance` | conditional-oce-declaration | EXT-02 / EXT-25 |
| `New-CsConfiguration` | conditional-oce-declaration | EXT-01 / EXT-23 |
| `New-CsSdgDeviceTaggingRequest` | conditional-config-export | EXT-01 / EXT-24 |
| `New-CsSDGDeviceTransferRequest` | conditional-config-export | EXT-01 / EXT-24 |
| `Remove-CsConfiguration` | conditional-oce-declaration | EXT-01 / EXT-23 |
| `Set-CsConfiguration` | conditional-oce-declaration | EXT-01 / EXT-23 |
| `Set-CsNotifyCache` | unresolved-export | EXT-01 / EXT-23 |
| `Set-CsOCEContext` | conditional-config-export | BASE-02 / AUTH-06 |
| `Set-CsOnlineDialInConferencingUserDefaultNumber` | conditional-config-export | VOICE-02 / VOICE-24 |
| `Set-CsPublishPolicySchemaDefaults` | conditional-oce-declaration | EXT-01 / EXT-23 |
| `Set-CsRegionContext` | conditional-config-export | BASE-02 / AUTH-06 |
| `Set-CsTenantUserBackfill` | conditional-config-export | EXT-02 / EXT-26 |
| `Set-CsUssUserSettings` | conditional-config-export | EXT-02 / EXT-26 |

`Set-CsNotifyCache` is the exact unresolved original export: it is declared in the root
manifest but absent from this runtime observation and the per-command export configuration.
Its owner remains EXT-01 / EXT-23; no REST operation has been inferred.

## Additional nested candidates

These 15 names occur in nested manifests, outside the Gallery/default-session union.
They retain BASE-02 research ownership and no implementation claim. The five Teams app
installation/subscription names are also commented `#preview` declarations in the root;
comments do not establish active exports. Each record identifies its nested source.
Family contract research must resolve public support before assigning implementation.

- `Add-TeamsAppInstallation`
- `Get-CsConversationRole`
- `Get-CsConversationRolesSetting`
- `Get-CsTeamsAppPolicyConfiguration`
- `Get-CsTenantConfiguration`
- `Get-LicenseReportForChangeNotificationSubscription`
- `Get-TeamsAppInstallation`
- `New-CsConversationRole`
- `Remove-CsConversationRole`
- `Remove-TeamsAppInstallation`
- `Set-CsConversationRole`
- `Set-CsConversationRolesSetting`
- `Set-CsTeamsAppPolicyConfiguration`
- `Set-CsTenantConfiguration`
- `Update-TeamsAppInstallation`

## Classifications and limitations

The previously approved `Invoke-CsInternalPSTelemetry` exclusion is the only exclusion.
Nine session-equivalent candidates retain AUTH-06 ownership. Seven attendant-builder
candidates retain their CALL family ownership; their local behaviour still needs the
family contract proof. Remaining observed exports retain management ownership.
Case folding reconciles four spelling differences while preserving both spellings:

- `Get-CsOnlineDialInConferencingBridge` → `Get-CsOnlineDialinConferencingBridge`
- `Get-CsOnlineDialInConferencingLanguagesSupported` → `Get-CsOnlineDialinConferencingLanguagesSupported`
- `Get-CsSDGBulkSignInRequestsSummary` → `Get-CsSdgBulkSignInRequestsSummary`
- `Invoke-CsInternalPSTelemetry` → `Invoke-CsInternalPsTelemetry`

This is metadata qualification, not contract or live qualification. Per-command packaged
help presence is recorded; absence of packaged help is not a support decision. Remote
API shapes, public-documentation status, privileged-session behaviour and local-builder
semantics remain family research tasks. No tenant connection was made.

The baseline workflow now extracts twice on Linux, macOS and Windows and preserves
platform snapshots as CI artifacts. Platform evidence must be read from successful runs;
a configured job alone is not a passed observation. Authentication-dependent exports
remain unobserved until the lab phase. No privileged context is enabled automatically.

## Reproduce

Acquire the pinned package using `tools/baseline/baseline.py`, then reconcile against its
verified extracted module directory:

```sh
python tools/baseline/reconcile.py --snapshot /tmp/snapshot.json \
  --module /tmp/MicrosoftTeams --output /tmp/reconciliation.json
python -m unittest discover -s tools/baseline/tests -v
```

The command refuses to replace different existing evidence. Ownership comes from the
approved backlog coverage map. New runtime exports are retained with BASE-02 research
ownership, so a discovery change cannot silently disappear from the union.
