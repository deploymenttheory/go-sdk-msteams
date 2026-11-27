# Teams guest meeting configuration

Use `sdk.TeamsAdminAPI.TeamsGuestMeetingConfigurations` for the compiled public
MicrosoftTeams 8.1.0 Get-CsTeamsGuestMeetingConfiguration and
Set-CsTeamsGuestMeetingConfiguration cmdlets.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; explicit Global or full scoped identity |
| `List` | Get with omitted Identity; bare collection, exact Global first |
| `ListFiltered` | Get Filter; local Identity/Tag wildcard matching |
| `PatchByIdentity` | One sparse PATCH with at least one bound setting |
| `UpdateByIdentity` | Set's preparation GET and sparse PATCH, retaining both receipts |

The exported commands are compiled Core cmdlets that use the direct PolicyRP
provider. The module's unexported ConfigAPI proxy entries for the same type are
not the selected protocol. Requests send `MPACmdlet: true` and command metadata,
with no modern ConfigAPI discovery. Scoped identities split at the first slash.
Empty or invalid identities, empty or malformed filters, and canceled contexts
fail before credentials are acquired. Filter whitespace remains literal.

`Configuration` has all eight getters as nullable values. These are
`AllowIPVideo`, `ScreenSharingMode`, `AllowMultipleScreenshare`, `AllowMeetNow`,
`LiveCaptionsEnabledType`, `AllowTranscription`,
`AllowParticipantGiveRequestControl` and
`AllowExternalParticipantGiveRequestControl`. It also keeps unknown `RawFields`
and an owned `RawJSON` snapshot. `Settings` uses `policy.Field[T]`, so omission,
explicit null and false or empty values stay distinct. The two string settings
have no invented enum restriction. The documented values, such as `EntireScreen`
and `DisabledUserOverride`, are service values and are not checked locally.

Set requires an Identity and has no Global default. Bound Force keeps the
compiled string form. `UpdateResult` keeps the preparation read, mutation
invocation and HTTP acceptance separate. A 204 receipt does not show that a
setting persisted or took effect. Mutations never replay automatically. The
assigned baseline has no public create or delete command for this singleton.

The SDK lab identity has Teams Administrator. Certificate and client-secret
acceptance tests read and list Global and compare all eight getters with pinned
PowerShell. They toggle `AllowMeetNow`, verify that the other seven getters are
preserved, and restore the original value with an independent read. Recovery
journals the original value before the mutation. CI serializes these Global reads
and writes. Permanent sanitized acceptance `testdata` supplies the unit response
fixtures.

These tests do not qualify live writes of the other settings, explicit
null/empty persistence, effective guest meeting behavior, delegated
authentication or other clouds.

See [qualification](../../../docs/acceptance/guest-meeting-configuration-lab-acceptance.md).
