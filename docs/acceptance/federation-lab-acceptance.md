# Federation tenant tests

The SDK exposes external-access policy, ACS federation, Teams external-access and
tenant-federation services. Each has JSON-backed request/model/error regressions.
Current configured tenant read jobs pass for ACS, Teams external-access and tenant
federation. The older dedicated-SPN 403 observations no longer describe those runs.

`TestAccExternalAccessPolicyCRUD` creates one random unassigned policy, checks
non-default description/false values, sparsely updates its description while
preserving the omitted Boolean, verifies its list entry and confirms selected/list
absence after deletion. Reads compare selected projections with the existing
PowerShell reader. Writes require mutation opt-in and a private bound journal;
cleanup cannot delete an unrelated name or silently ignore an uncertain result.

Edits to tenant/Global federation settings and user/group/tenant policy grants may
change effective access. They require controlled subjects, prior-state snapshots,
consequential opt-in and independently verified restoration. The unassigned-policy
test does not establish federation propagation or effective external access.
See [lifecycle setup](lifecycle-and-recovery.md).
