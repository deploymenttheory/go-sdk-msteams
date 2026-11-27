# Teams feedback policy lab acceptance

On 2026-11-21, TestAccFeedbackPolicyReadOnly attempted a compiled
TeamsFeedbackPolicy Global read with the dedicated lab SPN using client-secret
and certificate authentication. Both actual resource reads returned HTTP 403.
The harness verifies the response request path to distinguish resource denial
from preparation/discovery failure. No response payload, tenant/client ID or
credential was retained. List/filter and optional pinned comparison were not
reached. No policy, assignment, role, consent or privilege changed.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; explicit approval remains pending. No workaround or privilege change was
applied. A denied read does not establish successful tenant CRUD behavior.

The harness gates credential access, bounds execution and optional oracle output,
binds comparisons to the canonical tenant and requires all six typed nullable
settings. It rejects incomplete/wrong types, wrong tenant, trailing output and
missing/ambiguous Global selections. Canonical getter projections preserve
case-insensitive last values. The optional certificate oracle pins the 8.1.0
module/Core/PolicyRp baseline, performs one compiled Global read and disposes its
certificate/session. No writes run.

Remaining qualification includes authorized compiled reads/list/filter and pinned
comparison; disposable create/update/read-back/delete and cleanup; modern grants
to safe targets with read-back/restoration; omission/null/empty/Force/Rank and
server constraints; cloud/authentication coverage; and effective survey, feedback
and collection behavior. HTTP acceptance or a group 202 result does not prove
effective assignment. Issue #119 remains open for full command/tenant acceptance.
