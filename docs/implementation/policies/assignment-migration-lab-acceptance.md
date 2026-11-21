# Assignment migration lab acceptance

MSTEAMS_LAB_READ_ONLY=1 and MSTEAMS_LAB_IDENTITY_FILE enable
TestLabAssignmentMigrationReadOnly using the existing dedicated identity.
MSTEAMS_LAB_AUTH_METHOD selects client-secret or certificate (default). The test
lists status and reads returned job IDs. It never starts migrations or clears
assignments. Both actual credential-mode attempts returned resource HTTP 403;
read-by-ID and PowerShell comparison were not reached. No IDs/payloads are retained.

Optional MSTEAMS_LAB_POWERSHELL_ORACLE=1 with the absolute pinned module directory
enables certificate-authenticated Get-DirectToGroupAssignmentsMigrationStatus.
The script checks version/assembly hashes and requested tenant binding, rejects
ordinary output strings as errors, and projects job IDs/states. Go bounds/clears
transient output and requires a complete tenant-bound shape. Numeric enum states
render to pinned names; unknown states remain open. Sequential reads can differ
while jobs change. This projection proves neither full response equality nor
completion/effective assignments. Recommendation/conflict acceptance needs explicit
tenant-bound EntityType/PolicyType/GroupID fixtures and remains unrun.

Start/Clear change effective assignments. Their future acceptance must require
read-only + mutation + consequential opt-ins, a private tenant/identity-bound journal,
an owned group with controlled members and non-default policies, explicit
allow-non-empty decisions, assignment snapshots and a recovery plan. Record intent
before submission; investigate ambiguous results instead of blindly resubmitting.
Observe jobs and effective assignments separately with a bounded explicit wait.
Verify cleanup preserves unrelated assignments and the group. No general rollback
or job Delete is established. These mutation scenarios are not implemented or run;
the qualification issue remains open.
