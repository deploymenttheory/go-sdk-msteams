# Meeting-template permission policies

Use `sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies` for the five
public compiled MicrosoftTeams 8.1.0 commands and their nine parameter sets.

| Operation | Methods |
|---|---|
| Read/list/filter | `GetByIdentity`, `List`, `ListFiltered` |
| Create | `CreateByIdentity` |
| Update | `PatchByIdentity` for one sparse PATCH; `UpdateByIdentity` for Get then local list edits then PATCH |
| Delete | `DeleteByIdentity` |
| Scoped grant | `AssignUser`, `AssignGroup`, `AssignTenantDefault`, or exactly one scope in `Assign` |

This family uses direct compiled PolicyRP routing for CRUD and grants. Grant's
parameterless Identity set has no source operation and rejects locally. User,
group and tenant scopes preserve nullable versus empty policy names. Group Rank
is a required signed Int32; zero and negative values are sent unchanged. The
tenant grant preserves the source's capitalized `PolicyName`, null
`additionalParameters` and Boolean `forceSwitchPresent`. HTTP acceptance does not
prove propagation or effective assignment.

`ScalarSettings` preserves omitted/null/value Description and
DefaultMeetingTemplateId. These source strings have no inferred enum, GUID or
length validation. `HiddenTemplate` preserves Identity, nullable/missing Id and
unknown response fields; an explicit null Id setter fails like the pinned class.
`NewHiddenTemplate` is a pure local constructor accepting a Go string.

A nil `HiddenTemplateEdit` leaves the list unbound. `ClearHiddenTemplates` clears
it, `ReplaceHiddenTemplates` replaces it, and `ChangeHiddenTemplates` removes the
first ordinal Id match per Remove item then appends Add items without deduplication.
Results own their pointers and snapshots. An empty delta clears the list like the
source's empty modifier. Native requests send complete replacement arrays, not
server-side Add/Remove operators. A bound nil slice sends `[]`; explicit array
null is not a source DTO representation.

```go
_, response, err := sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies.CreateByIdentity(
    ctx, "Tag:Lab", teamsmeetingtemplatepermissionpolicy.CreateRequest{
        ScalarSettings: teamsmeetingtemplatepermissionpolicy.ScalarSettings{
            Description: policy.Set("Lab meeting templates"),
        },
        HiddenMeetingTemplates: teamsmeetingtemplatepermissionpolicy.ClearHiddenTemplates(),
    },
)
```

Update preparation and HTTP acceptance remain available in `UpdateResult` on
partial failure. Force alone is not a setting. Bound CRUD Force is serialized as
the compiled string representation. Mutations never replay automatically.

The dedicated lab SPN uses Teams Administrator; no additional permissions were
granted for this service. Global read/list and owned create/read/update/list/delete
passed under certificate and client-secret authentication. Created/updated
getters matched the pinned module; a description-only update preserved the bound
empty hidden-template list. Deletion required GET404 and list absence. Tests use
uniquely named unassigned policies and private recovery journals.

Permanent acceptance responses in `testdata` are used by read/CRUD unit tests.
Failure injection separately checks source-defined grant requests, authorization,
errors, cancellation and replay safety. Local list-edit tests are not tenant
responses. Live nonempty hidden lists, default-template changes, grants/clear,
propagation, effective meeting behavior, delegated flows and other clouds remain
unqualified; issue #132 remains open.

See [qualification details](../../../docs/acceptance/meeting-template-policies-lab-acceptance.md).

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
