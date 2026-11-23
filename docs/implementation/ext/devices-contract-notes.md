# Devices management contracts

Owner: #87 (EXT-24). Baseline: MicrosoftTeams 8.1.0. Go packages:
`sip_devices_gateway`, `teams_remote_log_collection_configuration` and
`teams_remote_log_collection_device`.

## SIP devices gateway (AutoRest)

CmdletConfig marks the five SDG cmdlets as AutoRest. The bulk sign-in commands
export to `TeamsGA`; `New-CsSDGDeviceTaggingRequest` and
`New-CsSDGDeviceTransferRequest` export only to `Torus` and `SelfHost`, which is
why the unauthenticated import did not list them. All five have proxies in
`ProxyCmdletDefinitionsWithHelp.ps1`.

The IL of each `ConfigApiBasedCmdlets.Sdg*` method gives the route and verb, and
its `_Call` method dispatches only status 200 to the success deserializer:

| Operation | Verb and route | Success model |
|---|---|---|
| SdgBulkSignInRequestsSummaryGet | GET `/Teams.SipDevicesGatewayManagement/v1/batchsummaries` | SdgBulkSignInRequestsSummaryResponse |
| SdgBulkSignInRequestStatusGet | GET `…/v1/batchstatus?batchid=` | SdgBulkSignInRequestStatusResult |
| SdgBulkSignInRequestNew | POST `…/v1/createbatch?TargetRegion=`, JSON array body | SdgBulkSignInResponse |
| SdgDeviceTaggingRequestNew | POST `…/devicetagging/tagnewdevice?TargetRegion=`, object body | SdgDeviceTaggingResponse |
| SdgDeviceTransferRequestNew | POST `…/devicetransfer/transfer?SourceRegion=&TargetRegion=&SdhDeviceId=`, no body | SdgDeviceTransferResponse |

Every other status goes to `SdgErrorResponse`. The model serializers (via
`ToJsonString` on populated instances) give the JSON names:

- request item: `userName`, `hardwareId`;
- summary: `{"requests": [{batchId, batchStatus, batchItemsCount, userId, requestTime, region}]}`;
- status: `batchId`, `batchStatus`, `userId`, `requestTime`, `region`, `batchItems: [{userName, hardwareId, status, error}]`;
- created batch: `batchId`, `batchItemsCount`;
- tagging request: `OCEUserName`, `HardwareId`, `TenantId`, `SdhRegion`, `IcmId`;
- tagging response: `correlationId`, `status`; transfer response: `Message`.

Null properties are omitted. The public `New-CsSdgBulkSignInRequest` is a custom
wrapper: it reads `-DeviceDetailsFilePath` with `Import-Csv`, builds one
`SdgBulkSignInRequestItem` per row from the `Username` and `HardwareId` columns,
and calls the internal cmdlet with `-Body` and `-TargetRegion $Region`. Both
wrapper parameters are mandatory. The SDK keeps the CSV step as the local
`ReadBulkSignInItems` helper and takes the items directly.

Lab observations (client-secret, app-only Teams Administrator): every SDG route
returns HTTP 403 `{"code":"Forbidden","message":"Access Denied.","action":"Provide
different credential or request access."}`. Pinned PowerShell fails the same
way (`Forbidden,…GetCsSdgBulkSignInRequestsSummary_Get`). Bulk sign-in, tagging
and transfer were sent only with an `example.invalid` account, a documentation
MAC address, an invalid region and reserved IDs, so nothing could act on a real
device.

## Remote log collection (compiled policy provider)

`Get-CsTeamsRemoteLogCollectionConfiguration` and the four
`*-CsTeamsRemoteLogCollectionDevice` cmdlets are compiled cmdlets in
`Microsoft.Teams.Policy.Administration.Cmdlets.Core`, on the same provider base
as custom banner text. The provider's route strings are
`/Skype.Policy/configurations/{name}` and `/configuration/`. Raw probes settled
the two resource names:

- `TeamsRemoteLogCollectionConfiguration`: the collection returns
  `[{"Devices":[],"Identity":"Global","SchemaVersion":"0.0.1","ConfigId":"Global","ConfigMetadata":{…}}]`
  and `/configuration/Global` returns the object.
- `TeamsRemoteLogCollectionDevice`: the collection returns `[]`, and
  `/configuration/Global/{id}` returns 404 with
  `"\"{id}\" not found"` and `errorCode` 40407, the same message the cmdlet shows.
  `Devices` is not a child path of the configuration (404).

Reflection gives the parameter sets: Get has Identity, ParentAndRelativeKey
(`-Id`) and Filter; New has Identity and ParentAndRelativeKey with mandatory
`DeviceId`, `ExpireAfter` and `UserId` strings and optional `-Id`; Set has the
same two sets with all three optional; Remove has Identity and
ParentAndRelativeKey. The DTO is `TeamsRemoteLogCollectionDevice {Identity, Id,
DeviceId, ExpireAfter, UserId}`. The help asks for GUID user and device IDs and
an ISO 8601 UTC expiry of now plus three days. The configuration Set cmdlet
exists in the assembly but is not exported.

The compiled cmdlets do not print their HTTP traffic, so the body shapes follow
the shared provider base already settled for custom banner text (POST to the
collection with Identity and Id metadata, read-then-PATCH for Set, DELETE with
an Id or Identity body). Raw probes confirmed the routes:

- POST with `DeviceId: "not-a-guid"` returns 400 `errorCode` 40013 with inner
  `ErrorTeamsRemoteLogCollectionConfigurationInvalidDeviceId`;
- PATCH and DELETE of `Global/00000000-0000-4000-8000-000000000001` return 404
  40407;
- pinned `Set-` and `Remove-CsTeamsRemoteLogCollectionDevice -Id` of the same ID
  fail with the same 404.

## Acceptance

One lab package per service under `tests/integration/lab`:

- `teams_remote_log_collection_configuration`:
  `TestAccRemoteLogCollectionConfigurationReads` (read-only) lists, reads and
  filters the configuration.
- `teams_remote_log_collection_device`: `TestAccRemoteLogCollectionDeviceReads`
  (read-only) lists devices and reads one missing device, then compares the
  configuration and device counts and the not-found code with
  `oracle/read-remote-log-collection.ps1` ("PowerShell 8.1.0 / Go remote log
  collection parity: passed"). `TestAccRemoteLogCollectionDeviceRejectedWrites`
  (serial) creates a device with an invalid DeviceId (400) and updates, patches
  and deletes the reserved missing device (404; Update stops at its read).
- `sip_devices_gateway`: `TestAccSIPDevicesGatewayDenied` (read-only) checks the
  403 for the summary and status reads; `TestAccSIPDevicesGatewayRejectedWrites`
  (serial) checks the 403 for bulk sign-in, tagging and transfer.

Their sanitized responses feed the unit contract tests. Unqualified: every
successful SDG operation (needs delegated credentials with SDG access, and real
devices); successful device create, update and delete (the lab creates no
device records); and non-Global configuration instances.
