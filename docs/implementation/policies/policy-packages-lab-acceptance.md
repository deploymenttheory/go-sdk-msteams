# POL-28 delegated tenant acceptance

Issue #101 remains open. Native methods, request/model observations and exact
parameter-set mappings do not establish successful tenant qualification.

TestAccPolicyPackagesReadOnly requires MSTEAMS_LAB_READ_ONLY=1 and the private
MSTEAMS_LAB_IDENTITY_FILE for the expected lab tenant. Without a separately supplied
delegated context it verifies the existing SPN's documented local mode rejection,
then skips tenant qualification. The actual configured-SPN run took this path;
no package request, token acquisition, role/consent change or tenant write occurred.
This is an unavailable delegated prerequisite, not a successful live read.

To exercise reads later, set MSTEAMS_LAB_PACKAGE_DELEGATED_FILE to a private JSON
file (same private-file permissions as other lab credentials). Required fields are
tenant_id, identity_id, access_token and expires_at (RFC3339). Optional cloud selects
the existing environment profile; optional user_id enables selected user/recommendation
reads. The supplied Teams token must belong to a delegated identity in the same
expected tenant, and expire more than one minute in the future. Token mode is the
trusted caller's declaration, not a JWT-inspection or entitlement guarantee.

The loader rejects wrong tenants, expired/empty tokens, unknown JSON fields and
trailing documents without retaining output. It never prints tokens or identifiers.
The read test lists packages and reads at most five returned names; optional user
reads do not grant/modify anything. It never invokes custom CRUD, group/user grants
or asynchronous batch assignment. PowerShell projection comparison remains a
separate qualification step; the SDK itself needs no PowerShell runtime.

Mutation qualification needs explicit gates, disposable tenant-bound package/user/
group fixtures, non-default read-back and cleanup. Keep synchronous partial failures
separate from asynchronous HTTP acceptance and effective policy convergence. No
SPN permission grant can substitute for the documented delegated-mode requirement.
