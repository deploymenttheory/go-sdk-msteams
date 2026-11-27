# Teams migration configuration

Use `sdk.TeamsAdminAPI.TeamsMigrationConfigurations` for the public 8.1.0
Get and Set functions. These functions select modern ConfigAPI remoting even
though the package also contains older compiled cmdlet classes.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; one escaped configuration-name segment |
| `List` | Omitted/blank Get Identity; bare list, case-insensitive Global first |
| `ListFiltered` | Local Identity/Tag wildcard filter; blank selects all |
| `UpdateByIdentity` | Sparse PUT, explicit Identity, requires HTTP 204 |
| `UpdateGlobal` | Explicitly applies Set's omitted-Identity Global default |

The shared modern adapter performs tenant service discovery and supplies target
routing. Requests use `Accept: application/json` and `X-MS-CmdletName` without
`MPACmdlet`. This service does not use the compiled PATCH provider. The whole
configuration name is escaped as one route segment. Context/cancellation and
invalid input checks precede credential acquisition.

`UpdateRequest.EnableLegacyClientInterop` is `*bool`: nil omits the property,
false/true bind it. The public proxy declares a non-nullable Boolean, so no JSON
null request is introduced. Identity-only Set is permitted by the pinned wrapper;
there is no implicit preparation read or Force setting. PUT never replays.
`Configuration` preserves nullable read values, unknown `RawFields` and an owned
`RawJSON` snapshot. There are no public create/delete commands for this singleton.

Certificate and client-secret lab runs using Teams Administrator read/list Global,
compare its getter with PowerShell 8.1.0, toggle the Boolean, verify read-back,
restore its original value, and independently verify restoration. CI requires
three pinned comparisons and a cleaned journal entry. Global read/update tests
share the tenant lock. Permanent acceptance responses provide unit decoding and
request fixtures. These checks do not establish effective legacy-client behavior,
delegated authentication or other clouds.

See [qualification](../../../docs/acceptance/messaging-migration-configuration-lab-acceptance.md).
