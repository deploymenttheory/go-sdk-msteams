# SIP devices configuration

Use `TeamsAdminAPI.TeamsSipDevicesConfigurations` for the public compiled
MicrosoftTeams 8.1.0 Get/Set commands and their three parameter sets.

`GetByIdentity`, `List` and `ListFiltered` use direct PolicyRP with stable
Global-first ordering and local Identity/Tag wildcard filtering. The bare
collection is not treated as an invented paged envelope. `PatchByIdentity` sends
one sparse PATCH; `UpdateByIdentity` explicitly composes Set's preparation GET and
PATCH and retains both preparation and HTTP acceptance on partial failure.

`BulkSignIn` is an opaque nullable string. `policy.Field[string]` distinguishes
omission, JSON null and a bound empty string. The pinned setter imposes no local
enum. An empty Set or Force alone is rejected before authentication. Optional
Force preserves the compiled string representation. The pinned runtime requires
an explicit Set Identity; callers can select Global. No New, Remove or Grant
operation is invented. Caller context and no-replay mutations use shared transport.

Read models own complete raw snapshots and project nullable Identity/BulkSignIn.
Unknown server fields remain available. Construction performs no I/O.

Both certificate/client-secret modes passed Global read/list and a journaled
Enabled/Disabled update/restoration with three pinned PowerShell comparisons.
Recovery `sip-devices-configuration` accepts only Global and the original bound
Enabled/Disabled lab snapshot, restores it and independently reads it back.
The tests do not register devices, sign in users or exercise effective SIP behavior.
Their Global reads and writes share the CI tenant lock. Five actual sanitized
responses supply unit tests. Null/empty persistence, other identities, delegated
modes and other clouds remain separately qualified scenarios.

Tested permission: dedicated Teams Administrator application role. Primary sources:
pinned Core TeamsSipDevicesConfiguration/Get/Set classes, BaseTpm/provider code and
runtime metadata. Current [Microsoft reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/get-csteamssipdevicesconfiguration?view=teams-ps)
may show a narrower syntax than the pinned runtime. See the repository
[qualification](../../../docs/implementation/policies/sip-template-lab-acceptance.md).
