# Teams room video teleconferencing policies

Use `sdk.TeamsAdminAPI.TeamsRoomVideoTeleConferencingPolicies`. The public 8.1.0
Get/New/Set/Remove/Grant functions select modern ConfigAPI helpers. Compiled classes
in the same package do not replace the public proxy wire contract.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Modern GET, bare array, stable case-insensitive Global first, local Identity/Tag wildcards |
| `CreateByIdentity`, `Create` | Native POST/201 or explicit POST/Get composition retaining partial success |
| `UpdateByIdentity`, `UpdateGlobal` | One sparse PUT/204; explicit helper preserves omitted-Identity Global default |
| `DeleteByIdentity` | Bodyless DELETE/204 |
| `Assign`, `AssignToUser`, `AssignToGroup`, `AssignTenantDefault` | Modern per-scope grant requests and explicit receipts |

Settings includes optional Description, AreaCode, ReceiveExternalCalls,
ReceiveInternalCalls, PlaceExternalCalls and PlaceInternalCalls strings, plus
*bool Enabled. Nil omits a setting; empty/false remain explicit. Public Enabled
is non-nullable. There is no public Force setting; the low-level tenant grant can
represent the underlying helper's forceSwitchPresent, while Assign uses false.
Public PassThru controls PowerShell output; native operations return their
receipts directly. Group Rank is nullable Int32: Assign omits zero like the
wrapper, and AssignToGroup can explicitly send zero. A 202 receipt is distinct
from a 204 and from effective assignment.

The shared adapter performs tenant discovery and supplies X-MS-Target-Uri, Accept
and command metadata without MPACmdlet. Identities are escaped as whole segments;
blank filters select all. Mutations never replay. Typed nullable read projections
retain the complete owned RawJSON/RawFields, including unknown/unprojectable data.

Both application auth modes passed Global read/list and owned native CRUD:
non-default description creation with Enabled=false, a description-only PUT
preserving the Boolean, two pinned getter comparisons, list/delete and independent
GET404/list absence. Intent is journaled before create. The tested role is Teams
Administrator; policies remain unassigned and Global is not changed. Actual
sanitized responses feed unit tests. Grant persistence/propagation, effective room
calling, null/empty persistence, delegated modes and other clouds remain separate.

See [qualification](../../../docs/implementation/policies/room-shared-calling-lab-acceptance.md).
