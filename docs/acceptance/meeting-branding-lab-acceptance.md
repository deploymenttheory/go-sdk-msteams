# Meeting branding tenant tests

The SDK supports reads, owned policy CRUD, sparse prepared updates, typed theme
builders and explicit assignments. `TestAccMeetingBrandingPolicyCRUD` already has
successful certificate and client-secret runs, with captured JSON regressions.
It creates an unassigned policy with theme override disabled, enables that field
while preserving the omitted background-effect setting, verifies list membership,
and checks selected/list absence after deletion. Both scalar snapshots are compared
with the existing PowerShell reader. See the [SDK reference](../../msteams/teams_admin_api/meeting_branding/api_docs.md).

Reads use `MSTEAMS_LAB_READ_ONLY=1`; CRUD also requires mutation opt-in and a private
journal. Recovery validates the random owned name, retains ambiguous writes and
uses a separate bounded context for cleanup. No Global settings or assignments
change. Historical resource 403 attempts predate these successful CRUD runs.

Theme assets, Premium rendering, user/group/tenant grants and effective meeting
behaviour are outside this owned-policy test. Those scenarios require controlled
subjects, the relevant tenant entitlement, original-state snapshots and an
explicit restoration plan. HTTP acceptance alone does not prove propagation.
