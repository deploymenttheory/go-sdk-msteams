# Tenant migration configuration qualification

Owner: #160 (POL-87). Baseline: MicrosoftTeams 8.1.0.

The pinned public runtime metadata (`runtime.json` commands 189 and 646) has
Get-CsTenantMigrationConfiguration Identity/Filter and
Set-CsTenantMigrationConfiguration Identity. Set binds one setting with no
validation attributes, the non-nullable `System.Boolean` MeetingMigrationEnabled;
Identity is an optional positional string. The help's Instance, Tenant and Force
parameters are not present in the public runtime. Both commands are public proxy
`Function`s whose `custom/CmdletConfig.json` entry (`reconciliation.json` records
213 and 708) is `CmdletType: Remoting`,
`AutoRestModuleName: Microsoft.Teams.ConfigAPI.Cmdlets`,
`DefaultAutoRestParameters.ConfigType: TenantMigrationConfiguration` and
`ModernCmdlet: Get-CsConfigurationModern` / `Set-CsConfigurationModern`. This is
the same dispatch already qualified for TeamsMigrationConfiguration and the guest
configurations, so the generated ConfigAPI routes apply:

- GET `/Skype.Policy/configurations/TenantMigrationConfiguration` (bare list)
- GET `/Skype.Policy/configurations/TenantMigrationConfiguration/configuration/{Identity}`
- PUT the same instance route with a sparse body that always carries `Identity`;
  HTTP 204 with an empty body is the success receipt.

This family is distinct from `teams_migration_configuration`
(Get/Set-CsTeamsMigrationConfiguration, POL-62): that ConfigType is
`TeamsMigrationConfiguration` with the `EnableLegacyClientInterop` setting,
while this one is `TenantMigrationConfiguration` (XML schema
`Settings.TenantMigration.2016`) with MeetingMigrationEnabled, the Meeting
Migration Service (MMS) switch. The packages share no code or recovery kind.

The modern wrapper defaults Set Identity to Global and exposes no Force; Filter
is applied locally to Identity and `Tag:`Identity with wildcard matching. The
Go service does not use the compiled PolicyRP PATCH provider and no policyrp
registration is involved. Oracle checks pin these module assembly SHA256 values:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`
- ConfigAPI: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

On 2026-11-23 the dedicated SDK service principal (Teams Administrator) read and
listed Global with both client-secret and certificate authentication, and the
typed getter matched pinned PowerShell 8.1.0. The lab's original Global had Host
authority and MeetingMigrationEnabled true.

Write safety: per the 8.1.0 help, setting MeetingMigrationEnabled to false
disables MMS, which otherwise updates users' meetings in the background when
their user-level settings change. Disabling it cannot start a migration, and
re-enabling it restores the default; the toggle itself is not an MMS trigger and
has no licensing effect. The update test therefore only runs from an enabled
original and refuses otherwise, and the recovery handler only ever restores an
enabled journaled value (it rejects a false snapshot), so the lab never moves
MMS from disabled to enabled on its own.

The update test persists the original in a bound private recovery journal,
disables MMS through one sparse `UpdateGlobal`, waits for persisted read-back and
compares the pinned getter. It then restores the original through the journal
handler, independently reads it back and compares again. Both application auth
modes passed; each run left exactly one `tenant-migration-configuration` journal
entry in the cleaned state. After restoration MeetingMigrationEnabled is true
again, but the authority is Tenant rather than Host, as for every other
singleton written by the service. No licenses, roles, users, groups or policy
assignments were changed.

CI runs these as the `tenant_migration_configuration` partition: a serialized
`global-reads` step and an `update-restore` step under the tenant lock, requiring
three pinned comparisons, the update and restoration assertions and one cleaned
restore entry. The partition uses its own read-only oracle script,
`read-tenant-migration-configuration.ps1`, so it does not touch the shared
configuration oracle or lab kit.

Five permanent sanitized response files come directly from these acceptance
reads, the update and the read-backs; the real 204 body is empty. Unit tests use
them for native routing, headers, typed getters, snapshot ownership, filter
selection, sparse PUT bodies (unbound, false and true) and identity segment
escaping. Common denial propagation reuses the actual privacy-service 403
envelope and is not a live denial claim for this service. Deliberately malformed
replies test decoding failures separately from successful fixtures.

Still unqualified: enabling MMS from a disabled original (unit-tested only;
not a safe lab write), non-Global identities (the service documents Global only),
effective meeting migration behavior, delegated authentication and other clouds.
No create/delete methods are invented; none are exported by the pinned module.
