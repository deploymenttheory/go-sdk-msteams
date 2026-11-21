# CORE-25 targeting implementation progress

Candidate branch work, not full ticket parity or qualification. Issue #79 remains
open. No lab authentication or tenant mutation was performed for this work.

The native `TeamsAdminAPI.TargetingPolicies` resource now has List,
GetByIdentity, ListFiltered, UpdateByIdentity, explicit UpdateGlobal and
DeleteByIdentity. The pinned modern ConfigAPI routes and generated request
fixtures are implemented using the existing per-session discovery adapter.
Identity is one escaped segment. PUT includes that identity and only bound
properties; nil omits, explicit empty strings survive. No guessed setting
defaults, strict enum closure, new generic PATCH, Create or assignment is exposed.
Received null/future fields remain in native snapshots. Global-first stable
ordering and local Identity/Tag:Identity wildcard matching compose List without
a server Filter query. The established UTF-16 matcher now lives in a shared
internal policy helper; existing application-access-policy cases exercise it.

This is the **modern native REST** resource. The pinned targeting cmdlets may
select TRPS when a tenant is unflighted; the Go SDK does not start PowerShell or
claim universal exported-command parity. Preserve the conditional disposition
from [routing evidence](policy-routing-evidence.md) and
[policy evidence](template-policy-deployment-evidence.md) when recording later
validation. Neither a missing flight nor an available modern method proves
service acceptance/rejection. Creating a new policy is still an unsettled
contract despite the user's true-CRUD goal; do not invent POST or call PUT a
proven upsert. Global deletion/reset/default assignment effects require lab
observation before destructive acceptance.

Remaining implementation: v1/v2 hierarchy native operations and typed status
projections, source-informed multipart upload
composition and partial outcomes, examples/API documentation, complete command
ledger mapping, retained structural fixtures and gated lab acceptance. This
progress checkpoint does not update ledger rows as complete.

Hierarchy belongs under TeamsAdminAPI but uses distinct service profiles:
commercial targeting v1 origin/audience `https://retailservices.teams.microsoft.com`;
GCC changes only the request origin to
`https://retailservices.gcc.teams.microsoft.com`. Hierarchy v2 uses request origin
`https://hierarchy.teams.microsoft.com` and audience
`https://chatsvcagg.teams.microsoft.com`. Other clouds lack settled v2 mappings.
The candidate now extends the closed native profile catalog with these exact
pairs, and native providers accept only their evidenced resource audiences in
the selected cloud. Graph/ConfigAPI endpoint overrides retain their existing
meaning; the new special-service profiles have no custom override or fallback.
Targeting scope is restricted to `/api/rest/v1/nodes`; hierarchy scope to
`/api/v1/hierarchies`. Unsupported maps and foreign paths fail before token
acquisition. ConfigAPI tokens are never substituted. The complete source contract is
[hierarchy evidence](hierarchy-migration-contract-evidence.md) and
`contracts/endpoints/8.1.0/special-service-profiles.json`.

Offline checks exercise these pairs against that manifest, fixed namespace and
cloud rejection, typed/redacted backend errors, secret/PFX/PEM MSAL token/cache
isolation for all four current audiences, and delegated acquisition with one
selected account and separate resource tokens. This extends the existing
`audience + /.default` contract only. Other services with explicit delegated
scopes such as Region.Read.All still require a scope-aware token/cache extension;
they are not accepted by this change. None of these synthetic checks establishes
tenant grants, application eligibility, regional service deployment or live
hierarchy behavior.

Checkpoint verification after integrating main: full build, vet, shuffled race
suite and repository-wide lint passed with all tenant lab flags unset. These are
local offline results; no CORE-25 PR or cross-platform qualification is claimed.
