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
