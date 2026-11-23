# Teams remote log collection configuration

`Client.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations` reads the
compiled `TeamsRemoteLogCollectionConfiguration` resource. The tenant has one
`Global` instance whose `Devices` lists pending remote log collection requests.

| Go method | Cmdlet / parameter set | Wire |
|---|---|---|
| GetByIdentity, List | Get-CsTeamsRemoteLogCollectionConfiguration / Identity | GET …/configuration/{identity}, GET collection |
| ListFiltered | Get / Filter | GET collection, local wildcard match on Identity |

The collection is `/Skype.Policy/configurations/TeamsRemoteLogCollectionConfiguration`.
List returns the bare array with Global first and no invented paging. An empty
filter fails with `ErrInput`. Models keep unknown fields, such as
`SchemaVersion` and `ConfigMetadata`, in `RawFields`, and redact `String`/`GoString`.

The compiled assembly contains a Set cmdlet for this configuration, but 8.1.0
does not export it, so the SDK has no write method. Add and remove
devices with `TeamsRemoteLogCollectionDevices`. See the
[contract notes](../../../docs/implementation/ext/devices-contract-notes.md).
