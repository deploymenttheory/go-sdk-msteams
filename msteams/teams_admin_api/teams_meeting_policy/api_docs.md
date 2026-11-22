# Teams meeting policies

Use `sdk.TeamsAdminAPI.TeamsMeetingPolicies`. The service runs native Go; PowerShell
8.1.0 is used only by opt-in acceptance comparisons.

| Operation | Method | Request behavior |
|---|---|---|
| Read | `GetByIdentity` | Direct compiled-provider GET with explicit identity |
| List | `List` | Bare collection, exact `Global` first |
| Filter | `ListFiltered` | Local Identity/Tag wildcard matching; whitespace stays literal |
| Create | `CreateByIdentity` | POST with identity and selected settings; no implicit read |
| Update | `PatchByIdentity` | One sparse PATCH; at least one bound setting required |
| Set composition | `UpdateByIdentity` | Set-metadata GET followed by sparse PATCH |
| Delete | `DeleteByIdentity` | DELETE with Identity and optional Force JSON body |
| Grant | `Assign`, `AssignToUser`, `AssignTenantDefault`, `AssignToGroup` | Generated ConfigAPI routing, separate from direct CRUD |

All 123 settings from the pinned public New/Set parameter sets have typed
`policy.Field[T]` request fields and nullable read fields. Zero fields omit a
setting, `policy.Set` binds false/empty/zero, and `policy.Null[T]` binds JSON null.
A Force-only update is rejected before authentication. Bound CRUD Force uses the
compiled string representation. Read models preserve unknown properties in
`RawFields` and their complete owned `RawJSON`; marshal returns the received
snapshot. Build write requests from settings rather than modifying a read model.

Identities are explicit (`Global`, `Tag:Name`, or an evidenced scoped identity).
A scoped identity splits at the first slash; the remainder stays one escaped
route segment. Update results retain preparation and HTTP acceptance even when
later decoding fails. Writes never replay automatically.

Grants follow the public wrapper's modern helper, including optional signed
Int32 group Rank. `Assign` requires exactly one scope and omits zero Rank as the
PowerShell wrapper does; `AssignToGroup` can explicitly send zero. Empty policy
names clear a direct assignment. Tenant `Assign` does not expose Force; the
lower-level tenant method maps `forceSwitchPresent`. Group 202 preserves an
operation receipt independently of 204; it does not prove effective assignment.

`TestAccMeetingPolicyReadOnly` reads Global and a filtered collection.
`TestAccMeetingPolicyCRUD` creates a uniquely named unassigned policy with a
non-default description and `AllowIPVideo: false`, checks persistence, updates
only its description, checks the omitted Boolean survives, lists, deletes and
verifies both typed GET404 and list absence. Both created/updated reads compare
all 123 getters with the pinned module. Cleanup uses the private recovery
journal. Certificate and client-secret lab runs passed on 2026-11-22.

The service's unit tests use these acceptance responses in `testdata`, including
all typed settings, raw snapshots, CRUD status codes and the actual deleted
resource error. Injected empty authorization/throttling/server errors separately
check request contracts and replay safety for application/delegated sessions.
They do not claim live delegated or assignment qualification.

See [qualification details](../../../docs/implementation/policies/meeting-policy-lab-acceptance.md).
Issue #130 remains open: grant/clear persistence, effective behavior, remaining
parameter scenarios, delegated authentication and other clouds still require
qualification. No tenant-wide defaults or existing user/group assignments were
changed by the CRUD acceptance tests.

The meeting model includes `KnowledgeGenerationMeetingTypes`,
`TranscriptionWithMeetingAI` and `MeetingAI`. Signed expiration/opacity values
use `int64`; `MediaBitRateKb` uses `uint32`. Only evidenced setter constraints
are enforced: a bound non-null bit rate must be at least 1, and the three GUID
settings must match the source pattern. Null remains distinct. No enum, length,
opacity range or cross-field restriction is invented.
