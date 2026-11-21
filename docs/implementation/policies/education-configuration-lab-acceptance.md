# Teams education configuration lab acceptance

On 2026-11-21, TestLabEducationConfigurationReadOnly attempted a compiled
TeamsEducationConfiguration Global read using the dedicated lab SPN with
client-secret and certificate authentication. Both actual resource reads returned
HTTP 403. No response payload, tenant/client ID or credential was retained.
Filter/list and optional pinned comparison were not reached. No education setting,
policy, role, consent or privilege changed. This authorization failure does not
establish server availability or successful tenant read/update behavior.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; explicit approval remains pending. No workaround or privilege change was
applied.

The harness gates credential access, bounds execution and optional oracle output,
binds comparisons to the canonical tenant and requires all three typed nullable
settings. It rejects incomplete/wrong types, wrong tenant, trailing output and
missing/ambiguous Global selections. Getter projections preserve case-insensitive
last values rather than raw property spelling. The optional certificate oracle
pins the 8.1.0 module/Core/PolicyRp baseline, performs one Global read and disposes
its certificate/session. No writes run.

Remaining qualification includes authorized read/list/filter and pinned comparison;
non-default setting updates/read-back/restoration; source null/empty/Force and
server constraints; cloud/authentication coverage and effective parent contact,
generative AI enhancement and parent information behavior. Only Get/Set public
operations are assigned in this baseline. Issue #116 and full command/tenant
qualification remain open.
