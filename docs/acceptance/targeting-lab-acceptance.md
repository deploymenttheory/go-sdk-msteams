# Targeting lab acceptance

Status: prepared, not run. The existing dedicated SPN is present; exact tenant
role/Graph approval requests remain pending. New resource audiences and valid
token acquisition are not proofs of authorization or application eligibility.

Both prepared cases require `MSTEAMS_LAB_READ_ONLY=1` and the existing private
`MSTEAMS_LAB_IDENTITY_FILE`. Without opt-in they skip before reading credentials
or fixture files. `TestAccTargetingPolicyReadOnly` reads Global and a local
Global-filtered modern list; it changes no policy/assignment. It does not prove
modern flight eligibility or unflighted TRPS parity.

`TestAccTargetingHierarchyStatusReadOnly` additionally requires a 0600 non-symlink
`MSTEAMS_LAB_HIERARCHY_STATUS_FILE`, containing private JSON `tenant_id`, explicit
`api_version` (1 or 2), and `request_id`. Tenant must match the selected identity;
v2 needs the canonical tenant GUID bound to that session. Blank request_id selects
latest. Use an independently known existing operation for positive visibility;
404 is not evidence of an empty hierarchy or insufficient permissions. The case
does one status read without polling or any import/create/delete.

Optional `MSTEAMS_LAB_POWERSHELL_ORACLE=1` and absolute
`MSTEAMS_LAB_MODULE_PATH` enable the existing hash-guarded read-only child oracle.
It compares Get-CsTeamsTargetingPolicy Global or the selected
Get-TeamTargetingHierarchyStatus. Tenant-scoped requests must match the child
identity. Certificate app auth is noninteractive. Stdout is bounded/transient,
stderr discarded; identifiers, response/error text and tokens never enter retained
logs/artifacts. Strict report decoding rejects mixed/missing projections. This is
prepared behavior, not a successful live or exported-command comparison.

With those private variables explicitly set, select only the intended probe:

```sh
go test -count=1 -run '^TestAccTargetingPolicyReadOnly$' ./tests/integration/lab/targeting_policy
go test -count=1 -run '^TestAccTargetingHierarchyStatusReadOnly$' ./tests/integration/lab/targeting_hierarchy
```
