# Teams work location detection policy

`TeamsAdminAPI.TeamsWorkLocationDetectionPolicies` implements the public compiled
MicrosoftTeams 8.1.0 Get/New/Set/Remove/Grant-CsTeamsWorkLocationDetectionPolicy
commands through PolicyRP
(`/Skype.Policy/configurations/TeamsWorkLocationDetectionPolicy`). Construction
performs no I/O.

`GetByIdentity`, `List` and `ListFiltered` return the nullable typed
`EnableWorkLocationDetection` Boolean and `UserSettingsDefault` string plus an
owned complete raw snapshot, including server fields such as Key, ConfigMetadata
and ConfigId. Lists keep a stable Global-first order; filters are local
Identity/Tag wildcards. There is no evidenced pagination contract.

The pinned New and Set cmdlets bind `EnableWorkLocationDetection` as
`Nullable<Boolean>` and `UserSettingsDefault` as `System.String`. The pinned model
setters run no enum or range rule, so `UserSettingsDefault` stays opaque; Microsoft
documents `Disabled` (default, Ask mode) and `Enabled` (Inform mode), and tenant
validation stays authoritative. `policy.Field` distinguishes omission, JSON null,
`false` and empty strings.

`CreateByIdentity` sends one POST; `PatchByIdentity` sends one sparse PATCH.
`UpdateByIdentity` explicitly prepares Set with a GET then PATCH, retaining the
preparation and HTTP acceptance independently of response decoding. An empty Set
or Force alone is rejected before authentication. Optional Force keeps the
compiled string encoding. DELETE includes its compiled Identity/control body.
Mutations never replay automatically; all methods preserve caller context.

The service sends through the shared compiled policy transport
(`internal/policytransport`) and reuses the shared PolicyRP identity, body and
decode helpers; its route is bound to the single
`TeamsWorkLocationDetectionPolicy` configuration name in this package.

Grant methods support user, group with signed Int32 rank, and tenant default with
optional Force, including explicit clears, through the shared compiled
policy-assignment routes (`users/{id}/policies`,
`groupPolicyAssignments/{id}/policyTypes`, `tenants/policies`) and the shared
`policy.Assignment` model. `Assign` requires exactly one scope; the parameterless
compiled Identity parameter set has no source operation (the shared
`BaseTpmGrantCmdlet` reports an error for it) and is rejected locally. Request
shapes and errors are contract-tested; grant persistence and effective user
behavior remain unqualified.

Both app auth modes passed Global read/list/filter and owned CRUD in the
authorized lab on 2026-11-24. The test creates `EnableWorkLocationDetection=true`
with `UserSettingsDefault=Enabled` (Global is `false`/`Disabled`), updates only
`EnableWorkLocationDetection` to `false`, requires `UserSettingsDefault` and every
other returned field to stay unchanged, compares both public getters with pinned
PowerShell twice, lists, deletes and independently verifies GET 404 and list
absence. Sanitized responses in `testdata` supply the unit tests. No assignments,
existing policies or user work-location settings change. Recovery is bound to the
private `work-location-detection-policy` intent.

Permissions: the tested dedicated Teams Administrator application role, with
certificate or client-secret credentials. Delegated credentials, other clouds,
locale-specific filters, explicit-null persistence, grants and effective
work-location detection behavior need separate lab qualification.

Primary sources: pinned TeamsWorkLocationDetectionPolicy and
Get/New/Set/Remove/Grant Core cmdlet classes, BaseTpm provider implementations
and checked runtime metadata. See
[Microsoft New-CsTeamsWorkLocationDetectionPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsworklocationdetectionpolicy?view=teams-ps)
and the repository qualification document.
