# Conditional conversation-role settings

`TeamsAdminAPI.ConversationRoleSettings` uses compiled resource
`ConversationRolesSetting`, with Get Identity/Filter and Set Identity sets.
`GetByIdentity`, `List` and `ListFiltered` perform direct compiled reads. Empty
explicit Filter is rejected by its declared attribute; omitted Filter uses List.
No create, delete or grant contract is invented for this configuration.

`Configuration` preserves full raw response snapshots. The typed role collection
normalizes omitted and null to an empty collection, matching actual pinned JSON
deserialization; RawFields still distinguishes those wire forms. Null elements
inside a collection remain null. Nested roles use the typed local permission/date
models from `conversation_role`.

## Explicit list composition

`PrepareUpdateByIdentity` performs a Set-metadata GET and the exact local
PSListModifier transition, then returns an opaque request with an owned
`Projection()`, `Before()` and read receipt. It does not PATCH. Inspect the body
and OmittedPaths before applying with `UpdateByIdentity`. A prepared request is
bound to its preparing service/client and exact resource identity; copying or
modifying exposed views cannot alter the frozen body or transfer it to another
principal/target.

- Unbound ConversationRoles is NoOp, including Force alone.
- Bound null or all-empty modifier clears the current list.
- Nonempty Replace wins, bypassing Add and Remove.
- Otherwise Remove on an empty current list does nothing; Add then appends in
  order without deduplication. Remove on a nonempty list invokes the compiled
  throwing predicate and fails after preparation, even for an absent target.
- The standalone `ProjectListValue` rejects a null current collection, as the
  actual helper does. This differs from the response DTO normalization above.

The resulting body replaces the full ConversationRoles array; it is not a
server-side Add/Remove patch. Source projection omits nested Chat/Calling fields
from retained and new roles. The projection reports those omissions explicitly;
it accepts the supplied values and does not invent a successful full-permission
wire contract. Full flag persistence and successful management writes remain
unqualified. Request vectors are actual offline pipeline captures; replies are
authored model vectors, not tenant responses.

Apply retains Before, ReadResponse, Projection, PatchInvoked and HTTPAccepted on
failure, including accepted malformed replies and uncertain lost responses.
Writes use ReplayNever; a receipt is not persistence proof. The read/modify/apply
sequence is not atomic and provides no ETag or concurrent-edit guarantee.
String/GoString redact all request, result and model data. No live Global update,
new role/consent or license operation is configured.

See [source notes](../../../docs/implementation/ext/conversation-role-contract-notes.md)
for assembly/reproducer hashes, the thirty-case helper matrix and full-pipeline
observations. The issue and command qualification remain open.
