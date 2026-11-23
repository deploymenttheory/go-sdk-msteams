# Unified apps contracts

Owner: #92 (EXT-30). Baseline: MicrosoftTeams 8.1.0.

## Implementing assembly

The seven M365 app cmdlets aren't in `custom/CmdletConfig.json` or the ConfigAPI
wrappers. They are compiled cmdlets in
`netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll` (SHA256
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`):

| Cmdlet | Type |
|---|---|
| Get-AllM365TeamsApps | `GetUnifiedApps` |
| Get-M365TeamsApp | `GetUnifiedApp` |
| Get-M365UnifiedCustomPendingApps | `GetUnifiedStagedApps` |
| Get-M365UnifiedTenantSettings | `GetTenantSettings` |
| Update-M365TeamsApp | `UpdateUnifiedApp` |
| Update-M365UnifiedCustomPendingApp | `UpdateStagedApp` |
| Update-M365UnifiedTenantSettings | `UpdateTenantSettings` |

## Backend and authorization

`TeamsCmdlet.PerformAuthorization` sets `SubstrateResource` from the session
resource map. It then requests the delegated scope
`08ff1ce2-4973-4b08-86a3-ebed13badc7f/AdminAppCatalog.ReadWrite.All`.

`Microsoft.TeamsCmdlets.PowerShell.Connect.dll` lists `https://substrate.office.com`
only in its commercial endpoint table. The SDK models this as a closed
`AppCatalog` service profile:

- Origin: `https://substrate.office.com`
- Prefix: `/AdminAppCatalog`
- Audience: `08ff1ce2-4973-4b08-86a3-ebed13badc7f`
- Delegated scope: the exact scope above
- Clouds: commercial only

The module's capability documentation already lists every one of these cmdlets
as an application-authentication exception (`app-exception-07` to `-13`).

## Routes

The routes come from IL string literals, and the generic arguments come from the
`HttpUtilities` call sites. All routes sit under
`https://substrate.office.com/AdminAppCatalog`.

| Cmdlet | Request | Model |
|---|---|---|
| Get-AllM365TeamsApps | `GET /ps/v2/admin/unifiedApp?$expand=DeployedTo,AvailableTo,metadata` | `IList<UnifiedAppsResponse>` |
| Get-M365TeamsApp | `GET /ps/v2/admin/unifiedApp/{Id}?$expand=metadata` | `UnifiedAppResponse` |
| Get-M365UnifiedCustomPendingApps | `GET /ps/v2/admin/unifiedApp/stagedApps` | `IList<UnifiedStagedAppsResponse>` |
| Get-M365UnifiedTenantSettings | `GET /ps/v2/admin/unifiedApp/settings[?SettingNames=a,b]` | `List<UnifiedAppsTenantSettingsResponse>` |
| Update-M365TeamsApp | `PATCH /ps/v2/admin/unifiedApp` | `UnifiedAppUpdateRequestBody` |
| Update-M365UnifiedTenantSettings | `PATCH /ps/v2/admin/unifiedApp/settings` | `UnifiedAppsTenantSettingsRequest` |
| Update-M365UnifiedCustomPendingApp | `GET …/stagedApps`, then `GET /v2/admin/unifiedApps/stagedApps/{Id}/details`, then `PUT /ps/v2/admin/unifiedApp/stagedApps/review` | `StagedAppDetailsResponse`, `StagedAppUpdateRequestBody` |

All three list reads decode as bare JSON arrays.

## Serialization

Serialization uses Newtonsoft with a camel-case contract resolver and explicit
`JsonProperty` names:

- `UnifiedAppUpdateRequestBody` is `{id, isBlocked, availableTo, deployTo}`.
  - `availableTo` is `{type, users, groups, operation}`.
  - `deployTo` adds `version`.
- The tenant settings request is `{settingName, settingValue, users, groups, operation}`.
- The review request is `{Id, ReviewStatus, StagedAppDefinitionEtag}`, in
  PascalCase via explicit names.

Response names:

- App:
  - `id`, `isBlocked`, `appType`, `appName`.
  - `availableTo` with `availableToType`, `users`, `groups`, `lastUpdatedDate`,
    `assignedBy`.
  - `deployedTo` with `deployedToType`, `deployedBy`, `deployReason`, `version`,
    `users`, `groups`.
  - Assignment entries carry `id` and `assignedBy`/`deployedBy`, typed as
    string arrays.
- Pending app:
  - `id`, `externalId`, `iteration`, `createdBy`, `lastUpdateDateTime`,
    `reviewStatus`.
  - `metadata` with the manifest fields.
- Details: `stagedAppEntitlement.iteration`.
- Error: `{error:{code,message,innerError:{request-id,date}}}`.

The serializer's null-handling constant isn't recoverable from the call listing.
The SDK omits unset optional request fields.

## Validation mirrored before I/O

**Update-M365TeamsApp:**

- At least one of IsBlocked, AppAssignmentType or AppInstallType is required.
- `UsersAndGroups` requires an operation and at least one user or group. Members
  are de-duplicated.
- `Everyone` and `Noone` forbid members and unconditionally set operation `add`.
- Any other type is rejected.
- The operation setters lowercase their value.

**Update-M365UnifiedTenantSettings:** the ValidateSets are case-sensitive:

- names: DefaultApp, GlobalApp, PrivateApp, EnableCopilotExtensibility
- values: Some, All, None
- operation: add, remove

DefaultApp, GlobalApp and PrivateApp reject `Some` and any users, groups or
operation.

**Update-M365UnifiedCustomPendingApp:**

- The status is title-cased, and must be Published or Rejected.
- The ID must be in the staged list ("Please provide the Id for an existing
  Staged App").
- The details must carry an iteration ("The provided app does not have a Staged
  App Definition Etag").

The SDK returns `ErrPendingAppNotFound` or `ErrMissingIteration` for the last
two, and never sends the review.

## Lab observations (2026-11-23)

The lab identity is client-secret and certificate, app-only, Teams Administrator.

- Raw app-only tokens for `https://substrate.office.com/.default` and
  `08ff1ce2-4973-4b08-86a3-ebed13badc7f/.default` have no roles. With either
  token, GET on the list, stagedApps and settings routes returns an empty HTTP 401.
- Pinned PowerShell connected with `-AccessTokens` fails each read locally with
  `KeyNotFoundException: The given key 'https://substrate.office.com' was not
  present in the dictionary`. No request is sent.
- Pinned PowerShell connected with a certificate (app-only) fails each read with
  `ApiException` ErrorCode 401 and an empty message.

The SDK therefore refuses application mode locally with
`client.ErrAuthenticationMode`, as for Multi-Geo region.
`TestAccUnifiedAppsApplicationIdentityRefused` checks that every method refuses
application mode, and that pinned PowerShell refuses the same reads (outcome
`local` or `401`).

## Unqualified

- **All successful reads and response shapes.** Unit fixtures are synthetic,
  built from the pinned model names. `TestAccUnifiedAppsDelegatedReadOnly` runs
  only with a private `MSTEAMS_LAB_APP_CATALOG_FILE` delegated scope-token
  fixture, which is pending in `tools/acceptance/pending.json`.
- **All writes.** The planned reversible `Update-M365TeamsApp` toggle needs
  delegated credentials, so it wasn't run and no recovery handler or mutation
  proof was added.
- **Write success statuses.** 200, 202 and 204 are accepted but none was observed.
- **Query encoding.** The SDK percent-encodes `$expand` as `%24expand`, while
  the module sends it literally. Server equivalence is unobserved.
