# POL-27 policy assignment tenant acceptance

Issue #100 remains open. The ordinary service has offline request/model checks;
eight conditional OCE commands are metadata-reconciled and still need native
implementation. No successful live policy-assignment qualification is claimed.

`TestLabPolicyAssignmentsReadOnly` requires MSTEAMS_LAB_READ_ONLY=1 and the private
MSTEAMS_LAB_IDENTITY_FILE. MSTEAMS_LAB_AUTH_METHOD selects client-secret or certificate.
It lists generated jobs, then reads at most five returned operation IDs. It never
submits work, edits assignments/rank, changes permissions or invokes OCE actions.
Failures retain only HTTP status and whether discovery failed, without payloads.

Actual dedicated-SPN runs in both credential modes completed discovery and received
resource HTTP 403 on the list. Per-ID reads were not reached. The Get command's live
state is unavailable; no successful contract/live parity state is promoted.

Successful future qualification needs list/status comparison against the pinned
module and explicit tenant-bound fixtures for group/user effective reads. Preserve
rank and direct/group source provenance without inventing a precedence rule.
Group creation/update/deletion and batch submission need explicit mutation gates,
owned disposable fixtures, non-default read-back and defined cleanup. A 201/202/204
alone proves acceptance, not effective assignment convergence or per-user success.

OCE reads/actions have additional contributor eligibility and explicit tenant,
environment and regional context. They must be exercised separately from ordinary
tenant policy operations. No OCE profile is activated, no role is granted and no
eligibility is inferred from a denial. Reprocessing/refresh may affect existing
users and require a separately scoped consequential test, not automatic cleanup.
