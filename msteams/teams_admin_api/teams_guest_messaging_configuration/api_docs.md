# Teams guest messaging configuration

Use `sdk.TeamsAdminAPI.TeamsGuestMessagingConfigurations` for the public 8.1.0
Get-CsTeamsGuestMessagingConfiguration and Set-CsTeamsGuestMessagingConfiguration
functions. Both are `Remoting` proxies whose pinned `CmdletConfig.json` selects
`Get-CsConfigurationModern`/`Set-CsConfigurationModern` with
`ConfigType=TeamsGuestMessagingConfiguration`, i.e. modern ConfigAPI remoting.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; one escaped configuration-name segment |
| `List` | Omitted/blank Get Identity; bare list, case-insensitive Global first |
| `ListFiltered` | Get Filter; local Identity/Tag wildcard filter; blank selects all |
| `UpdateByIdentity` | Set Identity; sparse PUT, explicit Identity, requires HTTP 204 |
| `UpdateGlobal` | Explicitly applies Set's omitted-Identity Global default |

Routes are `/Skype.Policy/configurations/TeamsGuestMessagingConfiguration` and
`.../configuration/{Identity}`. The shared modern adapter performs tenant service
discovery and supplies target routing. Requests use `Accept: application/json`
and `X-MS-CmdletName` without `MPACmdlet`. This service does not use the compiled
PATCH provider. Context/cancellation and invalid input checks precede credential
acquisition.

`UpdateRequest` has one field per public Set setting: the eight Booleans
AllowUserEditMessage, AllowUserDeleteMessage, AllowUserDeleteChat, AllowUserChat,
AllowGiphy, AllowMemes, AllowImmersiveReader and AllowStickers (`*bool`), and
GiphyRatingType (`*string`). nil omits the property; false, true and the empty
string are sent as bound values. The public proxy declares non-nullable Boolean
and String parameters, so no JSON null request is introduced. GiphyRatingType has
no pinned validation attribute; the help names Moderate (default), Strict and
NoRestriction, and the SDK sends the value verbatim rather than inventing a local
enum. Identity-only Set is permitted by the pinned wrapper; there is no implicit
preparation read or Force setting. PUT never replays.

`Configuration` types every getter setting as a nullable value, including the
read-only UsersCanDeleteBotMessages that the service returns but Set cannot bind.
It also keeps unknown `RawFields` and an owned `RawJSON` snapshot. There are no
public create/delete commands for this singleton.

Certificate and client-secret lab runs using Teams Administrator read/list Global,
compare all ten getter settings with PowerShell 8.1.0, change AllowMemes and
GiphyRatingType, verify read-back and that the other settings were untouched,
restore the originals and independently verify restoration. CI requires three
pinned comparisons and a cleaned journal entry. Global read/update tests share
the tenant lock. Permanent sanitized acceptance responses provide unit decoding
and request fixtures. These checks do not establish effective guest chat
behavior, delegated authentication or other clouds.

See [qualification](../../../docs/acceptance/guest-messaging-configuration-lab-acceptance.md).
