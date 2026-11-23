# Guest meeting configuration qualification

Owner: #123 (POL-50). Baseline: MicrosoftTeams 8.1.0.

The pinned public runtime metadata (`runtime.json` commands 127 and 600) has
Get-CsTeamsGuestMeetingConfiguration Identity/Filter and
Set-CsTeamsGuestMeetingConfiguration Identity. Both are `Cmdlet`s, not proxy
functions. In the pinned module the default session resolves them to the compiled
Core classes `GetTeamsGuestMeetingConfigurationCmdlet`
(`BaseTpmGetCmdlet<TeamsGuestMeetingConfiguration>`) and
`SetTeamsGuestMeetingConfigurationCmdlet` (`BaseTpmSetCmdlet<...>`), which use
the direct PolicyRP provider. `custom/CmdletConfig.json` (`reconciliation.json`
records 150 and 661) and `exports/ProxyCmdletDefinitionsWithHelp.ps1` also contain
modern ConfigAPI `Remoting` entries for this type. The shipped module does not
export or resolve those entries. They declare only six settings, as non-nullable
Booleans and strings, and an optional Identity. The exported runtime instead has
eight nullable settings, Force and a mandatory Identity. This is the same
compiled-over-config selection already qualified for TeamsMobilityPolicy and
TeamsEventsPolicy.

The SDK therefore uses the compiled direct protocol, as for
TeamsMessagingConfiguration and TeamsMultiTenantOrganizationConfiguration:

- GET `/Skype.Policy/configurations/TeamsGuestMeetingConfiguration` (bare list)
- GET `.../configuration/{Identity}`
- Set: a preparation GET, then a sparse PATCH of the same instance route with the
  read Identity and only bound settings. The receipt is HTTP 204 with an empty
  body.

Requests carry `MPACmdlet: true` and `X-MS-CmdletName`. Filter applies locally
to Identity and `Tag:` Identity with wildcard matching, and its whitespace is
literal. The compiled type `TeamsGuestMeetingConfiguration` has Identity and eight
getters. `AllowIPVideo`, `AllowMultipleScreenshare`, `AllowMeetNow`,
`AllowTranscription`, `AllowParticipantGiveRequestControl` and
`AllowExternalParticipantGiveRequestControl` are `Nullable<bool>`.
`ScreenSharingMode` and `LiveCaptionsEnabledType` are strings with no source enum
validation. Oracle checks pin these module assembly SHA256 values:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`
- ConfigAPI: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

On 2026-11-23 the dedicated SDK service principal (Teams Administrator) read and
listed Global with both client-secret and certificate authentication. All eight
typed getters matched pinned PowerShell 8.1.0. The lab's original Global had Host
authority and AllowMeetNow true.

The update test refuses a missing or null original AllowMeetNow. It persists the
original in a bound private recovery journal and toggles only AllowMeetNow through
`UpdateByIdentity`. It waits for a persisted read-back in which the other seven
getters are unchanged, and compares that read-back with the pinned getter. It then
restores the original value through the journal handler, independently reads the
complete original projection back, and compares again. Both application auth modes
passed. Each run left exactly one `guest-meeting-configuration` journal entry, in
the cleaned state. After restoration every getter equals the original, but Global
authority is now Tenant rather than Host, as for every other singleton written
through the service. No licenses, roles, users, groups or policy assignments were
changed.

CI runs these tests as the `teams_guest_meeting_configuration` partition. A
serialized `global-reads` step and an `update-restore` step run under the tenant
lock. The update step requires three pinned comparisons, the update and
restoration assertions and one cleaned restore entry. The partition uses its own
read-only oracle script, `read-guest-meeting-configuration.ps1`, so it does not
touch the shared configuration oracle or lab kit.

Five permanent sanitized response files come directly from these acceptance
reads, the update and the read-backs. The real 204 body is empty. Unit tests use
them for native routing, headers, typed getters, snapshot ownership, filter
selection, the Set preparation read, and sparse PATCH bodies. The bodies cover
every setting, explicit null, false, empty string and Force. Common denial
propagation reuses the actual privacy-service 403 envelope; it is not a live
denial claim for this service. Deliberately malformed replies test decoding
failures separately from successful fixtures.

Still unqualified:

- non-Global identities
- live writes of the other seven settings
- explicit null or empty persistence
- effective guest meeting behavior and propagation
- delegated authentication and other clouds

No create or delete methods are invented; the pinned module exports none.
