# Teams app setup policy lab acceptance

MSTEAMS_LAB_READ_ONLY=1 and MSTEAMS_LAB_IDENTITY_FILE enable
TestAccAppSetupPolicyReadOnly with the existing private dedicated lab identity.
MSTEAMS_LAB_AUTH_METHOD selects certificate (default) or client-secret. The test
reads Global and a locally filtered list, selecting the Global identity while
allowing Tag:Global matches. Known scalar and six-list projections are compared.
No policy, layout, assignment, app installation, role or consent changes run.
No response payload is retained in reports.

On 2026-11-21, the Global read returned resource HTTP 403 with both actual
credential modes. Filter/list and optional PowerShell comparison were not reached.
No unsupported-mode, exact-role, migration or licensing cause is inferred.

Future write acceptance requires explicit mutation/consequential opt-ins, owned
policy/principal fixtures, a private tenant/identity-bound journal, original state
and an authorized cleanup/restore plan. Exercise non-default Boolean/text settings,
omission/null/false/empty distinctions, every list and key type, Int64 boundaries,
persisted read-back, all assignment scopes and cleanup. Correlate ambiguous writes
before retry/restoration. Observe actual Teams client layout separately from HTTP
acceptance. Those write/effect scenarios remain unimplemented/unrun, and #105
stays open for complete command, permission/cloud and tenant qualification.
