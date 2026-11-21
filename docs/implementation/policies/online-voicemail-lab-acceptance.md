# Online voicemail lab acceptance

MSTEAMS_LAB_READ_ONLY=1 and MSTEAMS_LAB_IDENTITY_FILE enable
TestAccOnlineVoicemailReadOnly using the existing private dedicated identity.
MSTEAMS_LAB_AUTH_METHOD selects certificate (default) or client-secret. The test
reads Global and a locally filtered list. Both actual credential-mode attempts
returned resource HTTP 403; list and optional PowerShell comparison were not
reached. No unsupported mode or feature/RBAC cause is inferred from denial.

Optional MSTEAMS_LAB_POWERSHELL_ORACLE=1 and the absolute pinned module path enable
a certificate-authenticated Global read. The script checks Core/PolicyRp/baseline
hashes and tenant binding, and projects Description, EnableTranscription and the
constant duration string. Child output is bounded/cleared and never retained.
Mixed/incomplete/wrong-tenant projections fail locally. Scalar equality would not
establish voicemail delivery, audio-file resolution or effective assignments.

Future policy persistence acceptance requires an owned unassigned policy, explicit
mutation opt-in and a private tenant/identity-bound journal. Record intent before
Create; verify a non-default Boolean/duration, sparse update omission and explicit
null/zero, list, delete and confirmed absence. Audio references and languages must
use controlled fixtures with independently verified prerequisites. Resolve ambiguous
writes through ownership-bound reads instead of replay/compensation.

User/group/tenant assignment changes and Global edits require consequential opt-in,
controlled subjects, assignment snapshots and an explicit restore plan. Observe
group job status separately from policy/effective behavior, with bounded caller-chosen
waits and unknown states preserved. Native null/empty and Priority zero/omission need
separate cases from the PowerShell binding composition. Those write/assignment
scenarios remain unimplemented/unrun qualification work. Issue #98 remains open.
