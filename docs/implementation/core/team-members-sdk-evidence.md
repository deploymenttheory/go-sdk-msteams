# CORE-26 SDK progress

Status: in progress. Native conversation-membership CRUD/list and owner
promotion/demotion are implemented through `sdk.GraphAPI.TeamMembers`, with
separate read/write models, one-page list and explicit bounded traversal.
Backing-group member/owner reference primitives are also implemented through
separate services. PowerShell compositions and journalled lab acceptance remain
required. This does not close
[CORE-26 / #182](https://github.com/deploymenttheory/go-sdk-msteams/issues/182)
or promote any command coverage.

The models implement the accepted
[membership resource boundary](membership-resource-contract.md): opaque native
membership IDs and directory user IDs are distinct Go types. Read models retain
unknown subtype/role/property information, nullable AAD fields and another
tenant's identity without decoding or substituting the membership ID. Chat-only
history is a read field and does not enter a team role patch.

Role-patch serialization distinguishes omission, explicit null, explicit empty
array and owner promotion. Only the empty array represents demotion. Resource
methods reject null/unsupported role writes, preserve native membership IDs in
escaped routes and return the updated member on HTTP 200. Server-assigned guest
roles are not writable; Graph enforces guest-promotion restrictions. A create
input supplies a directory user ID, and the service constructs its binding from
the selected Graph cloud.

Tests use the real root client through the shared CORE-15 offline runner. They
check exact native CRUD/action routes and statuses, escaped opaque IDs, cloud
bindings, omitted/empty role writes, malformed responses, typed authorization and
server failures, mutation non-replay, native single-page results, continuation
queries, traversal bounds/partial results and local cancellation/validation.
Configured application/delegated routing is checked across all seven built-in
profiles; no cloud, mode or permission eligibility is inferred.

The [native operation contract](../../../contracts/endpoints/8.1.0/fixtures/core-native-team-members.json)
is authored from current primary documentation. It is not a PowerShell capture
or a tenant recording. Nine CRUD/action fixture cases emit the sanitized
`team-member-service.json` CI report using the tested revision. They have no
PowerShell command mappings, explicitly set `coverage_claim: false`, and leave
live qualification `not-run`. Additional error/paging/model cases run as normal
tests. The CI uploader includes this exact report name alongside existing
primitive evidence, without widening the artifact allowlist to arbitrary files.

Native page traversal adopts the shared expected-status option and requires
HTTP 200 before decoding, so a collection-shaped 202/206 cannot masquerade as a
complete list. Source/read models and compile-only examples perform no tenant
actions in ordinary tests. No native membership lab calls or mutations were
made; the existing dedicated SPN's prior teams-discovery read was forbidden.
Successful lab qualification remains outstanding. See the
[API notes](../../../msteams/graph_api/team_member/api_docs.md).

Primary sources verified on 2026-11-20:
[AAD member](https://learn.microsoft.com/en-us/graph/api/resources/aaduserconversationmember?view=graph-rest-1.0),
[create](https://learn.microsoft.com/en-us/graph/api/team-post-members?view=graph-rest-1.0),
[update team member](https://learn.microsoft.com/en-us/graph/api/team-update-members?view=graph-rest-1.0),
[list and opaque identities](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0).

The added `GroupMembers` and `GroupOwners` services expose AddReference, List,
RemoveReference and explicit bounded traversal. They share internal request
mechanics while retaining separate public types. Tests use the real root client
and shared offline runner to verify default selected-cloud bindings, explicit
directory-object/user/service-principal references, exact 204 responses and the
fixed `/$ref` delete suffix. They also cover null/limited directory data, opaque
paging/header snapshots, partial results, status guards, local validation and
received 400/403/404/412/429/503 outcomes without mutation replay or false duplicate
success. Configured mode/cloud checks remain routing evidence only.

These primitives do not preflight, resolve names, add a second relationship,
enforce a client-side last-owner rule, wait for synchronization or compose the
pinned Add/Get/Remove-TeamUser commands. Actual server errors remain distinct
from the pinned local synthetic 412 branch. Explicit owner directoryObjects
binding is retained separately from the native default users binding. Native
endpoint service-principal omissions and limited data remain documented limits;
they are not converted into absence or a fabricated user role. No reference
tenant calls or mutations were made. See the
[reference API notes](../../../msteams/graph_api/group_relationship/api_docs.md).
