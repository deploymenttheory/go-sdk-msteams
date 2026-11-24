# Teams meeting broadcast policies

Use `sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies`. The public 8.1.0
Get/New/Set/Remove/Grant-CsTeamsMeetingBroadcastPolicy functions are remoting
proxies whose `CmdletConfig.json` entries select the modern ConfigAPI helpers
(`ConfigType`/`PolicyType` `TeamsMeetingBroadcastPolicy`) and Grant-CsTeamsPolicy.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Modern GET, bare array, stable case-insensitive Global first, local Identity/Tag wildcards |
| `CreateByIdentity`, `Create` | Native POST/201 or explicit POST/Get composition retaining partial success |
| `UpdateByIdentity`, `UpdateGlobal` | One sparse PUT/204; explicit helper preserves omitted-Identity Global default |
| `DeleteByIdentity` | Bodyless DELETE/204 |
| `Assign`, `AssignToUser`, `AssignToGroup`, `AssignTenantDefault` | Modern per-scope grant requests and explicit receipts |

Settings has optional Description, BroadcastAttendeeVisibilityMode and
BroadcastRecordingMode strings plus *bool AllowBroadcastScheduling and
AllowBroadcastTranscription. Nil omits a setting; empty/false remain explicit.
The public Booleans are non-nullable. The pinned proxy declares no enum validation,
so mode strings are sent unchanged and server validation applies.

Grant selects one scope: user (`Identity` set), group (`GrantToGroup`, nullable
Int32 Rank) or tenant (`GrantToTenant`). There is no public Force setting; the
low-level tenant grant can represent the underlying helper's forceSwitchPresent,
while Assign uses false. Public PassThru controls PowerShell output; native
operations return their receipts directly. Assign omits a zero group Rank like the
wrapper; AssignToGroup can explicitly send zero. A 202 receipt is distinct from a
204 and from effective assignment.

The shared adapter performs tenant discovery and supplies X-MS-Target-Uri, Accept
and command metadata without MPACmdlet. Identities are escaped as whole segments;
blank filters select all. Mutations never replay. Typed nullable read projections
retain the complete owned RawJSON/RawFields, including unknown/unprojectable data.

Both application auth modes passed Global read/list/filter and owned native CRUD:
creation with AllowBroadcastScheduling=false and BroadcastRecordingMode=AlwaysDisabled,
a sparse description/attendee-visibility PUT preserving every other field, two
pinned getter comparisons, list/delete and independent GET404/list absence. Intent
is journaled before create. The tested role is Teams Administrator; policies remain
unassigned and Global is not changed. Actual sanitized responses feed unit tests.
Grant persistence/propagation, effective live-event behavior, null/empty
persistence, delegated modes and other clouds remain separate.

See [qualification](../../../docs/implementation/policies/meeting-broadcast-policy-lab-acceptance.md).
