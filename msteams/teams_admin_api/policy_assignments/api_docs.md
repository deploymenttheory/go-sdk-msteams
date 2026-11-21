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

This is an ordinary-protocol implementation checkpoint for issue #100. Eight OCE
commands, their fifteen conditional sets, privileged-mode eligibility, binding
details, cloud support and successful tenant CRUD are still unqualified. Their
distinct regional/direct protocol must not be routed through the generated job
adapter. [Pinned evidence](../../../docs/implementation/policies/policy-assignments-evidence.md)
records the exact captures and remaining work.
