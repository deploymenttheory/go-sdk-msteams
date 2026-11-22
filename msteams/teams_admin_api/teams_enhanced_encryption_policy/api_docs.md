# Teams enhanced encryption policy

`Client.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies` exposes native modern
configuration CRUD/list and user/tenant/group assignments. PowerShell is used
only by development captures and separately enabled lab comparisons.

| Go method | MicrosoftTeams 8.1.0 command / parameter set |
|---|---|
| GetByIdentity, List | Get-CsTeamsEnhancedEncryptionPolicy / Identity |
| ListFiltered | Get / Filter |
| CreateByIdentity, Create | New / Identity; POST primitive or POST/Get composition |
| UpdateByIdentity, UpdateGlobal | Set / Identity; sparse PUT or explicit omitted-identity default |
| DeleteByIdentity | Remove / Identity |
| AssignToUser, Assign.User | Grant / Identity |
| AssignTenantDefault, Assign.Tenant | Grant / GrantToTenant |
| AssignToGroup, Assign.Group | Grant / GrantToGroup |

Settings contains optional Description, CallingEndToEndEncryptionEnabledType and
MeetingEndToEndEncryption strings. Nil omits a field; a pointer to an empty string
sends it explicitly. The public declaration binder converts null strings to empty.
Private calling-mode metadata lists Disabled and DisabledUserOverride, but the
captured modern property-bag validator and declaration binder do not enforce that
set. Future mode strings remain representable. Server and legacy constraints
still require qualification.

Explicit identity methods use one escaped route component, preserving literal
whitespace. Empty identity rejects as the private command's attributes do. The
mapped Get helper's whitespace/omitted selection is available through List;
ListFiltered treats whitespace as the complete collection and otherwise applies
local Identity/Tag wildcard matching. Global-first partitioning is stable and
case-insensitive. UpdateGlobal applies only the omitted-identity default. A bound
blank identity is not silently rewritten to Global.

CreateByIdentity POSTs once and retains HTTP creation metadata. Create explicitly
composes POST and a selected Get. ReadBackError retains CreationResponse if the
separate read fails; the returned partial Policy retains the submitted identity.
PUT sends only bound settings plus Identity and permits Identity-only updates.
Delete is bodyless. Mutations never replay, implicitly roll back, wait or add
read-back beyond the explicit Create composition.

Assign selects exactly one User/Tenant/Group scope. Empty PolicyName sends the
captured clear form; no effective-clear guarantee is implied. User/tenant grants
require HTTP 200. The tenant primitive exposes the generated helper's Boolean
Force; the public family does not expose a Force parameter. Group grants retain
202 acceptance and operation metadata separately from 204. Only an actual POST
response can establish acceptance, including after an accepted malformed reply;
discovery metadata cannot. No polling is added.

Assign omits absent/zero group Rank as the mapped helper does. The lower-level
AssignToGroup accepts explicit zero priority, matching the generated DTO, and
preserves negative int32 values without guessing server constraints. PassThru is
not supported by the captured mapped grant function; Go returns typed results.

Policy snapshots own complete original JSON and unknown fields. Known values
project independently; unknown/unprojectable shapes stay raw with nil typed
projections rather than fabricated values. Ordinary policy/settings/result
formatting is redacted; raw snapshots/responses remain caller-owned sensitive
configuration. Requests send known settings without forwarding received fields.

The selected Teams ConfigAPI backend supplies forest discovery, audience/origin
validation and headers. The family uses modern headers without MPACmdlet. Native
API methods are independent of PowerShell release numbers and internal flight
switches. Synthetic application/delegated comparisons establish wire behavior,
not tenant authorization, persistence or effective encryption. See
[unit tests](service_test.go)
and the [lab acceptance record](../../../docs/implementation/policies/enhanced-encryption-lab-acceptance.md).
Issue #117 and full command/tenant qualification remain open.

## Owned CRUD acceptance

`TestAccEnhancedEncryptionPolicyCRUD` creates a uniquely named unassigned policy with
non-default settings, verifies read-back, updates its description, checks list
membership, deletes it and confirms GET404 plus list absence. `CallingEndtoEndEncryptionEnabledType: Disabled` survives a description-only update.
Created/updated reads are compared with MicrosoftTeams 8.1.0. Certificate and
client-secret lab runs passed with independently verified cleanup.

`TestUnitOwnedCRUDUsesAcceptanceResponses` loads seven actual responses from
`testdata`, checking requests, typed projections, sparse updates, native statuses
and deleted-resource errors. CI requires four CRUD milestones, two pinned
comparisons per test and verified cleanup.

Assignments/clear, effective client behavior, delegated authentication, other
clouds and remaining parameter scenarios still need acceptance; the feature
ticket remains open.
