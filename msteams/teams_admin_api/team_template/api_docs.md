# Team templates

Use `sdk.TeamsAdminAPI.TeamTemplates` for the native ConfigAPI template resource.
These operations manage reusable templates, not teams created from templates.
They use the shared client/auth/transport controls and need no PowerShell runtime.
Constructing the root client or this service performs no network I/O.

| Resource method | Wire operation | Declared successful statuses |
| --- | --- | --- |
| `Create(ctx, locale, definition)` | POST `/Teams.Templates/api/TeamTemplates/v1.0/Tenant/{locale}` | 201 |
| `GetByID(ctx, odataID)` | GET `/Teams.Templates/{odataID}` | 200 |
| `UpdateByID(ctx, odataID, definition)` | PUT `/Teams.Templates/{odataID}` | 200, 201 |
| `DeleteByID(ctx, odataID)` | DELETE `/Teams.Templates/{odataID}` | 200, 204 |
| `List(ctx, options)` | GET `/Teams.Templates/api/TeamTemplates/v1.0/{locale}` | 200 |

Every method takes a context and returns the actual Resty response alongside its
value/error. Mutations are never replayed automatically. Create/update return
the immediate typed receipt; no polling operation or automatic read-back is
invented. Verify managed state with Get/List when a workflow requires it.

## Identity, locale and collection results

Use a returned `@odata.id`, not a bare summary `id`, for Get/Update/Delete. The
whole reference is escaped once as one path segment, including its slashes. It
is never followed as an arbitrary URL. `Identity` supplies the selectors used
by InputObject parameter sets: extract `ODataID` or `Locale` and call the same
native method. No PowerShell session object is required.

`List(ctx, nil)` and an omitted `ListOptions.PublicTemplateLocale` use `en-US`.
An explicitly empty/whitespace locale fails locally instead of silently replacing
an already-bound parameter. Create requires an explicit nonempty locale. Local
path/control validation runs before discovery/credential acquisition; other
locale, field and service constraints remain server-authoritative.

List returns a `ListResult` with typed summaries, raw envelope fields and service
diagnostics. Locale selects Microsoft templates; the module's list contract
includes tenant custom templates regardless of locale and does not filter out
hidden templates. There is no evidenced pagination contract. `IsIncomplete`
is true when `X-Teams-TemplateSvc-Diagnostics` contains
`TemplatesRetrievalFailure`, case-insensitively. Check that flag before using a
list to establish absence; an incomplete collection must not become an empty
success during cleanup/reconciliation.

## Definitions and response snapshots

`NewDefinition(displayName, shortDescription)` constructs sparse expanded input
and validates the two mandatory strings. Optional strings/booleans use pointers
to distinguish omission from empty/false. Nil collections are omitted; explicit
empty collections remain `[]`. Null/empty object items are dropped by the pinned
array serializer. Empty settings objects retain their known false defaults.

`ParseDefinition` and JSON decoding into `Definition` reproduce the generated
Body constructor. They initialize five settings objects with false defaults,
even when those settings are absent/null. Known properties of another JSON type
are ignored as in the pinned typed readers. Root fields win over
`teamDefinition`; `extendedProperties` fills fields still absent. A root null
also masks a nested value. Unknown Body fields are not arbitrary write
passthrough. Use sparse struct construction when omitted settings must stay
omitted, and choose the JSON Body path deliberately.

Create appends a General channel if none has that display name under
case-insensitive comparison. The inserted channel has ID/displayName `General`
and `isFavoriteByDefault: true`. It copies the slice before appending and does
not mutate caller input or force an existing General channel's settings. Update
sends the supplied definition without this insertion. PUT is the evidenced
method; the SDK does not promise PATCH semantics or silently read/merge state.

`Template`, `TemplateSummary` and error models project known fields while
retaining unknown/null/empty values in `RawFields`. JSON encoding a received
read model preserves its response snapshot. `Template.Definition` is the known
flattened projection for deliberate write construction. Editing typed fields
does not rewrite an existing raw snapshot; use an explicit Definition for writes.

## Errors, routing and authentication

Unexpected successful statuses return `ErrResponse`; malformed successful
envelopes return `client.ErrDecode` with response metadata retained. A genuine
empty delete 204 succeeds. A delete 200 Any payload remains in `response.Body()`.

HTTP failures return `*teamtemplate.Error`, wrapping the shared `*client.APIError`
for `errors.As` and auth/error classification. Its optional `Envelope` retains
the operation ID and code/message/action objects, including unknown fields. An
empty or malformed error body does not erase its HTTP status or become a decode
success. Error formatting prints no service bodies, messages or identifiers.
Raw response/envelope values are for explicit inspection. Unlike PowerShell's
console handlers, named non-success responses are returned as Go errors.

The collection shares one lazy ConfigAPI forest-discovery adapter with other
ConfigAPI resources. The normalized forest is `X-MS-Target-Uri` routing data;
bearers stay bound to the configured ConfigAPI origin/audience. Each operation
supplies the corresponding `X-MS-CmdletName`; the SDK retains its own user-agent.

Use the root client's supported auth provider patterns. For app-only `*-Cs`
management, Microsoft's setup requires Graph `Organization.Read.All` and an
appropriate Entra directory role. Teams Administrator is the lab role requested
for template/policy management. Additional Graph write permissions are not a
substitute for that role, and Microsoft says not to configure an API permission
for the Skype/Teams Tenant Admin resource.

Sources: [application authentication](https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication),
[Teams roles](https://learn.microsoft.com/en-us/microsoftteams/using-admin-roles),
[pinned template contract](../../../docs/implementation/core/template-policy-deployment-evidence.md).
Offline cloud/mode routing tests do not claim actual sovereign service eligibility.

## Baseline command and parameter-set mapping

| PowerShell 8.1.0 command/sets | Native SDK composition |
| --- | --- |
| Get-CsTeamTemplate: Get, GetViaIdentity | GetByID with the direct or Identity.ODataID selector |
| Get-CsTeamTemplateList: DefaultLocaleOverride | List with omitted/explicit locale options |
| New-CsTeamTemplate: New, NewViaIdentity | Definition/ParseDefinition + Create with direct or Identity.Locale selector |
| New-CsTeamTemplate: NewExpanded, NewViaIdentityExpanded | NewDefinition + optional typed fields + Create |
| Update-CsTeamTemplate: Update, UpdateViaIdentity | Definition/ParseDefinition + UpdateByID |
| Update-CsTeamTemplate: UpdateExpanded, UpdateViaIdentityExpanded | NewDefinition + optional typed fields + UpdateByID |
| Remove-CsTeamTemplate: Delete, DeleteViaIdentity | DeleteByID with the direct or Identity.ODataID selector |
| New-CsTeamsHiddenTemplate: Identity | NewHiddenTemplate, a local ID value only |
| New-CsTeamsHiddenMeetingTemplate: Identity | NewHiddenMeetingTemplate, a local ID value only |

Local hidden values do not persist or assign permission policies. Their eventual
policy payload serialization belongs to the corresponding permission-policy
service. Source/model/request evidence and implementation qualification are
recorded in [the SDK evidence note](../../../docs/implementation/core/team-template-sdk-evidence.md).

## Lab acceptance

`TestLabTeamTemplateCRUD` is opt-in and uses the dedicated lab identity. Its
expanded and JSON Body variants each journal intent, create one random tenant
template, resolve its reference from
a complete list and verifies ownership, reads non-default fields and General,
updates explicit fields, lists, deletes and verifies absence. Cleanup rechecks
the owned name and tenant reference; failures/ambiguity retain private state.
It creates no team and changes no permission policy or Microsoft template.

Required inputs are `MSTEAMS_LAB_READ_ONLY=1`, `MSTEAMS_LAB_MUTATIONS=1`, an
explicit private `MSTEAMS_LAB_JOURNAL_DIRECTORY`, and
`MSTEAMS_LAB_IDENTITY_FILE`. `MSTEAMS_LAB_AUTH_METHOD` selects certificate
(default) or client-secret. `MSTEAMS_LAB_TIMEOUT_SECONDS` accepts 1–300.

Set `MSTEAMS_LAB_POWERSHELL_ORACLE=1` and an explicitly verified 8.1.0 module path
in `MSTEAMS_LAB_MODULE_PATH` to compare the same owned template's fields/channel
order after create and update. The oracle connects and reads only that fixture;
its output is compared in memory and not logged. Ordinary tests skip before
reading credentials. No live CRUD success is claimed while the role grant is pending.

After an interrupted run, select its private file in
`MSTEAMS_LAB_RECOVERY_FILE` and run `TestLabTeamTemplateRecover` with the same
mutation opt-in, tenant/cloud/identity binding and private journal directory.
An incomplete list, failed lookup, foreign reference, renamed fixture or
ambiguous match remains unresolved rather than authorizing deletion.
An unresolved creation with no verified reference also remains retained when a
list is empty; recovery does not turn that uncertainty into verified absence.
