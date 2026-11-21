# Recording messages lab acceptance

MSTEAMS_LAB_READ_ONLY=1 and MSTEAMS_LAB_IDENTITY_FILE enable
TestAccRecordingMessagesReadOnly with the existing private dedicated lab identity.
MSTEAMS_LAB_AUTH_METHOD selects certificate (default) or client-secret. The test
lists the resource, compares a locally filtered list and reads at most five
returned full identities. Where a selected record has a canonical Id and matching
Global identity, it also compares the relative-Id read. Known typed projections
are compared. It performs no message, policy, assignment, directory
permission or role write and retains no response payload.

On 2026-11-21, the collection read returned resource HTTP 403 with both actual
credential modes. Selected identity reads and optional PowerShell comparison were
not reached. No unsupported-mode, exact role or licensing cause is inferred.

MSTEAMS_LAB_POWERSHELL_ORACLE=1 and the absolute pinned module directory optionally
enable comparison against certificate-authenticated PowerShell reads for those
same bounded identities. The script checks version, Core/PolicyRp hashes and
tenant binding. All known localization fields are projected, including nullable
CLR defaults; raw future data is independently preserved by SDK snapshots.
Go bounds/clears transient output and rejects wrong tenants, duplicate/unrequested
identities, incomplete or mistyped fields and mixed output. Actual success remains
unqualified until the comparison runs successfully.

Future write acceptance needs explicit mutation/consequential opt-ins, a private
tenant/identity-bound journal, owned test records, original snapshots and an
authorized cleanup/restore plan. Exercise all Identity/Id Create/Update/Delete
forms, generated-ID retention, non-default text, explicit null/empty distinctions,
list remove/add order and duplicates, persisted read-back and cleanup. Correlate
ambiguous writes before deciding whether to retry or restore; no automatic replay
or effective-policy guarantee is implied. Those scenarios remain unimplemented
and unrun, and #103 stays open for full command/permission/cloud/tenant parity.
