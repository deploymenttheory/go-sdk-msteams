# Team operation status

`sdk.GraphAPI.TeamOperations` provides the explicit `/beta/teams/{teamID}/operations/{operationID}`
status resource used by MicrosoftTeams 8.1.0 Get-Operation. It is separate from
the existing native team service's v1.0 operation handles. Construction is local.

| Method | Behavior |
| --- | --- |
| `GetByID` | Read one status snapshot; require HTTP 200; preserve failed/future status as data |
| `ResumeByID` | Construct a polling handle locally from explicit team/opaque operation IDs |
| `WaitByID` | Explicit bounded waiting using the shared waiter and beta status decoder |
| `Operation.State` | Interpret recognized states; return ErrOperationState for absent/unknown values |

Get is a read, so a snapshot whose status is failed is returned successfully.
Wait reports ErrOperationFailed and retains normalized diagnostics/last response.
Unknown status is never successful completion. HTTP 202 alone cannot establish
completion. Missing identities, malformed known fields and unexpected successful
Get statuses fail while retaining the response.

Team IDs must be UUIDs. Operation IDs remain opaque individually escaped path
segments, distinct from a target resource ID. Nullable timestamps/zero attempts,
open operation type/status values, target identity/location and failure fields
remain available, with complete RawFields for original presence/future data.
Snapshot/error formatting is redacted. Inspect server diagnostics explicitly.
TargetResourceLocation remains opaque and is never parsed or followed.

Get uses bounded shared read retries with a 31-second retry floor. Resumed waits
make their first observation immediately; later polls and retries observe at
least 31 seconds and longer Retry-After values. Caller cancellation/session close
stop local work while snapshots retain the remote identity. Waiting does not
resubmit the initiating mutation. Persist handles privately, and preserve the
last observation time if resuming outside an existing in-memory handle.

Resource mutations initiate operations through their own services. The pinned
command establishes a status read, rather than direct operation-record mutations.
Graph-generated operation navigation wrappers include additional verbs, but
their service-specific accepted inputs, permissions and lifecycle effects still
need separate evidence. They are not emitted as guessed create/update/delete or
remote-cancel requests here. The true SDK lifecycle requirement remains tracked;
no generic unsupported implementation substitutes for missing contracts.

The pinned assembly establishes the team route. The current
[resource reference](https://learn.microsoft.com/en-us/graph/api/resources/teamsasyncoperation?view=graph-rest-beta)
describes fields and a polling interval greater than 30 seconds. Its linked
[Get reference](https://learn.microsoft.com/en-us/graph/api/teamsasyncoperation-get?view=graph-rest-beta)
currently documents chats and chat permissions. Those permissions are not
transferred to this team route. Exact team-operation authorization and current
tenant behavior remain unqualified; no role/grant is assigned by this service.
The strict 200 Get contract is an explicit SDK expectation alongside the pinned
helper's broad successful-status handling, and requires tenant qualification.

See [source evidence](../../../docs/implementation/core/channel-contract-evidence.md),
[SDK evidence](../../../docs/implementation/core/team-operation-sdk-evidence.md) and
[read-only lab procedure](../../../docs/implementation/core/team-operation-lab-acceptance.md).
