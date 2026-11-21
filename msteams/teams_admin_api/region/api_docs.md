# Multi-Geo region observations

`sdk.TeamsAdminAPI.Regions.GetForUser` and `GetForGroup` perform one native
observation read for an explicitly supplied ID, escaped as one path segment.
No Graph identity lookup, regional mutation, list or inferred relocation workflow
is added. The response retains Region, nullable PreferredDataLocation and original
timestamp text, plus every raw future/null/omitted field. Unknown region values
remain readable. The native object is consistent even when PowerShell displays
only Region for an observation without optional properties.

The fixed profile uses teams.cloud.microsoft with migrationsvc.teams.microsoft.com
as audience and the exact delegated scope
`https://migrationsvc.teams.microsoft.com/Region.Read.All`. Token acquisition,
coalescing, cache and rejected-token refresh preserve that scope. Region does not
use Graph tokens, ConfigAPI discovery/header or resource /.default. Application
mode is documented unsupported and fails locally before credential acquisition.
The evidenced commercial/GCC maps are enabled; other clouds fail locally without
fallback. That local restriction is not a claim of remote service denial.

Supplied-token callers must configure the scope-specific inventory key returned
by `binding.RequestScope(audience, scope).TokenKey()`. An audience-only supplied
token is never silently reused for an explicit scope. Injected providers receive
TokenRequest.Scope and must preserve it in their own credential/cache behavior.
Native delegated providers request exactly one resource scope and retain the
selected account across resources. Opaque tokens are not decoded to infer consent.

The SDK retains successful HTTP/model metadata and typed HTTP failures. No live
mode/consent/region/tenant behavior is qualified by offline tests. See
[unit tests](service_test.go)
and [unit tests](service_test.go).
