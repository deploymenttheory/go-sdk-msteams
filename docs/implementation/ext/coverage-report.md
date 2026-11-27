# EXT command and parameter-set evidence inventory

Owner: EXT-90 (#191). Baseline: MicrosoftTeams 8.1.0.

The ledger assigns 111 commands, 108 default-session observed sets and 27 separately declared conditional sets to 17 EXT families.
Unowned nested candidates retaining EXT research ownership: 0.
Regenerate with `python3 tools/coverage/ext_report.py`; `--check` rejects stale output.

## Evidence boundaries

These are declared symbol/test mappings and recorded ledger states, not proof that a test ran
or that every behavior is complete. Conditional commands with no observed parameter sets remain
visible; a missing runtime signature is not an empty supported surface. No ledger state is promoted.

The [integrated suite](../../../tests/integration/lab/ext_integration/contracts_test.go) constrains
generic/typed configuration identity and presence, independent recovery after typed authorization
failure, tag-template identity handed to an agent configuration, and independent template
cleanup after agent authorization failure. These compositions use
synthetic replies. The separate opt-in `TestAccGenericTypedConfigurationComposition` creates an
owned unassigned feedback policy, crosses the generic/typed adapters, requires three pinned
comparisons, and independently cleans its one journaled resource. The second opt-in
`TestAccAgentTargetTemplateComposition` links an owned target template to an owned agent,
verifies the reference after replacement with two pinned comparisons, and cleans both.
Neither scenario assigns users/accounts or enables a live call path.
The [sanitized scoped CI evidence](evidence/integrated-1ed243c.json) records both passes
at merge revision `ade85818e2968e9a3ee2c2550e463511d17e0828`, source head
`1ed243c75ec239560f94456f64506d07ba3c26c8`: five pinned comparisons and three
independently cleaned entries, zero unresolved. No response fixtures were retained;
the initial offline replies remain synthetic. A later [captured run](evidence/integrated-25ad385.json)
at merge `b7ab5738454c13ce82666337faf883cf8fbc49c6` / head `25ad385` again passed
five comparisons with three cleanups and retained seven digest-verified response files.
The captured replay cases consume those unchanged sanitized bodies with controlled request
surrogates; they do not reconstruct private requests or qualify unobserved installer/settings calls.
Additional authored Shifts connection → instance → async-operation cases check distinct
returned identities, per-detail failures and independent parent cleanup after sync denial.
They do not qualify live WFM access or treat 202 acceptance as successful persistence.

## Live prerequisites and scope blockers

- Fresh [Graph app-token claims](evidence/graph-permission-claims.json) show the listed broad
  AppCatalog/Sites/Files roles absent from the bound CI identity; a fresh read-only catalog
  request returned HTTP 403. This denial does not qualify catalog success or establish
  every scoped permission alternative for Sites/Files.
- Unified apps and meeting knowledge need delegated credentials and exact service scopes.
- Shifts success needs service-authorized credentials, a WFM system and a mapped team.
- Device success needs real device inventory and delegated gateway access.
- Support/provisioning success needs support credentials; irreversible tenant moves, licensing
  changes and service-wide actions are not made safe by a synthetic success fixture.
- Conditional nested declarations are not normal root exports. Implementation ownership
  does not establish public availability, tenant success or approved exclusion.

## Conditional nested declarations

Read-only reflection against the verified pinned package found a compiled type for every one
of the fifteen audited nested candidates and none exported by a normal macOS root-module import. The
[assembly hashes, types and parameter declarations](evidence/nested-command-types.json)
are source observations only; they do not settle public/support entitlement or wire behavior.
The hashed manifest declarations make the conditional boundary explicit: five installation/license
commands are commented `#preview` in the root manifest; four conversation-role commands
are exported by preview/OCE manifests, and six remaining settings commands by OCE only.
All fifteen commands now have in-progress implementation ownership across app_installation,
tenant/app configuration, conversation roles/settings and the license report. Separate [compiled declaration metadata](evidence/conditional-command-metadata.json)
now captures all fifteen types and twenty-seven declared sets without invoking them; eight
TeamScope/UserScope sets map to the installer, six sets to tenant/app settings, and twelve to
conversation reads and explicit inspectable compatibility projections, plus one Period set
for bounded native license composition and a separate pinned projection. Full permission writes
remain unresolved: actual local builders retain flags, while the observed management projection
drops them. No guessed positive flag-persistence contract is added. This does not establish
root-export availability or a live pass, and does not authorize
excluding declarations or assuming the ordinary lab identity can exercise them.
Three additional [compiled local permission-builder declarations](evidence/conversation-builder-metadata.json)
each have one supplemental Identity set; they are absent from root/OCE/preview cmdlet export
lists and root discovery. They remain separately classified scope candidates, not silently
counted as accepted exports, owned commands, supported public cmdlets or approved exclusions.
The license-report composition and documented delegated-only license stage are detailed in [research notes](license-report-contract-notes.md).


## ai_agents

Contract and live limitations: [ai-agents-contract-notes.md](ai-agents-contract-notes.md).

### Get-AIGeneratedKnowledgeContainer

Owner: `EXT-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListKnowledgeContainers`](../../../msteams/teams_admin_api/ai_agents/service.go) | [`TestUnitAIGeneratedKnowledgeUsesMCPSProfile`](../../../msteams/teams_admin_api/ai_agents/service_test.go)<br>[`TestUnitAIGeneratedKnowledgeFailsLocallyOffProfile`](../../../msteams/teams_admin_api/ai_agents/service_test.go) | [`TestAccAIAgentRejectedWrites`](../../../tests/integration/lab/ai_agents/ai_agents_test.go) |

Recorded live limitation: MCPS rejects app-only tokens with HTTP 401 (the pinned module fails before sending in app-only sessions); only fabricated-thread rejection probes ran, so successful knowledge requests need delegated credentials and are unqualified.

### Get-CsAgent

Owner: `EXT-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.List`](../../../msteams/teams_admin_api/ai_agents/service.go)<br>[`Service.GetByID`](../../../msteams/teams_admin_api/ai_agents/service.go) | [`TestUnitAIAgentLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go)<br>[`TestUnitAIAgentReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go)<br>[`TestUnitAIAgentListOptionsAndCatalogFilters`](../../../msteams/teams_admin_api/ai_agents/service_test.go) | [`TestAccAIAgentReads`](../../../tests/integration/lab/ai_agents/ai_agents_test.go)<br>[`TestAccAIAgentLifecycle`](../../../tests/integration/lab/ai_agents/ai_agents_test.go) |

Recorded live limitation: Client-secret lab run created, read, listed, renamed and deleted one owned journaled agent configuration with pinned PowerShell parity, and proved rejections for unknown types, tag templates and configurations.

### Get-CsAiAgents

Owner: `EXT-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.QueryPlatformAgents`](../../../msteams/teams_admin_api/ai_agents/service.go) | [`TestUnitAIAgentReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go)<br>[`TestUnitAIAgentListOptionsAndCatalogFilters`](../../../msteams/teams_admin_api/ai_agents/service_test.go) | [`TestAccAIAgentReads`](../../../tests/integration/lab/ai_agents/ai_agents_test.go) |

Recorded live limitation: Client-secret lab run proved the catalog route through an empty unknown-provider page and a non-GUID 400 with PowerShell count parity; no published catalog entries exist in the lab, so populated pages are unqualified.

### New-CsAgent

Owner: `EXT-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Create`](../../../msteams/teams_admin_api/ai_agents/service.go) | [`TestUnitAIAgentLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go)<br>[`TestUnitAIAgentRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go) | [`TestAccAIAgentLifecycle`](../../../tests/integration/lab/ai_agents/ai_agents_test.go)<br>[`TestAccAIAgentRejectedWrites`](../../../tests/integration/lab/ai_agents/ai_agents_test.go) |

Recorded live limitation: Client-secret lab run created, read, listed, renamed and deleted one owned journaled agent configuration with pinned PowerShell parity, and proved rejections for unknown types, tag templates and configurations.

### Remove-AIGeneratedKnowledge

Owner: `EXT-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteKnowledge`](../../../msteams/teams_admin_api/ai_agents/service.go) | [`TestUnitAIGeneratedKnowledgeUsesMCPSProfile`](../../../msteams/teams_admin_api/ai_agents/service_test.go)<br>[`TestUnitAIGeneratedKnowledgeFailsLocallyOffProfile`](../../../msteams/teams_admin_api/ai_agents/service_test.go) | [`TestAccAIAgentRejectedWrites`](../../../tests/integration/lab/ai_agents/ai_agents_test.go) |

Recorded live limitation: MCPS rejects app-only tokens with HTTP 401 (the pinned module fails before sending in app-only sessions); only fabricated-thread rejection probes ran, so successful knowledge requests need delegated credentials and are unqualified.

### Remove-CsAgent

Owner: `EXT-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteByID`](../../../msteams/teams_admin_api/ai_agents/service.go) | [`TestUnitAIAgentLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go)<br>[`TestUnitAIAgentRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go) | [`TestAccAIAgentLifecycle`](../../../tests/integration/lab/ai_agents/ai_agents_test.go)<br>[`TestAccAIAgentRejectedWrites`](../../../tests/integration/lab/ai_agents/ai_agents_test.go) |

Recorded live limitation: Client-secret lab run created, read, listed, renamed and deleted one owned journaled agent configuration with pinned PowerShell parity, and proved rejections for unknown types, tag templates and configurations.

### Set-CsAgent

Owner: `EXT-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateByID`](../../../msteams/teams_admin_api/ai_agents/service.go) | [`TestUnitAIAgentLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go)<br>[`TestUnitAIAgentRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/ai_agents/service_test.go) | [`TestAccAIAgentLifecycle`](../../../tests/integration/lab/ai_agents/ai_agents_test.go)<br>[`TestAccAIAgentRejectedWrites`](../../../tests/integration/lab/ai_agents/ai_agents_test.go) |

Recorded live limitation: Client-secret lab run created, read, listed, renamed and deleted one owned journaled agent configuration with pinned PowerShell parity, and proved rejections for unknown types, tag templates and configurations.

## app_catalog

Contract and live limitations: [app-catalog-contract-notes.md](app-catalog-contract-notes.md).

### Get-CsTeamsSettingsCustomApp

Owner: `EXT-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Get`](../../../msteams/teams_admin_api/custom_app_settings/service.go) | [`TestUnitCustomAppSettingsGetDecodesTheLabShape`](../../../msteams/teams_admin_api/custom_app_settings/service_test.go)<br>[`TestUnitCustomAppSettingsProjectionReportsNullAsFalse`](../../../msteams/teams_admin_api/custom_app_settings/service_test.go)<br>[`TestUnitCustomAppSettingsFailuresStayTyped`](../../../msteams/teams_admin_api/custom_app_settings/service_test.go) | [`TestAccAppCatalogReads`](../../../tests/integration/lab/custom_app_settings/app_catalog_test.go) |

Recorded live limitation: Client-secret lab run read the settings with pinned PowerShell sideloading-flag parity; full release qualification is separate.

### Get-TeamsApp

Owner: `EXT-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.List`](../../../msteams/graph_api/app_catalog/service.go)<br>[`Service.NewPager`](../../../msteams/graph_api/app_catalog/service.go) | [`TestUnitAppCatalogListBuildsTheModuleFilterAndFollowsPages`](../../../msteams/graph_api/app_catalog/service_test.go)<br>[`TestUnitAppCatalogListDenialUsesLabResponse`](../../../msteams/graph_api/app_catalog/service_test.go) | [`TestAccAppCatalogReads`](../../../tests/integration/lab/custom_app_settings/app_catalog_test.go) |

Recorded live limitation: Client-secret lab run confirmed the app-only HTTP 403 with pinned PowerShell parity. The lab identity lacks AppCatalog permissions, so a successful list needs separately authorized credentials.

### Get-TeamsArtifacts

Owner: `EXT-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.List`](../../../msteams/graph_api/teams_artifact/service.go) | [`TestUnitTeamsArtifactsScanFollowsTheModuleRules`](../../../msteams/graph_api/teams_artifact/service_test.go)<br>[`TestUnitTeamsArtifactsSharePointDenialUsesLabResponse`](../../../msteams/graph_api/teams_artifact/service_test.go)<br>[`TestUnitTeamsArtifactsSelectorsAndPaging`](../../../msteams/graph_api/teams_artifact/service_test.go)<br>[`TestUnitTeamsArtifactsRejectsInvalidOptions`](../../../msteams/graph_api/teams_artifact/service_test.go) | [`TestAccAppCatalogReads`](../../../tests/integration/lab/custom_app_settings/app_catalog_test.go) |

Recorded live limitation: Client-secret lab run confirmed the SharePoint site search HTTP 403 with pinned PowerShell parity (no artifacts, one error); scanning real drives needs Sites.Read.All and Files.Read.All; full release qualification is separate.

### New-TeamsApp

Owner: `EXT-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Create`](../../../msteams/graph_api/app_catalog/service.go) | [`TestUnitAppCatalogPackageWrites`](../../../msteams/graph_api/app_catalog/service_test.go)<br>[`TestUnitAppCatalogFailuresStayTyped`](../../../msteams/graph_api/app_catalog/service_test.go) | [`TestAccAppCatalogWrites`](../../../tests/integration/lab/custom_app_settings/app_catalog_test.go) |

Recorded live limitation: Client-secret lab run confirmed the app-only HTTP 403 for an empty package. The lab identity lacks AppCatalog permissions, so a successful create needs separately authorized credentials.

### Remove-TeamsApp

Owner: `EXT-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteByID`](../../../msteams/graph_api/app_catalog/service.go) | [`TestUnitAppCatalogPackageWrites`](../../../msteams/graph_api/app_catalog/service_test.go)<br>[`TestUnitAppCatalogFailuresStayTyped`](../../../msteams/graph_api/app_catalog/service_test.go) | [`TestAccAppCatalogWrites`](../../../tests/integration/lab/custom_app_settings/app_catalog_test.go) |

Recorded live limitation: Client-secret lab run confirmed the app-only HTTP 403 against a reserved app ID. The lab identity lacks AppCatalog permissions, so a successful delete needs separately authorized credentials.

### Set-CsTeamsSettingsCustomApp

Owner: `EXT-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.SetSideloadedAppsInteraction`](../../../msteams/teams_admin_api/custom_app_settings/service.go)<br>[`Service.Update`](../../../msteams/teams_admin_api/custom_app_settings/service.go) | [`TestUnitCustomAppSettingsSetUsesTheModuleReadReplaceFlow`](../../../msteams/teams_admin_api/custom_app_settings/service_test.go)<br>[`TestUnitCustomAppSettingsUpdateSendsTheGivenDocument`](../../../msteams/teams_admin_api/custom_app_settings/service_test.go)<br>[`TestUnitCustomAppSettingsFailuresStayTyped`](../../../msteams/teams_admin_api/custom_app_settings/service_test.go) | [`TestAccAppCatalogWrites`](../../../tests/integration/lab/custom_app_settings/app_catalog_test.go) |

Recorded live limitation: Client-secret lab run wrote the current sideloading flag back (HTTP 204, settings unchanged); changing the flag to a different value is not exercised; full release qualification is separate.

### Set-TeamsApp

Owner: `EXT-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateByID`](../../../msteams/graph_api/app_catalog/service.go) | [`TestUnitAppCatalogPackageWrites`](../../../msteams/graph_api/app_catalog/service_test.go)<br>[`TestUnitAppCatalogFailuresStayTyped`](../../../msteams/graph_api/app_catalog/service_test.go) | [`TestAccAppCatalogWrites`](../../../tests/integration/lab/custom_app_settings/app_catalog_test.go) |

Recorded live limitation: Client-secret lab run confirmed the app-only HTTP 403 against a reserved app ID. The lab identity lacks AppCatalog permissions, so a successful update needs separately authorized credentials.

## app_installation

Contract and live limitations: [app-installation-contract-notes.md](app-installation-contract-notes.md).

### Add-TeamsAppInstallation

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `TeamScope` (compiled declaration) | [`Service.Create`](../../../msteams/graph_api/app_installation/service.go) | [`TestUnitTeamAndUserInstallationLifecycle`](../../../msteams/graph_api/app_installation/service_test.go)<br>[`TestUnitCatalogResolutionRequiresCompleteUniqueInstallation`](../../../msteams/graph_api/app_installation/service_test.go) | None recorded |
| `UserScope` (compiled declaration) | [`Service.Create`](../../../msteams/graph_api/app_installation/service.go) | [`TestUnitTeamAndUserInstallationLifecycle`](../../../msteams/graph_api/app_installation/service_test.go)<br>[`TestUnitCatalogResolutionRequiresCompleteUniqueInstallation`](../../../msteams/graph_api/app_installation/service_test.go) | None recorded |

Recorded live limitation: Compiled declarations expose TeamScope/UserScope, separately from the immutable default-session snapshot. No live installation attempted: requires an approved owned catalog app and authorized team/user scope. Native consent live acceptance and legacy cmdlet permission equivalence remain unqualified; fresh catalog read was HTTP 403.

### Get-TeamsAppInstallation

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `TeamScope` (compiled declaration) | [`Service.List`](../../../msteams/graph_api/app_installation/service.go)<br>[`Service.GetByID`](../../../msteams/graph_api/app_installation/service.go) | [`TestUnitTeamAndUserInstallationLifecycle`](../../../msteams/graph_api/app_installation/service_test.go)<br>[`TestUnitCatalogResolutionRequiresCompleteUniqueInstallation`](../../../msteams/graph_api/app_installation/service_test.go) | None recorded |
| `UserScope` (compiled declaration) | [`Service.List`](../../../msteams/graph_api/app_installation/service.go)<br>[`Service.GetByID`](../../../msteams/graph_api/app_installation/service.go) | [`TestUnitTeamAndUserInstallationLifecycle`](../../../msteams/graph_api/app_installation/service_test.go)<br>[`TestUnitCatalogResolutionRequiresCompleteUniqueInstallation`](../../../msteams/graph_api/app_installation/service_test.go) | None recorded |

Recorded live limitation: Compiled declarations expose TeamScope/UserScope, separately from the immutable default-session snapshot. No live installation attempted: requires an approved owned catalog app and authorized team/user scope. Native consent live acceptance and legacy cmdlet permission equivalence remain unqualified; fresh catalog read was HTTP 403.

### Remove-TeamsAppInstallation

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `TeamScope` (compiled declaration) | [`Service.DeleteByID`](../../../msteams/graph_api/app_installation/service.go)<br>[`Service.DeleteByCatalogID`](../../../msteams/graph_api/app_installation/service.go) | [`TestUnitTeamAndUserInstallationLifecycle`](../../../msteams/graph_api/app_installation/service_test.go)<br>[`TestUnitCatalogResolutionRequiresCompleteUniqueInstallation`](../../../msteams/graph_api/app_installation/service_test.go) | None recorded |
| `UserScope` (compiled declaration) | [`Service.DeleteByID`](../../../msteams/graph_api/app_installation/service.go)<br>[`Service.DeleteByCatalogID`](../../../msteams/graph_api/app_installation/service.go) | [`TestUnitTeamAndUserInstallationLifecycle`](../../../msteams/graph_api/app_installation/service_test.go)<br>[`TestUnitCatalogResolutionRequiresCompleteUniqueInstallation`](../../../msteams/graph_api/app_installation/service_test.go) | None recorded |

Recorded live limitation: Compiled declarations expose TeamScope/UserScope, separately from the immutable default-session snapshot. No live installation attempted: requires an approved owned catalog app and authorized team/user scope. Native consent live acceptance and legacy cmdlet permission equivalence remain unqualified; fresh catalog read was HTTP 403.

### Update-TeamsAppInstallation

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `TeamScope` (compiled declaration) | [`Service.UpgradeByID`](../../../msteams/graph_api/app_installation/service.go)<br>[`Service.UpgradeByCatalogID`](../../../msteams/graph_api/app_installation/service.go) | [`TestUnitTeamAndUserInstallationLifecycle`](../../../msteams/graph_api/app_installation/service_test.go)<br>[`TestUnitCatalogResolutionRequiresCompleteUniqueInstallation`](../../../msteams/graph_api/app_installation/service_test.go) | None recorded |
| `UserScope` (compiled declaration) | [`Service.UpgradeByID`](../../../msteams/graph_api/app_installation/service.go)<br>[`Service.UpgradeByCatalogID`](../../../msteams/graph_api/app_installation/service.go) | [`TestUnitTeamAndUserInstallationLifecycle`](../../../msteams/graph_api/app_installation/service_test.go)<br>[`TestUnitCatalogResolutionRequiresCompleteUniqueInstallation`](../../../msteams/graph_api/app_installation/service_test.go) | None recorded |

Recorded live limitation: Compiled declarations expose TeamScope/UserScope, separately from the immutable default-session snapshot. No live installation attempted: requires an approved owned catalog app and authorized team/user scope. Native consent live acceptance and legacy cmdlet permission equivalence remain unqualified; fresh catalog read was HTTP 403.

## change_notification_license_report

Contract and live limitations: [license-report-contract-notes.md](license-report-contract-notes.md).

### Get-LicenseReportForChangeNotificationSubscription

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` (compiled declaration) | [`Service.Get`](../../../msteams/graph_api/change_notification_license_report/report.go)<br>[`ProjectPinnedRows`](../../../msteams/graph_api/change_notification_license_report/pinned.go) | [`TestUnitNativeCountRequiresJSONGrammarAndBoundsOnlyComputation`](../../../msteams/graph_api/change_notification_license_report/models_internal_test.go)<br>[`TestUnitNativePlanUnknownIsNeverNegative`](../../../msteams/graph_api/change_notification_license_report/models_internal_test.go)<br>[`TestUnitNativeModelRetainsNumberWidthAndNestedPresence`](../../../msteams/graph_api/change_notification_license_report/models_internal_test.go)<br>[`TestUnitPinnedRowsKeepSourceWrapMissingNullAndDisplayFallback`](../../../msteams/graph_api/change_notification_license_report/pinned_test.go)<br>[`TestUnitPinnedNestedActualDefaultsAndOrdinalShortCircuit`](../../../msteams/graph_api/change_notification_license_report/pinned_test.go)<br>[`TestUnitPinnedCountRepresentabilityDoesNotConstrainNativeDTO`](../../../msteams/graph_api/change_notification_license_report/pinned_test.go)<br>[`TestUnitLicenseReportNativeIDCompositionAndUnboundedWireCount`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportAllPeriodValuesUseDocumentedJSONRoute`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportUnsafeContinuationAndPartialReadNeverStartsJoin`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportDeniedLicenseIsUnknownAndRetainsReceipt`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportDeclaredIDCannotFallBackToDifferentUPNOwner`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportModeAndInputBoundariesDoNotAcquireCredentials`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportTwentyItemBatchesCorrelateReorderedReplies`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportLicenseContinuationIsSubjectBoundAndPageBounded`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportPartialDirectoryCannotEstablishResolution`](../../../msteams/graph_api/change_notification_license_report/service_test.go)<br>[`TestUnitLicenseReportOpaqueDirectoryIDRemainsAuthoritativeAndSubjectBound`](../../../msteams/graph_api/change_notification_license_report/service_test.go) | None recorded |

Recorded live limitation: Conditional preview declaration; native documented reads and separate pinned local projection have offline tests only. Full composition needs delegated license/directory permissions and role; application licenseDetails is documented unsupported. No tenant success, current entitlement or default-session export is claimed.

## cloud_call_data

Contract and live limitations: [cloud-call-data-contract-notes.md](cloud-call-data-contract-notes.md).

### Get-CsCloudCallDataConnection

Owner: `EXT-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Get`](../../../msteams/teams_admin_api/cloud_call_data/service.go) | [`TestUnitCloudCallDataGetDenialUsesAcceptanceResponse`](../../../msteams/teams_admin_api/cloud_call_data/service_test.go)<br>[`TestUnitCloudCallDataReturnsTokenWithoutPrintingIt`](../../../msteams/teams_admin_api/cloud_call_data/service_test.go) | [`TestAccCloudCallDataConnectionAuthorizationDenied`](../../../tests/integration/lab/cloud_call_data/cloud_call_data_test.go) |

Recorded live limitation: Client-secret lab run confirmed the app-only HTTP 403 denial with pinned PowerShell parity, before and after the approved live create (also denied); a successful token read needs separately authorized credentials.

### New-CsCloudCallDataConnection

Owner: `EXT-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Create`](../../../msteams/teams_admin_api/cloud_call_data/service.go) | [`TestUnitCloudCallDataCreateDenialUsesAcceptanceResponse`](../../../msteams/teams_admin_api/cloud_call_data/service_test.go)<br>[`TestUnitCloudCallDataReturnsTokenWithoutPrintingIt`](../../../msteams/teams_admin_api/cloud_call_data/service_test.go)<br>[`TestUnitCloudCallDataFailuresStayTyped`](../../../msteams/teams_admin_api/cloud_call_data/service_test.go) | [`TestAccCloudCallDataConnectionCreateOutcome`](../../../tests/integration/lab/cloud_call_data/cloud_call_data_test.go) |

Recorded live limitation: Sent live with approval (tenant singleton, no remove cmdlet): the client-secret app-only Teams Administrator gets HTTP 403 Forbidden, the tenant is unchanged, and pinned PowerShell New-CsCloudCallDataConnection fails Forbidden too; a successful create and the already-exists outcome need separately authorized credentials.

## configuration_diagnostics

Contract and live limitations: [configuration-diagnostics-contract-notes.md](configuration-diagnostics-contract-notes.md).

### Clear-CsCacheOperation

Owner: `EXT-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.ClearCache`](../../../msteams/teams_admin_api/configuration_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent with inert inputs (a nonexistent key for a nonexistent region); a successful call needs support credentials and acts service-wide, so it remains unqualified. Pinned PowerShell, run in the module's own scope, fails the same way.

### Get-CsConfiguration

Owner: `EXT-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.ListConfigurations`](../../../msteams/teams_admin_api/configuration_diagnostics/service.go)<br>[`Service.GetConfiguration`](../../../msteams/teams_admin_api/configuration_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run qualified an owned TeamsFeedbackPolicy instance through the generic route (create/read-back, replace/read-back, remove/absence) and the Global/missing/unknown-type reads, with pinned PowerShell parity via the module's internal command; the module exports this command only in support sessions, and no CI test run is recorded as evidence yet.

### New-CsConfiguration

Owner: `EXT-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.CreateConfiguration`](../../../msteams/teams_admin_api/configuration_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run qualified an owned TeamsFeedbackPolicy instance through the generic route (create/read-back, replace/read-back, remove/absence) and the Global/missing/unknown-type reads, with pinned PowerShell parity via the module's internal command; the module exports this command only in support sessions, and no CI test run is recorded as evidence yet.

### Remove-CsConfiguration

Owner: `EXT-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.RemoveConfiguration`](../../../msteams/teams_admin_api/configuration_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run qualified an owned TeamsFeedbackPolicy instance through the generic route (create/read-back, replace/read-back, remove/absence) and the Global/missing/unknown-type reads, with pinned PowerShell parity via the module's internal command; the module exports this command only in support sessions, and no CI test run is recorded as evidence yet.

### Set-CsConfiguration

Owner: `EXT-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.SetConfiguration`](../../../msteams/teams_admin_api/configuration_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run qualified an owned TeamsFeedbackPolicy instance through the generic route (create/read-back, replace/read-back, remove/absence) and the Global/missing/unknown-type reads, with pinned PowerShell parity via the module's internal command; the module exports this command only in support sessions, and no CI test run is recorded as evidence yet.

### Set-CsNotifyCache

Owner: `EXT-23`. Implementation: `not-started`; contract: `metadata-only`; live: `not-run`.

Recorded approved exclusion: Approved: no evidenced wire contract in 8.1.0 (manifest-only export whose request helper has no caller): docs/implementation/ext/configuration-diagnostics-contract-notes.md

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | None recorded | No set mapping | No set mapping |

Recorded live limitation: Not implemented: the module manifest lists the command but no cmdlet or function defines it, and its only request helper has no caller, so no wire contract is evidenced.

### Set-CsPublishPolicySchemaDefaults

Owner: `EXT-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.PublishPolicySchemaDefaults`](../../../msteams/teams_admin_api/configuration_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent with inert inputs (an undefined policy type); a successful call needs support credentials and acts service-wide, so it remains unqualified. The compiled cmdlet needs an OCE login context, so it has no PowerShell parity run.

## conversation_role

Contract and live limitations: [conversation-role-contract-notes.md](conversation-role-contract-notes.md).

### Get-CsConversationRole

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.GetByIdentity`](../../../msteams/teams_admin_api/conversation_role/service.go)<br>[`Service.List`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationRoleAllGetSetsAndRawSnapshot`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |
| `ParentAndRelativeKey` (compiled declaration) | [`Service.GetByID`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationRoleAllGetSetsAndRawSnapshot`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |
| `Filter` (compiled declaration) | [`Service.ListFiltered`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationRoleAllGetSetsAndRawSnapshot`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |

Recorded live limitation: Conditional preview/OCE declaration; source-exact projection and authored reply vectors are not successful server or effective-permission proof. Nested Chat/Calling write semantics, root availability and authorization remain unqualified; no live role/Global mutation is configured.

### New-CsConversationRole

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `ParentAndRelativeKey` (compiled declaration) | [`ProjectCreateByID`](../../../msteams/teams_admin_api/conversation_role/projection.go)<br>[`Service.Create`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationCreateSetsUseInspectableFrozenObservedProjection`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |
| `Identity` (compiled declaration) | [`ProjectCreateByIdentity`](../../../msteams/teams_admin_api/conversation_role/projection.go)<br>[`Service.Create`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationCreateSetsUseInspectableFrozenObservedProjection`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |

Recorded live limitation: Conditional preview/OCE declaration; source-exact projection and authored reply vectors are not successful server or effective-permission proof. Nested Chat/Calling write semantics, root availability and authorization remain unqualified; no live role/Global mutation is configured.

### Remove-CsConversationRole

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.DeleteByIdentity`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationDeleteBothSetsUseOriginalBoundParameters`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |
| `ParentAndRelativeKey` (compiled declaration) | [`Service.DeleteByID`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationDeleteBothSetsUseOriginalBoundParameters`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |

Recorded live limitation: Conditional preview/OCE declaration; source-exact projection and authored reply vectors are not successful server or effective-permission proof. Nested Chat/Calling write semantics, root availability and authorization remain unqualified; no live role/Global mutation is configured.

### Set-CsConversationRole

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.PrepareUpdateByIdentity`](../../../msteams/teams_admin_api/conversation_role/service.go)<br>[`Service.UpdateByIdentity`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationSetBothSelectorsPrepareThenApplyFrozenBody`](../../../msteams/teams_admin_api/conversation_role/service_test.go)<br>[`TestUnitConversationSelectedPermissionSetReportsLossWithoutRejectingInput`](../../../msteams/teams_admin_api/conversation_role/service_test.go)<br>[`TestUnitConversationPreparationDenialAndAcceptedDecodeFailureRetainStages`](../../../msteams/teams_admin_api/conversation_role/service_test.go)<br>[`TestUnitConversationPreparedSelectorCannotRetargetOrChangeSet`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |
| `ParentAndRelativeKey` (compiled declaration) | [`Service.PrepareUpdateByID`](../../../msteams/teams_admin_api/conversation_role/service.go)<br>[`Service.UpdateByID`](../../../msteams/teams_admin_api/conversation_role/service.go) | [`TestUnitConversationSetBothSelectorsPrepareThenApplyFrozenBody`](../../../msteams/teams_admin_api/conversation_role/service_test.go)<br>[`TestUnitConversationRelativeSetKeepsMissingModelIDNull`](../../../msteams/teams_admin_api/conversation_role/service_test.go)<br>[`TestUnitConversationUnboundSetRejectsBeforeCredentialsAndEmptyFilterIsLocalError`](../../../msteams/teams_admin_api/conversation_role/service_test.go) | None recorded |

Recorded live limitation: Conditional preview/OCE declaration; source-exact projection and authored reply vectors are not successful server or effective-permission proof. Nested Chat/Calling write semantics, root availability and authorization remain unqualified; no live role/Global mutation is configured.

## conversation_roles_setting

Contract and live limitations: [conversation-role-contract-notes.md](conversation-role-contract-notes.md).

### Get-CsConversationRolesSetting

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.GetByIdentity`](../../../msteams/teams_admin_api/conversation_roles_setting/service.go)<br>[`Service.List`](../../../msteams/teams_admin_api/conversation_roles_setting/service.go) | [`TestUnitConversationSettingsGetSetsPreservePresenceAndLocalFilter`](../../../msteams/teams_admin_api/conversation_roles_setting/service_test.go) | None recorded |
| `Filter` (compiled declaration) | [`Service.ListFiltered`](../../../msteams/teams_admin_api/conversation_roles_setting/service.go) | [`TestUnitConversationSettingsGetSetsPreservePresenceAndLocalFilter`](../../../msteams/teams_admin_api/conversation_roles_setting/service_test.go) | None recorded |

Recorded live limitation: Conditional preview/OCE declaration; source-exact projection and authored reply vectors are not successful server or effective-permission proof. Nested Chat/Calling write semantics, root availability and authorization remain unqualified; no live role/Global mutation is configured.

### Set-CsConversationRolesSetting

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.PrepareUpdateByIdentity`](../../../msteams/teams_admin_api/conversation_roles_setting/service.go)<br>[`Service.UpdateByIdentity`](../../../msteams/teams_admin_api/conversation_roles_setting/service.go)<br>[`ProjectListValue`](../../../msteams/teams_admin_api/conversation_roles_setting/projection.go) | [`TestUnitConversationSettingsFullTransitionPipelineMatchesPinnedBodies`](../../../msteams/teams_admin_api/conversation_roles_setting/service_test.go)<br>[`TestUnitConversationSettingsNullOmittedAndEmptyResponseCollections`](../../../msteams/teams_admin_api/conversation_roles_setting/service_test.go)<br>[`TestUnitConversationSettingsDenialAcceptedMalformedAndRetargetRetainStages`](../../../msteams/teams_admin_api/conversation_roles_setting/service_test.go)<br>[`TestUnitConversationSettingsUnboundRejectsBeforeCredentialsAndEmptyFilterIsLocal`](../../../msteams/teams_admin_api/conversation_roles_setting/service_test.go) | None recorded |

Recorded live limitation: Conditional preview/OCE declaration; source-exact projection and authored reply vectors are not successful server or effective-permission proof. Nested Chat/Calling write semantics, root availability and authorization remain unqualified; no live role/Global mutation is configured.

## devices

Contract and live limitations: [devices-contract-notes.md](devices-contract-notes.md).

### Get-CsSDGBulkSignInRequestsSummary

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.ListBulkSignInRequests`](../../../msteams/teams_admin_api/sip_devices_gateway/service.go) | [`TestUnitSIPDevicesGatewayDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/sip_devices_gateway/service_test.go)<br>[`TestUnitSIPDevicesGatewaySuccessShapesPreserveFields`](../../../msteams/teams_admin_api/sip_devices_gateway/service_test.go) | [`TestAccSIPDevicesGatewayDenied`](../../../tests/integration/lab/sip_devices_gateway/sip_devices_gateway_test.go) |

Recorded live limitation: Client-secret lab run: the app-only Teams Administrator is denied (403) by every SIP devices gateway route; successful gateway operations need delegated SDG access and real devices and remain unqualified.

### Get-CsSdgBulkSignInRequestStatus

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.GetBulkSignInRequestStatus`](../../../msteams/teams_admin_api/sip_devices_gateway/service.go) | [`TestUnitSIPDevicesGatewayDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/sip_devices_gateway/service_test.go)<br>[`TestUnitSIPDevicesGatewaySuccessShapesPreserveFields`](../../../msteams/teams_admin_api/sip_devices_gateway/service_test.go) | [`TestAccSIPDevicesGatewayDenied`](../../../tests/integration/lab/sip_devices_gateway/sip_devices_gateway_test.go) |

Recorded live limitation: Client-secret lab run: the app-only Teams Administrator is denied (403) by every SIP devices gateway route; successful gateway operations need delegated SDG access and real devices and remain unqualified.

### Get-CsTeamsRemoteLogCollectionConfiguration

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Filter` | [`Service.ListFiltered`](../../../msteams/teams_admin_api/teams_remote_log_collection_configuration/service.go) | [`TestUnitConfigurationEmbeddedDevicesAndFilter`](../../../msteams/teams_admin_api/teams_remote_log_collection_configuration/service_test.go) | [`TestAccRemoteLogCollectionConfigurationReads`](../../../tests/integration/lab/teams_remote_log_collection_configuration/teams_remote_log_collection_configuration_test.go) |
| `Identity` | [`Service.GetByIdentity`](../../../msteams/teams_admin_api/teams_remote_log_collection_configuration/service.go)<br>[`Service.List`](../../../msteams/teams_admin_api/teams_remote_log_collection_configuration/service.go) | [`TestUnitConfigurationReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/teams_remote_log_collection_configuration/service_test.go) | [`TestAccRemoteLogCollectionConfigurationReads`](../../../tests/integration/lab/teams_remote_log_collection_configuration/teams_remote_log_collection_configuration_test.go) |

Recorded live limitation: Client-secret lab run passed the reads with pinned PowerShell parity against a tenant without remote log collection devices; populated reads remain unqualified.

### Get-CsTeamsRemoteLogCollectionDevice

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Filter` | [`Service.ListFiltered`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceListFilterMatchesIdentityLocally`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | None recorded |
| `Identity` | [`Service.GetByIdentity`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go)<br>[`Service.List`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | [`TestAccRemoteLogCollectionDeviceReads`](../../../tests/integration/lab/teams_remote_log_collection_device/teams_remote_log_collection_device_test.go) |
| `ParentAndRelativeKey` | [`Service.GetByID`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | [`TestAccRemoteLogCollectionDeviceReads`](../../../tests/integration/lab/teams_remote_log_collection_device/teams_remote_log_collection_device_test.go) |

Recorded live limitation: Client-secret lab run passed the reads with pinned PowerShell parity against a tenant without remote log collection devices; populated reads remain unqualified.

### New-CsSdgBulkSignInRequest

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.CreateBulkSignInRequest`](../../../msteams/teams_admin_api/sip_devices_gateway/service.go)<br>[`ReadBulkSignInItems`](../../../msteams/teams_admin_api/sip_devices_gateway/service.go) | [`TestUnitSIPDevicesGatewayDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/sip_devices_gateway/service_test.go)<br>[`TestUnitSIPDevicesGatewaySuccessShapesPreserveFields`](../../../msteams/teams_admin_api/sip_devices_gateway/service_test.go)<br>[`TestUnitReadBulkSignInItemsMatchesImportCsv`](../../../msteams/teams_admin_api/sip_devices_gateway/service_test.go) | [`TestAccSIPDevicesGatewayRejectedWrites`](../../../tests/integration/lab/sip_devices_gateway/sip_devices_gateway_test.go) |

Recorded live limitation: Client-secret lab run: the app-only Teams Administrator is denied (403) by every SIP devices gateway route; successful gateway operations need delegated SDG access and real devices and remain unqualified.

### New-CsSdgDeviceTaggingRequest

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.CreateDeviceTaggingRequest`](../../../msteams/teams_admin_api/sip_devices_gateway/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run: the app-only Teams Administrator is denied (403) by every SIP devices gateway route; successful gateway operations need delegated SDG access and real devices and remain unqualified.

### New-CsSDGDeviceTransferRequest

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.CreateDeviceTransferRequest`](../../../msteams/teams_admin_api/sip_devices_gateway/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run: the app-only Teams Administrator is denied (403) by every SIP devices gateway route; successful gateway operations need delegated SDG access and real devices and remain unqualified.

### New-CsTeamsRemoteLogCollectionDevice

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.CreateByIdentity`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceWriteShapes`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | None recorded |
| `ParentAndRelativeKey` | [`Service.CreateByID`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go)<br>[`Service.Create`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go)<br>[`TestUnitDeviceWriteShapes`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go)<br>[`TestUnitDeviceCreateGeneratesAndKeepsID`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | [`TestAccRemoteLogCollectionDeviceRejectedWrites`](../../../tests/integration/lab/teams_remote_log_collection_device/teams_remote_log_collection_device_test.go) |

Recorded live limitation: Client-secret lab run proved the route through rejected or missing-target probes (400 invalid DeviceId, 404 missing device); a successful write remains unqualified.

### Remove-CsTeamsRemoteLogCollectionDevice

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.DeleteByIdentity`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceWriteShapes`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | None recorded |
| `ParentAndRelativeKey` | [`Service.DeleteByID`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | [`TestAccRemoteLogCollectionDeviceRejectedWrites`](../../../tests/integration/lab/teams_remote_log_collection_device/teams_remote_log_collection_device_test.go) |

Recorded live limitation: Client-secret lab run proved the route through rejected or missing-target probes (400 invalid DeviceId, 404 missing device); a successful write remains unqualified.

### Set-CsTeamsRemoteLogCollectionDevice

Owner: `EXT-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.UpdateByIdentity`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go)<br>[`Service.PatchByIdentity`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceWriteShapes`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | None recorded |
| `ParentAndRelativeKey` | [`Service.UpdateByID`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go)<br>[`Service.PatchByID`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service.go) | [`TestUnitDeviceRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/teams_remote_log_collection_device/service_test.go) | [`TestAccRemoteLogCollectionDeviceRejectedWrites`](../../../tests/integration/lab/teams_remote_log_collection_device/teams_remote_log_collection_device_test.go) |

Recorded live limitation: Client-secret lab run proved the route through rejected or missing-target probes (400 invalid DeviceId, 404 missing device); a successful write remains unqualified.

## migration

Contract and live limitations: [migration-contract-notes.md](migration-contract-notes.md).

### Get-CsMeetingMigrationStatus

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListMeetingMigrationStatuses`](../../../msteams/teams_admin_api/migration/service.go)<br>[`Service.GetMeetingMigrationStatusSummary`](../../../msteams/teams_admin_api/migration/service.go) | [`TestUnitMigrationDeniedOperationsUseAcceptanceResponses`](../../../msteams/teams_admin_api/migration/service_test.go)<br>[`TestUnitMigrationSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/migration/service_test.go) | [`TestAccMigrationReads`](../../../tests/integration/lab/migration/migration_test.go) |

Recorded live limitation: Client-secret lab run: records and -SummaryOnly both return HTTP 401 (empty body) to the app-only identity, matched by pinned PowerShell (Unauthorized); successful reads are unqualified.

### Get-CsMeetingMigrationTransactionHistory

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.ListMeetingMigrationTransactionHistory`](../../../msteams/teams_admin_api/migration/service.go) | No set mapping | No set mapping |

Recorded live limitation: Torus-only export. Client-secret lab run returned HTTP 403 for a reserved nonexistent user; successful reads are unqualified.

### Get-CsMoveTenantServiceInstanceTaskStatus

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.ListTenantMoveTaskStatuses`](../../../msteams/teams_admin_api/migration/service.go) | No set mapping | No set mapping |

Recorded live limitation: SelfHost/Torus-only export. Client-secret lab run returned HTTP 403; successful reads are unqualified.

### Get-NgtProvInstanceFailOverStatus

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.GetOCENgtProvFailoverStatus`](../../../msteams/teams_admin_api/policy_assignments/oce_ngtprov.go) | No set mapping | No set mapping |

Recorded live limitation: OCE declaration. Client-secret lab run of the NOAM status read returned HTTP 403; successful reads are unqualified.

### Move-CsAvsTenantPartition

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.MoveAVSTenantPartition`](../../../msteams/teams_admin_api/migration/service.go) | No set mapping | No set mapping |

Recorded live limitation: Proved only by a move between partitions of a reserved nonexistent tenant in a nonexistent container (HTTP 403); successful moves are unqualified.

### Move-CsTenantCrossRegion

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.MoveTenantCrossRegion`](../../../msteams/teams_admin_api/migration/service.go) | No set mapping | No set mapping |

Recorded live limitation: Never sent live: the request names no target the service could reject before acting, so any accepted call would irreversibly move the lab tenant to another region; covered by unit contract tests from module evidence only.

### Move-CsTenantServiceInstance

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.MoveTenantServiceInstance`](../../../msteams/teams_admin_api/migration/service.go) | No set mapping | No set mapping |

Recorded live limitation: Never sent live: the request names no target the service could reject before acting, so any accepted call would irreversibly start moving the lab tenant between service instances; covered by unit contract tests from module evidence only.

### Move-NgtProvInstance

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.MoveOCENgtProvInstance`](../../../msteams/teams_admin_api/policy_assignments/oce_ngtprov.go) | No set mapping | No set mapping |

Recorded live limitation: Never sent live: the request names no target the service could reject before acting, so any accepted call would irreversibly shift service-wide provisioning traffic between failover sides; covered by unit contract tests from module evidence only.

### Start-CsExMeetingMigration

Owner: `EXT-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `UserId` | [`Service.StartMeetingMigration`](../../../msteams/teams_admin_api/migration/service.go) | [`TestUnitMigrationDeniedOperationsUseAcceptanceResponses`](../../../msteams/teams_admin_api/migration/service_test.go)<br>[`TestUnitMigrationSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/migration/service_test.go) | [`TestAccMigrationRejectedWrites`](../../../tests/integration/lab/migration/migration_test.go) |

Recorded live limitation: Proved only by a request for a reserved nonexistent user ID (HTTP 401, empty body); no real user was migrated, so successful starts are unqualified.

## provisioning_diagnostics

Contract and live limitations: [provisioning-diagnostics-contract-notes.md](provisioning-diagnostics-contract-notes.md).

### Get-CsBusinessVoiceDirectoryDiagnosticData

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.GetDirectoryDiagnosticData`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Get-CsLoginInfo

Owner: `EXT-26`. Implementation: `not-started`; contract: `metadata-only`; live: `not-run`.

Recorded approved exclusion: Approved: returns the module's own access token and sends no service request; SDK credentials come from msteams/auth: docs/implementation/ext/provisioning-diagnostics-contract-notes.md

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | None recorded | No set mapping | No set mapping |

Recorded live limitation: Not implemented: the cmdlet returns the module's own ConfigAPI access token and sends no service request; SDK credentials come from msteams/auth.

### Get-CsMasObjectChangelog

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.ListObjectChangelog`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Get-CsMasVersionedSchemaData

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.ListVersionedSchemaData`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Get-CsUserProvHistory

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.GetUserProvisioningHistory`](../../../msteams/teams_admin_api/provisioning_diagnostics/oce.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Get-CsUssUserSettings

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.GetUserSetting`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Invoke-CsCustomHandlerCallBackNgtprov

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.InvokeCustomHandlerCallback`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 401 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Invoke-CsCustomHandlerNgtprov

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.InvokeCustomHandler`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 401 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Invoke-CsDirectObjectSync

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.SyncDirectObjects`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Invoke-CsDirectoryObjectSync

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.SyncDirectoryObject`](../../../msteams/teams_admin_api/provisioning_diagnostics/oce.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Invoke-CsGenericNgtProvCommand

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.InvokeNgtProvHandler`](../../../msteams/teams_admin_api/provisioning_diagnostics/oce.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Invoke-CsMsodsSync

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.SyncMsodsObjects`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Invoke-CsRehomeuser

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.RehomeUser`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Move-CsInternalHelper

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `BeginAndCompleteMove` | [`Service.BeginUserMove`](../../../msteams/teams_admin_api/provisioning_diagnostics/move.go)<br>[`Service.CompleteUserMove`](../../../msteams/teams_admin_api/provisioning_diagnostics/move.go) | [`TestUnitProvisioningDiagnosticsDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/provisioning_diagnostics/service_test.go)<br>[`TestUnitProvisioningDiagnosticsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/provisioning_diagnostics/service_test.go) | [`TestAccProvisioningDiagnosticsActionsDenied`](../../../tests/integration/lab/provisioning_diagnostics/provisioning_diagnostics_test.go) |
| `MoveResourcedata` | [`Service.SetMovedResourceData`](../../../msteams/teams_admin_api/provisioning_diagnostics/move.go) | [`TestUnitProvisioningDiagnosticsDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/provisioning_diagnostics/service_test.go)<br>[`TestUnitProvisioningDiagnosticsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/provisioning_diagnostics/service_test.go) | [`TestAccProvisioningDiagnosticsActionsDenied`](../../../tests/integration/lab/provisioning_diagnostics/provisioning_diagnostics_test.go) |
| `Rehome` | [`Service.RehomeUserMove`](../../../msteams/teams_admin_api/provisioning_diagnostics/move.go)<br>[`Service.GetUserMovePolicies`](../../../msteams/teams_admin_api/provisioning_diagnostics/move.go)<br>[`Service.RollbackUserMove`](../../../msteams/teams_admin_api/provisioning_diagnostics/move.go) | [`TestUnitProvisioningDiagnosticsDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/provisioning_diagnostics/service_test.go)<br>[`TestUnitProvisioningDiagnosticsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/provisioning_diagnostics/service_test.go) | [`TestAccProvisioningDiagnosticsActionsDenied`](../../../tests/integration/lab/provisioning_diagnostics/provisioning_diagnostics_test.go) |
| `Validate` | [`Service.ValidateUserMove`](../../../msteams/teams_admin_api/provisioning_diagnostics/move.go) | [`TestUnitProvisioningDiagnosticsDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/provisioning_diagnostics/service_test.go)<br>[`TestUnitProvisioningDiagnosticsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/provisioning_diagnostics/service_test.go) | [`TestAccProvisioningDiagnosticsReadsDenied`](../../../tests/integration/lab/provisioning_diagnostics/provisioning_diagnostics_test.go) |

Recorded live limitation: Client-secret lab run proved every HostedMigration route only through HTTP 403 denials sent for an example.invalid SIP address; in a tenant session the pinned export fails before any request because its private target is not loaded, so no PowerShell parity or successful move step is qualified.

### Set-CsTenantUserBackfill

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.SetTenantUserBackfill`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

### Set-CsUssUserSettings

Owner: `EXT-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.SetUserSetting`](../../../msteams/teams_admin_api/provisioning_diagnostics/service.go) | No set mapping | No set mapping |

Recorded live limitation: Client-secret lab run proved the route only through the app-only Teams Administrator HTTP 403 denial, sent against reserved targets; the pinned module does not load this support command in a tenant session, so no PowerShell parity or successful call is qualified.

## shifts_connections

Contract and live limitations: [shifts-connections-contract-notes.md](shifts-connections-contract-notes.md).

### Get-CsTeamsShiftsConnection

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.ListConnections`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `Get1` | [`Service.GetConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsSyncDenialPreservesReceiptsAndIndependentParentCleanup`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsConnectionsAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity` | [`Service.GetConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Get-CsTeamsShiftsConnectionConnector

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.ListConnectors`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Get-CsTeamsShiftsConnectionInstance

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.ListInstances`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `Get1` | [`Service.GetInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsSyncDenialPreservesReceiptsAndIndependentParentCleanup`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsConnectionsAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity` | [`Service.GetInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### New-CsTeamsShiftsConnection

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `New` | [`Service.CreateConnection`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `NewExpanded` | [`Service.CreateConnection`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionInstanceAndAsyncOperationKeepDistinctIDs`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go)<br>[`TestUnitShiftsSyncDenialPreservesReceiptsAndIndependentParentCleanup`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### New-CsTeamsShiftsConnectionInstance

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `New` | [`Service.CreateInstance`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `NewExpanded` | [`Service.CreateInstance`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionInstanceAndAsyncOperationKeepDistinctIDs`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go)<br>[`TestUnitShiftsSyncDenialPreservesReceiptsAndIndependentParentCleanup`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Remove-CsTeamsShiftsConnection

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Remove` | [`Service.DeleteConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsSyncDenialPreservesReceiptsAndIndependentParentCleanup`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `RemoveViaIdentity` | [`Service.DeleteConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Remove-CsTeamsShiftsConnectionInstance

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Remove` | [`Service.DeleteInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsSyncDenialPreservesReceiptsAndIndependentParentCleanup`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `RemoveViaIdentity` | [`Service.DeleteInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Set-CsTeamsShiftsConnection

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Set` | [`Service.ReplaceConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `SetExpanded` | [`Service.ReplaceConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `SetViaIdentity` | [`Service.ReplaceConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `SetViaIdentityExpanded` | [`Service.ReplaceConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Set-CsTeamsShiftsConnectionInstance

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Set` | [`Service.ReplaceInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `SetExpanded` | [`Service.ReplaceInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `SetViaIdentity` | [`Service.ReplaceInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `SetViaIdentityExpanded` | [`Service.ReplaceInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Test-CsTeamsShiftsConnectionValidate

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Test` | [`Service.Validate`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `TestExpanded` | [`Service.Validate`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Update-CsTeamsShiftsConnection

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Update` | [`Service.UpdateConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `UpdateExpanded` | [`Service.UpdateConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `UpdateViaIdentity` | [`Service.UpdateConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `UpdateViaIdentityExpanded` | [`Service.UpdateConnectionByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Update-CsTeamsShiftsConnectionInstance

Owner: `EXT-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Update` | [`Service.UpdateInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `UpdateExpanded` | [`Service.UpdateInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go)<br>[`TestUnitShiftsConnectionsSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `UpdateViaIdentity` | [`Service.UpdateInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `UpdateViaIdentityExpanded` | [`Service.UpdateInstanceByID`](../../../msteams/teams_admin_api/shifts_connections/service.go) | [`TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/service_test.go) | [`TestAccShiftsConnectionsRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

## shifts_sync

Contract and live limitations: [shifts-sync-contract-notes.md](shifts-sync-contract-notes.md).

### Get-CsTeamsShiftsConnectionErrorReport

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.ListErrorReports`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `Get1` | [`Service.GetErrorReportByID`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity` | [`Service.GetErrorReportByID`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Get-CsTeamsShiftsConnectionOperation

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.GetOperationByID`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsConnectionInstanceAndAsyncOperationKeepDistinctIDs`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity` | [`Service.GetOperationByID`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Get-CsTeamsShiftsConnectionSyncResult

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.GetSyncResult`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity` | [`Service.GetSyncResult`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Get-CsTeamsShiftsConnectionTeamMap

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.ListTeamMaps`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsConnectionInstanceAndAsyncOperationKeepDistinctIDs`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity` | [`Service.ListTeamMaps`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Get-CsTeamsShiftsConnectionWfmTeam

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.ListInstanceWfmTeams`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `Get1` | [`Service.ListConnectionWfmTeams`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity` | [`Service.ListInstanceWfmTeams`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity1` | [`Service.ListConnectionWfmTeams`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Get-CsTeamsShiftsConnectionWfmUser

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Get` | [`Service.ListWfmUsers`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `GetViaIdentity` | [`Service.ListWfmUsers`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncAuthorizationDenied`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### New-CsTeamsShiftsConnectionBatchTeamMap

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Create` | [`Service.MapTeams`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `CreateExpanded` | [`Service.MapTeams`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsConnectionInstanceAndAsyncOperationKeepDistinctIDs`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go)<br>[`TestUnitShiftsSyncDenialPreservesReceiptsAndIndependentParentCleanup`](../../../tests/integration/lab/ext_integration/shifts_composition_test.go) | [`TestAccShiftsSyncRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `CreateViaIdentity` | [`Service.MapTeams`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `CreateViaIdentityExpanded` | [`Service.MapTeams`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Remove-CsTeamsShiftsConnectionTeamMap

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Remove` | [`Service.DeleteTeamMap`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `RemoveViaIdentity` | [`Service.DeleteTeamMap`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

### Remove-CsTeamsShiftsScheduleRecord

Owner: `EXT-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Remove` | [`Service.ClearSchedule`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |
| `RemoveExpanded` | [`Service.ClearSchedule`](../../../msteams/teams_admin_api/shifts_connections/sync.go) | [`TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go)<br>[`TestUnitShiftsSyncSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/shifts_connections/sync_test.go) | [`TestAccShiftsSyncRejectedWrites`](../../../tests/integration/lab/shifts_connections/shifts_connections_test.go) |

Recorded live limitation: Client-secret lab run proved the routes only through HTTP 403 denials: the Shifts connector service denies app-only Teams Administrator credentials before validation, so successful reads and writes are unqualified.

## tenant_identity

Contract and live limitations: [tenant-identity-contract-notes.md](tenant-identity-contract-notes.md).

### Find-CsGroup

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.SearchGroups`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentitySuccessShapesPreserveFields`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`ExampleService_SearchGroups`](../../../msteams/teams_admin_api/tenant_identity/example_test.go) | [`TestAccTenantIdentityReads`](../../../tests/integration/lab/tenant_identity/tenant_identity_test.go) |

Recorded live limitation: Client-secret lab run searched the owned acceptance group (exact match, unmatched and short-query rejection) with pinned PowerShell parity. Qualification across delegated callers has not run.

### Get-CsAadTenant

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.GetAadTenant`](../../../msteams/teams_admin_api/tenant_identity/service.go) | No set mapping | No set mapping |

Recorded live limitation: Support-only (SelfHost/Torus) export. Client-secret lab run confirmed the app-only HTTP 403; a successful read is unqualified.

### Get-CsAadUser

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.GetAadUser`](../../../msteams/teams_admin_api/tenant_identity/service.go) | No set mapping | No set mapping |

Recorded live limitation: Support-only (SelfHost/Torus) export. Client-secret lab run confirmed the app-only HTTP 403; a successful read is unqualified.

### Get-CsCloudTenant

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.GetCloudTenant`](../../../msteams/teams_admin_api/tenant_identity/service.go) | No set mapping | No set mapping |

Recorded live limitation: Support-only (SelfHost/Torus) export reported retired by the service. Client-secret lab run confirmed the app-only HTTP 403; a successful read is unqualified.

### Get-CsCloudUser

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.GetCloudUser`](../../../msteams/teams_admin_api/tenant_identity/service.go) | No set mapping | No set mapping |

Recorded live limitation: Support-only (SelfHost/Torus) export reported retired by the service. Client-secret lab run confirmed the app-only HTTP 403; a successful read is unqualified.

### Get-CsHostingProvider

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| No observed set; conditional scope requires evidence | [`Service.ListHostingProviders`](../../../msteams/teams_admin_api/tenant_identity/service.go)<br>[`Service.ListHostingProvidersFiltered`](../../../msteams/teams_admin_api/tenant_identity/service.go)<br>[`Service.GetHostingProviderByIdentity`](../../../msteams/teams_admin_api/tenant_identity/service.go) | No set mapping | No set mapping |

Recorded live limitation: Support-only (SelfHost/Torus) export. Client-secret lab run confirmed the app-only HTTP 403 (errorCode 40301); a successful read is unqualified.

### Get-CsOnlineDirectoryTenant

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetOnlineDirectoryTenant`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go) | [`TestAccTenantIdentityReads`](../../../tests/integration/lab/tenant_identity/tenant_identity_test.go) |

Recorded live limitation: The module dispatches to Invoke-CsDeprecatedError; the lab run confirmed the deprecation in Go and pinned PowerShell. No service call exists to qualify.

### Get-CsTenant

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.GetTenant`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`ExampleService_GetTenant`](../../../msteams/teams_admin_api/tenant_identity/example_test.go) | [`TestAccTenantIdentityReads`](../../../tests/integration/lab/tenant_identity/tenant_identity_test.go) |

Recorded live limitation: Client-secret lab run read the extended tenant record with pinned PowerShell parity. The module Filter/Identity/ResultSize parameters fail binding in 8.1.0 and are unqualified.

### Get-CsTenantLicensingConfiguration

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Filter` | [`Service.ListTenantLicensingConfigurations`](../../../msteams/teams_admin_api/tenant_identity/service.go)<br>[`Service.ListTenantLicensingConfigurationsFiltered`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentitySuccessShapesPreserveFields`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go) | [`TestAccTenantIdentityReads`](../../../tests/integration/lab/tenant_identity/tenant_identity_test.go) |
| `Identity` | [`Service.GetTenantLicensingConfigurationByIdentity`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityDenialsUseAcceptanceResponses`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go) | [`TestAccTenantIdentityReads`](../../../tests/integration/lab/tenant_identity/tenant_identity_test.go) |

Recorded live limitation: Client-secret lab run confirmed the app-only HTTP 403 (errorCode 40301) with pinned PowerShell parity; a successful read needs separately authorized credentials.

### Remove-CsUserLicenseGracePeriod

Owner: `EXT-29`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Remove` | [`Service.RemoveUserLicenseGracePeriod`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityGracePeriodRemovalWire`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go) | None recorded |
| `RemoveExpanded` | [`Service.RemoveUserLicenseGracePeriod`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityGracePeriodRemovalWire`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go) | None recorded |
| `RemoveViaIdentity` | [`Service.RemoveUserLicenseGracePeriod`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityGracePeriodRemovalWire`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go) | None recorded |
| `RemoveViaIdentityExpanded` | [`Service.RemoveUserLicenseGracePeriod`](../../../msteams/teams_admin_api/tenant_identity/service.go) | [`TestUnitTenantIdentityGracePeriodRemovalWire`](../../../msteams/teams_admin_api/tenant_identity/service_test.go)<br>[`TestUnitTenantIdentityRejectsInputBeforeIO`](../../../msteams/teams_admin_api/tenant_identity/service_test.go) | None recorded |

Recorded live limitation: Unqualified: ending a license grace period changes licensing state irreversibly, so the lab never calls it. The wire contract comes from the pinned module model and IL only.

## tenant_configuration

Contract and live limitations: [conditional-settings-contract-notes.md](conditional-settings-contract-notes.md).

### Get-CsTenantConfiguration

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.GetByIdentity`](../../../msteams/teams_admin_api/tenant_configuration/service.go)<br>[`Service.List`](../../../msteams/teams_admin_api/tenant_configuration/service.go)<br>[`Service.ListFiltered`](../../../msteams/teams_admin_api/tenant_configuration/service.go) | [`TestUnitConditionalGetIdentityFilterAndNullableProjection`](../../../msteams/teams_admin_api/tenant_configuration/service_test.go) | None recorded |
| `Filter` (compiled declaration) | [`Service.GetByIdentity`](../../../msteams/teams_admin_api/tenant_configuration/service.go)<br>[`Service.List`](../../../msteams/teams_admin_api/tenant_configuration/service.go)<br>[`Service.ListFiltered`](../../../msteams/teams_admin_api/tenant_configuration/service.go) | [`TestUnitConditionalGetIdentityFilterAndNullableProjection`](../../../msteams/teams_admin_api/tenant_configuration/service_test.go) | None recorded |

Recorded live limitation: OCE-only compiled declaration; authorization, server acceptance and root-export availability are unqualified. No live Global setting mutation is configured.

### Set-CsTenantConfiguration

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.PatchByIdentity`](../../../msteams/teams_admin_api/tenant_configuration/service.go)<br>[`Service.UpdateByIdentity`](../../../msteams/teams_admin_api/tenant_configuration/service.go) | [`TestUnitConditionalSetPreservesExplicitNullAndPreparationReceipt`](../../../msteams/teams_admin_api/tenant_configuration/service_test.go)<br>[`TestUnitConditionalSetRetainsReadDenialAndAcceptedDecodeFailure`](../../../msteams/teams_admin_api/tenant_configuration/service_test.go)<br>[`TestUnitConditionalSparsePatchOmitsUnboundFieldsAndKeepsFalseForce`](../../../msteams/teams_admin_api/tenant_configuration/service_test.go)<br>[`TestUnitConditionalSelectedSettingValidationPrecedesCredentials`](../../../msteams/teams_admin_api/tenant_configuration/service_test.go) | None recorded |

Recorded live limitation: OCE-only compiled declaration; authorization, server acceptance and root-export availability are unqualified. No live Global setting mutation is configured.

## teams_app_policy_configuration

Contract and live limitations: [conditional-settings-contract-notes.md](conditional-settings-contract-notes.md).

### Get-CsTeamsAppPolicyConfiguration

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.GetByIdentity`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service.go)<br>[`Service.List`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service.go)<br>[`Service.ListFiltered`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service.go) | [`TestUnitConditionalGetIdentityFilterAndNullableProjection`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service_test.go) | None recorded |
| `Filter` (compiled declaration) | [`Service.GetByIdentity`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service.go)<br>[`Service.List`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service.go)<br>[`Service.ListFiltered`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service.go) | [`TestUnitConditionalGetIdentityFilterAndNullableProjection`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service_test.go) | None recorded |

Recorded live limitation: OCE-only compiled declaration; authorization, server acceptance and root-export availability are unqualified. No live Global setting mutation is configured.

### Set-CsTeamsAppPolicyConfiguration

Owner: `EXT-90`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` (compiled declaration) | [`Service.PatchByIdentity`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service.go)<br>[`Service.UpdateByIdentity`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service.go) | [`TestUnitConditionalSetPreservesExplicitNullAndPreparationReceipt`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service_test.go)<br>[`TestUnitConditionalSetRetainsReadDenialAndAcceptedDecodeFailure`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service_test.go)<br>[`TestUnitConditionalSparsePatchOmitsUnboundFieldsAndKeepsFalseForce`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service_test.go)<br>[`TestUnitConditionalSelectedSettingValidationPrecedesCredentials`](../../../msteams/teams_admin_api/teams_app_policy_configuration/service_test.go) | None recorded |

Recorded live limitation: OCE-only compiled declaration; authorization, server acceptance and root-export availability are unqualified. No live Global setting mutation is configured.

## unified_apps

Contract and live limitations: [unified-apps-contract-notes.md](unified-apps-contract-notes.md).

### Get-AllM365TeamsApps

Owner: `EXT-30`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.List`](../../../msteams/teams_admin_api/unified_apps/service.go) | [`TestUnitUnifiedAppReadsDecodeCatalogShapes`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppFailuresStayTyped`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppProfileRejectsBeforeTokens`](../../../msteams/teams_admin_api/unified_apps/service_test.go) | [`TestAccUnifiedAppsApplicationIdentityRefused`](../../../tests/integration/lab/unified_apps/unified_apps_test.go)<br>[`TestAccUnifiedAppsDelegatedReadOnly`](../../../tests/integration/lab/unified_apps/unified_apps_test.go) |

Recorded live limitation: Lab run confirmed application-mode refusal with pinned PowerShell parity (the service returns 401 to app-only tokens); delegated AdminAppCatalog.ReadWrite.All reads and all writes need a separately provisioned delegated identity.

### Get-M365TeamsApp

Owner: `EXT-30`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Get`](../../../msteams/teams_admin_api/unified_apps/service.go) | [`TestUnitUnifiedAppReadsDecodeCatalogShapes`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppFailuresStayTyped`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppInvalidInputFailsLocally`](../../../msteams/teams_admin_api/unified_apps/service_test.go) | [`TestAccUnifiedAppsApplicationIdentityRefused`](../../../tests/integration/lab/unified_apps/unified_apps_test.go)<br>[`TestAccUnifiedAppsDelegatedReadOnly`](../../../tests/integration/lab/unified_apps/unified_apps_test.go) |

Recorded live limitation: Lab run confirmed application-mode refusal with pinned PowerShell parity (the service returns 401 to app-only tokens); delegated AdminAppCatalog.ReadWrite.All reads and all writes need a separately provisioned delegated identity.

### Get-M365UnifiedCustomPendingApps

Owner: `EXT-30`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListCustomPending`](../../../msteams/teams_admin_api/unified_apps/service.go) | [`TestUnitUnifiedAppReadsDecodeCatalogShapes`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppProfileRejectsBeforeTokens`](../../../msteams/teams_admin_api/unified_apps/service_test.go) | [`TestAccUnifiedAppsApplicationIdentityRefused`](../../../tests/integration/lab/unified_apps/unified_apps_test.go)<br>[`TestAccUnifiedAppsDelegatedReadOnly`](../../../tests/integration/lab/unified_apps/unified_apps_test.go) |

Recorded live limitation: Lab run confirmed application-mode refusal with pinned PowerShell parity (the service returns 401 to app-only tokens); delegated AdminAppCatalog.ReadWrite.All reads and all writes need a separately provisioned delegated identity.

### Get-M365UnifiedTenantSettings

Owner: `EXT-30`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetTenantSettings`](../../../msteams/teams_admin_api/unified_apps/service.go) | [`TestUnitUnifiedAppReadsDecodeCatalogShapes`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppInvalidInputFailsLocally`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppProfileRejectsBeforeTokens`](../../../msteams/teams_admin_api/unified_apps/service_test.go) | [`TestAccUnifiedAppsApplicationIdentityRefused`](../../../tests/integration/lab/unified_apps/unified_apps_test.go)<br>[`TestAccUnifiedAppsDelegatedReadOnly`](../../../tests/integration/lab/unified_apps/unified_apps_test.go) |

Recorded live limitation: Lab run confirmed application-mode refusal with pinned PowerShell parity (the service returns 401 to app-only tokens); delegated AdminAppCatalog.ReadWrite.All reads and all writes need a separately provisioned delegated identity.

### Update-M365TeamsApp

Owner: `EXT-30`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Update`](../../../msteams/teams_admin_api/unified_apps/service.go) | [`TestUnitUnifiedAppWritesSendModuleBodies`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppFailuresStayTyped`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppInvalidInputFailsLocally`](../../../msteams/teams_admin_api/unified_apps/service_test.go) | [`TestAccUnifiedAppsApplicationIdentityRefused`](../../../tests/integration/lab/unified_apps/unified_apps_test.go) |

Recorded live limitation: Lab run confirmed application-mode refusal with pinned PowerShell parity (the service returns 401 to app-only tokens); delegated AdminAppCatalog.ReadWrite.All reads and all writes need a separately provisioned delegated identity.

### Update-M365UnifiedCustomPendingApp

Owner: `EXT-30`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ReviewCustomPending`](../../../msteams/teams_admin_api/unified_apps/service.go) | [`TestUnitUnifiedAppReviewReadsEtagThenPuts`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppInvalidInputFailsLocally`](../../../msteams/teams_admin_api/unified_apps/service_test.go) | [`TestAccUnifiedAppsApplicationIdentityRefused`](../../../tests/integration/lab/unified_apps/unified_apps_test.go) |

Recorded live limitation: Lab run confirmed application-mode refusal with pinned PowerShell parity (the service returns 401 to app-only tokens); delegated AdminAppCatalog.ReadWrite.All reads and all writes need a separately provisioned delegated identity.

### Update-M365UnifiedTenantSettings

Owner: `EXT-30`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateTenantSettings`](../../../msteams/teams_admin_api/unified_apps/service.go) | [`TestUnitUnifiedAppWritesSendModuleBodies`](../../../msteams/teams_admin_api/unified_apps/service_test.go)<br>[`TestUnitUnifiedAppInvalidInputFailsLocally`](../../../msteams/teams_admin_api/unified_apps/service_test.go) | [`TestAccUnifiedAppsApplicationIdentityRefused`](../../../tests/integration/lab/unified_apps/unified_apps_test.go) |

Recorded live limitation: Lab run confirmed application-mode refusal with pinned PowerShell parity (the service returns 401 to app-only tokens); delegated AdminAppCatalog.ReadWrite.All reads and all writes need a separately provisioned delegated identity.
