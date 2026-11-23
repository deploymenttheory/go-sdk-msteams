# Teams compliance recording

Use `sdk.TeamsAdminAPI.ComplianceRecording`. One scope-bound service manages
compliance recording policies, the recording applications attached to them and
policy grants. Construction performs no I/O.

| PowerShell | Method | Transport |
| --- | --- | --- |
| `Get-CsTeamsComplianceRecordingPolicy` | `GetPolicyByIdentity`, `ListPolicies`, `ListPoliciesFiltered` | Compiled GET `TeamsComplianceRecordingPolicy`, Global first, local Identity/Tag filter |
| `New-CsTeamsComplianceRecordingPolicy` | `CreatePolicyByIdentity` | Compiled POST to the collection |
| `Set-CsTeamsComplianceRecordingPolicy` | `PatchPolicyByIdentity`, `UpdatePolicyByIdentity` | One sparse PATCH; Update reads first and applies list edits |
| `Remove-CsTeamsComplianceRecordingPolicy` | `DeletePolicyByIdentity` | Compiled DELETE with Identity/optional Force body |
| `Grant-CsTeamsComplianceRecordingPolicy` | `AssignToUser`, `AssignTenantDefault`, `AssignToGroup`, `Assign` | Modern ConfigAPI grant with forest discovery |
| `Get-CsTeamsComplianceRecordingApplication` | `GetApplicationByIdentity`, `GetApplicationByID`, `ListApplications`, `ListApplicationsFiltered` | Compiled GET `TeamsComplianceRecordingApplication` |
| `New-CsTeamsComplianceRecordingApplication` | `CreateApplicationByIdentity`, `CreateApplicationByID` | Compiled POST; body carries Identity and Id |
| `Set-CsTeamsComplianceRecordingApplication` | `PatchApplicationByIdentity`, `UpdateApplicationByIdentity`, `UpdateApplicationByID` | Sparse PATCH; Update reads first |
| `Remove-CsTeamsComplianceRecordingApplication` | `DeleteApplicationByIdentity`, `DeleteApplicationByID` | Compiled DELETE |
| `New-CsTeamsComplianceRecordingPairedApplication` | `NewPairedApplication` | Local builder, no I/O |

## Identities

Policies use `Global` or `Tag:{name}`; create accepts a bare name and the
service tags it. An application's identity is `{policy}/{id}`, where `id` is the
recorder's resource account (application instance) object ID. The `ByID` forms
build that identity from `-Parent` and `-Id`. Application and paired IDs must be
GUIDs; other values fail with `ErrInput` before any request.

## Settings

`PolicySettings` and `ApplicationSettings` use `policy.Field`: unset omits,
`policy.Null` sends null and `policy.Set` sends the value. A set
`ComplianceRecordingApplications` or `ComplianceRecordingPairedApplications`
replaces the whole list; a set nil list sends `[]`.

The module's list parameters are PSListModifiers. `PolicyUpdate.Applications`
and `ApplicationUpdate.PairedApplications` take a `ListEdit` (Replace, Add,
Remove by Id, case-insensitive) applied to the value read before the PATCH. A
list edit and a set list together are rejected.

`-Priority` on the application cmdlets is accepted by the module but never
copied into the request, so it has no SDK equivalent.
`ConcurrentInvitationCount` is a `uint32`; the service rejects a value above one
together with paired applications (40013).

## Service behaviour seen in the lab

- An application whose Id is not a resource account is rejected with 400 /
  40013 `ErrorComplianceRecordingApplicationIdentityNotAValidAppInstance`.
- Reads, updates and removes of a missing policy or application are 404 /
  40407. A failed preparation read in `Update…` sends no PATCH.
- Policy create, update and delete are accepted with app-only credentials.

Update results keep `Before`, `ReadResponse`, `PatchInvoked` and `HTTPAccepted`
on partial failure. Models keep `RawFields` and `RawJSON`, and their `String`
and `GoString` output is redacted.

Grant follows the shared modern pattern: user PATCH
`Skype.Policy/users/{user}/policies/TeamsComplianceRecordingPolicy`, tenant
PATCH `Skype.Policy/tenants/policies/…` and group POST
`Skype.Policy/groupPolicyAssignments/{group}/policyTypes/…` (202 or 204). An
empty policy name clears a user or group grant; `Assign` omits a zero group
Rank like the wrapper.
