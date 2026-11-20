# CORE-28 SDK progress

Status: in progress. Native team create/read/settings-update/delete/list,
create-from-group, archive/unarchive, picture upload and resumable operation
handling are implemented under `sdk.GraphAPI.Teams`. Supporting group CRUD,
PowerShell compositions and successful tenant acceptance remain required.
This does not close [CORE-28 / #81](https://github.com/deploymenttheory/go-sdk-msteams/issues/81)
or promote its command coverage.

Supporting group models are now started under `graph_api/group`: separate native
read/create/update inputs, a local standard Unified-group constructor, nullable
read fields and retained future extension data. Group update serialization is
checked against the actual pinned omitted/explicit-empty captures. The native
constructor does not inject the PowerShell Private default, user references or
undocumented provisioning flags. No Groups root service or management methods
are exposed yet; group CRUD and the source-derived compositions remain required.
Primary [create](https://learn.microsoft.com/en-us/graph/api/group-post-groups?view=graph-rest-1.0),
[get](https://learn.microsoft.com/en-us/graph/api/group-get?view=graph-rest-1.0),
[update](https://learn.microsoft.com/en-us/graph/api/group-update?view=graph-rest-1.0)
and [list](https://learn.microsoft.com/en-us/graph/api/group-list?view=graph-rest-1.0)
documentation was reviewed on 2026-11-20. No group tenant operations were made.

Native settings preserve omitted versus non-nil empty containers and explicit
false values. A caller can construct the five empty containers observed in the
module without forcing that behavior into every native update. Tests compare
those containers and explicit flags against the actual pinned serializer cases.

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

Before CORE-28 can close, implement the group-first and enriched PowerShell
compositions from the [existing decision record](core-composition-evidence.md),
including directory owner resolution, education-group constants, bounded group
propagation/reconciliation, group-plus-settings update outcomes, archive-state
checks, owned-resource recovery and partial hydration/list behavior. Add the
service contract reports and journalled tenant acceptance for those paths.
Successful lab validation and least-privilege access setup remain outstanding.
