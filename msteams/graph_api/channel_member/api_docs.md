# Channel members

`sdk.GraphAPI.ChannelMembers` owns native conversationMember resources under
`/beta/teams/{team}/channels/{channel}/members`, with `Create`, `GetByID`,
`UpdateByID`, `DeleteByID`, one-page `List`, lazy `NewPager`, complete bounded
`ListAll`, `PromoteOwner`, `DemoteOwner` and explicit `ResolveByUser`.

Construction acquires no credentials. Every request uses the selected Graph
cloud/audience. Team and directory-user IDs are UUIDs; channel and conversation
membership IDs remain opaque, independently escaped segments. The returned
membership ID is retained for subsequent updates/deletion. It is never decoded
or synthesized from a private channel and user ID. Native membership deletion
does not delete the directory user or a backing group.

Create requires 201 and a member response; Get/Update require 200 plus a member;
Delete requires 204. Create defaults to explicit `roles:[]` for an ordinary
member; `["owner"]` selects an owner. Update nil Roles omits the field, explicit
empty roles demotes and `["owner"]` promotes. JSON-null role arrays, `member`,
`guest`, unknown and repeated write roles are rejected locally. Read role/type
values stay open and RawFields retain future properties and null/absent fields.
An optional explicit tenant UUID is retained on creation; it does not trigger
a host-directory lookup or assert that the member belongs to the host tenant.

ResolveByUser requires complete bounded membership traversal and selects the
actual returned opaque ID. With no tenant selector, multiple matching directory
users are ambiguous. With a tenant selector, returned tenantId must match. An
incomplete traversal never becomes a missing/successful resolution. Nil/empty
collections, malformed resource envelopes and unexpected successful status codes
fail while retaining response metadata. Listing supports filter/select/top,
with documented maximum page size 999. Paging validates continuation origins;
failure returns partial items and a nonnil error. Writes are not replayed.

Current documented app permissions are ChannelMember.Read.Group (resource-specific
consent) or ChannelMember.Read.All/ChannelMember.ReadWrite.All for reads;
ChannelMember.ReadWrite.Group or ChannelMember.ReadWrite.All for mutations.
Private/shared eligibility, shared-channel member access, existing team membership,
external-user access and owner rules remain service restrictions. A Teams role
does not replace Graph consent. A disabled/blocked user can yield 404 on create;
that is not proof that an earlier creation did not happen.

These native contracts use documented empty/owner roles and configured-cloud
bindings. Pinned PowerShell sent `["member"]`, retained supplied Owner casing,
used commercial binding strings and synthesized private membership IDs. The
native resource API does not claim byte-equivalence to those variants.
`GraphAPI.ChannelUserManagement` composes the three channel-name/user operations
with explicit add/promote/demote/delete selection and UPN enrichment. Its
[API notes](../channel_user_management/api_docs.md) document the remaining
PowerShell differences and qualification limits. CORE-20 remains open.
No live acceptance or unsupported-feature claim is made.

Sources reviewed 2026-11-21:
[list](https://learn.microsoft.com/en-us/graph/api/channel-list-members?view=graph-rest-beta),
[create](https://learn.microsoft.com/en-us/graph/api/channel-post-members?view=graph-rest-beta),
[get](https://learn.microsoft.com/en-us/graph/api/channel-get-members?view=graph-rest-beta),
[update](https://learn.microsoft.com/en-us/graph/api/channel-update-members?view=graph-rest-beta),
[delete](https://learn.microsoft.com/en-us/graph/api/channel-delete-members?view=graph-rest-beta),
[pinned evidence](../../../docs/implementation/core/channel-contract-evidence.md).
