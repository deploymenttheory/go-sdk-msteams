# Teams channels policies

Use `sdk.TeamsAdminAPI.TeamsChannelsPolicies` to manage TeamsChannelsPolicy.
The MicrosoftTeams 8.1.0 exports select compiled CRUD and modern grants. Native
SDK calls require no PowerShell runtime.

| Operation | API | Protocol |
|---|---|---|
| Read/list/filter | GetByIdentity / List / ListFiltered | Compiled GET; local Identity/Tag wildcards, exact Global first |
| Create | CreateByIdentity | One compiled POST |
| Sparse update | PatchByIdentity | One compiled PATCH; at least one setting |
| Set composition | UpdateByIdentity | Explicit Set-metadata GET then sparse PATCH |
| Delete | DeleteByIdentity | Compiled DELETE with Identity/optional Force JSON body |
| Assign | AssignToUser / AssignTenantDefault / AssignToGroup / Assign | Modern generated grants with session forest discovery |

Pass full identities explicitly, such as Global or Tag:Lab. Compiled filter
whitespace is literal. Create does not perform the source Global comparison used
for ShouldProcess presentation or implicit read-back. Set composition retains
Before, ReadResponse, PatchInvoked and HTTPAccepted on partial failure; the read
model Identity, including null, goes in the body while the route stays requested.
Only bound fields are patched. No ETag or atomicity guarantee is added.

## Settings

All 20 source settings are represented independently by `policy.Field[T]`: 14
strings and six Booleans. A zero Field omits a value; `policy.Set` binds a value;
`policy.Null` binds JSON null. Empty strings and false remain explicit.
AllowPrivateChannelCreation is Boolean; AllowCreatePrivateChannel is a separate
string mode. Newer creation, outside-team membership/sharing and default-type
fields are preserved alongside the legacy permissions. The captured setters do
not enforce enums, Description length or cross-field rules. Future string values
are allowed locally; server rules and effective behavior need tenant validation.

```go
settings := teamschannelspolicy.Settings{
    AllowPrivateChannelCreation: policy.Set(false),
    AllowCreatePrivateChannel: policy.Set("FutureMode"),
    Description: policy.Null[string](),
}
_, response, err := sdk.TeamsAdminAPI.TeamsChannelsPolicies.PatchByIdentity(
    ctx, "Tag:Lab", teamschannelspolicy.PatchRequest{Settings: settings},
)
```

FutureMode demonstrates type/presence, not a qualified server value. Read models
retain nullable known fields, RawFields and an owned RawJSON snapshot. Marshaling
a read model returns the snapshot; change request settings to update a resource.
Known JSON fields require actual string/Boolean types. Arbitrary Newtonsoft
primitive coercions and the public binder are separate from this native contract.

## Grants and acceptance

User/tenant grants require HTTP 200. Group grants retain 202 acceptance separately
from 204, including malformed accepted payload errors and unknown operation
fields, without polling. Discovery 202 can never establish group submission.
PolicyName is an explicit string; empty sends the mapped helper's captured clear
form, but clearing/effective behavior is unqualified. Assign requires one scope
and omits nil/zero Rank like the mapped wrapper. AssignToGroup accepts optional
Int32 priority and can send explicit zero. Tenant helper Force is Boolean; the
public source family does not expose Force. Compiled CRUD Force is a bound string.

All mutations disable replay. HTTP submission/acceptance does not prove
persistence or effective channel restrictions. There is no implicit restore,
elevation or completion guarantee. Runtime PowerShell PassThru/flight behavior is
not inferred from isolated mapped function observations.

See [unit tests](service_test.go)
and [lab acceptance](../../../docs/acceptance/channels-policy-lab-acceptance.md).
All five commands/eight sets have native mappings. Full command/tenant
qualification and issue #111 remain open.

## Owned CRUD acceptance

`TestAccChannelsPolicyCRUD` creates a uniquely named unassigned policy with a
non-default description and `AllowPrivateChannelCreation: false`. A description-only update
preserves that omitted flag. The test verifies list membership, deletes the
policy and confirms GET404 plus list absence. Both created/updated reads are
compared against MicrosoftTeams 8.1.0. Certificate and client-secret lab runs
passed; cleanup is independently verified.

`TestUnitOwnedCRUDUsesAcceptanceResponses` loads seven actual acceptance
responses from `testdata`, covering requests, read projections, sparse updates,
native statuses and typed absence. CI requires each owned CRUD test to reach
all four milestones, two PowerShell comparisons and verified cleanup.

User/tenant/group assignments and clear operations, effective client behavior,
delegated authentication, other clouds and remaining parameter scenarios still
require acceptance. The feature ticket stays open.
