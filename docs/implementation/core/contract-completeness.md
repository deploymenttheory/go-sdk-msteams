# CORE contract completeness audit

## Current reviewed repairs (2026-11-26)

This section supersedes the baseline gap descriptions below for the changed tests. Every one of the 52 parameter sets is accounted for; “represented” means the stated local/offline branch, not full command qualification. Existing malformed, denial, cancellation and partial-result expectations remain. Source-backed authored responses carry per-exchange primary provenance; captured template replies retain their existing fixture provenance. No command implementation, contract or live state is promoted in the ledger.

All affected packages passed targeted guarded `TestUnit` runs. The current matrices distinguish successful SDK behavior from external prerequisites and unresolved upstream contracts. Additional permutations or untested fields are not silently claimed.

| Command | Parameter set | Current positive/result/error evidence represented offline | Remaining qualification boundary |
|---|---|---|---|
| Add-TeamChannelUser | `__AllParameterSets` | Channel lookup, resolved user/opaque membership IDs, member/owner writes and retained mutation stages; denied later write and incomplete resolution execute. | Graph private/shared eligibility, ownership grants and propagation. |
| Add-TeamUser | `__AllParameterSets` | Member/owner preflight, successful/skipped add, UPN and later-page resolution; member prefix survives owner denial. | Graph grants, owner eligibility and propagation. |
| Get-AssociatedTeam | `__AllParameterSets` | ID/UPN/mail fallback/delegated-current resolution and complete associated-team rows with native ID/tenant/raw fields. | App/delegated visibility and shared-channel tenant topology. |
| Get-CsBatchTeamsDeploymentStatus | `Get` | Accepted submission identity is reused through ID and Identity status APIs; exact opaque CRLF text retained without terminal inference. | Real deployment/status identities, CSV semantic eligibility and partial deployment outcomes. |
| Get-CsBatchTeamsDeploymentStatus | `GetViaIdentity` | Accepted submission identity is reused through ID and Identity status APIs; exact opaque CRLF text retained without terminal inference. | Real deployment/status identities, CSV semantic eligibility and partial deployment outcomes. |
| Get-CsTeamsTargetingPolicy | `Filter` | Global typed/raw/null object and bare-array filtered read, including case-insensitive tag selection. | Global read-only live proof is scoped to run 45874930078; other identities/flight/cloud/delegated evidence remains. |
| Get-CsTeamsTargetingPolicy | `Identity` | Global typed/raw/null object and bare-array filtered read, including case-insensitive tag selection. | Global read-only live proof is scoped to run 45874930078; other identities/flight/cloud/delegated evidence remains. |
| Get-CsTeamTemplate | `Get` | Existing captured expanded template readback and typed/default projection for both identity input variants. | Current owned-template live settings proof pending; locale/tenant/cloud and other fields remain scoped. |
| Get-CsTeamTemplate | `GetViaIdentity` | Existing captured expanded template readback and typed/default projection for both identity input variants. | Current owned-template live settings proof pending; locale/tenant/cloud and other fields remain scoped. |
| Get-CsTeamTemplateList | `DefaultLocaleOverride` | Existing captured 16-item list, retained raw envelope, typed first entry and incomplete diagnostic flags. | Locale-specific and cross-tenant/cloud live list behavior. |
| Get-MultiGeoRegion | `Identity` | User/group routes decode all three populated or null typed fields and future raw data using pinned model spelling. | Delegated Region.Read.All and actual user/group regional observations. |
| Get-Operation | `__AllParameterSets` | Native notStarted/inProgress/succeeded/failed/future state and typed error/time/raw values; denied/cancelled/retried polls retain handles. | Live team-scoped operation identity, permission and propagation. |
| Get-SharedWithTeam | `__AllParameterSets` | Single object and complete paged shared-team collections; nullable/future raw values retained. | Real shared relationship visibility and cross-tenant topology. |
| Get-SharedWithTeamUser | `__AllParameterSets` | Allowed roster with owner/member/future roles, directory enrichment, stale UPN replacement and nullable UPN; incomplete enrichments retain native data. | Shared roster authorization and actual directory visibility. |
| Get-Team | `Filters` | Paired group/team read preserves group property precedence and both raw snapshots; tenant batch correlation and user joined-team hydration retain successful prefix on later denial. | Licensed owner, Graph grants and eligible owned team lifecycle; search/flight/cloud live conformance. |
| Get-Team | `Identity` | Paired group/team read preserves group property precedence and both raw snapshots; tenant batch correlation and user joined-team hydration retain successful prefix on later denial. | Licensed owner, Graph grants and eligible owned team lifecycle; search/flight/cloud live conformance. |
| Get-TeamAllChannel | `__AllParameterSets` | Complete all-channel pages and typed nullable/future projections with reviewed envelopes. | Hosted/incoming/shared channel topology and access. |
| Get-TeamChannel | `__AllParameterSets` | Read/list/name resolution, query options, complete and empty arrays with reviewed bodies. | Live channel and visibility/permission parity. |
| Get-TeamChannelUser | `__AllParameterSets` | Complete native roster with roles plus chunked directory enrichment; malformed/duplicate/unrequested/later-denied lookup preserves partial native result. | Private/shared roster eligibility, grants and propagation. |
| Get-TeamIncomingChannel | `__AllParameterSets` | Incoming-channel complete paged results use reviewed team-list-incomingchannels response contract. | Cross-tenant shared-channel topology and visibility. |
| Get-TeamTargetingHierarchyStatus | `__AllParameterSets` | v1 and v2 operation wire models feed actual GetStatus; raw v2 Succeeded remains distinct from normalized Successful. | Eligible read-only operation fixture, nested-error live parity and flight/cloud behavior. |
| Get-TeamUser | `__AllParameterSets` | Complete member/owner collection, role filter and resolved directory UPN preserve nullable/future data; partial paging/enrichment execute. | Graph grants and complete tenant roster parity. |
| Get-TenantPrivateChannelMigrationStatus | `__AllParameterSets` | Pinned native model fields, null timestamps and opaque Details text; malformed caller DecodeDetails does not invalidate native 200. | Actual bound native secret/certificate and pinned certificate reads still fail HTTP 401; exact backend prerequisite unresolved. |
| New-CsBatchTeamsDeployment | `__AllParameterSets` | 202 receipt retains instanceId/executionId/future raw fields and addresses opaque status. Does not assert CSV validity or deployment completion. | Licensed owned-team/deployment recipients, tenant execution and real recovery outcomes. |
| New-CsTeamsHiddenMeetingTemplate | `Identity` | Local valid identity construction only; no network operation is implied. | No live service assertion applies to this local constructor. |
| New-CsTeamsHiddenTemplate | `Identity` | Local valid identity construction only; no network operation is implied. | No live service assertion applies to this local constructor. |
| New-CsTeamTemplate | `New` | All four bound Body/Expanded/Identity variants execute captured 201 responses; General-channel insertion and caller immutability remain asserted. | Owned template live field proof is limited to recorded revision; all other fields/cloud/locale remain separate. |
| New-CsTeamTemplate | `NewExpanded` | All four bound Body/Expanded/Identity variants execute captured 201 responses; General-channel insertion and caller immutability remain asserted. | Owned template live field proof is limited to recorded revision; all other fields/cloud/locale remain separate. |
| New-CsTeamTemplate | `NewViaIdentity` | All four bound Body/Expanded/Identity variants execute captured 201 responses; General-channel insertion and caller immutability remain asserted. | Owned template live field proof is limited to recorded revision; all other fields/cloud/locale remain separate. |
| New-CsTeamTemplate | `NewViaIdentityExpanded` | All four bound Body/Expanded/Identity variants execute captured 201 responses; General-channel insertion and caller immutability remain asserted. | Owned template live field proof is limited to recorded revision; all other fields/cloud/locale remain separate. |
| New-Team | `CreateTeam` | Owned group and supplied group both reach teamification and paired readback, preserving distinct ownership and cleanup disposition; existing compensation integration remains. | Licensed owner, Graph grants, propagation and real owned lifecycle cleanup. |
| New-Team | `MigrateGroup` | Owned group and supplied group both reach teamification and paired readback, preserving distinct ownership and cleanup disposition; existing compensation integration remains. | Licensed owner, Graph grants, propagation and real owned lifecycle cleanup. |
| New-TeamChannel | `__AllParameterSets` | Standard/private 201 typed results and shared 202 Content-Location→poll→read; initial share inputs, priority/trust/cancellation remain covered. | Standard/private/shared live lifecycle, grants and eligible channel topology. |
| Remove-CsTeamsTargetingPolicy | `Identity` | Exact DELETE with bodyless 204 and one-attempt failures; not a reset or assignment API. | Owned/disposable policy semantics and live mutation persistence. |
| Remove-CsTeamTemplate | `Delete` | Both identity inputs execute captured empty 204 plus pinned Any-compatible authored 200 {} variant, retaining response metadata. | Owned cleanup live proof at exact tested revision and other cloud/locale behavior. |
| Remove-CsTeamTemplate | `DeleteViaIdentity` | Both identity inputs execute captured empty 204 plus pinned Any-compatible authored 200 {} variant, retaining response metadata. | Owned cleanup live proof at exact tested revision and other cloud/locale behavior. |
| Remove-SharedWithTeam | `__AllParameterSets` | Native DELETE 204 after trusted relationship identity; denied failures remain. | Real shared relationship and permission topology; no guessed create/update relationship route. |
| Remove-Team | `__AllParameterSets` | Native backing-group DELETE 204 with exact ID/path, response and one-attempt count; unrelated picture mapping removed. | Actual owned team/group absence and recoverable live cleanup. |
| Remove-TeamChannel | `__AllParameterSets` | Resolved exact/name channel DELETE 204 with complete lookup and distinct ambiguity/missing failures. | Owned channel absence, propagation and actual deletion authorization. |
| Remove-TeamChannelUser | `__AllParameterSets` | Resolved opaque membership DELETE and owner demotion retain stage/identity; denied mutation and incomplete resolver execute. | Live role eligibility, complete roster and authorization. |
| Remove-TeamTargetingHierarchy | `__AllParameterSets` | Both native v1/v2 DELETE protocols execute bodyless 204, without entity decoding or replay. | Disposable hierarchy and live persistence/absence confirmation. |
| Remove-TeamUser | `__AllParameterSets` | Complete owner preflight guards last owner; owner/member reference removals and partial errors preserve confirmed prefix. | Real owner constraints, complete roster and propagation. |
| Set-CsTeamsTargetingPolicy | `Identity` | Sparse exact PUT executes 204 with explicit empty description, separate from PATCH/create/assignment contracts. | Existing policy upsert/creation semantics remain source-unknown; live disposable mutation and persistence needed. |
| Set-Team | `__AllParameterSets` | Group PATCH precedes five-container team PATCH and paired readback; team denial retains confirmed group stage without fabricated readback. | Live independent backing-group/native effects and partial failure recovery. |
| Set-TeamArchivedState | `__AllParameterSets` | Archive/unarchive pre-read→202→operation succeeded→fresh paired readback; mismatched state remains unconfirmed despite successful operation. | Actual asynchronous lifecycle and independent state convergence. |
| Set-TeamChannel | `__AllParameterSets` | Resolved exact/name PATCH 204 and nullable typed readbacks; malformed/missing/ambiguous resolution remains error. | Owned channel mutation and live field parity. |
| Set-TeamPicture | `__AllParameterSets` | 200/204 upload verifies original bytes, media type, one attempt and exactly-once source closure; 503 remains failure. | Actual group-photo persistence and PowerShell parity. |
| Set-TeamTargetingHierarchy | `__AllParameterSets` | v1 optional-handle import and v2 existing/create preparation→202 import; later denied import retains successful preparation receipts. | Real ingestion/error/recovery against a disposable tenant hierarchy; its published column layout is represented offline. |
| Update-CsTeamTemplate | `Update` | Four input variants execute captured 200 and source-backed 201 receipt; response decoding distinct from definition serialization. | Current owned-template settings proof pending; other field/locale/cloud live parity remains. |
| Update-CsTeamTemplate | `UpdateExpanded` | Four input variants execute captured 200 and source-backed 201 receipt; response decoding distinct from definition serialization. | Current owned-template settings proof pending; other field/locale/cloud live parity remains. |
| Update-CsTeamTemplate | `UpdateViaIdentity` | Four input variants execute captured 200 and source-backed 201 receipt; response decoding distinct from definition serialization. | Current owned-template settings proof pending; other field/locale/cloud live parity remains. |
| Update-CsTeamTemplate | `UpdateViaIdentityExpanded` | Four input variants execute captured 200 and source-backed 201 receipt; response decoding distinct from definition serialization. | Current owned-template settings proof pending; other field/locale/cloud live parity remains. |

Exact test function references are maintained per set in `contracts/coverage/8.1.0.json`; the baseline appendix below explains why the old mappings alone overstated actual execution. Newly added admin/native/composed-team success functions are explicitly mapped. The MigrateGroup set no longer cites CreateTeam-only input validation/request tests, and Remove-Team no longer cites photo upload.

### Source-unknown versus external-only work

Initial shared-channel `sharedWithTeams` creation is evidenced and implemented; this does not establish POST/PATCH of an existing relationship. The current primary relationship surface only establishes read/delete/allowed-members. Targeting policy PUT is evidenced for updates; creation/upsert semantics are not inferred from the verb. Hierarchy and batch wrappers preserve/stream CSV bytes and receipts; their public column layouts are available and used in authored inputs. Accepted HTTP responses do not establish tenant identity eligibility or completed processing. These source limits stay explicit instead of becoming invented SDK routes or success claims.

Graph success vectors derive from the cited primitive contracts and retained raw response shapes; they are authored offline. Template success uses already retained captures and source-backed 200/201/204 variants. Admin response schemas/handlers were independently settled against pinned 8.1.0 assemblies in [authored-protocol-sources.md](authored-protocol-sources.md), including exact hashes and IL offsets. Unknown/null preservation beyond pinned round-tripping is labeled as SDK behavior.

## Immutable pre-repair baseline

Source baseline: `22f01edc43d70cdbeffdf5ea850d38b2a2337952` (2026-11-26). This is a read-only source audit, not test execution. Every source citation below is pinned to that commit so concurrent repairs cannot change the meaning of this baseline. Scope is the 40 commands / 52 parameter sets whose implementation owner starts `CORE-` in the [coverage ledger](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/contracts/coverage/8.1.0.json#L1). The appendix accounts for all 78 distinct mapped functions across 26 files. Counts describe mappings, not qualified contracts. No ledger promotion is warranted by this audit.

## What is established offline

The nine shared-root integration functions reach authored success and failure responses without the per-package request-boundary transformation. They establish selected owner composition, native membership identities/recovery, owned-group compensation, shared-channel acceptance/poll/read and trust/cancellation paths. Native channel membership integration uses primitives; mapping it to a command does not establish the command composition's lookup, enrichment or complete result shape. The [integrated qualification report](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/integrated-qualification.md#L1) records primary Graph provenance, live prerequisites and the exact limited execution evidence already obtained by its author. Its execution statements are not new execution evidence from this audit.

Only the two hidden-template Identity constructors have their entire intended local-only operation represented by mapped valid-result tests. The mapped function does not include invalid IDs; separate local tests do. Local definition/setting serialization, ZIP construction and hierarchy projection tests also genuinely exercise selected values. Every network command still has either a missing success/result path, incomplete variants/models/errors, or live qualification outstanding. A synthetic success proves conformance to that vector; it cannot prove tenant permission, eligibility, propagation or PowerShell parity.

## Runner semantics used in this audit

- **B**: package `run`/`runCase` calls `requestBoundary`, including from another test file. For nonlocal cases it replaces the first 2xx response with403, truncates later exchanges, invokes with a boundary context, verifies forbidden classification/available response status, then returns an empty observation and replaces the original expectation with empty. Local/no-step cases and any single-step case with `Want.Error != nil` bypass this rewrite. Such negative exceptions really can decode a successful page before a local semantic error. Several positive response bodies are blank even before transformation; merely disabling the wrapper would not supply a valid reviewed vector.
- **D**: direct `contracttest.Run`. Region/migration/hierarchy vectors explicitly supply403. Batch/targeting/template use successful config discovery followed by explicit403, so discovery success is not management success.
- **I**: shared-root integration runner calls the direct harness and preserves all authored exchanges. **L**: direct local construction, projection, serialization or no-step validation. **F**: native team package custom `fixture`; ordinary200/201 exchanges become403 and truncate, but204 survives. Names containing Reject/Invalid/Validation/Malformed/Failure/Failures/Envelope/Local/Wrong/Error preserve exchanges. Follow actual conditions, not the descriptive test name.

Runner primary sources: [channel boundary](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel/request_boundary_test.go#L1), [team-user boundary](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/request_boundary_test.go#L1), [channel-user boundary](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel_user_management/request_boundary_test.go#L1), [shared-management boundary](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/shared_channel_management/request_boundary_test.go#L1), [operation boundary](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_operation/request_boundary_test.go#L1), [team fixture](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team/service_test.go#L1), [shared harness](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/internal/contracttest/runner.go#L1), [integration runner](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L1). Each appendix entry names its actual runner and behavior.

## Per-parameter-set matrix

`S` means successful command path represented; `R/M` means returned result/model represented. `partial` always names its limit. `no` concerns mapped functions only; additional unmapped tests are discussed separately. `E` records mapped local/protocol/error paths. The last column separates work implementable offline from external/source evidence still required. Exact mapped functions are linked by appendix ID; all sets are listed separately, even where they reuse the same Go method.

| Command / owner | Parameter set | S; R/M; E established by mapped source | Implementable offline gap; external/source limit | Tests |
|---|---|---|---|---|
| Add-TeamChannelUser / CORE-20 | `__AllParameterSets` | S partial native create/promote; R/M partial identity/role; E primitive promotion/recovery denial | Complete Add composition channel/user lookup, add/demotion variants and stages; live private/shared eligibility and grants | [T01](#t01), [T02](#t02), [T03](#t03) |
| Add-TeamUser / CORE-26 | `__AllParameterSets` | S owner composition; R/M selected stages/identity; E preflight/partial-owner/recovery/local/bounded page | Add member/skip/UPN/paging variants with reviewed replies; live grants/propagation/parity | [T04](#t04), [T05](#t05), [T06](#t06), [T07](#t07), [T08](#t08), [T09](#t09) |
| Get-AssociatedTeam / CORE-24 | `__AllParameterSets` | S no; R/M no; E controlled denial, direct404 fallback and local mode/options | Complete associated-team list/paging and resolver result source variants; live delegated/app visibility | [T10](#t10), [T11](#t11) |
| Get-CsBatchTeamsDeploymentStatus / CORE-22 | `Get` | S no; R/M no; E actual403/400/404 after discovery | Test raw successful status and ID/identity equivalence only against pinned opaque-text contract; live orchestration evidence | [T12](#t12), [T13](#t13) |
| Get-CsBatchTeamsDeploymentStatus / CORE-22 | `GetViaIdentity` | S no; R/M no; E actual403/400/404 after discovery | Test raw successful status and ID/identity equivalence only against pinned opaque-text contract; live orchestration evidence | [T12](#t12), [T13](#t13) |
| Get-CsTeamsTargetingPolicy / CORE-25 | `Filter` | S no; R/M no; E403 and local filter validation | Successful list/entity/filter values and nullable/open fields need reviewed vectors; live config grants and result evidence | [T14](#t14), [T15](#t15) |
| Get-CsTeamsTargetingPolicy / CORE-25 | `Identity` | S no; R/M no; E403 and local filter validation | Successful list/entity/filter values and nullable/open fields need reviewed vectors; live config grants and result evidence | [T14](#t14), [T15](#t15) |
| Get-CsTeamTemplate / CORE-27 | `Get` | S no; R/M no; E403 after discovery | Decode valid detail/defaults for both identity inputs; pinned module captures/first-party evidence required, then live parity | [T16](#t16) |
| Get-CsTeamTemplate / CORE-27 | `GetViaIdentity` | S no; R/M no; E403 after discovery | Decode valid detail/defaults for both identity inputs; pinned module captures/first-party evidence required, then live parity | [T16](#t16) |
| Get-CsTeamTemplateList / CORE-27 | `DefaultLocaleOverride` | S no; R/M no; E403 after discovery | Valid list/default/fr-FR and incomplete diagnostic projection; pinned valid envelope source, live locale/tenant evidence | [T17](#t17) |
| Get-MultiGeoRegion / CORE-22 | `Identity` | S no; R/M no; E403 exact delegated scope, local mode/cloud errors | Review valid region/nullable/time/raw result vectors; real delegated Region.Read.All grant and region response | [T18](#t18), [T19](#t19) |
| Get-Operation / CORE-23 | `__AllParameterSets` | S selected succeeded read/poll; R/M target ID/state; E malformed/status/auth/denied resumable handle | Native accepted/running/failed/future typed error/time/raw fields need nonblank reviewed replies; live operation identity/cadence | [T20](#t20), [T21](#t21), [T22](#t22), [T23](#t23) |
| Get-SharedWithTeam / CORE-24 | `__AllParameterSets` | S no; R/M no; E selector-route denial | Collection/single results, host/tenant/null metadata and paging; live sharing prerequisite/grants | [T24](#t24) |
| Get-SharedWithTeamUser / CORE-24 | `__AllParameterSets` | S no; R/M no; E first native-read403 and local options | Native rows, role filter/enrichment/nulls/paging and later enrichment failures; live external membership visibility | [T25](#t25), [T26](#t26), [T11](#t11) |
| Get-Team / CORE-28 | `Filters` | S no; R/M no; E discovery/identity403 and local validation | Discovery+batch/hydration/filter or paired identity reads, completeness/partial results; live grants/search propagation | [T27](#t27), [T28](#t28) |
| Get-Team / CORE-28 | `Identity` | S no; R/M no; E discovery/identity403 and local validation | Discovery+batch/hydration/filter or paired identity reads, completeness/partial results; live grants/search propagation | [T29](#t29), [T30](#t30) |
| Get-TeamAllChannel / CORE-21 | `__AllParameterSets` | S no; R/M no; E controlled403 (foreign/bounds subcases use hosted collection) | Successful multi-page allChannels traversal and wrapper model result; live channel visibility | [T31](#t31), [T32](#t32) |
| Get-TeamChannel / CORE-21 | `__AllParameterSets` | S selected shared-channel GET; R/M ID/type; E403/partial-page foreign/bounds | Hosted full collection/options/paging and complete field presence/unknown values; live private/shared visibility | [T31](#t31), [T23](#t23) |
| Get-TeamChannelUser / CORE-20 | `__AllParameterSets` | S partial native empty roster; R/M empty primitive list only; E native recovery failure | ChannelUserManagement.List lookup/role order/enriched nulls and later failure retention; live roster/directory grants | [T33](#t33), [T34](#t34), [T03](#t03) |
| Get-TeamIncomingChannel / CORE-21 | `__AllParameterSets` | S no; R/M no; E controlled403 (foreign/bounds subcases use hosted collection) | Successful incomingChannels traversal/wrapper and typed result; live incoming shared channel prerequisite | [T31](#t31), [T32](#t32) |
| Get-TeamTargetingHierarchyStatus / CORE-25 | `__AllParameterSets` | S no service success; R/M local v1/v2 projections; E route403 | Wire operation success then service projection, future/raw/null/malformed nested status; reviewed pinned protocol and live hierarchy evidence | [T35](#t35), [T36](#t36), [T37](#t37) |
| Get-TeamUser / CORE-26 | `__AllParameterSets` | S no; R/M no; E first owner/member403; metadata mapping local | Owner/member/guest filtering, dedup, nullable projection and retained owner prefix after later failure; live directory grants | [T38](#t38), [T39](#t39), [T07](#t07) |
| Get-TenantPrivateChannelMigrationStatus / CORE-22 | `__AllParameterSets` | S no; R/M no; E cloud/mode403, local destination/cancel | Valid native status/Details/time/raw response only from reviewed contract; live migration prerequisite and service audience grants | [T40](#t40), [T41](#t41) |
| New-CsBatchTeamsDeployment / CORE-22 | `__AllParameterSets` | S no submission; R/M local archive; E403/400/404/429/503 submission | Submission receipt optional identities/raw fields; reviewed pinned response and real orchestration/notify prerequisites | [T42](#t42), [T43](#t43), [T13](#t13) |
| New-CsTeamsHiddenMeetingTemplate / CORE-27 | `Identity` | S local yes; R/M exact ID JSON; E no mapped invalid-ID case | Local empty-ID error exists outside mapping; no management/live requirement for constructor | [T44](#t44) |
| New-CsTeamsHiddenTemplate / CORE-27 | `Identity` | S local yes; R/M exact ID JSON; E no mapped invalid-ID case | Local empty-ID error exists outside mapping; no management/live requirement for constructor | [T44](#t44) |
| New-CsTeamTemplate / CORE-27 | `New` | S no HTTP success; R/M definition input local; E403 | Valid create receipt and Body/Expanded/Identity variants; reviewed pinned source then live template permissions/parity | [T45](#t45), [T46](#t46) |
| New-CsTeamTemplate / CORE-27 | `NewExpanded` | S no HTTP success; R/M definition input local; E403 | Valid create receipt and Body/Expanded/Identity variants; reviewed pinned source then live template permissions/parity | [T45](#t45), [T46](#t46) |
| New-CsTeamTemplate / CORE-27 | `NewViaIdentity` | S no HTTP success; R/M definition input local; E403 | Valid create receipt and Body/Expanded/Identity variants; reviewed pinned source then live template permissions/parity | [T45](#t45), [T46](#t46) |
| New-CsTeamTemplate / CORE-27 | `NewViaIdentityExpanded` | S no HTTP success; R/M definition input local; E403 | Valid create receipt and Body/Expanded/Identity variants; reviewed pinned source then live template permissions/parity | [T45](#t45), [T46](#t46) |
| New-Team / CORE-28 | `CreateTeam` | S owned group creation only; R/M partial create receipt/cleanup ownership; E primary403/compensation | Complete provisioning, hydration/settings/readback and failed cleanup variants; live group/team propagation/grants/parity | [T47](#t47), [T48](#t48), [T49](#t49), [T50](#t50) |
| New-Team / CORE-28 | `MigrateGroup` | S no provisioning; R/M supplied-group retention on403; E prerequisite denial; create-validation mapping does not invoke MigrateGroup | Complete provisioning, hydration/settings/readback and failed cleanup variants; live group/team propagation/grants/parity | [T47](#t47), [T51](#t51), [T50](#t50) |
| New-TeamChannel / CORE-21 | `__AllParameterSets` | S shared accepted→poll→read; R/M selected shared model/handle; E trust/cancel/local sharing/403 | Standard/private created models, owner-UPN composition and more status/model branches; live owner/channel eligibility/shared propagation | [T52](#t52), [T53](#t53), [T54](#t54), [T23](#t23), [T55](#t55), [T56](#t56), [T57](#t57), [T58](#t58) |
| Remove-CsTeamsTargetingPolicy / CORE-25 | `Identity` | S no; R/M no success response; E403/429/503 and mapped negative read errors | Successful delete receipt/status variants from reviewed contract; live policy grants | [T14](#t14), [T59](#t59) |
| Remove-CsTeamTemplate / CORE-27 | `Delete` | S no; R/M no receipt; E403 | Declared successful status/optional receipt body and both identity forms; pinned receipt source and live parity | [T60](#t60) |
| Remove-CsTeamTemplate / CORE-27 | `DeleteViaIdentity` | S no; R/M no receipt; E403 | Declared successful status/optional receipt body and both identity forms; pinned receipt source and live parity | [T60](#t60) |
| Remove-SharedWithTeam / CORE-24 | `__AllParameterSets` | S no; R/M no success; E403/429/503 no replay | Successful wrapper/native unshare acknowledgement; live relationship ownership/grants | [T24](#t24), [T61](#t61) |
| Remove-Team / CORE-28 | `__AllParameterSets` | S no delete; R/M no delete receipt; E403/503, local invalid ID | Group DELETE204 actual success and returned metadata; picture test mapping is unrelated to delete; live owned group cleanup | [T62](#t62), [T63](#t63), [T64](#t64), [T65](#t65) |
| Remove-TeamChannel / CORE-21 | `__AllParameterSets` | S no; R/M no delete success; E first403 and actual missing/ambiguous/page-limit lookup | Successful complete unique lookup followed by DELETE204 and explicit ID path; live owned channel cleanup | [T52](#t52), [T66](#t66) |
| Remove-TeamChannelUser / CORE-20 | `__AllParameterSets` | S partial primitive DELETE/recovery; R/M primitive identity only; E recovery denial | Complete Remove composition/demotion and membership resolution stages; live private/shared relationship ownership | [T01](#t01), [T02](#t02), [T03](#t03) |
| Remove-TeamTargetingHierarchy / CORE-25 | `__AllParameterSets` | S no; R/M no success; E v1/v2 403/429/503 | Successful declared delete status/receipt; reviewed version-specific protocol and live hierarchy grants | [T35](#t35), [T67](#t67) |
| Remove-TeamUser / CORE-26 | `__AllParameterSets` | S no composition delete; R/M selected last-owner model; E last-owner guard/local/first403 | Complete owner traversal and demotion/member-reference deletes, skip/partial results; live owner invariants/grants | [T68](#t68), [T69](#t69), [T06](#t06), [T07](#t07) |
| Set-CsTeamsTargetingPolicy / CORE-25 | `Identity` | S no; R/M sparse request input; E403/local identity | Successful update status/receipt from reviewed contract; live policy permissions | [T14](#t14), [T70](#t70) |
| Set-Team / CORE-28 | `__AllParameterSets` | S no; R/M local settings serialization only; E first PATCH403 | Paired group/team writes and readback including partial stage failure; live directory/team grant and consistency | [T71](#t71), [T29](#t29), [T48](#t48), [T49](#t49) |
| Set-TeamArchivedState / CORE-28 | `__AllParameterSets` | S no; R/M no successful archive; E preread403/local options | Paired preread, archive/unarchive acceptance/poll/readback and state flags; live archive/SPO permission and propagation | [T29](#t29), [T72](#t72) |
| Set-TeamChannel / CORE-21 | `__AllParameterSets` | S no; R/M sparse request only; E first403 | PATCH204 and complete unique-name lookup→PATCH; returned metadata and live channel mutation grants | [T52](#t52), [T66](#t66) |
| Set-TeamPicture / CORE-28 | `__AllParameterSets` | S204 upload; R/M response metadata only; E503/local/size bound | Ordinary200 success still rewritten, review acknowledgement variants; live group-photo grants | [T65](#t65), [T73](#t73), [T74](#t74), [T62](#t62) |
| Set-TeamTargetingHierarchy / CORE-25 | `__AllParameterSets` | S no import success; R/M actual negative preparation flags; E403/read404→create403/local/source closure | v1/v2 valid import handle and preparation-existing/create→import chains; reviewed protocol source plus live version/grants | [T75](#t75), [T76](#t76), [T77](#t77) |
| Update-CsTeamTemplate / CORE-27 | `Update` | S no HTTP success; R/M definition input local; E403 | Valid update receipt for Body/Expanded/Identity variants; reviewed pinned source then live permissions/parity | [T78](#t78), [T46](#t46) |
| Update-CsTeamTemplate / CORE-27 | `UpdateExpanded` | S no HTTP success; R/M definition input local; E403 | Valid update receipt for Body/Expanded/Identity variants; reviewed pinned source then live permissions/parity | [T78](#t78), [T46](#t46) |
| Update-CsTeamTemplate / CORE-27 | `UpdateViaIdentity` | S no HTTP success; R/M definition input local; E403 | Valid update receipt for Body/Expanded/Identity variants; reviewed pinned source then live permissions/parity | [T78](#t78), [T46](#t46) |
| Update-CsTeamTemplate / CORE-27 | `UpdateViaIdentityExpanded` | S no HTTP success; R/M definition input local; E403 | Valid update receipt for Body/Expanded/Identity variants; reviewed pinned source then live permissions/parity | [T78](#t78), [T46](#t46) |

## Remaining evidence boundaries

**Implementable offline:** extend selected Graph authored vectors using the primary links and provenance already reviewed in the integrated report; supply valid bodies and assert actual result models, stage state, pagination, identity and response metadata. Existing boundary tests remain useful denial tests. Repairing or adding a successful vector must include the entire ordered chain, and later-stage failure cases must actually pass their successful prerequisites. Do not count an input serializer as a service response model. Do not count a successful discovery response as command success. Do not infer parity across parameter sets from shared symbols; inspect their distinct caller-input variants.

**External or source-unknown:** regional/migration/batch/hierarchy/template successful schemas, status acceptance and result projection must be grounded in the pinned module/first-party materials identified by each acceptance note before authoring values. Where those materials do not establish a field/route/receipt, obtain the missing primary source or sanitized tenant capture rather than inventing it. This audit does not independently verify external documentation or introduce protocol claims beyond those repository sources. Tenant roles/scopes, delegated service-specific consent, dedicated SDK principal grants, eligible preexisting team/channel/hierarchy/orchestration state and actual propagation/cleanup remain external. Bootstrap application permissions only enable provisioning; they are not evidence that the dedicated SDK application/delegated identity has the tested service grant.

Primary repository acceptance/provenance notes: [baseline-inspection.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/baseline-inspection.md#L1), [membership-resource-contract.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/membership-resource-contract.md#L1), [deployment-migration-lab-acceptance.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/deployment-migration-lab-acceptance.md#L1), [targeting-lab-acceptance.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/targeting-lab-acceptance.md#L1), [team-lifecycle-lab-acceptance.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/team-lifecycle-lab-acceptance.md#L1), [channel-lab-acceptance.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/channel-lab-acceptance.md#L1), [channel-membership-lab-acceptance.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/channel-membership-lab-acceptance.md#L1), [membership-lab-acceptance.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/membership-lab-acceptance.md#L1), [shared-channel-lab-acceptance.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/shared-channel-lab-acceptance.md#L1), [team-operation-lab-acceptance.md](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/docs/implementation/core/team-operation-lab-acceptance.md#L1). Live evidence requirements are enumerated in the integrated report; none are satisfied by this source inspection.

### Additional tests are not mapped evidence

The inspected files contain useful unmapped tests: region malformed responses, migration local Details parsing/malformed responses, batch submission local optional identities and local validation, targeting local raw-fields model and cancellation, hierarchy malformed status/import handles and tenant binding, template malformed-success/typed error contracts and local empty-ID constructors. These are separate from the 78 mapped functions and must not silently inflate a per-set ledger claim. Likewise, channel/team operation local model tests outside these files are not included merely because their package has a ledger reference. Any mapping repair should choose exact affected commands/sets and preserve qualification status until its criteria are met.

## Appendix: every mapped function and actual flow

Each entry is pinned to its function definition. B/D/I/L/F refer to the runner definitions above. This is inspection of code paths and configured exchanges; no claim is made that tests were executed during this audit.

### T01

[TestUnitAddRemoveAndOwnerCompositions](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel_user_management/service_test.go#L79) — B: all add/promote/demote/delete compositions stop at channel lookup403; neither user/member resolution nor final writes/stage success assertions execute.

### T02

[TestUnitFailuresStopBeforeFurtherChanges](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel_user_management/service_test.go#L151) — B: first channel lookup403 truncates later write403/429/503 and missing/ambiguous/incomplete roster branches; those intended later errors are unexercised.

### T03

[TestUnitIntegratedCoreChannelMembershipOwnershipAndRecovery](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L107) — I: actual native team-member GET200, channel-member POST201/PATCH200/DELETE204, complete empty roster GET200 and unchanged team-member reread; promotion403 with successful or denied explicit recovery also executes. Uses primitives, not ChannelUserManagement composition or enriched List.

### T04

[TestUnitAddCompositionStagesAndCorrectedPreflight](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/service_test.go#L80) — B: positive/preexisting/partial owner chains become first member-read403. Dedicated one-step forbidden-preflight retains actual unconfirmed stage assertion; later owner denial is unreachable.

### T05

[TestUnitAddCompletePreflightAndResolvedIdentity](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/service_test.go#L107) — B plus preserved single-step negative: partial-cannot-establish-absence decodes first member page and ErrPageLimit; later-member paging and UPN success become first request403.

### T06

[TestUnitCompositionInvalidInputAndCancellationAreLocal](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/service_test.go#L176) — L: add/remove/list invalid role/group/bounds, cancellation, closed and nil service, with no declared HTTP steps.

### T07

[TestUnitCompositionMappingsExistInPinnedRuntime](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/list_test.go#L79) — L: reads pinned runtime and asserts three command/set names exist. No SDK call or request/response/model evidence.

### T08

[TestUnitIntegratedCoreOwnerAndNativeMembershipIdentity](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L73) — I: full TeamUserManagement.Add owner chain (member preflight/write, owner preflight/write), native team-membership GET and DELETE; asserts distinct directory/native IDs and successful stages.

### T09

[TestUnitIntegratedCorePartialOwnerFailureAndExplicitRecovery](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L203) — I: successful member reference write then owner403; asserts five stages and partial state; explicit reference DELETE204 plus independent empty GET, or DELETE403 with no confirmation read.

### T10

[TestUnitAssociatedTeamsSelectionAndResolution](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/shared_channel_management/service_test.go#L95) — B: explicit UUID or delegated /me request403; UPN resolution403; fallback retains direct404 then mail-query403. Application current-user selection is a separate L error. No associated-team decoding.

### T11

[TestUnitCompositionsRejectOptionsAndCancellationBeforeCredentials](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/shared_channel_management/service_test.go#L167) — L: invalid allowed-user role/lookup/traversal options, invalid associated traversal/resolve options and cancelled unshare before credentials.

### T12

[TestUnitBatchStatusIsOpaqueAndIdentityInputIsNotAURI](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/batch_deployment/service_test.go#L80) — D: discovery200 then actual status GET403 for ID and identity sets; escaped opaque orchestration identity asserted; value.Text callback has no successful value.

### T13

[TestUnitBatchErrorsRemainErrorsAndSubmissionIsNotReplayed](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/batch_deployment/service_test.go#L107) — D: discovery200 then create400/404/429/503 or status400/404; raw bodies/APIError retained; rejecting sequence prevents submission replay. No success model.

### T14

[TestUnitTargetingRequestVectors](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_policy/service_test.go#L50) — D: discovery200 then list/read/PUT/DELETE403; exact request/body/header checks, no policy result decoding.

### T15

[TestUnitTargetingFilterRequestAndLocalValidation](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_policy/service_test.go#L120) — D plus L: discovery200 then list-filter403; invalid wildcard range rejected locally. Successful local filtering never executes.

### T16

[TestUnitServiceGetByID](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/team_template/service_test.go#L138) — D: serviceCase discovery200 then template GET403 for both sets; generated detail-default assertion is guarded by err==nil and never executes.

### T17

[TestUnitServiceList](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/team_template/service_test.go#L164) — D: discovery200 then template list403 for default/fr-FR/partial diagnostics cases. IsIncomplete/diagnostic success checks never execute.

### T18

[TestUnitRegionUsesExactDelegatedScopeAndDirectEntityRoutes](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/region/service_test.go#L34) — D: user/group region GET403 in commercial/GCC with exact delegated Region.Read.All token scope; region/time/raw-fields callback never obtains a model.

### T19

[TestUnitApplicationRegionAndUnqualifiedCloudsRejectBeforeTokens](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/region/service_test.go#L60) — L: application mode refused and five unqualified cloud profiles rejected before token acquisition.

### T20

[TestUnitGetOperationPreservesNativeStatesAndFutureData](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_operation/service_test.go#L47) — B: all five native operation state vectors become GET403; raw/future/error-model assertions never execute (success reply is blank even before rewrite).

### T21

[TestUnitGetRejectsMalformedAndUnexpectedResponses](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_operation/service_test.go#L71) — B bypass: single-step malformed/null/array/missing-ID/invalid time/error responses retain decode errors; unexpected202/206 and explicit401/403 retain typed failures.

### T22

[TestUnitWaitDenialRetainsResumableHandle](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_operation/service_test.go#L95) — B bypass: explicit first poll403, retaining resumable operation ID and location; no successful poll/terminal state.

### T23

[TestUnitIntegratedCoreSharedChannelOperationAndRead](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L291) — I: shared channel POST202 via Location/Content-Location; wait GET200 then typed operation GET200 then resulting channel GET200. Omitted/empty/host-team sharing variants, accepted-versus-completed distinction and target identity assertions execute.

### T24

[TestUnitSharedTeamOptionalSelectorAndUnshare](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/shared_channel_management/service_test.go#L68) — B: collection/single GET and unshare DELETE become403; optional selector routes are checked, successful item/page/host assertions are skipped.

### T25

[TestUnitAllowedUsersProjectionFilterAndNullableEnrichment](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/shared_channel_management/service_test.go#L131) — B: allowedMembers first GET403; role filtering, native projections and directory enrichment requests/results never execute.

### T26

[TestUnitAllowedUserFailuresRetainNativeState](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/shared_channel_management/service_test.go#L146) — B: allowedMembers first GET403 truncates directory403/duplicate/unexpected-user/malformed branches; native partial rows and later failure assertions are unexercised.

### T27

[TestUnitManagementListDiscoveryDenial](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/service_test.go#L114) — D: tenant/search beta group-discovery GET403; asserts no items and incomplete discovery/result, but no batching/hydration/result success.

### T28

[TestUnitListingLocalValidationAndCurrentUserMode](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/validation_test.go#L116) — L: list batch/attempt/timeout/delay/paging/search validation and application current-user rejection before credentials.

### T29

[TestUnitManagementFirstRequestDenials](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/service_test.go#L74) — D: paired-read GET, group PATCH, migration prerequisite GET, archive/unarchive prerequisite GET each403; no later team request/readback or successful archive path.

### T30

[TestUnitInvalidInputAndCancellationPerformNoIO](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/validation_test.go#L160) — L: management get/update invalid ID, nil/cancelled context, closed/nil service and wrong backend with no HTTP.

### T31

[TestUnitCollectionsAndPaging](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel/service_test.go#L121) — B: hosted/all/incoming positive pages and filter/select/prefer become first403; single-step foreign-link and MaxPages negatives bypass rewrite and decode a partial page, then fail with ErrDestination/ErrPageLimit.

### T32

[TestUnitExplicitCollectionWrappers](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel/names_test.go#L58) — B: ListAll/ListIncoming wrappers become first collection GET403; blank replies and expected null do not prove successful empty collection.

### T33

[TestUnitListRoleOrderPresenceAndPagedEnrichment](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel_user_management/list_test.go#L36) — B: channel lookup403 stops roster, role ordering and paged UPN enrichment; no projected row or completeness result proof.

### T34

[TestUnitEnrichmentFailureNeverProducesACompleteList](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel_user_management/list_test.go#L93) — B: channel lookup403 truncates roster and every intended enrichment denial/decode/status/ambiguity/foreign/bounds error; post-enrichment native-state assertions are skipped by context guard.

### T35

[TestUnitNativeStatusRoutesAndIndependentHierarchyPrimitives](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_hierarchy/service_test.go#L92) — D: v1/v2 latest/opaque status GET403, native hierarchy GET/POST403 and hierarchy DELETE403; only routes/errors, not operation/receipt decoding.

### T36

[TestUnitStatusProjectionFromLocalOperationValues](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_hierarchy/projection_test.go#L25) — L: unmarshals local HierarchyOperation values then projects native statuses/time/file fields; maps known states and preserves selected unknowns, without service GetStatus.

### T37

[TestUnitStatusErrorProjectionFromLocalOperationValues](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_hierarchy/projection_test.go#L48) — L: local v1/v2 operation decoding and projection; ordered/grouped nested errors, nulls, numeric/case/combined/whitespace statuses; no service HTTP success.

### T38

[TestUnitListCompositionOwnerPrecedenceProjectionAndFilters](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/list_test.go#L33) — B: owner-only/member-only/all/guest listing each first read403; successful dedup/order/nullable role projection never executes.

### T39

[TestUnitListCompositionPreservesOwnerPrefixOnMemberReadFailure](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/list_test.go#L69) — B: owner GET rewritten403, so later member403 and retained-owner prefix/completeness assertion never execute.

### T40

[TestUnitMigrationCloudRequestPrefixesAndAudiences](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/private_channel_migration/service_test.go#L32) — D: ten cloud/mode migration GET403 routes/audiences/cmdlet headers; nullable status/details/raw callbacks receive no model.

### T41

[TestUnitMigrationProfilesDoNotFallbackAndWrongBackendIsLocal](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/private_channel_migration/service_test.go#L59) — L: France/Germany destination rejection, wrong-backend rejection and cancellation; no HTTP fallback.

### T42

[TestUnitArchivePreservesOriginalBytesFixedNamesAndReaderOwnership](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/batch_deployment/archive_test.go#L64) — L: actual BuildArchive result, ZIP fixed names/Deflate/mode/original bytes, borrowed reader ownership, Base64, defensive byte copy and formatting redaction.

### T43

[TestUnitBatchSubmissionRequestAndDenial](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/batch_deployment/service_test.go#L71) — D: archive creation then discovery200 and actual submission POST403; outgoing encoded CSVs/notify body asserted, no submission receipt.

### T44

[TestUnitHiddenBuildersHaveNoManagementOrAuthIO](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/team_template/local_test.go#L50) — L: both hidden constructors return exact ID JSON through local contract harness with no token/HTTP steps; this mapped function covers valid ID, not empty-ID errors.

### T45

[TestUnitServiceCreate](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/team_template/service_test.go#L200) — D plus L inputs: definition construction/parse succeeds; all four sets discovery200 then POST403. General channel append request checked, create receipt unexercised.

### T46

[TestUnitDefinitionSerializationVectors](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/team_template/definition_test.go#L71) — L: actual ParseDefinition/NewDefinition JSON serialization for at least12 selected vectors: defaults, sparse/false/empty arrays, nested/root precedence, null handling and tab/channel shape. Does not establish HTTP result models.

### T47

[TestUnitManagementCreateRequestVariantsDenied](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/service_test.go#L134) — D: ordinary/class/PLC group POST403; exact presets/owner/member binds asserted and no fabricated group ownership/cleanup. Supplied-group path covered elsewhere, not by this function.

### T48

[TestUnitSettingsInputsCoverBothPinnedApplicationParameterSets](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/creation_settings_test.go#L64) — L: pinned runtime/coverage metadata checks all settings parameters for New-Team CreateTeam and Set-Team; no HTTP or result model.

### T49

[TestUnitSettingsSerializeEveryInputField](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/creation_settings_test.go#L124) — L: each typed Boolean true/false and Giphy rating serializes to independent JSON field/value expectation; no request delivery or result.

### T50

[TestUnitIntegratedCoreCreationOwnershipControlsCompensation](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L248) — I: CreateTeam POST201 group identity then prerequisite GET403 and owned DELETE204 compensation. MigrateGroup prerequisite GET403 retains supplied group. No successful team provisioning/readback.

### T51

[TestUnitCreationValidationPrecedesCredentialsAndWrites](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/validation_test.go#L18) — L: create invalid name/alias/description/classification/template/visibility/timing/options and application missing owner; no HTTP. Function body targets CreateTeam despite MigrateGroup ledger reference.

### T52

[TestUnitNativeCRUDContracts](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel/service_test.go#L54) — B: standard/private POST, GET, PATCH and DELETE vectors become first403, including204 writes; no native CRUD success or future enum read.

### T53

[TestUnitSharedCreationAndOperation](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel/service_test.go#L324) — B: shared POST202 becomes403; operation acceptance/poll assertions and later polling request do not execute.

### T54

[TestUnitOwnerResolutionComposition](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel/service_test.go#L364) — B: owner UPN GET becomes403; no resolved-owner binding POST or created channel result.

### T55

[TestUnitIntegratedCoreSharedChannelContentLocationTrustBoundary](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L355) — I negative: accepted POST202 with foreign Content-Location or malformed primary Location returns ErrOperation with no usable handle/channel and no poll.

### T56

[TestUnitIntegratedCoreSharedChannelCancellationRetainsOperation](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L381) — I: accepted POST202 yields handle; locally cancelled Wait returns context.Canceled while retaining accepted ID/state and sends no poll/repeated create.

### T57

[TestUnitIntegratedCoreInitialSharesValidateBeforeHTTP](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L403) — L: invalid initial shared team ID rejected before credentials/HTTP through direct local harness.

### T58

[TestUnitIntegratedCoreInitialShareShapesRemainServerValidated](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/tests/contract/core/integration_test.go#L430) — I denial: explicit-null sharing and standard-with-share input serialized unchanged, POST403 propagated; no acceptance claim.

### T59

[TestUnitTargetingFailuresDoNotInventSuccessOrReplayWrites](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_policy/service_test.go#L133) — D: discovery200 then unexpected202/null200/missing-identity200 read errors and DELETE429/503 single attempts; actual decode/status/APIError negative branches execute.

### T60

[TestUnitServiceDeleteByID](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/team_template/service_test.go#L256) — D: discovery200 then DELETE403 for Delete/DeleteViaIdentity; status==200 receipt branch is unreachable because loop contains403 only.

### T61

[TestUnitUnshareFailureIsNotReplayed](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/shared_channel/service_test.go#L87) — B bypass: DELETE403/429/503 with typed APIError, exact path and one declared mutation; no successful unshare.

### T62

[TestUnitNativeCRUDRequestDenials](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team/service_test.go#L146) — F: PUT create-from-group, team GET/PATCH, group DELETE403 with response metadata and exactly one request.

### T63

[TestUnitMutationFailuresAreNotReplayed](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team/service_test.go#L281) — F: native PUT/PATCH/DELETE503 with APIError and exactly one attempt; name-triggered negative mode preserves exchanges.

### T64

[TestUnitLocalValidationAndCancellationPerformNoIO](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team/service_test.go#L234) — F local: cancellation/nil context/invalid identities/bounds/projection, wrong backend and closed client; verifies zero provider/transport calls. Includes delete identity validation.

### T65

[TestUnitUploadPictureUsesGroupPhotoRouteAndOriginalBytes](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team/picture_test.go#L23) — F: picture200 becomes403 and returns early; picture204 genuinely succeeds, picture503 genuinely fails. Original JPEG route/bytes/type, source close and single attempt checked for204/503. Does not exercise Remove-Team despite ledger mapping.

### T66

[TestUnitNameCompositionsUseCompleteUniqueLookup](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/channel/names_test.go#L14) — B plus bypass: name update/delete positive page chains become first403; one-step missing/ambiguous/bounded DeleteByDisplayName retain actual list decoding and lookup errors.

### T67

[TestUnitHierarchyDeletionFailuresAreNotReplayed](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_hierarchy/failure_test.go#L18) — D: v1/v2 DELETE403/429/503 with backend/status/code and no replay.

### T68

[TestUnitRemoveCompositionPreservesLastOwnerAndPartialWrites](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/service_test.go#L136) — B plus bypass: one-step last-owner guard decodes successful owner list and ErrLastOwner; positive/demotion/no-op/partial delete chains become first owners403, including intended later member denial.

### T69

[TestUnitRemoveReadsAllOwnersBeforeLastOwnerGuard](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_user_management/service_test.go#L164) — B: first owners page403 truncates second page and owner-reference DELETE; complete traversal guard success never executes.

### T70

[TestUnitSparseIdentityUpdateAndOpenValues](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_policy/service_test.go#L82) — D plus L: sparse identity PUT403 preserves future mode/explicit empty body and identity path; malformed identities rejected locally, no successful update.

### T71

[TestUnitManagementSparsePropertyAndSettingsInput](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/service_test.go#L172) — D: sparse group PATCH403; asserts no readback and no mutation of caller settings, but team-settings PATCH and final paired reads never execute.

### T72

[TestUnitArchiveValidationStopsBeforeAnyReadOrWrite](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team_management/validation_test.go#L74) — L: archive option/mode/wait/propagation validation before any read/write; no accepted archive/wait success.

### T73

[TestUnitUploadPictureValidationClosesSourcesWithoutIO](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team/picture_test.go#L41) — F local: invalid ID/cancelled request/close failure/typed nil/nil service preserve close/error/redaction and zero token calls.

### T74

[TestUnitUploadPictureEnforcesGraphSizeLimit](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/graph_api/team/picture_test.go#L70) — F: name has none of the fixture preservation words, so configured200 becomes403; the Invoke still expects ErrUploadTooLarge, source closure and one transport attempt. The transport reads the streaming body before returning its denial. This is an authored size/error-path assertion; source inspection alone does not establish that it passes or a successful upload acknowledgement.

### T75

[TestUnitNativeImportsStreamOriginalCSVWithoutPreparation](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_hierarchy/service_test.go#L143) — D: v1/v2 multipart native import POST403 checks original CSV/file metadata and exactly-once source close; RequestID callback has no result.

### T76

[TestUnitUploadCompositionDenialsRetainPreparationState](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_hierarchy/service_test.go#L171) — D: v1 import403, v2 read403, v2 read404 then create403; actual UploadResult preparation/attempt/acceptance flags and source ownership asserted; import-after-prepare success unexercised.

### T77

[TestUnitHierarchyOptionsProfilesAndSourceValidationAreLocal](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/targeting_hierarchy/failure_test.go#L35) — L: bad version, unsupported profiles and filename/budget/nil source validation with source close counts and no preparation/import attempts.

### T78

[TestUnitServiceUpdateByID](https://github.com/deploymenttheory/go-sdk-msteams/blob/22f01edc43d70cdbeffdf5ea850d38b2a2337952/msteams/teams_admin_api/team_template/service_test.go#L226) — D plus L inputs: both body/expanded definition paths and identity carriers for four update sets; discovery200 then PUT403 only, no result receipt.

