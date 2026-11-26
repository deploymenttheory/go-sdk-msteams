# Find-CsGroup finite decision review

Independent review passed. Exactly one implementation/offline-contract record
and its sole ordinary set are approved; all 745 live records/references remain
identical. This decision covers Find-CsGroup and its accepted
__AllParameterSets signature, not other tenant/support/licensing commands.

The [versioned decision packet](evidence/find-group-3ab0f62-contract.json) binds
head 3ab0f62, implementation f2eeaf7, exact runtime/reconciliation pointers,
source/test hashes and actual CI execution. The [portable primary packet](evidence/find-group-source.json)
contains own source and reproducers: 31 model, 11 envelope and 21 binding cases,
81 selected IL bodies, zero unresolved metadata operands. Independent root and
peer reviews approved draft publication; root independently verified all 12 file hashes, raw CI hash, checkout revision,
18 actual package/step records and exact metadata ownership before approving
this bounded offline-contract decision.

## Source, inputs and request

The default root export is an AutoRest function. Its custom wrapper removes
Force, forwards to the internal proxy, then private Find-CsGroup_Find. Its own
invocation getter/default pipeline is traced before reusing the common header
writer proof; the private header is source-derived, not captured request data.
The only route is GET /Teams.VoiceApps/groups; exactly HTTP 200 is declared.
SearchQuery is mandatory string data; native empty input fails before I/O,
while whitespace/control data is encoded. The observed three-character server
minimum stays a server 400, not an invented client restriction.

Public MaxResults is Nullable UInt32 but the internal script parameter is Int32.
The three observed phases preserve a critical distinction: direct compiled
Int32 null binding fails, while the actual internal script proxy converts bound
null to bound zero. Only its dispatch target was replaced by an inert compiled
cmdlet; the public stage is a matching inert Nullable UInt32 script. Omission,
zero, one, MaxInt32 and overflow are all recorded. Native nil means unbound;
non-nil zero expresses the source null-to-zero wire outcome. Values above
MaxInt32 fail early with ErrInput; that is a native representation of the later
source script-binding error, not identical telemetry/session sequencing.

ExactMatchOnly and MailEnabledOnly preserve explicit false: the private caller
checks BoundParameters.ContainsKey and constructs nullable booleans from the
bound switches. Unbound options are omitted. Source uses Uri.EscapeDataString;
Go form encoding differs in spelling of spaces but preserves decoded query
values. No arbitrary origin/path or mutation ID is added.

## Models, output, errors and controls

The own generated Group/GetGroupsResponse paths use exact-case/type and reject
exact duplicate JSON keys recursively. Missing/null/wrong-type Groups in an
object yields empty output. Empty/whitespace/array roots yield null generated
output; literal null fails. Null/nonmodel rows and failed public ID conversion
fail the buffered group result, with no partial group slice. Every successful
HTTP receipt remains available on projection errors or nil/empty output.

Actual GroupModel.ParseFrom calls Guid.Parse. The compatible public ID remains
*string but normalizes D/N/B/P/X inputs to canonical D text; original properties
stay independently in RawFields. Missing/null/wrong-type/invalid IDs fail
projection. This proves a response GUID conversion only; no stricter GUID
constraint is imposed on search strings, TenantId or other directory inputs.
The actual runtime helpers and observations settle D prefix/plus compatibility,
trailing-NUL fallback, X plus/second-prefix handling, UInt32-to-UInt16 conversion
and byte/UInt32 overflow. DisplayName wrong types become nil. Typed-versus-raw
and complete model/envelope projections are source-derived tests, not expected
values copied from the implementation.

Write-AdminServiceDiagnostic precedes buffered public group output. Info/default
messages may already have appeared when later group projection fails; Warning
and Verbose use their streams. Native results contain groups only and retain
the raw response without default tenant-content logging. This is an explicit
presentation boundary. PowerShell common stream/error/session controls map to
native client configuration and Go errors/context. There is no WhatIf/Confirm
surface; Force is removed. Native standard JSON and typed API inputs do not
pretend to reproduce arbitrary PowerShell language argument conversion.

Auth remains the shared Teams Admin audience/discovered-forest client with no
fallback identity or new permissions. Native errors keep response/status and
normal API classification; the real short-query 400 fixture is replayed.
Shared ready/client context, nil and closed controls remain; the existing local
input test checks zero I/O and nil service through the common ready path, not
an invented per-Find cancellation execution. Reads are replay-safe under client
HTTP policy, while projection failure causes no additional request: the new
real-root-client cases contain exactly one GET after discovery. No ownership or
absence is inferred from nil generated output. The lab negative-result check
now requires an actual empty slice rather than a nil response model.

## Actual execution and boundaries

Quality run 45950006235 at head 3ab0f62 completed SUCCESS. All three platforms
(ubuntu/windows/macos) passed build, vet, offline unit tests and race detection;
the vulnerability job passed. Checkout logs bind the tested PR merge to
0b87d92734ce5c90e6b9017f0335a211eba8a1bb. The packet records all 18 affected
package/step results, raw-log SHA and job URLs. Tenant identity lab compiles but
explicitly selects no tests in those offline steps; no individual test-pass
count is invented from this non-verbose log. SDK tenant identity and the root
EXT suite do execute their selected tests. Local Go was not started below the
16 GiB admission floor.

The separate current-head TestAcc tenant_identity job 140343772513 in run
45950006368 succeeded. This is scoped job evidence, not blanket successful live
qualification, delegated coverage, artifact/assertion/cleanup review or all
parameter behavior proof. All live states/references remain unchanged.

The old lint job exhausted its five-minute package-loading deadline without
code diagnostics. Main `39bdb6a` adds a separate export-loading step. Fetched main then advanced
to `155c2b2` with an external lint-action pin update; merge `c40d2cf` integrates
those workflow-only changes. The exact tree diff changes only go-lint.yml and
all 12 packet SDK/source/metadata hashes remain identical to 3ab0f62. Fresh full PR
gates remain necessary before merge. No manual tenant call, duplicate dispatch,
role/license purchase, consent change or additional implementation ticket is
requested by this decision. #191 stays open; release readiness remains false.


The ledger promotion changes only Find-CsGroup and its one __AllParameterSets
mapping. EXT now has 14 approved ordinary command/set contracts. The remaining
97 without passing contracts include 15 conditional declarations and two already
approved exclusions; they are not silently marked complete by this decision.
All 745 live objects and reference lists are unchanged, #191 remains open, and
full fresh PR gates are still required before merge.
