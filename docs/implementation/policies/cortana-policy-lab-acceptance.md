# Teams Cortana policy lab acceptance

On 2026-11-21, TestLabCortanaPolicyReadOnly attempted a modern TeamsCortanaPolicy
Global read with the dedicated lab SPN using client-secret and certificate auth.
Both actual resource reads returned HTTP 403. No response payload, tenant/client
identifier or credential was retained in the repository. Filter/list and optional
pinned comparison were not reached. No policy, voice setting, assignment, role,
consent or privilege changed. Feature availability is not inferred from retained
baseline exports or this authorization failure.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; explicit approval remains pending. No workaround or privilege change was
applied. This unavailable evidence does not establish successful tenant CRUD.

The read-only harness gates credential access, bounds execution and optional
oracle output, binds the oracle to the canonical tenant, requires all five typed
settings and rejects missing/wrong/ambiguous Global and trailing output. The
optional oracle pins the module/private/wrapper/config/export hashes, selects
ForceModern, performs one Global read and disposes its certificate/session.
Canonical comparisons use known nullable projections. No tenant writes run.

Remaining qualification includes authorized reads/list/filter and pinned source
comparison; unique tagged create/read-back/update/delete and cleanup; source
null/empty and server mode constraints; runtime/flight/PassThru behavior; effective
voice invocation, ambient listening and in-context suggestions; user/tenant/group
grants, clear/rank semantics and cloud/authentication coverage. Issue #113 and
full command/tenant qualification stay open.
