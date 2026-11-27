# Teams events policy lab acceptance

On 2026-11-21, TestAccEventsPolicyReadOnly attempted a compiled TeamsEventsPolicy
Global read using the dedicated lab SPN with client-secret and certificate
authentication. Both actual resource reads returned HTTP 403. The harness checks
the response request path to distinguish a resource denial from preparation or
discovery failure. No response payload, tenant/client ID or credential was
retained. Filter/list and optional pinned comparison were not reached. No policy,
role, consent or privilege changed.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; explicit approval remains pending. No workaround or privilege change was
applied. A denied read does not establish successful tenant CRUD behavior.

The harness gates credential access, bounds execution and optional oracle output,
binds comparisons to the canonical tenant and requires all 25 typed nullable
settings. It rejects incomplete/wrong types, wrong tenant, trailing output and
missing/ambiguous Global selections. Getter projections use case-insensitive
last values rather than raw property spelling. The optional certificate oracle
pins the 8.1.0 module/Core/PolicyRp baseline, performs one Global read and disposes
its certificate/session. No writes run.

Remaining qualification includes authorized read/list/filter and pinned comparison;
disposable policy create/update/read-back/delete; safe assignment targets and
restoration; null/empty/Force/Rank and server constraints; cloud/authentication
coverage; and effective webinar, townhall, recording/transcription, chat, ECDN and
integration behavior. HTTP acceptance or a group 202 result does not prove an
effective assignment. Issue #118 and remaining tenant testing remain open.
