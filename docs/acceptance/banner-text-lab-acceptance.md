# Banner text tenant tests

`teams_custom_banner_text` runs bounded list/full-identity/relative-ID reads against
the configured tenant. Current reads pass with the lab application identity; the
earlier 403 and administrator-grant notes are historical, not current prerequisites.

`TestAccBannerTextCRUDByID` and `TestAccBannerTextCRUDByIdentity` create one owned,
unassigned record with non-default text, read it back, update only text while
preserving the description, verify its list entry, delete it and independently
check selected/list absence. The existing PowerShell reader compares both states
when enabled, including the list count and the selected scalar projection.

Writes require mutation opt-in and a private recovery journal. The journal's random
fixture determines a UUID and an exact description ownership marker. Recovery
refuses foreign IDs, changed ownership, denied reads and inconsistent absence.
It uses a fresh bounded context after cancellation and retains unresolved state.
These tests edit no existing banner, Global setting or policy assignment. Displayed
banner behaviour and effective user settings require separate controlled subjects.

CI exports only bounded, sanitized response JSON and empty receipts. Review new
`crud_id_*` and `crud_identity_*` captures before using them in offline tests.
See [test setup](../../tools/acceptance/README.md) and [SDK methods](../../msteams/teams_admin_api/teams_custom_banner_text/api_docs.md).
