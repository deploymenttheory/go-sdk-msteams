# POL-05 pinned primary-source audit

Reviewed 2026-11-26 against repository baseline `bb56e7b` and the retained MicrosoftTeams **8.1.0** module. Scope is exactly the 36 commands / 61 runtime parameter sets assigned `research_owner: POL-05` in [coverage](../../../contracts/coverage/8.1.0.json), across 11 families. This is an offline source audit: no authentication, tenant requests, token probes, Go execution, or new response capture occurred. Issue #36 remains open for the qualification gaps below.

## Reproducible evidence

The module was imported from `/private/tmp/msteams-core90-pinned/module/MicrosoftTeams.psd1`; every assigned name was resolved with `Get-Command`, independently of candidate mappings. All 23 resolved Core `ProcessRecord` methods were decoded: each calls its matching `BaseTpmGet/New/Set/Remove/GrantCmdlet<T>` provider at decimal **IL1**. Binary citations below use declaring type, method and decimal IL byte offset. Script citations use pinned-file line numbers.

The six shared SHA-256 hashes were recomputed and match the [published POL-04 audit and portable IL reader at commit dc2bb9a](https://github.com/deploymenttheory/go-sdk-msteams/blob/dc2bb9a/docs/implementation/policies/pol04-primary-source-audit.md#portable-reproduction):

| Module-relative file | SHA-256 |
|---|---|
| `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll` | `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49` |
| `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll` | `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51` |
| `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll` | `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6` |
| `custom/Merged_custom_PsExt.ps1` | `03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255` |
| `custom/CmdletConfig.json` | `754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906` |
| `exports/ProxyCmdletDefinitionsWithHelp.ps1` | `5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5` |
| `internal/Merged_internal.ps1` | `6ff95158174bab031dc1fc24112a2bc6ae81cf1e2f59db4ecb58afbefc8c2bd9` |

Obtain the exact Gallery package using [baseline acquisition/provenance](../../../tools/baseline/README.md), check these hashes, then use that published reader with a caller-selected `ModuleRoot`. Generic methods must resolve metadata tokens with the owning declaring type and method generic arguments; the published reader does so. No host-specific helper is required to reproduce this report. `Core`, `PolicyRp`, `Private`, `Helpers`, `Mapping` and `Proxies` below refer to the six shared files; `Internal` names the additional internal-wrapper file hashed above.

To reproduce public resolution without executing a tenant command:

```powershell
Import-Module (Join-Path $ModuleRoot 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$records = (Get-Content ./contracts/coverage/8.1.0.json -Raw | ConvertFrom-Json).records
$records | Where-Object research_owner -eq 'POL-05' | ForEach-Object {
    $command = Get-Command $_.name
    [pscustomobject]@{
        Name = $command.Name
        Kind = $command.CommandType
        Type = $command.ImplementingType.FullName
        Sets = $command.ParameterSets.Name -join ','
    }
}
```

## Public surface, PowerShell selection, SDK branch, qualification

The result is **23 compiled public cmdlets and 13 public functions**. `C` below means the actual resolved Core cmdlet and compiled PolicyRp provider; `P` means an exported function forwarding to a private `TrpsPSCmdlet`, whose execution is conditional modern ConfigAPI **or remote TRPS**. This is not a classification from `CmdletConfig.json` alone. Verb names are exactly the assigned public surface, not every internal model/class in the package.

| Family | Actual public verbs / selection | SDK at `bb56e7b` and pending implementation | Live qualification |
|---|---|---|---|
| MediaConnectivityPolicy | Get/New/Set/Remove/Grant: C | Absent on baseline; [PR #333](https://github.com/deploymenttheory/go-sdk-msteams/pull/333), head `9aabde3`, compiled implementation | Pending implementation/source parity does not establish tenant behavior. |
| MediaLoggingPolicy | Get/Grant: P | Modern Get/Grant | [Existing record](media-logging-lab-acceptance.md) only within its recorded scenarios. No inferred CRUD. |
| MeetingBroadcastConfiguration | Get/Set: P | Modern Get/Set | [Existing record](meeting-broadcast-configuration-lab-acceptance.md); ExposeSDN and token handling retain separate boundaries. |
| MeetingBroadcastPolicy | Get/New/Set/Remove/Grant: P | Modern CRUD/Grant | [Existing record](meeting-broadcast-policy-lab-acceptance.md); legacy and assignment effects remain separate. |
| MeetingConfiguration | Get/Set: C | Absent on baseline; [PR #332](https://github.com/deploymenttheory/go-sdk-msteams/pull/332), head `d1a4fbe`, compiled implementation with the source validation correction below | No promotion of pending implementation into baseline or tenant qualification. |
| MeetingPolicy | Get/New/Set/Remove: C; Grant: P | Compiled CRUD; modern Grant | [Existing record](meeting-policy-lab-acceptance.md); individual settings/auth/clouds/assignment effects remain bounded. |
| MeetingTemplateConfiguration | Get: C | Compiled read only | [Existing template record](meeting-template-policies-lab-acceptance.md); no public Set contract inferred. |
| MeetingTemplatePermissionPolicy | Get/New/Set/Remove/Grant: C | Compiled CRUD/Grant | Same template record; nested operations and effective assignment need their own evidence. |
| MessagingConfiguration | Get/Set: C | Compiled Get/Set | [Existing record](messaging-migration-configuration-lab-acceptance.md); not every list/null/enum combination. |
| MessagingPolicy | Get/New/Set/Remove: C; Grant: P | Compiled CRUD; modern Grant | [Existing record](messaging-policy-lab-acceptance.md); individual settings and assignments remain distinct. |
| MigrationConfiguration | Get/Set: P | Modern Get/Set | [Existing record](messaging-migration-configuration-lab-acceptance.md); legacy remains unqualified. |

The separate [contract inventory](pol05-contract-inventory.md) owns full command/set and implementation-symbol mapping. This audit found no SDK backend mismatch in the reviewed baseline/pending branch selections. In particular, compiled MeetingPolicy/MessagingPolicy CRUD does not make their Grants compiled. Public wildcard resolution returned only `Get-CsTeamsMediaLoggingPolicy`, `Grant-CsTeamsMediaLoggingPolicy`, and `Get-CsTeamsMeetingTemplateConfiguration` for those two families, despite other internal classes. No missing public verbs were invented.

For every public Get, the resolved parameter sets are `Identity,Filter`; New/Set/Remove have `Identity`. Compiled Grants have `Identity,GrantToUser,GrantToGroup,GrantToTenant`; proxy Grants have `Identity,GrantToTenant,GrantToGroup` (MeetingPolicy reports the same three in another order). Their totals are the 61 assigned sets. The complete binding types remain owned by [runtime metadata](../../../contracts/powershell/8.1.0/runtime.json), not by route resemblance.

## Conditional modern and concrete remote boundary

The 13 P commands are mapped to `Get/New/Set/Remove-CsConfigurationModern` and `Grant-CsTeamsPolicy` with the corresponding `ConfigType`/`PolicyType`: Mapping 1697–1751, 2290–2373, 3105–3119, 3720–3734, 4441–4496. Public proxies select private same-named cmdlets, not their helper directly: e.g. MediaLogging Get Proxies 9215–9278; Grant 15854–15938; BroadcastConfiguration Get 9279–9347 and Set 39639–39731; MeetingPolicy Grant 16024–16108; MessagingPolicy Grant 16109–16193. Each proxy builds a steppable pipeline, then calls `Begin`, `Process($_)` and `End`. Pipeline orchestration is an observable part of these public functions, not evidence of batch REST semantics.

The exact shared dispatch is established in Private and is reproduced in the [published source audit](https://github.com/deploymenttheory/go-sdk-msteams/blob/dc2bb9a/docs/implementation/policies/pol04-primary-source-audit.md#conditional-dispatch-and-the-precise-legacy-boundary):

- `TrpsPSCmdlet.BeginProcessing` IL55/60 checks hidden `MsftInternalProcessingMode`; IL91/99/107 supplies `Legacy`, `TryModern`, `ForceModern`; IL155/160 removes the control before forwarding. `TryGetFlightedApi` IL46/64/91 compares modes and IL128 calls `FlightingUtils.GetFlightingCommandInfo`. The latter consults `SessionConfiguration.RemotingCmdletsFlightedForAutoRest` IL59 and modern availability IL98. Omitted hidden control still leaves session-dependent flight selection.
- `TrpsPSCmdlet.getProcessor` IL52/72 constructs modern processors; IL82 calls `CreateTrpsPipelineProcessor` for the remote path. The latter takes public name IL21–31, `TrpsSessionPsCmds.GetSession().Runspace` IL36/41 and cleaned parameters IL53, constructing `RemoteSyncProcessor` IL58 or `RemoteCommandPipelineProcessor` IL139.
- `RemoteSyncProcessor.ReceiveOutput` adds the public command IL16, attaches the remote runspace IL42, invokes `PowerShell.Invoke()` IL49. Pipeline processing calls `BeginInvoke` in `Begin` IL47 and `EndInvoke` in `WaitForAndCollectRemainingOutput` IL25.
- `TrpsSessionBuilder.CreatePSSessionAndUpdateSessionStore` builds `/OcsPowershellOAuth/` IL66 plus `?AdminDomain=` IL82 and calls `IPSCommands.NewPSSession` IL448. `PSCommands`' explicit implementation adds `New-PSSession` IL15 and invokes IL105. The builder's OAuth credential label IL233 does not prove ordinary username/password support.

This is a concrete remote PowerShell invocation boundary, not an unidentified alternative REST route. The local package does not contain the remote server's binding, result schema, validation, storage/effect contract or effective authorization. Modern paths and local compiled DTOs cannot settle that branch. First-party server contracts or a separately authorized pinned legacy comparison are required. The hidden control can have an internal disposition; the default public command's conditional legacy branch remains **unknown/unqualified**, not excluded.

## Routes, methods and audience boundaries

Compiled PolicyRp `HttpRequestHelper<T>` async state machines settle GET selected configuration (`GetAsync>d__2.MoveNext` IL30/60/77), POST collection (`CreateAsync>d__4` IL55/298), PATCH selected configuration (`SetAsync>d__6` IL55/72/284), and DELETE selected configuration (`RemoveAsync>d__5` IL55/72/146) under `/Skype.Policy/configurations/{type}/configuration/{identity}` (collection omits `/configuration/{identity}`). They apply to the C cells above, including pending compiled families; they do not establish P command selection.

The modern generated source is separately traceable in Private `Microsoft.Teams.ConfigAPI.Cmdlets.Generated.ConfigApiBasedCmdlets` nested state machines:

| Operation | Exact `MoveNext` method / offsets | Contract |
|---|---|---|
| Get selected | `<ConfigurationGetConfiguration>d__226` IL82/104/423 | GET `/Skype.Policy/configurations/{type}/configuration/{name}` |
| Get collection | `<ConfigurationGetConfigurations>d__230` IL82/401 | GET `/Skype.Policy/configurations/{type}` |
| New | `<ConfigurationCreateConfiguration>d__218` IL91/410 | POST collection |
| Remove | `<ConfigurationDeleteConfiguration>d__222` IL74/90/301 | DELETE selected configuration |
| Set | `<ConfigurationUpdateConfiguration>d__234` IL91/113/432 | **PUT** selected configuration |
| Grant user | `<UserGrantPolicy>d__1179` IL83/99/310 | PATCH `/Skype.Policy/users/{identity}/policies/{type}` |
| Grant tenant | `<TenantGrantPolicy>d__1091` IL83/294 | PATCH `/Skype.Policy/tenants/policies/{type}` |
| Grant group | `<GroupGrantPolicyAssignment>d__294` IL83/99/310 | POST `/Skype.Policy/groupPolicyAssignments/{group}/policyTypes/{type}` |

Generated methods escape path segments (`EscapeDataString`, e.g. Get IL96/118; Grant user IL94/110). Generated methods support `apiVersion` and `schemaVersion` query names (Get IL147/199, Set IL156/208), but the selected public helper path **passes null for both and omits both URI parameters**. Helpers Get builds ConfigType/optional ConfigName at 342–345 and invokes internal Get at 367; Set invokes the internal setter at 1046, without binding either version argument. Internal Get wrappers (3107+, duplicate19004+) and Set (37642+) declare these parameters without defaults and forward bound parameters only. Private `GetCsConfiguration_Get1+<ProcessRecordAsync>d__62.MoveNext` checks absent bound keys at IL836/841 and IL878/883, loads null at IL848/890, and calls the API at IL955. Collection Get `d__58` loads null at IL837/879. Set `d__69` loads null at IL848/890 and calls the selected PUT API at IL966. Generated Get selected `d__226` tests IsNullOrEmpty at IL140/192, collection `d__230` at IL118/170, and Set `d__234` at IL149/201; empty fragments are removed at IL243/221/252. The SDK's omission matches that actual path; Broadcast ExposeSDN is a separate helper-added query, not a version default. The selected Set cmdlet `SetCsConfiguration_Update+<ProcessRecordAsync>d__69.MoveNext` IL966 calls `ConfigurationUpdateConfiguration`, the PUT method. A generated sibling `ConfigurationUpdateConfiguration1>d__235` uses PATCH at IL432; its mere presence must not replace the actual helper/cmdlet call chain.

The generated request source contains `https://api.interfaces.records.teams.microsoft.com` (e.g. Get IL396, Set IL405). Environment selection, discovery/forest targeting and token acquisition remain separate from that literal default. Compiled `BaseTpmGetCmdlet<T>.ProcessRecord` IL21–89 obtains an environment-selected ConfigAPI token, with `/user_impersonation` at IL79; Grant IL228–295 has that suffix at IL285. `HttpClientHelper+<SendRequestAsync>d__6.MoveNext` IL32–48 applies Bearer. These settle source audience/scope construction and transport, not delegated-only restrictions, minimum role, application-credential success, licensing or sovereign-cloud support. Existing tenant evidence cannot be expanded from these literals.

## Grants: binding, DTOs and effects are different contracts

Compiled MediaConnectivity and MeetingTemplatePermission Grants use the shared PolicyRp scopes: user PATCH (`GrantToUserAsync>d__7` IL55/69/77/186), tenant PATCH (`GrantToTenantAsync>d__8` IL41/159), group POST (`GrantToGroupAsync>d__9` IL66/80/88/210). `SingleGrantPayload` and `GroupGrantPayload` have camel-case naming attributes; `GrantToTenantPayload` does not. The [published DTO reflection/serialization proof](https://github.com/deploymenttheory/go-sdk-msteams/blob/dc2bb9a/docs/implementation/policies/pol04-primary-source-audit.md#compiled-route-and-grant-proof) establishes user `policyName`/`additionalParameters`, group `policyName`/nullable `priority`, and tenant **`PolicyName`**/`additionalParameters`/`forceSwitchPresent`. This is local source-derived JSON, not a captured request or clear-operation result.

Modern Grants use Helpers 832–835 to reject **unbound** PolicyName, even though its declaration is not mandatory; bound null/empty is a separate state. Helpers 839–848 dispatch group and only add Rank when `$Rank` is truthy, so zero is omitted; 850–863 require Global when Identity is whitespace/omitted, otherwise dispatch user. The helper declares no PassThru. Public proxy exposure of a parameter does not prove its modern helper supports it, nor does a successful assignment receipt prove convergence or client effects. Do not copy modern rank omission into compiled priority DTOs.

Public runtime reflection additionally confirms MediaLogging Grant Identity is a string accepting pipeline and property-name input in its `Identity` set. Compiled MediaConnectivity Grant Identity accepts those forms in `GrantToUser`; its other public `Identity` set is distinct. MediaLogging Get, MeetingTemplateConfiguration Get and MeetingConfiguration Set do not declare pipeline binding in the inspected public parameter attributes. Native per-call methods and caller loops do not by themselves settle PowerShell binder/steppable-pipeline behavior.

## Hard types, defaults and null/list behavior

**MeetingConfiguration.** Core `SetTeamsMeetingConfigurationCmdlet.UpdateProperties` tests bound-key presence (PublishedEntraAuthenticationContexts IL11/16) then calls `PSListModifierHelper.ApplyTo<PublishedEntraAuthenticationContext>` IL66. Its `<UpdateProperties>b__2_0` comparator throws `ErrorCannotRemoveElementsOf("PublishedEntraAuthenticationContexts")` IL0–20. This proves failure when that comparer is invoked, not an unconditional preflight rejection of every Remove-shaped input (an empty existing collection need not compare). The helper clears a null/all-empty modifier at IL14–62, gives Replace precedence IL63–199, otherwise processes Remove then Add IL200–406. Add/replace, omission, bound null and empty are distinct operations.

`PublishedEntraAuthenticationContext` fields are Identity, Id, PublishedName, PublishedDescription, EntraName, EntraDescription and nullable Boolean Published. `set_Id` rejects null IL3/8 and validates regex IL17–28; `.cctor` IL0 initializes **`^c([1-9]|[1-9][0-9])$`**. This does not impose that restriction on every route identity. Meeting configuration ports/ranges are nullable UInt32 in the source (`UpdateProperties` e.g. ClientAudioPort IL446/451), not arbitrary signed numbers. Pending PR #332 must remain separately identified from baseline support.

The audit found a concrete **DTO validation mismatch in the earlier pending head**, separate from correct backend selection: the .NET regex accepts a final LF under `$`, while [PR #332 head 36e72bd `models.go`](https://github.com/deploymenttheory/go-sdk-msteams/blob/36e72bd/msteams/teams_admin_api/teams_meeting_configuration/models.go#L16) uses Go `contextID.MatchString` without accounting for that behavior. Offline reflection created a fresh `PublishedEntraAuthenticationContext` for each value and invoked its Id setter; no cmdlet or transport was invoked:

| Input (JSON notation) | Source setter accepts | Retained value |
|---|---|---|
| `"c1"` | yes | `"c1"` |
| `"c1\n"` | yes | exact `"c1\n"` |
| `"c1\r\n"` | no | null |
| `"c100"` | no | null |

Thus calling the source predicate simply “c1 through c99” loses an observed local edge case. The earlier pending Go validator rejected the accepted final-LF case. The finding was reported before mutation, then fixed separately on [PR #332 at d1a4fbe](https://github.com/deploymenttheory/go-sdk-msteams/commit/d1a4fbe5c53971caa62cc27949ff915f65c090e3): exactly one final LF is ignored for regex comparison, while the Id pointer/string and serialized request retain the original value. The guarded affected-package TestUnit run passed, including rejection of CRLF/repeated LF/c100 and authored PATCH preservation. This audit PR does not import that SDK change. Server acceptance is still unknown: local setter acceptance does not prove persistence or policy behavior.

**MeetingTemplateConfiguration.** Actual public resolution is Get only. Its model has Identity/Description and a getter-only `List<TeamsMeetingTemplateType>` TeamsMeetingTemplates; each template has Identity, Description, Name, Category and getter-only `List<TeamsMeetingOption>` TeamsMeetingOptions. Each option has Identity/Name/Value strings and nullable Boolean IsLocked/IsHidden. Constructors/container initialization and nullable field shapes are source facts, not a license to invent a public write verb. Core `TeamsMeetingTemplateConfiguration..ctor` IL7 calls `CreateNewTeamsMeetingTemplates`, which creates its container IL1/6. Parameterized `TeamsMeetingTemplateType..ctor` IL7 initializes options and IL14 sets Name, whereas its parameterless constructor has only the base call IL1: construction path matters. Raw nested responses and absent/null/empty lists require evidence separate from a scalar Get.

**MeetingTemplatePermissionPolicy.** Set's HiddenMeetingTemplates binding is `PSListModifier<HiddenMeetingTemplate>`; UpdateProperties IL11/16 checks binding, IL66 invokes the shared helper. Its comparator reads both Id properties IL1/7 and calls string `op_Equality` IL12, giving case-sensitive ordinal equality. The helper removes the **first** matching item (`ApplyTo` comparator overload IL264/274/282); it does not remove every duplicate. `HiddenMeetingTemplate.set_Id` rejects null IL3/8, then stores the string IL16, without the published-context regex. Getter-only hidden lists, default template ID omission/null, duplicate IDs and missing-removal behavior must not be flattened into generic JSON-patch semantics.

**MessagingConfiguration.** Storyline is a nullable enum with underlying **Int32**, `Disabled=0`, `Enabled=1`, not a string-only type. FileTypeBlockList is a model `List<string>` and Set uses `PSListModifier<string>`. `SetTeamsMessagingConfigurationCmdlet.UpdateProperties` tests the bound key IL11/16; a nonnull empty modifier takes `List.Clear` IL37/50; otherwise it invokes PowerShell's own `PSListModifier.ApplyTo(IList)` IL69. Its null-path behavior therefore must not be borrowed from the custom helper used by the template/context families. Storyline is copied only after bound-key testing IL190/195/204/209. Unknown numeric/server enum values and representation on responses remain distinct from the two named local enum constants.

**BroadcastConfiguration and modern defaults.** Helpers 353–356 append `?ExposeSDNConfigurationJsonBlob=true` only for this family and a true PropertyBag flag. Helpers 394 project the extended read fields, including **SdnName**, **SdnRuntimeconfiguration** and **SdnApiToken**. Public Set Proxies 39639–39695 instead declares **SdnProviderName**, **SdnRuntimeConfiguration**, and SdnApiToken. This establishes read/write name differences; token-bearing read data must be treated as sensitive too, despite the baseline SDK comment describing a write-only token. This research PR corrects that comment and the matching family-note wording without changing formatting/redaction/serialization behavior; the actual empty-token lab observation is retained. Source does not prove token echo/redaction behavior by the server. Helpers 342–345 omit ConfigName for whitespace Identity; 372–375 apply PowerShell `-Like` to Identity and `Tag:$Filter`, not regex. Set Helpers 1041–1046 inserts Global only when Identity is **absent** from PropertyBag; present-null/present-empty do not take that default. New Helpers 923–928 performs New then a separate selected Get, so read-back can fail after a successful create.

## What remains open

Pinned public exports, local dispatch selection, route/method shapes and the selected type/default predicates above are source-settled. No local reflection supplies server-side constraints, minimum roles, license gates, eventual convergence, effective client policy, successful policy clearing, null semantics for every field, or untested auth/cloud combinations. The 13 conditional proxy commands retain the explicit remote-server dependency above. Pending PR #332/#333 support remains separate from baseline main. Existing family lab/API records retain their own dates, identities, request provenance and scenario limits; this audit neither repeats nor broadens them.

The main inventory tracks all 36 commands/61 sets and evidence links. This source note provides bounded evidence for that ledger; it does not authorize closing #36, mark a modern-only SDK implementation as complete PowerShell parity, or promote controlled fixtures/source interception into live tenant captures.
