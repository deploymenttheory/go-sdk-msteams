# Guest calling configuration qualification

Owner: #122 (POL-49). Baseline: MicrosoftTeams 8.1.0.

The pinned public runtime metadata (`runtime.json` commands 126 and 599) has
Get-CsTeamsGuestCallingConfiguration Identity/Filter and
Set-CsTeamsGuestCallingConfiguration Identity. Set binds one setting,
AllowPrivateCalling, as a non-nullable `System.Boolean`; Identity is an optional
positional string. Both commands are public proxy `Function`s whose
`custom/CmdletConfig.json` entry (`reconciliation.json` records 149 and 660) is
`CmdletType: Remoting`, `AutoRestModuleName: Microsoft.Teams.ConfigAPI.Cmdlets`,
`DefaultAutoRestParameters.ConfigType: TeamsGuestCallingConfiguration` and
`ModernCmdlet: Get-CsConfigurationModern` / `Set-CsConfigurationModern`. This is
a session-flight-dependent dispatch with an explicit modern candidate. The
observed native/module acceptance below covers these modern routes; mapping
alone does not qualify the legacy server contract:

- GET `/Skype.Policy/configurations/TeamsGuestCallingConfiguration` (bare list)
- GET `/Skype.Policy/configurations/TeamsGuestCallingConfiguration/configuration/{Identity}`
- PUT the same instance route with a sparse body that always carries `Identity`;
  HTTP 204 with an empty body is the success receipt.

The modern wrapper defaults Set Identity to Global and exposes no Force; Filter
is applied locally to Identity and `Tag:`Identity with wildcard matching. The
Go service does not use the compiled PolicyRP PATCH provider and no
policyrp registration is involved. Oracle checks pin these module assembly
SHA256 values:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`
- ConfigAPI: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

On 2026-11-23 the dedicated SDK service principal (Teams Administrator) read and
listed Global with both client-secret and certificate authentication, and the
typed AllowPrivateCalling getter matched pinned PowerShell 8.1.0. The lab's
original Global had Host authority and AllowPrivateCalling true.

The update test refuses a missing/null original Boolean, persists the original in
a bound private recovery journal, toggles AllowPrivateCalling through
`UpdateGlobal`, waits for persisted read-back and compares the pinned getter,
restores the original value through the journal handler, then independently
reads it back and compares again. Both application auth modes passed; each run
left exactly one `guest-calling-configuration` journal entry in the cleaned
state. After restoration the Global value equals the original, but its authority
is Tenant rather than Host, as for every other singleton written by the service.
No licenses, roles, users, groups or policy assignments were changed.

CI runs these as the `teams_guest_calling_configuration` partition: a serialized
`global-reads` step and an `update-restore` step under the tenant lock, requiring
three pinned comparisons, the update and restoration assertions and one cleaned
restore entry. The partition uses its own read-only oracle script,
`read-guest-calling-configuration.ps1`, so it does not touch the shared
configuration oracle or lab kit.

Five permanent sanitized response files come directly from these acceptance
reads, the update and the read-backs; the real 204 body is empty. Unit tests use
them for native routing, headers, typed getters, snapshot ownership, filter
selection, sparse PUT bodies for unbound/false/true and the identity segment
escaping. Common denial propagation reuses the actual privacy-service 403
envelope and is not a live denial claim for this service. Deliberately malformed
replies test decoding failures separately from successful fixtures.

Still unqualified: non-Global identities (the service documents Global only),
effective guest private-calling behavior and propagation, delegated
authentication and other clouds. No create/delete methods are invented; none are
exported by the pinned module.
