# CORE-23 team-operation SDK evidence

Status: implementation candidate with offline contracts. Issue #78 remains open
for authorization, tenant and PowerShell qualification. The dedicated identity
authenticates, but the exact role/Graph-grant requests remain pending. No tenant
operation was read or initiated by this implementation work.

`GraphAPI.TeamOperations` owns the pinned beta team-operation read. It exposes
GetByID, local ResumeByID and explicit WaitByID; operation snapshots retain all
pinned fields plus nullable/future values and typed error diagnostics. The
four recognized polling states have explicit meanings; unknown/missing states
remain readable metadata and fail polling interpretation. Targets stay opaque.

Route and projection provenance: MicrosoftTeams 8.1.0 TeamsCmdlets assembly SHA256
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`,
GetOperation.ProcessRecord and Model.TeamsAsyncOperation, as recorded in
[CORE-01 evidence](channel-contract-evidence.md). Parameter-set mapping remains
Get-Operation / __AllParameterSets. Synthetic request/reply fixtures exercise the
real root client without sockets or real credentials; they are source-informed,
not captured tenant traffic or exported-cmdlet execution.

The current [resource reference](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-beta)
supports the richer status/error fields and polling interval. The linked Get page
still describes chats. Its permissions do not settle the team endpoint, and the
native strict-200 Get expectation remains explicitly distinct from the pinned
helper's broad-success handling. Investigation of the archived beta Get reference
at docs commit `03c4645cc08d7623a901721b85a94db5c3b7fb77` also found a chat route;
this did not resolve team authorization. The generated
[Graph Get wrapper](https://learn.microsoft.com/en-us/powershell/module/microsoft.graph.beta.teams/get-mgbetateamoperation?view=graph-powershell-beta)
establishes navigation commands, including a collection form, without a matching
service permission table on the inspected page. Generated additional verbs
remain candidate contracts, not assumed operation-record lifecycle APIs.

Tests compare exact method, escaped beta path, Graph audience, selected headers,
body absence, response status and selected native values. They exercise all
recognized/future reads, failed-operation data, null/zero/future properties,
malformed/unauthenticated/forbidden responses, resumed success/failure,
identity changes and contradictory acceptance. Fake-clock cases verify the
31-second poll/retry floor, retained failure diagnostics and cancellation before
another read. Local tests verify credential-free construction/validation,
closed/wrong backend and configured cloud/mode routing. Examples compile without
HTTP. Neither a token nor synthetic cloud routing proves service availability.

CI retains only structural `team-operation-service.json` reports with revision,
case results, `coverage_claim:false` and `live:not-run`. The ledger remains
in-progress/metadata-only/not-run. The prepared read-only acceptance case is
[documented separately](team-operation-lab-acceptance.md). It grants no permission
and initiates no operation. The optional PowerShell conversion was extracted
from its AST and exercised against the actual pinned CLR model with synthetic
values, without importing the module or executing management cmdlets; parsing
and the disabled guard also pass offline. This is projection verification,
rather than a successful tenant or PowerShell/Go comparison.
