# CORE-01 special profile source audit

Audited 2026-11-20 against the local pinned MicrosoftTeams 8.1.0 package.
The endpoint/audience values in the [profile contract](../../../contracts/endpoints/8.1.0/special-service-profiles.json)
and [capture script](../../../tools/contracts/capture-special-profiles.ps1) match
the methods below. Selector rejection must remain an **isolated helper result**;
it does not establish complete command or service availability in a cloud. In
particular, GCC region-command availability remains unqualified.

## Primary source identity

Package `/private/tmp/msteams-baseline-810/module.zip`, SHA256
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
The following package-relative assembly hashes were independently rechecked:

| Package path | SHA256 | Local decompilation |
| --- | --- | --- |
| `netcoreapp3.1/Microsoft.TeamsCmdlets.PowerShell.Connect.dll` | `c1c2b2d1c4a5306fb70f373df7ed5bfca4f9db58e90be358afd8a1d48ed46aea` | `/private/tmp/msteams-connect-decompiled` |
| `netcoreapp3.1/Microsoft.Teams.PowerShell.TeamsCmdlets.dll` | `3f9f5bfe36beb0b3d378f606db58c55183221eb5c1465889b106444423086d24` | `/private/tmp/msteams-core-decompiled` |

Connect source types below are `Microsoft.TeamsCmdlets.Powershell.Connect.Models.AzureEnvironment`,
`TeamsEnvironment`, `AzureRMProfile`, `AzureSessionProvider`, and
`Microsoft.TeamsCmdlets.Powershell.Connect.ConnectMicrosoftTeams`. Core source
types are in `Microsoft.Teams.PowerShell.TeamsCmdlets`. This audit reads those
decompiled methods and checked artifacts; it does not execute their methods,
authenticate, acquire tokens, contact a tenant, or qualify SDK implementation.

## GCC: supplied base and actual caller order

`TeamsEnvironment.GetAzureEnvironmentName` has no `TeamsGCC` branch and throws
`ArgumentException("Invalid TeamsEnvironmentName")` for that input. The capture
script catches the GCC `ArgumentException`, records `name_mapper_rejected=true`,
and explicitly supplies `AzureCloud`; `azure_environment_setup` identifies this
source-supplied base. Consequently the GCC `azure_environment` and base-map reads
are source-informed setup, **not a successful mapper observation**. Its targeting
and region selector calls and private base-URI call are separate isolated probes.

The ordinary omitted-name source path, using the raw map without an environment
override, explains this setup:

1. `ConnectMicrosoftTeams.AzureEnvironmentName` defaults to enum zero,
   `AzureEnvironment.EnvironmentName.AzureCloud`. `BeginProcessing` calls
   `setAzureEnvironment()`; `AzureRMProfile` populated that environment from
   `AzureEnvironment.PublicEnvironments`.
2. `ProcessRecord` calls `RMProfileClient.Login(..., AzureEnvironment, ...)`, then
   updates tenant information, then invokes `DetermineTeamsTargetingEndPoint`
   with the tenant domain when custom endpoints were not supplied.
3. `DetermineTeamsTargetingEndPoint` requests OIDC configuration under the current
   ActiveDirectory endpoint. `tenant_region_scope == "NA"` and
   `tenant_region_sub_scope == "GCC"` select
   `setTeamsTargetingEndPoint("TeamsGCC", AzureEnvironment)` and the GCC meeting
   content endpoint. The targeting setter changes the targeting request endpoint
   and static Teams name; it does not change the targeting token resource or
   either migration resource.

The caller expression is `DetermineTeamsTargetingEndPoint(...).GetAwaiter()`;
there is no `await` or `GetResult()`. This proves invocation after login, not that
the async detection has completed before Connect returns. The offline capture
does not reproduce login or OIDC discovery. It bypasses the override-capable
`PublicEnvironments` getter, clones endpoint dictionaries for setters, explicitly
sets `TeamsEnvironment.Name` for the private selector, and restores that name in
`finally`.

## Isolated selectors and route qualification

`TeamsEnvironment.setTeamsTargetingEndPoint` accepts `TeamsCloud` and `TeamsGCC`,
choosing the commercial and GCC retail origins respectively. It explicitly
throws for `TeamsGCCH`, `TeamsDOD`, `TeamsChina`, `TeamsFrance`, and `TeamsGermany`.
Only the request endpoint is assigned; the GCC AzureCloud token audience remains
`https://retailservices.teams.microsoft.com`.

`TeamsEnvironment.setMigrationEndPoint` accepts only `TeamsCloud`; every other
name throws. Its sole Connect call is in `BeginProcessing`, under explicit
`TeamsEnvironmentName != null` and `ModulePreReleaseVersion != null` guards,
before login/detection. The latter guard tests null, not empty text; the pinned
`MicrosoftTeams.psd1` contains `Prerelease = ''`.

For omitted-name GCC login, AzureCloud already contains the commercial migration
origin and token resource; subsequent GCC detection leaves both intact.
`RMProfileClient.Login` retains that environment in `AzureContext`, and
`AzureSessionProvider` retains it for resource lookup.
`GetMultiGeoRegion.ProcessRecord` reads these resources directly and never calls
the migration selector. Thus the fixture's rejected GCC helper call must not be
presented as a GCC command denial. The contract's GCC region route describes this
inherited tuple while command/service availability remains unqualified. Its
`isolated_selector_rejects` and the capture's `origin_after_attempt` correctly keep
helper outcomes separate from route and availability claims.

## Request origin, token resource and exact scope

`AzureEnvironment.environments` supplies the following commercial map entries.
`TeamsCmdlet.PerformAuthorization` reads the token resource, builds the exact
scope, and supplies both to `GetAccessToken` and `AuthorizeRequest`. The hierarchy
commands select the v1/v2 token key in their `ProcessRecord` methods;
`GetMultiGeoRegion.ProcessRecord` separately reads the migration request key.

| Profile | Request origin | Token resource | Scope |
| --- | --- | --- | --- |
| Targeting v1 | `https://retailservices.teams.microsoft.com` | `https://retailservices.teams.microsoft.com` | resource + `/.default` |
| Hierarchy v2 | `https://hierarchy.teams.microsoft.com` | `https://chatsvcagg.teams.microsoft.com` | resource + `/.default` |
| Regions | `https://teams.cloud.microsoft` | `https://migrationsvc.teams.microsoft.com` | resource + `/Region.Read.All` |

Region scope is exactly
`https://migrationsvc.teams.microsoft.com/Region.Read.All`; replacing it with
`.default` would change the evidenced request. There is no hierarchy-v2 cloud
selector here. The commercial entries, including their presence in the supplied
GCC base, do not qualify deployment or application access in any cloud.

## Private-channel migration

`GetTenantPrivateChannelMigrationStatus.ProcessRecord` first authorizes against
`TeamsAndChannelsServiceResourceId`, then appends
`/channels/privateChannelMigrationStatus` to `GetTemplatesServiceApiBaseUrl()`.
`TeamsCmdlet.PerformAuthorization` uses that resource plus `/.default`.

| Teams name | Pure base-URI result | Raw mapped token resource |
| --- | --- | --- |
| TeamsCloud | `https://teams.microsoft.com/fabric/templates/api` | `https://templates.teams.microsoft.com` |
| TeamsGCC | `https://teams.microsoft.com/fabric/gcc/templates/api` | `https://templates.teams.microsoft.com` from explicitly supplied AzureCloud |
| TeamsGCCH | `https://gov.teams.microsoft.us/fabric/gov/templates/api` | `https://templates.teams.microsoft.us` |
| TeamsDOD | `https://dod.teams.microsoft.us/fabric/dod/templates/api` | `https://templates.teams.microsoft.us` |
| TeamsChina | `https://teams.microsoftonline.cn/fabric/templates/api` | `https://templates.teams.microsoftonline.cn` |
| TeamsFrance / TeamsGermany | `https://teams.microsoft.com/fabric/templates/api` via default branch | absent from AzureFranceGovCloud / AzureGermanyGovCloud |

`AzureEnvironment.GetEndpoint` returns null for absent keys;
`AzureEnvironment.GetResource` throws for missing/empty resources, as delegated by
`AzureSessionProvider.GetResource`. With those raw France/Germany maps, command
authorization therefore fails before the HTTP step. Their pure commercial URI
fallback establishes neither a usable token profile nor cloud support. The
remaining rows likewise describe source selection only, not live availability.
