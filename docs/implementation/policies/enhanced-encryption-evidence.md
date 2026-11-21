# Teams enhanced encryption policy evidence (POL-44 / #117)

MicrosoftTeams 8.1.0 exports five Functions backed by private TrpsPSCmdlet classes.
All five pinned CmdletConfig mappings select modern configuration/grant functions
with ConfigType/PolicyType TeamsEnhancedEncryptionPolicy. The Core class/provider
does not establish the selected exported path. Get has Filter/Identity sets;
Grant has GrantToGroup, GrantToTenant and Identity; New/Set/Remove have Identity,
for eight exact command/set pairs.

## Captured source seams

capture-enhanced-encryption-wrappers.ps1 verifies the pinned wrapper source and
all five mappings, then executes unchanged mapped AST functions under synthetic
telemetry/in-memory internal-command seams. Its fixture contains 22 observations
for ordering/filtering, whitespace, create/read-back, sparse/default updates,
delete and grants. Five expected failures cover create failure, missing grant
scope/name and both PassThru forms. No exported Trps processing, flights,
credentials or network run.

capture-enhanced-encryption-requests.ps1 pins the private assembly and invokes
actual generated configuration/grant helpers with actual DTOs and a memory sender.
Seventeen observations cover list/read and escaped identity, POST 201, sparse PUT
204 including Identity-only and empty strings, bodyless DELETE 204, failure
decoding, pre-cancellation and grants. All non-canceled cases send once; the
pre-canceled case sends zero times. Session discovery/header rewrite and the
production pipeline remain separate from this source request seam.

capture-enhanced-encryption-binding.ps1 pins export/private sources, extracts
actual New/Set declarations into a renamed function with a bound-argument body,
and performs actual pure Xds dictionary serialization. Twenty-two observations
yield 20 successes and two mandatory New Identity validation failures. Optional
strings bind null to empty and coerce numeric/Boolean input to strings. This seam
does not execute the original export lifecycle or runtime attribute processor.

The same artifact separately records actual private calling-mode ValidSet metadata
and invokes only the inherited property-bag validator after verifying its exact
pinned empty IL body (nop, branch-to-next-instruction, ret). An uninitialized
command avoids constructor/lifecycle execution; the empty method leaves future
calling-mode text unchanged. No authentication/networking path is entered.

Promoted fixtures are enhanced-encryption-wrappers.json, -requests.json and
-binding.json under contracts/endpoints/8.1.0/fixtures. Microsoft binaries and
decompiled source remain private; capture tools/PowerShell are development-only.

## Semantics and metadata distinction

The public settings are Description, CallingEndtoEndEncryptionEnabledType and
MeetingEndToEndEncryption, all strings. Calling-mode private metadata lists
Disabled and DisabledUserOverride. That metadata is distinct from public
declaration binding and the modern property-bag validation path: neither of
those captured seams imposes that set. Future calling/meeting strings and long
Description pass the isolated source paths. No server or legacy enum behavior
is inferred from these observations. The export's runtime message helper handles
preview/breaking-change messages; it does not turn this metadata into ValidateSet.

Modern create has a one-POST primitive and explicit POST/Get composition; failure
stops read-back. PUT is sparse and permits Identity-only; omitted Identity maps
to Global in the mapped helper. Whitespace Get/filter selects the collection,
with stable case-insensitive Global-first ordering. Typed native explicit
identity operations should reject invalid selectors and expose List/default
operations separately.

User/tenant assignments PATCH and require HTTP 200. Group assignment POST
distinguishes 202 acceptance from 204. Null PolicyName becomes an empty string
through the mapped string binder. The wrapper omits absent/zero Rank while the
generated DTO can send explicit zero priority; negative Rank also reaches the
mapped helper. Public PassThru does not bind to the captured mapped grant
function. Tenant helper Force is Boolean, but this public family exposes no
Force parameter. No effective clear, polling or assignment guarantee follows.

## Native implementation checkpoint and remaining qualification

TeamsEnhancedEncryptionPolicies now has native modern CRUD/list, explicit
create/read-back partial results and user/tenant/group grants, all three typed
settings, owned Xds projections and root wiring. Initial race tests compare all
17 generated request observations in synthetic application/delegated modes plus
four modern filter cases; model tests preserve known/unknown values, owned raw
snapshots and redacted formatting. Metadata hints remain separate from proven
runtime validation. Group results retain generated acceptance statuses, and
Assign distinguishes source Rank omission from explicit native priority.

Before the implementation PR, finish direct source binder/wrapper comparisons,
partial create/read-back and assignment/discovery failure coverage, no-replay and
context/input guards, compiling examples/API docs, gated read-only lab support,
full repository checks and reviewed sanitized artifacts. Exact ledger mappings
remain in-progress/metadata-only; this checkpoint is not full qualification.

No tenant reads/writes have run for this family. Successful authorization,
server/cloud/flight availability, write/read-back/cleanup, legacy behavior and
effective encryption/assignment behavior remain unqualified. Issue #117 stays
open; the captures do not establish full command or tenant qualification.
