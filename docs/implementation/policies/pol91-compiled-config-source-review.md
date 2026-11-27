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

The initial audit records exact four-package proofs at 42fd88e (87948/29006), clean root0df43de/93343 for success/denial/malformed/cancellation/recovery and distinct retained versus authored bytes. New review identified two bounded evidence gaps: privacy's duplicate filter test never exercised successful filtering/actual whitespace, and voicemail Set never reached an accepted submission. The pending source-authored replacements close those branches without fabricating a tenant capture or nondefault privacy scope. Client's new typed PATCH200/malformed200 full-package result is separately bound to 01e469b/33122; its initial outer-whitespace test assertion failure is a corrected test oracle, not a production regression. Final new privacy/validation execution and individual status decisions are recorded only after their source checkpoint passes.

Live authorization/minimum role, actual tenant change/readback/oracle/restore and client effects remain unqualified. Nothing in this source review promotes live status or closes #185.
