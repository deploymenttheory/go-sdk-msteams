# Policy assignment tenant tests

The SDK implements group assignment CRUD, effective-user reads and batch
submission/status, with separate conditional OCE operations. Native tests cover
requests, omitted/null/empty values, decoded outcomes, partial failures and HTTP
receipts. See the [SDK reference](../../msteams/teams_admin_api/policy_assignments/api_docs.md).

The dedicated application-access policy scenarios already exercise user grant,
read and clear/restoration, plus group create/read/update/delete and group grant.
Their tenant runs passed in certificate and client-secret modes. Actual response
JSON drives the service's unit regressions. Group tests preserve the prior empty
assignment and remove both owned policies; independent cleanup and six selected
PowerShell comparisons are required by the runner.

Generic batch and OCE read tests remain pending: a suitable tenant/job fixture and
OCE contributor eligibility are not configured. Earlier 403 responses do not prove
all authentication modes are unsupported. Batch submission needs controlled users
and original-state snapshots; OCE refresh/reprocessing requires an explicit target
tenant, environment/region, eligibility and consequential opt-in. No test activates
a privileged profile or grants roles automatically. Effective-user precedence,
member propagation and cross-group ranking remain separate from HTTP acceptance.
