# Teams IP phone policies

Use `sdk.TeamsAdminAPI.TeamsIPPhonePolicies`. The public 8.1.0
Get/New/Set/Remove/Grant-CsTeamsIPPhonePolicy functions are remoting
proxies whose `CmdletConfig.json` entries select the modern ConfigAPI helpers
(`ConfigType`/`PolicyType` `TeamsIPPhonePolicy`) and Grant-CsTeamsPolicy.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Modern GET, bare array, stable case-insensitive Global first, local Identity/Tag wildcards |
| `CreateByIdentity`, `Create` | Native POST/201 or explicit POST/Get composition retaining partial success |
| `UpdateByIdentity`, `UpdateGlobal` | One sparse PUT/204; explicit helper preserves omitted-Identity Global default |
| `DeleteByIdentity` | Bodyless DELETE/204 |
| `Assign`, `AssignToUser`, `AssignToGroup`, `AssignTenantDefault` | Modern per-scope grant requests and explicit receipts |

Settings has optional Description, AllowBetterTogether, AllowHomeScreen,
SearchOnCommonAreaPhoneMode and SignInMode strings, a `*bool` AllowHotDesking
and a `*int64` HotDeskingIdleTimeoutInMinutes. Nil omits a setting; empty, false
and zero remain explicit. The public Boolean and Int64 are non-nullable. The
pinned proxy declares no enum or range validation, so mode strings and timeouts
are sent unchanged and server validation applies.

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
creation with SignInMode=CommonAreaPhoneSignIn, AllowHotDesking=false and a
30-minute HotDeskingIdleTimeoutInMinutes, a sparse description/AllowHomeScreen/
AllowBetterTogether PUT preserving every other field, two pinned getter
comparisons, list/delete and independent GET404/list absence. Intent is journaled
before create. The tested role is Teams Administrator; policies remain unassigned
and Global is not changed. Actual sanitized responses feed unit tests. Grant
persistence/propagation, effective device behavior, SearchOnCommonAreaPhoneMode on
a live write, null/empty persistence, delegated modes and other clouds remain
separate.

See [qualification](../../../docs/acceptance/ip-phone-policy-lab-acceptance.md).
