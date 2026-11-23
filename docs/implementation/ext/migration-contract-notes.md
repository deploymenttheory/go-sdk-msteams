# Migration contracts

Owner: #88 (EXT-25). Baseline: MicrosoftTeams 8.1.0.

The family covers nine ledger rows. Only two are default-session exports:

| Command | Export | Dispatch | SDK symbol |
| --- | --- | --- | --- |
| Get-CsMeetingMigrationStatus | TeamsGA | Remoting → custom `Get-CsMmsStatus` | `migration.Service.ListMeetingMigrationStatuses`, `GetMeetingMigrationStatusSummary` |
| Start-CsExMeetingMigration | TeamsGA | Remoting → `Start-CsMeetingMigrationModern` | `migration.Service.StartMeetingMigration` |
| Get-CsMeetingMigrationTransactionHistory | Torus | custom → `Get-CsMeetingMigrationTransactionHistoryModern` | `migration.Service.ListMeetingMigrationTransactionHistory` |
| Get-CsMoveTenantServiceInstanceTaskStatus | SelfHost, Torus | custom → internal generated cmdlet | `migration.Service.ListTenantMoveTaskStatuses` |
| Move-CsTenantCrossRegion | SelfHost, Torus | custom → `New-CsTenantCrossMigration` | `migration.Service.MoveTenantCrossRegion` |
| Move-CsTenantServiceInstance | SelfHost, Torus | custom → `New-CsTenantCrossMigration` | `migration.Service.MoveTenantServiceInstance` |
| Move-CsAvsTenantPartition | SelfHost, Torus | AutoRest | `migration.Service.MoveAVSTenantPartition` |
| Get-NgtProvInstanceFailOverStatus | OCE manifest | `Microsoft.Teams.Policy.Administration.Cmdlets.OCE.dll` | `policyassignments.Service.GetOCENgtProvFailoverStatus` |
| Move-NgtProvInstance | OCE manifest | same OCE assembly | `policyassignments.Service.MoveOCENgtProvInstance` |

The two NgtProv commands share the OCE base (`IC3DirectoryTorusCmdletBase`:
mandatory `-Tenant`, optional `-Environment`, `X-MS-Tenant-Id` and a fresh
`X-MS-Correlation-Id`) with the group policy assignment OCE commands, so they
reuse that package's `OCEContext` and profile selection.

## Routes (generated `ConfigApiBasedCmdlets` IL and `CmdletSettings.json`)

- Status: GET `/Teams.Mms/GetMeetingMigrationStatus` with optional
  `identity`, `state`, `migrationType`, `startTime`, `endTime`. Each is added
  only when non-empty (`String.IsNullOrEmpty`), with `Uri.EscapeDataString`.
  Reply HTTP 200, JSON array of `MeetingMigrationStatusResponse`.
- Summary (`-SummaryOnly`): same query on
  `/Teams.Mms/GetMeetingMigrationStatusSummary`; reply is one object
  `{MigrationType, SummaryOutput:[{State, UserCount}]}`. The wrapper prints it
  with `ConvertTo-Json`; the SDK returns it decoded.
- History: GET `/Teams.Mms/GetMeetingMigrationTransactionHistory?userIdentity=…`
  plus optional `CorrelationId`, `startTime`, `endTime`. Reply array.
- Start: POST `/Teams.Mms/StartMeetingMigration`, body
  `{Identity, SourceMeetingType?, TargetMeetingType?}`; `onCreated` (201,
  `{OperationId}`) and `onNoContent` (204).
- Task status: GET `/Teams.Tenant/TenantMigrationDetail/`, reply array of
  free-form `GmtSchemaItem`; the wrapper outputs each `AdditionalProperties`.
- Cross-region and service-instance moves: POST
  `/Teams.Tenant/CrossRegionMoveTask`, body `ITenantMigration
  {MoveOption, TargetServiceInstance}`, reply 200 `{TenantId, JobName, OrderId}`.
- AVS partition move: POST `/Teams.AVS/MovePartition`, body
  `{SourcePartitionKey, TargetPartitionKey, Workload, ContainerName,
  BasePartitionKey, NumberOfDocuments, PercentageOfPartition, BatchSize}`
  (all optional in the module), declared success `on207` with an array of
  `{documentId, sourcePartitionKey, targetPartitionKey, isSuccessful,
  lastCompletedStage, exception:{ErrorId, CorrelationId, Message}}`; default
  error body `{errorCode, message}`.
- NgtProv status: GET `Teams.InternalSupport/NgtProvOperations/FailoverStatus?region=R`.
- NgtProv move: POST `Teams.InternalSupport/NgtProvOperations/Failover/{A}/{B}?region=R`
  with no body. `DeploymentRegion` is NOAM, EMEA or APAC; A/B are
  `PercentTrafficSideA/B` (Int32, unvalidated). The reply is decoded by
  Newtonsoft into `FailOverResponseWrapper{ResponseMessage}`.

## Module quirks kept or documented

- Date filters: the public cmdlets take `DateTime`, the modern ones `String`.
  PowerShell converts with the invariant `MM/dd/yyyy HH:mm:ss` form (verified:
  `[string][datetime]'2026-11-01T10:00:00'` is `11/01/2026 10:00:00`); the SDK
  formats `time.Time` the same way without zone conversion.
- `Get-CsMeetingMigrationTransactionHistory` declares `[DateTime]` filters
  without defaults, so the module sends `01/01/0001 00:00:00` when they are
  omitted. The SDK omits unset filters.
- `Move-CsTenantServiceInstance` forwards an unbound `[string]` as `""`, so the
  module always sends `TargetServiceInstance`; the SDK sends `""` when nil.
- `Move-CsTenantCrossRegion` sends `{}`: it has no parameters and moves the
  tenant the credential belongs to.
- `Start-CsExMeetingMigration` publicly declares `-CleanupSipDisabled` and
  `-EnqueueSourceType`, but the modern cmdlet does not; pinned PowerShell fails
  with `NamedParameterNotFound` before any request (captured). The SDK does not
  expose them.
- The task status route ends with `/`. The SDK transport sends no empty path
  segment, so it requests `/Teams.Tenant/TenantMigrationDetail`; the gateway
  answered both forms with the same HTTP 403, against 404 for an unknown
  `Teams.Tenant` route.
- Query keys are sent in sorted order with `+` for spaces (Go encoding),
  rather than the module's fixed order and `%20`.
- `MoveAVSTenantPartition` requires source and target partition keys locally;
  the module would send an empty move.

## Lab observations (client-secret, app-only Teams Administrator)

- Status and summary: HTTP 401 with an empty body. Pinned PowerShell fails
  with `Unauthorized,…GetCsMeetingMigrationStatus(Summary)Modern_Get`; the
  oracle (`read-meeting-migration.ps1`) matches both outcomes.
- History (reserved user ID), task status, NgtProv NOAM status: HTTP 403
  `Forbidden`.
- Start for the reserved ID `00000000-0000-4000-8000-000000000001`: HTTP 401.
  The ID cannot name a directory object, so no user or meeting can be acted on.
- AVS partition move between partitions of the reserved tenant in a
  nonexistent container `msteams-sdk-probe`, zero documents: HTTP 403.

## Not sent live

`Move-CsTenantCrossRegion`, `Move-CsTenantServiceInstance` and
`Move-NgtProvInstance` were never sent. The first takes no input and moves the
caller's own tenant; the second's only field the service could reject is a
free-form `MoveOption` with no proof the service validates it before acting;
the third shifts service-wide provisioning traffic. No request exists that the
service is certain to reject before acting, so they are covered by unit
contract tests from module evidence only and their live status is
unqualified. Every successful migration path is likewise unqualified.

Acceptance: `TestAccMigrationReads` (read-only, with PowerShell parity) and
`TestAccMigrationRejectedWrites` (serial, inert probes only).
