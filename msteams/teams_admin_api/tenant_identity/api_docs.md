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
array is not mixed into the native group slice; the original HTTP response is
retained for inspection. PowerShell emits Info/default diagnostic text on the
success stream and Warning/Verbose separately; diagnostics can precede a
later group-projection failure. Missing, null or wrong-type
`Groups` in an object yields an empty slice; an empty/whitespace/array generated
root yields a nil slice with its receipt. A null/nonmodel row or failed public
ID projection fails the buffered result and retains that receipt.

The source path is public wrapper → internal proxy → private
`Find-CsGroup_Find`; the request header uses that private default name. `Force`
is removed by the wrapper, with no native option. There is no WhatIf/Confirm
surface for this read. Query whitespace/control characters are accepted and
encoded. Only an empty required query is rejected locally; the three-character
minimum is the server's observed validation, not an invented client limit.

`MaxResults` remains `*uint32` like the public parameter, but a value above
MaxInt32 fails its actual forwarded Int32 binder. The SDK reports this as an
early `ErrInput` before I/O. Nil means omitted. The actual internal script proxy
converts an explicitly
bound PowerShell null to bound zero before reaching the compiled cmdlet; use a
non-nil zero pointer for that wire result. Direct compiled Int32 binding has a
different null outcome and is not substituted for the script path. Explicit
zero and false options are sent.

`Group.ID` stays `*string` for API compatibility, containing the public
GroupModel's canonical D-format GUID text. The actual `Guid.Parse` projection
accepts N/D/B/P/X forms, canonicalizes them, and rejects missing/null/invalid IDs.
The bounded source packet pins the actual runtime parser, including X numeric
component quirks. `DisplayName` uses exact generated case/type semantics;
wrong-type values project nil. Exact duplicate JSON keys fail recursively.
`RawFields` retains the original wire values independently, including the
original ID spelling, without changing the typed projected value. This changes
only the group model/search path, not other tenant/support DTOs. Standard JSON
and native context/replay/error handling remain explicit SDK boundaries.

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

## Full tenant wire view (unreleased model correction)

`Tenant` types all 35 pinned wire fields plus the existing native enrichments.
`CompanyPartnerships` and `SyncInLyncADInfo` have typed known fields and original
`RawFields`. `SIPDomains`, `CompanyTags` and `Pools` contain nullable string
entries. `LastProvisionTimeStamps` and `LastPublishTimeStamps` now use
`map[string]json.RawMessage`, preserving arbitrary JSON values and numeric
lexemes; this changes the earlier string-only map API. Existing consumers must
decode individual values according to their actual type.

Timestamp fields retain the service's original text, and dictionary values retain
their JSON types and numeric lexemes. Consumers can decode the values they need
using ordinary Go types. The SDK does not emulate CLR presentation or mutate
values merely to reproduce PowerShell formatting. Unit tests cover the native
wire view, malformed replies and retained HTTP receipts.

## Default Modern tenant request identity

`GetTenant` sends `X-MS-CmdletName: Get-CsTenantObou_Get` for the Extended
ConfigAPI read. The public wrapper selects Get-CsTenantPoint for its default
Modern route; that helper invokes the internal proxy, whose private cmdlet name
passes through its invocation information to the final header writer. This
header is source evidence, independently of historical response captures.
See the [unit tests](service_test.go).
The SDK keeps its own user-agent and shared audience/routing controls.

The returned native wire view and actual HTTP receipt remain available
separately. Server-side availability and permissions still require tenant tests.
