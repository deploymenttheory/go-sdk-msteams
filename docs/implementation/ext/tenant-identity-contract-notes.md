# Tenant identity contracts

Owner: #91 (EXT-29). Research: EXT-03. Baseline: MicrosoftTeams 8.1.0.

Package: `msteams/teams_admin_api/tenant_identity`, wired as
`sdk.TeamsAdminAPI.TenantIdentity` on the ConfigAPI backend.

## Dispatch evidence

| Cmdlet | CmdletConfig / export | Route (ConfigApiBasedCmdlets IL) |
| --- | --- | --- |
| Find-CsGroup | AutoRest, GA | GET `/Teams.VoiceApps/groups?SearchQuery=&MaxResults=&ExactMatchOnly=&MailEnabledOnly=`. Each bound value goes through `Uri.EscapeDataString`; unbound values are omitted. Only 200 is declared. |
| Get-CsTenant | Remoting → `Get-CsTenantPoint` → internal `Get-CsTenantObou -DefaultPropertySet Extended` | GET `/Teams.Tenant/tenants?defaultpropertyset=Extended` (200; `$select` unbound) |
| Get-CsOnlineDirectoryTenant | Remoting → `Invoke-CsDeprecatedError` | none |
| Get-CsTenantLicensingConfiguration | Remoting → `Get-CsConfigurationModern {ConfigType: TenantLicensingConfiguration}`, GA | GET `/Skype.Policy/configurations/TenantLicensingConfiguration[/configuration/{identity}]` |
| Get-CsHostingProvider | Remoting → `Get-CsConfigurationModern {ConfigType: HostingProvider}`, SelfHost/Torus | GET `/Skype.Policy/configurations/HostingProvider[/configuration/{identity}]` |
| Get-CsAadTenant | AutoRest, SelfHost/Torus | GET `/Teams.Tenant/AadTenant/` (200) |
| Get-CsAadUser | AutoRest, SelfHost/Torus | GET `/Teams.User/AadUser/{EscapeDataString(identity)}` (200) |
| Get-CsCloudTenant | AutoRest, SelfHost/Torus | GET `/Teams.InternalSupport/GetAADTenantData/` (200) |
| Get-CsCloudUser | AutoRest, SelfHost/Torus | GET `/Teams.InternalSupport/GetAADUserData/?userId=` (200) |
| Remove-CsUserLicenseGracePeriod | AutoRest, GA | PATCH `/Teams.User/users/{EscapeDataString(identity)}/AssignedPlans`. Only 204 (`onNoContent`) is declared. |

Notes:

- **Find-CsGroup.** The custom wrapper (Merged_custom_PsExt.ps1) removes
  `Force`. It writes `Diagnostic` to the admin diagnostic stream and projects
  each `Group` (`Id`, `DisplayName`) into `GroupModel`. The service rejects
  queries shorter than three characters with HTTP 400 (`code` "100001"). The
  module's booleans bind as `System.Boolean` and are spelled `True`/`False`.
  The SDK encodes the query with Go's form encoding: spaces become `+`, where
  the module sends `%20`. The service decodes both the same way; the lab search
  for a name containing spaces matched.
- **Get-CsTenant.** The public proxy declares `Filter`, `Identity`, `ResultSize`
  and `MsftInternalProcessingMode`. `Get-CsTenantPoint` declares none of them,
  so PowerShell fails each with `NamedParameterNotFound` (captured live). The
  SDK therefore exposes only the parameterless read. The PowerShell projection
  renames service properties (`objectId` → `TenantId`, `sipDomains` →
  `SipDomain`, `assignedPlans` → `AssignedPlan`) and flattens
  `lastProvisionTimeStamps`/`lastPublishTimeStamps`. The SDK keeps service
  names and `RawFields`.
- **Get-CsOnlineDirectoryTenant.** It always writes "This cmdlet is no longer
  supported, please consult public documentation or use Get-CsTenant and
  Get-CsOnlineDialinConferencingBridge to get the relevant attributes." The SDK
  returns `ErrOnlineDirectoryTenantDeprecated` with no I/O.
- **Configuration reads.** The `-Filter` parameter set is the shared local
  `Identity`/`Tag:Identity` wildcard match over the full list, Global first, as
  in application_meeting_configuration. Get-CsHostingProvider's `LocalStore`
  switch has no service behaviour.
- **Support-only routes.** The module's generated paths for the four support
  routes end in `/`. `client.Request.Path` cannot carry an empty trailing
  segment, so the SDK sends them without it. Live, app-only, without the slash,
  all four return 403. With it, AadTenant and AadUser return 403, and the two
  InternalSupport routes return 400 "This legacy OCE cmdlet is no more
  supported." Either way, no app-only caller can read them. The module outputs
  each response's `AdditionalProperties` with no fixed model, so the SDK types
  `objectId`/`displayName` only.
- **Grace period removal body.** The `UserDelicensingAccelerationPatch.ToJson`
  output, serialized through the module, is
  `{"capabilities":[…],"action":"…"}`. Unset properties are omitted, and
  `Capability` allows an empty collection. SupportsShouldProcess and PassThru
  are PowerShell-only.

## Lab observations (client-secret, app-only Teams Administrator)

`TestAccTenantIdentityReads` (read-only step `tenant_identity/reads`):

- `GetTenant` returns 200 with the extended record bound to the lab tenant.
- `SearchGroups` finds exactly the owned acceptance policy group with
  `ExactMatchOnly=True`. An unmatched query returns `{"Groups":[],…}`, and a
  two-character query returns 400.
- `GetAadTenant`, `GetAadUser`, `GetCloudTenant` and `GetCloudUser` (reserved
  probe ID) return 403 `Forbidden`.
- `ListHostingProviders`, `ListTenantLicensingConfigurations` and
  `GetTenantLicensingConfigurationByIdentity("Global")` return 403 `Forbidden`
  (errorCode 40301).
- Pinned PowerShell oracle (`read-tenant-identity.ps1`) compares the tenant ID,
  the verified domain, SIP domain and assigned plan counts, and the owned group
  ID. It also checks the Get-CsOnlineDirectoryTenant deprecation message and
  the Forbidden outcome of Get-CsTenantLicensingConfiguration. Parity passed.
  The forbidden configuration cmdlet writes a correlation ID line straight to
  the console, so the oracle redirects `[Console]::Out` around that call.

The sanitized responses feed the unit contract tests. labkit fixture redaction
now also replaces directory address, contact and naming fields (city, street,
postal code, state, telephone, UPN, mail nickname and similar).

## Unqualified

- **Remove-CsUserLicenseGracePeriod.** Never sent live: it changes licensing
  state and cannot be undone. Only a synthetic wire contract covers it.
- **Successful support-only and configuration reads.** Get-CsAadTenant/User,
  Get-CsCloudTenant/User, Get-CsHostingProvider and
  Get-CsTenantLicensingConfiguration would need credentials that the service
  authorizes. Their success shapes are synthetic.
- **Find-CsGroup `MailEnabledOnly` filtering, and delegated callers generally.**

## EXT-90 finite Find-CsGroup repair

This follow-up concerns only Find-CsGroup's one __AllParameterSets record;
other tenant/support/configuration/licensing contracts are unchanged. The
[portable source packet](evidence/find-group-source.json) binds the verified
8.1.0 default export, custom/internal scripts, own private header/pipeline and
query/status/callback bodies, generated response/Group DTO and real public
GroupModel.ParseFrom. Its runtime Guid helpers are separately hashed because
that actual public projector calls Guid.Parse; no generic UUID parser or GUID
restriction on unrelated directory inputs is inferred.

Read-only actual observations cover 31 model cases, 11 envelopes and 21
PowerShell binder cases across three phases in an empty runspace. Model parsing and
GroupModel.ParseFrom are actual pinned data operations; the binder proves
type conversion/presence, not a management request. The final phase executes
the actual pinned internal proxy with only its private target replaced by an
inert compiled cmdlet. No MicrosoftTeams compiled management lifecycle,
auth/HTTP or tenant operation is invoked. Embedded reproducers and all raw
hashes are retained, with 81 relevant IL bodies and no unresolved metadata
operands across the three selected method groups.

The public proxy/custom wrapper takes UInt32? MaxResults, then forwards to an
Int32 internal parameter. Omission stays unbound; explicit zero survives. The actual internal script
proxy converts bound null to bound zero, whereas a direct compiled Int32 binder
rejects null. Both phases are retained rather than assuming they are equivalent.
Values above MaxInt32 fail the internal script transformation. The unchanged
native uint32 pointer rejects that range early with ErrInput; nil represents
omission and non-nil zero represents the source bound-null wire result. Required
empty search is rejected locally,
whereas query whitespace/control characters are safely encoded. The source
header is Find-CsGroup_Find, status exactly 200. Force is removed and no
WhatIf/Confirm is declared. Native early errors/context/transport controls do
not claim identical PowerShell telemetry or session lifecycle stages.

Generated fields use exact names/types and recursive exact-duplicate rejection.
Missing/null/wrong-type Groups in an object yields empty public output. An
empty/whitespace/array root yields null generated output; a literal null scalar
fails. A null row or failed public GUID conversion fails the buffered group
result, never a partial/negative match. Public IDs are canonical D-format GUID
strings; original wire fields remain independent in RawFields. Source Guid
N/D/B/P/X forms, D prefix/plus compatibility, X UInt32-to-UInt16 conversion and
strict byte/first-field ranges are source-backed. A bounded D trailing-NUL
compatibility case is captured rather than silently rejected. DisplayName
wrong types project null. Raw-field enrichment and native standard-JSON/error
and diagnostic-stream boundaries remain explicit.

New real-client offline cases cover retained 200 receipts on failed/null/empty
projection, no automatic second request, typed versus original raw identities,
encoded query/zero/false/max boundaries and undeclared 201/204. Historical
sanitized live response fixtures stay response-only evidence for these corrected
source-derived requests. Full affected execution and independent review are
required before any Find-CsGroup implementation/contract decision; all live
states remain unchanged and #191 remains open.
