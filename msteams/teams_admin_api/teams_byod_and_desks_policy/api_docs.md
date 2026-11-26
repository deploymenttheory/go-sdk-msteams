# Teams BYOD and desks policy

`TeamsAdminAPI.TeamsBYODAndDesksPolicies` implements the public compiled
MicrosoftTeams 8.1.0 Get/New/Set/Remove/Grant-CsTeamsBYODAndDesksPolicy
commands through PolicyRP (`/Skype.Policy/configurations/TeamsBYODAndDesksPolicy`).
Construction performs no I/O.

`GetByIdentity`, `List` and `ListFiltered` return the nullable typed
`DeviceDataCollection` string plus an owned complete raw snapshot, including server
fields such as Key, ConfigMetadata and ConfigId. Lists keep a stable Global-first
order; filters are local Identity/Tag wildcards. There is no evidenced pagination
contract.

The only setting is the nullable opaque string `DeviceDataCollection`. The pinned
New and Set cmdlets bind it as `System.String` with no enum; the lab Global value
is `Enabled`, an owned policy accepted `Disabled`, and tenant validation stays
authoritative. `policy.Field` distinguishes omission, JSON null and empty strings.

`CreateByIdentity` sends one POST; `PatchByIdentity` sends one sparse PATCH.
`UpdateByIdentity` explicitly prepares Set with a GET then PATCH, retaining the
preparation and HTTP acceptance independently of response decoding. An empty Set
or Force alone is rejected before authentication. Optional Force keeps the
compiled string encoding. DELETE includes its compiled Identity/control body.
Mutations never replay automatically; all methods preserve caller context.

The service sends through the shared compiled policy transport
(`internal/policytransport`) and reuses the shared PolicyRP identity, body and
decode helpers; its route is bound to the single `TeamsBYODAndDesksPolicy`
configuration name in this package.

Grant methods support user, group with signed Int32 rank, and tenant default,
including explicit clears, through the shared compiled policy-assignment routes
(`users/{id}/policies`, `groupPolicyAssignments/{id}/policyTypes`,
`tenants/policies`) and the shared `policy.Assignment` model. `Assign` requires
exactly one scope; the otherwise unbound compiled Identity parameter set is
rejected locally. Request shapes and errors are contract-tested; grant
persistence and effective user behavior remain unqualified.

Both app auth modes passed Global read/list/filter and owned CRUD in the
authorized lab on 2026-11-23. The test creates `DeviceDataCollection=Disabled`,
updates it to `Enabled`, requires every other returned field to stay unchanged,
compares the public getter with pinned PowerShell twice, lists, deletes and
independently verifies GET 404 and list absence. Sanitized responses in
`testdata` supply the unit tests. No assignments or existing policies change.
Recovery is bound to the private `byod-and-desks-policy` intent.

Permissions: the tested dedicated Teams Administrator application role, with
certificate or client-secret credentials. Delegated credentials, other clouds,
locale-specific filters, explicit-null persistence, grants and effective device
data collection behavior need separate lab qualification.

Primary sources: pinned TeamsBYODAndDesksPolicy and Get/New/Set/Remove/Grant
Core cmdlet classes, BaseTpm provider implementations and checked runtime
metadata. See [Microsoft New-CsTeamsBYODAndDesksPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsbyodanddeskspolicy?view=teams-ps)
and the repository qualification document.
