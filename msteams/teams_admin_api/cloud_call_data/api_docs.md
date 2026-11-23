# Cloud call data connection

Use `sdk.TeamsAdminAPI.CloudCallData` for the tenant's cloud call data
connection, which lets on-premises Skype for Business Server deployments upload
call data to the cloud. Operations use the shared client/auth/transport controls
through the ConfigAPI backend and need no PowerShell runtime. Constructing the
root client or this service performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `Get(ctx)` | `Get-CsCloudCallDataConnection` | GET `/Teams.InternalSupport/CloudCallDataConnection` | 200 |
| `Create(ctx)` | `New-CsCloudCallDataConnection` | POST `/Teams.InternalSupport/CloudCallDataConnection` (no body) | 200, 201 |

Both cmdlets take no parameters. Every method takes a context and returns the
actual Resty response alongside its value/error. `Get` is replay-safe; `Create`
is never replayed automatically. Any other HTTP status is an
`*client.APIError`; an undeclared success status is `ErrResponse`.

## Token

`Connection.Token` is the credential the on-premises deployment uses. Treat it
as a secret: `Connection` formats as `cloud call data connection (redacted)`
with `%v`, `%+v` and `%#v`. `Token` is nil when the service sends null or omits
it, and `RawFields` keeps every service property.

## Restrictions

- App-only Teams Administrator credentials receive HTTP 403 `Forbidden` on
  both `Get` and `Create` (observed live), and the tenant is unchanged;
  `errors.Is(err, auth.ErrForbidden)` reports it.
- The connection is a tenant singleton and the module has no cmdlet to remove
  it, so `Create` is a one-way operation.
