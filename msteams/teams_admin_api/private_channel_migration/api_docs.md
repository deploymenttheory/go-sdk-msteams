# Tenant private-channel migration status

`sdk.TeamsAdminAPI.PrivateChannelMigration.Get` reads the singleton migration
status with no caller tenant, URI or destination parameter. It retains TenantId,
MigrationStatus, Details, nullable original timestamp strings and raw future
fields. Unknown migration status values remain readable; no completion/success
interpretation, list, mutation or cancellation contract is invented.

Details is JSON text. `Status.DecodeDetails()` optionally decodes an object of
raw fields without assuming a closed count schema. A malformed detail string
does not invalidate the native status read and remains explicitly available;
the optional decoder returns an error. Ordinary status formatting is redacted.

The selected cloud profile fixes both origin and prefix: commercial/GCC use
distinct paths under teams.microsoft.com; GCCH/DoD use their own gov/dod host and
fabric prefix; China uses teams.microsoftonline.cn. The token audiences are the
matching templates resources. Exact namespaces are enforced before credentials,
including rejecting another cloud's prefix on the same host. No ConfigAPI forest
header, Graph token or commercial fallback is used. France/Germany lack settled
maps and are locally rejected rather than assigned default commercial routing.

Native providers use the selected audience's /.default contract. Exact app
eligibility, tenant administrator enforcement and least-privilege grants remain
unqualified. A valid token or configured route proves neither. Source:
[unit tests](service_test.go).
Requests include the pinned `X-MS-CmdletName` value
`Get-TenantPrivateChannelMigrationStatus` and retain the SDK User-Agent.
Read-only tenant qualification currently observed HTTP 401 before this header
repair; this does not establish application-mode eligibility or ineligibility.
Certificate-connected pinned comparison runs in the credential-owning auth scope;
the published module accepts exactly two Graph/Skype supplied tokens and cannot
use that mode to acquire the separate Templates resource. See the
[unit tests](service_test.go)
for exact assembly provenance, live evidence and remaining qualification gaps.
No migration mutation is performed.
