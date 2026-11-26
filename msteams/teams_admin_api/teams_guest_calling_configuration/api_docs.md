# Teams guest calling configuration

Use `sdk.TeamsAdminAPI.TeamsGuestCallingConfigurations` for the public 8.1.0
Get-CsTeamsGuestCallingConfiguration and Set-CsTeamsGuestCallingConfiguration
functions. Both are `Remoting` proxies whose pinned `CmdletConfig.json` selects
`Get-CsConfigurationModern`/`Set-CsConfigurationModern` with
`ConfigType=TeamsGuestCallingConfiguration`, i.e. modern ConfigAPI remoting.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; one escaped configuration-name segment |
| `List` | Omitted/blank Get Identity; bare list, case-insensitive Global first |
| `ListFiltered` | Get Filter; local Identity/Tag wildcard filter; blank selects all |
| `UpdateByIdentity` | Set Identity; sparse PUT, explicit Identity, requires HTTP 204 |
| `UpdateGlobal` | Explicitly applies Set's omitted-Identity Global default |

Routes are `/Skype.Policy/configurations/TeamsGuestCallingConfiguration` and
`.../configuration/{Identity}`. The shared modern adapter performs tenant service
discovery and supplies target routing. Requests use `Accept: application/json`
and `X-MS-CmdletName` without `MPACmdlet`. This service does not use the compiled
PATCH provider. Context/cancellation and invalid input checks precede credential
acquisition.

`UpdateRequest.AllowPrivateCalling` is `*bool`: nil omits the property,
false/true bind it. The public proxy declares a non-nullable Boolean, so no JSON
null request is introduced. Identity-only Set is permitted by the pinned wrapper;
there is no implicit preparation read or Force setting. PUT never replays.
`Configuration` preserves the nullable read value, unknown `RawFields` and an
owned `RawJSON` snapshot. There are no public create/delete commands for this
singleton.

Certificate and client-secret lab runs using Teams Administrator read/list Global,
compare its getter with PowerShell 8.1.0, toggle AllowPrivateCalling, verify
read-back, restore the original value and independently verify restoration.
CI requires three pinned comparisons and a cleaned journal entry. Global
read/update tests share the tenant lock. Permanent sanitized acceptance responses
provide unit decoding and request fixtures. These checks do not establish
effective guest private-calling behavior, delegated authentication or other clouds.

See [qualification](../../../docs/implementation/policies/guest-calling-configuration-lab-acceptance.md).
