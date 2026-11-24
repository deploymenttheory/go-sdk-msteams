# Teams AI policy

`TeamsAdminAPI.TeamsAIPolicies` implements the public compiled MicrosoftTeams
8.1.0 Get/New/Set/Remove/Grant-CsTeamsAIPolicy commands through PolicyRP
(`/Skype.Policy/configurations/TeamsAIPolicy`). Construction performs no I/O.

`GetByIdentity`, `List` and `ListFiltered` return the five nullable typed settings
plus an owned complete raw snapshot, including server fields such as Key,
ConfigMetadata and ConfigId. Lists keep a stable Global-first order; filters are
local Identity/Tag wildcards. There is no evidenced pagination contract.

The settings are `Description`, `EnrollFace`, `EnrollVoice`,
`SpeakerAttributionForBYOD` and `PassiveVoiceEnrollment`. The pinned New and Set
cmdlets bind each as an optional `System.String` with no enum, and the pinned
model setters run no enum or range rule. The `Description` setter evaluates
1/1040 length constants but discards the result and never throws, so the SDK
does not invent length validation; tenant validation stays authoritative. The
lab Global returns `Description` null and the four enrollment/attribution values
`Enabled`; an owned policy accepted `EnrollFace=Disabled`. `policy.Field`
distinguishes omission, JSON null and empty strings.

`CreateByIdentity` sends one POST; `PatchByIdentity` sends one sparse PATCH.
`UpdateByIdentity` explicitly prepares Set with a GET then PATCH, retaining the
preparation and HTTP acceptance independently of response decoding. An empty Set
or Force alone is rejected before authentication. Optional Force keeps the
compiled string encoding. DELETE includes its compiled Identity/control body.
Mutations never replay automatically; all methods preserve caller context.

The service sends through the shared compiled policy transport
(`internal/policytransport`) and reuses the shared PolicyRP identity, body and
decode helpers; its route is bound to the single `TeamsAIPolicy` configuration
name in this package.

Grant methods support user, group with signed Int32 rank, and tenant default
with Force, including explicit clears, through the shared compiled
policy-assignment routes (`users/{id}/policies`,
`groupPolicyAssignments/{id}/policyTypes`, `tenants/policies`) and the shared
`policy.Assignment` model. `Assign` requires exactly one scope; the otherwise
unbound parameterless compiled Identity parameter set (which the shared
BaseTpmGrantCmdlet reports as an error) is rejected locally. Request shapes and
errors are contract-tested; grant persistence and effective user behavior remain
unqualified.

Both app auth modes passed Global read/list/filter and owned CRUD in the
authorized lab on 2026-11-24. The test creates an owned policy with a
`Description` and `EnrollFace=Disabled`, then in one sparse update sets
`EnrollFace=Enabled` and clears `Description` with explicit JSON null, requires
every other returned field to stay unchanged, compares all five public getters
with pinned PowerShell twice, lists, deletes and independently verifies GET 404
and list absence. Sanitized responses in `testdata` supply the unit tests. No
assignments, existing policies or user enrollment settings change. Recovery is
bound to the private `ai-policy` intent.

Permissions: the tested dedicated Teams Administrator application role, with
certificate or client-secret credentials. Delegated credentials, other clouds,
locale-specific filters, other enrollment values, grants and effective
face/voice enrollment or speaker attribution behavior need separate lab
qualification.

Primary sources: pinned TeamsAIPolicy and Get/New/Set/Remove/Grant Core cmdlet
classes, BaseTpm provider implementations and checked runtime metadata. See
[Microsoft New-CsTeamsAIPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsaipolicy?view=teams-ps)
and the repository qualification document.
