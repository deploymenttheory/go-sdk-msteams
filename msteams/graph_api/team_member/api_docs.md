# Native team members

`sdk.GraphAPI.TeamMembers` manages Graph `conversationMember` resources. Its
membership IDs are opaque and distinct from Entra directory-user IDs. It does
not implement the backing-group semantics of Add/Get/Remove-TeamUser; those
compositions and the separate group member/owner reference services remain
CORE-26 work.

| Method | Graph v1.0 route | Success |
| --- | --- | --- |
| `Create` | `POST /teams/{teamID}/members` | 201 and the created member |
| `GetByID` | `GET /teams/{teamID}/members/{membershipID}` | 200 and one member |
| `UpdateByID` | `PATCH /teams/{teamID}/members/{membershipID}` | 200 and the updated member |
| `DeleteByID` | `DELETE /teams/{teamID}/members/{membershipID}` | 204, no resource |
| `List` | `GET /teams/{teamID}/members` | 200 and one page |
| `PromoteOwner`, `DemoteOwner` | The same membership PATCH route | 200 and the updated member |
| `NewPager`, `ListAll` | The members collection and its opaque next links | Bounded traversal; partial results on error |

Methods take the caller's context and return the shared Resty response metadata,
including on HTTP, decode or unexpected-success-status failures. Construction and
pager construction acquire no credentials. Mutations never replay, including
received 404/429/503 responses. A 404 during create can identify a disabled or
blocked user; it does not justify retrying or creating a different identity.

Team and directory-user inputs are UUIDs. `MembershipID` is a distinct Go type;
only use identities returned by native membership endpoints. The SDK does not
decode them, infer their format, copy `userId` into them, or assume all members
belong to this client's tenant. Their original contents are escaped as one path
segment. Manually converting a directory ID does not make it a membership ID.

`CreateRequest` supplies an Entra user ID; the service constructs the AAD
discriminator and user binding from the configured Graph cloud. A nil roles
pointer omits the property and accepts the ordinary-member default. An explicit
empty slice sends `[]`; owner creation sends `["owner"]`. A pointer to a nil
slice would be JSON null and is rejected locally. `guest`, `member`, duplicate
owner qualifiers and future write values are rejected rather than being silently
converted into a different role.

Role updates preserve omission (`{}`) separately from explicit demotion
(`{"roles":[]}`). `PromoteOwner` and `DemoteOwner` call `UpdateByID`; neither
deletes membership nor performs a hidden read. Guest is assigned by Graph and
cannot be changed with this patch. Graph rejects promotion of a guest; the SDK
does not infer guest status from an opaque identity or add a directory pre-read.
An update returns a resource, unlike the teams settings endpoint's empty 204.

Read models retain the AAD discriminator, membership ID, roles, display name,
user/tenant/email fields and unknown future subtype/role/property information.
Null and absent AAD fields remain unknown; empty strings/arrays remain explicit.
`RawFields` retains the complete read object. Chat-only history and source
membership URLs are readable metadata and never enter a team role write or
become an automatically followed destination.

`List` returns exactly one page and retains its next link without following it.
It accepts the documented filter/select/top parameters. Zero Top leaves Graph's
default of 100; values outside 0–999 fail locally. `ListAll` explicitly follows
continuations with shared same-origin checks, loop detection and traversal bounds.
`PagingOptions.Style` may be zero or `LinkPaging`, `Parameter` must be empty,
and `ExpectedStatus` may be zero or 200; this resource always requires 200.
Zero limits use 1,000 pages and 100,000 aggregate items. `MaxItems` limits
`Collect`/`ListAll` aggregation; callers using `Next` process pages individually.
Every page must satisfy HTTP 200 before decoding. Failure retains earlier items
and the last response with an error; that result is incomplete. Pagination never
replaces the opaque continuation query with the initial filters.

Native membership changes may take time to become visible. The documented
two-second staggering recommendation is caller scheduling guidance; no async
operation, hidden wait, retry or rollback is inferred. Create followed by role
promotion is a two-stage operation; retain the created membership identity if
the second stage fails.

Documented read permissions are delegated `TeamMember.Read.All` or application
`TeamMember.Read.Group` (resource-specific consent), with alternatives listed in
the endpoint documentation. Non-owner writes use
`TeamMember.ReadWriteNonOwnerRole.All`; that permission does not imply owner
authority. Owner management requires the appropriate `TeamMember.ReadWrite.All`
authority. Application guest addition is unsupported. These are documented
permission boundaries; successful lab access and least-privilege qualification
remain outstanding. The SDK grants nothing and never switches identities after
denial.

Primary sources verified on 2026-11-20:
[create](https://learn.microsoft.com/en-us/graph/api/team-post-members?view=graph-rest-1.0),
[get](https://learn.microsoft.com/en-us/graph/api/team-get-members?view=graph-rest-1.0),
[update](https://learn.microsoft.com/en-us/graph/api/team-update-members?view=graph-rest-1.0),
[delete](https://learn.microsoft.com/en-us/graph/api/team-delete-members?view=graph-rest-1.0),
[list](https://learn.microsoft.com/en-us/graph/api/team-list-members?view=graph-rest-1.0),
[AAD fields](https://learn.microsoft.com/en-us/graph/api/resources/aaduserconversationmember?view=graph-rest-1.0).
