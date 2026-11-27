# Education assignments app policy lab acceptance

On 2026-11-21, TestAccEducationAssignmentsReadOnly attempted a compiled
TeamsEducationAssignmentsAppPolicy Global read using the dedicated lab SPN with
client-secret and certificate authentication. Both actual resource reads returned
HTTP 403. No response payload, tenant/client ID, API key or credential was retained
in the repository. Filter/list and optional pinned comparison were not reached.
No policy, assignment app setting, role, consent or privilege changed. This
authorization failure does not establish server availability or successful CRUD.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; explicit approval remains pending. No workaround or privilege change was
applied.

The read-only harness gates credential access, bounds execution and optional
oracle output, binds source comparisons to the canonical tenant and requires all
five typed nullable settings. It rejects wrong/missing fields, wrong tenant,
invalid URL models, trailing output and missing/ambiguous Global selections.
Canonical comparisons use typed getter projections, including last valid
case-insensitive property values, rather than received key spelling. The optional
certificate oracle pins the 8.1.0 module/Core/PolicyRp baseline, performs one
Global read, and disposes its certificate/session. No writes run.

Remaining qualification includes authorized read/list/filter and pinned comparison;
non-default setting updates/read-back and restoration; source URL/null/empty/Force
and server validation; cloud/authentication coverage and effective parent digest,
MakeCode and TurnItIn behavior. The public baseline has Get/Set only; private Core
New/Remove/Grant classes do not imply public create/delete/assignment operations.
Issue #115 and remaining tenant testing remain open.
