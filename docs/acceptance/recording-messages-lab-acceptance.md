# Recording message tenant tests

`recording_messages` runs bounded list/full-identity/relative-ID reads against the
configured tenant. The current read scenario passes with the lab application
identity; the earlier 403 observation no longer describes current access.

`TestAccRecordingMessagesCRUDByID` and `TestAccRecordingMessagesCRUDByIdentity`
exercise an owned, unassigned message: create an `en-US` localization, read it
back, replace its notice text while preserving the omitted description, verify
its list entry, delete it and independently verify selected/list absence. Both
states are compared with the existing PowerShell reader when enabled.

Writes require mutation opt-in and a private recovery journal. The random journal
fixture determines a UUID before submission; its exact description is an ownership
marker. Recovery refuses a different UUID, changed marker, denied read or
inconsistent absence. Intent is written before creation, and a fresh bounded
context handles cleanup after cancellation. No policy is assigned and no existing
record or Global setting is edited. Microsoft documents that the message takes
effect only after its ID is associated with a policy; these tests leave it
unassigned. See [message configuration](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsrecordingandtranscriptioncustommessage).

CI exports only bounded, sanitized response JSON and empty HTTP receipts. Review
new `crud_id_*` and `crud_identity_*` fixtures before committing them as ordinary
unit regressions; do not infer a successful tenant run from an offline example.
See [test setup](../../tools/acceptance/README.md) and [SDK methods](../../msteams/teams_admin_api/recording_messages/api_docs.md).

## Executed owned lifecycle

Both ID and full-identity scenarios passed with client-secret authentication in
[run 45984997202](https://github.com/deploymenttheory/go-sdk-msteams/actions/runs/45984997202)
on 2026-11-27, at the PR merge revision `de1aec55e7f5d3020c9cee875588d6058b22358d`.
Each scenario completed create/read-back, sparse update/omission, list and
independent deletion/absence checks, with two successful PowerShell comparisons.
Both recovery entries were cleaned; the finalizer reported zero unresolved entries.
Sanitized response JSON and empty receipts now drive `TestUnitCapturedOwnedLifecycle`
in the service package. The unit test clearly labels its synthetic PATCH acknowledgement.
