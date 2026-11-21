# Teams files policy lab acceptance

On 2026-11-21, TestLabFilesPolicyReadOnly attempted a compiled TeamsFilesPolicy
Global read with the dedicated lab SPN using client-secret and certificate
authentication. Both actual resource reads returned HTTP 403. The harness checks
the response request path to distinguish resource denial from preparation or
discovery failure. No response payload, tenant/client ID or credential was
retained. List/filter and optional pinned comparison were not reached. No policy,
assignment, role, consent or privilege changed.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; explicit approval remains pending. No workaround or privilege change was
applied. Denied reads do not establish successful tenant CRUD behavior.

The harness gates credential access, bounds execution and optional oracle output,
binds comparisons to the canonical tenant and requires all four typed nullable
strings. It rejects incomplete/wrong types, wrong tenant, trailing output and
missing/ambiguous Global selections. Canonical getter projections preserve
case-insensitive last values and the exact source JSON names despite normalized
Go names. The optional certificate oracle pins the 8.1.0 module/Core/PolicyRp
baseline, reads Global once and disposes its certificate/session. No writes run.

Remaining qualification includes authorized compiled read/list/filter and pinned
comparison; disposable create/update/read-back/delete and cleanup; grants to safe
targets with read-back/restoration; null/empty/Force/Rank and server constraints;
cloud/authentication coverage; and effective native file entry points, channel
files tabs, upload application and external chat file-sharing behavior. HTTP
acceptance or a group 202 result does not prove effective assignment. Issue #120
remains open for full command/tenant acceptance.
