# SIP devices gateway

`Client.TeamsAdminAPI.SIPDevicesGateway` drives the SIP devices gateway (SDG)
through ConfigAPI. Requests go to the discovered admin forest with
`Accept: application/json` and the cmdlet name in `X-MS-CmdletName`.

| Go method | Cmdlet / parameter set | Wire | Success |
|---|---|---|---|
| ListBulkSignInRequests | Get-CsSdgBulkSignInRequestsSummary / Get | GET /Teams.SipDevicesGatewayManagement/v1/batchsummaries | 200 |
| GetBulkSignInRequestStatus | Get-CsSdgBulkSignInRequestStatus / Get | GET …/v1/batchstatus?batchid= | 200 |
| CreateBulkSignInRequest | New-CsSdgBulkSignInRequest / __AllParameterSets | POST …/v1/createbatch?TargetRegion= | 200 |
| ReadBulkSignInItems | New-CsSdgBulkSignInRequest -DeviceDetailsFilePath (local) | none | — |
| CreateDeviceTaggingRequest | New-CsSdgDeviceTaggingRequest / NewExpanded, New | POST …/devicetagging/tagnewdevice?TargetRegion= | 200 |
| CreateDeviceTransferRequest | New-CsSdgDeviceTransferRequest / New | POST …/devicetransfer/transfer?SourceRegion=&TargetRegion=&SdhDeviceId= | 200 |

Bulk sign-in sends a JSON array of `{userName, hardwareId}`; nil fields are
omitted, as the module omits null values. `ReadBulkSignInItems` does what the
module's wrapper does with `Import-Csv`: it maps the `Username` and `HardwareId`
columns by name (any case, BOM stripped, extra columns ignored), and a missing
column leaves that field nil. An empty batch, a blank region or batch ID, and
any blank transfer value fail with `ErrInput` before I/O.

The status result keeps every device's `status` and `error`. A completed batch
can still contain failed devices: check `BulkSignInItemStatus.Failed`. The
summary is the `{"requests": [...]}` envelope; the module writes its items.
Tagging sends the module's mixed-case `OCEUserName`, `HardwareId`, `TenantId`,
`SdhRegion` and `IcmId`. Transfer has no body.

Every operation accepts only 200, as the module does. Any other 2xx returns
`ErrResponse`; errors return `*client.APIError`. The POSTs are never replayed.
Models keep unknown fields in `RawFields`, and the status, item and tagging
models redact `String`/`GoString` because they carry account names.

Tagging and transfer are exported only to Microsoft support environments
(`Torus`, `SelfHost`). The lab's app-only Teams Administrator gets 403 for every
gateway operation, so success shapes are proven from the module's serializers
only. See the [contract notes](../../../docs/implementation/ext/devices-contract-notes.md).
