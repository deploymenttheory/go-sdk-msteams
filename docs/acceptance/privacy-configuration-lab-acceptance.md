# Privacy configuration lab acceptance

`MSTEAMS_LAB_READ_ONLY=1` and `MSTEAMS_LAB_IDENTITY_FILE` enable
TestAccPrivacyConfigurationReadOnly with the existing private dedicated lab identity.
MSTEAMS_LAB_AUTH_METHOD selects certificate (default) or client-secret. The test
reads Global and a locally filtered list. It changes no privacy setting, directory
permission or role. Actual success/denial must be recorded separately from offline
mode routing; neither a token nor an HTTP acceptance proves effective behavior.

On 2026-11-21, both actual credential modes returned resource HTTP 403 for Global.
List and the optional PowerShell comparison were not reached. No response payload
was retained. This establishes a denial with the configured identity, without
inferring an unsupported mode, exact role requirement or feature/license cause.

Optional MSTEAMS_LAB_POWERSHELL_ORACLE=1 with the absolute pinned module directory
enables a certificate-authenticated Global read. The script checks version and
Core/PolicyRp hashes and tenant binding, then projects the four nullable Boolean
settings. Go bounds/clears transient output and rejects mixed output, incomplete
fields, non-Boolean values or a different tenant/identity. No tenant data is retained
in contract reports. This comparison is unqualified until actually run successfully.

Privacy writes can affect tenant contact initiation, published location and photo
behavior. Future acceptance requires explicit mutation/consequential opt-ins, a
private tenant/identity-bound journal, original settings, authorized restoration
and post-write read-back. Verify a non-default value, omission of other settings,
explicit null and false; observe feature behavior separately. Investigate ambiguous
writes before retry/restoration. No write/read-back or feature-effect scenario has
run. Issue #102 remains open for full management, permission/cloud and tenant parity.
