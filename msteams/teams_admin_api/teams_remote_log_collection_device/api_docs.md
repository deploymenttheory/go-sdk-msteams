# Teams remote log collection devices

`Client.TeamsAdminAPI.TeamsRemoteLogCollectionDevices` manages the compiled
`TeamsRemoteLogCollectionDevice` resource. Each record asks one user's Teams
client to upload diagnostic logs until `ExpireAfter`.

| Go method | Cmdlet / parameter set | Wire |
|---|---|---|
| GetByIdentity, List | Get-CsTeamsRemoteLogCollectionDevice / Identity | GET …/configuration/{scope}/{id}, GET collection |
| GetByID | Get / ParentAndRelativeKey | GET …/configuration/Global/{id} |
| ListFiltered | Get / Filter | GET collection, local wildcard match |
| CreateByIdentity | New / Identity | POST collection |
| CreateByID, Create | New / ParentAndRelativeKey; bound or generated Id | POST collection |
| UpdateByIdentity, PatchByIdentity | Set / Identity | GET with Set metadata, then PATCH; or PATCH only |
| UpdateByID, PatchByID | Set / ParentAndRelativeKey | as above, Global/{id} |
| DeleteByIdentity | Remove / Identity | DELETE with Identity body |
| DeleteByID | Remove / ParentAndRelativeKey | DELETE with Id body |

The collection is `/Skype.Policy/configurations/TeamsRemoteLogCollectionDevice`.
Requests use the compiled policy headers (`MPACmdlet`, `X-MS-CmdletName` and a
fresh correlation ID) without forest discovery.

New requires `DeviceID`, `UserID` and `ExpireAfter`; an empty value fails with
`ErrInput` before I/O. The service, not the SDK, checks that the IDs are GUIDs
and returns 400 (`ErrorTeamsRemoteLogCollectionConfigurationInvalidDeviceId`).
The module's help asks for `ExpireAfter` in ISO 8601 UTC, now plus three days.
CreateByIdentity adds the identity's second component as `Id`; CreateByID sends
`Id: null`; Create generates a UUID, omits `Id` and keeps the generated ID in
its result even if submission fails. `Force` is sent as the string `True` or
`False` when set.

Set's fields use `policy.Field`, so omitted, null and empty stay distinct.
UpdateByID/Identity read the record with Set's cmdlet name, then PATCH once
with the read model's `Id` or `Identity`; `UpdateResult` reports the read and
whether the PATCH was sent and accepted. A missing record stops at the read
with 404. PatchByID needs an explicit `BodyID`. Mixed body selectors fail locally.

No write is replayed, read back or rolled back. A 204 returns no device.
Models keep every received field in `RawFields` and redact `String`/`GoString`.
The lab proves the read routes, the 400 for an invalid device ID and the 404s
for Set and Remove of a missing record; a successful create, update or delete
is unqualified. See the [contract notes](../../../docs/implementation/ext/devices-contract-notes.md).
