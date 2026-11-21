# CORE-28 SDK progress

Status: in progress. Native team create/read/settings-update/delete/list,
create-from-group, archive/unarchive, picture upload and resumable operation
handling are implemented under `sdk.GraphAPI.Teams`. Supporting group CRUD is
implemented under `sdk.GraphAPI.Groups`. New-Team creation/migration, Get-Team
identity/enriched listing, Set-Team and Set-TeamArchivedState compositions are implemented
under `sdk.GraphAPI.TeamManagement`. Command mappings and contract review are in
progress; successful tenant acceptance remains outstanding.
This does not close [CORE-28 / #81](https://github.com/deploymenttheory/go-sdk-msteams/issues/81)
or promote its command coverage.

Supporting `UserResolver` reads are implemented for explicit directory IDs,
UPNs, optional delegated current user and direct-404-only canonical mail fallback.
All-page uniqueness, OData escaping, pending reads and authorization failures are
distinct. Tests exercise the actual root client through the shared offline
runner, including later-page ambiguity, bounded/foreign continuations and zero-I/O
UUID selection. This corrects the pinned alias/first-result behavior explicitly;
it is not a claim of identical wire traffic or live user-resolution qualification.
See the [resolver notes](../../../msteams/graph_api/directory_user/api_docs.md).

Supporting groups now expose native create/get/update/delete, one-page list and
explicit bounded traversal, including an explicitly selected beta collection for
the pinned tenant-search path. Separate read/create/update inputs retain nullable
read fields and future extension data. Create and update tests compare the actual
pinned ordinary-create and omitted/explicit-empty captures. The local standard
Unified-group constructor does not inject the PowerShell Private default, user
references or undocumented provisioning flags. Supplied directory extensions
cannot overwrite ordinary fields. The [group API notes](../../../msteams/graph_api/group/api_docs.md)
describe lifecycle, query and access boundaries. Source-derived compositions,
their service evidence reports and successful journalled tenant acceptance remain
required; no group command qualification is inferred from these primitives.
Primary [create](https://learn.microsoft.com/en-us/graph/api/group-post-groups?view=graph-rest-1.0),
[get](https://learn.microsoft.com/en-us/graph/api/group-get?view=graph-rest-1.0),
[update](https://learn.microsoft.com/en-us/graph/api/group-update?view=graph-rest-1.0)
and [list](https://learn.microsoft.com/en-us/graph/api/group-list?view=graph-rest-1.0)
documentation was reviewed on 2026-11-20. Read-only first-page group-list probes
using the dedicated lab SPN returned typed Graph HTTP 403 with both secret and
certificate authentication. No group mutations or permission changes were made;
this establishes denial handling, not successful group CRUD or least privilege.

Native settings preserve omitted versus non-nil empty containers and explicit
false values. A caller can construct the five empty containers observed in the
module without forcing that behavior into every native update. Tests compare
those containers and explicit flags against the actual pinned serializer cases.

The additional [field capture](../../../contracts/endpoints/8.1.0/fixtures/core-team-settings-fields.json)
contains 36 actual isolated model observations: false/true for all 17 boolean
settings and both Giphy ratings. Its guarded [offline generator](../../../tools/contracts/capture-team-settings-fields.ps1)
checks the pinned assembly hash before import and invokes no management command.
Seventy-two real-root create/update executions compare every field with those
observations, including five-container placement and typed non-default readback.
Two reports, `team-create-settings-service.json` and
`team-update-settings-service.json`, retain the tested revision and exact set
mappings. The parameter audit checks both settings-bearing runtime sets so no
application settings parameter can disappear unnoticed. Numeric Giphy captures
remain immutable; only the test's native comparison uses the reviewed strings.
These executions qualify the offline field mapping, not tenant acceptance.

The paired management service verifies both returned IDs, projects group-owned
properties and retains both native snapshots. Update patches group properties,
then all five settings containers, then reads group/team state. Later failures
retain completed stages, the known ID and latest response without rollback or
replay. Three source-derived workflow cases emit `team-get-service.json` and
`team-update-service.json` through the actual root SDK and shared runner, mapped
only to Get-Team/Identity and Set-Team/__AllParameterSets. Additional tests check
the actual empty-container capture, denied stages, mismatched IDs, cancellation,
closed clients and local validation. These are synthetic SDK executions, not
runtime cmdlet captures. Reports retain `coverage_claim: false` and live
qualification `not-run`. See the [paired management notes](../../../msteams/graph_api/team_management/api_docs.md).

The archive-state composition now performs paired pre-read, unknown/unchanged
state rejection, one archive/unarchive submission, native operation waiting and
paired confirmation. It distinguishes acceptance, remote completion and observed
archive state; stale/missing readback is not synthesized into success. A failed
wait retains the accepted handle, latest snapshot, pre-read and stages without
replaying the action. Three positive cases emit `team-archive-service.json`,
mapped to Set-TeamArchivedState/__AllParameterSets. Virtual time checks the
31-second first-poll cadence and recovery after a shorter timeout, including
preserving cadence across a second wait. Additional cases exercise cancellation,
client close, remote/unknown/failed status, poll limits, denied reads, unsafe
handles, incomplete readback, unchanged/unknown precheck states and local option
validation. No propagation exception is inferred and no tenant call was made.
The actual pinned composition submissions and serializer differences remain
distinct from these synthetic SDK executions; reports retain no coverage/live
promotion. Primary archive/unarchive/operation documentation was reviewed on
2026-11-21 and is linked in the API notes.

Group-first Create resolves an owner, creates the group with pinned defaults and
provisioning presets, probes availability, teamifies and returns paired state.
MigrateGroup has a separate input surface and never owns/deletes the supplied
group. Actual EDU class/PLC model captures supply the expected request constants.
Explicit aliases remain unchanged; generated aliases require complete successful
availability reads, with bounded candidates and paging. Current API guidance
informs configurable received-404 retries; the pinned five/20-second sequence
remains separately characterized. Three representative real-root cases emit
`team-create-service.json`, mapped to both New-Team parameter sets. Mode/cloud
checks are routing evidence only, not PowerShell app-auth eligibility.

Compensation is limited to this call's confirmed created group. Positive team
reconciliation retains it; a received rejection plus observed 404 can permit
cleanup. Response loss, 5xx or decode failures cannot become absence after a
single 404. Primary and cleanup failures retain separate diagnostics and are
joined. Foreground timeout preserves identity and uses a separate bounded cleanup
context when eligible. A successful team is retained after failed enrichment.
Custom in-memory tests verify dynamic aliases, later-page collisions and lost
group/team responses without sockets, replay, fabricated IDs or blind deletes.
These source-derived/synthetic executions do not establish successful tenant
creation, education provisioning, SharePoint completion or command qualification.
See the [creation notes](../../../msteams/graph_api/team_management/creation_api_docs.md).

The enriched listing service now discovers beta team groups or v1 direct joined
teams, traverses every required source before selection, and retains incomplete
source prefixes. Combined tenant search uses separate queries and complete-ID
intersection; user discovery uses directory resolution and paired reads. Tenant
hydration uses the independently extracted batch engine from PR #238, preserves
response correlation/input order and retries only throttled items within explicit
bounds. Successful peers remain retained on later failure/cancellation. Typed
subrequest errors keep codes/request IDs without printing payloads; a batch's
outer HTTP 200 cannot hide inner denial. Missing selected filter properties remain
unknown, and partial discovery never establishes complete absence/intersection.
Three source-derived cases emit `team-list-service.json` mapped to Get-Team/Filters.
Additional offline tests cover batch width, mixed/malformed outcomes, statuses,
normalized diagnostics, exact filtering, throttles, deadlines and client close.
No live calls were made. See the [listing notes](../../../msteams/graph_api/team_management/list_api_docs.md).

Remove-Team maps to `Teams.DeleteByID`, which deletes only through the backing
group route. Set-TeamPicture maps to `Teams.UploadPicture`. Both now exercise the
immutable source-derived fixture through the real root client and shared runner,
with additional denied-response cases. Reports `team-delete-service.json` and
`team-picture-service.json` retain only structural metadata. The picture fixture's
original synthetic bytes are deliberately not a valid image: the test proves
exact transfer, JPEG content type and single source close, without claiming that
Graph accepts those bytes as an image. The shared binary assertion prerequisite
is separate from production transfer behavior. Existing native tests additionally
cover the size bound, local validation, cancellation and mutation non-replay.

All six assigned commands and eight parameter sets now link their actual Go
symbols and tests in the coverage ledger. Implementation remains `in-progress`,
contract status remains `metadata-only`, and live status remains `not-run` while
the candidate evidence and remaining acceptance work are reviewed. These links
do not infer a contract pass or tenant success from green CI alone.

Opt-in `TestLabTeamLifecycle` and `TestLabRecoverTeamLifecycle` are now prepared
for disposable group-first and migration fixtures. They skip before credential
loading without all three lab flags, require a private same-tenant owner fixture,
and use a bound recovery journal with positive paired ownership checks before
deletion. Offline tests constrain denied/incomplete/ambiguous discovery and
cleanup. These Go-only scenarios cover lifecycle readbacks, archive transitions
and valid photo transfer/readback when explicitly run; they are not PowerShell
comparisons and have not run against the tenant. See the
[acceptance procedure](team-lifecycle-lab-acceptance.md) for exact scope and limits.

The primary native Giphy payload uses the documented string enum. A test first
requires the pinned artifact to retain its observed numeric 1, then records the
explicit native string difference without rewriting the capture. Unknown future
string values remain representable. This follows the already accepted
[composition decision](core-composition-evidence.md), not a claim that the pinned
serializer emitted a string.

Verified primary sources on 2026-11-20:
[Graph settings](https://learn.microsoft.com/en-us/graph/api/resources/teamfunsettings?view=graph-rest-1.0),
[update](https://learn.microsoft.com/en-us/graph/api/team-update?view=graph-rest-1.0),
[create](https://learn.microsoft.com/en-us/graph/api/team-post?view=graph-rest-1.0),
[list](https://learn.microsoft.com/en-us/graph/api/teams-list?view=graph-rest-1.0).
Create is asynchronous; a 202 is not completion. Native list's basic projections
remain distinct from the enriched Get-Team composition. The resource methods
and remaining scope are described in the [API notes](../../../msteams/graph_api/team/api_docs.md).

Offline tests exercise the actual root client through an injected HTTP transport
that cannot open sockets or fall back to any other endpoint. They check CRUD
routes, sparse settings, unknown read fields, native list projections,
continuation queries and partial results, mutation non-replay, typed failures,
malformed envelopes, input validation and cancellation. Native template/user
bindings are checked for application and delegated configuration across all
seven built-in routing profiles. These are routing checks, not cloud/mode
eligibility or live support claims.

Native list traversal selects the shared HTTP 200 expectation. Tests reject
collection-shaped 202/206 responses before accepting items or completion, while
preserving their response metadata. The optional shared guard was independently
verified and merged through [PR #236](https://github.com/deploymenttheory/go-sdk-msteams/pull/236).

Operation tests retain accepted response metadata on invalid handles, reject
foreign/unrelated/wrong-team status routes, and check known/unknown/failed states
and opaque target metadata. Virtual time verifies the 31-second cadence,
Retry-After, terminal stickiness and timeout recovery without wall-clock sleeps.
Picture tests check original bytes, the group-photo route, JPEG content type,
four-MB bound, non-replay and immediate stream ownership on local failures.
The existing public Graph operation decoder delegates to this resource decoder;
its behavior is unchanged.

Actual read-only lab probes on 2026-11-20 used the existing dedicated lab SPN
with both client-secret and certificate authentication. A native first-page
`GET /v1.0/teams?$top=1` returned HTTP 403 in both cases, classified by the
shared typed Graph API error. No resource, role or permission changes were made.
No response bodies, tenant identifiers, tokens or credentials were recorded in
the repository. This verifies denial handling; it does not validate successful
team CRUD, operation completion, provisioning, archive transitions or photos.

Before CORE-28 can close, finish reviewing command evidence and journalled tenant
acceptance across the [existing decision record](core-composition-evidence.md),
including successful resource/composition scenarios.
Successful lab validation and least-privilege access setup remain outstanding.
