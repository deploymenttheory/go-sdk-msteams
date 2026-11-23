# Update-management policy

`TeamsAdminAPI.TeamsUpdateManagementPolicies` implements the eight public pinned
8.1.0 parameter sets. CRUD uses compiled PolicyRP; Grant is the public modern
Grant-CsTeamsPolicy Function and uses discovery, Accept and X-MS-Target-Uri.
One injected identity-bound Teams scope supplies both providers. No auth or I/O
occurs at construction; writes never replay and caller context is preserved.

GetByIdentity/List/ListFiltered own complete raw snapshots and project all ten
settings. CreateByIdentity applies list edits to a fresh empty list. PatchByIdentity
sends a sparse native PATCH. UpdateByIdentity composes Set's GET/local edit/PATCH
and retains preparation and HTTP acceptance independently. DeleteByIdentity sends
the compiled Identity/optional Force body. Empty updates and Force alone fail
locally. Scalar fields distinguish omission/null/false/zero/empty values.

The source UpdateDayOfWeek setter enforces 0 through 6; this is validated before
HTTP and during decoding. Other strings remain opaque. DisabledInProductMessages
is a lazy get-only string list with replacement, first ordinal removal, Add and
clear behavior; an empty/null bound modifier clears, while a nil edit is unbound.
No unused regex becomes extra validation.

DateTime preserves Unspecified, UTC and Local kinds at CLR 100ns precision.
NewDateTime selects kind explicitly; ParseDateTime accepts provider ISO text.
Unspecified retains a wall clock without timezone, UTC converts and emits Z, and
Local/offset input uses the host timezone like the pinned parser. Time/Kind expose
the value deliberately. Finer Go nanoseconds and out-of-range years are rejected.
The zero DateTime is the unspecified CLR minimum. JSON snapshots retain original
wire fields; typed DateTime marshaling uses the pinned trimmed seven-digit format.
[Json.NET defaults](https://www.newtonsoft.com/Json/help/html/T_Newtonsoft_Json_JsonSerializerSettings.htm)
and offline vectors against the loaded pinned serializer support this contract.

Grant methods select user, tenant or group with optional signed nullable Int32 rank.
The higher Assign wrapper omits rank zero like PowerShell; native AssignToGroup
can explicitly send zero. PolicyName is an explicit string; empty clears. Public
Force is absent; the lower tenant helper exposes its evidenced native force field.
202 group acceptance is retained separately from 204 and is not propagation proof.
PassThru is PowerShell output behavior. Grant persistence remains unqualified.

Both auth modes passed Global reads/list and owned CRUD. A non-default description,
managed-update schedule, public-preview value and actual documented message category
were created; description-only PATCH preserved all omitted getters, DateTime clock
change persisted, Add/Remove and clear persisted, five pinned comparisons passed,
and deletion required independent GET404/list absence. Existing policies/assignments
were untouched. Recovery kind `update-management-policy` owns its journaled policy.
Fifteen actual sanitized response files supply read/model/full CRUD unit replay.
Source/fault vectors remain separate from successful server responses. Other binds,
grants, effective client behavior, delegated modes and other clouds need separate
qualification. [Microsoft reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsupdatemanagementpolicy?view=teams-ps)
can include fields absent from the pinned baseline; none are added speculatively.
