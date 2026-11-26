# Teams AI policy qualification

Owner: #157 (POL-84). Research owner: POL-08. Baseline: MicrosoftTeams 8.1.0.

## Contract

All five public commands resolve as compiled cmdlets in
`Microsoft.Teams.Policy.Administration.Cmdlets.Core` (`GetTeamsAIPolicyCmdlet` and
siblings), deriving from the PolicyRP `BaseTpmGet/New/Set/Remove/GrantCmdlet<T>`
bases with `T = TeamsAIPolicy`. CmdletConfig has no modern mapping for them.
Reflection over the pinned assemblies (Core SHA-256 `68dbcb1f…1e49`, PolicyRp
`a5c8bc39…0e51`, the hashes already pinned by the other PolicyRP oracles) shows
`TeamsAIPolicy : BasePolicy` with six `System.String` properties: `Identity`,
`Description`, `EnrollFace`, `EnrollVoice`, `SpeakerAttributionForBYOD` and
`PassiveVoiceEnrollment`. New and Set bind the five settings as optional
parameters with no validation attributes, plus Identity and Force.

The model setters for the four enrollment/attribution settings only store the
value (8-byte IL bodies). The `Description` setter loads the string length and
the constants 1 and 1040, then pops them without comparing or throwing, so the
length constants do not prove validation. The SDK keeps all five values opaque
and nullable and adds no local enum or length rule
(`TestUnitDescriptionLengthIsNotLocallyValidated`).

The checked runtime metadata (`runtime.json` commands 100, 229, 338, 453 and 573)
agrees: Get has Filter/Identity sets, Grant has GrantToGroup (required `Group`
and signed Int32 `Rank`), GrantToTenant (required `Global`, optional `Force`),
GrantToUser and the parameterless default Identity set, and New/Set/Remove have
one Identity set each, nine sets in all. The parameterless Identity Grant set has
no matching operation in the shared BaseTpmGrantCmdlet, which reports an error;
the SDK rejects a missing or mixed scope before HTTP.

The configuration name in the PolicyRP route is the type name,
`/Skype.Policy/configurations/TeamsAIPolicy`, and the live tenant returned the
`TeamsAIPolicy.2024` schema. The SDK sends through the shared compiled policy
transport and reuses the shared PolicyRP identity/body/decode helpers and the
shared compiled grant routes. The route binding to the one configuration name
lives in the package, so no shared SDK file changed.

## Lab results

On 2026-11-24 both certificate and client-secret application modes passed:

- `TestAccAIPolicyReadOnly`: Global read, full list (Global first, identical raw
  fields) and the local `Global` filter, plus all five public getters compared
  with pinned PowerShell. The lab Global has `Description` null and the other
  four settings `Enabled`.
- `TestAccAIPolicyCRUD`: created an owned, unassigned, journaled
  `msteams-sdk-<run>-ai-policy` with a description and `EnrollFace=Disabled`,
  read it back, then in one sparse update set `EnrollFace=Enabled` and cleared
  `Description` with explicit JSON null (read back as null), required every other
  returned field unchanged, compared with PowerShell after create and after
  update, listed, deleted and independently required GET 404 and list absence.
  Every journal entry ended `cleaned`.

No Global settings, existing policies, assignments or user enrollment settings
were changed. The tests do not grant the policy, so there are no licensing
implications.

Nine sanitized response files from the client-secret run supply the unit tests.
Request shapes for every grant scope, error statuses, no-replay mutations, nil and
cancelled contexts, wrong backend and invalid identities are contract-tested.

## Not yet qualified

- Grant persistence for user, group and tenant scopes (contract-tested only).
- `EnrollVoice`, `SpeakerAttributionForBYOD` and `PassiveVoiceEnrollment` writes,
  values other than `Enabled`/`Disabled`, and empty or over-length descriptions.
- Effective face/voice enrollment and speaker attribution behavior, delegated
  credentials and other clouds.

Current Microsoft reference:
[New-CsTeamsAIPolicy](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsaipolicy?view=teams-ps).
