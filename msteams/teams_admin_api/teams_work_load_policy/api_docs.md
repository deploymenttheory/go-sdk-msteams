# Workload policy

`TeamsAdminAPI.TeamsWorkLoadPolicies` implements the eight public pinned 8.1.0
parameter sets with caller context, complete owned RawFields/RawJSON snapshots,
local Identity/Tag wildcard selection and Global-first lists. Construction has no
network or auth side effects. Seven settings: description and six publicly non-nullable booleans for calling, meetings, messaging and their pinned flags. Nil request pointers omit; false and empty strings stay explicit.

All public operations use modern ConfigAPI discovery. CreateByIdentity returns
POST acceptance; Create composes POST/Get and retains the successful creation in
ReadBackError if the read fails. UpdateByIdentity sends one sparse PUT without a
preparation read; UpdateGlobal explicitly selects the mapped omitted-Identity
default. Identity-only PUT remains possible. DeleteByIdentity is bodyless. Public
string null maps to an explicit empty string; Boolean null has no request bind.
The unused compiled model does not determine modern write behavior.

Assign selects exactly one user/tenant/group; individual native grant methods
remain available. Empty PolicyName clears, and the argument is required despite
optional public metadata. Group rank is signed optional Int32: the wrapper omits
zero; native AssignToGroup can explicitly send zero. Public Force is absent; the
lower tenant helper exposes its evidenced native force field. HTTP202 acceptance
is separate from HTTP204 and is not propagation proof. PassThru concerns
PowerShell output. No write replays or polls implicitly.

Both app-auth modes passed Global Get/List and owned unassigned CRUD with exact
sparse fields, all typed getter/omission checks, two pinned PowerShell comparisons,
list/delete and independent GET404/list-absence cleanup. Nine permanent sanitized
actual responses supply unit model/read/CRUD replay. Source/fault vectors are
separate from live server successes. All raw metadata is retained per response;
internal XML Key differences are not immutable state across separate reads.

The existing lab Teams Administrator application and secrets suffice. Successful
Grant persistence/clear, other setting persistence, delegated tenant flows and
other clouds remain unqualified. No existing policy or assignment changed.
See the [testing guide](../../../docs/guides/testing.md).
