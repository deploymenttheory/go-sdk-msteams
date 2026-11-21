# Enriched team listing

`sdk.GraphAPI.TeamManagement.List` implements Get-Team/Filters. Identity reads
remain a separate `GetByID` method, so list filters cannot accidentally change
identity lookup behavior. Native basic collections remain on `Teams.List` and
`Teams.ListJoined`.

Nil `ListOptions.User` selects tenant discovery. A non-nil selector uses the
shared directory resolver; a non-nil empty selector explicitly requests the
current delegated user. Application mode cannot infer a current user. Resolved
users are read through `/v1.0/users/{id}/joinedTeams` without initial OData query
parameters. Joined-team projections retain their nullable/raw data, and the
composition enriches each distinct ID through paired group/team reads.

Tenant discovery starts at `/beta/groups` with the pinned team-provisioning
filter, top 999, count true and eventual consistency. Each continuation retains
headers and uses the opaque next link. DisplayName search is tokenized by the
server; MailNickname search has the documented prefix behavior. Combined filters
run separate complete queries and intersect case-insensitive group IDs, retaining
the first query's group snapshots. The pinned repeated-$search request remains a
characterization; the SDK does not reproduce its ambiguous combination. Search
terms are quoted, fixed to known fields and URL encoded; they are not raw query
fragments. Beta and eventual consistency do not provide an atomic tenant snapshot.

Hydration uses v1.0 JSON batches containing only GET /teams/{id}. BatchSize
defaults to 20 and accepts 1..20; it is a request width rather than a goroutine
count. Response IDs are correlated independently of server order. Directory IDs
are validated before constructing request paths. HTTP 200 is required for both
the outer hydration response and individual team reads; malformed or mismatched
team identities cannot establish a successful item.

Only subrequest 429s are retried, preserving successful/permanently failed peers.
MaxThrottleAttempts defaults to four (initial plus three retries); one disables
retries. A positive Retry-After supplies the delay; otherwise ThrottleDelay
defaults to ten seconds. Multiple pending throttles wait for the longest delay.
The outer batch is never replayed, including 401, 429, transport failure or
malformed envelopes. The pinned one-auth-retry branch is not permission for
credential fallback or replaying arbitrary batch operations. Received failures
remain typed and explicit. The generic batch engine's existing policy is unchanged.

Timeout defaults to five minutes for the aggregate workflow. Traversal defaults
to sixteen pages per source, with shared aggregate item bounds; callers can
increase bounds explicitly. Caller cancellation and client close apply to the
whole operation, including waits. No worker survives the call. Malformed outer
responses stop later batches and retain their unattempted identities.

Tenant visibility/archive predicates are applied locally after hydration. User
queries additionally apply exact case-sensitive display-name comparison and
case-insensitive mail-nickname comparison. Visibility ignores case; Archived
uses a pointer so false remains a selected filter. Missing selected fields are
unknown and produce ErrFilterUnavailable unless another known nonmatch already
excludes the item. Basic joined projections do not override later full reads.

ListResult retains source discovery prefixes, candidate IDs, ordered outcomes,
successful Items, latest response and stages. Each outcome preserves its value,
known filtering result, attempts, batch subresponses and typed ItemFailure.
Normalized subrequest errors preserve service codes/request IDs without treating
the outer HTTP 200 as success. Incomplete discovery does not trigger hydration or
establish a complete intersection/absence. Missing batch outcomes keep known
successful observations while failing the overall list. Complete means all
selected work finished without a final failure; it is not a consistency guarantee.
Default formatting redacts identities and raw server data.

Access must cover directory resolution when selected, source collections and
every group/team hydration operation. No consent or role is granted. Direct
joined teams exclude shared-channel host teams; associated teams are a different
API. Primary documentation reviewed 2026-11-21: [beta groups](https://learn.microsoft.com/en-us/graph/api/group-list?view=graph-rest-beta),
[joined teams](https://learn.microsoft.com/en-us/graph/api/user-list-joinedteams?view=graph-rest-1.0),
[JSON batching](https://learn.microsoft.com/en-us/graph/json-batching).

Three source-derived real-root cases emit `team-list-service.json` mapped to
Get-Team/Filters. Additional tests exercise paging/status guards, partial batch
responses, normalized diagnostics, batch width/order, exact filters, throttles,
timeouts and client close. These are synthetic SDK executions, not cmdlet or
tenant recordings. Reports retain `coverage_claim: false` and live qualification
`not-run`; successful journalled tenant acceptance remains outstanding.
