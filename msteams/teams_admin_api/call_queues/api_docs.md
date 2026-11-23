# Call queues

Use `sdk.TeamsAdminAPI.CallQueues` for Teams call queues. Operations use the
shared client/auth/transport controls through the ConfigAPI backend and need
no PowerShell runtime. Constructing the root client or this service performs
no network I/O.

| Resource method | PowerShell command | Wire operation | Declared successful statuses |
| --- | --- | --- | --- |
| `List(ctx, options)` | `Get-CsCallQueue` | GET `/Teams.VoiceApps/callqueues?FilterInvalidObos=False[&First=&Skip=&ExcludeContent=&NameFilter=&SortBy=&Descending=]` | 200 |
| `GetByID(ctx, identity)` | `Get-CsCallQueue -Identity` | GET `/Teams.VoiceApps/callqueues/{identity}?FilterInvalidObos=False` | 200 |
| `Create(ctx, settings, options)` | `New-CsCallQueue` | POST `/Teams.VoiceApps/callqueues[?channelUserObjectId=]` | 200, 201 |
| `UpdateByID(ctx, identity, settings, options)` | `Set-CsCallQueue` | PUT `/Teams.VoiceApps/callqueues/{identity}[?channelUserObjectId=]` | 200, 204 |
| `DeleteByID(ctx, identity)` | `Remove-CsCallQueue` | DELETE `/Teams.VoiceApps/callqueues/{identity}` | 200, 204 |

Every method takes a context and returns the actual Resty response alongside
its value/error. Reads are replay-safe; writes are never replayed
automatically. Any other HTTP status is an `*client.APIError`; an undeclared
success status is `ErrResponse`. Invalid input fails with `ErrInput` before
discovery or credentials are used.

## Missing queues

An unknown or deleted identity is HTTP **400** with code `100001`
(`CallQueue_IdentityInvalid`), not 404, for both get and delete.

## Settings and updates

`Settings` is the create and update body. Nil fields are omitted, blank strings
are omitted as the module removes them, and a non-nil empty list is sent as
`[]`. Actions, routing methods and `NoAgentApplyTo` are integers; use the
package constants (`RoutingLongestIdle`, `OverflowForward`, …).

The service replaces the whole queue on update. `Set-CsCallQueue` reads the
queue and fills every unchanged property from it; do the same with
`CallQueue.Settings`:

```go
queue, _, err := sdk.TeamsAdminAPI.CallQueues.GetByID(ctx, id)
settings := queue.Settings()
settings.AgentAlertTime = &alert
_, _, err = sdk.TeamsAdminAPI.CallQueues.UpdateByID(ctx, id, settings, nil)
```

`Settings()` copies resource IDs (`WelcomeMusicResourceId` →
`WelcomeMusicAudioFileId`, `MusicOnHoldResourceId` →
`MusicOnHoldAudioFileId`, `OboResourceAccounts` → `OboResourceAccountIds`) and
reduces action targets to their identifiers. The update reply carries only
diagnostics, so `UpdateByID` usually returns a nil queue; read it back with
`GetByID`.

## Local rules

As in the module:

- `Name` is required for create and update.
- `*ActionCallPriority` applies only when the matching action is Forward
  (0–5, 0 = default 3); otherwise it must be 0.
- `AutoRecordingTemplateID` requires `ConferenceMode` not set to false.
- Create with `RoutingLongestIdle` forces `PresenceAwareRouting` false.

PowerShell's renamed parameters map to wire names: `PresenceBasedRouting` →
`PresenceAwareRouting`, `ChannelId` → `ThreadID`. `-LineUri` and `-Agents`
are not sent by the module and have no SDK field. `WriteOptions` carries the
`channelUserObjectId` query for channel-connected queues.

## Response fields

`CallQueue` types common properties and keeps every property, including
nulls and unknown ones such as `ETag`, in `RawFields`.
