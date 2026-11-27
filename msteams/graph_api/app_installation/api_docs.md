# Team and user app installations

Use `sdk.GraphAPI.AppInstallations` for apps installed in a team or a user's
personal scope. This resource changes installations of an existing catalog app;
`GraphAPI.AppCatalog` publishes the catalog app itself. Constructing either
resource does not authenticate or contact a service.

Set exactly one of `Target.TeamID` or `Target.UserID`. The corresponding collection
is `/v1.0/teams/{id}/installedApps` or `/v1.0/users/{id}/teamwork/installedApps`.
Chat scope is not part of this pinned command contract.

| Method | Pinned conditional command | Request | Success |
| --- | --- | --- | --- |
| `Create(ctx, target, CreateRequest)` | Add-TeamsAppInstallation | POST collection | Team: 200 or 201; user: 201; empty body |
| `List(ctx, target, ListOptions)` / `NewPager` | Get-TeamsAppInstallation | GET collection, expanding teamsAppDefinition | 200 |
| `GetByID(ctx, target, installationID)` | Get-TeamsAppInstallation | GET installation, expanding teamsAppDefinition | 200 |
| `UpgradeByID(ctx, target, installationID, *PermissionSet)` | Update-TeamsAppInstallation | POST installation/upgrade | 204 |
| `UpgradeByCatalogID(ctx, target, catalogID, *PermissionSet)` | Update-TeamsAppInstallation | Resolve catalog ID, then upgrade | 204 |
| `DeleteByID(ctx, target, installationID)` | Remove-TeamsAppInstallation | DELETE installation | 204 |
| `DeleteByCatalogID(ctx, target, catalogID)` | Remove-TeamsAppInstallation | Resolve catalog ID, then delete | 204 |

`Create` returns an HTTP receipt, not an installation object. Its catalog ID is
not an installation ID; reconcile with `ListOptions.CatalogID` after creation.
The resource has no PATCH operation: the pinned update command upgrades an app.
Mutations use `ReplayNever`; a lost response is ambiguous and must be reconciled
before retrying.

`ListOptions.CatalogID` filters `teamsAppDefinition/teamsAppId`. `List` follows
opaque next links under the bound Graph destination policy and returns partial
items plus the last response on a failed later page. `ByCatalogID` methods
require a complete traversal with exactly one result; `ErrNotFound`,
`ErrAmbiguous` or any traversal error prevents the write.

`CreateRequest.CatalogID` produces `teamsApp@odata.bind` using the bound cloud's
Graph origin, including explicit caller endpoint overrides. `CreateRequest.Consent`
and upgrade's `*PermissionSet` use the published v1.0 `consentedPermissionSet`
object containing typed `resourceSpecificPermissions` entries. Nil omits consent;
a nonnil empty collection sends `[]`. Each request permission has an explicit
`application` or `delegated` type and value. Known expanded definition properties
are typed; raw snapshots retain unknown, null and empty fields.

The pinned preview module instead emits a legacy property on its v1.0 route.
The SDK does not silently substitute that legacy request or assume provider
wire equivalence. Callers must explicitly choose native v1.0 consent after
checking the app definition; no requested consent is discarded.

Catalog-ID mutations return `CatalogMutationResult` and an error. Lookup and
action responses, complete traversal, the resolved installation ID and whether
the action method was invoked remain available on failure or cancellation.
`ActionInvoked` does not prove the service received a lost request; reconcile
that installation ID before retrying an ambiguous mutation.

Installation permissions differ by team/user scope and whether resource-specific
consent is requested. See the published [team installation permissions](https://learn.microsoft.com/en-us/graph/api/team-post-installedapps?view=graph-rest-1.0)
and [user installation permissions](https://learn.microsoft.com/en-us/graph/api/userteamwork-post-installedapps?view=graph-rest-1.0).
The SDK does not grant permissions, consent or buy licenses. An authorization
failure remains available through `errors.Is(err, auth.ErrForbidden)` and the
returned response. A denial is not qualification of successful installation.

The pinned root module comments these four cmdlets as preview exports. Separate
compiled-type declaration evidence captures real `TeamScope` and `UserScope`
sets; it does not replace the immutable default-session discovery snapshot.
See [unit tests](service_test.go).
Live installation, current acceptance of the legacy permission body and independent
installation cleanup remain unqualified; no live install runs in CI without an
explicitly approved owned catalog app and target.
