# Schedules

Use `sdk.TeamsAdminAPI.Schedules` for auto attendant schedules (fixed
date-time ranges and weekly recurrent hours), the service's range validation
drafts, and auto attendant holiday export and import. Operations use the shared
client/auth/transport controls through the ConfigAPI backend and need no
PowerShell runtime. Constructing the root client or this service performs no
network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `List(ctx, options)` | `Get-CsOnlineSchedule` | GET `/Teams.VoiceApps/schedules[?First=&Skip=&…]` | 200 |
| `GetByID(ctx, id)` | `Get-CsOnlineSchedule -Id` | GET `/Teams.VoiceApps/schedules/{id}` | 200 |
| `Create(ctx, request)` | `New-CsOnlineSchedule` | POST `/Teams.VoiceApps/schedules` | 200, 201 |
| `UpdateByID(ctx, id, schedule)` | `Set-CsOnlineSchedule -Instance` | PUT `/Teams.VoiceApps/schedules/{id}` | 200 |
| `DeleteByID(ctx, id)` | `Remove-CsOnlineSchedule` | DELETE `/Teams.VoiceApps/schedules/{id}` | 200, 204 |
| `NewDateTimeRange(ctx, start, end)` | `New-CsOnlineDateTimeRange` | POST `/Teams.VoiceApps/schedules/schedules/date-time-ranges/draft` | 200 |
| `NewTimeRange(ctx, start, end)` | `New-CsOnlineTimeRange` | POST `/Teams.VoiceApps/schedules/schedules/time-ranges/draft` | 200 |
| `ListHolidays(ctx, autoAttendantID, options)` | `Get-CsAutoAttendantHolidays` | GET `/Teams.VoiceApps/auto-attendants/{id}/holidays?ResponseType=0[&Years=&Names=]` | 200 |
| `GetHolidaysContent(ctx, autoAttendantID)` | `Export-CsAutoAttendantHolidays` | GET `/Teams.VoiceApps/auto-attendants/{id}/holidays?ResponseType=1` | 200 |
| `ImportHolidays(ctx, autoAttendantID, content)` | `Import-CsAutoAttendantHolidays` | PUT `/Teams.VoiceApps/auto-attendants/{id}/holidays` | 200 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; writes, including the drafts, are never
replayed automatically. Any other HTTP status is an `*client.APIError` (a
missing schedule or auto attendant is HTTP 404 with code `100004`); an
undeclared success status is `ErrResponse`. Invalid input fails with `ErrInput`
before discovery or credentials are used.

## Schedule bodies

`CreateRequest` takes exactly one of `FixedSchedule` or
`WeeklyRecurrentSchedule`; the module refuses a schedule with neither, and so
does `Create`. Empty weekdays are omitted from the body and an empty fixed
schedule is sent as `"DateTimeRanges": []`, as in the module.

`UpdateByID` replaces the schedule. Read it with `GetByID`, change it and pass
it back: the body carries `Name`, `Type` and the populated schedule part. `ID`
and `AssociatedConfigurationIDs` are read-only and are not sent; schedules are
attached to auto attendants through the auto attendant itself.

Range values are the service's text. Fixed ranges use `yyyy-MM-ddTHH:mm:ss`
and weekly hours `hh:mm:ss`, with `1.00:00:00` for the end of the day.
`Schedule.Type` is `TypeFixed` or `TypeWeeklyRecurrence`; other values are kept
as received. Response models keep nil for absent or null fields and retain
every service property in `RawFields`.

## Range drafts

`New-CsOnlineDateTimeRange` and `New-CsOnlineTimeRange` are not local
constructors: the module posts the text to the service, which validates and
normalizes it without storing anything. `NewDateTimeRange` accepts
`d/M/yyyy H:mm`, `d/M/yyyy` or `yyyy-MM-ddTHH:mm:ss`, read day-first; a nil end
ends the range at the start of the following day. The reply uses
`dd/MM/yyyy HH:mm`, which is returned unchanged; convert it before using it in
a fixed schedule. An invalid range is an HTTP 400 with a provider error code
such as `TimeRange_InvalidTimeRangeBound`.

## Holidays

`HolidayOptions.Years` and `Names` filter the visualisation. They are joined
into one query value exactly as the module's generated client does it, which
escapes every item after the first twice. `GetHolidaysContent` returns the
exported file bytes and `ImportHolidays` sends them back; per-holiday import
outcomes are in `HolidayImportResult.Statuses` (the service's
`DataTimeRanges` spelling is mapped to `DateTimeRanges`).
