# Call templates contracts

Owner: CALL-24, settling the CALL-01..03 research inline for this family.
Baseline: MicrosoftTeams 8.1.0 (package SHA256
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`).

## Evidence

The 24 commands are AutoRest cmdlets (`custom/CmdletConfig.json`) with custom
wrappers in `custom/Merged_custom_PsExt.ps1`. Routes, verbs and response
envelopes come from the generated `ConfigApiBasedCmdlets` operation bodies and
model `ToJson` members in `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll`,
read by IL inspection. Internal parameter sets come from
`internal/Merged_internal.ps1`. Every route and shape below was confirmed
against the lab with raw requests carrying the SDK's headers, and again
through the pinned module with `-Debug` capture.

| Template | Route under `/Teams.VoiceApps` | Create envelope | Get/Set envelope | List envelope |
| --- | --- | --- | --- | --- |
| Auto recording | `auto-recording` | `AutoRecordingTemplate` | `AutoRecordingTemplate` | `AutoRecordingTemplates` |
| Compliance recording for call queue | `compliance-recording` | `ComplianceRecordingForCallQueue` | `ComplianceRecording` | `ComplianceRecordings` |
| Shared call history | `shared-call-history-template` | `SharedCallHistoryTemplate` | `SharedCallHistoryTemplate` | `AllSharedCallHistoryTemplates` |
| Shared voicemail triage settings | `shared-voicemail-triage-settings` | `SharedVoicemailTriageSettings` | `SharedVoicemailTriageSettings` | `AllSharedVoicemailTriageSettings` |
| IVR tags | `ivr-tags-template` | `IvrTagsTemplate` | `IvrTagsTemplate` | `IvrTagsTemplates` |

Each type has list GET `/{route}`, get GET `/{route}/{id}`, create POST
`/{route}` (201), update PUT `/{route}/{id}` (200) and remove DELETE
`/{route}/{id}` (200, body `{"Diagnostics":[]}`). Missing IDs return 404 with
`{"code":"100004",...}` for get and delete. Every success body also
carries `Diagnostics`.

- The generated list operations accept `First`, `Skip`, `ExcludeContent`,
  `NameFilter`, `TypeFilter`, `SortBy`, `Descending` and (except tags)
  `IncludeStatus` queries, but every public `Get-*` wrapper binds only `Id`,
  so the module always sends a bare list. The SDK matches that.
- `Get/New/Set/Remove-CsSharedCallQueueHistoryTemplate` write a deprecation
  warning and call the `*-CsSharedCallHistoryTemplate` wrapper. Shared call
  history responses carry the template twice: under the names above and
  under `SharedCallQueueHistoryForCallQueue` (create),
  `SharedCallQueueHistory` (get/set) or `AllSharedCallQueueHistory` (list).
  The SDK maps the deprecated commands to `SharedCallHistory`.
- Model members: auto recording `Name`, `Id`, `Description`,
  `TranscriptionEnabled`, `RecordingEnabled`, `AgentViewPermission` (None=0,
  All=1), `SharePointHostName`, `SharePointSiteName`,
  `RecordingDocumentOwner`, `AutoRecordingAnnouncementAudioFileId`,
  `AutoRecordingAnnouncementAudioFileName`,
  `AutoRecordingAnnouncementTextToSpeechPrompt`; compliance `Name`, `Id`,
  `Description`, `BotId`, `RequiredDuringCall`, `RequiredBeforeCall`,
  `ConcurrentInvitationCount`, `PairedApplication`; shared call history
  `IncomingMissedCalls` and `AnsweredAndOutboundCalls` (None=0,
  AuthorizedUsersOnly=1, AuthorizedUsersAndAgents=2) and
  `IncomingRedirectedCalls` (None=0, AuthorizedUsersOnly=1,
  AuthorizedUsersAndGroupMembers=2); voicemail triage
  `EnableUrgencyDetection`, `EnableCategoryDetection`,
  `EnableCallToActionDetection`, `EnableVoiceToTextSummary`; tags `Tags` of
  `{TagName, TagDetails}` where `TagDetails` is a callable entity (`Id`,
  `Type`, `EnableTranscription`,
  `EnableSharedVoicemailSystemPromptSuppression`, `CallPriority`,
  `SharedVoicemailHistoryTemplateId`; reads add `DisplayName` and
  `ConfigurationEndpointType`).
- Compliance `GetByID` returns `BotRequiredDuringCall`,
  `BotRequiredBeforeCall` and `BotConcurrentInvitationCount` (the domain
  model); create, update and list use the unprefixed names. The SDK decodes
  both into the same fields.
- Tag targets are sent with the service entity type name (`ExternalPstn`,
  `User`, `ApplicationEndpoint`, ...); `New-CsTag` builds them through the
  `ivr-tag/draft` and `callable-entities/draft` helper routes, which are not
  CALL-24 commands and are not needed to send a template.

## Wrapper semantics

- `New-*` mandatory parameters: auto recording `Name`, `Description`,
  `TranscriptionEnabled`, `RecordingEnabled`, `AgentViewPermission`,
  `SharePointHostName`, `SharePointSiteName`, `RecordingDocumentOwner`;
  compliance `Name`, `Description`, `BotApplicationInstanceObjectId`; shared
  call history and voicemail triage `Name`, `Description`; tags `Name`,
  `Description`, `Tags`. The SDK enforces the same set locally (`ErrInput`).
- `New-CsComplianceRecordingForCallQueueTemplate` sends
  `ConcurrentInvitationCount` 0 as 1 and rejects values outside 1–2; the SDK
  rejects anything but 1 or 2 when the field is set. The wrapper's
  `PairedApplicationInstanceObjectId` handling references an undefined
  variable, so the module never sends it; the SDK sends `PairedApplication`
  when set.
- `Set-CsAutoRecordingTemplate -Identity` reads the template first (throwing
  "not found" when absent) and PUTs a full DTO merged from the bound
  parameters. Every other `Set-*` takes only `-Instance` and maps all of its
  properties. Both forms send a whole body; the service stores it as sent
  (omitted voicemail triage features become false). `UpdateByID` therefore
  documents read-modify-write and sends exactly the given value.
- **Module defect:** the object returned by
  `New-CsComplianceRecordingForCallQueueTemplate` swaps `RequiredDuringCall`
  and `RequiredBeforeCall` (the wire request and stored template are correct;
  `Get-*` reports them correctly). Passing that object to
  `Set-CsComplianceRecordingForCallQueueTemplate -Instance` stores the swap.
  The SDK decodes the service fields directly and is unaffected.

## Lab observations

- The lab had no templates of any type; tests only create and delete owned
  `msteams-sdk-*` templates.
- Compliance templates are stored without resolving `BotId`; an unassigned
  GUID is accepted.
- Voicemail triage features omitted on create default to true.
- Auto recording create validates `RecordingDocumentOwner` as a user object
  ID (400 `AutoRecordingTemplate_RecordingDocumentOwnerInvalidGuidFormat`,
  then `..._RecordingDocumentOwnerUserNotFound`) and then the SharePoint site
  with the caller's identity: the app-only lab identity receives 401, code
  `100012`, `SharePointClient_Unauthorized`. A PUT with recording and
  transcription both disabled is rejected 400 before the ID is looked up.

## Acceptance

- `TestAccCallTemplateReads` lists all five types and reads a missing ID.
- `TestAccComplianceRecordingTemplateLifecycle`,
  `TestAccSharedCallHistoryTemplateLifecycle`,
  `TestAccVoicemailTriageTemplateLifecycle` and
  `TestAccTagsTemplateLifecycle` each create one journaled template, read it
  back, update it (non-default values in every field), compare the read-back
  with pinned PowerShell `Get-*` (the shared call history oracle also reads
  through the deprecated alias), find it in the list, delete it, observe 404
  and confirm a repeated delete is 404.
- `TestAccAutoRecordingTemplateCreateDenied` journals a create with the lab
  user as document owner and verifies the SharePoint denial, the list, the
  update body validation and the missing-ID delete. **Auto recording
  create/get/update/delete success is unqualified** until a lab identity with
  SharePoint site-admin access is available.
- Recovery resolves a lost create by its unique name over the complete list
  (no match proves absence), re-checks the name before deleting, and treats a
  404 as absence.
