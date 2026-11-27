# Teams network roaming policies

Use `sdk.TeamsAdminAPI.TeamsNetworkRoamingPolicies` for the public 8.1.0
Get/New/Set/Remove functions. The public proxies select modern ConfigAPI helpers;
older compiled classes in the same package are not this service's wire contract.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Modern GET, bare array, case-insensitive Global first, local Identity/Tag wildcards |
| `CreateByIdentity` | One POST requiring HTTP 201; returns the native submission receipt |
| `Create` | Explicit POST + Get composition; ReadBackError retains accepted creation if Get fails |
| `UpdateByIdentity`, `UpdateGlobal` | One sparse PUT requiring 204; explicit Global helper preserves omitted-Identity Set default |
| `DeleteByIdentity` | Bodyless DELETE requiring 204 |

The shared adapter discovers Teams tenant routing and sends X-MS-Target-Uri,
Accept and X-MS-CmdletName. It does not send MPACmdlet. Whole identities remain
one escaped route segment. Blank filters select all; malformed filters and invalid
identities fail before authentication. Lists have no nextLink/pagination contract.

Settings uses *string Description, *bool AllowIPVideo and *int64 MediaBitRateKb.
Nil omits a setting, while empty/false/zero values bind explicitly. The public
Boolean/Int64 parameters are non-nullable; the SDK does not invent JSON-null
bindings or undocumented numeric limits. Identity-only Set is accepted by the
source wrapper. The public proxy has no Force parameter. Mutations never replay.

Policy projects known nullable response fields and retains the owned RawJSON and
RawFields, including future properties and values that cannot fit a known typed
projection. Marshaling a received model returns its original complete snapshot.
An HTTP receipt is separate from read-back and effective network behavior.

Both certificate and client-secret acceptance read/list Global and manage a
uniquely named, unassigned policy with a non-default description, AllowIPVideo=false
and an explicit bitrate. A description-only PUT must preserve both omitted
settings. The test compares all three getters with PowerShell before/after,
lists, deletes and independently verifies typed GET404 and collection absence.
The tested lab SPN has Teams Administrator. Actual sanitized responses supply
unit fixtures; read-back failures, replay safety, integer precision, invalid input
and context handling have separate contract checks. No Global configuration,
network site, user/group assignment, license or role is changed.

See [qualification](../../../docs/acceptance/network-notification-lab-acceptance.md).
