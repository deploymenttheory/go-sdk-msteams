# Upgrade configuration and policy contracts

Owners: #150 (POL-77), #151 (POL-78). Baseline: MicrosoftTeams 8.1.0.

Upgrade configuration exports compiled Get (Identity/Filter) and Set (Identity).
The inspected Core model and setter contain nullable DownloadTeams and
BlockLegacyAuthorization booleans and nullable opaque SfBMeetingJoinUx. Bound
fields alone are copied; no enum or cross-field validation is executed. Set uses
the shared compiled GET/sparse PATCH protocol and Force string encoding. There is
no public create/delete/grant for this configuration.

Upgrade policy exports modern Get (Identity/Filter) and Grant (Identity,
GrantToTenant, GrantToGroup), as verified in the public proxy and runtime metadata.
CmdletConfig maps Get to Get-CsConfigurationModern and Grant to Grant-CsTeamsPolicy.
The unused compiled New/Set/Remove classes are not public exports and do not add
baseline CRUD methods. The native Get/List use discovery, Accept and
X-MS-Target-Uri, retaining complete modern JSON snapshots and typed Description,
Mode, NotifySfbUsers and Action. Local filtering and stable Global-first ordering
follow the modern helper. Grant preserves tenant Force, optional signed Int32
rank, wrapper omission of zero rank and explicit empty policy clear. PolicyName
must be bound despite optional metadata. PassThru is PowerShell output behavior.

The public MigrateMeetingsToTeams boolean has a pinned modern dispatch defect.
GrantCsTeamsUpgradePolicy declares the argument, but mapped Grant-CsTeamsPolicy
has neither this argument nor a PropertyBag parameter. TrpsPSCmdlet only repacks
unknown arguments when the destination has PropertyBag; CustomCmdletUtils only
prepares telemetry/HTTP pipeline steps. Thus explicitly bound false and true both
fail NamedParameterNotFound before the native grant. Pinned authenticated WhatIf
probes confirmed both failures. The permanent TestAcc oracle repeats those two
binding checks with WhatIf and compares the SDK's ErrMigrationParameterBinding;
no successful meeting migration is claimed or invented. Omitting this field
leaves the native user, tenant or group grant available. This behavior is specific
to the inspected modern dispatch and must be reviewed if upstream changes it.

Both certificate and client-secret modes passed native Global Get/List and pinned
getter comparisons for both services. Upgrade configuration also passed a
journaled DownloadTeams toggle, full getter/omission checks, three pinned
comparisons and independent restoration of the original. Permanent actual
sanitized service responses feed unit read, model and update replay. Grant
persistence/effective behavior, delegated tenant flows and other clouds remain
unqualified; no tenant-wide upgrade policy or user meeting migration was tested.

Pinned sources inspected: runtime.json; public proxy functions; CmdletConfig;
Merged_custom_PsExt.ps1 Get-CsConfigurationModern/Grant-CsTeamsPolicy;
TeamsUpgradeConfiguration and SetTeamsUpgradeConfigurationCmdlet; TrpsPSCmdlet,
FlightingUtils, GrantCsTeamsUpgradePolicy and CustomCmdletUtils from the
checksum-verified assemblies. Shared-provider behavior follows the existing
compiled/modern contract adapters. No new roles, permissions or secrets were needed.
