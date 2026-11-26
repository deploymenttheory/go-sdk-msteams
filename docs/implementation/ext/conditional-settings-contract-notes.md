# Conditional tenant and app configuration contracts

Owner: EXT-90 (#191). Baseline: MicrosoftTeams 8.1.0.

These four commands occur in the pinned OCE manifest, with separate compiled
`CmdletInfo`/`CommandMetadata` declarations. They are absent from the immutable
default-session root-export snapshot; that absence is not an exclusion.

| Command | Resource | Declared sets | SDK operations |
| --- | --- | --- | --- |
| Get-CsTenantConfiguration | TenantConfiguration | Identity, Filter | GetByIdentity, List, ListFiltered |
| Set-CsTenantConfiguration | TenantConfiguration | Identity | PatchByIdentity, UpdateByIdentity |
| Get-CsTeamsAppPolicyConfiguration | TeamsAppPolicyConfiguration | Identity, Filter | GetByIdentity, List, ListFiltered |
| Set-CsTeamsAppPolicyConfiguration | TeamsAppPolicyConfiguration | Identity | PatchByIdentity, UpdateByIdentity |

The [IL observation](evidence/conditional-settings-il.json) records exact methods,
member references and offsets. Model assembly
`Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll` SHA256 is
`68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`;
provider `Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll`
SHA256 is `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`.

`ConfigurationNameNormalization<TPolicy>.GetNormalizedNameForPolicyRP` starts
with the model type name. Its static exception list contains only compliance
recording application types and NetworkSiteType, not either model here.
`HttpRequestHelper<TPolicy>.GetAsync` builds
`/Skype.Policy/configurations/{name}/configuration/{identity}`. `SetAsync`
uses PATCH and `BoundParametersContractResolver`; get-all uses the collection.
The SDK reuses the existing direct compiled PolicyRP transport and local filter
implementation, rather than generated-client forest discovery. Both Filter
properties explicitly declare `ValidateNotNullOrEmptyAttribute`: omission uses
List, explicit empty is rejected, and whitespace remains a literal pattern.
The separate metadata snapshot now includes every raw declared attribute; the
older ValidateSet-only view could not establish this restriction.

`SetTenantConfigurationCmdlet.AreParametersSpecified` considers only the two
limit settings; `UpdateProperties` copies only bound members. The same compiled
provider pattern applies to the two app-configuration URI settings. Native
`PatchByIdentity` exposes one sparse request. `UpdateByIdentity` models the
preparation GET and sparse PATCH, retaining its snapshot, response and acceptance
stages when preparation, decoding or cancellation fails. Force alone does not
select a setting; explicitly bound false remains the provider's `"False"` value.
HTTP acceptance does not establish persisted/effective configuration.

## Model validation and presence

The [offline setter/serializer observation](evidence/conditional-settings-serialization.json)
is reproduced by
[`inspect-conditional-ext-serialization.ps1`](../../../scripts/discovery/inspect-conditional-ext-serialization.ps1).
It constructs data models only, verifies model/provider hashes and serializes
through the real bound-parameter resolver. It never constructs or invokes a
cmdlet, signs in or contacts a tenant. These are model/request observations,
not successful tenant replies.

- Tenant limits are nullable UInt32 with minimum 1. Maximum UInt32, 1 and explicit
  CLR null are accepted; zero is rejected. Populated unbound members are omitted.
- URI fields admit CLR null but reject empty strings and non-HTTPS prefixes.
  Their CultureInvariant expression is
  `^(h|H)(t|T)(t|T)(p|P)(s|S)://.+$`. It checks a prefix/content pattern, not URL
  structure, and .NET's end anchor admits one terminal LF. The SDK preserves
  that exact scalar behavior; it never uses these data values as destinations.
- Direct PowerShell string-property assignment can coerce null to empty before
  a model setter. The observation uses `PropertyInfo.SetValue` to preserve actual
  CLR null. Native explicit JSON null is not claimed as identical to every
  PowerShell parameter-binding path.

Request settings distinguish omission, selected values and null. Response
models retain typed nullable fields plus owned complete raw snapshots. Known
range/type/URI violations fail decoding; unknown future fields and null/empty
presence survive snapshot round trips.

Colocated unit vectors cover identity/filter projections, sparse omission,
explicit null/Force false, invalid inputs before credentials, uint32 boundaries,
malformed successful replies and partial acceptance. The integration cancellation
case retains preparation and no-replay state. Authored replies remain model
vectors; no successful live OCE response has been captured.

There is no create/delete/grant surface in these four declarations, and none is
invented. OCE export availability, authorization and actual server acceptance
remain unqualified. No live Global configuration mutation or role/license change
is configured. The coverage ledger stays in-progress/metadata-only/not-run.
