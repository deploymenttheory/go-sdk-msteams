# Provisioning diagnostics

Use `sdk.TeamsAdminAPI.ProvisioningDiagnostics` for the module's internal
provisioning, directory sync, user settings and hosted migration operations.
They are Microsoft support (SelfHost, Torus and OCE) commands: the MicrosoftTeams
8.1.0 module does not load them in a normal tenant session, and app-only Teams
Administrator credentials are denied every one. Several of them resynchronize,
rehome or move users. Operations use the shared client/auth/transport controls
through the ConfigAPI backend and need no PowerShell runtime. Constructing the
root client or this service performs no network I/O.

## Reads

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `GetDirectoryDiagnosticData(ctx, query)` | `Get-CsBusinessVoiceDirectoryDiagnosticData` | GET `/Teams.Cpc/bvdDiagnosticData/{table}/{region}/{partitionKey}?rowKey=&resultSize=` | 200 |
| `ListObjectChangelog(ctx, query)` | `Get-CsMasObjectChangelog` | GET `/Teams.Tenant/GetMASObjectChangelog?Identity=&SchemaNames=&Versions=` | 200 |
| `ListVersionedSchemaData(ctx, query)` | `Get-CsMasVersionedSchemaData` | GET `/Teams.Tenant/GetSchemaDataFromMAS?SchemaName=&Identity=&Versions=` | 200 |
| `GetUserSetting(ctx, tenantID, userID, name)` | `Get-CsUssUserSettings` | GET `/Teams.UserSettings/api/v1/tenant/{tenantId}/user/{userId}/setting?name=` | 200 |
| `GetUserProvisioningHistory(ctx, tenantID, userID, region)` | `Get-CsUserProvHistory` (OCE) | GET `/Teams.InternalSupport/AdminStoreOperations/userHistory/{userId}/{region}` | 200–204 |
| `ValidateUserMove(ctx, request)` | `Move-CsInternalHelper -ActionType Validateuser` | POST `/Teams.HostedMigrationService/ValidateUser` | 200 |
| `GetUserMovePolicies(ctx, user)` | `Move-CsInternalHelper -ActionType Getpolicy` | POST `/Teams.HostedMigrationService/GetPolicies` | 200 |

## Actions

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `SetUserSetting(ctx, tenantID, userID, name, body)` | `Set-CsUssUserSettings` | PATCH `/Teams.UserSettings/api/v1/tenant/{tenantId}/user/{userId}/setting/{name}` | 200, 201 |
| `InvokeCustomHandler(ctx, payload)` | `Invoke-CsCustomHandlerNgtprov` | POST `/Teams.Tenant/CustomHandler` | 200 |
| `InvokeCustomHandlerCallback(ctx, request)` | `Invoke-CsCustomHandlerCallBackNgtprov` | POST `/Teams.Tenant/CustomHandlerCallback?id=&operation=&eventname=` (no body) | 200 |
| `SyncDirectObjects(ctx, request)` | `Invoke-CsDirectObjectSync` | POST `/Teams.InternalSupport/ProvisioningOperations/DsRequest` | 200 |
| `SyncMsodsObjects(ctx, request)` | `Invoke-CsMsodsSync` | POST `/Teams.Tenant/MsodsSync` | 200 |
| `RehomeUser(ctx, userID, body)` | `Invoke-CsRehomeuser` | POST `/Teams.User/RehomeUser/{userId}/moveUser` | 204 |
| `SetTenantUserBackfill(ctx, operation, body)` | `Set-CsTenantUserBackfill` | POST `/Teams.InternalSupport/SetBackfillToPubsub?operation=` | 200 |
| `BeginUserMove` / `CompleteUserMove(ctx, request)` | `Move-CsInternalHelper -ActionType Beginmove` / `Completemove` | POST `/Teams.HostedMigrationService/BeginMove` / `CompleteMove` | 200 |
| `RollbackUserMove(ctx, user)` | `Move-CsInternalHelper -ActionType Rollback` | POST `/Teams.HostedMigrationService/Rollback` | 200 |
| `RehomeUserMove(ctx, request)` | `Move-CsInternalHelper -ActionType Rehomeuser` | POST `/Teams.HostedMigrationService/RehomeUser` | 200 |
| `SetMovedResourceData(ctx, request)` | `Move-CsInternalHelper -ActionType Setmovedresourcedata` | POST `/Teams.HostedMigrationService/SetMovedResourceData` | 200 |
| `InvokeNgtProvHandler(ctx, request)` | `Invoke-CsGenericNgtProvCommand` (OCE) | POST `/Teams.InternalSupport/ProvisioningOperations/runHandler/{region}?handler=` | 200–204 |
| `SyncDirectoryObject(ctx, request)` | `Invoke-CsDirectoryObjectSync` (OCE) | POST `/Teams.InternalSupport/ProvisioningOperations/DsRequest` | 200–204 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe, except the OCE read, which makes one
attempt as the module does. Actions are never replayed automatically. Any other
HTTP status is an `*client.APIError` (`errors.Is(err, auth.ErrForbidden)` for
403, `auth.ErrUnauthenticated` for 401); an undeclared success status is
`ErrResponse`; invalid input is `ErrInput` before any request.

## Request details

- String `Body` parameters (`SetUserSetting`, `RehomeUser`,
  `SetTenantUserBackfill`, `InvokeNgtProvHandler`) must be JSON and are sent
  compacted.
- `Move-CsInternalHelper` actions send `X-MS-Target-Uri` as the discovered
  forest plus `HostedMigration` and the caller's tenant in `X-MS-Tenant-Id`.
  `UserMove.TenantID` is required for that header. The module always sends the
  three `TeamsData` flags and `moveToCloud`/`force`; so does the SDK.
- OCE commands skip forest discovery, send `X-MS-Tenant-Id` and a fresh
  `X-MS-Correlation-Id` per call, and no cmdlet-name header.
  `SyncDirectoryObject` sends enums as their module index and a non-empty
  `ObjectIDs` instead of `ObjectID`, exactly as the module's JSON serializer.
- `InvokeNgtProvHandler` sends the module's single
  `handler=NAME?tenant=ID?target=TARGET?serviceInstance=INSTANCE` value
  percent-encoded; the decoded value is identical.
- The module sends `GetSchemaDataFromMAS/`, `CustomHandlerCallback/`,
  `MsodsSync/` and `SetBackfillToPubsub/` with a trailing slash, and
  `DsRequest`/`RehomeUser` with a doubled leading slash. The SDK sends the
  canonical path; live probes reached the same gated routes either way.

## Restrictions

- App-only Teams Administrator credentials receive HTTP 403 on every
  operation, except HTTP 401 on both custom handler operations. Success paths
  are not qualified.
- `Get-CsLoginInfo` returns the module's own access token and sends no service
  request; it has no SDK method. Credentials come from `msteams/auth`.
