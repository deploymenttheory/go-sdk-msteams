# Tenant abuse and update management contracts

Owners: #148 (POL-75), #149 (POL-76). Baseline: MicrosoftTeams 8.1.0.
Selected tenant qualification is recorded below; full release qualification is not claimed.

The checked runtime and pinned Core classes confirm compiled Get/Set for tenant
abuse. Settings are seven nullable UInt64 thresholds, two nullable opaque strings
(TrialTenantValidation and SkipValidation), and a get-only BlockedTenants string
list. The constructor initializes the list. Set handles bound null/empty modifiers
as clear and applies other modifiers locally. Its helper's static GUID regex is
not executed by the inspected setter; no extra GUID validation should be invented.
Unsigned values must retain all 64 bits in typed models and JSON.

Update management has compiled Get/New/Set/Remove and a public Grant Function.
CmdletConfig maps only Grant to modern Grant-CsTeamsPolicy. Native CRUD and grant
therefore need distinct provider behavior. Public grant sets are Identity, group
with optional nullable Int32 rank, and tenant; no public Force parameter is exposed.
This must not be copied from the four-set compiled grant adapter.

Its ten settings include a get-only lazy-initialized DisabledInProductMessages
string list, three nullable Booleans, four nullable strings, nullable Int64
UpdateDayOfWeek and nullable DateTime UpdateTimeOfDay. The actual UpdateDayOfWeek
setter enforces 0 through 6, including during response deserialization; this is
source-backed validation. A dedicated DateTime type preserves the pinned provider's
kind and precision, as verified below. The source list helper's static regex is not called by its
setter. No additional enum, GUID or length validation is inferred.

Sources inspected: contracts/powershell/8.1.0/runtime.json; pinned CmdletConfig;
TeamsTenantAbuseConfiguration/Set, BlockedTenants,
TeamsUpdateManagementPolicy/DisabledInProductMessages and shared PSListModifier/
provider classes under the checksum-verified 8.1.0 module. Native models/adapters,
permanent sanitized response fixtures and unit/acceptance tests are implemented.
PRs merge only after every check passes for the tested commit.

Both auth modes now passed update-management Global reads/list and owned CRUD,
including sparse preservation, actual DateTime persistence, documented message
category Add/Remove and clear, five pinned comparisons and independent deletion.
The original lab identity receives HTTP403 / PolicyRP40301 for tenant-abuse in both
SDK and pinned PowerShell under both auth modes. Mandatory denial parity runs;
successful reads and UInt64 update/restoration remain exact-name configured skips.
The update test refuses mutation without a readable non-null original, journals
its full UInt64 value and independently verifies restoration when authorized.
No role, permission or existing tenant-abuse setting was changed.

DateTime source vectors were checked offline against the loaded pinned serializer.
Unspecified clocks remain timezone-free, UTC retains Z and offset values become
host-local. The native type preserves CLR 100ns precision and rejects finer Go
input explicitly. The live DateTime write changed the owned policy's clock and
matched the pinned module. Getter comparisons now use exact JSON numbers, with
regressions distinguishing neighboring integers beyond 2^53 and at UInt64 maximum.
