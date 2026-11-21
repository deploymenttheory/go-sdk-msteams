# Policy packages

Use `client.TeamsAdminAPI.PolicyPackages`. Construction performs no I/O. Methods
use native Go, the session's fixed Teams origin/audience and shared transport.
Generated forest discovery is cached per service/session; it supplies a routing
header, never a new credential-bearing destination. No runtime PowerShell is needed.

| Native method | Baseline command / set |
|---|---|
| List | Get-CsPolicyPackage / Get |
| GetByName | Get-CsPolicyPackage / Get1 |
| GetByIdentity | Get-CsPolicyPackage / GetViaIdentity |
| GetForUser | Get-CsUserPolicyPackage / Get |
| RecommendationsForUser | Get-CsUserPolicyPackageRecommendation / Get |
| Create | New-CsCustomPolicyPackage / RequiredPolicyList |
| Update | Update-CsCustomPolicyPackage / RequiredPolicyList |
| DeleteByName | Remove-CsCustomPolicyPackage / Remove |
| DeleteByIdentity | Remove-CsCustomPolicyPackage / RemoveViaIdentity |
| AssignGroup | Grant-CsGroupPolicyPackageAssignment / RequiredPolicyList |
| AssignUsers | Grant-CsUserPolicyPackage / GrantExpanded |
| SubmitBatch | New-CsBatchPolicyPackageAssignmentOperation / NewExpanded |

Microsoft's separately dated [application-authentication guidance](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication)
excludes PolicyPackage commands in the documented commercial/GCC/GCC High/DoD
profiles. The service checks that catalog evidence before discovery/token acquisition
and returns client.ErrAuthenticationMode for those application sessions. This applies
to client credentials, managed/workload identities and supplied tokens declared as
application mode. Unknown documentation profiles remain unqualified, without a
support claim or an inferred restriction. Delegated eligibility still needs tenant
qualification; acquiring a token does not grant service access.

Custom create/update requests require a bound valid Name and non-nil policy list.
An explicit empty list remains `[]`. Update is PUT to the custom-package collection
with PackageName in the body, not a sparse PATCH. Pointer fields follow generated
DTO behavior: nil omits a key, while a pointer to an empty string remains present.
PolicyPair is typed; ParsePolicyList is an optional local wrapper-compatible builder
that splits only the first delimiter and preserves empty pieces/order/duplicates.

Group rankings are typed int32 values. Rank always serializes, including zero;
absent Rankings defaults to the exported wrapper's empty list. The SDK does not
silently emulate the shell's numeric string casts. Group PackageName nil and empty
have different request bodies; no additional unassignment operation is inferred.
User IDs preserve order/duplicates and an explicit empty collection. Nil UserIDs
is an unbound required argument. User reads/recommendations use the service's `user`
header rather than inserting an identifier into a guessed URL.

AssignUsers and SubmitBatch are distinct synchronous/asynchronous operations.
Business failures can accompany a successful HTTP response. AssignmentResult retains
failed policy creation/assignment, failed users, known LegacyUser errors and unknown
raw error fields. Failed HTTP responses also retain decoded typed operation data;
discovery failures are never decoded as business results. Error classification keeps
the actual HTTP status, including the source's 520/521 cases.

HTTPAccepted indicates a recognized success status. It survives malformed success
bodies, with ErrDecode and response metadata returned separately. ResponseMessage
is untrusted data, not a success predicate. Missing/null batch OperationId stays nil;
no polling endpoint, terminal-state inference, automatic read-back, propagation wait,
mutation replay or rollback is fabricated. Create/Update/Delete and assignments
use single-send mutation requests.

Read models retain received null/absence, policy dictionaries and future fields.
Marshaling a received snapshot preserves its original object; use request types to
change data. String/GoString redact identifiers, messages and token-related data.
All methods propagate context and return Resty response metadata.

Issue #101 remains in-progress for qualification. Tests map all nine commands and
twelve sets, but full shell coercion, delegated live reads, mutation/read-back,
effective assignment and cloud/mode deployment remain unqualified. See the
[source evidence](../../../docs/implementation/policies/policy-packages-evidence.md)
and [lab plan](../../../docs/implementation/policies/policy-packages-lab-acceptance.md).
