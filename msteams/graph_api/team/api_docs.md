# Native teams

Use `sdk.GraphAPI.Teams` with an explicitly configured client. Construction is
local and lazy. Methods propagate the caller context and share the client's
tenant, cloud, identity, request budget and response/error handling.

| Method | Graph v1.0 route | Success | Behavior |
| --- | --- | --- | --- |
| `Create` | `POST /teams` | 202 | Returns an accepted operation; no implicit wait |
| `GetByID` | `GET /teams/{id}` | 200 | Native snapshot, without group hydration |
| `UpdateByID` | `PATCH /teams/{id}` | 204 | Sparse settings patch |
| `DeleteByID` | `DELETE /groups/{id}` | 204 | Deletes the backing group and associated team resources |
| `List`, `NewPager` | `GET /teams` | 200 | Basic projections, bounded opaque-link paging |
| `CreateFromGroup` | `PUT /groups/{id}/team` | 201 | Teamifies a caller-owned group; returns the team |
| `Archive`, `Unarchive` | `POST /teams/{id}/{action}` | 202 | Explicit asynchronous actions |
| `UploadPicture` | `PUT /groups/{id}/photo/$value` | 200/204 | Streams original JPEG bytes, at most four MB |
| `ResumeOperation` | `GET /teams/{id}/operations/{operationId}` | Status reads | Reconstructs a polling handle; never resubmits a mutation |

Team/group/user directory identifiers are UUIDs. Operation and template
identifiers accept letters, digits, hyphens and underscores. The SDK builds
template/user bindings from the configured Graph origin; the request cannot
supply a separate credential destination.

`CreateRequest.TemplateID` defaults to `standard`. An application request must
specify an initial user. `InitialMember.Owner` emits the owner role, while false
emits an empty roles array. Delegated requests can omit members. There is no
implicit `/me` lookup or directory search. Group-first creation, owner
resolution and education-group composition remain separate CORE-28 work.

Settings pointers preserve omission, explicit false and explicitly supplied
empty containers. `FunSettings.GiphyContentRating` uses Graph's string values
`moderate`/`strict`; future strings remain readable. The pinned module's numeric
serializer artifact is retained in its original fixture and tested as an
explicit native-API difference. A read `Team.RawFields` retains the complete
object, including future properties. Mutation models never resend read-only or
unknown fields automatically.

Only `id`, `displayName`, `description` and `visibility` are populated by native
list. Nil settings or archive state do not imply defaults. List supports the
documented filter, select, top, skiptoken and count parameters. A negative bound
or page size is rejected locally. Zero traversal bounds use the shared defaults
of 1,000 pages and 100,000 items. A failed or bounded traversal returns its
partial items and last response with an error. Treat that result as incomplete.
The next-link query remains opaque; the shared scope rejects foreign origins.

All mutations use `ReplayNever`, including create-from-group propagation
failures and failed uploads. Native methods do not perform compensation. In
particular, a group supplied to `CreateFromGroup` is never deleted on error.
HTTP 201 does not establish that background SharePoint provisioning completed.
Create-from-group can return 404 while a newly created group replicates; an
explicit composition must reconcile that state under its own bound and ownership
rules. An ambiguous create/delete must not be blindly retried.

`AcceptedOperation.Response` is retained even when a 202 response lacks a safe
operation link. A valid handle starts `accepted`. Only the documented teams
operation route is permitted, including Graph's root-relative OData form.
Unrelated same-origin routes, foreign origins, beta routes and query/fragment
locations are rejected. Archive/unarchive also require the location's team ID
to match the requested team.

Polling uses the shared serial operation handle. It waits at least 31 seconds
between checks and honors a longer `Retry-After`. The first read after
`ResumeOperation` is immediate. Callers must preserve the prior cadence when
resuming a handle immediately after an earlier poll. `notStarted`, `inProgress`,
`succeeded` and `failed` are mapped explicitly. Unknown states fail, terminal
failures retain typed diagnostics, and target-resource locations remain opaque
metadata. A local cancellation or timeout does not cancel the remote operation;
retain its handle and last response for recovery. Do not overlap `Poll`/`Wait` on
one handle.

Archive omits the request body when its SharePoint option is absent; unarchive
always omits it. These native requests deliberately differ from the pinned
PowerShell JSON-null/empty-string artifacts. Supplying the SharePoint option
with application authentication is rejected before acquiring a token, including
an explicit false. That option requires delegated authentication.

`UploadPicture` takes ownership of its source immediately, including validation
failure, and closes it exactly once. `Close` must unblock `Read`. The stream is
not buffered, converted, base64 encoded or replayed. A size/read/transport
failure may leave partial remote state. The chosen group-photo route follows
`Set-TeamPicture`; its permission requirements differ from `/teams/{id}/photo`.

The service does not grant permissions, infer consent from JWT claims, or switch
identities when access is denied. Native Graph application create is supported
by the documented API; this is distinct from the pinned PowerShell command's
app-auth restriction. Permission references include `Team.Create` for create,
`Team.ReadBasic.All` for list, `TeamSettings.ReadWrite.All` (or documented
resource-specific consent) for settings/archive, `Group.ReadWrite.All` for
teamifying/deleting a backing group, and `ProfilePhoto.ReadWrite.All` for group
photos, with alternatives documented per operation. Existing group grants have
broader effects and must be evaluated as such. These references are not lab
qualification or a request to provision those grants automatically.

Primary sources verified on 2026-11-20:
[create](https://learn.microsoft.com/en-us/graph/api/team-post?view=graph-rest-1.0),
[get](https://learn.microsoft.com/en-us/graph/api/team-get?view=graph-rest-1.0),
[update](https://learn.microsoft.com/en-us/graph/api/team-update?view=graph-rest-1.0),
[delete group](https://learn.microsoft.com/en-us/graph/api/group-delete?view=graph-rest-1.0),
[list](https://learn.microsoft.com/en-us/graph/api/teams-list?view=graph-rest-1.0),
[create from group](https://learn.microsoft.com/en-us/graph/api/team-put-teams?view=graph-rest-1.0),
[archive](https://learn.microsoft.com/en-us/graph/api/team-archive?view=graph-rest-1.0),
[unarchive](https://learn.microsoft.com/en-us/graph/api/team-unarchive?view=graph-rest-1.0),
[photo](https://learn.microsoft.com/en-us/graph/api/profilephoto-update?view=graph-rest-1.0),
[operation resource](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-1.0),
[operation states](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperationstatus?view=graph-rest-1.0).
