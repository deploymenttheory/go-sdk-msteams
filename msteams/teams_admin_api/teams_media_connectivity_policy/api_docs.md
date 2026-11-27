# Teams media connectivity policy

`TeamsAdminAPI.TeamsMediaConnectivityPolicies` implements the public compiled
MicrosoftTeams 8.1.0 Get/New/Set/Remove/Grant-CsTeamsMediaConnectivityPolicy
commands through PolicyRP (`/Skype.Policy/configurations/TeamsMediaConnectivityPolicy`).
Construction performs no I/O.

`GetByIdentity`, `List` and `ListFiltered` return the nullable typed
`DirectConnection` string plus an owned complete raw snapshot, including server
fields such as Key, ConfigMetadata and ConfigId. Lists keep a stable Global-first
order; filters are local Identity/Tag wildcards. There is no evidenced pagination
contract.

The only setting is the nullable opaque string `DirectConnection`. The pinned New
and Set cmdlets bind it as `System.String` with no enum; current Microsoft
documentation describes `Enabled` and `Disabled`, and tenant validation stays
authoritative. `policy.Field` distinguishes omission, JSON null and empty strings.

`CreateByIdentity` sends one POST; `PatchByIdentity` sends one sparse PATCH.
`UpdateByIdentity` explicitly prepares Set with a GET then PATCH, retaining the
preparation and HTTP acceptance independently of response decoding. An empty Set
or Force alone is rejected before authentication. Optional Force keeps the
compiled string encoding. DELETE includes its compiled Identity/control body.
Mutations never replay automatically; all methods preserve caller context.

Grant methods support user, group with signed Int32 rank, and tenant default,
including explicit clears, through the shared compiled policy-assignment routes
(`users/{id}/policies`, `groupPolicyAssignments/{id}/policyTypes`,
`tenants/policies`). `Assign` requires exactly one scope; the otherwise unbound
compiled Identity parameter set is rejected locally. Request shapes and errors
are contract-tested; grant persistence and effective user behavior remain
unqualified.

Both app auth modes passed Global read/list/filter and owned CRUD in the
authorized lab on 2026-11-23. The test creates `DirectConnection=Disabled`,
updates it to `Enabled`, requires every other returned field to stay unchanged,
compares the public getter with pinned PowerShell twice, lists, deletes and
independently verifies GET 404 and list absence. Sanitized responses in
`testdata` supply the unit tests. No assignments or existing policies change.
Recovery is bound to the private `media-connectivity-policy` intent.

Permissions: the tested dedicated Teams Administrator application role, with
certificate or client-secret credentials. Delegated credentials, other clouds,
locale-specific filters, explicit-null persistence, grants and effective media
path behavior need separate lab qualification.

Primary sources: pinned TeamsMediaConnectivityPolicy and Get/New/Set/Remove/Grant
Core cmdlet classes, BaseTpm provider implementations and checked runtime
metadata. See [Microsoft New-CsTeamsMediaConnectivityPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsmediaconnectivitypolicy?view=teams-ps)
and the repository qualification document.
