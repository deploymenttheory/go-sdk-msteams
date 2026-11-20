# POL-01 direct-to-group assignment migration evidence

Research date: 2026-11-20. Scope: six assignment-migration commands in
[POL-01](../backlog/issues/POL-01.md). Static package evidence only; no credentials,
tenant calls or management command execution. Exact parameter sets are recorded
in the [runtime snapshot](../../../contracts/powershell/8.1.0/runtime.json).

## Sources

Package SHA256 is recorded in [the inventory](baseline-inspection.md). Command
classes are in `netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Core.dll`,
SHA256 `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`, namespace
`Microsoft.Teams.Policy.Administration.Cmdlets.Core.Custom.GPA` (G below).
Provider helper source is
`netcoreapp3.1/Microsoft.Teams.Policy.Administration.Cmdlets.Providers.PolicyRp.dll`,
SHA256 `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`, namespace
`Microsoft.Teams.Policy.Administration.Cmdlets.Providers` (P below). ILSpy
11.2.0.10050 reconstructed method bodies into temporary files; no code is vendored.

## Backend and request construction

Each G ProcessRecord obtains the selected Config API application-resource token
through the host contract using its `/user_impersonation` delegated scope. The
resource is the existing cloud-mapped `48ac35b8-9aa8-4d74-927d-1f4a14a0b239`;
application-mode eligibility remains separately unqualified.
`P.CommonHttpHelperRequest.SendGPARequests` constructs
`https://{loginInfo.ConfigApiEndpoint.Host}/Skype.Policy/groupPolicyAssignments/`
plus the operation suffix. These are direct Config API calls, not Graph requests.

The helper adds X-MS-CmdletName, X-MS-Correlation-Id and a module user agent.
`P.HttpClientHelper.CreateHttpClient` adds `MPACmdlet: true`; SendRequestAsync adds
the bearer. Internal/OCE headers can additionally be injected under host context,
but they are not a general public tenant contract. This helper does not invoke
the generated pipeline's forest-discovery/target-URI callback. Therefore the
[generated Config API routing sequence](../core/policy-routing-evidence.md) must
not automatically be copied onto these six paths.

The non-generic `HttpClientHelper.SendAsync` returns the HttpResponseMessage
directly. The inspected paths contain no explicit retry or continuation loop.
Native SDK controls must preserve cancellation, normalize HTTP errors and avoid
automatic replay of a migration job when the submission result is ambiguous.

## Resource operations and command compositions

All paths below are relative to `/Skype.Policy/groupPolicyAssignments`.
Route parameters and query values must be encoded separately by the native SDK;
the original raw string concatenation is not permission for path/query injection.

| Command | Inputs beyond common parameters | Method/path | Canonical operation |
| --- | --- | --- | --- |
| Get-DirectToGroupAssignmentsMigrationStatus | optional MigrationEventId string | GET `/directAssignmentsMigration` when null; GET `/directAssignmentsMigration/{id}` otherwise | MigrationJobs.List / GetByID |
| Get-GroupAssignmentRecommendationsPerPolicyType | mandatory EntityType string; optional GroupThreshold string, default `500` | GET `/{entityType}/recommendations?groupThreshold={threshold}` | AssignmentRecommendations.ListPolicyTypes |
| Get-GroupAssignmentRecommendationsPerPolicyName | mandatory EntityType and PolicyType strings; optional GroupThreshold default `500` | GET `/{entityType}/policyTypes/{policyType}/recommendations?groupThreshold={threshold}` | AssignmentRecommendations.ListPolicyNames |
| Get-GroupPolicyAssignmentConflict | mandatory GroupId and PolicyType strings | GET `/{groupID}/policyTypes/{policyType}/conflict` | GroupAssignmentConflicts.Get |
| Invoke-StartDirectToGroupAssignmentMigration | mandatory Authority, PolicyType, PolicyName strings and GroupId GUID; optional AllowNonEmptyGroup Boolean, default false | POST `/directAssignmentsMigration/migrate?allowNonEmptyGroup=False` (True when selected) | MigrationJobs.Create |
| Invoke-ClearDirectToGroupAssignmentMigration | same parameter shape as Start | POST `/directAssignmentsMigration/cleanup?allowNonEmptyGroup=False` | MigrationJobs.ClearDirectAssignments, explicit action |

Sources: each G command ProcessRecord and parameter attributes; the Clear and
Start CLR type names omit the Invoke prefix. All six have one parameter set.
Get's null-versus-non-null MigrationEventId branch chooses collection versus single
response decoding; an explicitly empty string is not a proved list alias. Native
read-by-ID validation should make that distinction explicit. GroupThreshold has
no observed numeric/range validation; a string default does not prove a numeric
service limit or pagination control.

The Start/Clear body model `G.Models.DirectToGroupAssignmentMigration` explicitly
maps these keys through JsonProperty attributes:

```json
{"authority":"fixture","policytype":"FixturePolicy","policyname":"Fixture","groupid":"00000000-0000-4000-8000-000000000001"}
```

This is a synthetic source-derived shape, not a captured request. Preserve the
lower-case property spelling, including policytype/policyname/groupid. The query
Boolean is always sent by these cmdlets; it is not an omitted optional body field.
The request uses UTF-8 application/json. Clear is not an evidenced DELETE or
permission to delete a group. No migration-job Update/Delete endpoint is established
by these sources; keep those CRUD cells unevidenced rather than invent routes.

## Responses and failure differences

Recommendation results are bare arrays, not Graph value envelopes. Models retain
`policyType` or `policyName` plus integer `daCount`. The conflict model uses camel
case and contains totalCountUsersWithConflict, separate direct/group conflict
counts and detail arrays; each detail has userId, policyInstance and assignmentSource.

Migration status, Start and Clear decode
`Microsoft.Ic3.AdminConfig.RP.Policy.Group.Models.DirectToGroupAssignmentMigrationResponseModel`,
which is also defined in the same Core assembly above despite its different
namespace. Offline reflection and a targeted decompilation agree on camel-case
wire fields: migrationEventId, tenantId and groupId (GUIDs); groupName, policyType
and policyInstance (strings); startTime/completionTime (non-nullable DateTime);
directAssignmentsProcessed (Int32); and state with StringEnumConverter.
Absent non-nullable values produce CLR defaults, so model output does not prove
those defaults were present on the wire. Get without ID expects a list; with ID
it expects one object.

The same assembly's `DirectToGroupAssignmentMigrationState` names NotStarted,
PopulationInProgress, GpaCreationInProgress, WaitingForGpaProcessed,
CleanUpInProgress, Completed, Failed and ConnectionLost. These are model enum
values, not observed server-state transitions. A Go model should preserve unknown
future states and avoid inferring full tenant convergence from a local DTO.
Representative decode/error/state fixtures remain required.

Start checks success status before decoding. Its 404 emits a group-not-found
message; 401/403 emits a token-invalid message; 409/502 emits a group-already-in-use
message. Those are client presentation choices and do not prove that every 502
means a conflict or every 403 is an invalid token. Native errors retain the actual
HTTP status, with authentication and authorization distinct.

Several read/Clear paths deserialize response content without first checking
HTTP success. Recommendation/status JSON exceptions become normal output strings
such as no recommendations/no status. Native SDK list APIs must not convert
malformed or non-success responses into a successful empty result. Do not copy
PowerShell's async-void/Task.Run blocking mechanics; propagate caller context through
the shared request boundary.

## Remaining work

Current Microsoft [Start help](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/invoke-startdirecttogroupassignmentmigration?view=teams-ps),
checked 2026-11-20, describes moving direct assignments into the selected group and
restricts this workflow to tenants with the new group-policy-assignment adoption
features activated. AllowNonEmptyGroup permits an already assigned group. This
documents a tenant prerequisite; it does not identify a reliable capability probe
or qualify this lab. Treat migration/cleanup as consequential acceptance scenarios
under LAB-01's separate opt-in, not ordinary read-only preflight.

Add source-derived request/response/negative fixtures for the settled DTO and
trace any additional service status/error schema; verify required roles/app
eligibility from primary documentation, and settle exact cloud availability and
mutation replay constraints. No command is promoted to contract-complete or
live-tested by this note. These six contracts remain part of the full POL-01 scope.
