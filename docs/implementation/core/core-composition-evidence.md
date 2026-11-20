# Team and membership composition fixtures

Reviewed 2026-11-20. The [versioned scenarios](../../../contracts/endpoints/8.1.0/fixtures/core-team-compositions.json)
bind nine team/member commands and their eleven exact runtime parameter sets to
source-derived workflow examples. These are authored contract fixtures, not
recordings or executions of PowerShell, Go services or a tenant. A scenario can
exercise one segment under its stated precondition; it is not a complete command
emulator. The fixture validator checks provenance, references, parameter-set
bindings and request structure. It does not establish runtime behavior.

Pinned source identity and named methods are in the artifact and
[team wire evidence](team-wire-contract-evidence.md). Existing serializer captures
remain the authority for the shipped model output. All IDs/content are synthetic;
the picture bytes deliberately are not a valid image. Complete field/branch
fixtures for these commands and the other CORE-01 families are still required.

## Explicit native choices

The user-requested CRUD surface and [backend decision](../../decisions/0003-backend-contracts-and-compatibility.md)
allow native request models alongside PowerShell compositions. The scenario's
`variant` distinguishes `pinned-characterization` from `native-composition`.

- **Combined tenant search:** retain the pinned repeated `$search` keys as a
  characterization. The native enriched-list composition runs separate displayName
  and mailNickname queries, follows each collection completely, intersects group
  IDs, then hydrates and applies visibility/archive filters. Each request has one
  `$search`, `ConsistencyLevel: eventual` and `$count=true`. Search semantics remain
  those of each documented query; the intersection is an SDK composition choice.
  Partial pages or hydration failures cannot produce a complete successful list.
  [Current beta group query contract](https://learn.microsoft.com/en-us/graph/api/group-list?view=graph-rest-beta).
- **Team creation:** keep group creation, propagation probes, teamification and
  enrichment separate. The current group-to-team API documents `201` plus a team;
  this differs from native `POST /teams`' asynchronous contract. Preserve the
  pinned five-attempt/20-second propagation characterization separately from
  configurable, context-bounded native waiting. Never infer group ownership from
  an input ID. [Create from group](https://learn.microsoft.com/en-us/graph/api/team-put-teams?view=graph-rest-1.0).
- **Compensation and partial writes:** preserve created IDs and each completed
  stage. An ambiguous teamification result requires reconciliation before cleanup;
  a supplied group is never eligible for automatic deletion. Report failed
  cleanup alongside the primary error. Group/team updates and owner/member writes
  have no cross-request transaction. These are SDK lifecycle rules under the
  [accepted CRUD/composition amendment](../plan-amendments.md), not new server
  guarantees or an assertion that the pinned client verifies cleanup.
- **Archive/unarchive:** use documented bodies and return the accepted operation.
  Poll before a requested completion/readback. `202` is not completion. Unarchive
  sends no body; preserve the pinned JSON empty-string fixture separately.
  The archive SharePoint option retains explicit false versus omission and is
  documented unsupported in application context.
  [Archive](https://learn.microsoft.com/en-us/graph/api/team-archive?view=graph-rest-1.0),
  [unarchive](https://learn.microsoft.com/en-us/graph/api/team-unarchive?view=graph-rest-1.0).
- **Settings and membership:** use documented string Giphy values, preserve
  explicit false/empty fields, and retain the pinned numeric serializer case.
  Native `TeamMembers` CRUD uses opaque conversation membership IDs. Pinned
  Add/Get/Remove-TeamUser compositions use separate group relationships and
  directory IDs. A successful empty preflight permits native creation; a failed
  authorization response does not imply absence. Native preflight enumerates the
  selected IDs through every page before reporting absence, avoiding an
  unevidenced filtered existence query. Native owner-reference writes
  use the documented `/users/{id}` binding; the pinned `/directoryObjects/{id}`
  variant remains identified in the source note. See the
  [membership resource contract](membership-resource-contract.md) and
  [Giphy field contract](https://learn.microsoft.com/en-us/graph/api/resources/teamfunsettings?view=graph-rest-1.0).

## Remaining gates

The source-derived cases cover representative successes, validation, received
401/403 and partial failures. They deliberately do not claim all application
parameters, all fields, every branch, least privilege of whole compositions,
universal cloud/auth eligibility or live consistency. These nine command records
remain in progress. The [closure audit](closure-audit.md) continues to own the
finite CORE-01 acceptance gates; no coverage or delivery state changes here.
