# Teams channels policy lab acceptance

On 2026-11-21, the opt-in TestAccChannelsPolicyReadOnly attempted a compiled
TeamsChannelsPolicy Global read with the dedicated SPN, using client-secret and
certificate authentication. Both actual resource reads returned HTTP 403.
No response payload, tenant/client identifier or credential was retained in the
repository. Filter/list and the optional pinned comparison were not reached.
No policy, channel setting, assignment, privilege, role or consent changed.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; no workaround or privilege change was applied. Explicit approval remains
pending. This is unavailable tenant evidence, not a successful CRUD result.

The harness requires read-only opt-in before credentials, bounds execution and
optional oracle output, binds the oracle to the canonical tenant, requires all
20 typed settings, and rejects missing/wrong/ambiguous Global and trailing output.
Canonical comparison follows typed getter values for case variants and later
duplicates. Legacy Boolean and separate string creation settings remain distinct.
The optional oracle verifies the module/Core/provider hashes, performs one Global
read and disposes the certificate and session. Tenant writes are not performed.

Remaining qualification includes authorized reads/list/filter and pinned source
comparison; unique tagged create/read-back/update/delete and cleanup; server
string values and interactions with legacy permissions; effective organization,
private/shared/threaded/default channel and external sharing behavior; user,
tenant and group grants, clear/rank semantics; cloud/flight availability and
required authentication modes. Issue #111 and full qualification stay open.
