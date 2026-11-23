# Shifts sync contracts

Owner: EXT-28. Baseline: MicrosoftTeams 8.1.0. Package:
`msteams/teams_admin_api/shifts_connections` (shared with EXT-27; see
[shifts-connections-contract-notes.md](shifts-connections-contract-notes.md)).

All nine cmdlets are AutoRest cmdlets in
`Microsoft.Teams.ConfigAPI.Cmdlets.private.dll`. Evidence, in order of weight:

1. The pinned module's own requests, captured with `-HttpPipelinePrepend`
   against the lab (method, path, query and exact body; reserved IDs only).
2. The generated `ConfigApiBasedCmdlets` method signatures (declared response
   delegates give the success statuses and response models) and the
   `ConfigApiBasedCmdlets+<…>d__N.MoveNext` IL (HTTP method).
3. Model `ToJson` IL and reflected property types (wire property names and
   types).

## Routes

| Cmdlet (parameter sets) | Generated method | Wire | Success |
| --- | --- | --- | --- |
| `Get-CsTeamsShiftsConnectionErrorReport` (Get) | `TeamsGetShiftsConnectionsErrorReport` | GET `/Teams.ShiftsConnector/api/v2/reports/errors?activeness&connectorInstanceId&connectionId&teamId&operation&procedure&code&before&after` | 200 `GetAllErrorReportsResponse {errorReports}` |
| `Get-CsTeamsShiftsConnectionErrorReport` (Get1, GetViaIdentity) | `TeamsGetShiftsConnectionsErrorReport1` | GET `…/v2/reports/errors/{errorReportId}` | 200 `ErrorReportResponse` |
| `Get-CsTeamsShiftsConnectionOperation` | `TeamsGetShiftsConnectionOperation` | GET `…/v1/operations/{operationId}` | 200 `GetOperationResponse` |
| `Get-CsTeamsShiftsConnectionSyncResult` | `TeamsGetShiftsConnectionsSyncResults` | GET `…/v1/connectorinstances/{connectorInstanceId}/usersync/{teamId}` | 200 `GetUserSyncResponse` |
| `Get-CsTeamsShiftsConnectionTeamMap` | `TeamsGetShiftsConnectionsTeamMap` | GET `…/v1/connectorinstances/{id}/teamconnects` | 200 `GetTeamsConnectsResponse {teamConnects}` |
| `Get-CsTeamsShiftsConnectionWfmTeam` (Get, GetViaIdentity) | `TeamsGetShiftsConnectionsWfmTeam` | GET `…/v1/connectorinstances/{id}/wfmteams` | 200 `GetWfmTeamsResponse {wfmTeams: WfmTeam}` |
| `Get-CsTeamsShiftsConnectionWfmTeam` (Get1, GetViaIdentity1) | `TeamsGetShiftsConnectionWfmTeam` | GET `…/v2/connections/{connectionId}/wfmteams` | 200 `WfmTeamsResponse {wfmTeams: WfmTeamResponse}` |
| `Get-CsTeamsShiftsConnectionWfmUser` | `TeamsGetShiftsConnectionsWfmUsers` | GET `…/v1/connectorinstances/{id}/wfmteams/{wfmTeamId}` | 200 `GetWfmUsersResponse {wfmUsers}` |
| `New-CsTeamsShiftsConnectionBatchTeamMap` | `TeamsCreateShiftsConnectionBatchTeamMap` | PUT `…/v1/connectorinstances/{id}/teamconnects` | 202 `TeamConnectsResponse` |
| `Remove-CsTeamsShiftsConnectionTeamMap` | `TeamsRemoveShiftsConnectionsTeamMap` | DELETE `…/v1/connectorinstances/{id}/teamconnects/{teamId}` | 202, no body |
| `Remove-CsTeamsShiftsScheduleRecord` | `TeamsRemoveShiftsScheduleRecord` | POST `…/v1/clearschedule/clear` | 200, no body |

Every other status goes to the module's error delegates (`onBadRequest`,
`onUnauthorized`, `onForbidden`, `onNotFound`, `on424`, `on508`,
`onDefault`), so the SDK declares only the statuses above; any other 2xx is
`ErrResponse`. The error report list and the connection WFM team read also
take the optional `-Authorization` raw header, which the SDK does not expose
(as for EXT-27).

## Bodies

Captured from the module:

- Batch team map: `{"teamMappings":[{"teamId":…,"wfmTeamId":…,"timeZone":…}]}`;
  `timeZone` is omitted when not set.
- Clear schedule: `{"dateRange":{"startDate":…,"endDate":…},"teamId":…,
  "clearSchedulingGroup":false,"entityTypes":["shift","swapRequest"],
  "designatedActorId":…,"timeZone":…}`. `teamId`, `clearSchedulingGroup`
  (a mandatory switch, so always sent) and `entityTypes` are always present;
  `dateRange` appears only when a date is set. Dates use the .NET format
  `yyyy-MM-ddTHH:mm:ss.fffffffK`; the SDK formats a `time.Time` the same way
  (`Z` for UTC, otherwise the numeric offset).

## Response models

- `ErrorReportResponse`: `ttl`, `id`, `tenantId`, `wfmConnectorInstanceId`,
  `connectionId`, `teamId`, `operation`, `procedure`, `code`, `message`,
  `parameters` (string array), `referenceLinks` (string dictionary),
  `revisitIntervalInMinutes`, `errorNotificationSent`, `resolvedAt`,
  `resolvedNotificationSentOn`, `revisitedAt`, `totalIncidents`,
  `intermediateIncidents`, `createdAt`, `culture`, `scheduleSequenceNumber`
  (int64), `severity`, `errorType`. The SDK keeps timestamps as service text.
- `GetOperationResponse`: `Id` (capitalised on the wire), `type`, `tenantId`,
  `wfmConnectorInstanceId`, `createdDateTime`, `lastActionDateTime`, `status`,
  `details` of `{status, error, properties}`.
- `GetUserSyncResponse`: `successfulUsers` of `{userId, userName, wfmUserId,
  wfmUserName}`, `failedWfmUsers` and `failedAadUsers` of `{id, name}`.
- `TeamConnectResponse`: `teamId`, `teamName`, `wfmTeamId`, `wfmTeamName`,
  `timeZone`, `state`. `TeamConnectsResponse`: `operationId`,
  `createdDateTime`, `lastActionDateTime`, `status`.
- `WfmTeam`: `id`, `name`; `WfmTeamResponse` adds `connectorInstanceId` and
  `teamId`. The SDK uses one `WfmTeam` type for both.

PowerShell unwraps the collection envelopes and writes the items; the SDK list
methods do the same. Every model keeps unknown fields in `RawFields`.

## Input rules

The module asserts only non-null IDs and bodies, and an unanchored
`[A-Za-z0-9\-]+` match on `wfmTeamId`. The SDK rejects blank, control-character
and path-traversal IDs, an empty mapping batch, mappings without a team or WFM
team, and schedule clears without a team or entity types, all with `ErrInput`
before any request.

## Lab observations (client-secret, app-only Teams Administrator)

- Every route returns HTTP 403
  `{"code":"Forbidden","message":"Access Denied.","action":"Provide different credential or request access."}`,
  identical to EXT-27. Pinned PowerShell writes that body to the pipeline as
  `ErrorResponseAutoGenerated2` and records no error, for all nine cmdlets.
- The write probes (batch team map, team map removal, schedule clear) target
  the reserved connector instance `00000000-0000-4000-8000-000000000001` and
  team `00000000-0000-4000-8000-000000000002`. They are denied before
  validation, so no mapping or schedule record was created or removed.

Acceptance: `TestAccShiftsSyncAuthorizationDenied` (read step, with pinned
PowerShell output-type parity for every sync read) and
`TestAccShiftsSyncRejectedWrites` (serial step). Sanitized 403 responses from
these runs (`list_error_reports_403.json`, `list_team_maps_403.json`,
`map_teams_403.json`) feed the unit contract tests.

## Unqualified

Successful bodies for every read, the 202 batch-mapping receipt and operation
polling, team map removal, schedule clearing, the error report filters'
server-side semantics and the accepted `activeness` and entity type values all
need credentials the Shifts connector accepts plus a WFM test system with a
mapped team. Success shapes are modelled from the module's serializers and
covered only by synthetic contract tests.
