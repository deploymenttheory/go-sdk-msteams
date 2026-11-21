# Teams calling policy lab acceptance

On 2026-11-21, `TestLabCallingPolicyReadOnly` used the dedicated lab SPN to read
compiled TeamsCallingPolicy Global with client-secret and certificate auth.
Both actual resource reads returned HTTP 403. No response payload, tenant ID,
client ID, secret, certificate or token was retained in this repository.
Filter/list and optional pinned PowerShell comparisons were not reached.
No policy, calling setting, assignment, privilege, role or consent changed.

Automatic approval review previously rejected a tenant-wide Teams Administrator
grant; no workaround or privilege change was applied. Explicit approval remains
pending. This unavailable read result is not successful CRUD validation.

The opt-in harness checks enablement before credentials, bounds execution and
optional oracle output, binds the oracle to the canonical tenant, requires all
39 typed settings, rejects wrong/missing/ambiguous Global, trailing output,
fractional/overflow Int64 values and compares canonical known projections with
UseNumber. Case variants and later duplicates follow typed getter projections.
The optional oracle pins the module/Core/provider hashes and disposes the
certificate and session. It performs one Global read and no mutations.

Remaining acceptance includes authorized reads/list/filter and pinned source
comparison; unique tagged policy create/read-back/update/delete and cleanup;
all string choices and server ranges; calling, recording, transcription,
interpreter, spend and notification effects; tenant/user/group assignment,
clear/rank behavior, cloud/flight availability and authentication coverage.
Issue #110 remains open. This PR's offline fixtures do not satisfy tenant
qualification.
