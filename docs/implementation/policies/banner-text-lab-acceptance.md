# Teams custom banner text lab acceptance

On 2026-11-21, TestAccBannerTextReadOnly attempted the CustomBannerText collection
read using the dedicated lab SPN in client-secret and certificate modes. Both
actual resource reads returned HTTP 403. No response payload, tenant/client ID
or credential was retained. Filtered/selected reads and optional pinned comparison
were not reached. No banner, policy, assignment, role, consent or privilege changed.
This authorization failure does not establish feature availability or live CRUD.

Automatic approval review previously rejected the tenant-wide Teams Administrator
grant; explicit approval remains pending. No workaround or privilege change was
applied.

The harness gates credential access, bounds execution and optional oracle output,
and compares list/filter with at most five full/relative selector reads. Its
optional certificate oracle pins the 8.1.0 module/Core/PolicyRp hashes, binds to
the canonical tenant, compares source list/selected reads, and reports only known
nullable Id/Text/Description projections. Even an empty SDK list requires a source
list/count comparison when the oracle is enabled. Strict decoding rejects missing
fields, wrong types/tenant/count/selectors, duplicates and trailing output.
Certificates/sessions are disposed and output buffers cleared. No writes run.

Remaining qualification includes successful authorized list/filter/read and
pinned comparisons; uniquely tagged create/default-ID/read-back/update/delete and
cleanup; source null/empty/Force and server validation; read-derived selector
semantics, cloud/authentication coverage and actual displayed banner behavior.
Issue #114 and full command/tenant qualification remain open.
