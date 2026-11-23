# App catalog

Use `sdk.GraphAPI.AppCatalog` to list, publish, update and remove apps in the
tenant's Teams app catalog through Graph `/v1.0/appCatalogs/teamsApps`, as the
MicrosoftTeams module's compiled TeamsApp cmdlets do. Constructing the root
client or this service performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `List(ctx, ListOptions)` / `NewPager` | `Get-TeamsApp` | GET `/v1.0/appCatalogs/teamsApps[?$filter=…]`, following `@odata.nextLink` | 200 |
| `Create(ctx, CreateRequest)` | `New-TeamsApp` | POST `/v1.0/appCatalogs/teamsApps`, `application/zip` body | 201, 200 |
| `UpdateByID(ctx, id, pkg)` | `Set-TeamsApp` | PUT `/v1.0/appCatalogs/teamsApps/{id}`, `application/zip` body | 200, 201, 204 |
| `DeleteByID(ctx, id)` | `Remove-TeamsApp` | DELETE `/v1.0/appCatalogs/teamsApps/{id}` | 204 |

Graph requires `AppCatalog.Read.All` (list) or `AppCatalog.ReadWrite.All`
(writes); without them every call returns HTTP 403 `Forbidden`, which
`errors.Is(err, auth.ErrForbidden)` reports.

## Filters

`ListOptions` mirrors Get-TeamsApp's `-Id`, `-ExternalId`, `-DisplayName` and
`-DistributionMethod`. Each non-blank value adds `<property> eq '<value>'`, and
clauses are joined with ` and ` in that order. The SDK doubles single quotes;
the module sends them unescaped, which Graph rejects with HTTP 400.

## Packages

`Create` and `UpdateByID` take the app package bytes and send them unchanged as
`application/zip`. `Create` accepts only the `organization` distribution method
(any letter case) and fails locally with `ErrInput` otherwise, as the module
does. Uploads are never retried; after a lost response, find the app with
`List` before trying again. `UpdateByID` and `DeleteByID` reject a blank ID or
one containing `/` or `\`.
