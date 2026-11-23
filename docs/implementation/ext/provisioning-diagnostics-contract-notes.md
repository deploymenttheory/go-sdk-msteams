# Provisioning diagnostics contracts

Owner: #89 (EXT-26). Baseline: MicrosoftTeams 8.1.0.

## Availability

Only `Move-CsInternalHelper` is exported in a TeamsGA session. CmdletConfig
limits the AutoRest exports (Get-CsBusinessVoiceDirectoryDiagnosticData,
Get/Set-CsUssUserSettings, Get-CsMasObjectChangelog,
Get-CsMasVersionedSchemaData, Invoke-CsCustomHandlerNgtprov,
Invoke-CsCustomHandlerCallBackNgtprov, Invoke-CsDirectObjectSync,
Invoke-CsMsodsSync, Invoke-CsRehomeuser, Set-CsTenantUserBackfill) to SelfHost
and Torus. Get-CsLoginInfo, Get-CsUserProvHistory, Invoke-CsDirectoryObjectSync
and Invoke-CsGenericNgtProvCommand are compiled cmdlets in
`Microsoft.Teams.Policy.Administration.Cmdlets.OCE.dll`, declared only in the
manifest's OCE block (lines 798–805).

Connected to the lab with app-only credentials, the pinned module reports
`CommandNotFoundException` for every SelfHost/Torus export, and
`Move-CsInternalHelper` fails with a `NullReferenceException` from
`ProcessCustomAttributesAtRuntime`, because its target
`Microsoft.Teams.ConfigAPI.Cmdlets.private\Move-CsInternalHelper` is not
loaded. No PowerShell oracle is therefore possible. The wire contract below
comes from IL, model serializers and a capture of the module's own HTTP
pipeline, made by importing the private assembly directly and invoking each
generated cmdlet against reserved targets (all denied).

## Generated ConfigAPI routes

Each uses the discovered forest in `X-MS-Target-Uri`. Header values follow the
repository convention of the exported command name.

| Command | Route | Body | Success |
| --- | --- | --- | --- |
| Get-CsBusinessVoiceDirectoryDiagnosticData | GET `Teams.Cpc/bvdDiagnosticData/{table}/{region}/{partitionKey}?rowKey&resultSize` | – | 200 `IBvdTableEntity[]` (`Properties[{Name,Type,Value}]`) |
| Get-CsMasObjectChangelog | GET `Teams.Tenant/GetMASObjectChangelog?Identity&SchemaNames&Versions` | – | 200 `IMasChangelogItem[]` |
| Get-CsMasVersionedSchemaData | GET `Teams.Tenant/GetSchemaDataFromMAS/?Identity&SchemaName&Versions` | – | 200 `IMasSchemaItem[]` (free-form; the wrapper emits each item's properties) |
| Get-CsUssUserSettings | GET `Teams.UserSettings/api/v1/tenant/{t}/user/{u}/setting?name` | – | 200 text (404 text is also written) |
| Set-CsUssUserSettings | PATCH `…/setting/{name}` | Body string | 200, 201 text (400 text is also written) |
| Invoke-CsCustomHandlerCallBackNgtprov | POST `Teams.Tenant/CustomHandlerCallback/?id&operation&eventname` | none | 200 free-form object |
| Invoke-CsCustomHandlerNgtprov | POST `Teams.Tenant/CustomHandler` | `{HandlerFullyQualifiedName,Payload}` | 200 `{id}` |
| Invoke-CsDirectObjectSync | POST `/Teams.InternalSupport/ProvisioningOperations/DsRequest` | `IDsRequestBody` | 200 none |
| Invoke-CsMsodsSync | POST `Teams.Tenant/MsodsSync/` | `IResyncRequestBody` | 200 none |
| Invoke-CsRehomeuser | POST `/Teams.User/RehomeUser/{userId}/moveUser` | Body string | 204 none |
| Set-CsTenantUserBackfill | POST `Teams.InternalSupport/SetBackfillToPubsub/?operation` | Body string | 200 `{message}` |

Serializers omit unbound parameters; switches bound `:$false` are sent as
`false`. The DsRequest/Resync bodies use `ReSyncOptions` (plural) and
`SynchronizeTenantWithAllObjects`; Invoke-CsMsodsSync sends its `ObjectId`
array as `ObjectIds`.

SDK decisions:

- Trailing slashes and the module's doubled leading slash are not sent. Raw
  probes of both spellings reached the same gated routes (403/401), while
  unknown routes on the same prefixes return 404.
- Get-CsUssUserSettings 404 and Set-CsUssUserSettings 400 bodies are module
  output, not errors. The SDK returns them as `*client.APIError` with the
  response body intact, keeping one failure model.
- Queries use the shared sorted, `+`-for-space encoding.
- String Body parameters must be JSON; the transport compacts them.

## Move-CsInternalHelper

The compiled `MoveHelper` switches on `ActionType` (Beginmove, Completemove,
Getpolicy, Rehomeuser, Rollback, Setmovedresourcedata, Validateuser) and runs
`Invoke-CsInternal<Action>` with `UserSipUri` and the action's parameters,
always binding the three Teams data switches, `MoveToCloud` and `Force`. Each
calls `POST /Teams.HostedMigrationService/{BeginMove,CompleteMove,GetPolicies,
RehomeUser,Rollback,SetMovedResourceData,ValidateUser}` with
`X-MS-Target-Uri: {forest}HostedMigration` and `X-MS-Tenant-Id`. Captured
bodies:

- ValidateUser: `{localDeploymentInfo:{PresenceFqdn,RegistrarFqdn,
  HostingProviderFqdn?,MajorVersion:int},teamsData:{MoveToTeams,
  CheckEnterpriseVoice,CheckCpc},userSipUri,moveToCloud,force,cmdletVersion}`
  → `{ValidateUserResult:{…DeploymentInfo}}`.
- BeginMove/CompleteMove: `{userSipUri,majorVersion:string}`; BeginMove →
  `{BeginMoveResult:{Datastr}}`.
- GetPolicies/Rollback: `{userSipUri}`; GetPolicies →
  `{GetPoliciesResult:[{Key,Value}]}`.
- RehomeUser: `{teamsData,userSipUri,moveToCloud}`.
- SetMovedResourceData: `{resourceData:{Datastr},teamsData,userSipUri,
  majorVersion}`.

All declare only 200. The SDK maps the four parameter sets to seven methods
(Getpolicy and Rollback resolve to the default Rehome set).

## OCE commands

`IC3DirectoryTorusCmdletBase` reads `CmdletSettings.json`: service suffixes
under `Teams.InternalSupport` on the environment's records origin, a Torus
user-impersonation token, `X-MS-Tenant-Id` (the `-Tenant` GUID) and a new
`X-MS-Correlation-Id`. The SDK uses the bound ConfigAPI origin with one
attempt, as the existing policy-assignment OCE methods do; the module's
`-Environment` selection is not offered.

- Get-CsUserProvHistory: GET `AdminStoreOperations/userHistory/{userId}/{region}`
  (NOAM/EMEA/APAC) → `UserHistory{UserId,TenantId,VersionHistory[{Version,
  VersionCreatedTime,OrderHistory[…]}]}`. `-Mode` only formats console output.
- Invoke-CsGenericNgtProvCommand: POST
  `ProvisioningOperations/runHandler/{region}?handler={h}?tenant={t}?target=
  {target}?serviceInstance={si}` with the `-Payload` text. The SDK sends the
  same decoded `handler` value percent-encoded.
- Invoke-CsDirectoryObjectSync: POST `ProvisioningOperations/DsRequest` with
  Newtonsoft defaults: declared property order, enum indexes, nulls written,
  `ObjectId` left as the zero GUID when `ObjectIds` is non-empty. It rejects
  `SynchronizeTenantWithAllObjects` unless `ObjectClass` is Company, and
  `ServiceInstance` must match `^MicrosoftCommunicationsOnline/.+$`.
- Get-CsLoginInfo acquires and returns the module's own ConfigAPI token and
  sends no service request. It has no SDK method: credentials come from
  `msteams/auth`, and returning tokens is outside the SDK's surface.

## Lab observations (client-secret, app-only Teams Administrator)

Every operation was sent through the SDK against reserved targets
(`00000000-0000-4000-8000-000000000026`, `sip:…@example.invalid`, invalid
operation and handler names, validation-only sync requests); the lab user was
only read. Each returned HTTP 403 `{"code":"Forbidden",…}`, except the two
custom handler routes, which return HTTP 401 `{"Message":"Authorization has
been denied for this request."}`.

Acceptance: `TestAccProvisioningDiagnosticsReadsDenied` and
`TestAccProvisioningDiagnosticsActionsDenied` (read-only steps) assert the
classification; their sanitized responses feed the unit contract tests. Success
bodies are covered by synthetic contract cases only, and every command remains
unqualified for successful use.
