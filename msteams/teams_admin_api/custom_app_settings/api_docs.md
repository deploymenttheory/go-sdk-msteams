# Custom app settings

Use `sdk.TeamsAdminAPI.CustomAppSettings` for the tenant-wide Teams apps
settings, including whether users can interact with sideloaded (custom) apps.
Operations use the shared client/auth/transport controls through the ConfigAPI
backend and need no PowerShell runtime. Constructing the root client or this
service performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `Get(ctx)` | `Get-CsTeamsSettingsCustomApp` | GET `/Teams.MiddletierService/tenantWideAppsSettingsGlobal` | 200 |
| `SetSideloadedAppsInteraction(ctx, enabled)` | `Set-CsTeamsSettingsCustomApp -isSideloadedAppsInteractionEnabled` | GET, then PUT the same route | 204 |
| `Update(ctx, UpdateRequest)` | (the PUT the cmdlet sends) | PUT `/Teams.MiddletierService/tenantWideAppsSettingsGlobal` | 204 |

Every method returns the actual Resty response alongside its value/error. `Get`
is replay-safe; PUTs are never replayed automatically. Any other HTTP status is
an `*client.APIError`; an undeclared success status is `ErrResponse`.

## Reading

`Get` returns the whole document: the six `is…Enabled` switches, the `lob…`
branding strings, `appAccessRequestConfig` and the per-app `appSettingsList`,
with `RawFields` keeping every property. The cmdlet shows only
`isSideloadedAppsInteractionEnabled` and reports null as false; use
`Settings.SideloadedAppsInteractionEnabled()` for that exact-key projection.
It reads the canonical raw key when a snapshot is present, so an unexpected
uppercase alias cannot change the pinned public output. Native full-document
fields retain Go JSON matching and stricter wrong-type validation; they are an
explicit enhancement beyond the generated DTO/public one-property projection. The list
reflects the service-managed app catalog (thousands of entries in the lab) and
its length can change between reads.

## Writing

`SetSideloadedAppsInteraction` reads the settings, then
PUTs the six switches and `appAccessRequestConfig` from that read with the new
flag and an empty `appSettingsList`. The lab showed the empty list leaves the
stored per-app entries in place. The module's deserializer does not read the
service's camelCase `lob…` names, so it never sends them; the SDK matches that.
The read and write are not atomic. Source review identified additional nested
config/branding projection differences in this adapter; they remain tracked in
[source notes](../../../docs/implementation/ext/app-catalog-contract-notes.md).
Only the exact-key read projection and strict 204 status are corrected here;
Set is not marked contract-complete.

`Update` sends the given document unchanged (nil fields omitted, a nil
`AppSettingsList` sent as `[]`). The `Lob…` keys use the module's PascalCase
request names. Every list entry needs an `ID`.
