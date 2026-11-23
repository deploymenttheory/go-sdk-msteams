# Resource account contracts

Owner: #75 (CALL-27). Baseline: MicrosoftTeams 8.1.0.

CmdletConfig marks Find, Get/New/Remove association and association status as
AutoRest cmdlets, and Get/New/Set/Set-Voice/Sync-CsOnlineApplicationInstance
as Remoting cmdlets dispatched to their `…V2` modern cmdlets. The generated
ConfigApiBasedCmdlets methods give the wire routes:

- Get: GET `/Teams.PlatformService/v2/Legacy/ApplicationInstances` with
  optional `identity`, `resultSize`, `skip` queries.
- New: POST `/Teams.PlatformService/v2/ApplicationInstances`, body
  `userPrincipalName`, `applicationId`, `displayName`.
- Set: PATCH `…/ApplicationInstances/{identity}`, body `applicationId`,
  `displayName`, `onpremPhoneNumber`, `acsResourceId`.
- Set-Voice: PATCH `…/ApplicationInstances/{identity}/telephoneNumber`, body
  `telephoneNumber`.
- Sync: POST `…/ApplicationInstances/{objectId}/sync` with `applicationId` and
  `acsResourceId` queries.
- Find: GET `/Teams.VoiceApps/applicationinstances` with SearchQuery,
  MaxResults, ExactMatchOnly, IsAssociated, AssociatedOnly, UnAssociatedOnly;
  unbound values are omitted and booleans use .NET `True`/`False`.
- Associations: GET/DELETE `/Teams.VoiceApps/applicationinstanceassociations/{id}`,
  GET `…/{id}/status`, POST `…/applicationinstanceassociations` with
  `EndpointsIds`, `ConfigurationId`, `ConfigurationType`, `CallPriority`.

Model serializers give the property names: platform accounts are camelCase
(`objectId`, `tenantId`, `userPrincipalName`, `applicationId`, `acsResourceId`,
`displayName`, `phoneNumber`), Find results PascalCase (`Id`, `SipUri`,
`IsOnPrem`, …), associations `Id`, `ApplicationConfigurationType`,
`ConfigurationId`, `CallPriority`, and results wrap `ApplicationInstanceAssociation`
and `StatusRecord` (`Id`, `Status`, `ErrorCode`, `Message`, `StatusTimestamp`,
`AuxiliaryData`). The PowerShell projection renames `ErrorCode` to a
`StatusCode` enum without its provider prefix; the SDK keeps the wire value.

Custom wrappers: Get association and each Remove identity pass through
`EncodeSipUri` (`:` → `[COLON]`, `.` → `[DOT]`); status and New do not. New
rejects CallPriority outside 0–5 locally and omits it when unbound; the service
accepts 1–5. Remove loops, sending one DELETE per identity. Pinned live
dispatch fails `NamedParameterNotFound` for `Get -Identities`, `-Force` on every
remoting cmdlet and `Sync -CallbackUri`: TrpsPSCmdlet forwards them to V2
cmdlets that do not declare them. They have no SDK equivalent.

Lab observations (client-secret, app-only Teams Administrator, no resource
accounts present):

- List returned `[]`; Find returned `{"ApplicationInstances":[],"Diagnostics":[]}`.
- Get of an unknown object ID or UPN is a JSON 404 (`NotFound`).
- Unknown association is a 404 with top-level `code` 100004; status for an
  unknown endpoint is HTTP 200 with
  `ApplicationInstanceAssociationStatusProvider_AppEndpointNotFound`.
- New with an unverified domain is 400 `BadArgument`; with the tenant's own
  domain it is 401 `Unauthorized` ("user not authorized … in MsGraph API"), so
  app-only credentials cannot create resource accounts. No account was created.
- Set without `applicationId` is 400 ("ApplicationId needs to be a valid
  GUID"); with it, 400 ("Application based authentication is not supported …
  if updating the ApplicationId"). `onpremPhoneNumber` and the
  telephoneNumber route are 400 as deprecated in favour of
  Set/Remove-CsPhoneNumberAssignment.
- Sync without IDs is 400; with an application ID it is 204, even for an
  unknown object ID.
- Association create/remove for an unknown endpoint is HTTP 200 with one failed
  result (`CreateApplicationInstanceAssociation_AppEndpointNotFound`,
  `RemoveApplicationInstanceAssociation_AssociationNotFound`).

The successful Get-by-identity body shape was not observed (no accounts and no
create rights); GetByID accepts either the object or a one-item array.
Successful create/update/association bodies are modelled from the serializers
and covered by synthetic contract cases only.

Acceptance: `TestAccResourceAccountReads` (read-only, with pinned PowerShell
count and status parity) and `TestAccResourceAccountRejectedWrites` (serial; every
write targets the reserved IDs `00000000-0000-4000-8000-00000000000{1,2}` or an
`example.invalid` UPN, so nothing is created or changed). Sanitized responses
from these runs feed the unit contract tests. Creating, updating, associating
and syncing a real resource account, and phone number behaviour, remain
unqualified and need delegated or directory-write credentials.
