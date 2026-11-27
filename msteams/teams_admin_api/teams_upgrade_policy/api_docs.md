# Upgrade policy

`TeamsAdminAPI.TeamsUpgradePolicies` implements the five public pinned 8.1.0 Get
and Grant parameter sets through modern ConfigAPI discovery. GetByIdentity, List
and ListFiltered preserve full owned snapshots and project Description, Mode,
NotifySfbUsers and Action. Unknown modern fields and shapes remain available raw.
Whitespace filters select the entire collection; Identity/Tag wildcard filtering
and stable Global-first ordering match the mapped helper.

Assign selects exactly one user, tenant or group. AssignToUser,
AssignTenantDefault and AssignToGroup expose the native operations. PolicyName
is a required string argument despite optional PowerShell metadata; empty clears.
Tenant Force is explicit. Group rank is optional signed Int32; Assign omits zero
like the wrapper, while AssignToGroup can explicitly send zero. Accepted group
operations preserve 202 separately from 204; submission does not prove propagation.
Writes never replay and share caller context and identity-bound transport.

A nonnil Assignment.MigrateMeetingsToTeams returns ErrMigrationParameterBinding
before HTTP for either true or false. The pinned public proxy exposes this field
but forwards it to a modern function that cannot bind it. Authenticated pinned
WhatIf comparisons reproduce this defect. Omit it to use native grants. No
successful meeting migration is claimed. PassThru concerns PowerShell output;
hidden dispatch overrides are not SDK authentication or resource settings.

Both app-auth modes passed actual Global Get/List and pinned getter comparisons.
Permanent sanitized service responses feed unit models, reads and filters. Grant
wire tests cover discovery, headers, escaping, Force, signed rank and failure
receipts; grant persistence remains unqualified. There are no public New/Set/Remove
exports for this policy in the baseline. The lab uses its existing Teams
Administrator application; delegated flows and other clouds remain unqualified.
See the [testing guide](../../../docs/guides/testing.md) for source evidence.
