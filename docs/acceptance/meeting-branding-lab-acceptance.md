# Meeting branding lab acceptance

MSTEAMS_LAB_READ_ONLY=1 and MSTEAMS_LAB_IDENTITY_FILE enable
TestAccMeetingBrandingReadOnly with the existing private dedicated lab identity.
MSTEAMS_LAB_AUTH_METHOD selects certificate (default) or client-secret. The test
reads Global and its locally filtered list; it creates no policy/assets and changes
no Global setting or assignment. Failures retain only a fixed status message.

Optional MSTEAMS_LAB_POWERSHELL_ORACLE=1 with the absolute pinned module path enables
certificate-authenticated Get-CsTeamsMeetingBrandingPolicy. The script checks the
baseline/Core/PolicyRp hashes and tenant binding, reads only Global, and projects six
nullable scalars. Go bounds and clears transient output, rejects mixed/incomplete or
wrong-tenant observations, and compares only that selected projection. Nested asset
serialization is checked offline; scalar comparison does not prove Premium effects.

Both actual dedicated-SPN secret/certificate attempts returned resource HTTP 403.
List and optional PowerShell comparison were not reached. This establishes neither
unsupported application mode nor a feature/RBAC cause; successful qualification is
still required. No tenant mutation has run for this candidate.

Future persistence acceptance must use an owned unassigned policy with explicit
mutation opt-in and a private tenant/identity-bound lifecycle journal. Record intent
before Create; read a non-default scalar/asset, update a separate field while checking
omission/null/empty semantics, verify list, delete and confirm absence. Validate
ownership before recovery and investigate ambiguous writes instead of replaying them.
Use controlled asset URIs and document tenant/Premium prerequisites separately.

User/group/tenant grants and edits to Global can change effective meeting behavior.
They require consequential opt-in, controlled subjects, prior assignment snapshots,
an explicit restore plan and bounded propagation observation. Policy read-back and
HTTP acceptance do not prove meeting rendering or user-effective behavior. These
write/assignment scenarios remain unimplemented/unrun qualification work, so issue
#97 remains open.
