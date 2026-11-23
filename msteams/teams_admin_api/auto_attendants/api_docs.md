# Auto attendants

Use `sdk.TeamsAdminAPI.AutoAttendants` for Teams auto attendants. Operations
use the shared client/auth/transport controls through the ConfigAPI backend
and need no PowerShell runtime. Constructing the root client or this service
performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `List(ctx, options)` | `Get-CsAutoAttendant` | GET `/Teams.VoiceApps/auto-attendants[?First=&Skip=&ExcludeContent=&NameFilter=&SortBy=&Descending=&IncludeStatus=]` | 200 |
| `GetByID(ctx, identity)` | `Get-CsAutoAttendant -Identity` | GET `/Teams.VoiceApps/auto-attendants/{identity}` | 200 |
| `GetStatusByID(ctx, identity, options)` | `Get-CsAutoAttendantStatus` | GET `/Teams.VoiceApps/auto-attendants/status/{identity}[?includeResources=]` | 200 |
| `Create(ctx, settings)` | `New-CsAutoAttendant` | POST `/Teams.VoiceApps/auto-attendants` | 200, 201 |
| `UpdateByID(ctx, identity, settings)` | `Set-CsAutoAttendant` | PUT `/Teams.VoiceApps/auto-attendants/{identity}` | 200, 204 |
| `RefreshByID(ctx, identity)` | `Update-CsAutoAttendant` | POST `/Teams.VoiceApps/auto-attendants/{identity}/refresh` | 200 |
| `DeleteByID(ctx, identity)` | `Remove-CsAutoAttendant` | DELETE `/Teams.VoiceApps/auto-attendants/{identity}` | 200, 204 |
| `GetTenantInformation(ctx)` | `Get-CsAutoAttendantTenantInformation` | GET `/Teams.VoiceApps/information` | 200 |
| `ListSupportedLanguages(ctx)` | `Get-CsAutoAttendantSupportedLanguage` | GET `/Teams.VoiceApps/information` | 200 |
| `GetSupportedLanguageByID(ctx, identity)` | `Get-CsAutoAttendantSupportedLanguage -Identity` | GET `/Teams.VoiceApps/supported-languages/{identity}` | 200 |
| `ListSupportedTimeZones(ctx)` | `Get-CsAutoAttendantSupportedTimeZone` | GET `/Teams.VoiceApps/information` | 200 |
| `GetSupportedTimeZoneByID(ctx, identity)` | `Get-CsAutoAttendantSupportedTimeZone -Identity` | GET `/Teams.VoiceApps/supported-timezones/{identity}` | 200 |
| `DraftPrompt(ctx, prompt)` | `New-CsAutoAttendantPrompt` | POST `/Teams.VoiceApps/auto-attendants/prompts/draft` | 200 |
| `DraftMenuOption(ctx, option)` | `New-CsAutoAttendantMenuOption` | POST `/Teams.VoiceApps/auto-attendants/menus/menu-options/draft` | 200 |
| `DraftMenu(ctx, menu)` | `New-CsAutoAttendantMenu` | POST `/Teams.VoiceApps/auto-attendants/menus/draft` | 200 |
| `DraftCallFlow(ctx, flow)` | `New-CsAutoAttendantCallFlow` | POST `/Teams.VoiceApps/auto-attendants/call-flows/draft` | 200 |
| `DraftCallableEntity(ctx, entity)` | `New-CsAutoAttendantCallableEntity` | POST `/Teams.VoiceApps/auto-attendants/callable-entities/draft` | 200 |
| `DraftGroupDialScope(ctx, groupIDs)` | `New-CsAutoAttendantDialScope -GroupScope` | POST `/Teams.VoiceApps/auto-attendants/group-dial-scopes/draft` | 200 |
| `DraftCallHandlingAssociation(ctx, association)` | `New-CsAutoAttendantCallHandlingAssociation` | POST `/Teams.VoiceApps/auto-attendants/call-handling-associations/draft` | 200 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads and drafts are replay-safe; create, update, refresh
and delete are never replayed automatically. Any other HTTP status is an
`*client.APIError`; an undeclared success status is `ErrResponse`. Invalid
input fails with `ErrInput` before discovery or credentials are used.

## Missing identities and validation

An unknown attendant, language, time zone or dial-scope group is HTTP **404**
with code `100004`. A body the service rejects is HTTP 400 with code `100001`
on create and drafts (`100002` on update); `errorDetails[].ErrorCode` names
the rule, for example `Menu_PromptsMissingForUserActions` or
`CallHandlingAssociation_HolidayDisabled`.

## Building an attendant

The module's `New-CsAutoAttendant*` builders are not local constructors: each
posts its value to a `draft` route, where the service validates and
normalizes it without storing anything. In Go, build `Prompt`, `MenuOption`,
`Menu`, `CallFlow`, `CallableEntity`, `DialScope` and
`CallHandlingAssociation` values directly and pass them to `Create`; use the
`Draft*` methods when you want the service's validation (or a call flow ID)
up front. A call flow ID only links call handling associations to flows, so
any unique GUID you choose works as well as one from `DraftCallFlow`.

Enumerations are strings on the wire; use the package constants
(`PromptTextToSpeech`, `ActionTransferCallToTarget`, `DtmfTone1`,
`EntityExternalPstn`, `AssociationAfterHours`, …). Menus that accept caller
input need menu prompts, and an `Automatic` option must be the menu's only
option. Schedules are managed with the schedules service and referenced by
`CallHandlingAssociation.ScheduleID`.

## Settings and updates

`Settings` is the create and update body. Nil fields are omitted; a non-nil
empty list is sent as `[]`. `Create` needs `Name`, `LanguageID`,
`TimeZoneID` and `DefaultCallFlow`; the service picks the language's default
voice when `VoiceID` is nil. As in the module, `AutoRecordingTemplateID`
requires `MainlineAttendantEnabled`, and enabling the mainline attendant
turns `VoiceResponseEnabled` on. No resource account is needed to create an
attendant; associate one with the resource accounts service.

The service replaces the whole attendant on update, as
`Set-CsAutoAttendant -Instance` does. Start from `AutoAttendant.Settings`:

```go
attendant, _, err := sdk.TeamsAdminAPI.AutoAttendants.GetByID(ctx, id)
settings := attendant.Settings()
settings.TimeZoneID = &zone
_, _, err = sdk.TeamsAdminAPI.AutoAttendants.UpdateByID(ctx, id, settings)
```

`Settings()` copies every writable property, skipping nulls and empty lists;
schedules are carried as the raw objects read. The update reply carries the
attendant.

## Status

`GetStatusByID` returns the provisioning status records (a non-nil empty
slice for a healthy attendant). The pinned module's
`Get-CsAutoAttendantStatus` parses these records but never writes them to
the pipeline; the SDK returns them. `IncludeResources` values (for example
`AudioFile`, `DialByNameVoiceResponses`) are sent comma-joined.

## Response fields

`AutoAttendant`, `StatusRecord` and `TenantInformation` type common
properties and keep every property, including nulls and unknown ones such as
`AutoRecordingThread`, in `RawFields`. The module renames tenant information
fields (`DefaultLanguage` → `DefaultLanguageId`, `DefaultTimeZone` →
`DefaultTimeZoneId`, `EnabledFeatures` → `FlightedFeatures`); the SDK keeps
the wire names.
