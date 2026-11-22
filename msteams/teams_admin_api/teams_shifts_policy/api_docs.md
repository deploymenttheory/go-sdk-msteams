# Teams Shifts policy

`TeamsAdminAPI.TeamsShiftsPolicies` implements the public compiled MicrosoftTeams
8.1.0 Get/New/Set/Remove/Grant commands through PolicyRP, independently of the
modern Shifts app policy resource. Construction performs no I/O.

`GetByIdentity`, `List` and `ListFiltered` return nullable typed fields plus an
owned complete raw snapshot, including server extensions such as ShiftNoticeMessage.
Lists retain stable Global-first ordering; filters are local Identity/Tag wildcards.
There is no evidenced pagination contract.

Settings are four nullable opaque strings (ShiftNoticeFrequency, ShiftNoticeMessageType,
ShiftNoticeMessageCustom, AccessType), nullable Int64 AccessGracePeriodMinutes and
nullable Boolean EnableScheduleOwnerPermissions. `policy.Field` distinguishes
omission, JSON null, empty strings, false and numeric zero. Pinned setters enforce
no local enum or numeric range; tenant validation remains authoritative. Current
Microsoft documentation describes supported values without changing that baseline
binding contract.

`CreateByIdentity` sends one POST; `PatchByIdentity` sends one sparse PATCH.
`UpdateByIdentity` explicitly prepares Set with GET then PATCH, retaining the
preparation and HTTP acceptance independently of response decoding. An empty Set
or Force alone is rejected before authentication. Optional Force keeps the
compiled string encoding. DELETE includes its compiled Identity/control body.
Mutations never replay automatically; all methods preserve caller context.

Grant methods support user, group with signed Int32 rank, and tenant default,
including explicit clears. `Assign` requires exactly one scope; the otherwise
unbound compiled Identity parameter set is rejected locally. Request shapes and
errors are contract-tested; grant persistence and effective user behavior remain
unqualified.

Both app auth modes passed Global reads/list and owned CRUD in the authorized lab
on 2026-11-22. The test creates Always/5 minutes/schedule-owner=true, updates only
the grace period to 10, preserves omitted settings, compares all six public getters
with pinned PowerShell twice, lists, deletes and independently verifies absence.
Actual sanitized responses in `testdata` supply unit tests. No assignments or
existing policies change. Recovery is bound to the private `shifts-policy` intent.

Permissions: the tested dedicated Teams Administrator application role, certificate
or client-secret credentials. Delegated credentials, other clouds, locale-specific
filters, explicit-null persistence and effective access behavior need separate lab
qualification; selected tests do not establish whole-ticket/release completion.

Primary sources: pinned TeamsShiftsPolicy and Get/New/Set/Remove/Grant Core classes,
BaseTpm/provider implementations and checked runtime metadata. See
[Microsoft New-CsTeamsShiftsPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsshiftspolicy?view=teams-ps)
and the repository qualification document.
