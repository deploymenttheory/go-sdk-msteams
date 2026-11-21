# Deployment and migration lab acceptance

Status: prepared, not run. The dedicated lab SPN exists, but exact privilege
approvals remain pending. These cases grant nothing, create no account/team,
send no notification email and perform no migration. No ambient signed-in
browser or alternative identity is used to bypass approval boundaries.

All cases require `MSTEAMS_LAB_READ_ONLY=1` before credentials/fixtures are read.
Batch and private migration use the existing explicit private lab identity.
`TestAccBatchDeploymentStatusReadOnly` also requires private JSON at
`MSTEAMS_LAB_BATCH_STATUS_FILE`, with tenant_id and a known existing
orchestration_id. Tenant must match the selected identity. It reads one raw
status and does not poll, submit or interpret Completed as success.

`TestAccPrivateChannelMigrationReadOnly` reads the bound tenant observation and
checks the returned tenant identity. Optional `MSTEAMS_LAB_POWERSHELL_ORACLE=1`
and absolute `MSTEAMS_LAB_MODULE_PATH` enable the existing hash-guarded read-only
child for these two cases. It executes only the selected Get status command,
uses certificate app auth without prompts, bounds/transiently consumes stdout,
discards stderr and publishes no IDs/tokens/status/error contents. Batch text and
selected migration tenant/status/detail fields are compared; this does not claim
live timestamp normalization or every parameter/error branch.

`TestAccMultiGeoRegionReadOnly` requires a separate private
`MSTEAMS_LAB_REGION_FILE`; when absent it explicitly skips, even if other read-only
tests are enabled. Its JSON contains tenant_id, identity_id (selected user GUID),
cloud (commercial/gcc), entity_id, entity_type (User/Group), and tokens. The token
inventory must contain exactly the key returned by
`auth.TokenRequest{Audience: audience, Scope: audience+"/Region.Read.All"}.TokenKey()`
for audience https://migrationsvc.teams.microsoft.com, with access_token and
RFC3339 expires_at. Keep this private data outside the repository. Opaque token
claims/consent are not inferred; the caller explicitly provisions the delegated
identity and server remains authoritative. The app SPN cannot satisfy this case.
A matching delegated PowerShell session/oracle remains further qualification.

With explicitly selected private inputs, run only the intended case:

```sh
go test -count=1 -run '^TestAccBatchDeploymentStatusReadOnly$' ./tests/integration/lab
go test -count=1 -run '^TestAccPrivateChannelMigrationReadOnly$' ./tests/integration/lab
go test -count=1 -run '^TestAccMultiGeoRegionReadOnly$' ./tests/integration/lab
```

Never automate deployment submission merely to obtain a status fixture. Before
consequential acceptance, settle disposable teams/users, licensed owners,
notification recipients, a pre-write journal, accepted/lost-response recovery and
evidence-backed cleanup. No orchestration cancellation/deletion or per-team status
grammar has been assumed. Native tokens, configured cloud routes, offline tests
and empty/failed reads do not establish permission, feature availability or success.
