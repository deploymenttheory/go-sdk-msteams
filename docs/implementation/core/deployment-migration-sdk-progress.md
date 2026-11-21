# CORE-22 deployment/migration implementation progress

Issue #77 remains in progress. No tenant request, notification email, directory
grant or migration/deployment mutation has been performed for this work.

The first candidate component is a pure bounded deployment archive builder in
`teams_admin_api/batch_deployment`. It writes exactly Teams.csv/Users.csv using
fast Deflate, preserves original bytes, returns immutable ZIP/Base64 access and
keeps readers caller-owned. It neither parses CSV nor uses shared temporary
files/stale archive entries. Combined input and complete ZIP-output budgets are
explicit local limits, not claimed service limits. Tests cover framing/limits,
invalid/failed sources, cancellation, byte isolation and concurrent construction.

Remaining implementation: typed ConfigAPI submission and opaque status reads;
region/private-channel migration resources; exact additional profiles and scoped
token/cache support for delegated Region.Read.All; command/parameter-set ledger,
source-informed service tests, examples/API docs and gated lab acceptance.

Production routing must use the selected ConfigAPI origin and existing discovery
header adapter. The generated batch request fixture's
skypeanalyticsapi-dev.trafficmanager.net authority is pre-callback data; the pinned
session callback replaces it before sending. Do not publish or use that development
origin as a credential destination. Source/immutable hashes are recorded in
[batch evidence](template-policy-deployment-evidence.md#batch-deployment) and
[routing evidence](policy-routing-evidence.md).

Submission is one POST with DeploymentCsvs (Base64 ZIP) and UsersToNotify. It can
send notification email and must not be replayed. Preserve instanceId and
executionId on 202; a status lookup uses instanceId. Status is raw text plus HTTP
metadata, including named 400/404/500 failures in Go. No structured per-team model,
waiter terminal grammar or inference that Completed means success is established.

The other two resources have separate service contracts in
[hierarchy/migration evidence](hierarchy-migration-contract-evidence.md) and the
versioned special-service profile manifest. Region.Read.All is a delegated scope,
not /.default or an application permission; Get-MultiGeoRegion is documented
unsupported in application mode. Private-channel migration has exact cloud base
prefixes, and its Details field is JSON text. Their profiles are not ConfigAPI
or Graph, and neither native availability nor authorization is inferred from a
token. Reuse the CORE-25 closed catalog after its green PR merge, then extend scope
isolation explicitly; never widen audience acceptance or use a commercial fallback.
