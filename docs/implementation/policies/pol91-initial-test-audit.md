# POL91 configuration qualification checkpoint

Scope: [POL-91 #185](https://github.com/deploymenttheory/go-sdk-msteams/issues/185), 20 assigned resources, 39 commands and 59 exact parameter sets on main `2028f86758df816d89bc6bce136f1c61f461fbb3`. All referenced native symbol files exist; that inventory is not full implementation, contract or live qualification. The issue's historical recovery table is not a claim that no later read succeeded.

## First source-confirmed correction

`Get-CsOnlineVoicemailValidationConfiguration` resolves directly to Core's compiled `GetOnlineVoicemailValidationConfigurationCmdlet`, closed over `BaseTpmGetCmdlet<OnlineVoicemailValidationConfiguration>`. Effective public Filter metadata contains ParameterAttribute for the optional Filter set and inherited ValidateNotNullOrEmptyAttribute. The concrete command does not override Filter. The immutable runtime snapshot's empty validation array omitted this inherited constraint; it is not evidence that explicit empty binding is valid.

PolicyRp `BaseTpmGetCmdlet<T>.FilterPolicies` reads Filter at IL2, calls String.IsNullOrEmpty at IL7 and returns the original collection at IL17. That branch handles omission after binding. The native ListFiltered method previously reached it with an explicitly empty argument; the resource wrapper now returns its existing ErrInput before credentials/HTTP. List still represents omission, and literal whitespace/nonempty wildcard values are preserved. The shared helper and other families are unchanged.

The regression distinguishes three operations: explicit empty has zero requests/credentials; omitted List replays the permanent list response; a literal-space filter also reaches the successful list response and correctly produces no matching identities. Existing get/list response and filtering tests remain separate successful primitive evidence. A permanent fixture pointer does not make the newly authored compound test a captured tenant scenario.

## Bounded binder sweep

All 20 assigned public Get commands were inspected after importing only the pinned module. Fourteen resolve to compiled cmdlets with effective ValidateNotNullOrEmpty on Filter. Thirteen native wrappers already reject explicit empty; voicemail validation is the one missing guard. Six public functions have conditional proxy dispatch and no local inherited Filter validator; no new validator or unconditional backend guarantee is inferred for them.

The compiled group is voicemail validation, privacy, audio conferencing custom prompts, Teams client, education, first-party meeting templates, guest meeting, meeting, meeting templates, messaging, multi-tenant organization, SIP devices, tenant abuse and upgrade. The conditional proxy group is application meeting, guest calling, guest messaging, meeting broadcast, Teams migration and tenant migration.

## Reproduction and provenance

Pinned MicrosoftTeams 8.1.0 package SHA256 is `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Core.dll SHA256 is `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`; PolicyRp.dll is `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`.

On a verified extracted package, import MicrosoftTeams.psd1 and inspect `(Get-Command Get-CsOnlineVoicemailValidationConfiguration).Parameters['Filter'].Attributes`; enumerate ImplementingType.GetProperties with DeclaredOnly separately to verify no Filter override. Repeat Get-Command metadata inspection for all 20 names selected by the issue's ledger resources. The [portable hash-checked IL decoder](https://github.com/deploymenttheory/go-sdk-msteams/blob/de4218ca88f001a65117a206176886b52541e307/docs/implementation/policies/pol04-primary-source-audit.md#portable-reproduction) reads GetMethodBody and resolves operands with owning generic context; use it for the named shared helper. No command ProcessRecord, authentication, tenant operation or additional permission is needed.

## Remaining work

This first checkpoint changes no implementation/contract/live status. Integrated identity/type interactions, successful Set/partial-failure/recovery compositions, every command/set's full binding/model/control behavior and current live oracle/restore evidence still require review. Compiled source controls will be checked independently of retained transport request fixtures. No tenant mutation, extra workflow dispatch, grant, immutable preview regeneration or full-ticket completion claim is added.

## Checkpoint execution

At exact clean revision `8b02a8a85c3f4a05f03fd4c190c7616285ac31b9`, session `70009` exited 0 for the complete online_voicemail_validation_configuration package with `-p 1 -count=1 -run '^(TestUnit|Example)'`, GOMAXPROCS=2, an owned temporary directory and the 14 GiB runtime guard (50.85 GiB free at admission). This includes the existing typed get/list/filter/setter/error tests and the new three-way filter regression. No race, whole-module sweep or tenant test ran locally. The later documentation/metadata checkpoint does not change those compiled inputs; current-head remote CI remains separately required.

## Rejected-vector attribution and first successful write evidence

The exact overwrite pattern appeared in five files across four families: voicemail-validation, privacy, Teams client and education service tests, plus education binding tests. Their constructed HTTP probes use 400. Earlier conditional decode/success expectations were then overwritten by APIError400; they did not establish successful Set or malformed-success propagation. Those suites now explicitly name rejected request probes, assert their controlled400 scope, and remove dead expectations. Local education invalid/no-settings cases remain separate zero-request assertions. Other patterns across the 20 families still require review; this is not a blanket statement that all tests were truncated.

A separately authored privacy primitive test now executes typed PATCH200 with false/null/omission and bodyIdentity distinct from the Global route, followed by a malformed200 model field that retains ErrDecode / raw response / no fabricated value. No retained privacy 200 response was available; the existing permanent privacy fixture is 403, and newer live read success is independent evidence. Primary Core properties/shared PolicyRp contracts justify the authored response shape without treating request-denial probes as an upstream success sample.

Root verification at exact compiled-input revision `42fd88ec81d430ea496e0a90a27ed6473bf933d3`: privacy complete TestUnit/Example passed session `87948`; voicemail-validation, Teams client and education complete TestUnit/Example passed session `29006`. Both used Go 1.27.4, GOMAXPROCS=2, `-p 1 -count=1`, owned temporary directories and the 14 GiB guard. Post-run scoped diffs verified those packages and their shared client/auth/policy/provider/harness dependencies remained unchanged. No race or live test ran. These are four affected package proofs, not complete 39/59 qualification.

## First root configuration composition

`TestUnitConfigurationGlobalIdentitiesKeepPartialFailureAndExplicitRecoverySeparate` uses one native root client across Teams client, privacy and voicemail-validation. Retained client/validation primitive reads and a clearly authored privacy DTO feed sparse prepared Updates: separate false/null selections, distinct typed Global snapshots and 204 receipts remain independent. A later validation 403 retains its own preparation and both prior accepted receipts. Caller-selected independent reads and explicit native restoration PATCH/readback steps exercise recovery without automatic rollback, inferred server persistence or a guessed route. All three concrete Set guards were freshly inspected: AreParametersSpecified tests BoundParameters.ContainsKey, including bound false/null, rather than property HasValue. Root composition execution is recorded separately after its source checkpoint is frozen. No qualification status is promoted by these additions.

### Root execution

The complete tests/contract/policies TestUnit/Example suite passed at exact clean `5855bc586207fb0c95ddb620f2c0d0f29d28ad74`, session `66251`, with GOMAXPROCS=2, `-p 1 -count=1`, owned temporary directory and 14 GiB runtime floor (47.01 GiB admission). Its eleven expected HTTP exchanges establish typed namespace isolation, actual successful prepared Updates, later denied submission and explicit caller-selected restoration/readbacks. These replies prove SDK protocol behavior, not tenant persistence/effect or a successful live recovery run. Four SDK package proofs remain separately bound to `42fd88e` / `87948` and `29006`; no duplicate SDK run was added for this root-only file. All qualification statuses remain unchanged.
