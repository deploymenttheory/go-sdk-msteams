# Shifts policy qualification

Owners: #144 (POL-71) and #145 (POL-72). Baseline: MicrosoftTeams 8.1.0.

The public Shifts app Get/Set Functions dispatch through modern ConfigAPI helpers,
with optional Identity, local Filter and non-nullable Boolean setter. Native
Get/List/ListFiltered and sparse PUT/UpdateGlobal preserve those behaviors.
The public Shifts commands are compiled Core/PolicyRP, with six nullable settings,
CRUD/list/filter and four grant parameter sets. Source setters add no enums or
numeric range validation. Requests preserve omission/null/zero/false distinctions,
context, whole-name escaping and no-replay mutations. Typed/raw models own all
returned fields; unknown Shifts fields such as ShiftNoticeMessage remain retained.

Primary evidence is checked runtime metadata, public proxy definitions,
CmdletConfig, merged modern helpers, generated generic configuration methods,
TeamsShiftsPolicy and its Core/provider implementations. Oracles check pinned
8.1.0 module/assembly hashes and tenant/owned-identity binding. Existing published
hashes in the network/notification qualification apply to these same assemblies.

On 2026-11-22 both certificate and client-secret modes passed Global reads/list
for both services with complete raw-value agreement and pinned getter comparisons.
Owned Shifts CRUD created ShiftNoticeFrequency=Always, AccessGracePeriodMinutes=5
and EnableScheduleOwnerPermissions=true, changed only the grace period to10,
verified omitted frequency/Boolean retention, compared all six getters with
PowerShell twice, listed, deleted and independently required GET404/list absence.
Its `shifts-policy` intent was journaled before create and cleaned.

The Shifts app test journals the original non-null Global Boolean before toggling,
compares the pinned module before/after/restoration, restores it through the bound
`shifts-app-configuration` handler and independently verifies the original state.
Both auth runs completed cleanup. CI serializes its Global reads and updates with
all tenant writes. No existing Shifts policies or assignments were changed.

Fourteen permanent sanitized response files supply the service unit tests. Source
binding/decoder vectors and controlled failure replies test exact requests without
inventing successful server responses. Full ticket acceptance remains open for
additional field combinations, null/empty persistence, grant scopes/propagation,
effective location/access behavior, delegated modes, other clouds and locale
collation. Selected-operation evidence is not a complete release qualification.

Current Microsoft references: [Shifts policy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsshiftspolicy?view=teams-ps),
[Shifts app policy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamsshiftsapppolicy?view=teams-ps).
