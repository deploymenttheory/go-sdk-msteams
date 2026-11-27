# Media connectivity policy qualification

Owner: #125 (POL-52). Baseline: MicrosoftTeams 8.1.0.

## SDK behaviour

All five public commands are compiled cmdlets in
`Microsoft.Teams.Policy.Administration.Cmdlets.Core` (`GetTeamsMediaConnectivityPolicyCmdlet`
and siblings), deriving from the PolicyRP `BaseTpmGet/New/Set/Remove/GrantCmdlet<T>`
bases with `T = TeamsMediaConnectivityPolicy`. Reflection over the pinned
assemblies (Core SHA-256 `68dbcb1f…1e49`, PolicyRp `a5c8bc39…0e51`, the same
hashes already pinned by the Shifts and network/notification oracles) shows that
type has exactly two properties, `Identity` and `DirectConnection`, both
`System.String`. New and Set bind only `DirectConnection` (optional), plus
Identity and Force; the setters add no enum, so the SDK keeps the value opaque.

The checked runtime metadata (`runtime.json` commands 130, 248, 364, 474 and 603)
agrees: Get has Filter/Identity sets, Grant has GrantToGroup/GrantToTenant/
GrantToUser and the parameterless Identity set, New/Set/Remove have one
Identity set each. The configuration name in the PolicyRP route is the type name,
`/Skype.Policy/configurations/TeamsMediaConnectivityPolicy`, and the live tenant
returned the `TeamsMediaConnectivityPolicy.2023` schema for it.

The SDK reuses the shared compiled policy provider (`internal/policyrp`) and the
shared compiled grant routes used by every other PolicyRP policy. The only shared
SDK change is adding the type name to the provider's resource allowlist.

## Lab results

On 2026-11-23 both certificate and client-secret application modes passed:

- `TestAccMediaConnectivityPolicyReadOnly`: Global read, full list (Global first,
  identical raw fields) and the local `Global` filter, plus the public
  `DirectConnection` getter compared with pinned PowerShell.
- `TestAccMediaConnectivityPolicyCRUD`: created an owned, journaled
  `msteams-sdk-<run>-media-connectivity-policy` with `DirectConnection=Disabled`
  (Global is `Enabled`), read it back, updated only `DirectConnection` to
  `Enabled`, required every other returned field to stay unchanged, compared with
  PowerShell after create and after update, listed, deleted and independently
  required GET 404 and list absence. Both journal entries ended `cleaned`.

No Global settings, existing policies or assignments were changed. The tests do
not grant the policy, so there are no licensing implications.

Nine sanitized response files from the client-secret run supply the unit tests.
Request shapes for every grant scope, error statuses, no-replay mutations, nil and
cancelled contexts, wrong backend and invalid identities are unit-tested.

## Not yet qualified

- Grant persistence for user, group and tenant scopes (unit-tested only).
- Explicit-null and empty-string `DirectConnection` persistence.
- Effective media path behavior, delegated credentials and other clouds.

Current Microsoft reference:
[New-CsTeamsMediaConnectivityPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsmediaconnectivitypolicy?view=teams-ps).
