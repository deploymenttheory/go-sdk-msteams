# Teams Shifts app policy

`TeamsAdminAPI.TeamsShiftsAppPolicies` implements the public MicrosoftTeams 8.1.0
Get/Set Functions. CmdletConfig maps them to modern Get/Set-CsConfigurationModern.
Compiled classes with additional verbs are not public baseline exports for this
resource and do not determine its API family.

`GetByIdentity` reads one identity, `List` reads the bare array and stably orders
Global first, and `ListFiltered` applies the public wrapper's local Identity/Tag
wildcards. Blank filter selects all. Discovery and X-MS-Target-Uri use the shared
modern client; Filter is never invented as a server query.

`UpdateByIdentity` sends one sparse PUT with Identity and the explicitly bound
Boolean AllowTimeClockLocationDetection. `UpdateGlobal` applies the pinned omitted
Identity default. A nil setting pointer omits the field; false and true are bound
values. The public Boolean is non-nullable, so no JSON-null setter is invented.
Identity-only PUT is permitted by the public helper. There is no public Force,
New, Remove or Grant operation. Selected identity strings are escaped as one path
segment; the documented Global-only behavior remains service validated.

GET requires HTTP200; PUT requires HTTP204 and has no response model. Errors retain
HTTP metadata; writes never replay. Construction performs no discovery or auth.
The read model retains nullable Identity/Boolean plus complete owned RawFields and
RawJSON, including unknown metadata. Context cancellation is preserved.

Both certificate/client-secret modes passed Global read/list and a journaled
Boolean toggle/restoration on 2026-11-22. The original value is required non-null,
journaled before PUT, compared with PowerShell before/after/restoration and
independently read back after recovery. Its CI reads and writes share the tenant
write lock. Recovery kind `shifts-app-configuration` restores only this Boolean on
Global. Actual sanitized bodies and the empty204 receipt supply unit fixtures.
No create/delete semantics or effective location behavior is claimed.

The tested permission is the dedicated Teams Administrator application role.
Delegated modes, other clouds, additional identities, identity-only persistence
and effective location behavior need separate qualification.

Primary sources: pinned public proxy Functions, CmdletConfig, modern merged helpers
and generated Get/Set-CsConfiguration methods. Current reference:
[Microsoft Set-CsTeamsShiftsAppPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/set-csteamsshiftsapppolicy?view=teams-ps).
