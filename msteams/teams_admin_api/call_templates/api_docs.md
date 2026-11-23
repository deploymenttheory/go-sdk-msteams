# Call queue templates

Use `sdk.TeamsAdminAPI.CallTemplates` for the reusable templates that call
queues and auto attendants reference. Each template type is one field holding
a `*Templates[T]` with the same five methods. Operations use the shared
client/auth/transport controls through the ConfigAPI backend and need no
PowerShell runtime. Constructing the root client or this service performs no
network I/O.

| Field | Model | PowerShell noun | Route under `/Teams.VoiceApps` |
| --- | --- | --- | --- |
| `AutoRecording` | `AutoRecordingTemplate` | `CsAutoRecordingTemplate` | `auto-recording` |
| `ComplianceRecording` | `ComplianceRecordingTemplate` | `CsComplianceRecordingForCallQueueTemplate` | `compliance-recording` |
| `SharedCallHistory` | `SharedCallHistoryTemplate` | `CsSharedCallHistoryTemplate`, deprecated `CsSharedCallQueueHistoryTemplate` | `shared-call-history-template` |
| `SharedVoicemailTriageSettings` | `SharedVoicemailTriageSettingsTemplate` | `CsSharedVoicemailTriageSettingsTemplate` | `shared-voicemail-triage-settings` |
| `Tags` | `TagsTemplate` | `CsTagsTemplate` | `ivr-tags-template` |

| Method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `List(ctx)` | `Get-<noun>` | GET `/{route}` | 200 |
| `GetByID(ctx, id)` | `Get-<noun> -Id` | GET `/{route}/{id}` | 200 |
| `Create(ctx, value)` | `New-<noun>` | POST `/{route}` | 201, 200 |
| `UpdateByID(ctx, id, value)` | `Set-<noun>` | PUT `/{route}/{id}` | 200 |
| `DeleteByID(ctx, id)` | `Remove-<noun>` | DELETE `/{route}/{id}` | 200, 204 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads and updates are replay-safe; `Create` and `DeleteByID`
are never replayed automatically. A lost create response is reconciled by
listing for the template name.

## Requests

Nil model fields are omitted from the request. `Create` requires the module's
mandatory parameters and a nil `ID`:

- auto recording: `Name`, `Description`, `TranscriptionEnabled`,
  `RecordingEnabled`, `AgentViewPermission`, `SharePointHostName`,
  `SharePointSiteName` and `RecordingDocumentOwner` (a user object ID);
- compliance recording: `Name`, `Description` and `BotID` (the recording
  bot's resource account object ID);
- shared call history and voicemail triage: `Name` and `Description`;
- tags: `Name`, `Description` and at least one tag with a name and a target
  `ID` and `Type`.

`Description` may be empty. `ConcurrentInvitationCount` must be 1 or 2 when
set; shared call history values must be 0–2.

`UpdateByID` sends a PUT and the service stores the body as sent: omitted
voicemail triage features become false and omitted tags are removed. Read the
template with `GetByID`, change it and send it back whole. `Set-CsAutoRecordingTemplate`
does this merge itself; the other `Set-*` cmdlets take a whole `-Instance`
object, which is the same contract. A non-nil value `ID` must equal `id`.

## Responses

Each response wraps the template in a named property, which the SDK unwraps
and requires:

| Type | Create | Get, update | List |
| --- | --- | --- | --- |
| auto recording | `AutoRecordingTemplate` | `AutoRecordingTemplate` | `AutoRecordingTemplates` |
| compliance recording | `ComplianceRecordingForCallQueue` | `ComplianceRecording` | `ComplianceRecordings` |
| shared call history | `SharedCallHistoryTemplate` | `SharedCallHistoryTemplate` | `AllSharedCallHistoryTemplates` |
| voicemail triage | `SharedVoicemailTriageSettings` | `SharedVoicemailTriageSettings` | `AllSharedVoicemailTriageSettings` |
| tags | `IvrTagsTemplate` | `IvrTagsTemplate` | `IvrTagsTemplates` |

Shared call history responses also carry the same template under a
`SharedCallQueueHistory*` alias property, which is ignored. A compliance
recording `GetByID` returns `BotRequiredDuringCall`, `BotRequiredBeforeCall`
and `BotConcurrentInvitationCount`; they decode into the unprefixed fields.
Voicemail triage features omitted on create are enabled by the service.
Every model keeps all members in `RawFields`, including future ones.

## Errors

Local validation (blank, `.`/`..`, slash, backslash or control-character IDs;
missing mandatory fields; out-of-range values) returns `ErrInput` before
discovery or token acquisition. HTTP failures are `*client.APIError`; a
missing template is HTTP 404 with service code `100004`. Malformed success
bodies, including a missing or null envelope property, return
`client.ErrDecode`, and undeclared success statuses return `ErrResponse`.

The service validates auto recording templates against SharePoint with the
caller's identity: an app-only identity without SharePoint site-admin access
receives HTTP 401, code `100012` (`SharePointClient_Unauthorized`).
