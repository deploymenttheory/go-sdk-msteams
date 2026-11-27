# Teams recording rollout policies

Use `sdk.TeamsAdminAPI.TeamsRecordingRollOutPolicies` for the pinned compiled
Get/New/Set/Remove/Grant surface. Public Go method names do not include a module
version.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Direct compiled GET, bare collection, local wildcard selection |
| `CreateByIdentity` | One POST without implicit Global read |
| `PatchByIdentity`, `UpdateByIdentity` | Sparse PATCH or explicit Set read/PATCH composition |
| `DeleteByIdentity` | DELETE with Identity and optional Force body |
| `Assign`, `AssignUser`, `AssignGroup`, `AssignTenantDefault` | Explicit compiled user/group/tenant grants |

MeetingRecordingOwnership is an opaque nullable source string, with
policy.Field[string] preserving omission/null/empty bindings. No enum restrictions
or deprecation behavior are invented locally. The lab demonstrated persistence
of MeetingInitiator and MeetingOrganizer; this does not establish where an
actual meeting recording is stored.

The shared compiled provider preserves first-slash identity escaping, exact Global
ordering, literal filter whitespace, MPACmdlet/command headers and optional string
Force. It performs no modern tenant discovery. Writes never replay automatically.
RawFields and owned RawJSON retain the full response. UpdateResult keeps read
preparation and HTTP acceptance independently of later decode/read-back failures.
Grant requires one explicit scope and preserves nullable policy names, signed group
rank and tenant Force; propagation needs independent qualification.

Both application auth modes passed Global read/list and owned native CRUD. The
owned policy starts with MeetingInitiator and is updated to MeetingOrganizer;
read-back and two pinned PowerShell comparisons verify persistence. The test
checks that the actual sparse PATCH body contains only Identity and the selected
setting. It then lists, deletes and requires independent typed GET404 plus list
absence. Recovery intent precedes create. The tested role is Teams Administrator.
No Global/default assignment, user/group, license or privilege is changed.

Nine permanent sanitized service response files supply unit response fixtures.
Grant persistence, effective storage behavior, null/empty persistence, delegated
flows and other clouds remain separately unqualified. The pinned module/source
and actual API results are retained as the baseline contract; a current command
reference URL redirects to the general module index.

See [qualification](../../../docs/acceptance/personal-recording-lab-acceptance.md).
