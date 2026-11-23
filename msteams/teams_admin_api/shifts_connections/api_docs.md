# Shifts connections

Use `sdk.TeamsAdminAPI.ShiftsConnections` for the Teams Shifts connector:
connections to workforce management (WFM) systems such as Blue Yonder or UKG
Dimensions, and the connector instances that synchronize a connection with
Shifts. Operations use the shared client/auth/transport controls through the
ConfigAPI backend and need no PowerShell runtime. Constructing the root client
or this service performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `ListConnectors(ctx)` | `Get-CsTeamsShiftsConnectionConnector` | GET `/Teams.ShiftsConnector/api/v2/connectors` | 200 |
| `ListConnections(ctx)` | `Get-CsTeamsShiftsConnection` | GET `/Teams.ShiftsConnector/api/v2/connections` | 200 |
| `GetConnectionByID(ctx, id)` | `Get-CsTeamsShiftsConnection -ConnectionId` | GET `…/v2/connections/{id}` | 200 |
| `CreateConnection(ctx, request)` | `New-CsTeamsShiftsConnection` | POST `…/v2/connections` | 200, 201 |
| `ReplaceConnectionByID(ctx, id, update)` | `Set-CsTeamsShiftsConnection` | PUT `…/v2/connections/{id}` | 200 |
| `UpdateConnectionByID(ctx, id, update)` | `Update-CsTeamsShiftsConnection` | PATCH `…/v2/connections/{id}` | 200 |
| `DeleteConnectionByID(ctx, id)` | `Remove-CsTeamsShiftsConnection` | DELETE `…/v2/connections/{id}` | 200, 204 |
| `Validate(ctx, request)` | `Test-CsTeamsShiftsConnectionValidate` | POST `/Teams.ShiftsConnector/api/v1/connect` | 200, 204 |
| `ListInstances(ctx)` | `Get-CsTeamsShiftsConnectionInstance` | GET `…/v2/connectorinstances` | 200 |
| `GetInstanceByID(ctx, id)` | `Get-CsTeamsShiftsConnectionInstance -ConnectorInstanceId` | GET `…/v2/connectorinstances/{id}` | 200 |
| `CreateInstance(ctx, request)` | `New-CsTeamsShiftsConnectionInstance` | POST `…/v2/connectorinstances` | 200, 201 |
| `ReplaceInstanceByID(ctx, id, update)` | `Set-CsTeamsShiftsConnectionInstance` | PUT `…/v2/connectorinstances/{id}` | 200 |
| `UpdateInstanceByID(ctx, id, update)` | `Update-CsTeamsShiftsConnectionInstance` | PATCH `…/v2/connectorinstances/{id}` | 200 |
| `DeleteInstanceByID(ctx, id)` | `Remove-CsTeamsShiftsConnectionInstance` | DELETE `…/v2/connectorinstances/{id}` | 200, 204 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; writes and `Validate` are never replayed
automatically. Any other HTTP status is an `*client.APIError`; an undeclared
success status is `ErrResponse`. Invalid input fails with `ErrInput` before
discovery or credentials are used.

## Set versus Update, and etags

`Replace…ByID` (PUT, the module's Set) and `Update…ByID` (PATCH) send the same
body shape; nil fields are omitted. `Etag` is sent as the body `_etag`;
`IfMatch` is the `If-Match` header. The module makes `If-Match` optional for
connections and required for instances, and the SDK does the same.

## Secrets

`Settings` holds connector-specific settings, for example UKG Dimensions
`apiUrl`, `ssoUrl`, `appKey`, `clientId` and `clientSecret`, or Blue Yonder
URLs with `loginUserName` and `loginPwd`. They are sent as given. The request
types that carry them (`CreateConnectionRequest`, `ConnectionUpdate`,
`ValidateRequest`) print as a fixed redacted string with every fmt verb. The
service returns settings without secrets; `Connection.ConnectorSpecificSettings`
keeps them as raw JSON. `Validate` asks the service to sign in to the WFM
system with the settings and does not save them.

The module's optional `-Authorization` parameter adds a raw `Authorization`
header to the request. The SDK authenticates every request through its
configured credential and has no equivalent.

## Service restrictions

The Shifts connector service denies app-only Teams Administrator credentials:
every read and write returns HTTP 403 `Forbidden` ("Access Denied."), before
any validation. PowerShell returns that body as an output object rather than
an error. Successful response shapes are modelled from the module's
serializers and are not yet live-qualified.
