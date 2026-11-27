# Teams shared calling routing policies

Use `sdk.TeamsAdminAPI.TeamsSharedCallingRoutingPolicies` for the public compiled
8.1.0 Get/New/Set/Remove/Grant commands.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Direct compiled GET, bare array, exact Global first, local wildcards |
| `CreateByIdentity` | POST with scalar bindings and local list edits applied to a fresh empty list |
| `PatchByIdentity` | One sparse PATCH with complete replacement arrays |
| `UpdateByIdentity` | Set preparation GET, local list edits, sparse PATCH, independent receipts |
| `DeleteByIdentity` | DELETE with Identity/optional Force body |
| `Assign`, `AssignUser`, `AssignGroup`, `AssignTenantDefault` | Explicit compiled grant scopes |

Description and ResourceAccount use nullable strings and policy.Field[string]
requests. Settings.EmergencyNumbers is a *policy.StringList for native full-array
writes. CreateRequest/UpdateRequest use *policy.ListEdit: nil omits, ReplaceStrings
replaces/clears, and ChangeStrings removes first ordinal matches then appends.
The API receives an array rather than Add/Remove operators. Read/modify/write has
no atomicity or ETag guarantee. The source constructor creates an empty list;
missing input stays empty, and the source setter rejects an explicit null list.
Known/unknown fields and owned response bytes remain available in RawFields/RawJSON.

The compiled source initializes an emergency-number regex but the inspected setter
does not execute it. The SDK introduces no extra local phone/GUID/enum restrictions.
Server validation and typed API errors remain authoritative. Bound Force uses the
compiled string form. Routing splits a scoped identity at the first slash and
escapes the remainder; MPACmdlet/command headers do not use modern discovery.
Writes never replay. Group Rank is signed Int32 and scope selection is explicit.

Both auth modes passed Global read/list and CRUD for an unassigned policy without
telephone associations: non-default description, explicit empty emergency list,
sparse description update, two pinned comparisons, list/delete and independent
GET404/list absence. The tested role is Teams Administrator. This does not prove
populated callback-list behavior. The lab inventory reads succeeded and returned
zero resource accounts and zero phone numbers. Dial codes were rejected by the
service; they are not valid callback fixtures.

The full opt-in populated test remains pending an explicit same-tenant resource
account plus two real unassigned callback numbers. It creates with one number,
checks sparse preservation, Add/Remove to the second, clears, lists and deletes.
The native implementation and request-only list-edit tests remain available.
Fixtures and journals stay private; only sanitized actual responses enter service
testdata. No phone numbers are allocated, assigned or purchased by these tests.

[Microsoft's setup guide](https://learn.microsoft.com/en-us/microsoftteams/shared-calling-setup)
and [parameter reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamssharedcallingroutingpolicy?view=teams-ps)
describe resource accounts and emergency callback-number requirements. See
[qualification](../../../docs/acceptance/room-shared-calling-lab-acceptance.md).
