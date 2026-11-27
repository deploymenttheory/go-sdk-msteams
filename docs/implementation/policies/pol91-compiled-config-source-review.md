> Current approved decisions and corrected read-profile execution are in [the final report](pol91-final-contract-decisions.md); older canonical-only qualifications below remain bounded historical evidence. Live status is unchanged.

# POL91 first six compiled command contracts

Review scope: Get/Set for Teams client, privacy and online voicemail validation, six commands / nine sets. This is the source and execution basis for individual decisions; the wider POL91 scope remains 39 commands / 59 sets. No decision is inferred merely from a symbol or test mapping.

## Source binding and model inventory

All six resolve directly to Core cmdlets and closed PolicyRp Get/Set bases, without a public Trps/remoting selection. Effective Get Identity is optional with ValidateNotNullOrEmpty: omission selects a collection and maps to native List; GetByIdentity selects an explicit path. All three Filter sets inherit ValidateNotNullOrEmpty. Set Identity is mandatory with that same validator. These are string parameters. The SDK's additional path-boundary guards against blank/dot/backslash/control or invalid scoped segments are native path safety, not a claim that the public binder alone rejects whitespace or has identical early-error ordering.

| Model | Complete managed setting types | Source control |
|---|---|---|
| Core.TeamsClientConfiguration | 13 Nullable<Boolean>: AllowEmailIntoChannel, AllowDropBox, AllowBox, AllowGoogleDrive, AllowShareFile, AllowEgnyte, AllowOrganizationTab, AllowSkypeBusinessInterop, AllowResourceAccountSendMessage, AllowGuestUser, AllowScopedPeopleSearchandAccess, AllowRoleBasedChatPermissions, ExtendedWorkInfoInPeopleSearch. Four String: RestrictedSenderList, ContentPin, ResourceAccountContentAccess, UseUnifiedDomain. | All 17 effective public parameter types match the DTO property types; none has a parameter validator. Concrete AreParametersSpecified selects BoundParameters.ContainsKey for each setting; UpdateProperties assigns selected values. Strings are not invented arrays or local enums. |
| Core.PrivacyConfiguration | Four Nullable<Boolean>: EnablePrivacyMode, AutoInitiateContacts, PublishLocationDataDefault, DisplayPublishedPhotoDefault. | Public/model nullable types independently verified; all four bound-key settings count, including false/null. No service default is synthesized. |
| Core.OnlineVoicemailValidationConfiguration | AudioFileValidationEnabled:Nullable<Boolean>; AudioFileValidationUri:String. | Bound-key checks at IL16/39. URI model setter separately permits null and applies the ASCII-case HTTPS regex with one final-LF allowance; no URI parser, DNS, length or reachability rule is invented. The existing controlled setter observations exercise that rule independently of denied request probes. |

Identity and received non-management metadata remain separate. Native pointer/raw snapshots preserve omission/null/future properties instead of materializing CLR defaults. Client ContentPin may appear in sensitive received data; redacted formatting and owned snapshots are retained. All known model fields, nullable/false/empty behavior, request projections and raw ownership/type rejection are exercised by existing complete-package model tests. A raw model can retain response data; separate write requests prevent accidental full-object update.

## Null versus omission: actual serializer branch

Concrete Set guards use bound-key presence. Shared PolicyRp `HttpRequestHelper<T>+<SetAsync>d__6::MoveNext` constructs BoundParametersContractResolver at IL185, assigns it to JsonSerializerSettings.ContractResolver190 and calls JsonConvert.SerializeObject213. `BoundParametersContractResolver.CreateProperty` installs ShouldSerialize at IL46. Its `<>c__DisplayClass2_0::<CreateProperty>b__0` reads JsonProperty.PropertyName17 and returns boundParameters.ContainsKey22; it does not test the property's value. No NullValueHandling override is installed by SetAsync.

A pure local probe, using the actual hash-checked Core Privacy DTO and actual resolver, selected EnablePrivacyMode=null and AutoInitiateContacts=false. The result was exactly `{"EnablePrivacyMode":null,"AutoInitiateContacts":false}` with default NullValueHandling=Include; unbound other properties were omitted. This is authored local constructor/property/serializer evidence, not a cmdlet invocation, authentication, telemetry or HTTP capture.

Portable reproduction on the verified extracted module (no ProcessRecord):

```powershell
Import-Module (Join-Path $ModuleRoot 'MicrosoftTeams.psd1')
$t = (Get-Command Set-CsPrivacyConfiguration).ImplementingType.BaseType.GetGenericArguments()[0]
$dto = [Activator]::CreateInstance($t)
$t.GetProperty('EnablePrivacyMode').SetValue($dto, $null)
$t.GetProperty('AutoInitiateContacts').SetValue($dto, $false)
$bound = [Collections.Generic.Dictionary[string,object]]::new()
$bound.Add('EnablePrivacyMode', $null)
$bound.Add('AutoInitiateContacts', $false)
$r = [AppDomain]::CurrentDomain.GetAssemblies() | ForEach-Object {
    $_.GetType('Microsoft.Teams.Policy.Administration.Cmdlets.Providers.Helpers.BoundParametersContractResolver', $false)
} | Where-Object { $_ } | Select-Object -First 1
$settings = [Newtonsoft.Json.JsonSerializerSettings]::new()
$settings.ContractResolver = [Activator]::CreateInstance($r, [object[]]@($bound))
[Newtonsoft.Json.JsonConvert]::SerializeObject($dto, $settings)
```

Verify the package and loaded assembly paths against these hashes before model construction: package `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`; netcore Core.dll `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`; Providers.PolicyRp.dll `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`. The portable IL reader linked in the initial audit resolves operands with owning generic context. Effective metadata can be reproduced with Get-Command.Parameters and the closed model's GetProperties, independently of immutable preview validation arrays.

## Set preparation, errors and execution boundaries

The shared Set no-settings branch is WriteError, before its token path, not quiet success. Native ErrInput before I/O represents that error-return boundary; caller controls confirmation and continuation. Public Set composes a preparation read and selected sparse PATCH, using the received body Identity independently of the route. The native PatchByIdentity primitive is additional explicit behavior and is not mislabeled a public Set without preparation. HTTP acceptance, typed received output and effective persistence are separate; no automatic readback, restoration, replay or transaction is inferred.

The initial audit records exact four-package proofs at 42fd88e (87948/29006), clean root 0df43de / 93343 for success/denial/malformed/cancellation/recovery and distinct retained versus authored bytes. New review identified two bounded evidence gaps: privacy's duplicate filter test never exercised successful filtering/actual whitespace, and voicemail Set never reached an accepted submission. The pending source-authored replacements close those branches without fabricating a tenant capture or nondefault privacy scope. Client's new typed PATCH200/malformed200 full-package result is separately bound to 01e469b/33122; its initial outer-whitespace test assertion failure is a corrected test oracle, not a production regression. Final new privacy/validation execution and individual status decisions are recorded only after their source checkpoint passes.

Live authorization/minimum role, actual tenant change/readback/oracle/restore and client effects remain unqualified. Nothing in this source review promotes live status or closes #185.

## Shared endpoint, audience, response and control evidence

The [merged PolicyRp route/result/auth investigation](pol04-primary-source-audit.md#compiled-bodies-and-grant-casing) and its [authentication boundary](pol04-primary-source-audit.md#auth-provenance-and-unresolved-dispositions), plus the [independent normalization trace](pol08-primary-source-audit.md), are source evidence rather than SDK assertions. In the same pinned PolicyRp assembly, HttpRequestHelper GetAsync.MoveNext route fragments are at IL30/60/77; SetAsync route fragments55/72 and PATCH284. GetAll uses the collection route. These establish direct GET `/Skype.Policy/configurations/{type}` or `.../configuration/{identity}` and selected sparse PATCH, without generated forest discovery. Pure ConfigurationNameNormalization<T>.GetNormalizedNameForPolicyRP calls closed over these three actual Core types returned OnlineVoicemailValidationConfiguration, PrivacyConfiguration and TeamsClientConfiguration exactly. No ledger ID or inferred doubled prefix defines their paths.

BaseTpmGet.ProcessRecord selects an environment-mapped ConfigAPI token at IL21–89, appends `/user_impersonation`79 and acquires the host token89; HttpClientHelper.SendRequestAsync applies Bearer at IL32–48. MPACmdlet, command and correlation metadata distinguish this compiled protocol; there is no generated forest target. Native policytransport uses the session's fixed TeamsAdmin origin/audience and encoded path components, with credential/destination enforcement in the shared scope. These facts do not establish delegated-only access, minimum role, application success in every cloud or tenant eligibility.

Compiled HttpClientHelper.SendAsync<T>.MoveNext records IsSuccessStatusCode at IL212/217, reads content297 and deserializes the policy projection452. This is broad 2xx transport success, not a universal fixed 200/204 result guarantee. Native Scope.Do retains the HTTP receipt/error; policyrp.Decode returns no typed value for an empty successful body and ErrDecode plus response for malformed nonempty data. The new typed 200 and empty 204 tests exercise both branches without declaring server persistence. Public Set/WriteObject presentation is not a promise of a returned updated DTO; the additional native primitive can decode a received DTO, while a prepared Update keeps Before, ReadResponse, After, PatchInvoked and HTTPAccepted distinct. Caller-requested readback/restoration is separate.

Fresh BaseTpmGet.ProcessRecord IL101/106 checks Identity absence: selected Get versus collection. Collection handling invokes FilterPolicies167 then two source ForEach passes188/208. The corresponding b__1/b__2 lambdas compare Identity with literal Global using String.Equals(String) at IL12/17 and write the two groups at39/42, preserving original relative order. Native stable exact-Global-first order follows that source behavior. FilterPolicies script literal at IL64 is `$identity -like "$filter" -or $identity -like "Tag:$filter"`; effective public nonempty validation is separate from its omission branch. No paging or continuation envelope is fabricated. Privacy success vectors author Global only; they do not assert nondefault scope eligibility.

Shared no-settings Set WriteError and sparse selected-property serialization are described above. Force-only does not count as a setting; the source-intercepted request probes cover explicit Force metadata independently of successful model decoding. Native mutation requests use ReplayNever, with raw normalized HTTP/API/decode/cancellation errors; tests assert no forbidden next PATCH/replay and preserve earlier independent state. PowerShell warnings/ErrorAction/ShouldProcess ordering is not silently presented as an identical Go host pipeline: native errors/receipts and caller-controlled confirmation are explicit compatibility boundaries. Existing complete-package failure tests and clean root `93343` cover those propagation/recovery paths; neither a denied probe nor a source symbol is counted as successful submission.

The final new privacy/list and voicemail accepted branches passed full TestUnit/Example at clean `0ef95f7d34e6454e26abef8252e9d3f85b07c8df`, session `72461`, owned TMPDIR/GOMAXPROCS=2/`-p 1 -count=1` / 14 GiB guard (43.99 GiB admission). Client full-package proof remains independently scoped to `01e469be0e8929a89be6e4f89298a8438177964b` /33122, with shared input diffs unchanged. No extra test permutation or production change was needed for this shared-source review.

## Actual compiled public binder closure

The independent [13-row field-binding artifact](evidence/pol91-field-binding.json) and [parameterized reproducer](../../../scripts/discovery/observe-pol91-field-binding.ps1) use the actual PowerShell compiled-cmdlet binder with source-effective declared attributes, then actual pinned Core property setters and BoundParametersContractResolver. The helper [C# probe](../../../scripts/discovery/pol91-field-binding-fakes.cs) replaces the management lifecycle with recording only; no original management ProcessRecord, authentication or HTTP runs. It is not the different EXT two-stage script-function conversion.

All four client optional String null arguments remain null in bound parameters and actual DTO/resolver output; RestrictedSenderList uses an explicit literal-null invocation and the other three use compiled hashtable splatting. Empty and literal-space strings remain unchanged. Validation URI literal null remains null and serializes null; explicit empty reaches the actual URI setter's PropertyArgumentException, while HTTPS input succeeds. Client Nullable<Boolean> literal-null/literal-false and validation compiled-splat null/false remain null/false. The artifact records each binding_form; 13 final rows match the initial compiled-splat results after excluding that invocation-form annotation. This closes the public-binder versus nullable DTO distinction for these six commands rather than assuming model serialization alone settles argument conversion. Native early path safety and caller-controlled error/confirmation boundaries remain separately documented.

The 13-row artifact directly covers **client and validation**, not direct Privacy binder invocations. Privacy's own four Nullable<Boolean> effective parameter/model metadata (independent metadata check `37828`) and its actual DTO/resolver null/false/omission probe above form its direct source basis. The compiled Nullable<Boolean> binder rows are shared supporting behavior only. No direct Privacy literal-argument invocation is claimed by that artifact.

## Next finite review: Education accepted responses

The pinned `SetTeamsEducationConfigurationCmdlet` is another direct compiled
PolicyRp command. Fresh effective metadata resolves exactly three setting
parameters and Core DTO properties as `System.String`:
ParentGuardianPreferredContactMethod, EduGenerativeAIEnhancements and
UpdateParentInformation. Each has only ParameterAttribute; no extra enum,
nonempty or transformation attribute is present. Their actual DTO setters are
plain stores at IL2 followed by return7. The concrete AreParametersSpecified
checks BoundParameters.ContainsKey at IL16/39/62, so selected null and empty are
settings while omission is not. This uses the same verified Core/PolicyRp
hashes and portable metadata/IL reproduction above; it is not a server-value
eligibility claim or a substitute for literal public binder execution.

The existing retained selected/list read fixtures already reach typed Get
success. Controlled 400 request probes remain correctly labeled rejections;
prepared Update previously reached only read/patch failures. New
`TestUnitEducationSparsePatchAndPreparedUpdateAcceptedResponses` authors a
sparse typed PATCH200 reply and malformed string-property200, asserting distinct
route/body identity, null/empty/omitted fields, future/raw retention and
ErrDecode with the raw HTTP receipt. It separately replays the permanent Global
read before authored empty PATCH204 and checks Before/read receipt versus
HTTPAccepted, with no After or persistence claim. No tenant request, fixture
capture or ledger status promotion accompanies this added branch. Execution
will be bound to its committed source checkpoint before qualification decisions.

Full Education TestUnit/Example passed at clean `d34f149` (session `52700`),
Go 1.27.4, GOMAXPROCS=2, `-p 1 -count=1`, 14 GiB disk guard and
owned scratch directory removed after completion; admission 39.23 GiB. The first
sandbox attempt stopped at Go-cache access before any test ran; the authorized
cache-access run is the execution proof. No race or live test ran. This test
reference update leaves the command contract and every live status unchanged.

## Finite follow-up: Upgrade, SIP devices and MTO

These six Get/Set commands are direct compiled Core/PolicyRp operations with
exact Get Identity/Filter and Set Identity sets. Fresh effective metadata and
actual setters identify this complete write-field inventory:

| Family | All public setting/DTO types | Concrete AreParametersSpecified |
|---|---|---|
| TeamsUpgradeConfiguration | DownloadTeams and BlockLegacyAuthorization: Nullable<Boolean>; SfBMeetingJoinUx: String | ContainsKey at IL16/39/62 |
| TeamsSipDevicesConfiguration | BulkSignIn: String | ContainsKey16 |
| TeamsMultiTenantOrganizationConfiguration | CopilotFromHomeTenant: String | ContainsKey16 |

Every listed public setting has only ParameterAttribute and its DTO setter is
plain stfld2/ret7. There is no source enum, nonempty or cross-field validator in
these fields. Effective Filter nonempty attributes are part of the initial
14-compiled-get sweep. Identity binding versus native encoded-path safety and
early error ordering stay separate as for the first six decisions. No source
observation states that an arbitrary opaque string is accepted by a tenant.
The already recorded compiled String/Nullable<Boolean> binder observations are
supporting shared-type behavior; they did not directly invoke these three
families' binders. This table comes from their own concrete metadata/setters.
Bound-key guards and the actual shared resolver establish selected false/null/
empty versus omitted-property serialization; Force alone is not a setting.

Reproduce against the hash-checked module using Get-Command
`Get-CsTeamsUpgradeConfiguration`/`Set-CsTeamsUpgradeConfiguration` (substitute
TeamsSipDevicesConfiguration or TeamsMultiTenantOrganizationConfiguration),
inspect Parameters/ParameterSets and obtain the closed model from the Set
ImplementingType.BaseType generic argument. Use the portable IL reader on each
concrete `Set{type}Cmdlet.AreParametersSpecified` and Core `{type}.set_{field}`.
This reproduces offsets without depending on temporary host files. The shared
route/auth/header/list/filter/status/decoder/no-replay and resolver sections
above apply to the same exact compiled base/helper methods; none of the three
uses a conditional Trps dispatch. Minimum roles and live behavior are separate.

Existing full packages already contain retained selected/list/readback models,
successful filter operations, prepared reads followed by accepted empty204,
explicit null/empty/false/Force request cases, actual403 envelope retention,
zero-request context/destination/filter/identity guards and owned snapshots.
Upgrade additionally replays a caller-issued independent readback. SIP/MTO
prepare/acceptance tests do not claim independent effective readback. No new
permutation is added merely to count these existing success branches. Their
complete-package execution and individual command decisions are recorded below,
with live fixtures and effective behavior kept separate from contract proof.

The complete Upgrade/SIP/MTO package TestUnit/Example baseline passed at clean
`20490c0`, session `82938` (36.31 GiB admission, Go 1.27.4 / GOMAXPROCS 2,
`-p 1 -count=1`, owned scratch, 14 GiB guard). Review then found that their
existing local guard tests did not actually invoke either no-settings write
method. The bounded additions call both Patch and prepared Update with omitted,
false and true Force but no settings, asserting ErrInput, nil typed result and
nil response with zero HTTP. At clean
`d014e7827f31d57bb2d0be4ad1ae3adcdf0ae91c`, exact
TestUnitConfigurationInputAndContextRejectBeforeHTTP passed in all three
packages, session `65890` (36.80 GiB admission; same guard/owned scratch, removed
at completion). This supplemental execution covers only the changed local
cases; it is not mislabeled a second full-package run.

The null/empty SIP/MTO request combinations are primary-authored request
vectors paired with retained empty204 reply bytes; they are not captures of
those null/empty mutations succeeding in a tenant. Upgrade's503 request vectors
likewise establish selected serialization with controlled rejection, not
successful server eligibility for opaque strings. Actual accepted update
fixtures and independent readback snapshots remain separately identified.

### Eight individually reviewed command decisions

| Commands / exact sets | Complete reviewed behavior and execution | Remaining live boundary |
|---|---|---|
| Get/Set-CsTeamsEducationConfiguration / Get Identity+Filter, Set Identity | All three opaque strings, actual property setters, concrete bound-key control, inherited binder/Force/path and shared compiled route/result rules; retained get/list/filter, zero-request guards/no-settings, source-authored typed200/malformed200 and prepared204, plus later HTTP/preparation failures. Full package d34f149/52700. | Actual change/readback/oracle/restore, effective education behavior and minimum role remain unqualified. |
| Get/Set-CsTeamsUpgradeConfiguration / Get Identity+Filter, Set Identity | All three nullable settings, no extra validators, exact bound-key/no-settings semantics, optional collection/selected read and literal filtering; retained typed before/accepted204/caller readback and owned snapshots, false/null/opaque/Force rejected request proofs, real403/no-replay and shared source response/error rules. Full package 20490c0 / 82938; new no-settings cases d014e78 / 65890. | Each field's effective tenant behavior and full oracle/restore/permission/cloud qualification stay separate. |
| Get/Set-CsTeamsSipDevicesConfiguration / Get Identity+Filter, Set Identity | Complete single String BulkSignIn public/model inventory, null/empty/omitted distinctions, exact concrete guard, selected/list/filter, retained before/accepted204, typed snapshots, actual403/no-replay, source-authored null/empty/Force wire cases and zero-request guards. Full package 20490c0 / 82938; no-settings d014e78 / 65890. | No effective SIP-device behavior or broad live qualification is inferred from transport acceptance/replayed snapshots. |
| Get/Set-CsTeamsMultiTenantOrganizationConfiguration / Get Identity+Filter, Set Identity | Complete single String CopilotFromHomeTenant public/model inventory and same individually verified plain setter/bound-key rules; retained typed get/list/filter/prepared204 and snapshots, actual403/no-replay, authored null/empty/Force and no-settings guards. Full package 20490c0 / 82938; no-settings d014e78 / 65890. | Effective cross-tenant Copilot behavior, minimum role and all live qualification remain unchanged. |

This packet records eight reviewed implemented/contract-passed commands and
twelve exact sets. [Versioned execution records](../../../contracts/coverage/evidence/pol91-small-compiled-config.json) bind the decisions. It uses common-source evidence for unchanged shared model/status/filter/
control behavior, not a claim that each generic helper permutation was rerun in
every family. The earlier clean root 0df43de / 93343 compositions remain bounded to
client/privacy/validation; none is relabeled an Education/Upgrade/SIP/MTO
cross-family composition. Every live record and the other command states remain unchanged.
