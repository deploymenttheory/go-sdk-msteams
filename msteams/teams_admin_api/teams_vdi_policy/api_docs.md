# VDI policy

`TeamsAdminAPI.TeamsVDIPolicies` implements the nine public pinned 8.1.0 parameter
sets through compiled PolicyRP, with injected scope and no constructor I/O.
GetByIdentity/List/ListFiltered retain complete owned response snapshots and
nullable typed getters. Local Identity/Tag wildcard filtering and exact Global-first
ordering follow the compiled provider.

Settings expose DisableCallsAndMeetings/DisableAudioVideoInCallsAndMeetings (nullable bool) and VDI2Optimization (nullable opaque string).
`policy.Field` distinguishes omission, null, false and empty values. The inspected
setters execute no enums or cross-field rules. CreateByIdentity submits only bound
settings; it does not copy Global or make the display-only ShouldProcess read.
PatchByIdentity submits one sparse PATCH. UpdateByIdentity composes Set's
preparation GET and sparse PATCH and keeps preparation/HTTP acceptance receipts.
DeleteByIdentity sends the compiled Identity/optional Force body. No write replays.

Assign/AssignUser/AssignGroup/AssignTenantDefault preserve compiled scope-specific
casing, null additionalParameters, nullable PolicyName versus explicit empty,
required signed Int32 group rank and tenant Force. Missing/mixed scopes reject
before HTTP; the parameterless source Identity set reports an error and is not a
successful no-op. Group rank zero is sent explicitly. A successful HTTP submission
is distinct from persisted or effective assignment; Grant persistence is unqualified.

Both certificate/client-secret lab modes passed Global read/list and owned CRUD:
both disable flags true and optimization Disabled; sparse Set changes optimization to Enabled and preserves the flags. Each test verifies exact sparse fields, getter persistence, two pinned
PowerShell comparisons, list membership, delete and independent GET404/list absence.
Nine permanent actual sanitized response files feed model/read/filter/full CRUD
units. Fault and source-binding vectors are separate from live server successes.

The lab uses its existing Teams Administrator application. Existing/Global policies
and assignments remain unchanged. Other bind persistence, successful grants,
delegated flows and other clouds need separate qualification. See
`docs/implementation/policies/vdi-virtual-appointments-contract-notes.md` for source
and recovery details.
