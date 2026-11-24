# Work location detection policy qualification

Owner: #156 (POL-83). Research owner: POL-08. Baseline: MicrosoftTeams 8.1.0.

## Contract

All five public commands resolve as compiled Cmdlets in
`Microsoft.Teams.Policy.Administration.Cmdlets.Core`
(`GetTeamsWorkLocationDetectionPolicyCmdlet` and siblings), deriving from the
PolicyRP `BaseTpmGet/New/Set/Remove/GrantCmdlet<T>` bases with
`T = TeamsWorkLocationDetectionPolicy`. The pinned `custom/CmdletConfig.json` has
no entry for the family (the ConfigAPI private assembly only carries the
policy-type name as a string), so there is no modern mapping. Reflection over
the pinned assemblies (Core SHA-256 `68dbcb1f…1e49`, PolicyRp `a5c8bc39…0e51`, the
hashes already pinned by the Shifts and network/notification oracles) shows the model derives from
`BasePolicy` with exactly three public properties: `Identity` (`System.String`),
`EnableWorkLocationDetection` (`Nullable<Boolean>`) and `UserSettingsDefault`
(`System.String`). Their setters are trivial assignments (no enum or range
rule), and New/Set declare the two settings as optional parameters, plus Identity
and Force. The SDK therefore keeps `UserSettingsDefault` opaque; the pinned help
documents `Disabled` (default, Ask mode) and `Enabled` (Inform mode).

The checked runtime metadata (`runtime.json` commands 181, 269, 396, 505 and 642)
has nine public parameter sets: Get Filter/Identity; Grant GrantToGroup (required
`Group`, required signed `System.Int32` `Rank`, optional `PolicyName`),
GrantToTenant (required `Global`, optional `PolicyName` and `Force`), GrantToUser
(required `Identity`) and the parameterless default `Identity` set; and one
Identity set each for New, Set and Remove. The parameterless Grant set reaches
the shared `BaseTpmGrantCmdlet`, which reports an error, so the SDK rejects it
locally before HTTP. The configuration name in the PolicyRP route is the type
name, `/Skype.Policy/configurations/TeamsWorkLocationDetectionPolicy`, and the
live tenant returned the `TeamsWorkLocationDetectionPolicy.2024` schema for it.

The SDK sends through the shared compiled policy transport and reuses the shared
PolicyRP identity/body/decode helpers and the shared compiled grant routes used by
every other PolicyRP policy. The route binding to the one configuration name lives
in the package, so no shared SDK file changed.

## Lab results

On 2026-11-24 both certificate and client-secret application modes passed:

- `TestAccWorkLocationDetectionPolicyReadOnly`: Global read, full list (Global
  first, identical raw fields) and the local `Global` filter, plus both public
  getters compared with pinned PowerShell. Global is
  `EnableWorkLocationDetection=false`, `UserSettingsDefault=Disabled`.
- `TestAccWorkLocationDetectionPolicyCRUD`: created an owned, journaled, unassigned
  `msteams-sdk-<run>-work-location-detection-policy` with
  `EnableWorkLocationDetection=true` and `UserSettingsDefault=Enabled`, read it
  back, updated only `EnableWorkLocationDetection` to `false`, required
  `UserSettingsDefault` and every other returned field to stay unchanged, compared
  with PowerShell after create and after update, listed, deleted and independently
  required GET 404 and list absence. Every journal entry ended `cleaned`.

No Global settings, existing policies, assignments or user work-location settings
were changed. The tests do not grant the policy, so there are no licensing
implications.

Nine sanitized response files from the client-secret run supply the unit tests.
Request shapes for every grant scope, error statuses, no-replay mutations, nil and
cancelled contexts, wrong backend and invalid identities are contract-tested.

## Not yet qualified

- Grant persistence for user, group and tenant scopes (contract-tested only).
- Explicit-null, empty-string and other `UserSettingsDefault` values.
- Effective Wi-Fi/peripheral work-location detection behavior, delegated
  credentials and other clouds.

Current Microsoft reference:
[New-CsTeamsWorkLocationDetectionPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsworklocationdetectionpolicy?view=teams-ps).
