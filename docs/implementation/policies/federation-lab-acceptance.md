# Federation lab acceptance

Use the existing private lab identity and the [lab lifecycle gates](../lab/lifecycle-evidence.md).
Ordinary tests read no credentials. Explicit MSTEAMS_LAB_READ_ONLY=1 enables
the four federation read tests; MSTEAMS_LAB_IDENTITY_FILE names the private identity.
MSTEAMS_LAB_AUTH_METHOD selects certificate (default) or client-secret.

Each family has its own read test, run in its service's acceptance partition:
TestAccExternalAccessPolicyReadOnly, TestAccACSFederationConfigurationReadOnly,
TestAccTeamsExternalAccessConfigurationReadOnly and
TestAccTenantFederationConfigurationReadOnly. Each checks Global and a locally
filtered list; they share labkit.RunFederationGlobal. Optional
MSTEAMS_LAB_POWERSHELL_ORACLE=1 plus the absolute pinned module path enables a
certificate-authenticated read-only PowerShell comparison of selected nullable
scalars. The oracle checks version/assembly hashes and tenant binding, accepts
only Global or an explicitly owned external-policy fixture, bounds transient
output and retains no response data. Domain-choice presentation fixups are not
mistaken for a raw wire contract.

Both actual read attempts (secret/certificate) stopped with resource HTTP 403
for all four families. Listing and PowerShell comparison therefore remain unrun.
No successful native resource or full command qualification is claimed.

TestAccExternalAccessPolicyCRUD additionally requires MSTEAMS_LAB_MUTATIONS=1
and a private MSTEAMS_LAB_JOURNAL_DIRECTORY. This prepared test is unrun. It plans
creation before sending, creates one random unassigned policy, reads non-default
Description/false values, updates Description while checking the omitted Boolean,
lists it, deletes it and requires 404 absence. Optional PowerShell reads compare
that same owned fixture after create/update. It never grants a policy or changes
Global/tenant settings. Cleanup reconciles only the tenant/identity-bound owned
name, accepts verified 404, and retains private state if recovery is unresolved.
No blind compensation follows an ambiguous submission.

Tenant federation/ACS/external-user changes and user/group/tenant assignments can
alter effective access. They remain separately planned consequential acceptance
with explicit owned fixtures, snapshots/restore and propagation decisions. The
prepared safe policy CRUD test does not stand in for these scope-specific effects.
