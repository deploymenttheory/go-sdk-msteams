# Teams meeting broadcast configuration

Use `sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations` for the public 8.1.0
Get-CsTeamsMeetingBroadcastConfiguration and
Set-CsTeamsMeetingBroadcastConfiguration functions. Both are `Remoting` proxies
whose pinned `CmdletConfig.json` selects
`Get-CsConfigurationModern`/`Set-CsConfigurationModern` with
`ConfigType=TeamsMeetingBroadcastConfiguration`, i.e. modern ConfigAPI remoting.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; one escaped configuration-name segment |
| `List` | Omitted/blank Get Identity; bare list, case-insensitive Global first |
| `ListFiltered` | Get Filter; local Identity/Tag wildcard filter; blank selects all |
| `ReadOptions.ExposeSDNConfigurationJsonBlob` | Get switch (both sets); appends `?ExposeSDNConfigurationJsonBlob=true` |
| `UpdateByIdentity` | Set Identity; sparse PUT, explicit Identity, requires HTTP 204 |
| `UpdateGlobal` | Explicitly applies Set's omitted-Identity Global default |

Routes are `/Skype.Policy/configurations/TeamsMeetingBroadcastConfiguration` and
`.../configuration/{Identity}`. The shared modern adapter performs tenant service
discovery and supplies target routing. Requests use `Accept: application/json`
and `X-MS-CmdletName` without `MPACmdlet`. This service does not use the compiled
PATCH provider. Context/cancellation and invalid input checks precede credential
acquisition.

Default reads return Identity, SupportURL and AllowSdnProviderForBroadcastMeeting.
With the exposure switch the service also returns the SDN provider fields
`SdnName`, `SdnLicenseId`, `SdnAzureSubscriptionId`, `SdnApiTemplateUrl`,
`SdnApiToken`, `SdnRuntimeconfiguration` and `SdnAttendeeFallbackCount`. Note that
Set binds the provider as `SdnProviderName` and the runtime blob as
`SdnRuntimeConfiguration`; the SDK sends and decodes each exact wire spelling.

Every `UpdateRequest` field is a pointer: nil omits the property, and a non-nil
value (false or an empty string included) binds it. The public proxy declares a
non-nullable Boolean and plain strings, so no JSON null is sent. Identity-only Set
is permitted by the pinned wrapper; there is no implicit preparation read or Force
setting. PUT never replays. The service rejects enabling
AllowSdnProviderForBroadcastMeeting without `SdnProviderName` (HTTP 400, errorCode
40013, `SkypecastSdnMissingParameter`); the SDK returns that `client.APIError`
unchanged rather than duplicating the rule. `Configuration` preserves nullable
read values, unknown `RawFields` and an owned `RawJSON` snapshot. There are no
public create/delete commands for this singleton.

`SdnAPIToken` is a provider credential. `UpdateRequest` and `Configuration` never
print it through `String`/`GoString`/`fmt`, and the transport never logs bodies.
An exposed read returns it in `SdnAPIToken`, `RawFields` and `RawJSON`, so callers
must not log those.

Certificate and client-secret lab runs using Teams Administrator read/list Global
with and without the exposure switch, compare both getters with PowerShell 8.1.0,
update SupportURL, verify read-back and that the unbound Boolean and SDN settings
are unchanged, restore the original value and independently verify restoration.
CI requires three pinned comparisons and a cleaned journal entry. Global
read/update tests share the tenant lock. Permanent sanitized acceptance responses
provide unit decoding and request fixtures. Enabling the SDN provider, writing SDN
provider settings or tokens, effective live event behavior, delegated
authentication and other clouds are not qualified.

See [qualification](../../../docs/implementation/policies/meeting-broadcast-configuration-lab-acceptance.md).
