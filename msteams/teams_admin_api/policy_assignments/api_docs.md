# Policy assignments

Use `client.TeamsAdminAPI.PolicyAssignments` for native Go operations. Construction
performs no I/O. Runtime use needs no PowerShell installation or module import.
The service uses the session's configured Teams origin/audience and shared HTTP,
authentication and error handling. Generated forest discovery is cached within
this service's identity-bound adapter; its result is a routing header, never a
new credential-bearing destination.

| Method | Native operation / baseline mapping |
|---|---|
| GetGroup | Read one group and policy type; Get-CsGroupPolicyAssignment / Get |
| ListGroup | List one group's assignments; Get-CsGroupPolicyAssignment / Get1 |
| ListGroupsByPolicy | List a policy type, or all when empty; Get-CsGroupPolicyAssignment / Get2 |
| CreateGroup | PUT group assignment; New-CsGroupPolicyAssignment / NewExpanded |
| UpdateGroup | PATCH settings using the pinned generated GroupSetPolicyAssignment helper; additional operation, no exported baseline command claim |
| DeleteGroup | DELETE group assignment; Remove-CsGroupPolicyAssignment / Remove |
| GetUser | Read effective user policies and their sources; Get-CsUserPolicyAssignment / Get |
| GetUserByIdentity | Same operation using UserIdentity.Identity; Get-CsUserPolicyAssignment / GetViaIdentity |
| ListBatches | List generated jobs with an optional status query; Get-CsBatchPolicyAssignmentOperation / Get |
| GetBatch | Read a generated job and per-user outcomes; Get-CsBatchPolicyAssignmentOperation / Get1 |
| CreateBatch | Submit a batch once; New-CsBatchPolicyAssignmentOperation / NewExpanded |

`GroupSettings` follows the generated DTO serializer: unbound and null values are
omitted; bound empty policy names and zero/negative priorities remain present.
The low-level native create/update methods can send an empty DTO. PowerShell's
exported group creation additionally requires binding a policy name; complete
binding/coercion qualification remains open. Go returns typed data and response
metadata in place of the shell's optional PassThru display switch.

`BatchCreateRequest` requires a supplied identity slice; `[]string{}` is an allowed
empty collection, while nil is an unbound required argument. Order and duplicates
are retained. PolicyType and PolicyName are independent bound values inside the
`policy` object. An empty OperationName omits the query. No hard-coded help-text
policy allowlist, user-count limit, Graph lookup or client-side precedence rule is
inferred from the module description.

AdditionalParameters accepts JSON dictionary values and explicitly follows the
pinned generated JSON dictionary projection: object null members are omitted at
every depth, while a null inside an array becomes the string `"null"`. False, zero,
empty strings/collections and raw numeric precision are retained. Malformed JSON
fails before discovery or credential acquisition. PowerShell Hashtable coercion
and non-JSON CLR inputs remain separate qualification work.

Create accepts 201/204, update and delete accept 202/204, and batch submission
accepts 202. `OperationResult.HTTPAccepted` records a recognized success even when
its body cannot be decoded. A missing/null OperationId stays nil; a malformed ID
does not produce a usable handle. An empty ID remains an explicit server value.
No wait, terminal-state classification, implicit read-back, retry of a mutation,
rollback or effective-policy convergence claim is included.

Group/user/batch reads retain nullable fields, raw presence, nested future data and
timestamp strings. Treat read models as snapshots; marshaling a received snapshot
returns its original raw object. Use request types to change a resource. Effective
sources and batch outcomes are observations: their ordering does not establish a
local precedence algorithm. Bare-array responses do not imply a cursor or paging
envelope. Invalid/null read envelopes return ErrDecode with response metadata.

All methods propagate context and return Resty response metadata. Errors omit
service messages from diagnostic formatting; inspect response bodies deliberately.
401/403 do not establish unsupported authentication or a missing tenant feature.

## Conditional OCE operations

All OCE methods require an explicit OCEContext.TenantID in canonical Guid form.
Environment defaults to the profile matching the session's fixed origin. Explicit
environments must match that origin and cannot change credential destinations.
The eight source profiles and their regional order are retained. Custom sessions
can supply the corresponding trusted origin/authority/audience tuple. Routing
availability is not evidence of contributor eligibility, auth-mode or cloud support.

| Methods | Conditional baseline command |
|---|---|
| GetOCEBatchStatus, ListOCEBatchStatuses | Get-CsBatchOperationStatus |
| GetOCEBatchDefinition | Get-CsBatchOperationDefinition |
| ReprocessOCEBatch | Invoke-CsReprocessBatchOperation |
| GetOCEGroup, ListOCEGroup, ListOCEGroupsByPolicy, ListOCEGroups | Get-CsGroupPolicyAssignments |
| ListOCEGroupMembers | Get-GPAGroupMembers |
| ListOCEUserMembership | Get-GPAUserMembership |
| RefreshOCEGroupUsers | Invoke-CsRefreshGroupUsers |
| ReprocessOCEGroupAssignment | Invoke-CsReprocessGroupPolicyAssignment |

The direct OCE sender adds target-tenant and correlation headers. It never performs
forest discovery or adds MPACmdlet/generated command-name headers. One correlation
is shared across a regional operation; a new operation gets a fresh UUID. Each
request makes one attempt, including GET, so regional fallback retains failures.

An explicit region selects one batch request. Nil on a status/definition lookup
searches the source region order and stops at the first HTTP 200, retaining earlier
failed attempts. Nil on a status list gathers every region, retaining partial values
and per-attempt errors; the returned error joins failures. A malformed successful
response stops processing. Native Region annotations identify the actual request,
while RawFields retains the original wire object separately.

OCE GPA reads/actions accept every successful 2xx status, matching the pinned
generic helper's IsSuccessStatusCode check. This does not establish which statuses
the service actually emits. GPA reference responses can be
empty or JSON null; regional snapshots/lists require a decoded object/array.
Regional status/definition reads require 200. Reprocessing requires 202 and a Guid
string, with acceptance retained if decoding fails; it does not use a generated
operation-ID object envelope. No implicit wait or completion inference is included.

Refresh always includes userId and policyType query keys. Omitted UserID becomes
Guid.Empty and omitted PolicyType becomes an empty value. All=false/true uses the
source's False/True spelling. Path/query values are encoded separately rather than
copying unsafe raw interpolation. Refresh and reprocessing are explicit consequential
actions; no read or ordinary mutation invokes them or elevates eligibility.

Issue #100 remains in-progress for full qualification. Native methods/tests map all
fourteen commands and twenty-five sets. Successful tenant CRUD, OCE eligibility,
complete shell binding/coercion and cloud behavior remain unqualified.
[unit tests](service_test.go)
records the captures, actual lab denials and remaining work.

## Dedicated user assignment acceptance

`TestAccApplicationAccessPolicyUserAssignment` reads the dedicated SDK user's
application-access assignments before grant, after grant and after clear. All
three projections match MicrosoftTeams 8.1.0 in certificate and client-secret lab
runs, and cleanup verifies the user's restored state before policy deletion.
The real responses are stored in `testdata` and consumed by
`TestUnitUserReadsUseAcceptanceResponses` for both identity entry points. A
cross-service unit workflow also validates native grant/clear requests and the
complete direct-assignment source fields against those responses.

This covers these user reads only. Group/batch/OCE operations, effective client
behavior and other policy types remain separately qualified.

### Lab group assignment qualification

`TestAccApplicationAccessPolicyGroupAssignment` passed with certificate and
client-secret authentication against the dedicated empty security group. It
exercises assignment PUT, selected GET, group list, PATCH to a second owned
policy, DELETE/GET404/list absence, and application-access group POST grant,
then restores the original empty assignment and deletes both owned policies.
Each run compares six configuration states with public MicrosoftTeams 8.1.0
`Get-CsGroupPolicyAssignment`. Actual responses supply the committed per-service
unit fixtures, including empty HTTP204 replies and PolicyRP40404 absence.

Rank 1 and the policy-name update are verified. Member propagation, effective
user precedence, cross-group rank changes, tenant-default grants and batch/OCE
operations remain separately qualified. A successful write or HTTP204 does not
imply effective-user convergence or an asynchronous operation identifier.

CreateGroup requires a bound nonempty PolicyName before credentials or discovery, matching the public New binding. Whitespace is preserved. Priority omission differs from bound zero; UpdateGroup and raw GroupSettings DTO omission/null projection remain separate native behavior.
