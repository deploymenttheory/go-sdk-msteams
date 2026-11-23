# Resource accounts

Use `sdk.TeamsAdminAPI.ResourceAccounts` for Teams resource accounts
(application instances): the accounts that give auto attendants and call queues
their identity, and their associations with those configurations. Operations
use the shared client/auth/transport controls through the ConfigAPI backend and
need no PowerShell runtime. Constructing the root client or this service
performs no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `List(ctx, options)` | `Get-CsOnlineApplicationInstance` | GET `/Teams.PlatformService/v2/Legacy/ApplicationInstances[?resultSize=&skip=]` | 200 |
| `GetByID(ctx, identity)` | `Get-CsOnlineApplicationInstance -Identity` | GET `/Teams.PlatformService/v2/Legacy/ApplicationInstances?identity=` | 200 |
| `Search(ctx, request)` | `Find-CsOnlineApplicationInstance` | GET `/Teams.VoiceApps/applicationinstances?SearchQuery=…` | 200 |
| `Create(ctx, request)` | `New-CsOnlineApplicationInstance` | POST `/Teams.PlatformService/v2/ApplicationInstances` | 200, 201 |
| `UpdateByID(ctx, identity, request)` | `Set-CsOnlineApplicationInstance` | PATCH `/Teams.PlatformService/v2/ApplicationInstances/{identity}` | 200, 204 |
| `UpdateTelephoneNumberByID(ctx, identity, number)` | `Set-CsOnlineVoiceApplicationInstance` | PATCH `/Teams.PlatformService/v2/ApplicationInstances/{identity}/telephoneNumber` | 200, 204 |
| `SyncByID(ctx, objectID, request)` | `Sync-CsOnlineApplicationInstance` | POST `/Teams.PlatformService/v2/ApplicationInstances/{objectId}/sync[?applicationId=&acsResourceId=]` | 200, 204 |
| `GetAssociationByID(ctx, identity)` | `Get-CsOnlineApplicationInstanceAssociation` | GET `/Teams.VoiceApps/applicationinstanceassociations/{identity}` | 200 |
| `GetAssociationStatusByID(ctx, identity)` | `Get-CsOnlineApplicationInstanceAssociationStatus` | GET `/Teams.VoiceApps/applicationinstanceassociations/{identity}/status` | 200 |
| `CreateAssociations(ctx, request)` | `New-CsOnlineApplicationInstanceAssociation` | POST `/Teams.VoiceApps/applicationinstanceassociations` | 200 |
| `DeleteAssociationByID(ctx, identity)` | `Remove-CsOnlineApplicationInstanceAssociation` (once per identity) | DELETE `/Teams.VoiceApps/applicationinstanceassociations/{identity}` | 200 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; writes are never replayed automatically.
Any other HTTP status is an `*client.APIError`; an undeclared success status is
`ErrResponse`. Invalid input fails with `ErrInput` before discovery or
credentials are used.

## Identities and GUIDs

`identity` is an object ID or user principal name. Association get and remove
apply the module's `EncodeSipUri` rewrite (`.` → `[DOT]`, `:` → `[COLON]`) before
the identity enters the path; association status does not, as in the module.
Parameters the module types as `System.Guid` (sync object ID, application and
ACS resource IDs) must be GUIDs and are sent in lowercase hyphenated form.

## Optional fields

Nil request fields are omitted from the query or body. `SearchRequest` switches
are `*bool`: nil omits the switch and a non-nil value is sent as `True`/`False`.
Response models keep nil for absent or null fields and retain every service
property in `RawFields`.

## Per-account association results

`CreateAssociations` and `DeleteAssociationByID` return one
`AssociationOperationResult` per account. The service answers HTTP 200 even
when an account fails; check `Status.Failed()` and `Status.ErrorCode` (for
example `CreateApplicationInstanceAssociation_AppEndpointNotFound`).
`GetAssociationStatusByID` likewise reports an unknown account as a failed
`StatusRecord` with HTTP 200. A missing association on `GetAssociationByID` is
an HTTP 404.

## Service restrictions

- `Create` and `UpdateByID` need directory write rights. With app-only Teams
  Administrator credentials the service rejects create with HTTP 401 and
  rejects changing `ApplicationID` with HTTP 400; it also requires
  `ApplicationID` on every update, so app-only updates are not possible.
- `OnpremPhoneNumber` and `UpdateTelephoneNumberByID` are rejected by the
  service as deprecated in favour of phone number assignment.
- `SyncByID` requires an application or ACS resource ID.
- The module has no cmdlet to delete a resource account; it is an Entra ID user
  and is removed through the directory.

The module's `-Identities` on `Get-CsOnlineApplicationInstance`, `-Force` on
the remoting cmdlets and `-CallbackUri` on `Sync-CsOnlineApplicationInstance`
fail parameter binding in the pinned release and have no wire effect, so they
have no SDK equivalent.
