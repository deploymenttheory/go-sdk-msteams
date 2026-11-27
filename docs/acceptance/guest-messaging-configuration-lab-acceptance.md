# Guest messaging configuration qualification

Owner: #124 (POL-51). Baseline: MicrosoftTeams 8.1.0.

The pinned public runtime metadata (`runtime.json` commands 128 and 601) has
Get-CsTeamsGuestMessagingConfiguration Identity/Filter and
Set-CsTeamsGuestMessagingConfiguration Identity. Set binds nine settings with no
validation attributes: eight non-nullable `System.Boolean` values
(AllowUserEditMessage, AllowUserDeleteMessage, AllowUserDeleteChat,
AllowUserChat, AllowGiphy, AllowMemes, AllowImmersiveReader, AllowStickers) and
the `System.String` GiphyRatingType; Identity is an optional positional string.
Both commands are public proxy `Function`s whose `custom/CmdletConfig.json`
entry (`reconciliation.json` records 151 and 662) is `CmdletType: Remoting`,
`AutoRestModuleName: Microsoft.Teams.ConfigAPI.Cmdlets`,
`DefaultAutoRestParameters.ConfigType: TeamsGuestMessagingConfiguration` and
`ModernCmdlet: Get-CsConfigurationModern` / `Set-CsConfigurationModern`. This is
the same dispatch already qualified for TeamsMigrationConfiguration,
TeamsShiftsAppPolicy and TeamsGuestCallingConfiguration, so the generated
ConfigAPI routes apply:

- GET `/Skype.Policy/configurations/TeamsGuestMessagingConfiguration` (bare list)
- GET `/Skype.Policy/configurations/TeamsGuestMessagingConfiguration/configuration/{Identity}`
- PUT the same instance route with a sparse body that always carries `Identity`;
  HTTP 204 with an empty body is the success receipt.

The modern wrapper defaults Set Identity to Global and exposes no Force; Filter
is applied locally to Identity and `Tag:`Identity with wildcard matching. The
Go service does not use the compiled PolicyRP PATCH provider and no
policyrp registration is involved. The 8.1.0 help describes GiphyRatingType as
"Moderate" by default with "NoRestriction" and "Strict" as other options; with
no runtime validation set, the SDK sends the string verbatim. The service also
returns UsersCanDeleteBotMessages, which the pinned getter exposes but Set cannot
bind; the SDK types it as a read-only field. Oracle checks pin these module
assembly SHA256 values:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`
- ConfigAPI: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

On 2026-11-23 the dedicated SDK service principal (Teams Administrator) read and
listed Global with both client-secret and certificate authentication, and all ten
typed getter settings matched pinned PowerShell 8.1.0. The lab's original Global
had Host authority, AllowMemes true and GiphyRatingType Moderate.

The update test refuses an incomplete original or an undocumented rating,
persists the original AllowMemes and GiphyRatingType in a bound private recovery
journal, toggles AllowMemes and switches the rating (Moderate to Strict here)
through one sparse `UpdateGlobal`, waits for persisted read-back, checks that the
eight unbound getter settings are unchanged and compares the pinned getter. It
then restores the originals through the journal handler, independently reads
them back, checks every setting equals the original and compares again. Both
application auth modes passed; each run left exactly one
`guest-messaging-configuration` journal entry in the cleaned state. After
restoration the Global values equal the originals, but the authority is Tenant
rather than Host, as for every other singleton written by the service. No
licenses, roles, users, groups or policy assignments were changed.

CI runs these as the `teams_guest_messaging_configuration` partition: a
serialized `global-reads` step and an `update-restore` step under the tenant
lock, requiring three pinned comparisons, the update and restoration assertions
and one cleaned restore entry. The partition uses its own read-only oracle
script, `read-guest-messaging-configuration.ps1`, so it does not touch the shared
configuration oracle or lab kit.

Five permanent sanitized response files come directly from these acceptance
reads, the update and the read-backs; the real 204 body is empty. Unit tests use
them for native routing, headers, typed getters, snapshot ownership, filter
selection, sparse PUT bodies (unbound, false/true, empty string, the acceptance
pair and all nine settings) and identity segment escaping. Common denial
propagation reuses the actual privacy-service 403 envelope and is not a live
denial claim for this service. Deliberately malformed replies test decoding
failures separately from successful fixtures.

Still unqualified: the seven other Booleans and the NoRestriction rating that the lab
did not write (their request shapes are unit-tested only), non-Global
identities (the service documents Global only), effective guest chat behavior
and propagation, delegated authentication and other clouds. No create/delete
methods are invented; none are exported by the pinned module.
