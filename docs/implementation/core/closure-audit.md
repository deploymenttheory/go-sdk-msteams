# CORE-01 closure audit

Audit date: 2026-11-20. This audits [CORE-01 / issue 27](../backlog/issues/CORE-01.md),
the [8.1.0 inventory](../../../contracts/endpoints/8.1.0/core.json), and the evidence
available on the response-contract branch at `f8b939e`. It changes no acceptance,
implementation, coverage, or live-qualification state.

Follow-up in this change: the [generated request/dispatch artifact](../../../contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json)
now captures 24 primitive scenarios, including template CRUD, modern configuration,
assignment and deployment, with cancellation/error/empty-204 cases. Those cases
discharge the primitive capture part of the relevant next actions below; custom
wrapper normalization, command/set links and the remaining wire/composition gates
are still required. The audit's model-only baseline remains identified so earlier
observations are not confused with this subsequent capture.

**CORE-01 is still open. Its next deliverable is a finite set of executable wire
fixtures and accepted command/composition records, not another general survey.**
The inventory reconciles all **40 commands, 52 parameter sets, and 60 referenced
wire primitives**. Two local builders already have settled dispositions. For the
other 38 commands, the modern/static routes and application parameters are known.
The substantial remaining portability question concerns the legacy selection
branch of the three targeting-policy commands. Other remaining questions are
identified below as contract construction, shared implementation prerequisites,
or service qualification; they must not all be represented as unknown endpoints.

## Evidence and limits

The primary implementation source is [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0),
package SHA256 `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Reproducible type/method citations below refer to these package files, rather than
to temporary decompiler line numbers:

| Key | Primary source / SHA256 | Detailed evidence |
| --- | --- | --- |
| T | `netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll`; `3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24` | [Teams/members](team-wire-contract-evidence.md), [channels/sharing/operations](channel-contract-evidence.md), [hierarchy/migration](hierarchy-migration-contract-evidence.md) |
| P | `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll`; `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6` | [Templates/deployment/policies](template-policy-deployment-evidence.md), [policy routing](policy-routing-evidence.md) |
| C | `custom/Merged_custom_PsExt.ps1`; `03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255` | Exported custom wrappers; same family note |
| K | `netcoreapp3.1/Microsoft.TeamsCmdlets.PowerShell.Connect.dll`; `c1c2b2d1c4a5306fb70f373df7ed5bfca4f9db58e90be358afd8a1d48ed46aea` | Exact audience/origin selectors in hierarchy/migration and policy-routing notes |
| R | [Normalized runtime metadata](../../../contracts/powershell/8.1.0/runtime.json) | Exact exported sets, types, mandatory flags and validation; the inventory contains exact JSON references |
| M | [Eight team observations](../../../contracts/endpoints/8.1.0/fixtures/core-team-serialization.json), [30 core observations](../../../contracts/endpoints/8.1.0/fixtures/core-model-serialization.json) | [Response/model interpretation](response-model-evidence.md) |

T names below omit `Microsoft.Teams.PowerShell.TeamsCmdlets`. P generated methods
are on `Microsoft.Teams.ConfigAPI.Cmdlets.Generated.ConfigApiBasedCmdlets`; model
types are in its sibling `Generated.Models` namespace. Full routing source hashes,
including host-contract and command-configuration files, remain in the linked notes.

The 38 existing observations at this audit revision exercise shipped model
serialization/parsing and projection. They do not yet exercise complete generated
request construction, HTTP status dispatch, custom wrapper compositions, or all
52 parameter sets. For example, the deployment serialization case deliberately
contains synthetic Base64 rather than a valid ZIP. Synthetic HTTP responses are
acceptable contract fixtures when their source and purpose are explicit; they
must never be described as observed tenant responses. [M; endpoint evidence README](../../../contracts/endpoints/README.md).

## What actually needs resolving

| Current inventory question | Audit disposition and exact next action |
| --- | --- |
| `fixture-coverage` | A real acceptance gate for every remote command. Capture/build the request/status cases in the command table, associate them with every applicable parameter set and operation ID, and validate them in the offline verifier. Model round trips alone do not discharge this gate. |
| `graph-compatibility` | Mostly an explicit native-contract/composition choice, with source-known pinned differences. The already accepted [CRUD/composition amendment](../plan-amendments.md) and [backend decision](../../decisions/0003-backend-contracts-and-compatibility.md) allow documented native payloads and safe resource primitives. Record the differences and fixture both the pinned characterization and chosen native behavior; do not require reproducing client bugs as the primary Go API. See the choices below. |
| `config-routing`, templates/deployment | The destination/audience/discovery/header contract is source-known. Integrate and fixture same-origin service discovery, host substitution and `X-MS-Target-Uri`. Do not use the generated development deployment host. App-role minima and per-cloud deployment are qualification limits, not missing route semantics. [P `Module.SendAsync`, `SendAsyncCallback`; K/host-contract selectors; policy-routing note](policy-routing-evidence.md). |
| `config-routing`, three targeting policies | Modern HTTP is source-known; universal modern acceptance when the pinned client selects TRPS is unproven. The exact conditional disposition must retain the legacy branch and the native parity gap. Neither a missing flight nor a failed bootstrap proves that the service rejects REST. See the conditional gate below. |
| `deployment-status` | Resolve the assigned read contract as **raw text plus HTTP status/metadata**, which is what P `BatchTeamsDeploymentStatusGet_Call` and the exported handlers consume. Named 400/404/500 response strings remain failures in Go. Preserve unknown text. A structured per-team decoder/waiter needs further evidence, but is not a prerequisite for the evidenced raw read. `Completed` must not mean success. [Pinned handlers and official status help](template-policy-deployment-evidence.md#batch-deployment). |
| `operation-auth` | The team route and Graph audience are established by T `GetOperation.ProcessRecord`/`TeamsCmdlet.PerformAuthorization`. Document unpublished team-route least-privilege details and server-enforced authorization. Fixture received 401/403 failures. This does not prevent implementing the route; it prevents claiming a verified least-privilege grant or universal auth/cloud qualification. Chat grants are unrelated evidence. See below. |
| `additional-profiles` | Known shared implementation prerequisite, not unknown endpoints: define the exact four extra request-origin/audience profiles and scoped delegated-token/cache contract in the hierarchy/migration note. Preserve `Region.Read.All`; existing Graph/ConfigAPI-only provider checks cannot serve these calls unchanged. Fixture each command's selected profile and reject disallowed cross-profile requests. |
| `hierarchy-models` | T `SetTeamTargetingHierarchy.UploadCsvTargetingV2` only consumes **HTTP status** from supporting Get/Create, then a string map with `id` from import. T `Model.TeamHierarchy` proves create input `id`/`displayName`. A complete Get/Create response model is not necessary for that assigned upload composition. Keep those responses opaque/status-bearing until evidenced; do not advertise a complete standalone hierarchy read model. |

These dispositions narrow the questions; they do not turn the in-progress inventory
into accepted contracts. In particular, a documented source-known branch still needs
a fixture and a command mapping before the first acceptance box can be checked.

## Native Graph contract and PowerShell composition choices

The following choices can be encoded now under the accepted architecture. Their
remaining gate is an explicit contract/test mapping, not permission to guess how
the pinned anomalous requests behave on a tenant.

| Pinned observation | Native operation and composition contract to record |
| --- | --- |
| Numeric team `giphyContentRating`; `unknownFutureValue` omitted | Native team settings emit the documented string value; preserve the numeric M case as a baseline-client discrepancy. An unknown response value survives. The composition maps the supported user setting to the native field without claiming byte identity. [T `TeamSettings.ToTeam`; current Graph fun settings](https://learn.microsoft.com/en-us/graph/api/resources/teamfunsettings?view=graph-rest-1.0). |
| Channel types use `Standard`/`Private`; add/demotion uses `roles:["member"]`; discriminator casing differs | Native channel create uses documented type values and AAD discriminator; native membership creation/demotion uses an explicit empty roles array for ordinary membership, and promotion uses `owner`. Preserve omission versus `[]`. The baseline observations remain separate fixtures. [T New/Add/Remove channel methods; beta member update contract](https://learn.microsoft.com/en-us/graph/api/channel-update-members?view=graph-rest-beta). |
| Private membership ID synthesized from channel/user; first-page name resolver; first matching identity | Native membership CRUD accepts opaque IDs obtained from service reads; an explicit resolver follows pages and reports ambiguity. Characterize the pinned algorithm/first-page limit separately. Do not silently expose the synthetic ID as the resource identity. These resource reads are already evidenced in the [channel contract](channel-contract-evidence.md#sdk-crud-supporting-operations-from-current-primary-documentation). |
| Supplied Owner plus omitted type bypasses the explicit-Standard guard | Retain the exact pinned branch in the compatibility record. The native create input makes type/initial owner explicit and uses documented per-type validation. Do not claim the omitted-type-plus-owner request is accepted merely because it serializes. [T `NewTeamChannel.ProcessRecord`; Graph channel create](https://learn.microsoft.com/en-us/graph/api/channel-post?view=graph-rest-beta). |
| Commercial Graph URLs embedded in member/owner bindings | Construct native bindings from the selected, approved Graph profile; fixture profile-specific bindings. The pinned hard-coded URL is a compatibility limitation, not evidence of sovereign success. [T binding construction; shared routing decision](../http/routing-evidence.md). |
| Historical `myorganization/users` fallback on 404 | Preserve direct GUID/UPN resolution and its error distinction. A native email resolver can use the documented `/users` collection with an escaped mail filter and explicit ambiguity handling. Record this as a documented supporting operation rather than silently claiming the legacy alias is qualified. [T `GraphUtilities.GetUserId`; Graph users list](https://learn.microsoft.com/en-us/graph/api/user-list?view=graph-rest-1.0). |
| Add-TeamUser treats any 2xx preflight as existence | Native preflight inspects `value`; empty collection permits the create branch. Preserve HTTP failures; do not turn 403 into absence. Member and owner writes retain partial results. [T `AddTeamUser.AddUserToUnifiedGroup`; group member collection](https://learn.microsoft.com/en-us/graph/api/group-list-members?view=graph-rest-1.0). |
| Archive body is JSON null; unarchive body is JSON empty string; immediate readback | Native Archive/Unarchive use the documented request contract and return an operation. A composition may wait/read back while retaining the operation ID and partial outcome. Keep M's anomalous body literals as characterizations. [T `SetTeamArchivedState`; archive](https://learn.microsoft.com/en-us/graph/api/team-archive?view=graph-rest-1.0), [unarchive](https://learn.microsoft.com/en-us/graph/api/team-unarchive?view=graph-rest-1.0). |
| Group and team fields combined into one command result | Retain distinct typed resource primitives plus an explicit enriched projection/composition. List and Get remain different contracts; a partial batch cannot become a complete successful list. [T `GetTeam`, `TeamsCmdlet.GetTeamWithGroupAttributes`; team-wire note](team-wire-contract-evidence.md). |
| Two `$search` keys when both team filters supplied | Record the pinned generated URI; do not invent its server interpretation. Give the native list one documented search expression or compose independently evidenced searches with an explicit intersection, then fixture the chosen semantics and hydration failures. Graph documents one `$search` parameter, tokenization for displayName/description and startsWith behavior for other fields. [T `GetTeam.ProcessRecord`; current beta group query contract](https://learn.microsoft.com/en-us/graph/api/group-list?view=graph-rest-beta). |

The last choice must be made and tested in the contract implementation; it cannot
remain an unlabelled `graph-compatibility` placeholder. None of these decisions
licenses changing the pinned beta route to v1.0 without separately recording the
supporting operation's provenance. Exact reproduction of a pinned bug, if offered
as a compatibility mode, still needs an explicit limited contract.

### Get-Operation authorization and failure envelope

T `GetOperation.ProcessRecord` proves GET `/beta/teams/{teamID}/operations/{operationID}`
with default Graph authorization and no body. Current [team-create](https://learn.microsoft.com/en-us/graph/api/team-post?view=graph-rest-beta)
and [archive](https://learn.microsoft.com/en-us/graph/api/team-archive?view=graph-rest-1.0)
documentation direct the caller to poll the returned operation Location. That
supports continuing the originating workflow within its Graph identity/audience
boundary; **it does not publish a general rule that the initiating permission,
same initiating principal, or any particular read grant always authorizes every
team-operation read**. No such guarantee was found in the inspected sources.

The current [Get teamsAsyncOperation reference](https://learn.microsoft.com/en-us/graph/api/teamsasyncoperation-get?view=graph-rest-beta)
describes `/chats/...` and explicitly labels its permission table for chats.
Do not borrow those grants. Record team-operation least privilege as unpublished,
request authorization as server-enforced, and lab combinations as unobserved.
Keep the poll bound to the caller's configured identity; there is no reason for
the SDK to acquire a new, guessed permission to poll.

The [operation resource](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-beta)
establishes error diagnostics, open-ended operation/status strings and more than
30 seconds between polls. M demonstrates that the pinned class drops `error`
and rejects unknown string states; native preservation of them is already required
by decision 0003. Add a failed-status object with `error.code/message`, an unknown
status, and received 401/403/404/429 fixtures. These are document-derived/synthetic
contracts, not proofs of an application's actual permission grants. This resolves
the response-model design gate without requiring a tenant failure capture.

### Targeting policy conditional disposition

For each of Get/Set/Remove-CsTeamsTargetingPolicy, the accepted record must state:

1. The modern ConfigAPI route, payload, response/error dispatch and common discovery
   prerequisites are evidenced and implementable natively.
2. P `FlightingUtils`/`TrpsPSCmdlet` use server-provided client selection or a local
   override to choose modern execution; bootstrap failure and absent selection
   are distinct. The flight list is not proven to be a service authorization gate.
3. The other pinned branch executes the original command over a TRPS runspace.
   Its native protocol is not established by the extracted modern JSON route.
   An absent flight cannot be turned into a new unsupported-tenant allowlist.
4. Full native parity for that branch remains blocked by an evidenced native
   legacy contract or first-party service evidence settling modern coverage of
   the remaining cases. A modern fixture or one successful tenant does not prove
   universal coverage. [Primary call paths](policy-routing-evidence.md#native-implementation-disposition-and-closure).

CORE-01 explicitly allows an **evidenced conditional disposition**. Such a record
can discharge that research criterion for these three commands while retaining
the above implementation/release blocker; it cannot count as delivery, remove the
legacy cases from scope, or establish universal modern support. Merely changing
the inventory's status string is insufficient. The condition and exact outstanding
work must be linked from the command and downstream ticket/coverage records.

## Per-command remaining gate and next action

Every remote row also requires **F0**: a versioned record for the primitive and
composition, exact runtime-set mapping, required/optional validation and
presence/default semantics, selected profile, status/error handling, paging/async/
transfer behavior or explicit absence, provenance, and positive/validation/auth/
failure fixtures. This can reference shared adapter fixtures instead of copying
identical errors 38 times, but each command/set must identify the cases it uses.
The fixture names below describe required cases, not existing files or passed tests.

**W** means source-known wire behavior awaiting F0/fixtures; **P** additionally needs
the known extra profile/token design prerequisite; **C** needs the evidenced policy
conditional disposition above; **L** is already a settled local disposition.
The table preserves the exact CORE-01 assignment order. Set names/counts come from R.

| Command; parameter sets | Gate | Concrete remaining fixture/evidence | Exact next action / primary source |
| --- | --- | --- | --- |
| `Add-TeamChannelUser`; 1 | W | Add vs promote; Standard rejection; private/shared membership resolution; TenantId present/empty/ignored on promotion; empty roles vs pinned member role; received failures | Record Create/UpdateByID plus name/user resolution composition; fixture all branches against T `AddTeamChannelUser.ProcessRecord`, `UserUtilities.GetMembershipIdFromUserIdOrUPN` and documented native payload. |
| `Get-TeamChannelUser`; 1 | W | Filtered name lookup no-match; member paging; role precedence/order; enrichment groups of ten and failed enrichment; opaque ID retention | Fixture List plus explicit lookup/projection/enrichment composition from T `GetTeamChannelUser`, `UserUtilities`, `GraphUtilities.GetUpnForUsers`; link user-read permission dependency. |
| `Remove-TeamChannelUser`; 1 | W | Delete vs owner demotion; private/shared ID paths; Standard rejection; explicit empty roles; local vs HTTP failures | Fixture DeleteByID/UpdateByID and resolver composition from T `RemoveTeamChannelUser`; retain pinned demotion spelling as separate characterization. |
| `Get-TeamAllChannel`; 1 | W | All-pages list, optional membership filter; host-team projection from OData ID; absent/malformed OData ID | Fixture ListAll separately from hosted/incoming lists, with safe host derivation, from T `GetTeamAllChannel`, `ChannelEnhancedInfo`. |
| `Get-TeamChannel`; 1 | W | Hosted list with omitted/Standard/Private/Shared filter; continuation; missing/null collection; optional/unknown channel fields | Fixture List and page model from T `GetTeamChannel`/`HttpUtilities.GetAll`; do not claim an empty collection from malformed `value`. |
| `Get-TeamIncomingChannel`; 1 | W | Incoming path/paging; enhanced output and host identity; continuation failure | Fixture ListIncoming from T `GetTeamIncomingChannel`; keep recipient team and host team distinct. |
| `New-TeamChannel`; 1 | W | Default/explicit type; null/empty description; owner rules; 201 Channel vs 202 Location and missing Location; failed/unknown operation | Fixture Create union result and documented native inputs from T `NewTeamChannel`; preserve M's baseline casing; return accepted operation separately from channel. |
| `Remove-TeamChannel`; 1 | W | Name no-match/first-page limitation vs chosen full resolver; ID delete 204 and failure | Fixture DeleteByID and explicit resolver composition from T `RemoveTeamChannel`; ID-based delete itself needs no name lookup. |
| `Set-TeamChannel`; 1 | W | Omitted vs empty patch fields; name no-match; patch 204; failed resolver/patch | Use existing M patch cases, add HTTP UpdateByID and resolver fixtures from T `SetTeamChannel`; document native full-page resolver difference. |
| `Get-CsBatchTeamsDeploymentStatus`; `Get`, `GetViaIdentity` | W | Escaped orchestration ID; identity property missing; raw 200/400/404/500 strings and undeclared status; opaque text | Capture P `BatchTeamsDeploymentStatusGet` request/dispatch and wrapper identity extraction; accept raw status contract now, leave any structured decoder explicitly unevidenced. |
| `Get-MultiGeoRegion`; `Identity` | P | User vs Group route; exact `Region.Read.All`; optional timestamp/location; profile rejection and HTTP error | Encode Regions.GetForUser/GetForGroup with K/T profile tuple; fixture T `ExecuteMultiGeoRegionForEntity` and nullable result projection; retain documented application-auth exclusion. |
| `Get-TenantPrivateChannelMigrationStatus`; 1 | P | Each evidenced cloud prefix/audience; nullable timestamps; Details retained as string; RequiresAdminAttention/unknown status; HTTP failures | Define singleton Get profile then fixture T `GetTemplatesServiceApiBaseUrl`/ProcessRecord and migration model; do not force Details into a guessed object. |
| `New-CsBatchTeamsDeployment`; 1 | W | Valid ZIP with exactly Teams.csv/Users.csv; Base64/field casing; 202 both IDs; named/default error; no replay | Build synthetic CSV bytes/isolated archive fixture, then capture P submission request/dispatch plus C wrapper normalization; preserve instanceId and executionId. |
| `Get-Operation`; 1 | W | Beta team path; failed `error`; unknown state; 401/403; wait interval/cancellation through shared adapter | Add TeamOperations.GetByID and source-derived native response fixtures as above; document unpublished least-privilege table without borrowing chat grants. T `GetOperation`, M, Graph operation resource. |
| `Get-AssociatedTeam`; 1 | W | Omitted/empty User→me vs GUID/UPN explicit user; paging; application me rejection; lookup failure | Fixture both canonical associated-team lists and explicit user resolver from T `GetAssociatedTeam`; retain current documented delegated-me/application-explicit-user restriction. |
| `Get-SharedWithTeam`; 1 | W | Omitted/empty SharedWithTeamId list vs specified get; collection projection; paging/error | Fixture ChannelShares.List/GetByID from T `GetSharedWithTeam`; no Create/Update route is needed for this assigned command. |
| `Get-SharedWithTeamUser`; 1 | W | allowedMembers path; Owner/Member filtering/order; enrichment failure/cross-tenant absent user; pagination | Fixture ListAllowedMembers plus projection/enrichment from T `GetSharedWithTeamUser`/`UserUtilities`; retain documented roster exclusions as restrictions. |
| `Remove-SharedWithTeam`; 1 | W | Correct sharing identity; no body; 204; 403/404 structured failure | Fixture ChannelShares.DeleteByID from T `RemoveSharedWithTeam`; do not substitute deletion of the channel or target team. |
| `Get-CsTeamsTargetingPolicy`; `Filter`, `Identity` | C | Modern single/list routes; local wildcard and Tag: matching; Global first; unknown properties; discovery/flight branches | Capture P configuration GET dispatch and C local filtering/sorting, then write per-command conditional disposition retaining legacy branch. |
| `Get-TeamTargetingHierarchyStatus`; 1 | P | Default/1 vs 2; omitted/blank latest vs supplied ID; both status/error models; nested JSON error string; unknown state | Fixture T ProcessRecord and both Result constructors; reuse M v2 projection cases, add v1 errors/routes and exact profile selection. |
| `Remove-CsTeamsTargetingPolicy`; `Identity` | C | Escaped required identity; DELETE/no body/204; error envelope; WhatIf mapped outside service body | Capture P configuration DELETE and C wrapper input, then write conditional record; describe Global reset as documented server behavior without guessed reset values. |
| `Remove-TeamTargetingHierarchy`; 1 | P | Default/1 vs 2 DELETE; exact profile; no request body; success/error; no fabricated ClientRequestId input | Fixture T `RemoveTeamTargetingHierarchy.ProcessRecord`; treat deletion as singleton lifecycle with no automatic replay. |
| `Set-CsTeamsTargetingPolicy`; `Identity` | C | Omitted identity→Global; bound-only PUT body; null omitted/empty retained; 204; errors; unknown values | Reuse M sparse/unknown-property cases, add P/C request/dispatch and discovery fixtures; write conditional record, retain strings rather than an incomplete closed enum. |
| `Set-TeamTargetingHierarchy`; 1 | P | Both multipart routes/name/basename/application-json part; v2 200→import, 404→create→import, other read failure stops; partial create/import failure; missing response id | Encode status-only supporting Get/Create contract and known TeamHierarchy create body; fixture T upload methods with synthetic CSV; add profile/scope design before downstream implementation. |
| `Add-TeamUser`; 1 | W | GUID/UPN resolution; Member vs Owner; empty vs populated preflight; membership succeeds/owner fails; reference bindings | Fixture typed membership preflight and both reference writes from T `AddTeamUser`; document correction of pinned 2xx-existence bug and partial results. |
| `Get-TeamUser`; 1 | W | Owners-only/members-only/guest path; both lists paged; owner de-duplication; user-type classification; failed second list | Fixture T `GetTeamUser.ProcessRecord`/Transform with raw group-user wire models and separate flattened projection. |
| `Remove-TeamUser`; 1 | W | Last-owner local 412; Owner demotion vs full removal; owner deletion succeeds/member deletion fails; required `$ref` | Fixture T `RemoveTeamUser`; record explicit DemoteOwner vs DeleteByID composition and group/user identity. |
| `Get-CsTeamTemplate`; `Get`, `GetViaIdentity` | W | Whole OData-ID segment escaping; missing identity property; 200 detail flattening; named/default errors, empty/malformed bodies | Reuse M detail/flattening/error cases; capture P GET request/dispatch and wrapper identity extraction. |
| `Get-CsTeamTemplateList`; `DefaultLocaleOverride` | W | Omitted locale en-US; explicit whitespace/null wrapper quirk; list summaries/zero counts; partial-list diagnostic; HTTP error | Capture P GET-list route and C locale behavior; fixture diagnostic with returned partial values; expose incompleteness metadata. |
| `New-CsTeamsHiddenMeetingTemplate`; `Identity` | L | No route/auth unknown. Downstream local constructor needs nonempty ID validation and output mapping | Keep settled-local-builder. Add pure value-constructor cases when implementing; existing base-chain proof is sufficient research evidence. [Hidden builder sources](template-policy-deployment-evidence.md#hidden-template-builders-settled-local-disposition). |
| `New-CsTeamsHiddenTemplate`; `Identity` | L | Same local-only disposition; do not invent policy persistence or a management endpoint | Keep settled-local-builder and link eventual pure constructor symbol/tests; same primary base-chain/model evidence. |
| `New-CsTeamTemplate`; `New`, `NewExpanded`, `NewViaIdentity`, `NewViaIdentityExpanded` | W | Four input mappings; locale from identity; create-only General insertion; Body vs expanded defaults; arrays/false; 201/error dispatch | Reuse M nested/default cases; fixture C General normalization on a copy and P POST request/dispatch across all four sets. |
| `Remove-CsTeamTemplate`; `Delete`, `DeleteViaIdentity` | W | Escaped ID and identity extraction; empty 204 actual parser/dispatch; 200 Any; 400/default error | Exercise actual P `TeamTemplateDelete_Call`/Any parser with synthetic empty 204; its callback awaits parsing. Model error empty-text observation alone does not prove this path. Native delete returns metadata on success. |
| `Update-CsTeamTemplate`; `Update`, `UpdateExpanded`, `UpdateViaIdentity`, `UpdateViaIdentityExpanded` | W | Four input mappings; PUT; Body/expanded defaults; 200/201/error; no General insertion | Reuse M field/default cases; capture P PUT request/dispatch and each exported identity/body mapping; verify create normalization is absent. |
| `Get-Team`; `Filters`, `Identity` | W | Identity ignores filters; tenant vs user list; both search fields decision; batch width 1..20/default20; pagination/hydration per-item 429/failure; false Archived | Fixture T GetTeam/group-team merge/batch helpers; choose and record native dual-search behavior, retaining pinned URI characterization and partial identities. |
| `New-Team`; `CreateTeam`, `MigrateGroup` | W | Separate sets; explicit owner vs me; alias generation; ordinary/EDU defaults; bounded 404 stages; owned cleanup/retain/ambiguous outcome; partial ID | Reuse M ordinary/EDU/settings cases; fixture group→team composition plus existing-group branch, safe compensation result, and app-mode restriction/conflicting documentation. |
| `Remove-Team`; 1 | W | Backing-group DELETE and lifecycle effect; 204; 403/404; no body | Fixture T `RemoveTeam.ProcessRecord`; document deletion of backing group and team, no fake team-only DELETE. |
| `Set-Team`; 1 | W | Group PATCH then team PATCH then enriched readback; sparse values vs empty containers; string Giphy enum; each partial failure | Reuse M patch/settings cases; fixture T SetTeam sequence and native typed settings; surface group success/team failure without claiming atomicity. |
| `Set-TeamArchivedState`; 1 | W | Archive/unarchive actions; SharePoint flag validation/app restriction; already-state local result; 202 operation; wait failure/cancel | Reuse M pinned literals, add documented native body and operation fixtures from T/helper plus Graph docs; retain operation ID before readback. |
| `Set-TeamPicture`; 1 | W | Original bytes/content type; bounded stream/file failure; 200/204 success; received failure; correct group-photo grant | Fixture PUT `/groups/{id}/photo/$value` from T `SetTeamPicture`; document stream ownership and current 4 MB limit; use group-photo permission evidence below. |

## Authorization, qualification, and supporting CRUD limits

Complete the supporting Graph authorization references while writing the accepted
contracts; this is documentation integration, not an indefinite least-privilege
lab prerequisite. The endpoint's actual resource matters. Current primary docs
specify `GroupMember.ReadWrite.All` for adding user members and `Group.ReadWrite.All`
for adding owners, with delegated role/ownership restrictions. These are distinct
from channel or conversation-member permissions. [Add group members](https://learn.microsoft.com/en-us/graph/api/group-post-members?view=graph-rest-1.0),
[add owners](https://learn.microsoft.com/en-us/graph/api/group-post-owners?view=graph-rest-1.0).
Do not assert that those grants alone authorize a composed command's user lookups,
preflight lists, group/team updates, or cleanup.

For Set-TeamPicture's actual **group** photo route, the current least privilege is
`ProfilePhoto.ReadWrite.All` in delegated/application mode; `Group.ReadWrite.All`
is a broader listed alternative. The separate team-photo table uses a different
grant. The same current reference documents JPEG bytes and a 4 MB maximum.
[Update profilePhoto](https://learn.microsoft.com/en-us/graph/api/profilephoto-update?view=graph-rest-1.0).
This resolves a documentary route/grant association; it does not qualify a tenant
or prove that pinned PNG/GIF examples undergo conversion.

Record unobserved live authorization, exact service validation codes, propagation
latencies, throttling, licences and per-cloud deployment explicitly. Where no exact
role/grant is published, retain that limit and typed received authorization errors;
do not invent a role or infer eligibility from token acquisition. The ticket's
verification text expressly permits package/documentation contracts when the lab
is unavailable. These observations therefore do not block a correctly bounded
offline wire contract merely because no live 403/404/429 has been captured.
Documented exclusions and contradictions, including New-Team's app-auth conflict
and Get-MultiGeoRegion's app-auth exclusion, remain restrictions, not positive
qualification. [Teams documentation evidence](teams-documentation-evidence.md),
[hierarchy/migration profiles](hierarchy-migration-contract-evidence.md).

The extra resource matrix must remain honest. Sharing Create/Update have published
Graph PowerShell candidates but no settled service mutation contract in this
evidence; they remain research candidates for full family CRUD, not invented
unsupported operations and not prerequisites for the four assigned sharing
read/delete commands. Hierarchy standalone Get/Create response models likewise
remain incomplete while the assigned upload can use its evidenced status-only
primitives. Deployment, region/status resources and local builders do not acquire
fictional list/update/delete operations to fill a matrix. [Assessed sharing gap](channel-contract-evidence.md#sharing-creation-and-update-assessed-not-declared-unavailable),
[hierarchy/action matrix](hierarchy-migration-contract-evidence.md#hierarchy-crudaction-matrix),
[accepted CRUD rule](../plan-amendments.md).

## Concrete completion sequence

1. Add the generated ConfigAPI request/status-dispatch fixtures first: templates,
   modern configuration and deployment. Use injected in-memory sender/listener
   implementations; do not invoke module authentication/pipeline/management.
   Exercise empty template delete 204 through its actual parser/handler. Keep
   generated request construction separate from production host substitution.
2. Add the Graph and hierarchy/migration request/composition fixtures enumerated
   above, reusing the 38 model observations. Verify the 52 set mappings and all
   application fields, not just one happy case per command. Use controlled
   handlers and source/document-derived responses; no tenant credentials required.
3. Extend the versioned inventory into accepted primitive/composition records:
   fields/presence, status/error rules, profile/replay/paging/operation behavior,
   compatibility choices and fixture references. Add explicit conditional policy
   dispositions and the known extra-profile/token prerequisite. Update the
   verifier's deliberate in-progress-only status rules as part of that reviewed
   schema change; passing its current count check cannot accept the contracts.
4. Review the closure against the two CORE-01 criteria. Close the research ticket
   only when every row/set is linked to that concrete contract or evidenced local/
   conditional disposition, limitations and downstream gates are preserved, and
   representative positive/validation/auth/failure fixture mappings exist. Keep
   legacy native policy parity and unsettled supporting mutation contracts visible
   in downstream work; no full-release coverage claim follows from research closure.
5. Start the affected SDK tickets from those resource contracts and compositions.
   Additional profile/token implementation precedes the five P rows; modern
   ConfigAPI discovery implementation precedes its remote consumers. Ledger states
   advance to contract-tested only when SDK symbols and their controlled-server
   tests exist, and to live-tested only from actual permitted observations.

Audit verification: `python3 tools/contracts/verify_models.py` passed the existing
inventory check (40 commands, 60 primitives, status `in-progress`). The audit
independently counted 52 exact runtime-set references. No capture was rerun, no
management command or live call was executed, and no credentials were used.
