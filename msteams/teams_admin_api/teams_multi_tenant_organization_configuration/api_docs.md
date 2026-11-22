# Teams multi-tenant organization configuration

Use `sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations` for the pinned
8.1.0 public compiled Get and Set command surface.

| Method | Behavior |
|---|---|
| `GetByIdentity` | Explicit identity; direct compiled PolicyRP GET |
| `List`, `ListFiltered` | Bare collection, exact Global first, local wildcard matching |
| `PatchByIdentity` | One sparse PATCH with at least one bound setting |
| `UpdateByIdentity` | Set preparation GET and sparse PATCH, retaining both receipts |

CopilotFromHomeTenant is a nullable **string** getter and a `policy.Field[string]`
setting in the pinned runtime and model. Omitted/null/empty values remain distinct.
The current web reference inconsistently labels its parameter Boolean; the 8.1.0
public runtime and compiled source define this SDK contract. Enabled and Disabled
are documented service examples, not an invented local enum restriction. No
create/delete exports are assigned to this singleton.

Compiled MPACmdlet and command headers, first-slash scoped identities, literal
filter whitespace, optional compiled Force string and replay-disabled mutations
match the shared direct provider. Empty/invalid input and canceled contexts fail
before credentials. Read models retain unknown fields and owned RawJSON; create
write settings explicitly instead of editing the received snapshot.

Certificate and client-secret acceptance tests read/list Global, compare the
getter with PowerShell, journal its original Enabled/Disabled value, temporarily
toggle it, verify persistence, restore it and independently verify restoration.
CI serializes its Global reads and updates and requires three pinned comparisons
plus a cleaned restore entry. Its recovery handler rejects unrelated targets and
unsupported original snapshots. The tested identity has Teams Administrator.
No licenses, organizations, roles or assignments are changed by this test.

Permanent sanitized acceptance responses drive unit tests. Explicit null/empty
persistence, cross-tenant Copilot use, delegated modes and other clouds remain
separately unqualified.

[Microsoft's management guide](https://learn.microsoft.com/en-us/microsoftteams/copilot-mto)
shows the Enabled/Disabled commands. See [qualification](../../../docs/implementation/policies/mobility-multi-tenant-lab-acceptance.md).
