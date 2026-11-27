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
record or Global setting is edited. Creating a record does not establish its
rendering in meetings or effective user behaviour.

CI exports only bounded, sanitized response JSON and empty HTTP receipts. Review
new `crud_id_*` and `crud_identity_*` fixtures before committing them as ordinary
unit regressions; do not infer a successful tenant run from an offline example.
See [test setup](../../tools/acceptance/README.md) and [SDK methods](../../msteams/teams_admin_api/recording_messages/api_docs.md).
