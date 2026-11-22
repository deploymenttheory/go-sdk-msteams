# Teams personal attendant policies

Use `sdk.TeamsAdminAPI.TeamsPersonalAttendantPolicies` for the pinned compiled
Get/New/Set/Remove/Grant surface.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Direct GET, bare array, exact Global first, local Identity/Tag wildcards |
| `CreateByIdentity` | Native POST with only selected settings and optional Force |
| `PatchByIdentity` | One sparse PATCH; at least one bound setting required |
| `UpdateByIdentity` | Set preparation GET and sparse PATCH with separate receipts |
| `DeleteByIdentity` | DELETE with Identity/optional Force body |
| `Assign`, `AssignUser`, `AssignGroup`, `AssignTenantDefault` | Explicit compiled grant scope |

All eight source settings are nullable strings: PersonalAttendant, CallScreening,
CalendarBookings, InboundInternalCalls, InboundFederatedCalls, InboundPSTNCalls,
AutomaticTranscription and AutomaticRecording. policy.Field[string] distinguishes
omission, JSON null and bound empty strings. The source setters do not impose
local enums. Current service restrictions stay server-side and retain API errors.

The tested service returns an unprefixed owned identity, rather than Tag:Name.
Use the identity returned by reads; the SDK does not insert a Tag prefix. Unit
fixtures check that shape and its filter behavior explicitly. Scoped route names
still follow the shared compiled first-slash escaping contract. Filter whitespace
is literal. MPACmdlet/command headers are direct, without modern discovery.

Bound Force uses the compiled string form; mutations never replay. RawFields and
owned RawJSON retain complete responses and unknown properties. UpdateResult
separates preparation/HTTP acceptance from persisted and effective settings.
Grant requires exactly one scope, preserves nil/empty PolicyName and signed Int32
group rank, and treats backend acceptance separately from assignment propagation.

Certificate and client-secret lab CRUD use an unassigned SDK-owned policy:
AutomaticRecording=Disabled and AutomaticTranscription=EnabledUserOverride on
create, then a transcription-only update to Disabled. Read-back must preserve the
omitted recording setting; all eight getters match PowerShell twice. List/delete,
typed GET404 and collection absence independently verify cleanup. Intent is
journaled before create. The tested lab role is Teams Administrator.

[Microsoft's current reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamspersonalattendantpolicy?view=teams-ps)
describes a limited preview and temporary restrictions on the other six settings;
the acceptance scenario uses the two recording/transcription controls. No Global,
user/group assignment, license or role is changed. Full grant/effective behavior,
null/empty persistence, delegated modes and other clouds remain unqualified.
Permanent sanitized TestAcc responses feed the unit tests.

See [qualification](../../../docs/implementation/policies/personal-recording-lab-acceptance.md).
