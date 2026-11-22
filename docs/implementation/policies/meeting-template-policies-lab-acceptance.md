# Meeting-template configuration and permission qualification

Baseline: MicrosoftTeams 8.1.0. Owners: #131 (POL-58) and #132 (POL-59).

The pinned Core cmdlets and their models define configuration Get Identity/Filter,
and permission-policy Get Identity/Filter, Grant GrantToUser/GrantToGroup/
GrantToTenant/Identity, and New/Set/Remove Identity. Only Get is assigned/exported
for the configuration resource. Both use the direct compiled PolicyRP provider.
The permission grant derives from `BaseTpmGrantCmdlet`; the modern public-wrapper
protocol used by meeting/messaging policies is a different contract.

Primary local sources are the pinned Core `TeamsMeetingTemplateConfiguration`,
`TeamsMeetingTemplateType`, `TeamsMeetingOption`,
`TeamsMeetingTemplatePermissionPolicy`, `HiddenMeetingTemplate`, their cmdlets,
and provider `BaseTpm*`, `HttpRequestHelper` and local list-modifier extension.
The assembly hashes are checked before each PowerShell comparison:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`

On 2026-11-22, both Global read tests and permission-policy owned CRUD passed
under certificate and client-secret authentication with the dedicated SDK SPN
(Teams Administrator). Reads compare all public scalar and nested getter fields
with PowerShell 8.1.0. The current configuration has no custom meeting templates;
the permission policy's hidden list is empty. These results do not establish
live nonempty-list behavior or effective meeting behavior.

The owned policy was created with a non-default description and an explicit
empty hidden-template list. A description-only update persisted the new text
and preserved that list. List membership was checked, then deletion required
independent GET404 and collection absence. Recovery completed. No Global/default
configuration, existing policy, user/group assignment, license or privilege was
changed. Existing repository secrets and SDK permissions suffice.

Twelve permanent sanitized service response files come directly from these
acceptance tests. Unit tests read those bodies for get/list/filter, owned CRUD,
typed fields, owned snapshots, status codes and deleted-resource errors. Empty
injected failures and malformed replies exercise failure contracts separately;
local constructors and list edits operate on caller data without HTTP.

The CI registry selects both read tests, ordinary CRUD and the explicit populated-list authorization comparison for both auth methods. Full populated-list CRUD remains separately pending. Its
owned-CRUD proof requires create/read-back, update/read-back plus omission, list,
delete/absence, two pinned comparisons and a cleaned recovery entry. This proof
is scoped to those operations and the observed lab data.

Still open: nonempty nested template and hidden-list tenant fixtures, default
meeting-template mutation, scoped grant/clear persistence and propagation,
effective client behavior, remaining binding scenarios, delegated flows and
other clouds. Full ticket criteria and release qualification are audited apart
from merged code and green selected-operation checks.

## Populated-list authorization comparison

Populated-list create was also attempted using actual IDs read from the tenant's
first-party template configuration. With certificate and client-secret
application authentication, both the native SDK and pinned PowerShell 8.1.0
returned PolicyRP `40013` with an embedded downstream HTTP `401`. These matching
failures are not successful populated-list CRUD. The exact response is retained
as `testdata/create_populated_denied_400.json` and feeds a unit error test.

`TestAccMeetingTemplatePermissionPopulatedAccess` continuously checks that
specific native/module outcome and independently verifies absence for two
journaled SDK-owned intents. CI requires its native denial, pinned comparison
and cleanup markers, plus both cleaned entries. If the backend starts accepting
these requests, this check fails so qualification can be updated.

`TestAccMeetingTemplatePermissionPopulatedCRUD` retains the full nonempty-list
scenario: create with an existing ID, description-only update preserving that
ID, Add/Remove to another existing ID, clear, list and delete. It remains
explicitly pending, rather than being counted as a pass. Delegated
qualification is awaiting access to the signed-in browser; both the Chrome
connector and an approved local check encountered EPERM reading its debugging
connection file. No additional roles, permissions or licenses were changed.

The successful empty-list CRUD test separately checks the actual serialized
PATCH bytes, proving that the hidden list and default-template setting were
omitted from the description-only request. Its empty-list state alone is not
used to establish nonempty-list preservation.

[Microsoft's command reference](https://learn.microsoft.com/en-us/powershell/module/microsoftteams/new-csteamsmeetingtemplatepermissionpolicy?view=teams-ps)
requires existing template IDs. The test looks those up instead of assuming an
example ID exists in the lab.
