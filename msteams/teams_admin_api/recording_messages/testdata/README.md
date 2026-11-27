The `crud_id_*` and `crud_identity_*` files are sanitized responses from the
owned-resource tenant lifecycle tests in [run 45984997202](https://github.com/deploymenttheory/go-sdk-msteams/actions/runs/45984997202)
on 2026-11-27, using client-secret authentication against the commercial cloud.
Both snapshots were compared with the existing PowerShell reader. Create/delete
empty receipts and deleted-record errors are retained as returned. The ordinary
unit workflow labels its PATCH acknowledgement separately as synthetic.

Earlier fixtures retain their original meaning and are not overwritten by this capture.
