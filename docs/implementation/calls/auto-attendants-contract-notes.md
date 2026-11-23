# Auto attendant contracts

Owner: #179 (CALL-22). Baseline: MicrosoftTeams 8.1.0.

Management commands: Get/New/Set/Remove/Update-CsAutoAttendant,
Get-CsAutoAttendantStatus, Get-CsAutoAttendantSupportedLanguage,
Get-CsAutoAttendantSupportedTimeZone and
Get-CsAutoAttendantTenantInformation. Builder commands (ledger kind
local-builder, owner CALL-20): New-CsAutoAttendantCallFlow, -MenuOption,
-Prompt, -DialScope, -CallableEntity, -Menu and
-CallHandlingAssociation.

All routes are under `Teams.VoiceApps`, captured from the module's
HttpPipeline on the lab tenant:

- Get: GET `auto-attendants` with optional `First`, `Skip`,
  `ExcludeContent`, `NameFilter`, `SortBy`, `Descending` and
  `IncludeStatus` (.NET `True`/`False`) → `{"AutoAttendants":[…],
  "Diagnostics":[]}`; GET `auto-attendants/{id}` → `{"AutoAttendant":{…}}`.
- Status: GET `auto-attendants/status/{id}[?includeResources=a,b]` → a bare
  JSON array of status records (`[]` for a new attendant). Get, New and Set
  also call it after the main request.
- New: POST `auto-attendants` → **201** `{"AutoAttendant":{…}}`.
- Set: **PUT** `auto-attendants/{id}` with the whole attendant → 200
  `{"AutoAttendant":{…}}`.
- Update: POST `auto-attendants/{id}/refresh` → 200 `{"AutoAttendant":{…}}`.
- Remove: DELETE `auto-attendants/{id}` → 200 `{"Diagnostics":[]}`.
- Tenant information: GET `information` → `{"TenantInformation":{
  DefaultLanguage, DefaultTimeZone, SupportedLanguages, SupportedTimeZones,
  EnabledFeatures, MainlineAttendantSupportedLanguages},
  "MainlineAttendantTenantInformation", "Diagnostics"}`. The module shows
  `DefaultLanguageId`, `DefaultTimeZoneId` and `FlightedFeatures`. The
  supported language and time zone lists (no `-Identity`) come from this
  response.
- Supported language: GET `supported-languages/{id}` → `{"Language":{Id,
  DisplayName, Voices[{Id,Name}], VoiceResponseSupported}}`.
- Supported time zone: GET `supported-timezones/{id}` → `{"TimeZone":{Id,
  DisplayName}}`.

Unknown attendant, language, time zone or dial-scope group: HTTP **404** with
top-level `code` 100004. Validation failures: HTTP 400, `code` 100001 on create
and drafts and 100002 on update, with `errorDetails[].ErrorCode` (for example
`AutoAttendant_LanguageIdNull`, `Prompt_EmptyPromptData`,
`Menu_IncompatibleMenuOptions`, `Menu_PromptsMissingForUserActions`,
`MenuOption_DisconnectCallWithUserInput`, `GroupDialScope_GroupIdsEmpty`,
`CallFlow_MenuNull`, `CallHandlingAssociation_HolidayDisabled`).

## Builders are server drafts

The builder commands are not local constructors. Each POSTs its value to a
`draft` route and returns the service's normalized value; nothing is stored:

| Command | Route (under `auto-attendants/`) | Reply key |
| --- | --- | --- |
| New-CsAutoAttendantPrompt | `prompts/draft` | `Prompt` |
| New-CsAutoAttendantMenuOption | `menus/menu-options/draft` | `MenuOption` |
| New-CsAutoAttendantMenu | `menus/draft` | `Menu` |
| New-CsAutoAttendantCallFlow | `call-flows/draft` (assigns `Id`) | `CallFlow` |
| New-CsAutoAttendantCallableEntity | `callable-entities/draft` | `CallableEntity` |
| New-CsAutoAttendantDialScope -GroupScope | `group-dial-scopes/draft` (`{"GroupIds":[…]}`) | `DialScope` |
| New-CsAutoAttendantCallHandlingAssociation | `call-handling-associations/draft` (`Enabled` defaults true) | `CallHandlingAssociation` (adds `Priority`) |

The SDK offers plain Go types for these values plus `Draft*` methods for the
server validation. The live lifecycle creates an attendant whose call flow
IDs the test generated, so the draft round-trip is optional: a call flow ID
only links call handling associations to flows.

Enumerations are strings on the wire (reflected from the module's
`Microsoft.Rtc.Management.Hosted.OAA.Models` enums): ActionType,
DtmfTone, PromptType, CallableEntityType, CallHandlingAssociationType,
DirectorySearchMethod, DialScopeType (the service reports an unset scope as
`Default`), TimeoutDisconnectPromptType, AiDisclaimerType.
`AgentTargetType` is an Int32 in the generated model.

## Wrapper rules

- New: `AutoRecordingTemplateId` requires `EnableMainlineAttendant`;
  enabling the mainline attendant forces voice response on. Mirrored
  locally. The wrapper also adds a default greeting for a mainline
  attendant; the SDK does not invent content.
- New sends no `DirectoryLookupScope` or `VoiceId` unless given; the service
  chooses the language's default voice (`Male` for en-US).
- Set sends the whole attendant (`-Instance`): every property of the read
  object, `AuthorizedUsers []`, `UserNameExtension "None"`, Default scopes,
  and schedules with their recurrence range. `AutoAttendant.Settings`
  reproduces this from a read, skipping nulls and empty lists; the lab PUT
  without `Schedules` was accepted and preserved the other settings.
- Get-CsAutoAttendantStatus: the pinned wrapper parses the records but never
  emits them, so the PowerShell command prints nothing. **Module bug**; the
  SDK returns the records.

## Lab observations

Client-secret, app-only Teams Administrator; the tenant had no
pre-existing attendants.

- App-only credentials can create, update, refresh and delete attendants
  without a resource account, phone number or licence.
- Create answered 201; delete answered 200 and a repeat delete 404.
- Holidays (`auto-attendants/{id}/holidays`) belong to CALL-28 schedules and
  are not implemented here.

Acceptance: `TestAccAutoAttendantReads` (read-only) checks tenant information,
the supported language and time zone reads and their 404s, the missing
attendant answers for get, status, refresh and delete, and the draft routes,
including two service rejections; it compares the reads with pinned
PowerShell. `TestAccAutoAttendantLifecycle` (mutation, one journal entry)
creates an owned `msteams-sdk-…-attendant` with two call flows, reads it,
reads its status and lists it, replaces it from `AutoAttendant.Settings` with a new
time zone and greeting, proves the unchanged settings survived, compares the
read with pinned PowerShell, refreshes, deletes it and observes the 404.
Recovery kind `auto-attendant` resolves a lost create by exact name through
`NameFilter`. Sanitized responses feed the unit contract tests.

Unqualified: operators, user, group, shared voicemail and resource-account
targets, call handling associations bound to real schedules (and holidays,
pending CALL-28), dial-by-name group scopes, audio file prompts, voice
response, mainline attendant and auto recording. These need tenant users,
groups, schedules, uploaded audio, licensed resource accounts or features
not enabled in the lab.
