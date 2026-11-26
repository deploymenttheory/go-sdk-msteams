# CALL command and parameter-set evidence inventory

Owner: CALL-90 (#188). Baseline: MicrosoftTeams 8.1.0.

This inventory names 95 commands and 109 observed parameter sets
across all nine CALL families. Regenerate with
`python3 tools/coverage/call_report.py`; use `--check` to detect stale output.

## Interpretation

Symbol and test links are declared mappings, not proof that those tests ran
or that they cover every behavior. States and limitations below come from
the current ledger; this report never promotes them. A synthetic success
body is offline model evidence, not a successful tenant operation.

The integrated offline suite is
[`tests/integration/lab/call_integration`](../../../tests/integration/lab/call_integration/contracts_test.go).
It exercises schedule/association identities, audio IDs across prompt/queue
models, and independent recovery after a different service denies access.
The four offline composed scenarios are synthetic. The separate opt-in
`TestAccScheduleAutoAttendantComposition` journals an owned weekly schedule and
attendant, verifies the association before/after replacement against pinned
PowerShell, and independently removes both. Registration is not a live pass.
Existing child fixtures retain their provenance. Executed results and the tested
revision belong in the implementation PR/CI, separate from this inventory.

## Qualification boundaries

- Resource-account successful creation/association needs directory-write or
  delegated credentials unavailable to the existing app-only lab identity.
- Compliance recording application success needs a real resource account;
  grants affect calling behavior and are not qualified by policy CRUD.
- Auto-recording template success needs SharePoint site-admin access; an
  expected denial is not a successful create/get/update/delete lifecycle.
- Schedule/attendant references, audio prompt use and populated resource
  links require live compositions beyond the isolated child lifecycles.
- Commercial application-auth evidence does not qualify delegated flows,
  sovereign clouds or effective calling behavior.

CALL-90 remains incomplete until the integrated acceptance and complete
contract criteria are demonstrated. Child issue closure and mapped symbols
alone do not sign off CALL-90 or the CALL epic.

## attendant_builders

Contract and live limitations: [auto-attendants-contract-notes.md](auto-attendants-contract-notes.md).

### New-CsAutoAttendantCallableEntity

Owner: `CALL-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DraftCallableEntity`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitDraftBuildersPreserveNonDefaultMembers`](../../../msteams/teams_admin_api/auto_attendants/draft_variants_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: New-CsAutoAttendant* builders are server draft validations, not local constructors; a client-secret lab run passed the draft route for this builder (or its service rejection) and the owned lifecycle built from equivalent Go values. Full release qualification is separate.

### New-CsAutoAttendantCallFlow

Owner: `CALL-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DraftCallFlow`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitDraftBuildersPreserveNonDefaultMembers`](../../../msteams/teams_admin_api/auto_attendants/draft_variants_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: New-CsAutoAttendant* builders are server draft validations, not local constructors; a client-secret lab run passed the draft route for this builder (or its service rejection) and the owned lifecycle built from equivalent Go values. Full release qualification is separate.

### New-CsAutoAttendantCallHandlingAssociation

Owner: `CALL-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DraftCallHandlingAssociation`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitDraftBuildersPreserveNonDefaultMembers`](../../../msteams/teams_admin_api/auto_attendants/draft_variants_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: New-CsAutoAttendant* builders are server draft validations, not local constructors; a client-secret lab run passed the draft route for this builder (or its service rejection) and the owned lifecycle built from equivalent Go values. Full release qualification is separate.

### New-CsAutoAttendantDialScope

Owner: `CALL-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DraftGroupDialScope`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: New-CsAutoAttendantDialScope posts to a group dial scope draft route; the lab run qualified only the unknown-group HTTP 404 rejection because no qualified group fixture is used for dial scopes. Full release qualification is separate.

### New-CsAutoAttendantMenu

Owner: `CALL-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DraftMenu`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitDraftBuildersPreserveNonDefaultMembers`](../../../msteams/teams_admin_api/auto_attendants/draft_variants_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: New-CsAutoAttendant* builders are server draft validations, not local constructors; a client-secret lab run passed the draft route for this builder (or its service rejection) and the owned lifecycle built from equivalent Go values. Full release qualification is separate.

### New-CsAutoAttendantMenuOption

Owner: `CALL-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DraftMenuOption`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitDraftBuildersPreserveNonDefaultMembers`](../../../msteams/teams_admin_api/auto_attendants/draft_variants_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: New-CsAutoAttendant* builders are server draft validations, not local constructors; a client-secret lab run passed the draft route for this builder (or its service rejection) and the owned lifecycle built from equivalent Go values. Full release qualification is separate.

### New-CsAutoAttendantPrompt

Owner: `CALL-20`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `AudioFileParamSet` | [`Service.DraftPrompt`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitDraftBuildersPreserveNonDefaultMembers`](../../../msteams/teams_admin_api/auto_attendants/draft_variants_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |
| `DualParamSet` | [`Service.DraftPrompt`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitDraftBuildersPreserveNonDefaultMembers`](../../../msteams/teams_admin_api/auto_attendants/draft_variants_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |
| `TextToSpeechParamSet` | [`Service.DraftPrompt`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: New-CsAutoAttendant* builders are server draft validations, not local constructors; a client-secret lab run passed the draft route for this builder (or its service rejection) and the owned lifecycle built from equivalent Go values. Full release qualification is separate.

## audio_prompts

Contract and live limitations: [audio-prompts-contract-notes.md](audio-prompts-contract-notes.md).

### Export-CsOnlineAudioFile

Owner: `CALL-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetContentByID`](../../../msteams/teams_admin_api/audio_prompts/service.go) | [`TestUnitAudioFileLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/audio_prompts/service_test.go)<br>[`TestUnitAudioFileExportAcceptsURLSafeUnpaddedBase64`](../../../msteams/teams_admin_api/audio_prompts/service_test.go) | [`TestAccAudioFileLifecycle`](../../../tests/integration/lab/audio_prompts/audio_prompts_test.go) |

Live limitation: Client-secret lab run passed the owned TenantGlobal import/get/list/export/soft-delete lifecycle with pinned PowerShell export parity and journaled cleanup; full release qualification is separate.

### Get-CsOnlineAudioFile

Owner: `CALL-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.List`](../../../msteams/teams_admin_api/audio_prompts/service.go)<br>[`Service.GetByID`](../../../msteams/teams_admin_api/audio_prompts/service.go) | [`TestUnitAudioFileLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/audio_prompts/service_test.go)<br>[`TestUnitAudioFileListNormalizesApplicationAndPreservesNulls`](../../../msteams/teams_admin_api/audio_prompts/service_test.go)<br>[`TestUnitAudioFileGetSendsDurationOnlyWhenSet`](../../../msteams/teams_admin_api/audio_prompts/service_test.go) | [`TestAccAudioFileLifecycle`](../../../tests/integration/lab/audio_prompts/audio_prompts_test.go) |

Live limitation: Client-secret lab run passed the owned TenantGlobal import/get/list/export/soft-delete lifecycle with pinned PowerShell export parity and journaled cleanup; full release qualification is separate.

### Import-CsOnlineAudioFile

Owner: `CALL-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Create`](../../../msteams/teams_admin_api/audio_prompts/service.go) | [`TestUnitAudioFileLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/audio_prompts/service_test.go) | [`TestAccAudioFileLifecycle`](../../../tests/integration/lab/audio_prompts/audio_prompts_test.go) |

Live limitation: Client-secret lab run passed the owned TenantGlobal import/get/list/export/soft-delete lifecycle with pinned PowerShell export parity and journaled cleanup; full release qualification is separate.

### New-CsCustomPrompt

Owner: `CALL-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`NewPrompt`](../../../msteams/teams_admin_api/audio_conferencing_prompts/edits.go) | [`TestUnitTypedPromptConstructors`](../../../msteams/teams_admin_api/audio_conferencing_prompts/models_test.go)<br>[`TestUnitPromptsBuildersNeverAcquireCredentialsOrHTTP`](../../../msteams/teams_admin_api/audio_conferencing_prompts/failure_test.go) | None recorded |

Live limitation: Local constructors perform no I/O; they are exercised through the audio conferencing custom prompts configuration, whose prompt edits need the tenant flight.

### New-CsCustomPromptPackage

Owner: `CALL-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`NewPromptPackage`](../../../msteams/teams_admin_api/audio_conferencing_prompts/edits.go) | [`TestUnitTypedPromptConstructors`](../../../msteams/teams_admin_api/audio_conferencing_prompts/models_test.go)<br>[`TestUnitPromptsBuildersNeverAcquireCredentialsOrHTTP`](../../../msteams/teams_admin_api/audio_conferencing_prompts/failure_test.go) | None recorded |

Live limitation: Local constructors perform no I/O; they are exercised through the audio conferencing custom prompts configuration, whose prompt edits need the tenant flight.

### Remove-CsOnlineAudioFile

Owner: `CALL-21`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteByID`](../../../msteams/teams_admin_api/audio_prompts/service.go) | [`TestUnitAudioFileLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/audio_prompts/service_test.go)<br>[`TestUnitAudioFileServiceFailuresStayTyped`](../../../msteams/teams_admin_api/audio_prompts/service_test.go) | [`TestAccAudioFileLifecycle`](../../../tests/integration/lab/audio_prompts/audio_prompts_test.go) |

Live limitation: Client-secret lab run passed the owned TenantGlobal import/get/list/export/soft-delete lifecycle with pinned PowerShell export parity and journaled cleanup; full release qualification is separate.

## auto_attendants

Contract and live limitations: [auto-attendants-contract-notes.md](auto-attendants-contract-notes.md).

### Get-CsAutoAttendant

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `GetAllParamSet` | [`Service.List`](../../../msteams/teams_admin_api/auto_attendants/service.go)<br>[`Service.GetByID`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantOptionalQueriesAndNormalization`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantLifecycle`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go)<br>[`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |
| `GetSpecificParamSet` | [`Service.List`](../../../msteams/teams_admin_api/auto_attendants/service.go)<br>[`Service.GetByID`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantOptionalQueriesAndNormalization`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantLifecycle`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go)<br>[`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

### Get-CsAutoAttendantStatus

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetStatusByID`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantOptionalQueriesAndNormalization`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantLifecycle`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go)<br>[`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

### Get-CsAutoAttendantSupportedLanguage

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListSupportedLanguages`](../../../msteams/teams_admin_api/auto_attendants/service.go)<br>[`Service.GetSupportedLanguageByID`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

### Get-CsAutoAttendantSupportedTimeZone

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListSupportedTimeZones`](../../../msteams/teams_admin_api/auto_attendants/service.go)<br>[`Service.GetSupportedTimeZoneByID`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

### Get-CsAutoAttendantTenantInformation

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetTenantInformation`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

### New-CsAutoAttendant

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Create`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantOptionalQueriesAndNormalization`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantLifecycle`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

### Remove-CsAutoAttendant

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteByID`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantServiceRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantLifecycle`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go)<br>[`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

### Set-CsAutoAttendant

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateByID`](../../../msteams/teams_admin_api/auto_attendants/service.go)<br>[`AutoAttendant.Settings`](../../../msteams/teams_admin_api/auto_attendants/models.go) | [`TestUnitAutoAttendantLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go)<br>[`TestUnitAutoAttendantOptionalQueriesAndNormalization`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantLifecycle`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

### Update-CsAutoAttendant

Owner: `CALL-22`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.RefreshByID`](../../../msteams/teams_admin_api/auto_attendants/service.go) | [`TestUnitAutoAttendantLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/auto_attendants/service_test.go) | [`TestAccAutoAttendantLifecycle`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go)<br>[`TestAccAutoAttendantReads`](../../../tests/integration/lab/auto_attendants/auto_attendants_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/status/list/full-body update/read-back/refresh/delete lifecycle and the tenant information, supported language and time zone reads with pinned PowerShell read parity and journaled cleanup; operators, real schedule associations, dial-by-name scopes, audio prompts, voice response, mainline attendant and full release qualification are separate.

## call_queues

Contract and live limitations: [call-queues-contract-notes.md](call-queues-contract-notes.md).

### Get-CsCallQueue

Owner: `CALL-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.List`](../../../msteams/teams_admin_api/call_queues/service.go)<br>[`Service.GetByID`](../../../msteams/teams_admin_api/call_queues/service.go) | [`TestUnitCallQueueLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_queues/service_test.go)<br>[`TestUnitCallQueueMissingIdentityIsHTTP400`](../../../msteams/teams_admin_api/call_queues/service_test.go)<br>[`TestUnitCallQueueOptionalQueriesAndNormalization`](../../../msteams/teams_admin_api/call_queues/service_test.go) | [`TestAccCallQueueLifecycle`](../../../tests/integration/lab/call_queues/call_queues_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/list/full-body update/read-back/delete lifecycle with pinned PowerShell read parity and journaled cleanup; agents, resource-account (OBO) links, channel and callback settings, and full release qualification are separate.

### New-CsCallQueue

Owner: `CALL-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Create`](../../../msteams/teams_admin_api/call_queues/service.go) | [`TestUnitCallQueueLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_queues/service_test.go)<br>[`TestUnitCallQueueOptionalQueriesAndNormalization`](../../../msteams/teams_admin_api/call_queues/service_test.go)<br>[`TestUnitCallQueueInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/call_queues/service_test.go) | [`TestAccCallQueueLifecycle`](../../../tests/integration/lab/call_queues/call_queues_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/list/full-body update/read-back/delete lifecycle with pinned PowerShell read parity and journaled cleanup; agents, resource-account (OBO) links, channel and callback settings, and full release qualification are separate.

### Remove-CsCallQueue

Owner: `CALL-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteByID`](../../../msteams/teams_admin_api/call_queues/service.go) | [`TestUnitCallQueueLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_queues/service_test.go)<br>[`TestUnitCallQueueMissingIdentityIsHTTP400`](../../../msteams/teams_admin_api/call_queues/service_test.go) | [`TestAccCallQueueLifecycle`](../../../tests/integration/lab/call_queues/call_queues_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/list/full-body update/read-back/delete lifecycle with pinned PowerShell read parity and journaled cleanup; agents, resource-account (OBO) links, channel and callback settings, and full release qualification are separate.

### Set-CsCallQueue

Owner: `CALL-23`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateByID`](../../../msteams/teams_admin_api/call_queues/service.go)<br>[`CallQueue.Settings`](../../../msteams/teams_admin_api/call_queues/models.go) | [`TestUnitCallQueueLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_queues/service_test.go)<br>[`TestUnitCallQueueSettingsBackfillsFromRead`](../../../msteams/teams_admin_api/call_queues/service_test.go)<br>[`TestUnitCallQueueOptionalQueriesAndNormalization`](../../../msteams/teams_admin_api/call_queues/service_test.go) | [`TestAccCallQueueLifecycle`](../../../tests/integration/lab/call_queues/call_queues_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/list/full-body update/read-back/delete lifecycle with pinned PowerShell read parity and journaled cleanup; agents, resource-account (OBO) links, channel and callback settings, and full release qualification are separate.

## call_templates

Contract and live limitations: [call-templates-contract-notes.md](call-templates-contract-notes.md).

### Get-CsAutoRecordingTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.List`](../../../msteams/teams_admin_api/call_templates/service.go)<br>[`Templates.GetByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitMissingTemplateIsNotFound`](../../../msteams/teams_admin_api/call_templates/service_test.go)<br>[`TestUnitAutoRecordingRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccAutoRecordingTemplateCreateDenied`](../../../tests/integration/lab/call_templates/call_templates_test.go)<br>[`TestAccCallTemplateReads`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Unqualified: the service validates the SharePoint site with the caller's identity and denies the app-only lab identity (HTTP 401, code 100012, SharePointClient_Unauthorized). The client-secret lab run verified that denial, list, body validation and missing-template 404s only; successful writes need a SharePoint-capable identity.

### Get-CsComplianceRecordingForCallQueueTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.List`](../../../msteams/teams_admin_api/call_templates/service.go)<br>[`Templates.GetByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitComplianceRecordingLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccComplianceRecordingTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go)<br>[`TestAccCallTemplateReads`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Get-CsSharedCallHistoryTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.List`](../../../msteams/teams_admin_api/call_templates/service.go)<br>[`Templates.GetByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccSharedCallHistoryTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go)<br>[`TestAccCallTemplateReads`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Get-CsSharedCallQueueHistoryTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.List`](../../../msteams/teams_admin_api/call_templates/service.go)<br>[`Templates.GetByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccSharedCallHistoryTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go)<br>[`TestAccCallTemplateReads`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Deprecated alias of the shared call history template on the same route. Client-secret lab run passed the owned shared call history lifecycle with pinned PowerShell parity that also reads through the alias; full release qualification is separate.

### Get-CsSharedVoicemailTriageSettingsTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.List`](../../../msteams/teams_admin_api/call_templates/service.go)<br>[`Templates.GetByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitVoicemailTriageLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccVoicemailTriageTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go)<br>[`TestAccCallTemplateReads`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Get-CsTagsTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.List`](../../../msteams/teams_admin_api/call_templates/service.go)<br>[`Templates.GetByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitTagsLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccTagsTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go)<br>[`TestAccCallTemplateReads`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### New-CsAutoRecordingTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.Create`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitAutoRecordingRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccAutoRecordingTemplateCreateDenied`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Unqualified: the service validates the SharePoint site with the caller's identity and denies the app-only lab identity (HTTP 401, code 100012, SharePointClient_Unauthorized). The client-secret lab run verified that denial, list, body validation and missing-template 404s only; successful writes need a SharePoint-capable identity.

### New-CsComplianceRecordingForCallQueueTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.Create`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitComplianceRecordingLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccComplianceRecordingTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### New-CsSharedCallHistoryTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.Create`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccSharedCallHistoryTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### New-CsSharedCallQueueHistoryTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.Create`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccSharedCallHistoryTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Deprecated alias of the shared call history template on the same route. Client-secret lab run passed the owned shared call history lifecycle with pinned PowerShell parity that also reads through the alias; full release qualification is separate.

### New-CsSharedVoicemailTriageSettingsTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.Create`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitVoicemailTriageLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccVoicemailTriageTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### New-CsTagsTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.Create`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitTagsLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccTagsTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Remove-CsAutoRecordingTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.DeleteByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitMissingTemplateIsNotFound`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccAutoRecordingTemplateCreateDenied`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Unqualified: the service validates the SharePoint site with the caller's identity and denies the app-only lab identity (HTTP 401, code 100012, SharePointClient_Unauthorized). The client-secret lab run verified that denial, list, body validation and missing-template 404s only; successful writes need a SharePoint-capable identity.

### Remove-CsComplianceRecordingForCallQueueTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.DeleteByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitComplianceRecordingLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccComplianceRecordingTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Remove-CsSharedCallHistoryTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.DeleteByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccSharedCallHistoryTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Remove-CsSharedCallQueueHistoryTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.DeleteByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccSharedCallHistoryTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Deprecated alias of the shared call history template on the same route. Client-secret lab run passed the owned shared call history lifecycle with pinned PowerShell parity that also reads through the alias; full release qualification is separate.

### Remove-CsSharedVoicemailTriageSettingsTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.DeleteByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitVoicemailTriageLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccVoicemailTriageTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Remove-CsTagsTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.DeleteByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitTagsLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccTagsTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Set-CsAutoRecordingTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Templates.UpdateByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitAutoRecordingRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccAutoRecordingTemplateCreateDenied`](../../../tests/integration/lab/call_templates/call_templates_test.go) |
| `Instance` | [`Templates.UpdateByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitAutoRecordingRejectionsUseAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccAutoRecordingTemplateCreateDenied`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Unqualified: the service validates the SharePoint site with the caller's identity and denies the app-only lab identity (HTTP 401, code 100012, SharePointClient_Unauthorized). The client-secret lab run verified that denial, list, body validation and missing-template 404s only; successful writes need a SharePoint-capable identity.

### Set-CsComplianceRecordingForCallQueueTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.UpdateByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitComplianceRecordingLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccComplianceRecordingTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Set-CsSharedCallHistoryTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.UpdateByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccSharedCallHistoryTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Set-CsSharedCallQueueHistoryTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.UpdateByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitSharedCallHistoryLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccSharedCallHistoryTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Deprecated alias of the shared call history template on the same route. Client-secret lab run passed the owned shared call history lifecycle with pinned PowerShell parity that also reads through the alias; full release qualification is separate.

### Set-CsSharedVoicemailTriageSettingsTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.UpdateByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitVoicemailTriageLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccVoicemailTriageTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Set-CsTagsTemplate

Owner: `CALL-24`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Templates.UpdateByID`](../../../msteams/teams_admin_api/call_templates/service.go) | [`TestUnitTagsLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/call_templates/service_test.go) | [`TestAccTagsTemplateLifecycle`](../../../tests/integration/lab/call_templates/call_templates_test.go) |

Live limitation: Client-secret lab run passed the owned create/get/update/list/delete lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

## compliance_applications

Contract and live limitations: [compliance-recording-contract-notes.md](compliance-recording-contract-notes.md).

### Get-CsTeamsComplianceRecordingApplication

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Filter` | [`Service.ListApplicationsFiltered`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |
| `Identity` | [`Service.GetApplicationByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go)<br>[`Service.ListApplications`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingApplicationRejections`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go)<br>[`TestAccComplianceRecordingReadOnly`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |
| `ParentAndRelativeKey` | [`Service.GetApplicationByID`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingApplicationRejections`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |

Live limitation: Client-secret lab run proved the application routes through service rejections and absence only: a successful application needs a resource account, which app-only credentials cannot create.

### Get-CsTeamsComplianceRecordingPolicy

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Filter` | [`Service.ListPoliciesFiltered`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingPolicyReadsAndFilters`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |
| `Identity` | [`Service.GetPolicyByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go)<br>[`Service.ListPolicies`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingPolicyCRUDUsesAcceptanceResponses`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingPolicyReadsAndFilters`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go)<br>[`TestAccComplianceRecordingReadOnly`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* policy lifecycle with pinned PowerShell read parity; full release qualification is separate.

### Grant-CsTeamsComplianceRecordingPolicy

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `GrantToGroup` | [`Service.AssignToGroup`](../../../msteams/teams_admin_api/compliance_recording/assignment.go)<br>[`Service.Assign`](../../../msteams/teams_admin_api/compliance_recording/assignment.go) | [`TestUnitComplianceRecordingModernGrant`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |
| `GrantToTenant` | [`Service.AssignTenantDefault`](../../../msteams/teams_admin_api/compliance_recording/assignment.go)<br>[`Service.Assign`](../../../msteams/teams_admin_api/compliance_recording/assignment.go) | [`TestUnitComplianceRecordingModernGrant`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |
| `Identity` | [`Service.AssignToUser`](../../../msteams/teams_admin_api/compliance_recording/assignment.go)<br>[`Service.Assign`](../../../msteams/teams_admin_api/compliance_recording/assignment.go) | [`TestUnitComplianceRecordingModernGrant`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |

Live limitation: Not run live: a compliance recording grant changes call handling for its target, so the lab never grants it. Wire shapes are contract-tested.

### New-CsTeamsComplianceRecordingApplication

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.CreateApplicationByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingApplicationRejections`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |
| `ParentAndRelativeKey` | [`Service.CreateApplicationByID`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingApplicationRejections`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |

Live limitation: Client-secret lab run proved the application routes through service rejections and absence only: a successful application needs a resource account, which app-only credentials cannot create.

### New-CsTeamsComplianceRecordingPairedApplication

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`NewPairedApplication`](../../../msteams/teams_admin_api/compliance_recording/models.go) | [`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingApplicationRejections`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |

Live limitation: Local builder with no service call; contract tests prove it sends nothing.

### New-CsTeamsComplianceRecordingPolicy

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.CreatePolicyByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingPolicyCRUDUsesAcceptanceResponses`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* policy lifecycle with pinned PowerShell read parity; full release qualification is separate.

### Remove-CsTeamsComplianceRecordingApplication

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.DeleteApplicationByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingApplicationRejections`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |
| `ParentAndRelativeKey` | [`Service.DeleteApplicationByID`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingApplicationRejections`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |

Live limitation: Client-secret lab run proved the application routes through service rejections and absence only: a successful application needs a resource account, which app-only credentials cannot create.

### Remove-CsTeamsComplianceRecordingPolicy

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.DeletePolicyByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingPolicyCRUDUsesAcceptanceResponses`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* policy lifecycle with pinned PowerShell read parity; full release qualification is separate.

### Set-CsTeamsComplianceRecordingApplication

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.PatchApplicationByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go)<br>[`Service.UpdateApplicationByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingListEditsApplyToReadValue`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | None recorded |
| `ParentAndRelativeKey` | [`Service.UpdateApplicationByID`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingApplicationRejections`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingListEditsApplyToReadValue`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |

Live limitation: Client-secret lab run proved the application routes through service rejections and absence only: a successful application needs a resource account, which app-only credentials cannot create.

### Set-CsTeamsComplianceRecordingPolicy

Owner: `CALL-25`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `Identity` | [`Service.PatchPolicyByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go)<br>[`Service.UpdatePolicyByIdentity`](../../../msteams/teams_admin_api/compliance_recording/service.go) | [`TestUnitComplianceRecordingPolicyCRUDUsesAcceptanceResponses`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingListEditsApplyToReadValue`](../../../msteams/teams_admin_api/compliance_recording/service_test.go)<br>[`TestUnitComplianceRecordingInvalidInputHasNoHTTP`](../../../msteams/teams_admin_api/compliance_recording/service_test.go) | [`TestAccComplianceRecordingPolicyCRUD`](../../../tests/integration/lab/compliance_recording/compliance_recording_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* policy lifecycle with pinned PowerShell read parity; full release qualification is separate.

## mainline_attendants

Contract and live limitations: [mainline-attendants-contract-notes.md](mainline-attendants-contract-notes.md).

### Get-CsMainlineAttendantAppointmentBookingFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListAppointmentBookingFlows`](../../../msteams/teams_admin_api/mainline_attendants/service.go)<br>[`Service.GetAppointmentBookingFlowByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantAppointmentBookingFlowLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantReads`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go)<br>[`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Get-CsMainlineAttendantFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListFlows`](../../../msteams/teams_admin_api/mainline_attendants/service.go)<br>[`Service.GetFlowByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMissingResourcesAreTypedErrors`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantQuestionAnswerFlowLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMalformedResponsesFailDecode`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantReads`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go)<br>[`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Get-CsMainlineAttendantQuestionAnswerFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListQuestionAnswerFlows`](../../../msteams/teams_admin_api/mainline_attendants/service.go)<br>[`Service.GetQuestionAnswerFlowByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantReads`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go)<br>[`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Get-CsMainlineAttendantSpamDetectionTemplate

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListSpamDetectionTemplates`](../../../msteams/teams_admin_api/mainline_attendants/service.go)<br>[`Service.GetSpamDetectionTemplateByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantSpamDetectionTemplateLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMissingResourcesAreTypedErrors`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMalformedResponsesFailDecode`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantReads`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go)<br>[`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Get-CsMainlineAttendantSupportedLanguages

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetTenantInformation`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantReads`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed the tenant information, list and missing-resource reads with pinned PowerShell tenant information parity; full release qualification is separate.

### Get-CsMainlineAttendantSupportedVoices

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetTenantInformation`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMalformedResponsesFailDecode`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantReads`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed the tenant information, list and missing-resource reads with pinned PowerShell tenant information parity; full release qualification is separate.

### Get-CsMainlineAttendantTenantInformation

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetTenantInformation`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMalformedResponsesFailDecode`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantReads`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed the tenant information, list and missing-resource reads with pinned PowerShell tenant information parity; full release qualification is separate.

### New-CsMainlineAttendantAppointmentBookingFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.CreateAppointmentBookingFlow`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantAppointmentBookingFlowLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantCustomBookingSendsEnumsAsIntegers`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup. The Custom platform needs real API definitions and is contract-tested only; full release qualification is separate.

### New-CsMainlineAttendantQuestionAnswerFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.CreateQuestionAnswerFlow`](../../../msteams/teams_admin_api/mainline_attendants/service.go)<br>[`EncodeKnowledgeBase`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantQuestionAnswerFlowLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantEncodeKnowledgeBaseMatchesTheModule`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### New-CsMainlineAttendantSpamDetectionTemplate

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.CreateSpamDetectionTemplate`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantSpamDetectionTemplateLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantSpamDetectionTemplateSendsCallTargetAsTheModule`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Remove-CsMainlineAttendantAppointmentBookingFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteAppointmentBookingFlowByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantAppointmentBookingFlowLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMissingResourcesAreTypedErrors`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Remove-CsMainlineAttendantQuestionAnswerFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteQuestionAnswerFlowByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantQuestionAnswerFlowLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMalformedResponsesFailDecode`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Remove-CsMainlineAttendantSpamDetectionTemplate

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteSpamDetectionTemplateByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantSpamDetectionTemplateLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMissingResourcesAreTypedErrors`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Set-CsMainlineAttendantAppointmentBookingFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateAppointmentBookingFlowByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantAppointmentBookingFlowLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup. The Custom platform needs real API definitions and is contract-tested only; full release qualification is separate.

### Set-CsMainlineAttendantQuestionAnswerFlow

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateQuestionAnswerFlowByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantQuestionAnswerFlowLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantMissingResourcesAreTypedErrors`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

### Set-CsMainlineAttendantSpamDetectionTemplate

Owner: `CALL-26`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateSpamDetectionTemplateByID`](../../../msteams/teams_admin_api/mainline_attendants/service.go) | [`TestUnitMainlineAttendantSpamDetectionTemplateLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantSpamDetectionTemplateSendsCallTargetAsTheModule`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go)<br>[`TestUnitMainlineAttendantInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/mainline_attendants/service_test.go) | [`TestAccMainlineAttendantLifecycle`](../../../tests/integration/lab/mainline_attendants/mainline_attendants_test.go) |

Live limitation: Client-secret lab run passed an owned msteams-sdk-* spam detection template, Microsoft Bookings appointment booking flow and question and answer flow lifecycle with pinned PowerShell read parity and journaled cleanup; full release qualification is separate.

## resource_accounts

Contract and live limitations: [resource-accounts-contract-notes.md](resource-accounts-contract-notes.md).

### Find-CsOnlineApplicationInstance

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Search`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go)<br>[`TestUnitResourceAccountSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountReads`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run passed the read with pinned PowerShell parity against a tenant without resource accounts; populated reads and full release qualification are separate.

### Get-CsOnlineApplicationInstance

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.List`](../../../msteams/teams_admin_api/resource_accounts/service.go)<br>[`Service.GetByID`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go)<br>[`TestUnitResourceAccountSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountReads`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run passed the read with pinned PowerShell parity against a tenant without resource accounts; populated reads and full release qualification are separate.

### Get-CsOnlineApplicationInstanceAssociation

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetAssociationByID`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go)<br>[`TestUnitResourceAccountSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountReads`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run passed the read with pinned PowerShell parity against a tenant without resource accounts; populated reads and full release qualification are separate.

### Get-CsOnlineApplicationInstanceAssociationStatus

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetAssociationStatusByID`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountReadsUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountReads`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run passed the read with pinned PowerShell parity against a tenant without resource accounts; populated reads and full release qualification are separate.

### New-CsOnlineApplicationInstance

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.Create`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go)<br>[`TestUnitResourceAccountSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountRejectedWrites`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run proved the wire contract through service rejections and unknown-target outcomes only; app-only credentials cannot create or change resource accounts, so successful writes are unqualified.

### New-CsOnlineApplicationInstanceAssociation

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.CreateAssociations`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountAssociationWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountRejectedWrites`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run proved the wire contract through service rejections and unknown-target outcomes only; app-only credentials cannot create or change resource accounts, so successful writes are unqualified.

### Remove-CsOnlineApplicationInstanceAssociation

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteAssociationByID`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountAssociationWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountRejectedWrites`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run proved the wire contract through service rejections and unknown-target outcomes only; app-only credentials cannot create or change resource accounts, so successful writes are unqualified.

### Set-CsOnlineApplicationInstance

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateByID`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go)<br>[`TestUnitResourceAccountSuccessShapesPreserveFields`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountRejectedWrites`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run proved the wire contract through service rejections and unknown-target outcomes only; app-only credentials cannot create or change resource accounts, so successful writes are unqualified.

### Set-CsOnlineVoiceApplicationInstance

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateTelephoneNumberByID`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountRejectedWrites`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run proved the wire contract through service rejections and unknown-target outcomes only; app-only credentials cannot create or change resource accounts, so successful writes are unqualified.

### Sync-CsOnlineApplicationInstance

Owner: `CALL-27`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.SyncByID`](../../../msteams/teams_admin_api/resource_accounts/service.go) | [`TestUnitResourceAccountRejectedWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go)<br>[`TestUnitResourceAccountAssociationWritesUseAcceptanceResponses`](../../../msteams/teams_admin_api/resource_accounts/service_test.go) | [`TestAccResourceAccountRejectedWrites`](../../../tests/integration/lab/resource_accounts/resource_accounts_test.go) |

Live limitation: Client-secret lab run proved the wire contract through service rejections and unknown-target outcomes only; app-only credentials cannot create or change resource accounts, so successful writes are unqualified.

## schedules

Contract and live limitations: [schedules-contract-notes.md](schedules-contract-notes.md).

### Export-CsAutoAttendantHolidays

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.GetHolidaysContent`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleHolidaysUseAcceptanceAndModelResponses`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: The lab tenant has no auto attendant, so the route was proven only as HTTP 404 for a reserved ID; successful holiday bodies are modelled from the module serializers and remain unqualified (CALL-22).

### Get-CsAutoAttendantHolidays

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ListHolidays`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleHolidaysUseAcceptanceAndModelResponses`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: The lab tenant has no auto attendant, so the route was proven only as HTTP 404 for a reserved ID; successful holiday bodies are modelled from the module serializers and remain unqualified (CALL-22).

### Get-CsOnlineSchedule

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.List`](../../../msteams/teams_admin_api/schedules/service.go)<br>[`Service.GetByID`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/schedules/service_test.go)<br>[`TestUnitScheduleListAndBodyOptions`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: Client-secret lab run passed the owned, journaled schedule lifecycle with pinned PowerShell parity; auto attendant association and full release qualification are separate.

### Import-CsAutoAttendantHolidays

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.ImportHolidays`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleHolidaysUseAcceptanceAndModelResponses`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: The lab tenant has no auto attendant, so the route was proven only as HTTP 404 for a reserved ID; successful holiday bodies are modelled from the module serializers and remain unqualified (CALL-22).

### New-CsOnlineDateTimeRange

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.NewDateTimeRange`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleRangeDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: Client-secret lab run passed the service draft validation, including a rejected range; full release qualification is separate.

### New-CsOnlineSchedule

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `FixedScheduleParamSet` | [`Service.Create`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/schedules/service_test.go)<br>[`TestUnitScheduleListAndBodyOptions`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |
| `UnresolvedParamSet` | [`Service.Create`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleInvalidInputFailsBeforeIO`](../../../msteams/teams_admin_api/schedules/service_test.go) | None recorded |
| `WeeklyRecurrentScheduleParamSet` | [`Service.Create`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: Client-secret lab run passed the owned, journaled schedule lifecycle with pinned PowerShell parity; auto attendant association and full release qualification are separate.

### New-CsOnlineTimeRange

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.NewTimeRange`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleRangeDraftsUseAcceptanceResponses`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: Client-secret lab run passed the service draft validation, including a rejected range; full release qualification is separate.

### Remove-CsOnlineSchedule

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.DeleteByID`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: Client-secret lab run passed the owned, journaled schedule lifecycle with pinned PowerShell parity; auto attendant association and full release qualification are separate.

### Set-CsOnlineSchedule

Owner: `CALL-28`. Implementation: `in-progress`; contract: `metadata-only`; live: `not-run`.

| Parameter set | Go symbols | Contract test mappings | Live test mappings |
| --- | --- | --- | --- |
| `__AllParameterSets` | [`Service.UpdateByID`](../../../msteams/teams_admin_api/schedules/service.go) | [`TestUnitScheduleLifecycleUsesAcceptanceResponses`](../../../msteams/teams_admin_api/schedules/service_test.go)<br>[`TestUnitScheduleListAndBodyOptions`](../../../msteams/teams_admin_api/schedules/service_test.go) | [`TestAccScheduleLifecycle`](../../../tests/integration/lab/schedules/schedules_test.go) |

Live limitation: Client-secret lab run passed the owned, journaled schedule lifecycle with pinned PowerShell parity; auto attendant association and full release qualification are separate.
