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
These tests edit no existing banner, Global setting or policy assignment. Banner
activation requires associating its ID with a compliance-recording policy; see
[Microsoft's setup guide](https://learn.microsoft.com/en-us/microsoftteams/teams-compliance-recording-custom-banner).
Displayed behaviour requires separate controlled subjects.

CI exports only bounded, sanitized response JSON and empty receipts. Review new
`crud_id_*` and `crud_identity_*` captures before using them in offline tests.
See [test setup](../../tools/acceptance/README.md) and [SDK methods](../../msteams/teams_admin_api/teams_custom_banner_text/api_docs.md).

## Executed owned lifecycle

Both ID and full-identity scenarios passed with client-secret authentication in
[run 45984997202](https://github.com/deploymenttheory/go-sdk-msteams/actions/runs/45984997202)
on 2026-11-27, at the PR merge revision `de1aec55e7f5d3020c9cee875588d6058b22358d`.
Each scenario completed create/read-back, sparse update/omission, list and
independent deletion/absence checks, with two successful PowerShell comparisons.
Both recovery entries were cleaned; the finalizer reported zero unresolved entries.
Sanitized response JSON and empty receipts now drive `TestUnitCapturedOwnedLifecycle`
in the service package. The unit test clearly labels its synthetic PATCH acknowledgement.
