# POL-08 pinned primary-source audit

Reviewed 2026-11-26 against repository baseline `bb56e7b` and MicrosoftTeams **8.1.0**. Scope is exactly the **27 commands / 46 parameter sets / six families** assigned `research_owner: POL-08` in [coverage](../../../contracts/coverage/8.1.0.json). Refs #39; this audit does not close the issue. This was offline inspection only: no Go execution, authentication, token probe, tenant operation or new capture.

## Provenance and portable reproduction

All 27 assigned names were independently resolved with `Get-Command` after importing the retained Gallery module at `/private/tmp/msteams-core90-pinned/module/MicrosoftTeams.psd1`. All 15 resolved compiled public `ProcessRecord` methods were decoded: each calls the corresponding generic `BaseTpmGet/New/Set/Remove/GrantCmdlet<T>.ProcessRecord` at decimal **IL1**, then returns at IL6. The other 12 public names resolve to Functions. Merely finding a compiled model or CmdletConfig entry was not used to classify public selection.

The following hashes were independently recomputed. `Core`, `PolicyRp`, `Private`, `Helpers`, `Mapping`, `Proxies` and `Internal` below mean these exact module-relative files. Binary citations give declaring type, method and decimal IL byte offset, not source line numbers.

| Module-relative file | SHA-256 |
|---|---|
| `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll` | `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49` |
| `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll` | `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51` |
| `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll` | `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6` |
| `custom/Merged_custom_PsExt.ps1` | `03dbc3a85e12c98dd1807d604782c6a90d0656569541cd9540738825a7fcb255` |
| `custom/CmdletConfig.json` | `754d149c1a5904158e9f01eb55ba44d15c17b7e66f7556385ea894f4ff84d906` |
| `exports/ProxyCmdletDefinitionsWithHelp.ps1` | `5a8a5a50081b1c6986b93d5b15723e966a2d22b64459a9c300cb2a1bf69bb3b5` |
| `internal/Merged_internal.ps1` | `6ff95158174bab031dc1fc24112a2bc6ae81cf1e2f59db4ecb58afbefc8c2bd9` |

Acquisition and package provenance are documented in [baseline acquisition](../../../tools/baseline/README.md). The [published portable IL reader at commit dc2bb9a](https://github.com/deploymenttheory/go-sdk-msteams/blob/dc2bb9a/docs/implementation/policies/pol04-primary-source-audit.md#portable-reproduction) accepts a caller-selected `ModuleRoot`, checks the private assembly hash, and resolves tokens with declaring-type/method generic arguments. Use it against the types and methods cited here; no ephemeral host helper is required. For Core public resolution and the complete 46-set enumeration:

```powershell
Import-Module (Join-Path $ModuleRoot 'MicrosoftTeams.psd1') -Force -WarningAction SilentlyContinue
$records = (Get-Content ./contracts/coverage/8.1.0.json -Raw | ConvertFrom-Json).records
$records | Where-Object research_owner -eq 'POL-08' | ForEach-Object {
    $c = Get-Command $_.name
    [pscustomobject]@{Name=$c.Name; Kind=$c.CommandType;
        Type=$c.ImplementingType.FullName; Sets=$c.ParameterSets.Name -join ','}
}
```

To reproduce model setter inspection, resolve the Core type `Microsoft.Teams.Policy.Administration.Cmdlets.Core.TeamsAIPolicy` (or `TeamsBYODAndDesksPolicy`, `TeamsWorkLocationDetectionPolicy`, `TeamsWorkLoadPolicy`) and decode `set_<Property>`; reflection `GetProperties()` supplies the types. For shared async transport, decode the named nested state-machine `MoveNext`, not just its outer kickoff method.

## Exact assigned public selection

`C` means compiled Core public cmdlet, with type `Microsoft.Teams.Policy.Administration.Cmdlets.Core.<Verb><Family>Cmdlet` (`Cs` is absent from that type). `P` means public proxy Function, forwarded to the same-named Private cmdlet. Every assigned public name is listed, including the parameterless compiled Grant Identity set.

| Actual resolved command | Kind | Actual parameter sets |
|---|---|---|
| `Get-CsTeamsAIPolicy` | C | `Identity`, `Filter` |
| `Get-CsTeamsBYODAndDesksPolicy` | C | `Identity`, `Filter` |
| `Get-CsTeamsIPPhonePolicy` | P | `Identity`, `Filter` |
| `Get-CsTeamsWorkLoadPolicy` | P | `Identity`, `Filter` |
| `Get-CsTeamsWorkLocationDetectionPolicy` | C | `Identity`, `Filter` |
| `Get-CsTenantMigrationConfiguration` | P | `Identity`, `Filter` |
| `Grant-CsTeamsAIPolicy` | C | `Identity`, `GrantToUser`, `GrantToGroup`, `GrantToTenant` |
| `Grant-CsTeamsBYODAndDesksPolicy` | C | `Identity`, `GrantToUser`, `GrantToGroup`, `GrantToTenant` |
| `Grant-CsTeamsIPPhonePolicy` | P | `Identity`, `GrantToTenant`, `GrantToGroup` |
| `Grant-CsTeamsWorkLoadPolicy` | P | `Identity`, `GrantToTenant`, `GrantToGroup` |
| `Grant-CsTeamsWorkLocationDetectionPolicy` | C | `Identity`, `GrantToUser`, `GrantToGroup`, `GrantToTenant` |
| `New-CsTeamsAIPolicy` | C | `Identity` |
| `New-CsTeamsBYODAndDesksPolicy` | C | `Identity` |
| `New-CsTeamsIPPhonePolicy` | P | `Identity` |
| `New-CsTeamsWorkLoadPolicy` | P | `Identity` |
| `New-CsTeamsWorkLocationDetectionPolicy` | C | `Identity` |
| `Remove-CsTeamsAIPolicy` | C | `Identity` |
| `Remove-CsTeamsBYODAndDesksPolicy` | C | `Identity` |
| `Remove-CsTeamsIPPhonePolicy` | P | `Identity` |
| `Remove-CsTeamsWorkLoadPolicy` | P | `Identity` |
| `Remove-CsTeamsWorkLocationDetectionPolicy` | C | `Identity` |
| `Set-CsTeamsAIPolicy` | C | `Identity` |
| `Set-CsTeamsBYODAndDesksPolicy` | C | `Identity` |
| `Set-CsTeamsIPPhonePolicy` | P | `Identity` |
| `Set-CsTeamsWorkLoadPolicy` | P | `Identity` |
| `Set-CsTeamsWorkLocationDetectionPolicy` | C | `Identity` |
| `Set-CsTenantMigrationConfiguration` | P | `Identity` |

The three C families each own nine sets; WorkLoad and IPPhone each own eight; TenantMigration owns three. The independent enumeration totals 46 and agrees with [runtime metadata](../../../contracts/powershell/8.1.0/runtime.json). All six native implementations already exist at the baseline. Coverage status labels such as metadata-only/not-run do not erase the acceptance records summarized below, and acceptance prose does not establish untested branches.

| Family | Public selection and SDK backend | Exact implementation source |
|---|---|---|
| WorkLoadPolicy | P, conditional modern or remote TRPS; SDK modern ConfigAPI | [teams_work_load_policy/service.go](../../../msteams/teams_admin_api/teams_work_load_policy/service.go) |
| WorkLocationDetectionPolicy | C; SDK compiled PolicyRp | [teams_work_location_detection_policy/service.go](../../../msteams/teams_admin_api/teams_work_location_detection_policy/service.go), local [resource.go](../../../msteams/teams_admin_api/teams_work_location_detection_policy/resource.go) |
| AIPolicy | C; SDK compiled PolicyRp | [teams_ai_policy/service.go](../../../msteams/teams_admin_api/teams_ai_policy/service.go), local [resource.go](../../../msteams/teams_admin_api/teams_ai_policy/resource.go) |
| BYODAndDesksPolicy | C; SDK compiled PolicyRp | [teams_byod_and_desks_policy/service.go](../../../msteams/teams_admin_api/teams_byod_and_desks_policy/service.go), local [resource.go](../../../msteams/teams_admin_api/teams_byod_and_desks_policy/resource.go) |
| IPPhonePolicy | P, conditional modern or remote TRPS; SDK modern ConfigAPI | [teams_ip_phone_policy/service.go](../../../msteams/teams_admin_api/teams_ip_phone_policy/service.go) |
| TenantMigrationConfiguration | P, conditional modern or remote TRPS; SDK modern ConfigAPI | [tenant_migration_configuration/service.go](../../../msteams/teams_admin_api/tenant_migration_configuration/service.go) |

The unusual coverage resource IDs `teamsai_policy`, `teamsbyod_and_desks_policy`, `teamsip_phone_policy` are ledger identifiers, not Go directory names. The source links above are derived from the assigned implementation symbols. No SDK backend mismatch was found in this scope; no SDK mutation was made.

## PowerShell default branch and legacy boundary

Mapping independently selects Get/New/Set/Remove-CsConfigurationModern and Grant-CsTeamsPolicy for `TeamsWorkLoadPolicy` (1878–1890, 2458–2470, 3202–3214, 3832–3844, 4594–4606), `TeamsIPPhonePolicy` (1683–1695, 2276–2288, 3091–3103, 3706–3718, 4427–4439) and TenantMigration Get/Set (1959–1971, 4650–4662). Each uses its exact family `ConfigType` or `PolicyType`. There are no Mapping entries for the three C families. Actual Proxies definitions begin at WorkLoad 11757/16969/26460/32221/42257, IPPhone 9151/15769/24266/30760/39546, TenantMigration 12573/42634 (Get/Grant/New/Remove/Set order). They build a steppable pipeline around the Private command, rather than calling the helper unconditionally.

Private `Microsoft.Teams.ConfigApi.Cmdlets.TrpsPSCmdlet.BeginProcessing` IL55/60 checks hidden `MsftInternalProcessingMode`, IL91/99/107 accepts `Legacy`, `TryModern`, `ForceModern`, and IL155/160 removes the control before forwarding. `TryGetFlightedApi` IL46/64/91 checks modes and IL128 calls `FlightingUtils.GetFlightingCommandInfo`; the latter reads `SessionConfiguration.RemotingCmdletsFlightedForAutoRest` at IL59 and modern command availability at IL98. Therefore omission of the hidden control leaves session-dependent flighting; the SDK modern branch is not proof that ordinary default PowerShell always chooses modern.

`TrpsPSCmdlet.getProcessor` IL52/72 constructs modern processors and IL82 calls `CreateTrpsPipelineProcessor`. The latter obtains the public name at IL21–31, remote session/runspace IL36/41, cleaned parameters IL53 and remote processors IL58/139. The [published primary remote-chain proof](https://github.com/deploymenttheory/go-sdk-msteams/blob/dc2bb9a/docs/implementation/policies/pol04-primary-source-audit.md#conditional-dispatch-and-the-precise-legacy-boundary) follows that same pinned assembly through `RemoteSyncProcessor.ReceiveOutput` (`AddCommand` IL16, Runspace IL42, Invoke IL49), asynchronous BeginInvoke/EndInvoke, and `TrpsSessionBuilder.CreatePSSessionAndUpdateSessionStore` (`/OcsPowershellOAuth/` IL66, `?AdminDomain=` IL82, NewPSSession IL448). This is remote PowerShell execution. The remote server's validation, storage, results, authorization and effects are not present in the local module; this audit leaves that branch unknown/unqualified. Compiled sibling models cannot be promoted into its contract.

## Native routes, defaults and authority

Every route below is under `/Skype.Policy`; `{type}` is exactly `TeamsWorkLoadPolicy`, `TeamsWorkLocationDetectionPolicy`, `TeamsAIPolicy`, `TeamsBYODAndDesksPolicy`, `TeamsIPPhonePolicy`, or `TenantMigrationConfiguration`, as selected above. `{id}` is an escaped identity; collection routes omit `/configuration/{id}`.

| Operation | Compiled PolicyRp (three C families) | Selected modern (three P families) |
|---|---|---|
| Get/list | GET `/configurations/{type}/configuration/{id}` or GET `/configurations/{type}` | Same routes/methods |
| New | POST collection | POST collection (WorkLoad/IPPhone only) |
| Set | PATCH selected configuration | **PUT** selected configuration |
| Remove | DELETE selected configuration | DELETE selected configuration (WorkLoad/IPPhone only) |
| Grant user | PATCH `/users/{id}/policies/{type}` | Same |
| Grant tenant | PATCH `/tenants/policies/{type}` | Same |
| Grant group | POST `/groupPolicyAssignments/{id}/policyTypes/{type}` | Same |

Independent PolicyRp IL decoding: `HttpRequestHelper<T>+<GetAsync>d__2.MoveNext` route fragments IL30/60/77; `<CreateAsync>d__4` collection IL55, POST IL298; `<SetAsync>d__6` route IL55/72, PATCH IL284; `<RemoveAsync>d__5` route IL55/72, DELETE IL146. Grant user `<GrantToUserAsync>d__7` route IL55/69/77 and PATCH IL186; tenant `d__8` IL41/PATCH159; group `d__9` IL66/80/88/POST210. Configuration normalization comes from `ConfigurationNameNormalization<T>.GetNormalizedNameForPolicyRP`, not arbitrary coverage IDs. Its IL1/6/11 loads the model type name, IL17/23 checks the normalization list, and IL71–76 returns the unmodified name for types outside it. Independently invoking this pure static method closed over each of the three assigned Core model types returned exactly `TeamsAIPolicy`, `TeamsBYODAndDesksPolicy`, and `TeamsWorkLocationDetectionPolicy`; no doubled Teams prefix or ledger-ID route is inferred.

Private generated `Microsoft.Teams.ConfigAPI.Cmdlets.Generated.ConfigApiBasedCmdlets` selected state machines are `<ConfigurationGetConfiguration>d__226` route IL82/104 and GET423; `<ConfigurationGetConfigurations>d__230` collection IL82/GET401; `<ConfigurationCreateConfiguration>d__218` IL91/POST410; `<ConfigurationDeleteConfiguration>d__222` IL74/90/DELETE301; `<ConfigurationUpdateConfiguration>d__234` IL91/113/**PUT432**. Grants: `<UserGrantPolicy>d__1179` IL83/99/PATCH310, `<TenantGrantPolicy>d__1091` IL83/PATCH294, `<GroupGrantPolicyAssignment>d__294` IL83/99/POST310. Generated segment escaping uses `EscapeDataString` (Get IL96/118). The generated PATCH sibling is not the selected Set helper path.

Helpers Get 342–345 passes ConfigType/optional ConfigName, invokes Internal Get at 367, parses JSON 369–371, locally filters Identity or `Tag:`Identity at 374–377, then formats at 380. Helpers Set 1041–1044 defaults **absent** PropertyBag Identity to `Global`, and 1046 calls Internal Set with Body/ConfigName. An explicitly bound empty/null identity is not proven identical to omission. New 926–933 submits POST and, unless `$err`, performs a separate Get read-back; Remove 987 forwards identity without a body. Modern SDK Create wrappers likewise distinguish submission success from read-back errors; one sparse Update PUT is not a compiled preparatory GET/PATCH composition.

Modern generated source offers `apiVersion` and `schemaVersion`, but the selected Helpers do not bind either. Internal wrappers declare them without defaults. Private `GetCsConfiguration_Get1+<ProcessRecordAsync>d__62.MoveNext` tests bound keys at IL836/841 and IL878/883, then supplies null IL848/890 before API call IL955; collection d__58 supplies null IL837/879. `SetCsConfiguration_Update+<ProcessRecordAsync>d__69` supplies null IL848/890, calling selected PUT at IL966. Generated Get checks IsNullOrEmpty IL140/192 and Set IL149/201; empty query fragments are removed. The SDK correctly omits both version queries; do not invent a schema-year default from observed XML schema names.

Private generated requests contain default `https://api.interfaces.records.teams.microsoft.com` (Get IL396, Set IL405). This literal is not an authorization guarantee. PolicyRp `BaseTpmGetCmdlet<T>.ProcessRecord` selects the token environment IL21, appends `/user_impersonation` IL79, obtains host endpoint separately; Grant repeats selection IL228–295 and scope suffix IL285. SDK modern services use identity-bound `internal/configapi` discovery; compiled local resources use `policytransport` with TeamsAdmin scope and shared PolicyRp decoding. Audience construction/endpoint discovery, returned XML `Authority`, and effective tenant permission are separate facts. Existing application-mode success does not settle minimum role, delegated permissions, every cloud or licensing.

## Models, omission, null and constraints

Runtime reflection independently confirms these **public New/Set** bindings; no enum/range validation attributes appeared on their settings. Identity/Force/hidden processing controls are not settings.

| Family | Public settings and binding types | Primary model constraint proof |
|---|---|---|
| WorkLoad | **Seven**: Description String; AllowCalling, AllowCallingPinned, AllowMeeting, AllowMeetingPinned, AllowMessaging, AllowMessagingPinned **non-nullable Boolean** | Core `TeamsWorkLoadPolicy` exists with nullable Boolean model properties, but all five actual public commands are P. Its model does not alter public Boolean binding or justify JSON null writes. |
| WorkLocation | EnableWorkLocationDetection `Nullable<Boolean>`; UserSettingsDefault String | Core setters assign at IL2 and return IL7; no enum/range/throwing validation. |
| AI | Description, EnrollFace, EnrollVoice, SpeakerAttributionForBYOD, PassiveVoiceEnrollment String | Enrollment/attribution setters assign IL2, return IL7. Description loads lengths IL4/16, constants 1 IL9 and 1040 IL21, **pops** IL10/11/26/27, stores IL30; no comparison or throw. Constants alone do not establish a 1–1040 local restriction. |
| BYOD | DeviceDataCollection String | Core setter assigns IL2, returns IL7; no enum or range. |
| IPPhone | Description, AllowBetterTogether, AllowHomeScreen, SearchOnCommonAreaPhoneMode, SignInMode String; AllowHotDesking Boolean; HotDeskingIdleTimeoutInMinutes **Int64** | Public Proxies/runtime settle the binding. No invented Int32 conversion, positive minimum, timeout maximum or enumerated string validation. |
| TenantMigration | MeetingMigrationEnabled non-nullable Boolean | Public runtime lacks the Instance/Tenant/Force parameters found in some help; only Get/Set are exported in this scope. |

The actual Core `NewTeamsAIPolicyCmdlet.UpdateProperties` and `SetTeamsAIPolicyCmdlet.UpdateProperties` check `BoundParameters.ContainsKey` at IL16/51/86/121/156 and call model setters IL30/65/100/135/170, respectively for Description, EnrollFace, EnrollVoice, **SpeakerAttributionForBYOD**, **PassiveVoiceEnrollment**. BYOD New/Set checks DeviceDataCollection IL16 and calls its setter IL30. WorkLocation New/Set checks EnableWorkLocationDetection IL16 and UserSettingsDefault IL51, calling setters IL30/65. Thus omitted keys skip setter assignment, while a bound value reaches the model setter. These are exact pinned field mappings; a current documentation spelling or omitted help entry cannot rename or remove them.

All models include Identity. C model string properties can be null; WorkLocation's Boolean is nullable. In each C package, `models.go` uses nullable typed projections plus complete RawFields/RawJSON, and `policy.Field[T]` settings distinguish absent/null/bound empty or false. Modern WorkLoad/IPPhone `Settings` pointers express omission versus explicit false/empty/Int64; they do not manufacture a public null Boolean binding. `TenantMigrationConfiguration` likewise distinguishes omitted/false/true in its sparse PUT. The server remains authoritative for accepted values; absence of local validation does not prove every value persists or has an effect.

Compiled Set is a separate orchestration contract: PolicyRp `BaseTpmSetCmdlet<T>.ProcessRecord` checks any parameters IL22, emits "Nothing to update" IL52/WriteError115, obtains existing result IL257–277, calls `UpdateProperties` IL363, then writes IL382–401. Each C SDK service exposes direct `PatchByIdentity` separately from `UpdateByIdentity` read/merge/PATCH composition, preserving Before/After/ReadResponse and partial acceptance. Model `RawJSON` is an owned snapshot, not license to turn observed nullable response properties into modern public null inputs.

### Work-location constructor defaults and bound serialization

Fresh local `TeamsWorkLocationDetectionPolicy` instances have Identity, EnableWorkLocationDetection and UserSettingsDefault all null. Core model `.ctor` only calls `BasePolicy::.ctor` at IL1/ret6; BasePolicy only calls Object at IL1/ret7. The New cmdlet constructor adds no setting defaults; generic BaseTpmNew initializes Force=false at IL1–7 and then its base at IL13. `NewTeamsWorkLocationDetectionPolicyCmdlet.CreateNew` constructs the model at IL0/ret5. Consequently the help-described False/Disabled outcomes are not values injected by those local constructors.

Compiled Create filtering is also source-settled. PolicyRp `HttpRequestHelper<T>+<CreateAsync>d__4.MoveNext` reads BoundParameters at IL119, calls GetNestedBoundParameters with false at IL146–147, constructs `Providers.Helpers.BoundParametersContractResolver` at IL181, sets it at IL186 and calls JsonConvert.SerializeObject at IL209. The resolver's CreateProperty installs ShouldSerialize at IL35–46; its predicate tests boundParameters.ContainsKey(property.PropertyName) at IL17/22/27. It filters by binding, not blanket null removal. Pure local serialization through that exact resolver produced these controlled, non-tenant vectors:

```json
{"Identity":"sample"}
{"Identity":"sample","EnableWorkLocationDetection":null,"UserSettingsDefault":null}
{"Identity":"sample","EnableWorkLocationDetection":false}
```

These respectively represent omitted settings, explicitly bound null settings and a bound false Boolean. New/Set's bound-key checks therefore cannot be replaced by an inferred default injection or by dropping every null. The original model/setter/serializer evidence is local; actual default persistence still belongs to the server/recorded scenario. The portable reader independently reproduced the Create method offsets above without executing the target method.

## Grants, typed responses and errors

C Grants expose `GrantToUser` (Identity), `GrantToGroup` (Group and signed **Int32 Rank required**), `GrantToTenant` (Global, optional Force), plus the default parameterless Identity set. PolicyRp `BaseTpmGrantCmdlet<T>.ProcessRecord` checks AreParametersSpecified IL9, emits "Nothing to grant" IL39 and WriteError IL102 for no scope. Success only writes verbose text IL358–379; it does not establish assignment convergence.

Shared compiled DTO serialization is independently bounded by the [pinned primary DTO proof](https://github.com/deploymenttheory/go-sdk-msteams/blob/dc2bb9a/docs/implementation/policies/pol04-primary-source-audit.md#compiled-route-and-grant-proof): user `policyName`/`additionalParameters`; group `policyName`/nullable `priority`; tenant **`PolicyName`**/`additionalParameters`/`forceSwitchPresent`. Each C package's `assignment.go` mirrors this casing and signed rank. Missing/mixed scopes are rejected locally. Nil/empty PolicyName and signed/zero ranks remain representable; clear persistence is not inferred.

P Grants expose Identity (user), GrantToTenant and GrantToGroup, with nullable Int32 Rank, no public Force on WorkLoad/IPPhone. Helpers 832–835 rejects an **unbound** PolicyName even though declaration is optional; bound empty/null is a different state. Helpers 839–848 selects group, adds Rank only if truthy (zero omitted); 850–863 requires Global when Identity is absent/whitespace, otherwise selects user. The SDK high-level `Assign` preserves zero omission; low-level `AssignToGroup` can send explicit priority zero as a native operation. Low-level tenant force support is broader than these public family wrappers. Helpers do not declare PassThru. No family Grant was live-qualified by owned CRUD tests.

C HTTP results are decoded through `policyrp.Decode`/`PolicyResponse`, retaining native response/error metadata, and typed nullable Policy plus raw fields. Core Get writes the policy result (PolicyRp BaseTpmGet ProcessRecord IL298); New writes PolicyResult IL410/436; Set/Grant are not promises of returned updated objects. Modern Get parses bare list/object JSON; native POST201, PUT204 and DELETE204 may have empty bodies. WorkLoad/IPPhone `Create` reads back separately; group receipt distinguishes 202 acceptance (optional OperationId decoding) from 204 without polling. Malformed success bodies remain decoding failures even where submission was accepted. None is proof of effective assignment.

Each family service checks context/backend/input and retains HTTP response on failures; malformed model/list fixtures test `ErrDecode` separately from transport/API failures. Mutation replay is disabled. Native non-success envelopes are propagated by shared clients, not reclassified as “unsupported.” Existing tests that reuse another service's denial envelope prove parser propagation only, not a live denial for this family. Private/PolicyRp error processing (`BaseTpmSet` ProcessException IL500/WriteError505; Grant IL424/429; New IL481/486) preserves the distinction between command error behavior and SDK typed errors. Acceptance status alone does not prove persistence.

## Existing live qualification and remaining unknowns

These are existing repository records and their exact bounded scenarios; no new tenant evidence was generated in this audit.

| Family / existing evidence | Actually recorded qualification | Still open |
|---|---|---|
| [WorkLoad](voice-applications-workload-contract-notes.md) | Client-secret and certificate Global/list/filter/getter parity; owned unassigned policy created with all six Booleans false, sparse Description-only PUT preserving them, pinned comparisons, delete/independent GET404/list absence; recorded cleanup. | Grant persistence/clears, other setting writes/nulls, effective client behavior; legacy/default-flight behavior. |
| [WorkLocation](work-location-detection-policy-lab-acceptance.md) | Both application modes: read/list/filter; owned true/Enabled create, only Boolean false update preserving other fields, PowerShell parity, deletion/absence/cleaned journal. | UserSettingsDefault null/empty/other values, Grant persistence, effective Wi-Fi/peripheral detection, delegated/cloud coverage. |
| [AI](ai-policy-lab-acceptance.md) | Both application modes: all five getters, owned Description + EnrollFace Disabled create, EnrollFace Enabled and explicit-null Description update, parity and cleanup. | Other enrollment/attribution writes, arbitrary enum/length/null cases, grants, effective enrollment/attribution, delegated/cloud coverage. |
| [BYOD](byod-and-desks-policy-lab-acceptance.md) | Both application modes: read/list/filter and owned DeviceDataCollection Disabled→Enabled CRUD, parity, cleanup. | Explicit null/empty/other values, grants and effective device collection, delegated/cloud coverage. |
| [IPPhone](ip-phone-policy-lab-acceptance.md) | Both application modes: all seven getters; owned CommonAreaPhoneSignIn/false/30 create, sparse Description/HomeScreen/BetterTogether update, parity and cleanup. | Remaining mode writes, timeout boundaries, empty-string persistence, grants, effective device behavior; legacy/delegated/cloud coverage. |
| [TenantMigration](tenant-migration-configuration-lab-acceptance.md) | Both application modes: Global reads/list/getter parity; enabled-original MeetingMigrationEnabled false update and journaled restoration to true, three pinned comparisons, cleaned journal. Recorded authority changed Host→Tenant after writing. | Enabling from a disabled original, non-Global identities, effective MMS behavior; legacy/delegated/cloud coverage. |

TenantMigration is **Meeting Migration Service (MMS)** configuration (`TenantMigrationConfiguration`, observed schema `Settings.TenantMigration.2016`) and public `MeetingMigrationEnabled`. It is separate from `TeamsMigrationConfiguration` / `EnableLegacyClientInterop`: verify their distinct Mapping entries (TeamsMigration Get 1738, Set 4482; TenantMigration Get 1959, Set 4650) and [separate SDK package](../../../msteams/teams_admin_api/teams_migration_configuration/service.go). Similar route shapes do not merge their models, client interop effects or qualification. The TenantMigration owned read/update/restore evidence does not imply New/Remove support or successful triggering of meeting migration.

Remaining source gaps are precise: the server-owned legacy implementation; minimum role/auth/cloud/licensing contracts; server-accepted values beyond recorded vectors; all five policy-family Grant persistence/clear/convergence; effective end-user/device/MMS effects; and modern default PowerShell flight selection for any particular live session. Complete source parity and native SDK availability do not close these gaps. The retained pinned module is required for subsequent extension research and was preserved.
