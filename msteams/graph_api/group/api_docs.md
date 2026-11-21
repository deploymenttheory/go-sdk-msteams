# Backing groups

`sdk.GraphAPI.Groups` manages the groups used by team compositions. Group
creation and teamification are separate resource operations with separate
outcomes. A group create does not establish that a team or SharePoint site exists.

| Method | Route | Success |
| --- | --- | --- |
| `Create` | `POST /v1.0/groups` | 201 and a group snapshot |
| `GetByID` | `GET /v1.0/groups/{id}` | 200 and a group snapshot |
| `UpdateByID` | `PATCH /v1.0/groups/{id}` | 204 for the five supported properties |
| `DeleteByID` | `DELETE /v1.0/groups/{id}` | 204; deletes the group and associated resources |
| `List` | `GET /v1.0/groups` or explicit `/beta/groups` | 200 and one page |
| `NewPager`, `ListAll` | Collection plus opaque next links | Explicit bounded traversal |

Construction and pager construction are local. Methods use the root client's
Graph session and caller context, returning shared Resty metadata even on HTTP,
decode or unexpected-success-status failures. Mutations never replay. Neither
propagation waiting nor ownership inference nor compensation is hidden in these
primitives. A failed/ambiguous create may require reconciliation by its caller.

`NewMicrosoft365Group` creates local Unified-group inputs with mail enabled and
security disabled. It leaves visibility, owners, members and provisioning flags
omitted. The PowerShell composition chooses its own Private/education defaults.
The create service requires nonblank display name and mail nickname without
renaming either, maps user directory IDs to selected-cloud `/v1.0/users/{id}`
bindings and enforces the documented maximum of 20 initial relationships.
Nil reference slices omit the property; an explicit empty slice sends `[]`.
Required boolean values and groupTypes are serialized explicitly. Other type,
field and authorization restrictions remain Graph decisions.

Creation/resource-behavior options are sent only when explicitly supplied.
Directory extensions use explicitly supplied `extension_{appID}_{property}`
keys and valid raw JSON. They cannot overwrite ordinary fields or reference
properties. The pinned education extension and provisioning constants remain
captured evidence; the standard constructor does not inject them. Read snapshots
and unknown fields are never automatically resent in writes.

Create supports user owner/member references in this SDK input. In application
context, ownerless groups can have manageability/provisioning limitations. The
documented Group.Create scenario also requires permission to read referenced
object types (for user references, User.Read.All). No owner is resolved or
silently added by this service. The API enforces the configured grants and
principal restrictions; the SDK provisions no access.

`UpdateRequest` covers displayName, description, classification, visibility and
mailNickname, independently of team settings. Nil values are omitted; explicit
empty strings survive. This matches the actual pinned serializer captures and
does not establish a JSON-null clearing contract. Other Graph group property
families and their different response statuses are not exposed by this patch.
Reads preserve nullable fields and unknown extension data in `RawFields`.

List defaults to v1.0. `ListOptions.Version = group.Beta` explicitly selects the
beta collection used by the pinned tenant-search composition. Filter, search,
select, expand, orderby, count and top are mapped; skip is not supported. Zero
Top accepts the server default of 100; the maximum is 999. Advanced-query
consistency/count options are explicit because their requirements depend on the
query expression. An explicit false count is preserved. One search value is
encoded per request; combined tenant search composition remains separate work.

`List` returns one page and an opaque next link. `ListAll` follows all pages
within shared limits and returns partial items plus an error when traversal
fails. The pager freezes the original query/headers and requires HTTP 200 on
each page before decoding. Foreign continuations are rejected before credentials.
MaxPages/MaxItems default to 1,000/100,000; the item bound applies to aggregation.
Eventual consistency means a read/list outcome alone does not prove a recent
mutation was absent or that cleanup is safe.

Primary sources reviewed on 2026-11-20:
[create](https://learn.microsoft.com/en-us/graph/api/group-post-groups?view=graph-rest-1.0),
[get](https://learn.microsoft.com/en-us/graph/api/group-get?view=graph-rest-1.0),
[update](https://learn.microsoft.com/en-us/graph/api/group-update?view=graph-rest-1.0),
[delete](https://learn.microsoft.com/en-us/graph/api/group-delete?view=graph-rest-1.0),
[v1 list](https://learn.microsoft.com/en-us/graph/api/group-list?view=graph-rest-1.0),
[beta list](https://learn.microsoft.com/en-us/graph/api/group-list?view=graph-rest-beta).
These are documentation/fixture contracts; successful group lab CRUD remains
unqualified. Actual read-only first-page lab calls returned typed Graph HTTP 403
with both secret and certificate authentication. No group mutations or permission
changes were made.
