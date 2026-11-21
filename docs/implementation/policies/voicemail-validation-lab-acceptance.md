# Voicemail validation configuration lab acceptance

MSTEAMS_LAB_READ_ONLY=1 and MSTEAMS_LAB_IDENTITY_FILE enable
TestLabVoicemailValidationReadOnly with the existing private dedicated lab identity.
MSTEAMS_LAB_AUTH_METHOD selects certificate (default) or client-secret. The test
reads Global and a locally filtered list; it changes no validation setting or role.
Both actual credential-mode attempts returned resource HTTP 403. List and optional
PowerShell comparison were not reached; no unsupported mode or feature/RBAC cause
is inferred. No payload/URI is retained in the report.

Optional MSTEAMS_LAB_POWERSHELL_ORACLE=1 with the absolute pinned module directory
enables a certificate-authenticated Global read. The script checks baseline and
Core/PolicyRp hashes plus tenant binding, and projects the nullable Boolean/URI.
Go bounds/clears transient output and rejects mixed/incomplete or wrong-tenant data.
The URI is a returned string, never a fetch target.

Configuration writes can alter tenant voicemail validation. Future write acceptance
requires explicit mutation/consequential opt-ins, a private tenant/identity-bound
journal, original setting snapshots and an authorized restore plan. Verify a
non-default value, omission of the other setting, explicit null URI and read-back;
observe tenant behavior separately. Investigate ambiguous writes before deciding
whether to retry or restore, and never dereference an arbitrary validation URI.
Those write/effective-behavior scenarios remain unimplemented/unrun. Issue #99
stays open for full command, permission/cloud and tenant qualification.
