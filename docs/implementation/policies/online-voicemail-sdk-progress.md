# POL-25 online voicemail SDK candidate

Native methods cover issue #98's five baseline commands and eight observed
parameter sets. The qualification issue remains open, implementation in-progress
and contract metadata-only. No successful tenant behavior or full parity is claimed.

## Resource and protocol boundaries

TeamsAdminAPI.OnlineVoicemailPolicies exposes compiled direct Get/List/ListFiltered,
Create/DeleteByIdentity, native sparse PatchByIdentity, explicit UpdateByIdentity
read/modify/PATCH and optional CreateWithGlobalComparison. Settings cover all thirteen
New/Set fields with omission/null/false/empty retained. No audio-file bytes are read
or uploaded. Read snapshots preserve received presence/future data without client
recording/language defaults. TimeSpan preserves signed 100ns ticks across Int64
extrema and constant wire strings; Go duration conversion rejects sub-tick loss.
Broader CLR/PowerShell parsing/binding and service constraints remain unqualified.

Assign/AssignUser/AssignTenantDefault/AssignGroup use a separate selected generated
adapter with lazy private forest discovery and target-header routing. It shares the
bound Teams session but never reuses compiled Grant DTOs or MPACmdlet/bootstrap rules.
Route values use strict unreserved escaping. The forest is a header, never a token
forwarding destination. A bound name and exactly one explicit scope are required.
Native DTO null omits policyName; empty sends an empty string; native zero Priority
is retained. AssignWithPowerShellBinding explicitly converts bound null to empty
and zero Rank to omission. Typed Go data/metadata replaces an optional display flag;
exported PassThru-to-modern-helper binding remains a recorded gap.

202 group acceptance retains an optional capitalized OperationId. Missing IDs stay
missing; malformed accepted data retains acceptance/response and ErrDecode. The
additional GetAssignmentOperation reads lower-camel status/count fields and raw future
and user data. It claims no full implementation of separately owned batch-operation
commands. Status/timestamps remain open text; no terminal/wait/propagation assumption
is added. All writes use ReplayNever; CRUD composition retains preparation/acceptance
on failure without automatic read-back, rollback or atomicity guarantees.

## Pinned executable evidence

[Source research](federation-branding-voicemail-evidence.md) records package/assembly
hashes and exact compiled/generated/mapped methods. Production Go never loads the
PowerShell module, .NET or fixtures. The model capture has 16 observations covering
signed ticks, precision, days, extrema and null through pinned OnlineVoicemailPolicy
and Newtonsoft serialization. No client recording-length range or default is inferred.

The compiled provider capture checks Core/PolicyRp/host hashes and calls actual
HttpRequestHelper methods with synthetic InvocationInfo. A credential-refusing host,
restored client/version/host state and memory-only sender ensure no auth/network.
MPACmdlet comes from factory source. Its 13 observations cover CRUD/list, all fields,
Force false, tick/null durations, sparse/empty/denied/wrong-envelope and decode failure.
No cmdlet constructor/ProcessRecord or generated forest pipeline is executed.

The generated capture checks the private assembly hash and injects ISendAsync and
synthetic callbacks into actual generated methods. Fourteen observations settle
user/tenant/group DTO null/empty/zero, 202/204 acceptance and separate operation reads.
These are before host/header rewrite; no HttpClient, credential or pipeline is used.
A fresh untouched DTO represents actual native null: assigning PowerShell $null to
its typed string property would itself coerce to empty and conflate the two stages.

The isolated binding capture executes the unmodified pinned Grant-CsTeamsPolicy
function for OnlineVoicemailPolicy under synthetic host/internal command stubs.
Eleven observations cover scope/name/rank binding and rejected PassThru switches.
Nine supported/validation cases feed root-client composition tests. The exported
Trps flight/bridge and production HTTP pipeline are outside this isolated oracle.
Go CI consumes these goldens and does not run PowerShell or access credentials.

## Verification and outstanding qualification

Root-client contract tests cover both synthetic token modes, request path/body/
headers, discovery/cache and credential audience, exact list filtering, all sparse
request fields, TimeSpan, bound helpers, validation-before-tokens, cancellation,
wrong scope, denials/error redaction, non-replay and partial acceptance. Existing
compiled header spy includes voicemail. New/Set field inventories match pinned
metadata; examples compile. Reports retain structural data only and do not promote
full command coverage or live results.

TestLabOnlineVoicemailReadOnly ran with dedicated-SPN secret and certificate;
both Global requests returned resource HTTP 403. List and optional pinned read-only
PowerShell scalar comparison were not reached. No successful read or restriction
on mode/feature eligibility is inferred. The [lab plan](online-voicemail-lab-acceptance.md)
records separate owned policy CRUD and consequential assignment/Global scenarios.
Those writes, feature preview, effective behavior, cloud/RBAC availability and full
binding/coercion remain unrun qualification work. No tenant changes were made.
