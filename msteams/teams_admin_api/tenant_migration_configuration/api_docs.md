# Tenant migration configuration

Use `sdk.TeamsAdminAPI.TenantMigrationConfigurations` for the public 8.1.0
Get-CsTenantMigrationConfiguration and Set-CsTenantMigrationConfiguration
functions, which read and set whether the Meeting Migration Service (MMS) is
enabled. Both are `Remoting` proxies whose pinned `CmdletConfig.json` selects
`Get-CsConfigurationModern`/`Set-CsConfigurationModern` with
`ConfigType=TenantMigrationConfiguration`, i.e. modern ConfigAPI remoting.

This is not the Teams client interop setting managed by
`TeamsMigrationConfigurations` (Get/Set-CsTeamsMigrationConfiguration,
`EnableLegacyClientInterop`); the two commands use different ConfigTypes and
settings.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; one escaped configuration-name segment |
| `List` | Omitted/blank Get Identity; bare list, case-insensitive Global first |
| `ListFiltered` | Get Filter; local Identity/Tag wildcard filter; blank selects all |
| `UpdateByIdentity` | Set Identity; sparse PUT, explicit Identity, requires HTTP 204 |
| `UpdateGlobal` | Explicitly applies Set's omitted-Identity Global default |

Routes are `/Skype.Policy/configurations/TenantMigrationConfiguration` and
`.../configuration/{Identity}`. The shared modern adapter performs tenant service
discovery and supplies target routing. Requests use `Accept: application/json`
and `X-MS-CmdletName` without `MPACmdlet`. This service does not use the compiled
PATCH provider. Context/cancellation and invalid input checks precede credential
acquisition.

`UpdateRequest` has the one public Set setting, MeetingMigrationEnabled
(`*bool`). nil omits the property; false and true are sent as bound values. The
public proxy declares a non-nullable Boolean, so no JSON null request is
introduced. Identity-only Set is permitted by the pinned wrapper; there is no
implicit preparation read or Force setting. The help's Instance and Tenant
parameters are not bound by the public 8.1.0 runtime and are not modeled; the
internal `MsftInternalProcessingMode` switch is not a wire setting. PUT never
replays.

`Configuration` types Identity and MeetingMigrationEnabled as nullable values and
keeps unknown `RawFields` and an owned `RawJSON` snapshot. There are no public
create/delete commands for this singleton.

Certificate and client-secret lab runs using Teams Administrator read/list Global,
compare the getter with PowerShell 8.1.0, disable MMS from an enabled original,
verify read-back, re-enable it and independently verify restoration. CI requires
three pinned comparisons and a cleaned journal entry. Global read/update tests
share the tenant lock. Permanent sanitized acceptance responses provide unit
decoding and request fixtures. These checks do not establish effective meeting
migration behavior, delegated authentication or other clouds.

See [qualification](../../../docs/acceptance/tenant-migration-configuration-lab-acceptance.md).
