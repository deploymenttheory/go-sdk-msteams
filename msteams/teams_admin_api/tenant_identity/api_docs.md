# Tenant identity

Use `sdk.TeamsAdminAPI.TenantIdentity` to read the Teams tenant record, search
groups for voice application routing, and read the tenant-scoped hosting provider
and licensing configurations. The service also carries the support-only
directory reads and the licensing-impacting grace period removal. Operations use
the shared client/auth/transport controls through the ConfigAPI backend and need
no PowerShell runtime. Constructing the root client or this service performs no
network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `GetTenant(ctx)` | `Get-CsTenant` | GET `/Teams.Tenant/tenants?defaultpropertyset=Extended` | 200 |
| `SearchGroups(ctx, request)` | `Find-CsGroup` | GET `/Teams.VoiceApps/groups?SearchQuery=…[&MaxResults=&ExactMatchOnly=&MailEnabledOnly=]` | 200 |
| `GetOnlineDirectoryTenant(ctx)` | `Get-CsOnlineDirectoryTenant` | none: always `ErrOnlineDirectoryTenantDeprecated` | — |
| `ListHostingProviders(ctx)` | `Get-CsHostingProvider` | GET `/Skype.Policy/configurations/HostingProvider` | 200 |
| `ListHostingProvidersFiltered(ctx, pattern)` | `Get-CsHostingProvider -Filter` | same list, filtered locally | 200 |
| `GetHostingProviderByIdentity(ctx, identity)` | `Get-CsHostingProvider -Identity` | GET `/Skype.Policy/configurations/HostingProvider/configuration/{identity}` | 200 |
| `ListTenantLicensingConfigurations(ctx)` | `Get-CsTenantLicensingConfiguration` | GET `/Skype.Policy/configurations/TenantLicensingConfiguration` | 200 |
| `ListTenantLicensingConfigurationsFiltered(ctx, pattern)` | `Get-CsTenantLicensingConfiguration -Filter` | same list, filtered locally | 200 |
| `GetTenantLicensingConfigurationByIdentity(ctx, identity)` | `Get-CsTenantLicensingConfiguration -Identity` | GET `/Skype.Policy/configurations/TenantLicensingConfiguration/configuration/{identity}` | 200 |
| `GetAadTenant(ctx)` | `Get-CsAadTenant` | GET `/Teams.Tenant/AadTenant` | 200 |
| `GetAadUser(ctx, identity)` | `Get-CsAadUser` | GET `/Teams.User/AadUser/{identity}` | 200 |
| `GetCloudTenant(ctx)` | `Get-CsCloudTenant` | GET `/Teams.InternalSupport/GetAADTenantData` | 200 |
| `GetCloudUser(ctx, userID)` | `Get-CsCloudUser` | GET `/Teams.InternalSupport/GetAADUserData?userId=` | 200 |
| `RemoveUserLicenseGracePeriod(ctx, identity, request)` | `Remove-CsUserLicenseGracePeriod` | PATCH `/Teams.User/users/{identity}/AssignedPlans` | 204 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; the grace period removal is never
replayed automatically. Any other HTTP status is an `*client.APIError`; an
undeclared success status is `ErrResponse`. Invalid input fails with `ErrInput`
before discovery or credentials are used.

## Tenant record

`Tenant` types the fields the service returns for the extended property set,
including assigned and provisioned plans, verified and SIP domains, Teams
upgrade state and the provisioning/publishing timestamps (keyed by processor
name). Timestamps keep the service's text. `RawFields` keeps every property.
PowerShell's projection renames some of them (`objectId` → `TenantId`,
`sipDomains` → `SipDomain`, `assignedPlans` → `AssignedPlan`). The module's
`Filter`, `Identity` and `ResultSize` parameters fail parameter binding in 8.1.0,
so the SDK has no equivalent.

## Group search

`GroupSearchRequest.SearchQuery` is required; the service rejects queries
shorter than three characters with HTTP 400. Nil options are omitted. Booleans
are sent as `True`/`False`, the module's spelling. The response's `Diagnostics`
array goes to the module's diagnostic stream and is not returned. A missing or
null `Groups` array is an empty result.

## Configurations

Lists are ordered Global first, as in the module. `…Filtered` applies the
module's local wildcard match on `Identity` or `Tag:Identity`. No filter is sent
to the service, and blank input returns the full list. `Configuration` types
`Identity` only; `RawFields` keeps every setting. Get-CsHostingProvider's
`LocalStore` switch has no service behaviour.

## Support-only reads

`GetAadTenant`, `GetAadUser`, `GetCloudTenant` and `GetCloudUser` are exported
by the module only for support environments. Their output has no fixed model,
so `DirectoryObject` types `objectId` and `displayName` and keeps everything else
in `RawFields`. With app-only Teams Administrator credentials each returns
HTTP 403. The module sends these four routes with a trailing slash, which the
SDK's path model cannot express. Without the slash the service returns 403 to
app-only callers; with it, the Cloud routes return 400 "This legacy OCE cmdlet
is no more supported."

## Grace period removal

`RemoveUserLicenseGracePeriod` ends a user's license grace period immediately.
It changes licensing state and cannot be undone. `GracePeriodRemoval.Action` is
sent as `action` when non-nil. `Capabilities` is sent as `capabilities` when
non-nil; an empty non-nil slice sends `[]`.
