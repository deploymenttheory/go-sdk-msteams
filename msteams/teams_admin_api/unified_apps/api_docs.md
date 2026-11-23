# Unified apps

`sdk.TeamsAdminAPI.UnifiedApps` manages the Microsoft 365 unified Teams app
catalog. The module implements these cmdlets in
`Microsoft.Teams.PowerShell.TeamsCmdlets` (not ConfigAPI). They call the Substrate
AdminAppCatalog service directly.

| Method | Cmdlet | Wire | Success |
|---|---|---|---|
| `List` | Get-AllM365TeamsApps | `GET /AdminAppCatalog/ps/v2/admin/unifiedApp?$expand=DeployedTo,AvailableTo,metadata` | 200, JSON array |
| `Get` | Get-M365TeamsApp | `GET /AdminAppCatalog/ps/v2/admin/unifiedApp/{id}?$expand=metadata` | 200, JSON object |
| `ListCustomPending` | Get-M365UnifiedCustomPendingApps | `GET /AdminAppCatalog/ps/v2/admin/unifiedApp/stagedApps` | 200, JSON array |
| `GetTenantSettings` | Get-M365UnifiedTenantSettings | `GET /AdminAppCatalog/ps/v2/admin/unifiedApp/settings[?SettingNames=a,b]` | 200, JSON array |
| `Update` | Update-M365TeamsApp | `PATCH /AdminAppCatalog/ps/v2/admin/unifiedApp` | 200, 202, 204 |
| `ReviewCustomPending` | Update-M365UnifiedCustomPendingApp | `GET …/stagedApps`, `GET /AdminAppCatalog/v2/admin/unifiedApps/stagedApps/{id}/details`, `PUT …/stagedApps/review` | 200, 202, 204 |
| `UpdateTenantSettings` | Update-M365UnifiedTenantSettings | `PATCH /AdminAppCatalog/ps/v2/admin/unifiedApp/settings` | 200, 202, 204 |

Every request sends `Accept: application/json` and `X-MS-CmdletName`. Writes send
JSON. Other successful statuses return `ErrResponse` with the response.

## Authentication profile

The `AppCatalog` backend has these fixed values:

- Origin: `https://substrate.office.com`
- Namespace: `/AdminAppCatalog`
- Token audience: `08ff1ce2-4973-4b08-86a3-ebed13badc7f`
- Delegated scope: exactly `08ff1ce2-4973-4b08-86a3-ebed13badc7f/AdminAppCatalog.ReadWrite.All`

The pinned module lists Substrate only in its commercial endpoint table. Other
clouds fail locally with `client.ErrDestination`.

Microsoft documents these cmdlets as unsupported under application
authentication. The lab confirms it: an app-only token for either Substrate
audience gets an empty 401, as does the pinned module connected with a
certificate. Application mode therefore fails locally with
`client.ErrAuthenticationMode`, before any token is acquired.

Supplied-token callers must provide the scope-keyed token,
`binding.RequestScope(audience, scope).TokenKey()`.

## Behaviour

- **Read models.** All models keep absent and null fields as nil, keep timestamp
  text unparsed, and keep every service field in `RawFields`. The list model
  doesn't declare the expanded `metadata`, so it's only in `RawFields`.
- **Assignment entries.** The pinned model types assignment and installation
  entries' `id`, `assignedBy` and `deployedBy` as string arrays. The SDK follows
  that typing.
- **`Update` validation.** `Update` mirrors the module's checks before any I/O:
  - At least one of `IsBlocked`, `AssignmentType` or `InstallType` is required.
  - `UsersAndGroups` needs an `add` or `remove` operation, compared
    case-insensitively and sent lowercased, plus at least one user or group.
    Duplicate members are removed.
  - `Everyone` and `Noone` forbid members and always send operation `add`.
  - Types are matched case-insensitively and sent as supplied.
  - Member, operation or version fields without their type are rejected.
- **`ReviewCustomPending`.** It follows the module's three steps. If the ID isn't
  in the pending list it returns `ErrPendingAppNotFound`. If the details carry no
  `stagedAppEntitlement.iteration` it returns `ErrMissingIteration`. In both
  cases no review is sent. The status is case-insensitive and sent as
  `Published` or `Rejected`.
- **Tenant settings.** Setting names and values use the module's case-sensitive
  sets. `DefaultApp`, `GlobalApp` and `PrivateApp` accept only `All` or `None`,
  with no users, groups or operation.
- **Query encoding.** Query keys and values are percent-encoded, so `$expand`
  goes out as `%24expand`. The module sends it literally.

See [unit tests](service_test.go) and the
[contract notes](../../../docs/implementation/ext/unified-apps-contract-notes.md).
