# Team and user app installation contracts

Owner: EXT-90 (#191). Baseline: MicrosoftTeams 8.1.0.

`GraphAPI.AppInstallations` implements four nested preview candidates using the
verified `Microsoft.Teams.PowerShell.TeamsCmdlets.dll` (SHA256
`3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24`).
The root manifest comments these exports as `#preview`; observed runtime
parameter sets remain unknown. Implementing their evidenced operations does not
establish complete cmdlet conformance or authorize an exclusion.

The implementing classes are `AddTeamsAppInstallation`, `GetTeamsAppInstallation`,
`UpdateTeamsAppInstallation` and `RemoveTeamsAppInstallation` in namespace
`Microsoft.Teams.PowerShell.TeamsCmdlets`. Their shared `TeamsCmdletExtensions`
helpers `ValidateAndGetScope`, `CreateTeamsAppLifecycleManagementUri` and
`FetchTeamsAppInstallationsFromAppId` supply the target routes and filter.

| Cmdlet | SDK | Pinned request |
| --- | --- | --- |
| Add-TeamsAppInstallation | `Service.Install` | POST collection with catalog binding and optional permission list |
| Get-TeamsAppInstallation | `Service.List`, `Service.GetByID` | GET collection or installation, expanding teamsAppDefinition |
| Update-TeamsAppInstallation | `Service.UpgradeByID`, `Service.UpgradeByAppID` | POST installation/upgrade |
| Remove-TeamsAppInstallation | `Service.DeleteByID`, `Service.DeleteByAppID` | DELETE installation |

Collection paths are `/v1.0/teams/{teamId}/installedApps` and
`/v1.0/users/{userId}/teamwork/installedApps`. Targets require exactly one scope;
chat is explicitly unsupported by the pinned route helper. Installation IDs are
distinct from catalog app IDs. The catalog selector uses
`$filter=teamsAppDefinition/teamsAppId eq '{appId}'` and
`$expand=teamsAppDefinition`. The SDK doubles embedded quotes for OData.

The install body binds `teamsApp@odata.bind` to the bound cloud's Graph origin
and `/v1.0/appCatalogs/teamsApps/{appId}`. Install and upgrade preserve the pinned
`grantedResourceSpecificApplicationPermissions` property. Nil omits the property;
a nonnil empty slice explicitly sends `[]`. This legacy property is not silently
rewritten into the current Graph `consentedPermissionSet` model. Live acceptance
of the pinned permission body remains unqualified.

The SDK follows guarded opaque next links, retains partial list results and the
last response, and refuses an app-ID-based write unless a complete traversal
finds exactly one installation. It never automatically retries mutations.
Expanded definition JSON, nulls, empty arrays and future fields survive reads.

Credential-free tests in
[`service_test.go`](../../../msteams/graph_api/app_installation/service_test.go)
exercise both scopes, catalog versus installation identity, install/read/upgrade/
remove, explicit permission presence, unique/missing/duplicate resolution,
a denied continuation that prevents mutation, authorization denial and local
validation before token acquisition. They use synthetic replies and are not live
provider evidence.

The published [Graph resource](https://learn.microsoft.com/en-us/graph/api/resources/teamsappinstallation?view=graph-rest-1.0),
[team listing](https://learn.microsoft.com/en-us/graph/api/team-list-installedapps?view=graph-rest-1.0)
and [upgrade operation](https://learn.microsoft.com/en-us/graph/api/team-teamsappinstallation-upgrade?view=graph-rest-1.0)
independently confirm the resource identity, team route and upgrade protocol.
These current docs do not replace the pinned module's exact legacy request body.

No live installation is attempted: it changes a real team/user app surface and
requires an explicitly approved owned app plus suitable permissions. A fresh
read-only catalog probe with the bound CI identity returned HTTP 403. No role,
consent or license changes are included. Conditional runtime parameter-set
metadata, live body acceptance and independent installation cleanup remain open.
