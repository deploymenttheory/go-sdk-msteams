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

Future write acceptance needs explicit mutation/consequential opt-ins, a private
tenant/identity-bound journal, owned test records, original snapshots and an
authorized cleanup/restore plan. Exercise all Identity/Id Create/Update/Delete
forms, generated-ID retention, non-default text, explicit null/empty distinctions,
list remove/add order and duplicates, persisted read-back and cleanup. Correlate
ambiguous writes before deciding whether to retry or restore; no automatic replay
or effective-policy guarantee is implied. Those scenarios remain unimplemented
and unrun, and #103 stays open for full command/permission/cloud/tenant parity.
