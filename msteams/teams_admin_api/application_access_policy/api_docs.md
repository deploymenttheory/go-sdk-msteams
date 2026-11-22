# Application access policies

`sdk.TeamsAdminAPI.ApplicationAccessPolicies` manages the policy resource through
the selected ConfigAPI backend. Constructing the SDK/service performs no network
I/O. The first operation discovers the tenant forest at
`GET /Teams.Tenant/serviceDiscovery`; subsequent requests carry its normalized
HTTPS root in `X-MS-Target-Uri`. Requests and bearer tokens continue to use the
selected ConfigAPI origin/audience, including when discovery returns another host.
Each management request labels its native operation with the corresponding public
command in `X-MS-CmdletName`; this mirrors the existing template SDK convention,
without reconstructing a PowerShell invocation. Discovery has its own headers.

| Go operation | HTTP contract | PowerShell behavior |
| --- | --- | --- |
| `Create` | POST collection, then GET item | `New-CsApplicationAccessPolicy`; read-back failure preserves the successful creation receipt |
| `GetByIdentity` | GET item | `Get-CsApplicationAccessPolicy -Identity` |
| `List` | GET collection; bare array | `Get-CsApplicationAccessPolicy` collection read |
| `ListFiltered` | List then local Identity/Tag matching | Get's Filter set; no invented server query parameter |
| `UpdateByIdentity` | Sparse PUT item | `Set-CsApplicationAccessPolicy`; pass `"Global"` explicitly to reproduce omitted Identity's default |
| `DeleteByIdentity` | DELETE item, no body | `Remove-CsApplicationAccessPolicy` |
| `AssignToUser` | PATCH user policy | Grant Identity scope; empty policy name emits the user-clear request |
| `AssignTenantDefault` | PATCH tenant policy | Grant Global scope; force is always a JSON Boolean |
| `AssignToGroup` | POST group assignment | Grant Group scope; returns 202 acceptance/operation ID separately from 204, without implicit polling |

Identity/user/group inputs remain separate path segments. Colons can have a
different valid percent-encoding spelling from the pinned .NET client; encoded
slashes must never become path separators or be decoded twice.

`CreateRequest` and `UpdateRequest` are separate from the `Policy` read model.
Nil request fields remain absent. A pointer to an empty description sends `""`.
`ReplaceAppIDs(nil)` sends `[]`; `ChangeAppIDs(nil, nil)` sends `{}`. Constructors
copy input lists. Legacy object replacement/delta constructors remove the pinned
serialization metadata fields and retain data properties/false values. Their
local acceptance does not prove the backend accepts object-shaped application IDs.
Read models retain every field in `RawFields`, including unknown/non-string
AppIds shapes, rather than silently losing future response additions.

Assignments are actions, not policy updates/deletes. Native group assignment sends
an explicitly supplied zero priority; the pinned PowerShell adapter omits zero
Rank, so pass nil to mirror that adapter. Clearing user, tenant and group scope
must not be treated as interchangeable operations. Tenant/group clear acceptance,
rank constraints, propagation and async correlation remain service qualification
items. No mutation is automatically replayed or rolled back after an ambiguous
result. Read operations opt into the shared bounded retry policy.

Local filtering supports case-insensitive `*`, `?`, bracket sets/ranges and
backtick escapes. UTF-16 character positions preserve the observed string
wildcard behavior for surrogate pairs. Native matching uses portable Unicode
case folding; locale-specific PowerShell collation remains an explicit
compatibility limitation. See [PowerShell string wildcards](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_wildcards?view=powershell-7.5).
Both list methods place case-insensitive Global identities first, preserving
the relative order within each partition and all raw snapshot/response data.

Authentication uses the selected Teams/ConfigAPI resource. Microsoft documents
`Organization.Read.All` for module onboarding and an appropriate Teams directory
role for application management; downstream meeting-app Graph permissions are
separate. No Teams resource API permission or Global Administrator grant is
required by this implementation. See [Microsoft application authentication](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication)
and [Teams role capabilities](https://learn.microsoft.com/en-us/microsoftteams/using-admin-roles).

The pinned client's modern flight/legacy selection remains a compatibility
boundary; this service implements the evidenced native modern HTTP contract.
It does not invoke PowerShell at runtime, implement a remote runspace fallback,
or infer tenant eligibility from a missing flight or generic 403/404.

See the [request tests](service_test.go) and
[acceptance response tests](response_test.go).
Unit tests check ConfigAPI routing and token audience. Successful response
fixtures come from the service acceptance tests. Tenant coverage remains tracked
in the coverage ledger.

The opt-in `TestAccApplicationAccessPolicyCRUD` creates only a random, unassigned
policy, journals intent before creation, verifies non-default create/update
read-back, checks list membership, deletes it and verifies absence. Cleanup is
tenant/identity-bound and leaves unresolved private recovery data on failure.
It never grants the policy to an existing user/group/tenant or edits Global.
Set `MSTEAMS_LAB_READ_ONLY=1`, `MSTEAMS_LAB_MUTATIONS=1`,
`MSTEAMS_LAB_JOURNAL_DIRECTORY` and `MSTEAMS_LAB_IDENTITY_FILE` explicitly to run it.
`MSTEAMS_LAB_AUTH_METHOD` selects certificate (default) or client-secret.
`MSTEAMS_LAB_POWERSHELL_ORACLE=1` plus `MSTEAMS_LAB_MODULE_PATH` adds actual
PowerShell 8.1.0 reads of the same owned policy after create/update and compares
Identity, AppIds and Description in memory. Tenant values/credentials are not
written to repository artifacts or test logs.

`TestAccApplicationAccessPolicyUserAssignment` additionally qualifies user grant
and clear on the dedicated SDK-owned account. It compares the initial empty
state, assigned policy/source and restored state against MicrosoftTeams 8.1.0,
journals restoration before grant, then independently verifies restoration and
policy deletion. Certificate and client-secret runs passed; the account has no
license or administrator role. This qualifies assignment management, while
online-meeting access and client effects remain separate scenarios.

The unit assignment workflow loads the actual grant/clear replies and three
assignment-state responses from `policy_assignments/testdata`. Tenant/group grant
and clear, rank/async behavior and remaining parameter variants still need live
qualification on this ticket.

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
