# POL90 current test truth audit

Source-only audit for POL90/#184 at supplied baseline `bb56e7b`, dated 2026-11-26. Scope is exactly the 60 `coverage.records` commands and 95 parameter sets whose `implementation_owner` is POL-22, POL-23, POL-24, POL-27, POL-28, POL-30 or POL-41: assignment_migration, federation, meeting_branding, policy_assignments, policy_packages, recording_messages and teams_custom_banner_text. [Scope authority](../../../contracts/coverage/8.1.0.json). This review read mapped function bodies, their callback dispatch and package runners/helpers. No Go, live tests, credentials, tenant actions or repository operations ran; a later addendum explicitly records offline pinned module metadata/IL inspection. Thus an assertion is described as present/reachable, never newly passing.

## Findings that change qualification

1. The typed-provider request tables are primarily **400 with blank bodies**, not success/model tests. In branding, recording, banner, ACS and Teams external access, the final assignment to `c.Want` discards earlier JSON/decode/forbidden expectations. The GPA helper also unconditionally replaces its switch-selected expectations. Tenant federation instead directly expects APIError400. The runner will prove method/path/query/body/header/attempt order and APIError400, but none of the discarded result assertions. Names such as `read-empty`, `read-malformed`, `list-envelope`, `read-future-enum`, `group-accepted-opaque` and `acs-null-list` do not establish those behaviors. See the exact mapped-function appendix and each package's `helpers_test.go`/`service_test.go`.
2. Ordinary assignment, explicit-region OCE and package request vectors deliberately return **403 with blank bodies**. `TestUnitAssignmentInputObjectSharesOpaqueUserRoute` puts source-provenance checks behind `e == nil`, unreachable after its 403. `TestUnitOCEOptionalQueryAndScopedEscaping` (additional, not coverage-mapped) stores intended replies `[]`, `{}`, `null`, `""` but passes 403/empty to `oceWire`; nullable refresh success checks are therefore unreachable. Package `invoke` contains partial/domain failure/missing-ID checks keyed by `user-partial`, `group-domain-error`, `user-domain-error`, `batch-error`, `batch-missing-id`; **none of those names occurs in `requestCases()`**, and its type assertion would skip a nil/wrong result type. Add explicit cases and hard type/non-nil assertions before claiming these contracts. [Assignment callback](../../../msteams/teams_admin_api/policy_assignments/failure_test.go), [OCE callback](../../../msteams/teams_admin_api/policy_assignments/oce_service_test.go), [package runner/callback](../../../msteams/teams_admin_api/policy_packages/service_test.go).
3. A source fixture pointer is not capture provenance. Federation helpers force `c.Source` to `get_global.json` even for authored request tables; GPA forces `list_400.json` even for authored failures and local guards. Branding/recording/banner helpers often default to JSON fixture sources, but inline request vectors and adversarial statuses remain authored. Constructor wire literals are inline expectations. Keep **captured response bytes**, **authored offline adversarial/request cases**, **local builders**, **source-only live paths**, and **a current execution record** distinct.
4. Cross-set/function references overstate precision: filtering functions exercise collection retrieval, not Identity get. Both external/branding Grant `Identity` mappings have only invalid-union local checks, not a valid dispatch. Branding `TestUnitBrandingObjectEditsMatchTypedCoreExtension`, mapped under all three New asset commands, calls list `Apply` and internally maps Set policy/New theme; it never calls New slate. Its local builder's reaction rows map New theme while invoking `NewReaction`. Domain ownership function is mapped to AllowAll but does not invoke that constructor (the other constructor function does). Correct the reference scope, then retain supplementary composition evidence separately.
5. Empty success bodies must be interpreted per operation. Captured branding/external CRUD has blank Create201/Patch204/Delete204 and independent GET200/readback/GET404. This is valid acceptance-plus-readback evidence, not a nonempty created result. Recording/banner auto-create tests intentionally cover authored empty204 success and malformed200 accepted-but-decode-error with generated identity retained. They do not establish successful nonempty typed create decoding. GPA authored malformed202 is real decode-error coverage; ordinary request400 is not.

## What the shared runner actually checks

[`contracttest.Run` and `compareObservation`](../../../internal/contracttest/runner.go) construct the real root SDK with an in-memory transport, compare every scheduled request in order, reject missing/unexpected attempts, and compare only selected Expected fields. `JSON == nil` skips result comparison; `Body == nil` skips raw response comparison. With neither Error nor APIError, an unexpected callback error fails. Error checks use `errors.Is`; APIError checks use full `reflect.DeepEqual`. Local cases require no exchanges and zero provider acquisitions. Package `run` helpers derive expected status from the last reply; that does not add a result or cancellation assertion. Several failure callbacks deliberately validate error/status then return Response without Err; those checks are real callback assertions, not runner-normalized-error comparisons. [JSON comparison](../../../internal/contracttest/compare.go) preserves numeric spelling/type, object field presence and array order. This audit does not infer a successful test execution from source.

## Genuine current fixture evidence beyond mapped references

- Branding and external access each have `service_crud_response_test.go::TestUnitOwnedCRUDUsesAcceptanceResponses`, loading Create201, create_readback, Update204, update_readback, owned_list, Delete204 and get_deleted404 fixtures. Each callback verifies acceptance, non-default typed reads, separate sparse update, omission preservation, owned listing and typed absence. Branding also compares complete Before/After RawFields. They are **not in the 95 mapped references**; add appropriate references, rather than saying CRUD has no offline evidence. [Branding replay](../../../msteams/teams_admin_api/meeting_branding/service_crud_response_test.go), [external replay](../../../msteams/teams_admin_api/external_access_policy/service_crud_response_test.go), [branding live CRUD source](../../../tests/integration/lab/meeting_branding/meeting_branding_crud_test.go).
- Ordinary assignment has captured ApplicationAccessPolicy group create/update/delete204 and get/list/deleted404 plus user before/assigned/restored snapshots. `group_response_test.go` verifies complete raw and typed group fields (including priority, creator and time), empty204 HTTPAccepted with no invented handle, and absence; `response_test.go` verifies user typed policy fields/full raw snapshots with string and InputObject forms. These are additional real response replays, not broad OCE/batch/package proof. [Group replay](../../../msteams/teams_admin_api/policy_assignments/group_response_test.go), [user replay](../../../msteams/teams_admin_api/policy_assignments/response_test.go).
- Federation Global JSON pairs and branding Global/list have actual response replay; recording/banner currently have only captured `list.json`. An empty array supports collection decoding/filtering, not localization/banner field types or item selectors. [Recording replay](../../../msteams/teams_admin_api/recording_messages/response_test.go), [banner replay](../../../msteams/teams_admin_api/teams_custom_banner_text/response_test.go). GPA `list_400.json` is replayed as APIError400 plus `ErrTenantCapabilityUnavailable`; `access_outcome_test.go` has the specific 40003 status-message comparison path. [GPA replay](../../../msteams/teams_admin_api/assignment_migration/response_test.go), [access outcome](../../../tests/integration/lab/assignment_migration/access_outcome_test.go).
- File presence and deterministic replay establish retained response evidence, not the full execution context. Current lab-run provenance is a separate integration-plan responsibility. Existing `*-lab-acceptance.md` Nov21 “all403/no CRUD run” prose is stale against these artifacts and cannot erase them or qualify uncovered scopes.

## Implementable offline gaps

| Area | Concrete change and assertion | External/source unknown kept separate |
|---|---|---|
| Mapped request tables | Split rejected-wire cases from response cases; remove expectation overwrites, supply intended statuses/bodies explicitly; assert result nonnil and exact concrete type before field checks. Preserve current 400/403 route checks. | Captured successful body only where available; authored cases must be labeled authored. |
| Ordinary assignments | Reference existing group/user response tests to matching Get/Get1, NewExpanded, Remove and Get/GetViaIdentity sets. Add batch list/Get/Create successful typed handle/no-handle, malformed/overflow/future/raw-field cases; source-provenance array test with reachable200. Get2 all/by-type has no captured service response here. | Batch completion/per-user effective policy convergence and actual source precedence require source/tenant evidence. |
| OCE | Add valid200 status/definition/list/action responses with exact types, region-specific results; NOAM miss then EMEA success; mixed regional success/error retention; empty vs null action values; cancel after first regional attempt and during callbacks, asserting attempt order/no later requests. Retain all403/all404 and malformed200 tests already present. | True regional availability/eligibility, action side effects and server schema are not established by a 403. |
| Packages | Add explicit delegated success list/item/user/recommendation/create/update/delete bodies; user/group/batch partial/domain failures and missing handle; assert HTTPAccepted independently of domain failures. Test wrong/null result types and future failure fields. Keep application-mode local refusal separate. | No package capture directory or successful delegated lab evidence was found in this scope; entitled delegated behavior/effective assignments remain external. |
| GPA migration | Add well-formed list/Get/recommendation/conflict/Create/Clear success, open textual/numeric states and exact typed counters with future RawFields; malformed/null/envelope reads must actually be200. Preserve malformed202/no-replay and local-cancel tests. | Current capture is a capability40003 failure, not a migration job; successful service semantics and consequential effective migration/cleanup need external evidence. |
| Branding | Reference existing owned CRUD replay and live source to exact sets. Add nested asset update success/readback and comparison-success result, valid Grant Identity dispatch and each scope result. Add denied/malformed post-read cases where absent, and cancellation after preparation before PATCH with no patch replay. | Premium rendering, assets fetched by meetings and user/group/tenant effective behavior remain external. |
| Recording/banner | Keep empty captured collection tests; add authored nonempty collection/item/Get-by-Identity/relative result types, successful Create/Patch/Delete selector forms, malformed accepted result/partial metadata, successful composed list/text edit and post-read patch failure. Add cancellation between read and patch; assert preserved preparation/generated ID and no retry. | Empty capture cannot specify successful item fields; actual localization/banner CRUD/persistence/display remains unqualified. |
| Federation | Reference external CRUD replay; retain ACS/external-users captured preparation + denied400 patch. Add successful composed update result/After/snapshot independence, reachable typed Global Get, null/empty/future domains and valid assignment union routes; cancel between stages. Tenant PUT normalization currently proves request400 only. | Server-normalized domains, tenant/ACS/external-user persistence and access effects require actual reads/authorized observations. |
| Recovery | Add offline journal/cleanup callback tests for each implemented family: owned-name and binding rejection, ambiguous create,404 deletion, unrelated resources untouched, canceled caller cleanup via independent context, recoverable failure retained and second recovery idempotent. Existing source cleanup registration is not evidence that every branch ran. | Existing retained CRUD fixture absence proves a narrow lifecycle, not all crash/ambiguous-response recovery branches. |

## Exact command/set matrix

Every row below is one of the 95 sets; all 60 commands appear. T identifiers resolve to exact source function and behavioral audit below. `C` is coverage's contract reference, `L` its live reference (a path, not a passing execution). Additional real response evidence above must be reconciled into mappings. Reused broad functions do not imply every internal subcase targets the row's set.

| Command | Exact parameter set | C | L |
|---|---|---|---|
| Get-CsBatchOperationDefinition | `GetBatchOperationDefinition` | T23 | — |
| Get-CsBatchOperationDefinition | `GetBatchOperationDefinitionForRegion` | T24 | — |
| Get-CsBatchOperationStatus | `GetAllBatchOperationStatus` | T22 | T54 |
| Get-CsBatchOperationStatus | `GetAllBatchOperationStatusForRegion` | T24 | — |
| Get-CsBatchOperationStatus | `GetBatchOperationStatus` | T23 | — |
| Get-CsBatchOperationStatus | `GetBatchOperationStatusForRegion` | T24 | — |
| Get-CsBatchPolicyAssignmentOperation | `Get` | T25 | T55 |
| Get-CsBatchPolicyAssignmentOperation | `Get1` | T25 | — |
| Get-CsExternalAccessPolicy | `Filter` | T13, T11 | T52 |
| Get-CsExternalAccessPolicy | `Identity` | T13, T11 | T52 |
| Get-CsGroupPolicyAssignment | `Get` | T25 | — |
| Get-CsGroupPolicyAssignment | `Get1` | T25 | — |
| Get-CsGroupPolicyAssignment | `Get2` | T25 | — |
| Get-CsGroupPolicyAssignments | `GetAllGroupPolicyAssignments` | T24 | — |
| Get-CsGroupPolicyAssignments | `GetGroupPolicyAssignment` | T24 | — |
| Get-CsGroupPolicyAssignments | `GetGroupPolicyAssignments` | T24 | — |
| Get-CsGroupPolicyAssignments | `GetPolicyAssignmentForGroups` | T24 | — |
| Get-CsPolicyPackage | `Get` | T27 | T56 |
| Get-CsPolicyPackage | `Get1` | T27 | — |
| Get-CsPolicyPackage | `GetViaIdentity` | T26 | — |
| Get-CsTeamsAcsFederationConfiguration | `Filter` | T03, T01 | T49 |
| Get-CsTeamsAcsFederationConfiguration | `Identity` | T03, T01 | T49 |
| Get-CsTeamsCustomBannerText | `Filter` | T38 | T58 |
| Get-CsTeamsCustomBannerText | `Identity` | T39 | T58 |
| Get-CsTeamsCustomBannerText | `ParentAndRelativeKey` | T39 | T58 |
| Get-CsTeamsExternalAccessConfiguration | `Filter` | T42, T40 | T59 |
| Get-CsTeamsExternalAccessConfiguration | `Identity` | T42, T40 | T59 |
| Get-CsTeamsMeetingBrandingPolicy | `Filter` | T17 | T53 |
| Get-CsTeamsMeetingBrandingPolicy | `Identity` | T20 | T53 |
| Get-CsTeamsRecordingAndTranscriptionCustomMessage | `Filter` | T33 | T57 |
| Get-CsTeamsRecordingAndTranscriptionCustomMessage | `Identity` | T34 | T57 |
| Get-CsTeamsRecordingAndTranscriptionCustomMessage | `ParentAndRelativeKey` | T34 | T57 |
| Get-CsTenantFederationConfiguration | `Filter` | T45, T46 | T60 |
| Get-CsTenantFederationConfiguration | `Identity` | T45, T46 | T60 |
| Get-CsUserPolicyAssignment | `Get` | T25 | — |
| Get-CsUserPolicyAssignment | `GetViaIdentity` | T21 | — |
| Get-CsUserPolicyPackage | `Get` | T27 | — |
| Get-CsUserPolicyPackageRecommendation | `Get` | T27 | — |
| Get-DirectToGroupAssignmentsMigrationStatus | `__AllParameterSets` | T07, T05 | T50 |
| Get-GPAGroupMembers | `__AllParameterSets` | T24 | — |
| Get-GPAUserMembership | `__AllParameterSets` | T24 | — |
| Get-GroupAssignmentRecommendationsPerPolicyName | `__AllParameterSets` | T07, T05 | — |
| Get-GroupAssignmentRecommendationsPerPolicyType | `__AllParameterSets` | T07, T05 | — |
| Get-GroupPolicyAssignmentConflict | `__AllParameterSets` | T07, T05 | — |
| Grant-CsExternalAccessPolicy | `GrantToGroup` | T13, T08 | — |
| Grant-CsExternalAccessPolicy | `GrantToTenant` | T13, T08 | — |
| Grant-CsExternalAccessPolicy | `GrantToUser` | T13, T08 | — |
| Grant-CsExternalAccessPolicy | `Identity` | T13, T08 | — |
| Grant-CsGroupPolicyPackageAssignment | `RequiredPolicyList` | T27 | — |
| Grant-CsTeamsMeetingBrandingPolicy | `GrantToGroup` | T20 | — |
| Grant-CsTeamsMeetingBrandingPolicy | `GrantToTenant` | T20 | — |
| Grant-CsTeamsMeetingBrandingPolicy | `GrantToUser` | T20 | — |
| Grant-CsTeamsMeetingBrandingPolicy | `Identity` | T19 | — |
| Grant-CsUserPolicyPackage | `GrantExpanded` | T27 | — |
| Invoke-ClearDirectToGroupAssignmentMigration | `__AllParameterSets` | T07, T05, T06, T04 | — |
| Invoke-CsRefreshGroupUsers | `__AllParameterSets` | T24 | — |
| Invoke-CsReprocessBatchOperation | `__AllParameterSets` | T24 | — |
| Invoke-CsReprocessGroupPolicyAssignment | `__AllParameterSets` | T24 | — |
| Invoke-StartDirectToGroupAssignmentMigration | `__AllParameterSets` | T07, T05, T06, T04 | — |
| New-CsBatchPolicyAssignmentOperation | `NewExpanded` | T25 | — |
| New-CsBatchPolicyPackageAssignmentOperation | `NewExpanded` | T27 | — |
| New-CsCustomPolicyPackage | `RequiredPolicyList` | T27 | — |
| New-CsEdgeAllowAllKnownDomains | `__AllParameterSets` | T44, T43 | — |
| New-CsEdgeAllowList | `__AllParameterSets` | T44, T43 | — |
| New-CsEdgeDomainPattern | `__AllParameterSets` | T44, T43 | — |
| New-CsExternalAccessPolicy | `Identity` | T13, T09 | T51 |
| New-CsGroupPolicyAssignment | `NewExpanded` | T25 | — |
| New-CsTeamsCustomBannerText | `Identity` | T39 | — |
| New-CsTeamsCustomBannerText | `ParentAndRelativeKey` | T39, T35 | — |
| New-CsTeamsMeetingBackgroundImage | `Identity` | T14, T15 | — |
| New-CsTeamsMeetingBrandingPolicy | `Identity` | T20 | — |
| New-CsTeamsMeetingBrandingTheme | `Identity` | T14, T15 | — |
| New-CsTeamsNdiAssuranceSlate | `Identity` | T14, T15 | — |
| New-CsTeamsRecordingAndTranscriptionCustomMessage | `Identity` | T34 | — |
| New-CsTeamsRecordingAndTranscriptionCustomMessage | `ParentAndRelativeKey` | T34, T28 | — |
| New-CsTeamsRecordingAndTranscriptionLocalizationCustomMessage | `Identity` | T30 | — |
| Remove-CsCustomPolicyPackage | `Remove` | T27 | — |
| Remove-CsCustomPolicyPackage | `RemoveViaIdentity` | T26 | — |
| Remove-CsExternalAccessPolicy | `Identity` | T13 | T51 |
| Remove-CsGroupPolicyAssignment | `Remove` | T25 | — |
| Remove-CsTeamsCustomBannerText | `Identity` | T39 | — |
| Remove-CsTeamsCustomBannerText | `ParentAndRelativeKey` | T39 | — |
| Remove-CsTeamsMeetingBrandingPolicy | `Identity` | T20 | — |
| Remove-CsTeamsRecordingAndTranscriptionCustomMessage | `Identity` | T34 | — |
| Remove-CsTeamsRecordingAndTranscriptionCustomMessage | `ParentAndRelativeKey` | T34, T32 | — |
| Set-CsExternalAccessPolicy | `Identity` | T12, T10 | T51 |
| Set-CsTeamsAcsFederationConfiguration | `Identity` | T02 | — |
| Set-CsTeamsCustomBannerText | `Identity` | T39, T36 | — |
| Set-CsTeamsCustomBannerText | `ParentAndRelativeKey` | T39, T37 | — |
| Set-CsTeamsExternalAccessConfiguration | `Identity` | T41 | — |
| Set-CsTeamsMeetingBrandingPolicy | `Identity` | T20, T16, T18 | — |
| Set-CsTeamsRecordingAndTranscriptionCustomMessage | `Identity` | T34, T29 | — |
| Set-CsTeamsRecordingAndTranscriptionCustomMessage | `ParentAndRelativeKey` | T34, T31 | — |
| Set-CsTenantFederationConfiguration | `Identity` | T48, T47 | — |
| Update-CsCustomPolicyPackage | `RequiredPolicyList` | T27 | — |

## Mapped function appendix

All mapped functions were located. Each summary follows callback dispatch and actual replies, including helpers that set status/source or replace expectations.

| ID | Function source | Reachable evidence and limit |
|---|---|---|
| T01 | [TestUnitProviderFilteringIsLocalAndGlobalOrderIsExactFromRealResponses](../../../msteams/teams_admin_api/acs_federation_configuration/filter_test.go#L18) | 200 captured list replay; compares raw body and complete filtered JSON/order, Global/missing/whitespace and conditional Tag case. Only collection/filter path, not item selector. |
| T02 | [TestUnitComposedUpdatePreservesReadIdentityAndSparseFields](../../../msteams/teams_admin_api/acs_federation_configuration/service_test.go#L62) | Captured GET200 preparation + authored PATCH400; preserves Before/ReadResponse, sparse body, PatchInvoked true, HTTPAccepted false, After nil. |
| T03 | [TestUnitTypedProviderDispatch](../../../msteams/teams_admin_api/acs_federation_configuration/service_test.go#L17) | Controlled 400/empty request dispatch; no successful typed value. Earlier expected JSON/decode assertions are overwritten (tenant dispatch instead sets APIError400 directly). Banner retains its separate no-HTTP invalid cases. |
| T04 | [TestUnitGPAMalformedAcceptedResponsesStayErrorsWithoutResubmission](../../../msteams/teams_admin_api/assignment_migration/boundaries_test.go#L82) | Start/clear POST202 malformed/empty/wrong-type bodies -> ErrDecode, one attempt; no well-formed accepted job. |
| T05 | [TestUnitGPARejectsInvalidInputBeforeCredentialAcquisition](../../../msteams/teams_admin_api/assignment_migration/boundaries_test.go#L16) | Local validation with zero exchanges/credential acquisition; representative cancellation/backend/service guards. This is not valid command success. |
| T06 | [TestUnitGPAErrorsRetainHTTPStatusAndDoNotReplaySubmissions](../../../msteams/teams_admin_api/assignment_migration/service_test.go#L427) | Start/clear POST400/401/403/404/409/429/502/503; normalized status/redaction and no replay via one expected exchange. |
| T07 | [TestUnitTypedGPARequestHelper](../../../msteams/teams_admin_api/assignment_migration/service_test.go#L98) | Every table reply is 400/empty; final Expected APIError400 overwrites JSON/forbidden/decode intentions. |
| T08 | [TestUnitAssignmentsAndMutationsRejectInvalidInputBeforeTokens](../../../msteams/teams_admin_api/external_access_policy/failure_test.go#L83) | Local validation with zero exchanges/credential acquisition; representative cancellation/backend/service guards. This is not valid command success. |
| T09 | [TestUnitCreateComparisonIsExplicitAndDoesNotCloneGlobal](../../../msteams/teams_admin_api/external_access_policy/failure_test.go#L53) | GET200/403 comparison followed by POST400; Global skips comparison. Separate comparison/submission outcome, no create success. |
| T10 | [TestUnitUpdateFailurePreservesPreparationAndNeverReplays](../../../msteams/teams_admin_api/external_access_policy/failure_test.go#L20) | Read403; captured GET200 + PATCH429/503; accepted PATCH200 malformed -> ErrDecode/HTTPAccepted true. Preparation retained, payload redacted, ordered attempt budget. |
| T11 | [TestUnitProviderFilteringIsLocalAndGlobalOrderIsExactFromRealResponses](../../../msteams/teams_admin_api/external_access_policy/filter_test.go#L18) | 200 captured list replay; compares raw body and complete filtered JSON/order, Global/missing/whitespace and conditional Tag case. Only collection/filter path, not item selector. |
| T12 | [TestUnitExternalAccessComposedUpdate](../../../msteams/teams_admin_api/external_access_policy/service_test.go#L116) | Captured GET200 preparation + authored PATCH400; preserves Before/ReadResponse, sparse body, PatchInvoked true, HTTPAccepted false, After nil. |
| T13 | [TestUnitExternalAccessTypedProviderDispatch](../../../msteams/teams_admin_api/external_access_policy/service_test.go#L17) | Controlled 400/empty request dispatch; no successful typed value. Earlier expected JSON/decode assertions are overwritten (tenant dispatch instead sets APIError400 directly). Banner retains its separate no-HTTP invalid cases. |
| T14 | [TestUnitBrandingLocalBuildersMatchTypedSetters](../../../msteams/teams_admin_api/meeting_branding/assets_test.go#L17) | No-HTTP builders: background image/slate/reaction/theme URI and output JSON; reaction rows are labeled New-CsTeamsMeetingBrandingTheme although they call NewReaction. |
| T15 | [TestUnitBrandingObjectEditsMatchTypedCoreExtension](../../../msteams/teams_admin_api/meeting_branding/assets_test.go#L301) | No-HTTP Apply on background-image/reaction lists, duplicates/order/rejection; mappings are Set policy or New theme, not New background image/slate. Does not call policy mutation. |
| T16 | [TestUnitBrandingAssetInputsUseRealPreparationOnDeniedPatch](../../../msteams/teams_admin_api/meeting_branding/composition_test.go#L21) | Captured preparation GET200 + PATCH403; empty/append image and explicit null field; complete Before snapshot retained. No successful nested persistence. |
| T17 | [TestUnitBrandingListFilteringPreservesExactGlobalOrderingFromRealResponses](../../../msteams/teams_admin_api/meeting_branding/composition_test.go#L142) | 200 captured list replay; compares raw body and complete filtered JSON/order, Global/missing/whitespace and conditional Tag case. Only collection/filter path, not item selector. |
| T18 | [TestUnitBrandingUpdatePreparationFailures](../../../msteams/teams_admin_api/meeting_branding/composition_test.go#L59) | Read403/204-empty/200-malformed stop before patch. Branding also captured GET200 + PATCH403/429/503 and complete Before snapshot. Recording/banner define no post-read success cases. |
| T19 | [TestUnitBrandingInvalidInputRejectsBeforeCredentials](../../../msteams/teams_admin_api/meeting_branding/failure_test.go#L18) | Local validation with zero exchanges/credential acquisition; representative cancellation/backend/service guards. This is not valid command success. |
| T20 | [TestUnitBrandingTypedProviderRequests](../../../msteams/teams_admin_api/meeting_branding/service_test.go#L62) | Controlled 400/empty request dispatch; no successful typed value. Earlier expected JSON/decode assertions are overwritten (tenant dispatch instead sets APIError400 directly). Banner retains its separate no-HTTP invalid cases. |
| T21 | [TestUnitAssignmentInputObjectSharesOpaqueUserRoute](../../../msteams/teams_admin_api/policy_assignments/failure_test.go#L132) | 403 request route only; effective-policy/source assertions are behind e == nil and never reached. |
| T22 | [TestUnitOCERegionalListRetainsEveryFailedAttempt](../../../msteams/teams_admin_api/policy_assignments/oce_batch_test.go#L51) | Three 403 list attempts retained with errors and zero values; no mixed/partial regional success. |
| T23 | [TestUnitOCERegionalLookupRetainsFailedAttemptsAndOrder](../../../msteams/teams_admin_api/policy_assignments/oce_batch_test.go#L26) | Three 404 attempts NOAM/EMEA/APAC retained for status and definition; nil value. No regional success or cancellation. |
| T24 | [TestUnitOCEDirectRequestVectors](../../../msteams/teams_admin_api/policy_assignments/oce_service_test.go#L92) | Application/delegated direct routes all 403/empty: explicit region, groups/members/membership/actions. No result decode. |
| T25 | [TestUnitAssignmentRequestVectors](../../../msteams/teams_admin_api/policy_assignments/service_test.go#L149) | Application/delegated discovery + 403; batch/group/user/CRUD query/body dispatch. No successful result decode. |
| T26 | [TestUnitPackageInputObjectVariantsUsePackageName](../../../msteams/teams_admin_api/policy_packages/service_test.go#L208) | GetViaIdentity and RemoveViaIdentity name escaping with discovery + 403; no object result. |
| T27 | [TestUnitPackageRequestVectors](../../../msteams/teams_admin_api/policy_packages/service_test.go#L189) | Delegated discovery + 403 for list/get/user/recommendations/CRUD/group/user/batch routes; empty create DTO rejected locally. No typed partial/success response. |
| T28 | [TestUnitRecordingAutoCreateRetainsFreshIDWithoutReplay](../../../msteams/teams_admin_api/recording_messages/auto_create_test.go#L23) | Authored in-memory sender: two fresh UUIDs/correlations per status; 204 success, 403/429 failure, 200 malformed accepted decode error; no replay. No captured nonempty created model. |
| T29 | [TestUnitRecordingListEditsStopOnPreparationDenial](../../../msteams/teams_admin_api/recording_messages/composition_test.go#L55) | GET403 means Before nil/PatchInvoked false/HTTPAccepted false; list/text edit never materializes or reaches PATCH. |
| T30 | [TestUnitRecordingLocalBuilderHasNoCredentialsOrHTTP](../../../msteams/teams_admin_api/recording_messages/composition_test.go#L66) | No-HTTP localization constructor exact nullable output; Future-Language stays open. |
| T31 | [TestUnitRecordingUpdatePreparationFailures](../../../msteams/teams_admin_api/recording_messages/composition_test.go#L17) | Read403/204-empty/200-malformed stop before patch. Branding also captured GET200 + PATCH403/429/503 and complete Before snapshot. Recording/banner define no post-read success cases. |
| T32 | [TestUnitRecordingMutationErrorsKeepStatusAndNeverReplay](../../../msteams/teams_admin_api/recording_messages/failure_test.go#L69) | Create/patch/delete400/401/403/429/503, normalized status/redaction and one request per write. |
| T33 | [TestUnitRecordingListFiltersUseRealResponses](../../../msteams/teams_admin_api/recording_messages/service_test.go#L113) | 200 captured list replay; compares raw body and complete filtered JSON/order, Global/missing/whitespace and conditional Tag case. Only collection/filter path, not item selector. |
| T34 | [TestUnitRecordingTypedProviderRequests](../../../msteams/teams_admin_api/recording_messages/service_test.go#L22) | Controlled 400/empty request dispatch; no successful typed value. Earlier expected JSON/decode assertions are overwritten (tenant dispatch instead sets APIError400 directly). Banner retains its separate no-HTTP invalid cases. |
| T35 | [TestUnitBannerAutoCreateRetainsFreshIDWithoutReplay](../../../msteams/teams_admin_api/teams_custom_banner_text/auto_create_test.go#L23) | Authored in-memory sender: two fresh UUIDs/correlations per status; 204 success, 403/429 failure, 200 malformed accepted decode error; no replay. No captured nonempty created model. |
| T36 | [TestUnitBannerTextStopsOnPreparationDenial](../../../msteams/teams_admin_api/teams_custom_banner_text/composition_test.go#L55) | GET403 means Before nil/PatchInvoked false/HTTPAccepted false; list/text edit never materializes or reaches PATCH. |
| T37 | [TestUnitBannerUpdatePreparationFailures](../../../msteams/teams_admin_api/teams_custom_banner_text/composition_test.go#L17) | Read403/204-empty/200-malformed stop before patch. Branding also captured GET200 + PATCH403/429/503 and complete Before snapshot. Recording/banner define no post-read success cases. |
| T38 | [TestUnitBannerFiltersUseRealResponses](../../../msteams/teams_admin_api/teams_custom_banner_text/service_test.go#L121) | 200 captured list replay; compares raw body and complete filtered JSON/order, Global/missing/whitespace and conditional Tag case. Only collection/filter path, not item selector. |
| T39 | [TestUnitBannerTypedProviderAndSelectors](../../../msteams/teams_admin_api/teams_custom_banner_text/service_test.go#L18) | Controlled 400/empty request dispatch; no successful typed value. Earlier expected JSON/decode assertions are overwritten (tenant dispatch instead sets APIError400 directly). Banner retains its separate no-HTTP invalid cases. |
| T40 | [TestUnitProviderFilteringIsLocalAndGlobalOrderIsExactFromRealResponses](../../../msteams/teams_admin_api/teams_external_access_configuration/filter_test.go#L18) | 200 captured list replay; compares raw body and complete filtered JSON/order, Global/missing/whitespace and conditional Tag case. Only collection/filter path, not item selector. |
| T41 | [TestUnitComposedUpdatePreservesReadIdentityAndSparseFields](../../../msteams/teams_admin_api/teams_external_access_configuration/service_test.go#L62) | Captured GET200 preparation + authored PATCH400; preserves Before/ReadResponse, sparse body, PatchInvoked true, HTTPAccepted false, After nil. |
| T42 | [TestUnitTypedProviderDispatch](../../../msteams/teams_admin_api/teams_external_access_configuration/service_test.go#L17) | Controlled 400/empty request dispatch; no successful typed value. Earlier expected JSON/decode assertions are overwritten (tenant dispatch instead sets APIError400 directly). Banner retains its separate no-HTTP invalid cases. |
| T43 | [TestUnitAllowListOwnsOriginalOrderAndRejectsInvalidZeroChoice](../../../msteams/teams_admin_api/tenant_federation_configuration/domains_test.go#L79) | Local input ownership, duplicate/order serialization, invalid zero AllowedDomains and empty pattern. Does not exercise allow-all constructor itself. |
| T44 | [TestUnitDomainConstructorsAreLocal](../../../msteams/teams_admin_api/tenant_federation_configuration/domains_test.go#L14) | No-HTTP domain-pattern/list/allow-all exact JSON; no service persistence claim. |
| T45 | [TestUnitTenantFederationTypedDispatch](../../../msteams/teams_admin_api/tenant_federation_configuration/service_test.go#L21) | Controlled 400/empty request dispatch; no successful typed value. Earlier expected JSON/decode assertions are overwritten (tenant dispatch instead sets APIError400 directly). Banner retains its separate no-HTTP invalid cases. |
| T46 | [TestUnitTenantFiltersUseRealResponses](../../../msteams/teams_admin_api/tenant_federation_configuration/service_test.go#L125) | 200 captured list replay; compares raw body and complete filtered JSON/order, Global/missing/whitespace and conditional Tag case. Only collection/filter path, not item selector. |
| T47 | [TestUnitTenantInvalidInputsDoNotAcquireCredentials](../../../msteams/teams_admin_api/tenant_federation_configuration/service_test.go#L229) | Local validation with zero exchanges/credential acquisition; representative cancellation/backend/service guards. This is not valid command success. |
| T48 | [TestUnitTenantUpdateMatchesTypedNormalization](../../../msteams/teams_admin_api/tenant_federation_configuration/service_test.go#L103) | Authored normalization inputs vs inline expected bodies then PUT400: domain choices/operators/omission/false/empty; rejected model count only, covered separately by invalid-input function. |
| T49 | [TestAccACSFederationConfigurationReadOnly](../../../tests/integration/lab/acs_federation_configuration/federation_test.go#L14) | RunFederationGlobal callback: Global get + Global-filtered list complete fields, optional selected nullable scalar oracle; error stops downstream stages; no updates/grants. |
| T50 | [TestAccAssignmentMigrationReadOnly](../../../tests/integration/lab/assignment_migration/assignment_migration_test.go#L87) | List + every returned job GetByID (not capped at five), ID/state projection and optional oracle. List error stops all; no recommendations/conflicts/start/clear. |
| T51 | [TestAccExternalAccessPolicyCRUD](../../../tests/integration/lab/external_access_policy/external_access_policy_test.go#L17) | Gated journaled owned unassigned CRUD, non-default readbacks, omitted boolean, list and independent 404; optional selected scalar oracle. Source path and checked-in fixture evidence, not a new run. |
| T52 | [TestAccExternalAccessPolicyReadOnly](../../../tests/integration/lab/external_access_policy/federation_test.go#L14) | RunFederationGlobal callback: Global get + Global-filtered list complete fields, optional selected nullable scalar oracle; error stops downstream stages; no updates/grants. |
| T53 | [TestAccMeetingBrandingReadOnly](../../../tests/integration/lab/meeting_branding/meeting_branding_test.go#L99) | Global get + Global-filtered list complete fields; optional six scalar oracle, no nested assets/grants. |
| T54 | [TestAccPolicyAssignmentsOCEReadOnly](../../../tests/integration/lab/policy_assignments/policy_assignments_test.go#L51) | Lists status across regions; on error reports only last HTTP/attempt count. Success only checks nonnil wrapper, not typed values/parity. |
| T55 | [TestAccPolicyAssignmentsReadOnly](../../../tests/integration/lab/policy_assignments/policy_assignments_test.go#L15) | Lists batches then <=5 stable ID reads; list failure stops all selected reads. No oracle/effective assignment/CRUD. |
| T56 | [TestAccPolicyPackagesReadOnly](../../../tests/integration/lab/policy_packages/policy_packages_test.go#L90) | Application context checks local mode rejection then Skip; delegated branch lists and <=5 name reads, optional user reads. No oracle, mutations or partial-result assertion. |
| T57 | [TestAccRecordingMessagesReadOnly](../../../tests/integration/lab/recording_messages/recording_messages_test.go#L156) | List/filter known projection, <=5 identity/conditional canonical relative reads, optional selected localization oracle. Empty list gives no selector proof. |
| T58 | [TestAccBannerTextReadOnly](../../../tests/integration/lab/teams_custom_banner_text/banner_text_test.go#L135) | List/filter exact known projection, <=5 identity/conditional relative reads, optional oracle checks count/empty list. Empty list gives no selector proof. |
| T59 | [TestAccTeamsExternalAccessConfigurationReadOnly](../../../tests/integration/lab/teams_external_access_configuration/federation_test.go#L14) | RunFederationGlobal callback: Global get + Global-filtered list complete fields, optional selected nullable scalar oracle; error stops downstream stages; no updates/grants. |
| T60 | [TestAccTenantFederationConfigurationReadOnly](../../../tests/integration/lab/tenant_federation_configuration/federation_test.go#L14) | RunFederationGlobal callback: Global get + Global-filtered list complete fields, optional selected nullable scalar oracle; error stops downstream stages; no updates/grants. |

## Boundaries for acceptance claims

Mapped live callbacks are source paths gated before SDK access. Federation uses `RunFederationGlobal` and optional narrow scalar comparison; branding read-only compares six scalars, while owned CRUD separately checks persistence and cleanup. Recording/banner selector branches are conditional on actual list members, so empty lists cannot qualify selectors; their optional oracle shape/count checks are separate assertions. Ordinary batches list/get and OCE list do not assert PowerShell parity. Packages can skip after local application-mode rejection. GPA current capability denial and original successful-list harness are separate outcomes. [Federation helper/callback contract](../../../tests/integration/lab/internal/labkit/federation.go).

Local pre-canceled-context tests do establish zero-request cancellation for selected entry points. They do not establish mid-request cancellation, between-stage interruption, continuation of independent cleanup, or OCE partial-attempt retention under cancellation. Ordered mock exchanges do establish no unexpected replay for tested errors; they do not establish recovery from every ambiguous remote success. Neither a nonnil wrapper nor HTTPAccepted means effective assignment, feature rendering, asynchronous completion or parity of all output fields.

No new tenant execution, capture, authentication, workflow edit, fixture deletion or SDK modification is part of this audit. Exact integration execution and shared-root type plans belong to the parent POL90 plan; this file supplies source-bounded current test truth and actionable offline gaps.

## Bounded pinned upstream schema inspection (2026-11-26)

This addendum is offline primary-source evidence for **authored** positives, not a successful capture. Only module import, Get-Command metadata, CLR reflection and IL resolution ran; no cmdlet ProcessRecord, connection, credential acquisition or tenant request was invoked. The pinned 8.1.0 module is retained at `/private/tmp/msteams-core90-pinned/module`. Sources and SHA256:

| Local primary source | SHA256 | Role |
|---|---|---|
| `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll` | `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49` | GPA implementing cmdlets and response models; assembly version40.2610.19888.1 |
| `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll` | `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6` | Package generated models/command callback signatures; assembly version9.1116.2.0 |

Reflection procedure: import only the pinned MicrosoftTeams.psd1; resolve each Get-Command ImplementingType (GPA) or proxy mapping; enumerate declared properties and JsonProperty/JsonConverter attributes; resolve IL operand members with declaring/method generic arguments. IL offsets below are decimal byte offsets inside the named method, not file offsets. Existing read-only `/private/tmp/ext90-inspect-generic.ps1` supplied the opcode/member decoder; no generated script or JSON evidence file was added. The hashes plus full owner/method identify reproducible primary evidence even when the temporary inspection utility is absent later.

### GPA response grammar and owners

All GPA cmdlet owners below have prefix `Microsoft.Teams.Policy.Administration.Cmdlets.Core.Custom.GPA.` in Core.dll. The job model full owner is `Microsoft.Ic3.AdminConfig.RP.Policy.Group.Models.DirectToGroupAssignmentMigrationResponseModel`.

| Entry point / exact IL owner | Result binding |
|---|---|
| `GetDirectToGroupAssignmentsMigrationStatus+<ProcessRecord>d__6::MoveNext` | IL336 JsonConvert.DeserializeObject<List<job>>, IL341 WriteObject; IL458 DeserializeObject<job>, IL463 WriteObject. List vs selected MigrationEventId controls the branch. IL473 emits the no-status diagnostic on parse failure. |
| `StartDirectToGroupAssignmentMigration+<ProcessRecord>d__22::MoveNext` | IL277 `HttpResponseMessage.IsSuccessStatusCode`; success IL389 DeserializeObject<job>, IL394 WriteObject. It is **not a message-only successful output**. Error status handling begins IL403; literals401 at412,409 at441,502 at450, diagnostic strings at460/473/486/499. This establishes broad 2xx parse branch, not a captured server success status. |
| `ClearDirectToGroupAssignmentMigration+<ProcessRecord>d__22::MoveNext` | IL338 DeserializeObject<job>; cleanup also binds a typed job, not just a message. This inspected method does not itself establish a specific required 2xx status. |
| `GetGroupAssignmentRecommendationsPerPolicyType+<ProcessRecord>d__10::MoveNext` | IL279 DeserializeObject<List<Models.GroupAssignmentRecommendationsPerPolicyType>>, IL284 WriteObject; IL294 no-recommendations diagnostic. |
| `GetGroupAssignmentRecommendationsPerPolicyName+<ProcessRecord>d__14::MoveNext` | IL279 DeserializeObject<List<Models.GroupAssignmentRecommendationsPerPolicyName>>, IL284 WriteObject; IL294 no-recommendations diagnostic. |
| `GetGroupPolicyAssignmentConflict+<ProcessRecord>d__10::MoveNext` | IL279 DeserializeObject<Models.GroupPolicyAssignmentRankConflictResponseModel>. |

The job properties have no JsonProperty renaming attributes: canonical CLR names are `MigrationEventId:Guid`, `TenantId:Guid`, `GroupId:Guid`, `GroupName:string`, `PolicyType:string`, `PolicyInstance:string`, `State:DirectToGroupAssignmentMigrationState`, `StartTime:DateTime`, `CompletionTime:DateTime`, `DirectAssignmentsProcessed:Int32`. State alone carries `JsonConverter(typeof(StringEnumConverter))`. Enum names and values are `NotStarted=0`, `PopulationInProgress=1`, `GpaCreationInProgress=2`, `WaitingForGpaProcessed=3`, `CleanUpInProgress=4`, `Completed=5`, `Failed=6`, `ConnectionLost=7`. Use those exact canonical names in source-backed authored fixtures; wire lower-camel names accepted by the SDK remain a compatibility choice, not a captured response. Plain JsonConvert DeserializeObject supplies no per-call naming strategy here.

Recommendation `Models.GroupAssignmentRecommendationsPerPolicyType` explicitly has `[JsonProperty("policyType")] PolicyType:string` and `[JsonProperty("daCount")] DirectAssignmentsCount:Int32`. The policy-name equivalent explicitly uses `policyName:string` and `daCount:Int32`. Both return **bare arrays**, not an invented value envelope.

Conflict response properties, with no renaming attributes: `TotalCountUsersWithConflict:Int32`, `UsersWithConflictDueDirectAssignmentsCount:Int32`, `UsersWithConflictDueDirectAssignmentsDetail:List<GroupPolicyAssignmentRankConflictDetailModel>`, `UsersWithConflictDueGroupAssignmentsCount:Int32`, `UsersWithConflictDueGroupAssignmentsDetail:List<...>`. Detail fields are `UserId:string`, `PolicyInstance:string`, `AssignmentSource:string`. These are bare typed objects. This source is enough to author normal typed positive values and wrong-type/overflow adversarial cases; it is not sufficient to claim required presence of every property, actual server defaults, null response semantics or effective conflict precedence. CLR nonnullable scalar defaults differ from SDK pointer/raw-field preservation. Open future strings/numeric state variants and future/raw fields are explicit SDK compatibility enhancements, not an upstream CLR promise.

### Package response grammar and statuses

Package public commands are proxy functions in `exports/ProxyCmdletDefinitionsWithHelp.ps1`. The proxy maps to generated private cmdlets; custom package create/update and group assignment first route through `custom/Merged_custom_PsExt.ps1` (functions at lines5057,5135,693). Those custom wrappers only translate policy type/name or rank inputs and call internal generated operations; their OutputType(string) annotation does not define the wire JSON. Actual generated callback signatures and model ToJson/FromJson establish the following grammar. All model names below carry prefix `Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Models.` and command owners prefix `Microsoft.Teams.ConfigAPI.Cmdlets.Generated.Cmdlets.` in private.dll.

| Shape | Primary model grammar / exact keys |
|---|---|
| Package list/get/user/recommendation | FormattedPackageSummary, FormattedPackage and FormattedPackageRecommendation expose `Description:string`, `Name:string`, `PackageType:string`, `RecommendationType:string`, `Policies:<corresponding ...Policies interface>`. Policies implementation has `Item:Object` indexer and is an open JSON object map, not an array of PolicyPair. FormattedPackage::ToJson key literals at IL67 Description,130 Name,193 PackageType,241 Policies,304 RecommendationType. |
| Custom create/update message | `PackageServiceModelsCustomPackageResponse` has `ResponseMessage:string`; ToJson IL64 key ResponseMessage. SDK MessageResult is its own name, not an upstream class. |
| Synchronous user assignment | `PackageServiceModelsPostPackageResponse` wire keys `FailedUserErrorMessages` (ToJson IL52), `FailedAssignPolicies:string[]` (186), `FailedCreatePolicies:string[]` (312), `FailedUserIds:string[]` (438), `ResponseMessage:string` (491). CLR names singular FailedAssignPolicy/FailedCreatePolicy/FailedUserId are **not wire keys**. Nested `PackageServiceModelsPostPackageResponseFailedUserErrorMessages` has `LegacyUser:string[]` (ToJson IL138). Public flattened `FailedUserErrorMessageLegacyUser` property is projection convenience; it does not replace nested FailedUserErrorMessages.LegacyUser JSON. |
| Asynchronous batch assignment | `PackageServiceModelsBatchPostPackageResponse`: `FailedCreatePolicies:string[]` (ToJson IL138), `OperationId:string` (191), `ResponseMessage:string` (254). OperationId is a string, not source-constrained to UUID; presence is not guaranteed by reflection. |

The declared command callback signatures bind these statuses/types (each method accepts HttpResponseMessage plus Task<T>, except NoContent):

| Generated owner | Success callback and body | Additional typed error callbacks |
|---|---|---|
| GetCsPolicyPackage_Get | onOk /200, FormattedPackageSummary[] |400/404/500 PackageResponseBase |
| GetCsPolicyPackage_Get1 | onOk /200, FormattedPackage |400/404/500 PackageResponseBase |
| GetCsUserPolicyPackage_Get | onOk /200, FormattedPackageSummary **single** |400/404/500 PackageResponseBase |
| GetCsUserPolicyPackageRecommendation_Get | onOk /200, FormattedPackageRecommendation[] |400/404/500 PackageResponseBase |
| NewCsCustomPolicyPackage_NewExpanded | onCreated /201, CustomPackageResponse |400/404/500 same message model |
| UpdateCsCustomPolicyPackage_UpdateExpanded | onOk /200, CustomPackageResponse |400/404/500 same message model |
| RemoveCsCustomPolicyPackage_Remove | onNoContent /204, no body task |400/404/500 CustomPackageResponse |
| GrantCsUserPolicyPackage_GrantExpanded | onCreated /201, PostPackageResponse |400/404/500/**520/521** PostPackageResponse |
| NewCsBatchPolicyPackageAssignmentOperation_NewExpanded | onAccepted /202, BatchPostPackageResponse |400/404/500 same batch model |
| GrantCsGroupPolicyPackageAssignment_GrantExpanded | onOk /200, ApplyPackageGroupResponse |400/404/500/**520/521** ApplyPackageGroupResponse |

This establishes which upstream statuses carry typed data, not that the source treats error statuses as successful assignment. SDK AssignmentResult.HTTPAccepted and normalized APIError are its own representation of that distinction. Authored tests should cover 201 partial failed-user data separately from520/521 domain failure data and202 batch acceptance. Future unknown nested error keys, RawFields, arbitrary opaque handles and missing-vs-null preservation remain SDK behavior to label separately from these upstream shapes. No package capture or live delegated success was produced by this inspection.

`PackageServiceModelsApplyPackageGroupResponse` is the group-specific upstream shape: CLR FailedAssignPolicy:string[], FailedCreatePolicy:string[], ResponseMessage:string; wire ToJson literals are FailedAssignPolicies at138, FailedCreatePolicies at264, ResponseMessage at317. `PackageServiceModelsPackageResponseBase` also has only ResponseMessage:string. User onCreated (`GrantCsUserPolicyPackage_GrantExpanded+<onCreated>d__62::MoveNext`), user on520 (`...+<on520>d__59`) and group onOk (`GrantCsGroupPolicyPackageAssignment_GrantExpanded+<onOk>d__67`) all call WriteObject at IL339. Therefore a typed PowerShell output can occur at HTTP520; it must not be mistaken for HTTP success or per-user assignment success.

Additional exact source hashes: `custom/Merged_custom_PsExt.ps1` SHA256 `03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255`; `exports/ProxyCmdletDefinitionsWithHelp.ps1` SHA256 `5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5`.

GPA status handling qualification: `CommonHttpHelperRequest+<SendGPARequests>d__5::MoveNext` in PolicyRp.dll delegates at IL409 to HttpClientHelper.SendAsync; it contains no IsSuccessStatusCode/EnsureSuccessStatusCode call. `HttpClientHelper+<SendAsync>d__8::MoveNext` delegates at IL27 to SendRequestAsync and returns the response; `HttpClientHelper+<SendRequestAsync>d__6::MoveNext` calls System.Net.Http.HttpClient.SendAsync at65 and gets its response at155, with no success-status guard. PolicyRp.dll SHA256 is `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`. Thus Start explicitly gates job decoding on broad HTTP success; the inspected Get/list/recommendation/conflict/Clear callers decode content without an equivalent per-call success check. This is source behavior, not proof that an error-envelope body is a successful job. The SDK's normalized non-2xx APIError behavior is intentionally distinguishable from upstream parse/string diagnostic behavior. A specific successful server status for GPA reads/Clear remains uncaptured; use200 as a clearly authored transport scenario, not observed upstream service status.
