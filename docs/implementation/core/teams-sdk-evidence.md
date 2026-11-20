# CORE-28 SDK progress

Status: in progress. Only the native settings models are implemented in this
first slice. Team/group CRUD, paging, asynchronous creation, archive/picture
operations, PowerShell compositions, root wiring and tenant acceptance remain
required. This does not close [CORE-28 / #81](https://github.com/deploymenttheory/go-sdk-msteams/issues/81).

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
must remain distinct from the enriched Get-Team composition. These contracts
will guide the remaining service implementation; they are not delivered methods
or live qualification in this first slice.
