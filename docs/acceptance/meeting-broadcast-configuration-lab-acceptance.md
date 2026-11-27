# Meeting broadcast configuration qualification

Owner: #127 (POL-54). Baseline: MicrosoftTeams 8.1.0.

The pinned public runtime metadata (`runtime.json` commands 133 and 605)
has Get-CsTeamsMeetingBroadcastConfiguration Identity/Filter and
Set-CsTeamsMeetingBroadcastConfiguration Identity. Get adds one switch to both
parameter sets, `ExposeSDNConfigurationJsonBlob`. Set binds
AllowSdnProviderForBroadcastMeeting (`System.Boolean`), the strings SupportURL,
SdnProviderName, SdnLicenseId, SdnApiTemplateUrl, SdnApiToken and
SdnRuntimeConfiguration, and an optional positional Identity. Both commands are
public proxy `Function`s whose `custom/CmdletConfig.json` entry
(`reconciliation.json` records 156 and 666) is `CmdletType: Remoting`,
`AutoRestModuleName: Microsoft.Teams.ConfigAPI.Cmdlets`,
`DefaultAutoRestParameters.ConfigType: TeamsMeetingBroadcastConfiguration` and
`ModernCmdlet: Get-CsConfigurationModern` / `Set-CsConfigurationModern`. This is
the dispatch already qualified for TeamsMigrationConfiguration, so the generated
ConfigAPI routes apply:

- GET `/Skype.Policy/configurations/TeamsMeetingBroadcastConfiguration` (bare list)
- GET `/Skype.Policy/configurations/TeamsMeetingBroadcastConfiguration/configuration/{Identity}`
- PUT the same instance route with a sparse body that always carries `Identity`;
  HTTP 204 with an empty body is the success receipt.

The pinned `Get-CsConfigurationModern` (`custom/Merged_custom_PsExt.ps1`) has one
special case for this type: when `PropertyBag.ExposeSDNConfigurationJsonBlob` is
true it prepends a pipeline step that appends `?ExposeSDNConfigurationJsonBlob=true`
to the request URI, then projects Identity, SupportURL,
AllowSdnProviderForBroadcastMeeting, SdnName, SdnLicenseId,
SdnAzureSubscriptionId, SdnApiTemplateUrl, SdnApiToken, SdnRuntimeconfiguration
and SdnAttendeeFallbackCount. The Go `ReadOptions.ExposeSDNConfigurationJsonBlob`
sends the same query on GetByIdentity, List and ListFiltered and decodes the same
fields; RawFields keeps the full response. The service names the provider
`SdnName` on read while Set binds `SdnProviderName`, and spells the read field
`SdnRuntimeconfiguration`; the SDK keeps both wire spellings. Set defaults
Identity to Global and exposes no Force; Filter is applied locally to Identity and
`Tag:`Identity with wildcard matching. The compiled PolicyRP PATCH provider is
not used.

SdnApiToken is a provider credential that may also appear in exposed sensitive
reads. `UpdateRequest` and `Configuration` never format it, the transport does
not log bodies, the oracle
reports only whether the exposed token is non-empty, the recovery journal never
stores SDN values, and fixtures redact token-named fields. The lab has no SDN
provider configured; the exposed token was empty in every read.

Oracle checks pin these module assembly SHA256 values:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`
- ConfigAPI: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

On 2026-11-23 the dedicated SDK service principal (Teams Administrator) read and
listed Global, with and without the exposure switch, using client-secret and
certificate authentication. Both typed projections matched pinned PowerShell
8.1.0 (`Get-CsTeamsMeetingBroadcastConfiguration -Identity Global` and the same
call with `-ExposeSDNConfigurationJsonBlob`). The lab's original Global had Host
authority, AllowSdnProviderForBroadcastMeeting false, a non-empty SupportURL and
empty SDN provider settings.

A first update attempt that set AllowSdnProviderForBroadcastMeeting to true was
rejected by the service with HTTP 400, errorCode 40013 and inner error
`SkypecastSdnMissingParameter` naming `SdnProviderName`; nothing was applied. That
sanitized response is kept as `update_allow_without_provider_400.json` and is the
evidence for the service-side cross-field rule. The SDK does not duplicate it
locally. The lab has no SDN provider to configure, so enabling the provider is
not live-qualified.

The update test therefore binds only SupportURL. It refuses missing originals,
persists the original SupportURL and Boolean in a bound private recovery journal,
sets SupportURL to a reserved `example.com` address through `UpdateGlobal`, waits
for persisted read-back, verifies that the unbound Boolean and every exposed SDN
setting are unchanged, compares the pinned getters, restores the originals
through the journal handler, then independently reads them back and compares
again. Both application auth modes passed; each run left one
`meeting-broadcast-configuration` journal entry in the cleaned state. After
restoration the Global values equal the original, but authority is Tenant rather
than Host, as for every other singleton written by the service. No licenses,
roles, users, groups or policy assignments were changed.

CI runs these as the `teams_meeting_broadcast_configuration` partition: a
serialized `global-reads` step and an `update-restore` step under the tenant
lock, requiring three pinned comparisons, the update and restoration assertions
and one cleaned restore entry. The partition uses its own read-only oracle
script, `read-meeting-broadcast-configuration.ps1`, so it does not touch the
shared configuration oracle or lab kit.

Seven permanent sanitized response files come directly from these acceptance
reads, the rejected update, the update and the read-backs; the real 204 body is
empty. Unit tests use them for native routing, the exposure query on both
parameter sets, headers, typed getters, snapshot ownership, filter selection,
sparse PUT bodies (unbound, false, empty string and every setting), token
redaction, identity segment escaping and the service validation error. Common
denial propagation reuses the actual privacy-service 403 envelope and is not a
live denial claim for this service. Deliberately malformed replies test decoding
failures separately from successful fixtures.

Still unqualified: enabling AllowSdnProviderForBroadcastMeeting and writing
SdnProviderName, SdnLicenseId, SdnApiTemplateUrl, SdnApiToken and
SdnRuntimeConfiguration (they need a real SDN provider and credential, which the
lab does not have and which must not be overwritten), non-Global identities,
effective live event behavior, delegated authentication and other clouds. No
create/delete methods are invented; none are exported by the pinned module.
