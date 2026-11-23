# Call queue contracts

Owner: #180 (CALL-23). Baseline: MicrosoftTeams 8.1.0.

CmdletConfig marks Get/New/Set/Remove-CsCallQueue as AutoRest cmdlets with
custom wrappers in `Merged_custom_PsExt.ps1`. The generated
ConfigApiBasedCmdlets methods give the wire routes:

- Get: GET `/Teams.VoiceApps/callqueues` with optional `First`, `Skip`,
  `ExcludeContent`, `NameFilter`, `SortBy`, `Descending` and
  `FilterInvalidObos`; GET `/Teams.VoiceApps/callqueues/{identity}` with
  `FilterInvalidObos`. The wrapper always sends `FilterInvalidObos=False`,
  maps `-Sort` to `SortBy`, and rejects `-Identity` with any list parameter.
  Booleans use .NET `True`/`False`.
- New: POST `/Teams.VoiceApps/callqueues[?channelUserObjectId=]` with a
  `CreateCallQueueRequest` body.
- Set: **PUT** `/Teams.VoiceApps/callqueues/{identity}[?channelUserObjectId=]`
  with an `UpdateCallQueueRequest` body (the create properties minus `Agents`,
  92 properties).
- Remove: DELETE `/Teams.VoiceApps/callqueues/{identity}`. The wrapper GETs the
  queue first only to echo it.

Responses wrap the queue: `{"CallQueue":{…}}` for create/get and
`{"CallQueues":[…],"Diagnostics":[…]}` for list. The `CallQueue` serializer
has 126 properties, adding read-only ones (`Identity`, `TenantId`, `Agents`,
`ApplicationInstances`, `Statistics`, `*FileName`, `*ResourceId`,
`*FileDownloadUri`, `OboResourceAccounts`, …); the live body also carried
`ETag`, 135 properties in all. Enumerations are integers on the wire:
RoutingMethod Attendant=0, Serial=1, RoundRobin=2, LongestIdle=3;
OverflowAction DisconnectWithBusy=0, Forward=1, Voicemail=2,
SharedVoicemail=3; TimeoutAction Disconnect=0, Forward=1, Voicemail=2,
SharedVoicemail=3; NoAgentAction Queue=0, Disconnect=1, Forward=2,
Voicemail=3, SharedVoicemail=4; NoAgentApplyTo NewCalls=0, AllCalls=1.

Custom wrapper rules the SDK mirrors:

- New and Set rename `PresenceBasedRouting` to `PresenceAwareRouting` and
  `ChannelId` to `ThreadId`, drop `LineUri`, and remove blank optional
  strings. New forces `PresenceAwareRouting` false for LongestIdle.
- `*ActionCallPriority` is valid only with a Forward action and then 0–5
  (0 = default 3); with another action it must be 0. Checked locally
  ("…only applicable when the 'OverflowAction' is 'Forward'").
- `AutoRecordingTemplateId` requires ConferenceMode enabled.
- Set reads the queue and fills every unbound parameter from it, so the PUT
  carries the whole queue: `WelcomeMusicAudioFileId` from
  `WelcomeMusicResourceId`, `MusicOnHoldAudioFileId` from
  `MusicOnHoldResourceId`, `OboResourceAccountIds` from
  `OboResourceAccounts`, action targets from their IDs, and the remaining
  writable properties when non-blank. `CallQueue.Settings` reproduces this
  backfill; callers change the result and pass it to `UpdateByID`.

PowerShell `-Agents` is never bound by the wrappers and is not modelled. The
wrapper's PowerShell defaults (`PresenceBasedRouting`, `ConferenceMode` true)
are not bound, so they are not sent unless explicit; the service applies its
own defaults.

Lab observations (client-secret, app-only Teams Administrator, no
pre-existing queues):

- List returned `{"CallQueues":[],"Diagnostics":[]}`.
- Get or Delete of an unknown or deleted identity is HTTP **400** with
  top-level `code` 100001 and `errorDetails[0].ErrorCode`
  `CallQueue_IdentityInvalid`, not 404.
- App-only credentials can create, update and delete queues. Create answered
  201 with the queue; the service defaulted ConferenceMode true, AllowOptOut
  false and every action 0.
- Update answered 200 with `{"Diagnostics":[…]}` warnings and no queue, so the
  module (and the SDK) read the queue back. Delete answered 200
  `{"Diagnostics":[]}`.

Acceptance: `TestAccCallQueueLifecycle` (mutation, one journal entry) reads a
reserved missing ID, creates an owned `msteams-sdk-…-queue`, reads and lists
it, replaces its settings from `CallQueue.Settings` with three changes, proves
the unchanged settings survived the PUT, compares the read with pinned
PowerShell, deletes it and observes the 400 absence. Recovery kind
`call-queue` resolves a lost create by exact name through `NameFilter`.
Sanitized responses feed the unit contract tests. Agents, distribution lists,
users, resource-account (OBO) links, channel queues, shared voicemail,
callback and compliance-recording settings remain unqualified; they need
tenant users, groups, Teams channels or licensed resource accounts.
