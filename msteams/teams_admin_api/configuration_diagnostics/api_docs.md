# Configuration diagnostics

Use `sdk.TeamsAdminAPI.ConfigurationDiagnostics` to read and write any
configuration or policy type through the generic configuration route, and to
send the cache clear and schema-defaults publish support commands. Operations
use the shared client/auth/transport controls and need no PowerShell runtime.
Constructing the root client or this service performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `ListConfigurations(ctx, configType, options)` | `Get-CsConfiguration -ConfigType` | GET `/Skype.Policy/configurations/{configType}` | 200 |
| `GetConfiguration(ctx, configType, name, options)` | `Get-CsConfiguration -ConfigType -ConfigName` | GET `/Skype.Policy/configurations/{configType}/configuration/{name}` | 200 |
| `CreateConfiguration(ctx, configType, input, options)` | `New-CsConfiguration` | POST `/Skype.Policy/configurations/{configType}` | 201 (no body) |
| `SetConfiguration(ctx, configType, name, input, options)` | `Set-CsConfiguration` | PUT `/Skype.Policy/configurations/{configType}/configuration/{name}` | 204 |
| `RemoveConfiguration(ctx, configType, name)` | `Remove-CsConfiguration` | DELETE `/Skype.Policy/configurations/{configType}/configuration/{name}` | 204 |
| `ClearCache(ctx, input)` | `Clear-CsCacheOperation` | POST `/Teams.InternalSupport/CacheClear` | 200, 206 |
| `PublishPolicySchemaDefaults(ctx, policyType)` | `Set-CsPublishPolicySchemaDefaults` | POST `/Skype.Policy/schemadefaults/{policyType}` (no body, no forest discovery) | any 2xx |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; writes and support commands are never
replayed automatically. Any other HTTP status is an `*client.APIError`; an
undeclared success status is `ErrResponse`. Invalid input (a blank, `.`, `..`,
slash- or control-character path part; an empty version option; an
`Identity` key in `Settings`) fails with `ErrInput` before any I/O.

`Set-CsNotifyCache` is listed in the module manifest but no command in the
module implements it, so this package has no equivalent.

## Configuration instances

`Configuration` types only `Identity` and `ConfigId` because the schema
depends on the configuration type. `RawFields` keeps every property the
service sent. Instance names are `Global`, `Tag:Name` or the bare tag name;
the service accepts both tag forms in the path.

`ReadOptions` and `WriteOptions` set the optional `apiVersion` and
`schemaVersion` query values; nil fields are omitted.

`ConfigurationRequest` is the request body: `Identity` plus the type's own
properties in `Settings`. `SetConfiguration` is an HTTP PUT that replaces the
instance, so send every property you want to keep. The module's
`Set-CsConfiguration` wrapper defaults the instance to `Global`; in Go the
name is always explicit.

The module's wrapper moves `Global` to the front of a list and offers a
`-Filter` wildcard over identities; `ListConfigurations` returns the service
order and leaves filtering to the caller.

## Support commands

`ClearCache` removes cache keys service-wide. `KeysToDelete` is always sent
(nil as `[]`); `Region` is omitted when nil. A 206 response lists the keys that
could not be removed per region in `CacheClearResult.FailedKeys`.

`PublishPolicySchemaDefaults` publishes one policy type's schema defaults and
returns the response text unchanged.

## Restrictions

- The module exports these commands only in support (OCE) sessions; with
  tenant credentials the configuration route itself works.
- App-only Teams Administrator credentials receive HTTP 403 `Forbidden` from
  `ClearCache` and `PublishPolicySchemaDefaults`;
  `errors.Is(err, auth.ErrForbidden)` reports it.
- An unknown configuration type is HTTP 400 (`ClientError`, service error code
  40001); a missing instance is HTTP 404 (`ClientError`, 40407).
