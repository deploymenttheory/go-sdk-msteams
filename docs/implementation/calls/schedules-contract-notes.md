# Schedule contracts

Owner: #76 (CALL-28). Baseline: MicrosoftTeams 8.1.0.

CmdletConfig marks all nine commands as AutoRest cmdlets with custom
wrappers. The generated ConfigApiBasedCmdlets methods give the routes:

- Get-CsOnlineSchedule: GET `/Teams.VoiceApps/schedules` (list; the generated
  client also knows First, Skip, ExcludeContent, NameFilter, TypeFilter, SortBy
  and Descending, which the public wrapper never sends) or
  GET `/Teams.VoiceApps/schedules/{id}`.
- New-CsOnlineSchedule: POST `/Teams.VoiceApps/schedules`, body `Name` plus
  `FixedSchedule` or `WeeklyRecurrentSchedule`.
- Set-CsOnlineSchedule: PUT `/Teams.VoiceApps/schedules/{id}`.
- Remove-CsOnlineSchedule: DELETE `/Teams.VoiceApps/schedules/{id}`.
- New-CsOnlineDateTimeRange / New-CsOnlineTimeRange: POST
  `/Teams.VoiceApps/schedules/schedules/date-time-ranges/draft` and
  `…/time-ranges/draft`. They are service calls that validate and normalize a
  range; nothing is stored. They are not local constructors.
- Get/Export-CsAutoAttendantHolidays: GET
  `/Teams.VoiceApps/auto-attendants/{id}/holidays?ResponseType=0|1&Years=&Names=`;
  Import-CsAutoAttendantHolidays: PUT the same path with
  `SerializedHolidayRecords` (base64 of the file bytes).

Wrapper behaviour, confirmed by recording each request through
`-HttpPipelinePrepend` against the lab:

- New/Set convert range objects to `{Start,End}`; fixed ranges use
  `yyyy-MM-ddTHH:mm:ss`, weekly hours `hh:mm:ss`. Empty days are omitted;
  `-FixedSchedule` without ranges sends `"DateTimeRanges": []`. New sends
  `IsComplemented` only for `-Complement`; New sends no `Type`.
- Set (`-Instance`) sends Name, Type and the populated schedule part; it sends
  `IsComplemented` always for weekly schedules and RecurrenceRange when set.
  Id and AssociatedConfigurationIds are not sent.
- Draft builders send only bound fields (`{"Start":"1/1/2027"}`).
- Holiday arrays are aggregated as the first item verbatim plus each later item
  escaped, comma-joined, then escaped again
  (`Years=2026%2C2027&Names=a%20b%2Cc%252Cd`). ResponseType is 0 for Get and 1
  for Export. Import omits the unbound `Id`/`TenantId` body fields.
- The wrapper throws locally when neither schedule type switch is given.

Model serializers give `Schedule` (`Id`, `Name`, `Type`,
`WeeklyRecurrentSchedule`, `FixedSchedule`, `AssociatedConfigurationIds`),
`WeeklyRecurrentSchedule` (`MondayHours`…`SundayHours`, `IsComplemented`,
`RecurrenceRange`), `ScheduleRecurrenceRange` (`Type`, `Start`, `End`,
`NumberOfOccurrences`), holiday records (`CallFlow`, `Year`, `Name`,
`DateTimeRanges`), export `ExportHolidaysResult.SerializedHolidayRecords`
(byte[] as base64) and import statuses (`Name`, `DataTimeRanges`, `Succeeded`,
`FailureReason`).

Lab observations (client-secret, app-only Teams Administrator):

- The service returns `Type` as `Fixed` or `WeeklyRecurrence`; the PowerShell
  projection shows enum numbers. Create returns 201 with `{"Schedule":…}`;
  update 200; delete 200 `{"Diagnostics":[]}`; a missing schedule 404 with
  top-level `code` 100004.
- Create with neither schedule part is 400 (`Schedule_WeeklyRecurrentScheduleNull`).
- The date-time draft parses day-first (`3/1/2027` is 3 January) and, without
  End, ends at the start of the next day; its reply uses `dd/MM/yyyy HH:mm`.
  An invalid range or time is 400 with a provider error code.
- The tenant has no auto attendants, so holiday routes were proven only as 404
  for a reserved ID. Holiday visualisation, export and import bodies are
  modelled from the serializers and covered by synthetic contract cases.

Acceptance: `TestAccScheduleLifecycle` runs the drafts, creates one owned
weekly and one owned fixed schedule (journaled, with a recovery handler), reads,
lists, updates and deletes them with pinned PowerShell parity, and probes the
holiday routes with a reserved auto attendant ID. Associating schedules with an
auto attendant and successful holiday import/export remain unqualified until an
owned auto attendant exists (CALL-22).


## Holiday reply date projection

The read-only [pinned IL observation](evidence/holiday-date-projection-8.1.0.json)
uses `bin/Microsoft.Teams.ConfigAPI.Cmdlets.private.dll`, SHA256
`8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`.
`Generated.Models.GetAutoAttendantHolidaysResponse.ToJson` names the wire envelope
`HolidayVisualizationRecords`; `Generated.Models.DateTimeRange.ToJson` preserves
its `Start` and `End` strings without date normalization.

`Microsoft.Rtc.Management.Hosted.OAA.Models.HolidayVisRecord.ParseFrom` calls
`Microsoft.Rtc.Management.Hosted.Online.Models.DateTimeRange.ParseFrom(IDateTimeRange)`.
The latter loads `dd/MM/yyyy HH:mm`, obtains `CultureInfo.InvariantCulture`, calls
`DateTime.ParseExact` for each bound, then sets Start/End. The evidence records
method-body offsets and resolved member references. This distinguishes the reply
projection from the fixed-schedule request's ISO representation.

The SDK retains raw DTO strings. Integrated holiday assertions parse the pinned
invariant reply format and require exactly the original January 1 to January 2,
2027 calendar boundaries; they do not introduce offset or timezone conversion.
Wrong dates, reversed/equal boundaries, invalid dates and ISO reply strings fail.
The source observation does not itself prove a live holiday response: the
[failed distinct-flow run](evidence/schedule-attendant-542da26-failed.json) lacked
its response body. Bounded diagnostics and narrowly sanitized holiday capture
preserve evidence for the next existing owned-fixture run.


The export assertion proves a nonempty byte receipt only. It does not establish
that the serialized document contains the owned holiday, its dates or a call flow.
The published [Export command](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/export-csautoattendantholidays?view=teams-ps)
describes CSV holiday-name/date-pair rows, but the pinned wrapper only returns
the server-provided byte array. Exact server encoding/date rendering and a
revision-backed content projection remain unqualified. The proof marker therefore
says populated holiday read **and nonempty export receipt**. Export bytes and
base64 content remain in memory and are never fixture artifacts.
