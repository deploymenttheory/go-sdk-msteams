# Application-access-policy SDK evidence

Implementation work date: 2026-11-20. Workstream: POL-20 / issue 93.
The primary service has resource Create/GetByIdentity/UpdateByIdentity/
DeleteByIdentity/List, local wildcard-filtered listing, and separate user,
tenant-default and group assignment actions. The root collection injects the
ConfigAPI backend and lazy per-client forest discovery. Constructors do no I/O.

## Contract verification

Colocated tests consume the pinned
[request/dispatch artifact](../../../contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json)
for all eight wire primitives. They check the actual SDK transport destination,
Teams token audience, one-time same-origin discovery, normalized target header,
segment escaping, methods, JSON bodies and response dispatch. Additional cases
cover omitted/empty/delta/legacy AppIds, unknown response fields, failed creation
read-back without replay/automatic delete, malformed successful responses,
validation/cancellation without I/O and local Identity/Tag wildcard matching.

Discovery tests separately cover failure metadata, invalid/missing endpoints,
absence of a constant-host fallback, caller-header ownership and deterministic
queued cancellation. Go wildcard examples were checked against local PowerShell
string matching; portable Unicode case folding does not claim every host's
locale-specific PowerShell collation.

`go test -race -count=1 -shuffle=on ./...`, `go vet ./...`, `go build ./...` and
`golangci-lint run ./...` passed locally. The full test command required normal
loopback listener access for existing HTTP/TLS fixtures; tenant scenarios were
explicitly disabled. Contract inventory/resource and coverage validators passed.
No command/parameter-set ledger row is promoted by this implementation progress:
legacy/conditional scope and remaining qualification must still be reconciled.

## Actual tenant observations

These observations used the existing dedicated lab credentials. No credentials,
tokens, tenant IDs, policy contents or response bodies are included here.

| Actual check | Observed result |
| --- | --- |
| Dedicated principal's current Graph assignments | Organization.Read.All only |
| Dedicated principal's directory role assignments | None |
| Direct ConfigAPI service discovery | HTTP 200 |
| Direct application-access-policy collection read | HTTP 403 |
| New SDK List using client-secret authentication | HTTP 403, classified forbidden |
| New SDK List using certificate authentication | HTTP 403, classified forbidden |
| Existing SDK Graph organization probe | HTTP 200 |
| Existing SDK Graph users probe | HTTP 403 |
| Pinned PowerShell 8.1.0 certificate connection | Succeeded |
| Pinned Get-CsApplicationAccessPolicy / Get-CsTeamTemplateList | Failed; HTTP status not retained by this initial sanitized probe |

These are live authorization observations, **not successful management CRUD**.
The role inventory and Microsoft's application-onboarding guidance identify the
missing directory-role prerequisite; a generic 403 alone would not establish its
sole cause. A concrete request to assign Teams Administrator at tenant scope to
the dedicated lab SPN is pending approval. An earlier broader provisioning helper
was rejected by automatic approval review; it was removed without execution.
No extra Graph grants, directory-role assignments or policy mutations occurred.

Sources: [application authentication](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication),
[Teams roles](https://learn.microsoft.com/en-us/microsoftteams/using-admin-roles),
[versioned resource contract](../../../contracts/endpoints/8.1.0/application-access-policies.json).

## Live acceptance prepared, not yet executed

`TestLabApplicationAccessPolicyCRUD` requires explicit mutation opt-in and a
private recovery directory. It creates one random unassigned policy, records
intent before creation, checks non-default create/update read-back and omitted
field retention, lists it, deletes it and checks absence. Tenant/identity-bound
cleanup reconciles lost responses by the owned fixture name; unresolved cleanup
retains private state. It does not edit Global or assign policies to users,
groups or the tenant.

The optional pinned PowerShell oracle reads that same owned fixture after create
and update and compares Identity, AppIds and Description in memory. It verifies
module version/assembly identity before connection. Live create/update/delete and
successful PowerShell/Go field parity remain **not run**, pending lab access.

Chrome DevTools was also attempted for the user's signed-in Teams session; its
server could not read Chrome's debugging-port file (permission denied). Browser
inspection is unavailable until that connection is repaired; no browser change
or browser-derived service qualification is claimed.

## Mapped-wrapper and complete service evidence

The guarded [wrapper oracle](../../../tools/contracts/capture-policy-wrappers.ps1)
executes six unmodified function definitions from the hash-checked pinned source
under explicitly synthetic telemetry/host types and an in-memory internal command
module. It imports no MicrosoftTeams module, connects nowhere and has no HTTP
sender. Its [19 observations](../../../contracts/endpoints/8.1.0/fixtures/policy-wrapper-compositions.json)
confirm stable case-insensitive Global-first output, Identity/Tag filtering,
POST-then-GET creation, stopping after creation failure, the omitted-Identity
Global update default, empty user-clear policy names, tenant force false and Rank
omission/forwarding. Negative Rank is forwarded; a backend range is not invented.
Missing scope/PolicyName and explicit true/false PassThru reject at this mapped
function boundary. The exported TrpsPSCmdlet, flight selection and generated HTTP
pipeline were not executed; those boundaries are not silently promoted.

These observations exposed and fixed missing Global-first ordering in the Go
list projection. Raw HTTP bodies and within-partition order remain retained.
The policy operations now also supply explicit public-command operation labels
in X-MS-CmdletName, consistent with the existing template SDK convention. This
is native operation metadata, not a reconstructed PSInvocationInfo value or a
claim that the server requires exactly that string. Discovery stays unlabelled,
and the same-origin token/forest header boundary is unchanged.

The shared root-client runner now exercises all five commands/eight sets against
actual pinned request/status observations. Eleven positive CRUD/assignment
executions, three source-wrapper listing cases and twelve mutation failure cases
emit exact-name sanitized reports. The failure cases use the default transport
policy, so received 429/503 cannot be hidden by setting a one-attempt test policy.
All six mutation entry points make one attempt, and failed creation performs no
read-back. Set coverage is checked against the validated ownership ledger.
Go-symbol/test links now identify every set; implementation remains in-progress,
contract metadata-only and live not-run. These reports are candidate offline
evidence, not a promotion of complete command or tenant parity.
