# Teams mobility policy

Use `sdk.TeamsAdminAPI.TeamsMobilityPolicies` for the compiled public 8.1.0
Get/New/Set/Remove/Grant command surface.

| Method | Behavior |
|---|---|
| `GetByIdentity`, `List`, `ListFiltered` | Direct GET; bare list, exact Global first, local Identity/Tag wildcards |
| `CreateByIdentity` | One POST; omitted settings are not defaulted locally |
| `PatchByIdentity` | One sparse PATCH, requiring a bound setting |
| `UpdateByIdentity` | Explicit Set-metadata GET followed by sparse PATCH |
| `DeleteByIdentity` | DELETE with Identity and optional Force JSON body |
| `Assign`, `AssignUser`, `AssignGroup`, `AssignTenantDefault` | Compiled user/group/tenant grant routes |

Five nullable string getters and `policy.Field[string]` settings preserve
Description, IPAudioMobileMode, IPVideoMobileMode, MobileDialerPreference and
LinksInTeams. Omission, explicit null and an empty string remain distinct. The
pinned setters impose no enum/GUID/length validation, so documented examples are
not converted into local restrictions. Unknown response fields and owned complete
JSON remain available in RawFields/RawJSON; marshal returns the received snapshot.

The compiled transport uses MPACmdlet and command headers without modern service
discovery. A scoped configuration identity splits at its first slash; the
remaining name stays one escaped segment. Filter whitespace remains literal.
Bound Force uses the compiled string representation. UpdateResult distinguishes
preparation and HTTP acceptance from persisted state; mutations never replay.

Grant requires exactly one scope. The baseline's empty Identity set has no
operation and fails locally. User names are escaped whole route segments; nil and
empty PolicyName remain distinct. Group Rank is signed Int32 and explicit zero is
preserved. Tenant grant uses forceSwitchPresent. Grant HTTP acceptance alone does
not establish effective policy or propagation.

Acceptance uses an SDK-owned, unassigned policy with a non-default description and
explicit video setting. It verifies creation read-back, changes only the
description, checks the omitted setting survives, compares all five getters with
PowerShell twice, lists, deletes and independently checks GET404 and absence.
Recovery intent precedes create; the live responses provide permanent unit fixtures.
Teams Administrator is the tested lab role. Grants, effective client behavior,
delegated flows and other clouds remain separately unqualified.

See [qualification](../../../docs/implementation/policies/mobility-multi-tenant-lab-acceptance.md).
