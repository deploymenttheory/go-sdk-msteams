# Configuration diagnostics contracts

Owner: #86 (EXT-23). Baseline: MicrosoftTeams 8.1.0.

## Command availability

None of the seven commands loads in a tenant session (`Get-Command` finds
none of them after `Connect-MicrosoftTeams`). The manifest declares
Get/New/Set/Remove-CsConfiguration and Set-CsPublishPolicySchemaDefaults on
its OCE (support) export lines. CmdletConfig exports Clear-CsCacheOperation
only to SelfHost and Torus. The module still loads the implementations:

- Get/New/Set/Remove-CsConfiguration are AutoRest proxies in
  `internal/Merged_internal.ps1` (module
  `Microsoft.Teams.ConfigAPI.Cmdlets.internal`). The custom
  `*-CsConfigurationModern` wrappers call them with `-ConfigType`, the
  `PropertyBag` hashtable as `-Body`, and `PropertyBag['Identity']` as
  `-ConfigName`.
- Clear-CsCacheOperation is an AutoRest proxy in the
  `Microsoft.Teams.ConfigAPI.Cmdlets` script module.
- Set-CsPublishPolicySchemaDefaults is the compiled
  `Microsoft.Teams.Policy.Administration.Cmdlets.OCE.PublishPolicySchemaDefaults`.
- Set-CsNotifyCache has no implementation. No cmdlet class or script function
  defines it. `CommonHttpHelperRequest.SendNotifyCacheControllerRequests`
  (POST `/Skype.Policy/oce/refreshcachenotification` with a caller-built JSON
  body) has no caller, so its body is unknown. It is not implemented.

The lab oracle runs the internal commands inside their own module scope
(`& (Get-Module -All Microsoft.Teams.ConfigAPI.Cmdlets.internal) { … }`).

## Wire contracts

The generated `ConfigApiBasedCmdlets` methods give the following. The IL also
gives each status dispatch and each model serializer.

| Command | Method and path | Query | Body | Success |
| --- | --- | --- | --- | --- |
| Get-CsConfiguration (Get) | GET `/Skype.Policy/configurations/{configType}` | `apiVersion`, `schemaVersion` when bound | none | 200, body written as a JSON string |
| Get-CsConfiguration (Get1) | GET `…/{configType}/configuration/{configName}` | same | none | 200 |
| New-CsConfiguration | POST `/Skype.Policy/configurations/{configType}` | same | `XdsConfiguration` | 201, no body |
| Set-CsConfiguration | PUT `…/{configType}/configuration/{configName}` | same | `XdsConfiguration` | 204 |
| Remove-CsConfiguration | DELETE `…/{configType}/configuration/{configName}` | none | none | 204 |
| Clear-CsCacheOperation | POST `/Teams.InternalSupport/CacheClear` | none | `CacheClearRequest` | 200 and 206, both `CacheClearResponse[]` |

- The ViaIdentity parameter sets parse the same paths out of `InputObject`.
  The SDK takes the parts directly.
- `XdsConfiguration` serializes `Identity` plus its additional properties.
- `CacheClearRequest` serializes `KeysToDelete` (string array; the cmdlet
  allows an empty one) and `Region`.
- `CacheClearResponse` holds `FailedKeys` and `Region`.
- Update-CsConfiguration (PATCH on the same item path) is internal-only and
  not in the EXT-23 ledger.

Set-CsPublishPolicySchemaDefaults checks that `-PolicyType` is not empty. It
then calls `SendSchemaDefaultsControllerRequests`: POST
`https://{ConfigApiEndpoint host}/Skype.Policy/schemadefaults/{PolicyType}`
with no content and no forest discovery. The call adds `X-MS-CmdletName`,
`X-MS-Correlation-Id` and a User-Agent, plus OCE role/forest headers in
support hosts. Any 2xx status returns the response text, which the cmdlet
writes; any other status throws. The SDK sends it through the shared direct
policy transport, which also adds `MPACmdlet: true`.

## Lab observations

All observations used client-secret, app-only Teams Administrator credentials.

- Reads of `TeamsFeedbackPolicy` and `TeamsClientConfiguration` (list,
  `Global`, with and without `apiVersion`/`schemaVersion`) return 200 JSON.
- An unknown type returns 400 `ClientError`, errorCode 40001.
- A missing instance returns 404 `ClientError`, errorCode 40407.
- The pinned internal Get-CsConfiguration returns the same counts and values,
  and fails the missing read with `ClientError`.
- An owned `TeamsFeedbackPolicy` instance can be created (201), read back by
  both `name` and `Tag:name`, replaced with PUT (204, non-default values read
  back), and removed (204, then 404).
- Clear-CsCacheOperation was sent only with a nonexistent key for a
  nonexistent region; it returns 403 `Forbidden`. The pinned module fails the
  same way (`Forbidden,…ClearCsCacheOperation_ClearExpanded`). No cache was
  flushed.
- Set-CsPublishPolicySchemaDefaults was sent only for an undefined policy
  type; it returns 403 `Forbidden`. The compiled cmdlet needs an OCE login
  context, so it has no PowerShell parity run.

## Acceptance

- `TestAccConfigurationDiagnosticsReads` (read-only) covers the reads, with
  PowerShell parity.
- `TestAccConfigurationDiagnosticsSupportCommandsDenied` (read-only, serial)
  covers both support-command denials, with cache clear parity.
- `TestAccConfigurationLifecycle` (mutation, journaled) covers the owned
  instance through create/read-back, PowerShell parity, set/read-back and
  remove/absence. Recovery kind `generic-configuration` is registered.

Sanitized responses feed the unit contract tests.

## Unqualified

- Successful ClearCache (200/206 bodies) and PublishPolicySchemaDefaults
  calls are covered only by synthetic contract cases. They need support
  credentials, and they act service-wide.
- Set-CsNotifyCache is not implemented (see above).
