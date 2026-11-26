# Integrated core qualification evidence

Owner: [CORE-90 / #193](https://github.com/deploymenttheory/go-sdk-msteams/issues/193).
Baseline: MicrosoftTeams 8.1.0. Audited source base: `5d55f4a`.

**Status: incomplete.** This report inventories all 40 assigned core commands and
exact parameter sets. The command ledger remains in-progress / metadata-only;
new offline scenarios do not constitute live qualification or a release sign-off.

## Integrated protocol scenarios

The [integration suite](../../../tests/contract/core/integration_test.go) calls the
real root client, compositions and resource services through the shared offline
contract harness. The harness verifies every ordered request, body, audience and
response; extra requests (including unsafe cleanup) fail. No test changes the
existing per-resource request-boundary wrappers.

- `TestUnitIntegratedCoreOwnerAndNativeMembershipIdentity`: completes group member
  and owner writes, then reads and deletes a distinct opaque native membership.
  Directory identity and opaque membership identity must survive separately; native
  deletion must not turn into deletion of a directory object or group reference.
- `TestUnitIntegratedCorePartialOwnerFailureAndExplicitRecovery`: preserves a
  successful member write after an owner grant fails. Only the caller explicitly
  removes the owned relationship through `/$ref`, then independently reads absence.
  A denied recovery remains a failure and must not perform the confirmation read.
- `TestUnitIntegratedCoreCreationOwnershipControlsCompensation`: a confirmed group
  create followed by a denied prerequisite read compensates that owned group;
  the same denial for a supplied group retains it. Resource ID, primary error,
  create receipt and separate cleanup result remain available.

The shared-channel scenarios additionally exercise accepted creation through
both Location and documented Content-Location, terminal polling, a typed operation
read and the resulting channel read through the same root. They reject foreign
Content-Location and refuse to replace a malformed primary Location with a fallback.
Cancelling a local wait retains the accepted remote operation identity and sends no
poll or repeated creation request. The header-handling cases passed a guarded local test run. The initial-share
models/examples compile; integration execution awaits CI after correcting a
synthetic-clock test nesting error. No local Go build starts below the disk reserve.

The [beta channel creation contract](https://learn.microsoft.com/en-us/graph/api/channel-post?view=graph-rest-beta)
returns Content-Location in its shared-channel 202 example. Source inspection found
that Channels.Create read only Location; it now uses Content-Location only when
Location is absent, passing both through the existing origin/team validation.
This corrects a missing accepted-operation path, not a new endpoint or live claim.
The same beta Example8 supplies the typed initial `sharedWithTeams` write input:
`InitialSharedTeam` references are available on CreateRequest, independently from
read-only relationship snapshots. Tests distinguish omission, explicit empty and
host-team references, reject invalid team IDs locally, and preserve explicit-null
or nonshared inputs for server validation without a success claim.
The native route remains a channel create; no existing-relationship mutation is
invented. Polling tests use a synthetic clock to preserve the 31-second interval
without adding real wait time.

## Fixture provenance and permission boundaries

These are **authored offline protocol vectors**, not captured tenant replies.
Synthetic UUIDs and opaque membership IDs exercise documented schemas; empty 204
responses and 200 collection/member shapes follow the primary sources below.
They prove SDK composition against those selected contracts. They do not prove
propagation, tenant permission grants, pinned PowerShell parity or successful lab
cleanup. The existing [pinned composition contract](membership-resource-contract.md)
and SDK implementation define the group-first PowerShell mapping and provisioning
presets, separately from each primitive's documented wire contract.

Primary Microsoft Graph v1.0 sources reviewed on 2026-11-26:

| Primitive | Source and evidence |
|---|---|
| Group member list/add | [List members](https://learn.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0), [add member](https://learn.microsoft.com/en-us/graph/api/group-post-members?view=graph-rest-1.0): directory collections and reference writes |
| Group owners list/add | [List owners](https://learn.microsoft.com/en-us/graph/api/group-list-owners?view=graph-rest-1.0), [add owner](https://learn.microsoft.com/en-us/graph/api/group-post-owners?view=graph-rest-1.0): collection reads and empty 204 receipt; membership and ownership are separate |
| Relationship recovery | [Remove member](https://learn.microsoft.com/en-us/graph/api/group-delete-members?view=graph-rest-1.0): DELETE must retain `/$ref`; removing it can delete the directory object |
| Native membership | [Get team member](https://learn.microsoft.com/en-us/graph/api/team-get-members?view=graph-rest-1.0), [remove team member](https://learn.microsoft.com/en-us/graph/api/team-delete-members?view=graph-rest-1.0): native membership read/delete preserves opaque membership IDs separately from directory userId |
| Group ownership and compensation | [Create group](https://learn.microsoft.com/en-us/graph/api/group-post-groups?view=graph-rest-1.0), [delete group](https://learn.microsoft.com/en-us/graph/api/group-delete?view=graph-rest-1.0): 201 group identity and empty 204 delete receipt |

Group member changes require appropriate GroupMember.ReadWrite.All or higher
permission; owner changes require Group.ReadWrite.All. Delegated roles and consent
still apply. Native team membership has separate TeamMember permissions, including
resource-specific consent variants. These tests bind delegated synthetic credentials
and assert the Graph audience; they do not grant any role or establish app-only
compatibility. Consult the linked permission tables before provisioning a live lab.
Owner synchronization with Teams can be delayed: a synthetic subsequent native read
is an independently supplied response, not an assertion of immediate propagation.

## Remaining qualification requirements

| Family | Missing evidence for complete CORE-90 qualification |
|---|---|
| Teams | Owned live backing-group/team lifecycle, paired reads and management compositions; interrupted-run cleanup verification; all applicable parameter semantics |
| Channels | Authorized owner fixtures, standard/private/shared lifecycles, list membership, pinned comparisons and verified deletion |
| Team and channel members | Live CRUD, role convergence, exact identities, pinned comparison and complete-list absence/recovery |
| Operations | Real accepted operation through running and terminal transitions; exact permission contract and pinned comparison |
| Shared channels | Documented resource methods are read/list/unshare/allowed-members; generated POST/PATCH is not an accepted mutation contract. Live unshare and shared host/recipient fixtures remain unqualified |
| Deployment/migration | Published CSV column grammar is available; accepted orchestration outcomes, exact privileges and licensed owned-team/recipient fixtures remain unqualified |
| Targeting | Policy create/upsert and hierarchy full-resource/list contracts, conditional modern/TRPS behavior, import/delete persistence and recovery |
| Templates | Existing expanded/body CRUD/list, pinned comparisons and independent cleanup passed in CI; remaining field/parameter-set qualification and broader composition evidence still required |

Current template evidence: [tenant run 45415515397](https://github.com/deploymenttheory/go-sdk-msteams/actions/runs/45415515397)
at revision `fb04fe2b841c55ff41fcd728d37ec608031629c1` reports the
`TestAcc team_template` job successful. Its partition requires both expanded and
JSON-body children, four pinned comparisons and two cleaned journal entries.
Sixteen permanent sanitized responses are consumed by its resource tests. The
current oracle uses the bounded caller context and a one-second WaitDelay; its
cleanup verifies both selected GET404 and complete-list absence before returning.
Older issue notes describing these checks as missing are historical, not current
blockers. This evidence does not cover every template field or parameter set.

The [pending acceptance catalog](../../../tools/acceptance/pending.json) records
exact unavailable Graph fixtures and permission prerequisites. Its reasons are
limitations, not passes. Existing [acceptance notes](../../../tests/integration/lab)
and family notes in this directory describe scoped evidence. Child issue closure
and symbol existence are not substitutes for these missing outcomes. CORE-90
must remain open until all required integrated and available live evidence passes.

## Rechecked contract questions

These findings were checked against primary sources on 2026-11-26, rather than
carried forward from the historical issue text:

- **Sharing relationships:** the [beta resource method table](https://learn.microsoft.com/en-us/graph/api/resources/sharedwithchannelteaminfo?view=graph-rest-beta)
  lists reads, listing, unsharing and allowed members. It supplies no accepted
  relationship POST/PATCH request or success contract. The SDK implements those
  documented operations; generated navigation builders alone cannot establish
  additional mutations. [Channel creation](https://learn.microsoft.com/en-us/graph/api/channel-post?view=graph-rest-1.0)
  separately documents `sharedWithTeams` on a shared-channel create request; this
  is not evidence for mutating an existing sharing relationship. An owned shared
  channel, explicit recipient and suitable Graph consent remain live prerequisites.
- **Operation authorization:** the [current operation-get reference](https://learn.microsoft.com/en-us/graph/api/teamsasyncoperation-get?view=graph-rest-beta)
  explicitly scopes its permission table and route to chats. The [resource page](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-1.0)
  establishes asynchronous status and polling semantics, but does not establish
  team-route grants. These sources confirm the discrepancy; assigning chat grants
  to the pinned team route would be an unsupported inference. A real accepted
  team operation and authorization evidence are still required for live completion.
- **Targeting policy:** [Set-CsTeamsTargetingPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamstargetingpolicy?view=teams-ps)
  describes modification of existing tag settings. It does not establish an upsert
  or creation response. Current SDK GET/list/PUT/delete and omitted-identity Global
  handling follow the pinned mapping. No documented new standalone create contract
  was found. [Hierarchy upload](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-teamtargetinghierarchy?view=teams-ps)
  replaces the single tenant hierarchy and documents versions 1 and 2; it does not
  document a full hierarchy export/list needed to restore arbitrary existing data.
  The SDK retains opaque v2 preparation responses without inventing their schema.
  A disposable singleton or verified export/restore contract remains necessary for
  import/delete acceptance. Valid tokens alone do not establish its privileges.
- **Deployment grammar:** the [published deployment guide](https://learn.microsoft.com/en-us/microsoft-365/frontline/deploy-teams-at-scale)
  specifies ordered columns. Teams.csv uses `Team Name`, `Existing Team ID`,
  `Visibility`, `Team Template ID`. Users.csv uses `User Full Name`, `User UPN or ID`,
  `Team Name`, `ActionType`, `Owner or Member`. This resolves the caller's CSV input
  layout; `BuildArchive` deliberately preserves the bytes, as the pinned cmdlet
  does, instead of introducing a stricter parser. The [submission reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csbatchteamsdeployment?view=teams-ps)
  describes orchestration identity and notifications, not a per-team typed status
  response contract. The guide requires licensed owners. No owned licensed owner,
  deployment recipients or known accepted orchestration selector is configured in
  the CI inputs; submission would therefore lack both prerequisites and verified
  resource recovery. No notification email or bulk deployment is initiated here.

The configured CI secret names provide application authentication and dedicated
user/group fixtures. The workflow does not load a team-owner fixture, a delegated
region token or a known batch/hierarchy operation selector. A read of secret names
cannot establish the app's effective Graph consent; current notes record denied
Graph reads, and no new successful Graph live access is claimed by this PR.

## Available targeting read conformance

The former pending reason for `TestAccTargetingPolicyReadOnly` did not identify a
missing selector: its Global read/list needs only the existing application lab
identity. This PR moves that exact case into a serialized `targeting_policy` CI
partition. It requires one unambiguous Global row, matching typed fields between
selected read and filtered list, and one pinned PowerShell comparison. No policy,
assignment, directory permission or hierarchy is changed. The outcome is currently
unproven; a denial fails the positive read scenario and is not counted as success.
Mutation effects, additional identities, legacy flighting, cloud and delegated
qualification remain distinct. This revalidates the old pending disposition rather
than treating a historical note as a permanent fixture blocker.

The bound-tenant `TestAccPrivateChannelMigrationReadOnly` likewise needs no
additional operation selector or licensed owner: it reads only the configured
tenant's singleton status. It now has a read-only `private_channel_migration`
partition requiring matching tenant identity and one pinned status/details
comparison. No migration, deployment or tenant setting is changed. Denial,
missing tenant identity or mismatch fails the positive scenario; no live success
is claimed before the new CI result. Known orchestration, hierarchy and region
selectors remain genuine separate fixture prerequisites.

## Additional available template conformance

This PR extends the existing registered template lifecycle with an explicit-empty
Description update after the non-default description round trip. Expanded and JSON
Body paths each require the third pinned comparison, preserved updated summary and
General channel, and the existing independent deletion/absence checks. The CI guard
rejects the previous two-comparison-per-variant report. New response fixtures are
written by the live acceptance test; no tenant reply is manufactured for this case.
This scenario is prepared, not yet a live pass for this PR revision. Prior successful
template CI evidence above retains its original scope.

## Command and parameter-set inventory

This snapshot is taken directly from `contracts/coverage/8.1.0.json`. Each row
names every parameter set and its declared tests. Declarations identify where to
inspect evidence; they are not assertions that each success path ran. All 40
commands currently retain `implementation: in-progress`, `contract: metadata-only`
and `live: not-run`. No ledger success state is promoted by this change.

### Add-TeamChannelUser (CORE-20)

- **__AllParameterSets** — SDK: [Service.Add](../../../msteams/graph_api/channel_user_management/service.go), [Service.ResolveByDisplayName](../../../msteams/graph_api/channel/names.go), [Service.ListAll](../../../msteams/graph_api/channel_member/service.go), [Service.Create](../../../msteams/graph_api/channel_member/service.go), [Service.PromoteOwner](../../../msteams/graph_api/channel_member/service.go), [Service.ResolveByUser](../../../msteams/graph_api/channel_member/resolution.go), [Resolver.Resolve](../../../msteams/graph_api/directory_user/resolver.go).
  Contract tests: [TestUnitAddRemoveAndOwnerCompositions](../../../msteams/graph_api/channel_user_management/service_test.go), [TestUnitFailuresStopBeforeFurtherChanges](../../../msteams/graph_api/channel_user_management/service_test.go).
  Live tests: [TestAccChannelMembershipCRUD](../../../tests/integration/lab/graph/channel_membership_test.go).

### Add-TeamUser (CORE-26)

- **__AllParameterSets** — SDK: [Service.Add](../../../msteams/graph_api/team_user_management/service.go), [Resolver.Resolve](../../../msteams/graph_api/directory_user/resolver.go), [MemberService.AddReference](../../../msteams/graph_api/group_relationship/members.go), [MemberService.ListAll](../../../msteams/graph_api/group_relationship/members.go), [OwnerService.AddReference](../../../msteams/graph_api/group_relationship/owners.go), [OwnerService.ListAll](../../../msteams/graph_api/group_relationship/owners.go).
  Contract tests: [TestUnitAddCompositionStagesAndCorrectedPreflight](../../../msteams/graph_api/team_user_management/service_test.go), [TestUnitAddCompletePreflightAndResolvedIdentity](../../../msteams/graph_api/team_user_management/service_test.go), [TestUnitCompositionInvalidInputAndCancellationAreLocal](../../../msteams/graph_api/team_user_management/service_test.go), [TestUnitCompositionMappingsExistInPinnedRuntime](../../../msteams/graph_api/team_user_management/list_test.go), [TestUnitIntegratedCoreOwnerAndNativeMembershipIdentity](../../../tests/contract/core/integration_test.go), [TestUnitIntegratedCorePartialOwnerFailureAndExplicitRecovery](../../../tests/contract/core/integration_test.go).
  Live tests: [TestAccTeamMembershipCRUD](../../../tests/integration/lab/graph/team_membership_test.go).

### Get-AssociatedTeam (CORE-24)

- **__AllParameterSets** — SDK: [Service.ListAssociatedTeams](../../../msteams/graph_api/shared_channel_management/service.go), [Service.ListMine](../../../msteams/graph_api/associated_team/service.go), [Service.ListForUser](../../../msteams/graph_api/associated_team/service.go).
  Contract tests: [TestUnitAssociatedTeamsSelectionAndResolution](../../../msteams/graph_api/shared_channel_management/service_test.go), [TestUnitCompositionsRejectOptionsAndCancellationBeforeCredentials](../../../msteams/graph_api/shared_channel_management/service_test.go).
  Live tests: [TestAccSharedChannelReadOnly](../../../tests/integration/lab/graph/shared_channel_test.go).

### Get-CsBatchTeamsDeploymentStatus (CORE-22)

- **Get** — SDK: [Service.GetByID](../../../msteams/teams_admin_api/batch_deployment/service.go).
  Contract tests: [TestUnitBatchStatusIsOpaqueAndIdentityInputIsNotAURI](../../../msteams/teams_admin_api/batch_deployment/service_test.go), [TestUnitBatchErrorsRemainErrorsAndSubmissionIsNotReplayed](../../../msteams/teams_admin_api/batch_deployment/service_test.go).
  Live tests: [TestAccBatchDeploymentStatusReadOnly](../../../tests/integration/lab/batch_deployment/deployment_migration_test.go).
- **GetViaIdentity** — SDK: [Service.GetByIdentity](../../../msteams/teams_admin_api/batch_deployment/service.go).
  Contract tests: [TestUnitBatchStatusIsOpaqueAndIdentityInputIsNotAURI](../../../msteams/teams_admin_api/batch_deployment/service_test.go), [TestUnitBatchErrorsRemainErrorsAndSubmissionIsNotReplayed](../../../msteams/teams_admin_api/batch_deployment/service_test.go).
  Live tests: [TestAccBatchDeploymentStatusReadOnly](../../../tests/integration/lab/batch_deployment/deployment_migration_test.go).

### Get-CsTeamsTargetingPolicy (CORE-25)

- **Filter** — SDK: [Service.ListFiltered](../../../msteams/teams_admin_api/targeting_policy/filter.go).
  Contract tests: [TestUnitTargetingRequestVectors](../../../msteams/teams_admin_api/targeting_policy/service_test.go), [TestUnitTargetingFilterRequestAndLocalValidation](../../../msteams/teams_admin_api/targeting_policy/service_test.go).
  Live tests: [TestAccTargetingPolicyReadOnly](../../../tests/integration/lab/targeting_policy/targeting_test.go).
- **Identity** — SDK: [Service.GetByIdentity](../../../msteams/teams_admin_api/targeting_policy/service.go), [Service.List](../../../msteams/teams_admin_api/targeting_policy/service.go).
  Contract tests: [TestUnitTargetingRequestVectors](../../../msteams/teams_admin_api/targeting_policy/service_test.go), [TestUnitTargetingFilterRequestAndLocalValidation](../../../msteams/teams_admin_api/targeting_policy/service_test.go).
  Live tests: [TestAccTargetingPolicyReadOnly](../../../tests/integration/lab/targeting_policy/targeting_test.go).

### Get-CsTeamTemplate (CORE-27)

- **Get** — SDK: [Service.GetByID](../../../msteams/teams_admin_api/team_template/service.go).
  Contract tests: [TestUnitServiceGetByID](../../../msteams/teams_admin_api/team_template/service_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).
- **GetViaIdentity** — SDK: [Service.GetByID](../../../msteams/teams_admin_api/team_template/service.go).
  Contract tests: [TestUnitServiceGetByID](../../../msteams/teams_admin_api/team_template/service_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).

### Get-CsTeamTemplateList (CORE-27)

- **DefaultLocaleOverride** — SDK: [Service.List](../../../msteams/teams_admin_api/team_template/service.go).
  Contract tests: [TestUnitServiceList](../../../msteams/teams_admin_api/team_template/service_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).

### Get-MultiGeoRegion (CORE-22)

- **Identity** — SDK: [Service.GetForUser](../../../msteams/teams_admin_api/region/service.go), [Service.GetForGroup](../../../msteams/teams_admin_api/region/service.go).
  Contract tests: [TestUnitRegionUsesExactDelegatedScopeAndDirectEntityRoutes](../../../msteams/teams_admin_api/region/service_test.go), [TestUnitApplicationRegionAndUnqualifiedCloudsRejectBeforeTokens](../../../msteams/teams_admin_api/region/service_test.go).
  Live tests: [TestAccMultiGeoRegionReadOnly](../../../tests/integration/lab/region/deployment_migration_test.go).

### Get-Operation (CORE-23)

- **__AllParameterSets** — SDK: [Service.GetByID](../../../msteams/graph_api/team_operation/service.go), [Service.ResumeByID](../../../msteams/graph_api/team_operation/service.go), [Service.WaitByID](../../../msteams/graph_api/team_operation/service.go).
  Contract tests: [TestUnitIntegratedCoreSharedChannelOperationAndRead](../../../tests/contract/core/integration_test.go), [TestUnitGetOperationPreservesNativeStatesAndFutureData](../../../msteams/graph_api/team_operation/service_test.go), [TestUnitGetRejectsMalformedAndUnexpectedResponses](../../../msteams/graph_api/team_operation/service_test.go), [TestUnitWaitDenialRetainsResumableHandle](../../../msteams/graph_api/team_operation/service_test.go).
  Live tests: [TestAccTeamOperationReadOnly](../../../tests/integration/lab/graph/team_operation_test.go).

### Get-SharedWithTeam (CORE-24)

- **__AllParameterSets** — SDK: [Service.ListSharedWithTeams](../../../msteams/graph_api/shared_channel_management/service.go), [Service.GetByID](../../../msteams/graph_api/shared_channel/service.go), [Service.ListAll](../../../msteams/graph_api/shared_channel/service.go).
  Contract tests: [TestUnitSharedTeamOptionalSelectorAndUnshare](../../../msteams/graph_api/shared_channel_management/service_test.go).
  Live tests: [TestAccSharedChannelReadOnly](../../../tests/integration/lab/graph/shared_channel_test.go).

### Get-SharedWithTeamUser (CORE-24)

- **__AllParameterSets** — SDK: [Service.ListAllowedUsers](../../../msteams/graph_api/shared_channel_management/service.go), [Service.ListAllowedMembers](../../../msteams/graph_api/shared_channel/service.go), [Resolver.LookupPrincipalNames](../../../msteams/graph_api/directory_user/principal_names.go).
  Contract tests: [TestUnitAllowedUsersProjectionFilterAndNullableEnrichment](../../../msteams/graph_api/shared_channel_management/service_test.go), [TestUnitAllowedUserFailuresRetainNativeState](../../../msteams/graph_api/shared_channel_management/service_test.go), [TestUnitCompositionsRejectOptionsAndCancellationBeforeCredentials](../../../msteams/graph_api/shared_channel_management/service_test.go).
  Live tests: [TestAccSharedChannelReadOnly](../../../tests/integration/lab/graph/shared_channel_test.go).

### Get-Team (CORE-28)

- **Filters** — SDK: [Service.List](../../../msteams/graph_api/team_management/list.go), [Service.NewPager](../../../msteams/graph_api/group/pagination.go), [Service.ListJoined](../../../msteams/graph_api/team/joined.go), [Resolver.Resolve](../../../msteams/graph_api/directory_user/resolver.go), [Service.GetByID](../../../msteams/graph_api/team_management/service.go), [Service.GetByID](../../../msteams/graph_api/group/service.go), [Service.GetByID](../../../msteams/graph_api/team/service.go).
  Contract tests: [TestUnitManagementListDiscoveryDenial](../../../msteams/graph_api/team_management/service_test.go), [TestUnitListingLocalValidationAndCurrentUserMode](../../../msteams/graph_api/team_management/validation_test.go).
  Live tests: [TestAccTeamLifecycle](../../../tests/integration/lab/graph/team_lifecycle_test.go).
- **Identity** — SDK: [Service.GetByID](../../../msteams/graph_api/team_management/service.go), [Service.GetByID](../../../msteams/graph_api/group/service.go), [Service.GetByID](../../../msteams/graph_api/team/service.go).
  Contract tests: [TestUnitManagementFirstRequestDenials](../../../msteams/graph_api/team_management/service_test.go), [TestUnitInvalidInputAndCancellationPerformNoIO](../../../msteams/graph_api/team_management/validation_test.go).
  Live tests: [TestAccTeamLifecycle](../../../tests/integration/lab/graph/team_lifecycle_test.go).

### Get-TeamAllChannel (CORE-21)

- **__AllParameterSets** — SDK: [Service.ListAll](../../../msteams/graph_api/channel/service.go).
  Contract tests: [TestUnitCollectionsAndPaging](../../../msteams/graph_api/channel/service_test.go), [TestUnitExplicitCollectionWrappers](../../../msteams/graph_api/channel/names_test.go).
  Live tests: [TestAccChannelCRUD](../../../tests/integration/lab/graph/channel_test.go).

### Get-TeamChannel (CORE-21)

- **__AllParameterSets** — SDK: [Service.ListComplete](../../../msteams/graph_api/channel/service.go), [Service.List](../../../msteams/graph_api/channel/service.go).
  Contract tests: [TestUnitIntegratedCoreSharedChannelOperationAndRead](../../../tests/contract/core/integration_test.go), [TestUnitCollectionsAndPaging](../../../msteams/graph_api/channel/service_test.go).
  Live tests: [TestAccChannelCRUD](../../../tests/integration/lab/graph/channel_test.go).

### Get-TeamChannelUser (CORE-20)

- **__AllParameterSets** — SDK: [Service.List](../../../msteams/graph_api/channel_user_management/list.go), [Service.ResolveByDisplayName](../../../msteams/graph_api/channel/names.go), [Service.ListAll](../../../msteams/graph_api/channel_member/service.go).
  Contract tests: [TestUnitListRoleOrderPresenceAndPagedEnrichment](../../../msteams/graph_api/channel_user_management/list_test.go), [TestUnitEnrichmentFailureNeverProducesACompleteList](../../../msteams/graph_api/channel_user_management/list_test.go).
  Live tests: [TestAccChannelMembershipCRUD](../../../tests/integration/lab/graph/channel_membership_test.go).

### Get-TeamIncomingChannel (CORE-21)

- **__AllParameterSets** — SDK: [Service.ListIncoming](../../../msteams/graph_api/channel/service.go).
  Contract tests: [TestUnitCollectionsAndPaging](../../../msteams/graph_api/channel/service_test.go), [TestUnitExplicitCollectionWrappers](../../../msteams/graph_api/channel/names_test.go).
  Live tests: [TestAccChannelCRUD](../../../tests/integration/lab/graph/channel_test.go).

### Get-TeamTargetingHierarchyStatus (CORE-25)

- **__AllParameterSets** — SDK: [Service.GetStatus](../../../msteams/teams_admin_api/targeting_hierarchy/service.go), [V1Service.GetOperation](../../../msteams/teams_admin_api/targeting_hierarchy/native.go), [V2Service.GetOperation](../../../msteams/teams_admin_api/targeting_hierarchy/native.go).
  Contract tests: [TestUnitNativeStatusRoutesAndIndependentHierarchyPrimitives](../../../msteams/teams_admin_api/targeting_hierarchy/service_test.go), [TestUnitStatusProjectionFromLocalOperationValues](../../../msteams/teams_admin_api/targeting_hierarchy/projection_test.go), [TestUnitStatusErrorProjectionFromLocalOperationValues](../../../msteams/teams_admin_api/targeting_hierarchy/projection_test.go).
  Live tests: [TestAccTargetingHierarchyStatusReadOnly](../../../tests/integration/lab/targeting_hierarchy/targeting_test.go).

### Get-TeamUser (CORE-26)

- **__AllParameterSets** — SDK: [Service.List](../../../msteams/graph_api/team_user_management/list.go), [MemberService.ListAll](../../../msteams/graph_api/group_relationship/members.go), [OwnerService.ListAll](../../../msteams/graph_api/group_relationship/owners.go).
  Contract tests: [TestUnitListCompositionOwnerPrecedenceProjectionAndFilters](../../../msteams/graph_api/team_user_management/list_test.go), [TestUnitListCompositionPreservesOwnerPrefixOnMemberReadFailure](../../../msteams/graph_api/team_user_management/list_test.go), [TestUnitCompositionMappingsExistInPinnedRuntime](../../../msteams/graph_api/team_user_management/list_test.go).
  Live tests: [TestAccTeamMembershipCRUD](../../../tests/integration/lab/graph/team_membership_test.go).

### Get-TenantPrivateChannelMigrationStatus (CORE-22)

- **__AllParameterSets** — SDK: [Service.Get](../../../msteams/teams_admin_api/private_channel_migration/service.go).
  Contract tests: [TestUnitMigrationCloudRequestPrefixesAndAudiences](../../../msteams/teams_admin_api/private_channel_migration/service_test.go), [TestUnitMigrationProfilesDoNotFallbackAndWrongBackendIsLocal](../../../msteams/teams_admin_api/private_channel_migration/service_test.go).
  Live tests: [TestAccPrivateChannelMigrationReadOnly](../../../tests/integration/lab/private_channel_migration/deployment_migration_test.go).

### New-CsBatchTeamsDeployment (CORE-22)

- **__AllParameterSets** — SDK: [BuildArchive](../../../msteams/teams_admin_api/batch_deployment/archive.go), [Service.Create](../../../msteams/teams_admin_api/batch_deployment/service.go).
  Contract tests: [TestUnitArchivePreservesOriginalBytesFixedNamesAndReaderOwnership](../../../msteams/teams_admin_api/batch_deployment/archive_test.go), [TestUnitBatchSubmissionRequestAndDenial](../../../msteams/teams_admin_api/batch_deployment/service_test.go), [TestUnitBatchErrorsRemainErrorsAndSubmissionIsNotReplayed](../../../msteams/teams_admin_api/batch_deployment/service_test.go).
  Live tests: none declared.

### New-CsTeamsHiddenMeetingTemplate (CORE-27)

- **Identity** — SDK: [NewHiddenMeetingTemplate](../../../msteams/teams_admin_api/team_template/local.go).
  Contract tests: [TestUnitHiddenBuildersHaveNoManagementOrAuthIO](../../../msteams/teams_admin_api/team_template/local_test.go).
  Live tests: none declared.

### New-CsTeamsHiddenTemplate (CORE-27)

- **Identity** — SDK: [NewHiddenTemplate](../../../msteams/teams_admin_api/team_template/local.go).
  Contract tests: [TestUnitHiddenBuildersHaveNoManagementOrAuthIO](../../../msteams/teams_admin_api/team_template/local_test.go).
  Live tests: none declared.

### New-CsTeamTemplate (CORE-27)

- **New** — SDK: [Service.Create](../../../msteams/teams_admin_api/team_template/service.go), [Definition](../../../msteams/teams_admin_api/team_template/definition.go), [NewDefinition](../../../msteams/teams_admin_api/team_template/definition.go), [ParseDefinition](../../../msteams/teams_admin_api/team_template/definition.go).
  Contract tests: [TestUnitServiceCreate](../../../msteams/teams_admin_api/team_template/service_test.go), [TestUnitDefinitionSerializationVectors](../../../msteams/teams_admin_api/team_template/definition_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).
- **NewExpanded** — SDK: [Service.Create](../../../msteams/teams_admin_api/team_template/service.go), [Definition](../../../msteams/teams_admin_api/team_template/definition.go), [NewDefinition](../../../msteams/teams_admin_api/team_template/definition.go), [ParseDefinition](../../../msteams/teams_admin_api/team_template/definition.go).
  Contract tests: [TestUnitServiceCreate](../../../msteams/teams_admin_api/team_template/service_test.go), [TestUnitDefinitionSerializationVectors](../../../msteams/teams_admin_api/team_template/definition_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).
- **NewViaIdentity** — SDK: [Service.Create](../../../msteams/teams_admin_api/team_template/service.go), [Definition](../../../msteams/teams_admin_api/team_template/definition.go), [NewDefinition](../../../msteams/teams_admin_api/team_template/definition.go), [ParseDefinition](../../../msteams/teams_admin_api/team_template/definition.go).
  Contract tests: [TestUnitServiceCreate](../../../msteams/teams_admin_api/team_template/service_test.go), [TestUnitDefinitionSerializationVectors](../../../msteams/teams_admin_api/team_template/definition_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).
- **NewViaIdentityExpanded** — SDK: [Service.Create](../../../msteams/teams_admin_api/team_template/service.go), [Definition](../../../msteams/teams_admin_api/team_template/definition.go), [NewDefinition](../../../msteams/teams_admin_api/team_template/definition.go), [ParseDefinition](../../../msteams/teams_admin_api/team_template/definition.go).
  Contract tests: [TestUnitServiceCreate](../../../msteams/teams_admin_api/team_template/service_test.go), [TestUnitDefinitionSerializationVectors](../../../msteams/teams_admin_api/team_template/definition_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).

### New-Team (CORE-28)

- **CreateTeam** — SDK: [Service.Create](../../../msteams/graph_api/team_management/creation.go), [Resolver.Resolve](../../../msteams/graph_api/directory_user/resolver.go), [Service.Create](../../../msteams/graph_api/group/service.go), [Service.GetByID](../../../msteams/graph_api/group/service.go), [Service.DeleteByID](../../../msteams/graph_api/group/service.go), [Service.CreateFromGroup](../../../msteams/graph_api/team/service.go), [Service.GetByID](../../../msteams/graph_api/team_management/service.go), [Service.GetByID](../../../msteams/graph_api/group/service.go), [Service.GetByID](../../../msteams/graph_api/team/service.go).
  Contract tests: [TestUnitManagementCreateRequestVariantsDenied](../../../msteams/graph_api/team_management/service_test.go), [TestUnitSettingsInputsCoverBothPinnedApplicationParameterSets](../../../msteams/graph_api/team_management/creation_settings_test.go), [TestUnitSettingsSerializeEveryInputField](../../../msteams/graph_api/team_management/creation_settings_test.go), [TestUnitIntegratedCoreCreationOwnershipControlsCompensation](../../../tests/contract/core/integration_test.go).
  Live tests: [TestAccTeamLifecycle](../../../tests/integration/lab/graph/team_lifecycle_test.go).
- **MigrateGroup** — SDK: [Service.MigrateGroup](../../../msteams/graph_api/team_management/creation.go), [Service.CreateFromGroup](../../../msteams/graph_api/team/service.go), [Service.GetByID](../../../msteams/graph_api/team_management/service.go), [Service.GetByID](../../../msteams/graph_api/group/service.go), [Service.GetByID](../../../msteams/graph_api/team/service.go).
  Contract tests: [TestUnitManagementCreateRequestVariantsDenied](../../../msteams/graph_api/team_management/service_test.go), [TestUnitCreationValidationPrecedesCredentialsAndWrites](../../../msteams/graph_api/team_management/validation_test.go), [TestUnitIntegratedCoreCreationOwnershipControlsCompensation](../../../tests/contract/core/integration_test.go).
  Live tests: [TestAccTeamLifecycle](../../../tests/integration/lab/graph/team_lifecycle_test.go).

### New-TeamChannel (CORE-21)

- **__AllParameterSets** — SDK: [InitialSharedTeam](../../../msteams/graph_api/channel/models.go), [Service.Create](../../../msteams/graph_api/channel/creation.go), [Service.CreateWithOwner](../../../msteams/graph_api/channel/creation.go).
  Contract tests: [TestUnitIntegratedCoreInitialShareShapesRemainServerValidated](../../../tests/contract/core/integration_test.go), [TestUnitIntegratedCoreInitialSharesValidateBeforeHTTP](../../../tests/contract/core/integration_test.go), [TestUnitIntegratedCoreSharedChannelCancellationRetainsOperation](../../../tests/contract/core/integration_test.go), [TestUnitIntegratedCoreSharedChannelContentLocationTrustBoundary](../../../tests/contract/core/integration_test.go), [TestUnitIntegratedCoreSharedChannelOperationAndRead](../../../tests/contract/core/integration_test.go), [TestUnitNativeCRUDContracts](../../../msteams/graph_api/channel/service_test.go), [TestUnitSharedCreationAndOperation](../../../msteams/graph_api/channel/service_test.go), [TestUnitOwnerResolutionComposition](../../../msteams/graph_api/channel/service_test.go).
  Live tests: [TestAccChannelCRUD](../../../tests/integration/lab/graph/channel_test.go).

### Remove-CsTeamsTargetingPolicy (CORE-25)

- **Identity** — SDK: [Service.DeleteByIdentity](../../../msteams/teams_admin_api/targeting_policy/service.go).
  Contract tests: [TestUnitTargetingRequestVectors](../../../msteams/teams_admin_api/targeting_policy/service_test.go), [TestUnitTargetingFailuresDoNotInventSuccessOrReplayWrites](../../../msteams/teams_admin_api/targeting_policy/service_test.go).
  Live tests: none declared.

### Remove-CsTeamTemplate (CORE-27)

- **Delete** — SDK: [Service.DeleteByID](../../../msteams/teams_admin_api/team_template/service.go).
  Contract tests: [TestUnitServiceDeleteByID](../../../msteams/teams_admin_api/team_template/service_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).
- **DeleteViaIdentity** — SDK: [Service.DeleteByID](../../../msteams/teams_admin_api/team_template/service.go).
  Contract tests: [TestUnitServiceDeleteByID](../../../msteams/teams_admin_api/team_template/service_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).

### Remove-SharedWithTeam (CORE-24)

- **__AllParameterSets** — SDK: [Service.UnshareWithTeam](../../../msteams/graph_api/shared_channel_management/service.go), [Service.DeleteByID](../../../msteams/graph_api/shared_channel/service.go).
  Contract tests: [TestUnitSharedTeamOptionalSelectorAndUnshare](../../../msteams/graph_api/shared_channel_management/service_test.go), [TestUnitUnshareFailureIsNotReplayed](../../../msteams/graph_api/shared_channel/service_test.go).
  Live tests: none declared.

### Remove-Team (CORE-28)

- **__AllParameterSets** — SDK: [Service.DeleteByID](../../../msteams/graph_api/team/service.go).
  Contract tests: [TestUnitNativeCRUDRequestDenials](../../../msteams/graph_api/team/service_test.go), [TestUnitMutationFailuresAreNotReplayed](../../../msteams/graph_api/team/service_test.go), [TestUnitLocalValidationAndCancellationPerformNoIO](../../../msteams/graph_api/team/service_test.go), [TestUnitUploadPictureUsesGroupPhotoRouteAndOriginalBytes](../../../msteams/graph_api/team/picture_test.go).
  Live tests: [TestAccTeamLifecycle](../../../tests/integration/lab/graph/team_lifecycle_test.go).

### Remove-TeamChannel (CORE-21)

- **__AllParameterSets** — SDK: [Service.DeleteByID](../../../msteams/graph_api/channel/service.go), [Service.DeleteByDisplayName](../../../msteams/graph_api/channel/names.go).
  Contract tests: [TestUnitNativeCRUDContracts](../../../msteams/graph_api/channel/service_test.go), [TestUnitNameCompositionsUseCompleteUniqueLookup](../../../msteams/graph_api/channel/names_test.go).
  Live tests: [TestAccChannelCRUD](../../../tests/integration/lab/graph/channel_test.go).

### Remove-TeamChannelUser (CORE-20)

- **__AllParameterSets** — SDK: [Service.Remove](../../../msteams/graph_api/channel_user_management/service.go), [Service.ResolveByDisplayName](../../../msteams/graph_api/channel/names.go), [Service.ListAll](../../../msteams/graph_api/channel_member/service.go), [Service.DemoteOwner](../../../msteams/graph_api/channel_member/service.go), [Service.DeleteByID](../../../msteams/graph_api/channel_member/service.go), [Service.ResolveByUser](../../../msteams/graph_api/channel_member/resolution.go), [Resolver.Resolve](../../../msteams/graph_api/directory_user/resolver.go).
  Contract tests: [TestUnitAddRemoveAndOwnerCompositions](../../../msteams/graph_api/channel_user_management/service_test.go), [TestUnitFailuresStopBeforeFurtherChanges](../../../msteams/graph_api/channel_user_management/service_test.go).
  Live tests: [TestAccChannelMembershipCRUD](../../../tests/integration/lab/graph/channel_membership_test.go).

### Remove-TeamTargetingHierarchy (CORE-25)

- **__AllParameterSets** — SDK: [Service.Delete](../../../msteams/teams_admin_api/targeting_hierarchy/service.go), [V1Service.Delete](../../../msteams/teams_admin_api/targeting_hierarchy/native.go), [V2Service.Delete](../../../msteams/teams_admin_api/targeting_hierarchy/native.go).
  Contract tests: [TestUnitNativeStatusRoutesAndIndependentHierarchyPrimitives](../../../msteams/teams_admin_api/targeting_hierarchy/service_test.go), [TestUnitHierarchyDeletionFailuresAreNotReplayed](../../../msteams/teams_admin_api/targeting_hierarchy/failure_test.go).
  Live tests: none declared.

### Remove-TeamUser (CORE-26)

- **__AllParameterSets** — SDK: [Service.Remove](../../../msteams/graph_api/team_user_management/service.go), [Resolver.Resolve](../../../msteams/graph_api/directory_user/resolver.go), [MemberService.RemoveReference](../../../msteams/graph_api/group_relationship/members.go), [MemberService.ListAll](../../../msteams/graph_api/group_relationship/members.go), [OwnerService.RemoveReference](../../../msteams/graph_api/group_relationship/owners.go), [OwnerService.ListAll](../../../msteams/graph_api/group_relationship/owners.go).
  Contract tests: [TestUnitRemoveCompositionPreservesLastOwnerAndPartialWrites](../../../msteams/graph_api/team_user_management/service_test.go), [TestUnitRemoveReadsAllOwnersBeforeLastOwnerGuard](../../../msteams/graph_api/team_user_management/service_test.go), [TestUnitCompositionInvalidInputAndCancellationAreLocal](../../../msteams/graph_api/team_user_management/service_test.go), [TestUnitCompositionMappingsExistInPinnedRuntime](../../../msteams/graph_api/team_user_management/list_test.go).
  Live tests: [TestAccTeamMembershipCRUD](../../../tests/integration/lab/graph/team_membership_test.go).

### Set-CsTeamsTargetingPolicy (CORE-25)

- **Identity** — SDK: [Service.UpdateByIdentity](../../../msteams/teams_admin_api/targeting_policy/service.go), [Service.UpdateGlobal](../../../msteams/teams_admin_api/targeting_policy/service.go).
  Contract tests: [TestUnitTargetingRequestVectors](../../../msteams/teams_admin_api/targeting_policy/service_test.go), [TestUnitSparseIdentityUpdateAndOpenValues](../../../msteams/teams_admin_api/targeting_policy/service_test.go).
  Live tests: none declared.

### Set-Team (CORE-28)

- **__AllParameterSets** — SDK: [Service.UpdateByID](../../../msteams/graph_api/team_management/service.go), [Service.UpdateByID](../../../msteams/graph_api/group/service.go), [Service.UpdateByID](../../../msteams/graph_api/team/service.go), [Service.GetByID](../../../msteams/graph_api/team_management/service.go), [Service.GetByID](../../../msteams/graph_api/group/service.go), [Service.GetByID](../../../msteams/graph_api/team/service.go).
  Contract tests: [TestUnitManagementSparsePropertyAndSettingsInput](../../../msteams/graph_api/team_management/service_test.go), [TestUnitManagementFirstRequestDenials](../../../msteams/graph_api/team_management/service_test.go), [TestUnitSettingsInputsCoverBothPinnedApplicationParameterSets](../../../msteams/graph_api/team_management/creation_settings_test.go), [TestUnitSettingsSerializeEveryInputField](../../../msteams/graph_api/team_management/creation_settings_test.go).
  Live tests: [TestAccTeamLifecycle](../../../tests/integration/lab/graph/team_lifecycle_test.go).

### Set-TeamArchivedState (CORE-28)

- **__AllParameterSets** — SDK: [Service.SetArchivedState](../../../msteams/graph_api/team_management/archive.go), [Service.Archive](../../../msteams/graph_api/team/operations.go), [Service.Unarchive](../../../msteams/graph_api/team/operations.go), [Service.GetByID](../../../msteams/graph_api/team_management/service.go), [Service.GetByID](../../../msteams/graph_api/group/service.go), [Service.GetByID](../../../msteams/graph_api/team/service.go).
  Contract tests: [TestUnitManagementFirstRequestDenials](../../../msteams/graph_api/team_management/service_test.go), [TestUnitArchiveValidationStopsBeforeAnyReadOrWrite](../../../msteams/graph_api/team_management/validation_test.go).
  Live tests: [TestAccTeamLifecycle](../../../tests/integration/lab/graph/team_lifecycle_test.go).

### Set-TeamChannel (CORE-21)

- **__AllParameterSets** — SDK: [Service.UpdateByID](../../../msteams/graph_api/channel/service.go), [Service.UpdateByDisplayName](../../../msteams/graph_api/channel/names.go).
  Contract tests: [TestUnitNativeCRUDContracts](../../../msteams/graph_api/channel/service_test.go), [TestUnitNameCompositionsUseCompleteUniqueLookup](../../../msteams/graph_api/channel/names_test.go).
  Live tests: [TestAccChannelCRUD](../../../tests/integration/lab/graph/channel_test.go).

### Set-TeamPicture (CORE-28)

- **__AllParameterSets** — SDK: [Service.UploadPicture](../../../msteams/graph_api/team/picture.go).
  Contract tests: [TestUnitUploadPictureUsesGroupPhotoRouteAndOriginalBytes](../../../msteams/graph_api/team/picture_test.go), [TestUnitUploadPictureValidationClosesSourcesWithoutIO](../../../msteams/graph_api/team/picture_test.go), [TestUnitUploadPictureEnforcesGraphSizeLimit](../../../msteams/graph_api/team/picture_test.go), [TestUnitNativeCRUDRequestDenials](../../../msteams/graph_api/team/service_test.go).
  Live tests: [TestAccTeamLifecycle](../../../tests/integration/lab/graph/team_lifecycle_test.go).

### Set-TeamTargetingHierarchy (CORE-25)

- **__AllParameterSets** — SDK: [Service.UploadCSV](../../../msteams/teams_admin_api/targeting_hierarchy/service.go), [V1Service.ImportCSV](../../../msteams/teams_admin_api/targeting_hierarchy/native.go), [V2Service.GetHierarchy](../../../msteams/teams_admin_api/targeting_hierarchy/native.go), [V2Service.CreateHierarchy](../../../msteams/teams_admin_api/targeting_hierarchy/native.go), [V2Service.ImportCSV](../../../msteams/teams_admin_api/targeting_hierarchy/native.go).
  Contract tests: [TestUnitNativeImportsStreamOriginalCSVWithoutPreparation](../../../msteams/teams_admin_api/targeting_hierarchy/service_test.go), [TestUnitUploadCompositionDenialsRetainPreparationState](../../../msteams/teams_admin_api/targeting_hierarchy/service_test.go), [TestUnitHierarchyOptionsProfilesAndSourceValidationAreLocal](../../../msteams/teams_admin_api/targeting_hierarchy/failure_test.go).
  Live tests: none declared.

### Update-CsTeamTemplate (CORE-27)

- **Update** — SDK: [Service.UpdateByID](../../../msteams/teams_admin_api/team_template/service.go), [Definition](../../../msteams/teams_admin_api/team_template/definition.go), [NewDefinition](../../../msteams/teams_admin_api/team_template/definition.go), [ParseDefinition](../../../msteams/teams_admin_api/team_template/definition.go).
  Contract tests: [TestUnitServiceUpdateByID](../../../msteams/teams_admin_api/team_template/service_test.go), [TestUnitDefinitionSerializationVectors](../../../msteams/teams_admin_api/team_template/definition_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).
- **UpdateExpanded** — SDK: [Service.UpdateByID](../../../msteams/teams_admin_api/team_template/service.go), [Definition](../../../msteams/teams_admin_api/team_template/definition.go), [NewDefinition](../../../msteams/teams_admin_api/team_template/definition.go), [ParseDefinition](../../../msteams/teams_admin_api/team_template/definition.go).
  Contract tests: [TestUnitServiceUpdateByID](../../../msteams/teams_admin_api/team_template/service_test.go), [TestUnitDefinitionSerializationVectors](../../../msteams/teams_admin_api/team_template/definition_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).
- **UpdateViaIdentity** — SDK: [Service.UpdateByID](../../../msteams/teams_admin_api/team_template/service.go), [Definition](../../../msteams/teams_admin_api/team_template/definition.go), [NewDefinition](../../../msteams/teams_admin_api/team_template/definition.go), [ParseDefinition](../../../msteams/teams_admin_api/team_template/definition.go).
  Contract tests: [TestUnitServiceUpdateByID](../../../msteams/teams_admin_api/team_template/service_test.go), [TestUnitDefinitionSerializationVectors](../../../msteams/teams_admin_api/team_template/definition_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).
- **UpdateViaIdentityExpanded** — SDK: [Service.UpdateByID](../../../msteams/teams_admin_api/team_template/service.go), [Definition](../../../msteams/teams_admin_api/team_template/definition.go), [NewDefinition](../../../msteams/teams_admin_api/team_template/definition.go), [ParseDefinition](../../../msteams/teams_admin_api/team_template/definition.go).
  Contract tests: [TestUnitServiceUpdateByID](../../../msteams/teams_admin_api/team_template/service_test.go), [TestUnitDefinitionSerializationVectors](../../../msteams/teams_admin_api/team_template/definition_test.go).
  Live tests: [TestAccTeamTemplateCRUD](../../../tests/integration/lab/team_template/team_template_test.go).

