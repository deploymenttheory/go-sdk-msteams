# Conditional conversation-role contracts

Owner: EXT-90 (#191). Pinned baseline: MicrosoftTeams 8.1.0. These observations
settle local declaration and request-building behavior, not successful server
acceptance or permission persistence. No backend role command, sign-in or tenant
operation was executed.

## Declarations and scope

The original conditional snapshot contains six commands and twelve sets:

| Command | Declared sets | Observed management behavior |
| --- | --- | --- |
| Get-CsConversationRole | Identity, ParentAndRelativeKey, Filter | Bare collection, exact identity or local wildcard filtering; relative Id becomes Global/Id. |
| New-CsConversationRole | Identity, ParentAndRelativeKey | Required permission DTO and CreatedAt/UpdatedAt; POST after selected-field projection and required scope metadata. |
| Set-CsConversationRole | Identity, ParentAndRelativeKey | Preparation GET then selected-field PATCH; the relative selector also binds Id. |
| Remove-CsConversationRole | Identity, ParentAndRelativeKey | Exact identity DELETE; relative Id becomes Global/Id. |
| Get-CsConversationRolesSetting | Identity, Filter | Bare collection, exact identity or local wildcard filtering. |
| Set-CsConversationRolesSetting | Identity | Preparation GET, local list transition, then projected full-list PATCH. |

The existing fourteen owned conditional sets, these twelve, and the license
report's one set account for all twenty-seven declarations in the original
fifteen-candidate snapshot. They remain separate from the 108 immutable
ordinary-session sets.

The audit additionally found three compiled local builder types:
`New-CsTeamsChatPermissions`, `New-CsTeamsCallingPermissions`, and
`New-CsTeamsConversationRolePermissions`. Each declares one Identity set in the
[supplemental metadata](evidence/conversation-builder-metadata.json). None is
exported in the imported root session, and none is named by the pinned root,
OCE or preview `CmdletsToExport` lists. Exact-name searches of official command
documentation returned no results; negative search is not an exclusion proof.
These three supplemental declarations are not silently added to accepted runtime
coverage, classified as public exports, or assigned implementation qualification.
Their local behavior can still inform typed DTO composition.

## Local builder and serializer observation

The [reproducer](../../../scripts/discovery/inspect-conversation-role-serialization.ps1)
verifies Core and PolicyRp assembly hashes before using their data types. Core
SHA256 is `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`;
PolicyRp SHA256 is `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`.
The [capture](evidence/conversation-role-serialization.json) records its exact
extractor hash and controlled values. The [full IL snapshot](evidence/conversation-role-il.json)
records exact method-body hashes, instructions, exception regions and assembly
hashes for construction, scope adaptation and management request projection;
its read-only decoder never invokes the inspected methods.

Only the three local permission builders are invoked via their compiled
`CmdletInfo`. Their source paths construct data, validate properties and wrap
with `PSObject.AsPSObject`. The inherited lifecycle sets local invocation/logging
state and default JSON maximum depth; its telemetry callback dereferences the
host contract. The reproducer refuses invocation unless that host contract is
absent and verifies it remains absent afterward. It never invokes the management
role command or an authentication/HTTP method.

Actual local outputs retain explicit true/false values, with eighteen nullable
chat flags and thirteen nullable calling flags. There are no extra NoteProperty
or ScriptProperty members carrying hidden binding metadata. The typed role
constructor consumes the unwrapped CLR permission DTO and CLR DateTime values.
The latter preserve unspecified versus UTC clock forms; request dates are not
restricted to RFC3339 timestamps.

The controlled role parameter dictionaries and required metadata follow the
inspected declarations and `PrepareForScopeCheck`; they are not captured live
management invocation data. They pass through the actual
`GetNestedBoundParameters(..., false)`, `BoundParametersContractResolver`,
`AddAdditionalBodyContent`, and (for creation only)
`AddRequiredCmdletMetadata` methods.

The resulting observed request bodies contain `Permissions: {"Identity":null}`,
while the actual compiled local-builder output contained populated Chat/Calling flags.
Additional-body handling does not replace an already present Permissions key.
Thus this pinned pipeline observation demonstrates a local projection loss; it
supplies no positive evidence that supplied flags persist on the server.

The create selectors also differ. Full Identity projection omits unbound Id,
then required metadata adds the derived Id. Binding the relative Id causes an
existing `Id:null` property to be serialized; required metadata adds only absent
keys and leaves that null unchanged. Do not silently copy the relative selector
into the model body and call it pinned parity. Sparse name-null and false
IsAssignable projections are retained separately.

## List transitions and full-list projection

The [list helper reproducer](../../../scripts/discovery/inspect-conversation-role-list.ps1)
and [thirty-case observation](evidence/conversation-role-list-transitions.json)
exercise the exact compiled removal predicate and helper without a cmdlet or
backend operation:

- Null current collection rejects every modifier.
- Null or all-empty modifier clears a nonnull collection.
- Nonempty Replace takes precedence over Add and Remove.
- With no replacement, removal on an empty collection performs no comparison;
  addition can then append.
- Removal on a nonempty collection invokes the predicate and rejects, even if
  the requested element is absent. The predicate loads
  `ErrorCannotRemoveElementsOf("ConversationRoles")` and throws; the offline
  resource-message path's exact exception type is reported without claiming it
  matches a connected host's error rendering.

The [full Set-list reproducer](../../../scripts/discovery/inspect-conversation-role-list-serialization.ps1)
and [capture](evidence/conversation-role-list-serialization.json) use actual local permission output:
role scalar fields and dates survive, but nested Chat/Calling flags are omitted
from both retained and added roles. A client must not accidentally interpret
this as an independently proven native full-permission update contract.

## Implementation boundary

Typed reads and local DTO construction preserve flags and raw presence. Pure
`ProjectCreateByIdentity`/`ProjectCreateByID` helpers return an opaque request
whose owned `Projection()` exposes the exact body, omitted paths and changed
paths before mutation. `PrepareUpdateByIdentity`/`PrepareUpdateByID` performs
only the preparation read and local source projection; explicit
`UpdateByIdentity`/`UpdateByID` applies that frozen body. Settings uses the same
prepare/inspect/apply separation with the observed list transitions.

Accepted Chat/Calling values are not rejected or silently replaced by guessed
native JSON. Projection reports their source omission, and the caller can
inspect and explicitly apply it. Source selector differences, including relative
creation's null Id and relative Set's current model Id, remain intact. Prepared
requests bind to the exact preparing service/client, selector and parameter set;
owned copies prevent a changed input, before snapshot or projection from
retargeting the frozen wire. Results retain preparation and action receipts.
There is no extra approval flag or implicit delegated fallback.

RemoveAsync differs from creation: it starts with an empty object and adds
original bound parameters only; it never applies required-metadata body merge.
Identity deletion therefore sends Identity, while relative deletion sends Id.
Both optionally include the observed Force string. These variants have distinct
request capture cases and tests.

The [presence reproducer](../../../scripts/discovery/inspect-conversation-role-presence.ps1)
and [capture](evidence/conversation-role-presence.json) settle local JSON
conversion: omitted role Permissions/dates retain CLR defaults, but explicit
null Permissions, nested Chat/Calling and dates reject. Omitted/null setting
collections both normalize to empty, while null array elements survive. Typed
Go values follow these projections and separately retain original raw presence.

Offline request vectors are actual pinned local captures; HTTP replies remain
authored model vectors. Local TestUnit/Example checks exercise all twelve
management sets, the thirty-case list matrix, complete local-builder output,
client/identity binding, cancellation, uncertain-write receipt retention and
independent recovery after settings authorization denial. These checks do not
prove server acceptance, ownership, authorization or effective permissions.

No command is marked fully implemented or contract-complete. The six full
command contracts, full permission-write semantics and persistence, supplemental
builder public scope and actual tenant behavior remain unresolved until their
full criteria have revision-backed execution evidence.

## No-settings selection correction

The [exact concrete/base observation](evidence/conditional-set-selection.json)
and [reproducer](../../../scripts/discovery/inspect-ext-set-selection.ps1) verify
all four EXT conditional Set overrides. Each uses BoundParameters.ContainsKey,
not a truthy value or nullable HasValue test. Once bound, null, false, zero or
empty counts as selection; separate type/model validation still applies. Only
the actual mutable field names count, not Force or selector Id/Identity.

With no mutable key the shared BaseTpmSetCmdlet.ProcessRecord constructs
ProcessRecordFailed/category7, invokes WriteError at IL115, and branches out
before token handling at IL145. Its exact method hash is
`c06e7863f0da89a8ba5fd748b249139faf88fb6557fe7e01ad90a9fb14cf7b01`.
The native conversation Prepare helpers therefore return ErrInput before any
read/authentication, rather than the initially authored success/no-op outcome.
Tenant/app configuration already rejected this case; focused Force-only tests
keep all four resources aligned. Bound null ConversationRoles remains a selected
list clear, and explicit false/null scalar projections remain supported.
