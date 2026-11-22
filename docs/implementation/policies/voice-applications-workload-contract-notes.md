# Voice applications and workload contracts

Owners: #154 (POL-81), #155 (POL-82). Baseline: MicrosoftTeams 8.1.0.

Voice applications exports compiled Get/New/Set/Remove plus a modern Grant
Function. The runtime and CmdletConfig agree on eight public parameter sets,
including three modern grant scopes with optional nullable signed Int32 rank and
no public Force. The injected identity-bound scope must support both providers.
The model exposes 34 settings: 25 nullable booleans and nine nullable opaque
strings (description, queue monitoring and six metrics permissions). Inspected
model setters contain no executed validation exceptions; description length
constants alone do not establish a throwing range rule. The inspected New/Set classes copy the 34 bound properties without adding
other checks. Native CRUD uses compiled GET/sparse PATCH and optional string Force;
Grant uses modern discovery, Accept and X-MS-Target-Uri with the same scope.

Workload exports modern Get/New/Set/Remove/Grant Functions, also eight sets.
CmdletConfig maps these to the shared configuration and Grant-CsTeamsPolicy
helpers. Seven settings comprise description and six publicly non-nullable
booleans for meetings/messaging/calling and pinned variants. The unused compiled
Core model's nullable bools do not determine public request binding. The selected helpers use POST with optional explicit read-back, sparse PUT
without a preparation read and bodyless DELETE. Omitted Set Identity defaults to
Global; the SDK exposes that choice explicitly through UpdateGlobal. Nil request
pointers omit; false and empty stay bound. Public string-null binding becomes
empty, while Boolean arguments are non-nullable.

Both certificate and client-secret modes passed Global read/list and owned CRUD
for both services. Voice applications created a non-default description and two
true administrative flags; description-only Set preserved all other 33 getters.
Workload created all six Boolean values false and preserved them during a
sparse description-only PUT. Exact request-field assertions and full typed
projections verify omission, two pinned comparisons establish getter parity, and
list/delete/independent GET404 plus list absence establish cleanup. Eighteen
permanent actual sanitized response files feed unit model/read/full CRUD replay.

Unit setting metadata checks cover every public New/Set bind. Fault vectors cover
omission/null/false/empty without claiming arbitrary strings are server-valid.
Modern Grant unit cases cover both auth modes, separate discovery, three scopes,
rank omission/zero/signed values and failure receipts. Successful Grant persistence,
clear/propagation, other field persistence, delegated tenant modes and other clouds
remain unqualified. No existing/Global policy, assignment, queue, attendant, license,
permission or secret changed; the owned policies were never assigned.

Sources inspected: checked runtime.json; public proxies; CmdletConfig; Core
TeamsVoiceApplicationsPolicy and New/Set classes; shared BaseTpmNew/Set and HTTP
provider; Get/New/Set/Remove-CsConfigurationModern and Grant-CsTeamsPolicy;
checksum-verified ConfigAPI generated configuration routes and shared adapters.
The modern workload public surface takes precedence over unused compiled classes.
Preceding PR #296 also corrects cross-read comparison of internal XML Key metadata;
complete per-response snapshots and every management/extension value remain tested.
