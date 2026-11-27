# BYOD and desks policy qualification

Owner: #158 (POL-85). Baseline: MicrosoftTeams 8.1.0.

## SDK behaviour

All five public commands are compiled cmdlets in
`Microsoft.Teams.Policy.Administration.Cmdlets.Core` (`GetTeamsBYODAndDesksPolicyCmdlet`
and siblings), deriving from the PolicyRP `BaseTpmGet/New/Set/Remove/GrantCmdlet<T>`
bases with `T = TeamsBYODAndDesksPolicy`. Reflection over the pinned assemblies
(Core SHA-256 `68dbcb1f…1e49`, PolicyRp `a5c8bc39…0e51`, the same hashes already
pinned by the Shifts and network/notification oracles) shows that type has exactly
two properties, `Identity` and `DeviceDataCollection`, both `System.String`. New
and Set bind only `DeviceDataCollection` (optional), plus Identity and Force; the
setters add no enum, so the SDK keeps the value opaque.

The checked runtime metadata (`runtime.json` commands 105, 233, 342, 457 and 578)
agrees: Get has Filter/Identity sets, Grant has GrantToGroup/GrantToTenant/
GrantToUser and the parameterless Identity set, New/Set/Remove have one
Identity set each. The configuration name in the PolicyRP route is the type name,
`/Skype.Policy/configurations/TeamsBYODAndDesksPolicy`, and the live tenant
returned the `TeamsBYODAndDesksPolicy.2024` schema for it.

The SDK sends through the shared compiled policy transport and reuses the shared
PolicyRP identity/body/decode helpers and the shared compiled grant routes used by
every other PolicyRP policy. The route binding to the one configuration name lives
in the package, so no shared SDK file changed (adding the name to the shared
provider allowlist would have re-run every acceptance partition).

## Lab results

On 2026-11-23 both certificate and client-secret application modes passed:

- `TestAccBYODAndDesksPolicyReadOnly`: Global read, full list (Global first,
  identical raw fields) and the local `Global` filter, plus the public
  `DeviceDataCollection` getter compared with pinned PowerShell.
- `TestAccBYODAndDesksPolicyCRUD`: created an owned, journaled
  `msteams-sdk-<run>-byod-and-desks-policy` with `DeviceDataCollection=Disabled`
  (Global is `Enabled`), read it back, updated only `DeviceDataCollection` to
  `Enabled`, required every other returned field to stay unchanged, compared with
  PowerShell after create and after update, listed, deleted and independently
  required GET 404 and list absence. Every journal entry ended `cleaned`.

No Global settings, existing policies or assignments were changed. The tests do
not grant the policy, so there are no licensing implications.

Nine sanitized response files from the client-secret run supply the unit tests.
Request shapes for every grant scope, error statuses, no-replay mutations, nil and
cancelled contexts, wrong backend and invalid identities are unit-tested.

## Not yet qualified

- Grant persistence for user, group and tenant scopes (unit-tested only).
- Explicit-null, empty-string and other `DeviceDataCollection` values.
- Effective device/peripheral data collection behavior, delegated credentials and
  other clouds.

Current Microsoft reference:
[New-CsTeamsBYODAndDesksPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsbyodanddeskspolicy?view=teams-ps).
