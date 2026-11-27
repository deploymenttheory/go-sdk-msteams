# Teams call park policies

Use `sdk.TeamsAdminAPI.TeamsCallParkPolicies`. Construction injects the root
client's shared Teams ConfigAPI backend without authentication or network I/O.

| Method | Selected modern behavior |
| --- | --- |
| `GetByIdentity`, `List`, `ListFiltered` | GET with an opaque identity or bare collection and local Identity/Tag filtering |
| `CreateByIdentity` | One POST with creation metadata |
| `Create` | Explicit POST/Get composition matching mapped New |
| `UpdateByIdentity`, `UpdateGlobal` | Sparse PUT, with explicit Global default adapter |
| `DeleteByIdentity` | Bodyless DELETE |
| `AssignToUser`, `AssignTenantDefault`, `AssignToGroup` | Native generated scoped grant operations |
| `Assign` | One scope with mapped Rank-zero omission |

All five baseline exports use remoting/modern mappings. Calls share forest
discovery, Teams audience, command/correlation headers and replay controls.
ConfigAPI remains the credential destination; the discovered forest is a
routing header. No unconditional MPACmdlet, version query or PowerShell/TRPS
runtime fallback is introduced. PowerShell's default modern selection is
command/flight-dependent and tenant deployment remains separately qualified.

`Settings` has optional Description, AllowCallPark, ParkTimeoutSeconds,
PickupRangeStart and PickupRangeEnd. Nil omits a setting. Explicit empty, false,
zero and signed values stay present. The three numeric fields use Int64 pointers
through encoding and decoding; they never pass through float64 or native-size
int. PickupRangeStart/End remain independent fields. No local pickup ordering,
range or timeout rule is invented from permissive serializer observations.

The isolated public binder turns null Description into empty and null Int64
fields into zero; null Boolean and Int64 overflow fail. Native typed arguments
make the accepted wire values explicit. Numeric-string shell coercion does not
require a native PowerShell runtime. The serializer retains signed Int64 limits
and 9,007,199,254,740,993 exactly; this does not prove service acceptance.

Response Policy retains RawJSON/RawFields and projects known fields only when
the whole property is projectable. Missing/null values retain raw presence;
future/overflow shapes remain inspectable without projecting them as zero.
Marshaling a received policy returns the original raw snapshot. Write models
include only explicit known settings and do not copy unknown response data.

Create retains the requested identity and ReadBackError.CreationResponse if POST
succeeds but the separate read fails. Update sends a sparse PUT without reading
first, including identity-only updates. The wire form does not establish backend
replacement/default semantics. No successful mutation is automatically repeated,
deleted, restored or polled; mutations use ReplayNever and reads ReplaySafe.

PolicyName is an explicit native string argument. Empty is the source clear
form; source bound null string binding becomes empty. Assign omits Rank nil/zero;
AssignToGroup can send explicit native priority zero. Tenant force is an option
of the mapped helper/DTO, although this public Grant wrapper does not expose it.
Bound PassThru true/false is rejected by the mapped modern function.

Group Accepted means the actual POST returned 202, independently of completion.
Discovery 202 cannot establish acceptance. Malformed/empty accepted POSTs retain
acceptance with a decoding error; 204 remains distinct and raw future response
fields are retained. Identity/assignment paths use one complete escaped segment;
tests normalize equivalent Go/.NET colon spellings per segment once while
retaining encoded slashes. A server-specific byte-spelling rule is not claimed.

Actual Global reads returned 403 with client-secret and certificate credentials
on 2026-11-21. Authorized reads, cloud/flight availability, service range/timeout
rules, CRUD read-back/cleanup and effective parking/grants remain unqualified.
See [unit tests](service_test.go)
and [lab acceptance](../../../docs/acceptance/call-park-policy-lab-acceptance.md).

## Owned CRUD acceptance

`TestAccCallParkPolicyCRUD` creates a uniquely named unassigned policy, verifies
non-default read-back, updates the description, verifies omitted settings,
checks list membership, deletes the policy and confirms GET404 plus list absence.
Created/updated reads are compared with MicrosoftTeams 8.1.0. Certificate and
client-secret lab runs passed. CI runs both authentication modes and requires
verified cleanup and two PowerShell comparisons for each workflow.

Seven actual acceptance responses live in `testdata`; `TestUnitOwnedCRUDUsesAcceptanceResponses`
loads them to exercise requests, typed fields, update behavior, native statuses
and deleted-resource errors. The description-only update preserves `AllowCallPark: true` and `ParkTimeoutSeconds: 180`. User/tenant/group grants, clear operations, effective client behavior,
delegated authentication and other clouds remain unqualified.
