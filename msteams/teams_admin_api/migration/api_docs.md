# Migration

Use `sdk.TeamsAdminAPI.Migration` for meeting migration status and history and
for the module's meeting, tenant and AVS partition migration operations.
Operations use the shared client/auth/transport controls through the ConfigAPI
backend and need no PowerShell runtime. Constructing the root client or this
service performs no network I/O.

> **Irreversible operations.** `StartMeetingMigration`, `MoveTenantCrossRegion`,
> `MoveTenantServiceInstance` and `MoveAVSTenantPartition` change where a
> user's meetings or a tenant's data live. Neither this SDK nor the module can
> undo them. They are never replayed automatically, and every one except
> `StartMeetingMigration` is a Microsoft-internal (SelfHost/Torus) command that
> ordinary tenant administrators are denied. `MoveTenantCrossRegion` sends an
> empty body and moves the tenant the credential belongs to.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `ListMeetingMigrationStatuses(ctx, options)` | `Get-CsMeetingMigrationStatus` | GET `/Teams.Mms/GetMeetingMigrationStatus` | 200 |
| `GetMeetingMigrationStatusSummary(ctx, options)` | `Get-CsMeetingMigrationStatus -SummaryOnly` | GET `/Teams.Mms/GetMeetingMigrationStatusSummary` | 200 |
| `ListMeetingMigrationTransactionHistory(ctx, identity, options)` | `Get-CsMeetingMigrationTransactionHistory` | GET `/Teams.Mms/GetMeetingMigrationTransactionHistory?userIdentity=…` | 200 |
| `StartMeetingMigration(ctx, request)` | `Start-CsExMeetingMigration` | POST `/Teams.Mms/StartMeetingMigration` | 201, 204 |
| `ListTenantMoveTaskStatuses(ctx)` | `Get-CsMoveTenantServiceInstanceTaskStatus` | GET `/Teams.Tenant/TenantMigrationDetail` | 200 |
| `MoveTenantCrossRegion(ctx)` | `Move-CsTenantCrossRegion` | POST `/Teams.Tenant/CrossRegionMoveTask` (`{}`) | 200 |
| `MoveTenantServiceInstance(ctx, request)` | `Move-CsTenantServiceInstance` | POST `/Teams.Tenant/CrossRegionMoveTask` | 200 |
| `MoveAVSTenantPartition(ctx, request)` | `Move-CsAvsTenantPartition` | POST `/Teams.AVS/MovePartition` | 207 |

The OCE provisioning failover commands live with the other OCE commands on
`sdk.TeamsAdminAPI.PolicyAssignments`:
`GetOCENgtProvFailoverStatus` (`Get-NgtProvInstanceFailOverStatus`) and
`MoveOCENgtProvInstance` (`Move-NgtProvInstance`, an irreversible
service-wide traffic shift).

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe. Any other HTTP status is an
`*client.APIError`; an undeclared success status is `ErrResponse`; invalid
input is `ErrInput` before any I/O.

## Filters

`StatusOptions` and `TransactionHistoryOptions` send only non-nil fields.
`StartTime`/`EndTime` are formatted as `MM/dd/yyyy HH:mm:ss`, the form
PowerShell produces when it binds a `DateTime` to these string parameters; the
value is not converted between time zones. States are `Pending`,
`InProgress`, `Failed` and `Succeeded`; migration types are listed as
`MigrationType*` constants. The service validates both.

## Results

Models keep pointer fields (nil for absent or null) and `RawFields` with every
service property. `ListTenantMoveTaskStatuses` records have no fixed schema
and expose only `Fields`. `MoveAVSTenantPartition` returns one result per
document: check each `IsSuccessful` and `Exception`. `StartMeetingMigration`
returns the HTTP 201 receipt (`OperationId`) or nil for HTTP 204.

## Restrictions

- App-only Teams Administrator credentials receive HTTP 401 (empty body) for
  status, summary and start, and HTTP 403 for history, task status and the AVS
  move. `errors.Is(err, auth.ErrUnauthenticated)` / `auth.ErrForbidden`
  report them.
- `Start-CsExMeetingMigration -CleanupSipDisabled` and `-EnqueueSourceType`
  fail in the module with `NamedParameterNotFound`; they are not exposed.
- See the [testing guide](../../../docs/guides/testing.md) for evidence and
  what remains unqualified.
