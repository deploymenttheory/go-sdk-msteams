# Teams app permission policy lab acceptance

MSTEAMS_LAB_READ_ONLY=1 and MSTEAMS_LAB_IDENTITY_FILE enable
TestLabAppPermissionPolicyReadOnly with the existing private dedicated lab identity.
MSTEAMS_LAB_AUTH_METHOD selects certificate (default) or client-secret. The test
reads Global and a locally filtered list and compares canonical known fields.
It creates/edits no policy, assigns no app policy, installs no app and changes
no directory permission, consent or role. No payload is retained in reports.

On 2026-11-21, Global returned resource HTTP 403 with both actual credential modes.
Filtered-list and optional PowerShell comparison were not reached. No unsupported
mode, exact role, migration or licensing cause is inferred from that denial.

Optional MSTEAMS_LAB_POWERSHELL_ORACLE=1 and the absolute pinned module directory
enable a tenant-bound certificate read from the hash-checked baseline module.
Known nullable scalar fields and all three catalog lists are projected, including
CLR defaults. The SDK separately retains original presence, duplicate properties
and future response fields. Go bounds/clears transient output and rejects a wrong
tenant/identity, incomplete/mistyped fields or mixed output. Canonical null Id
presentation can represent an absent source setter and is not a transport claim.

Future write qualification requires explicit mutation/consequential opt-ins,
owned policy/principal fixtures, a private tenant/identity-bound journal, original
state and authorized cleanup/restoration. Exercise non-default values and mode
strings, omission/null/empty distinctions, all catalogs' remove/add ordering and
duplicates, persisted read-back, user/group/tenant assignments and cleanup.
Correlate ambiguous writes before retry/restoration; no automatic replay or
effective app-access guarantee is implied. Those scenarios remain unimplemented
and unrun; #104 stays open for full command/permission/cloud/tenant parity.
