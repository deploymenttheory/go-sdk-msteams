# Teams client configuration lab acceptance

On 2026-11-21, TestAccClientConfigurationReadOnly attempted a compiled
TeamsClientConfiguration Global read with the dedicated SPN in client-secret and
certificate modes. Both actual resource reads returned HTTP 403. No response
payload, tenant/client identifier or credential was retained in this repository.
Filter/list and the optional pinned comparison were not reached. No configuration,
client setting, role, consent or privilege changed.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; explicit approval remains pending. No workaround or privilege change was
applied. This is unavailable evidence, not a successful read/update validation.

The read-only harness gates credentials, bounds execution and optional oracle
output, binds the oracle to the canonical tenant, requires all 17 typed settings,
and rejects missing/wrong/ambiguous Global and trailing output. Canonical
comparison follows typed getter values for case variants and later duplicates.
The optional oracle verifies module/Core/provider hashes, performs one Global
read and disposes the certificate and session. It performs no tenant updates.

Remaining qualification includes authorized reads/list/filter and pinned source
comparison; non-default update/read-back and restoration; null/empty and Force
semantics; server/cloud/flight availability and effective email, external file
providers, guest, resource account, people-search and unified-domain behavior.
No create/delete/grant operation is evidenced for this configuration. Issue #112
and full command/tenant qualification remain open.
