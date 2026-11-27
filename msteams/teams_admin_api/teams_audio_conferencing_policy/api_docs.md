# Teams audio conferencing policies

Use `sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies`. The service uses the
root client's shared Teams ConfigAPI backend and forest discovery. Constructor
injection performs no I/O and requires a backend bound to the Teams audience.

| Method | Behavior |
| --- | --- |
| `GetByIdentity` | GET one explicit, opaque identity segment |
| `List`, `ListFiltered` | Bare collection; stable Global-first and local Identity/Tag wildcard filtering |
| `CreateByIdentity` | One POST; returns HTTP creation metadata |
| `Create` | Explicit POST/Get composition matching mapped New |
| `UpdateByIdentity`, `UpdateGlobal` | One sparse PUT; Global is an explicit omitted-Identity adapter |
| `DeleteByIdentity` | One bodyless DELETE |
| `AssignToUser`, `AssignTenantDefault`, `AssignToGroup` | Native scoped grant requests |
| `Assign` | Select one scope and apply mapped group Rank-zero omission |

This family is exported through the remoting/modern path. Its Core classes are
not the selected baseline export. Updates use PUT and send server-side list
operators; there is no implicit read/edit/PATCH. Ordinary modern selection in
the PowerShell module is command/flight-dependent. The native service calls the
evidenced REST operations directly and has no PowerShell/TRPS runtime fallback.
Deployment, authorization and behavior still require tenant qualification.

`Settings.AllowTollFreeDialin` is a pointer: nil omits it and `new(false)` sends
false. The public source parameter is a non-nullable Boolean. Nil
MeetingInvitePhoneNumbers omits that field. `ReplacePhoneNumbers()` sends `[]`;
`ChangePhoneNumbers(nil,nil)` sends `{}`. Nonempty Change Add/Remove arrays are
sent to the service without a local collection read or deduplication.
`ReplacePhoneValues` and `ChangePhoneValues` retain compatible string/legacy
object entries and remove the source metadata keys. Null/numeric items fail
locally; serializing a legacy object does not prove backend acceptance. Empty,
whitespace and future strings are preserved without an invented phone regex.

Response `Policy` retains the complete received RawJSON and RawFields. Known
Boolean/list projections are populated only if the entire field is projectable;
future shapes remain available in raw data without a misleading partial value.
Marshaling a received policy returns its raw snapshot. Writes are explicit
known-field request models and do not copy unknown response data implicitly.

`Create` returns the requested identity and `ReadBackError` when POST succeeds
but the separate read fails. CreationResponse retains the successful POST for
reconciliation. No successful mutation is replayed, deleted or restored
automatically. Sparse PUT identity-only requests remain possible; the wire
contract does not establish backend replacement/default semantics.

PolicyName is a required native string argument. Empty is the source clear form;
the mapped source binder converts an explicitly bound null string to empty.
`AssignToGroup` sends an explicit nonnil Int32 priority, including zero. `Assign`
uses Rank semantics and omits nil/zero. Tenant force is an explicit native option
of the mapped helper/DTO; this public command's wrapper does not expose Force.
Public PassThru is rejected by the mapped modern function, including bound false.

GroupAssignment.Accepted means the actual group POST returned 202. Discovery
202 responses cannot establish acceptance. Accepted POSTs with a malformed or
empty body retain acceptance plus a decoding error; 204 remains distinct. The
full accepted raw response is retained, including unknown fields. No polling or
effective assignment guarantee is implied. Mutations use ReplayNever; reads use
the client's replay-safe policy.

Identity is escaped as one complete path segment. Tests normalize the valid
Go/.NET colon spellings per segment once, retaining escaped slashes. No tenant
evidence establishes a byte-spelling-specific server requirement. No version
query, unconditional MPACmdlet, client ETag or implicit setting default is added.

On 2026-11-21 actual Global reads returned 403 using client-secret and certificate
credentials. Successful authorization, cloud/flight availability, writes,
read-back, cleanup and effective invite/assignment changes remain open. See
[unit tests](service_test.go)
and [lab acceptance](../../../docs/acceptance/audio-conferencing-policy-lab-acceptance.md).

### Executed owned CRUD

`TestAccAudioConferencingPolicyCRUD` passed with certificate and client-secret authentication.
It creates an owned unassigned policy, checks non-default read-back, updates a
setting while verifying omission, checks the list and independently verifies
selected HTTP404 and list absence after deletion. Each run has two public
MicrosoftTeams 8.1.0 read comparisons and one cleaned private journal entry.
Creation disables toll-free dial-in and explicitly sets an empty invite phone
list. A sparse PUT enables toll-free dial-in while retaining that empty list.
Both modern getter fields match the pinned public ForceModern read. No phone
numbers, calls or assigned users are involved.

The actual HTTP201/204 replies, created/updated JSON, owned list and deleted404
body supply `TestUnitOwnedCRUDUsesAcceptanceResponses`. Unit checks retain exact
requests, typed settings, complete snapshots and typed errors. CI requires every
CRUD milestone, both comparisons and verified journal cleanup. Global, grants,
media behavior and remaining command/parameter variants are separate scenarios.
