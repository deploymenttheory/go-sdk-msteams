# Assignment migration tenant tests

The configured tenant currently returns HTTP 400 / PolicyRP 40003 for migration
status. `TestAccAssignmentMigrationCapabilityUnavailable` checks the SDK's typed
`auth.ErrTenantCapabilityUnavailable`, original HTTP diagnostics and the existing
PowerShell observation. This access-outcome test passes; it is not a successful
migration read or submission.

Ordinary JSON-backed tests cover status, recommendations, conflicts, start/clear
requests, UUID validation, response decoding, cancellation and error receipts.
The SDK sends consequential submissions once and does not invent polling,
rollback or a job-delete operation. See the [SDK reference](../../msteams/teams_admin_api/assignment_migration/api_docs.md).

Successful migration scenarios require a tenant with the feature enabled and
explicit tenant-bound policy/group fixtures. Start/clear also need consequential
mutation opt-in, controlled membership, original assignment snapshots and a
recovery plan. Observe job state and effective assignment separately. Never retry
an ambiguous submission blindly or replace the current denial test with a skip.
