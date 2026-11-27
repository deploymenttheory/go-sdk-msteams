# Channel user management

`sdk.GraphAPI.ChannelUserManagement` composes independent native channel and
membership services. All dependencies share one root-client Graph scope.
Construction performs no authentication, discovery or HTTP.

| Method / option | Operation |
| --- | --- |
| `Add`, omitted Role | Resolve a unique channel and directory user, then create ordinary channel membership |
| `Add`, Role Owner | Resolve an existing membership and promote it; does not add a missing member |
| `Remove`, omitted Role | Resolve the existing membership and delete it |
| `Remove`, Role Owner | Resolve the existing membership and demote it, retaining membership |
| `List` | Resolve a unique channel, completely traverse native members, project/filter roles and enrich selected user names |

Only omitted Role or Owner is valid for writes. Read filters accept Owner,
Member or Guest. Recognized role names compare without case. Standard-channel
membership changes return the local ErrStandardChannel; reads remain allowed.
Server authorization, private/shared eligibility and last-owner restrictions
remain authoritative. An unknown/missing membership type is never silently
treated as a known writable kind.

Channel names use complete bounded unique lookup. Private owner/removal actions
resolve UUID/UPN through the shared directory resolver, then select the actual
returned opaque membership ID from a complete roster. Shared actions match the
returned user ID or email, retaining foreign/null user information without a
host-directory lookup. Multiple matches are ambiguous; a selected tenant UUID
must match returned tenantId. Creation with an explicit TenantID requires an
explicit user UUID, avoiding a host-directory lookup for a foreign identity.
Native bindings follow the selected Graph cloud.

Write results retain channel/user/membership IDs, the native member when known,
response metadata and succeeded/unconfirmed stages. A failed later step never
causes automatic rollback or replay. Unconfirmed does not establish absence of
a remote mutation. Stage/result formatting is redacted; inspect server data only
explicitly. Caller deadlines and session cancellation span every composed step.

Read projection preserves membership identity, user ID, tenant, raw membership,
nullable name and initial UPN/email. Owner takes precedence over guest; absent,
empty or unknown role qualifiers project as Member. Without a filter, ordering
is stable owner/member/guest. Filtering precedes enrichment. Selected present
UUIDs are deduplicated and queried through `/beta/users` in groups of at most
ten, including users with an initial UPN. All user pages must complete. A returned
null UPN clears the initial projection; an absent directory user retains it.
Unknown members with no user ID retain their available native/email fields.
Duplicate/unrequested directory results, denied reads and invalid continuations
fail the workflow. Complete is false on any failure; native partial members and
the last response remain available for inspection. An empty selected set sends
no directory query. It does not use Graph `$batch` for this enrichment.

The pinned PowerShell functions used first-page channel lookup, private
membership-ID synthesis, commercial bind URLs, `["member"]` write qualifiers
and retained Owner casing. Add's normal branch bound User directly without
directory resolution; its owner branch ignored TenantId. The SDK deliberately
uses complete unique resolution, returned IDs, explicit tenant disambiguation,
current documented role payloads and UUID bindings. Its conventional distinct-ID
chunks differ from the pinned helper's first-short-chunk ordering. These choices
preserve intended administration while remaining separate from byte-equivalence
and current-tenant acceptance claims.

Channel reads need effective channel-read access. Member reads/writes need the
corresponding ChannelMember application/delegated permissions. UPN resolution
and read enrichment additionally need directory-user read permission; app
enrichment uses User.Read.All. A Teams directory role does not replace Graph
consent. No permission is granted by the SDK or examples.

See [unit tests](service_test.go),
[unit tests](service_test.go)
and [gated lab procedure](../../../docs/acceptance/channel-membership-lab-acceptance.md).
