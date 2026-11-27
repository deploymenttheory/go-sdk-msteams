# Meeting configuration qualification

Owner: #129 (POL-56). Baseline: MicrosoftTeams 8.1.0.

The pinned public runtime metadata (`runtime.json` commands 135 and 607) has
Get-CsTeamsMeetingConfiguration Identity/Filter and
Set-CsTeamsMeetingConfiguration Identity. Both are `Cmdlet`s, not proxy
functions, and `reconciliation.json` records 158 and 668 carry no
`CmdletConfig.json` entry (`configuration: null`), so there is no
Get-/Set-CsConfigurationModern dispatch for this type. In the pinned module the
default session resolves them to the compiled Core classes
`GetTeamsMeetingConfigurationCmdlet` (`BaseTpmGetCmdlet<TeamsMeetingConfiguration>`)
and `SetTeamsMeetingConfigurationCmdlet` (`BaseTpmSetCmdlet<...>`), which use the
direct PolicyRP provider. This was checked by reflecting the hash-verified pinned
assemblies.

The SDK therefore uses the compiled direct protocol, as for
TeamsMessagingConfiguration and TeamsGuestMeetingConfiguration:

- GET `/Skype.Policy/configurations/TeamsMeetingConfiguration` (bare list)
- GET `.../configuration/{Identity}`
- Set: a preparation GET, then a sparse PATCH of the same instance route with the
  read Identity and only bound settings. The receipt is HTTP 204 with an empty
  body.

`TeamsMeetingConfiguration` is registered in the shared `policyrp` resource
allow-list. Requests carry `MPACmdlet: true` and `X-MS-CmdletName`. Filter
applies locally to Identity and `Tag:` Identity with wildcard matching, and its
whitespace is literal.

The compiled type has Identity and 21 getters: four strings (`LogoURL`,
`LegalURL`, `HelpURL`, `CustomFooterText`), eight `Nullable<bool>`, six
`Nullable<uint>` client ports and ranges, two strings
(`FeedbackSurveyForAnonymousUsers`, `ReportMeeting`) and the get-only,
constructor-initialized `List<PublishedEntraAuthenticationContext>`. Set binds
all 21, including the 8.1.0 `EnableWorkIQTranscriptAccess`, plus Force and a
mandatory positional Identity. Scalar strings have no enum, URL or length
validation. The six nullable UInt32 DTO setters require values from 1 through
UInt32 max (or null), and reject zero; their public parameter setters simply
store values, so this error occurs after the preparation read.

`SetTeamsMeetingConfigurationCmdlet.UpdateProperties` applies the
`PublishedEntraAuthenticationContexts` PSListModifier through
`PSListModifierHelper.ApplyTo` with an element comparer that always throws
`ErrorCannotRemoveElementsOf`. Running the pinned helper offline showed: an empty
modifier and an empty Replace clear the list; a nonempty Replace replaces it and
ignores Add; Add appends without deduplication (even for a repeated Id); any
Remove against a populated list fails at the comparator; a nonnull Remove
against an empty list succeeds. Null Remove fails even when empty. The SDK
now exposes ChangeAuthenticationContexts as well as Replace and Add. Null
container members are allowed, while typed writes project known fields and
drop unknown raw-response data. The element type
`PublishedEntraAuthenticationContext` has seven nullable members (Identity, Id,
PublishedName, PublishedDescription, EntraName, EntraDescription, Published); its
Id setter enforces the culture-invariant pattern `^c([1-9]|[1-9][0-9])$`, which
the SDK applies to assigned Ids. The separate Get/New/Set/Remove
PublishedEntraAuthenticationContext cmdlets are not part of this family.

Oracle checks pin these module assembly SHA256 values:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`
- ConfigAPI: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

On 2026-11-23 the dedicated SDK service principal (Teams Administrator) read and
listed Global with both client-secret and certificate authentication. All 21
typed getters matched pinned PowerShell 8.1.0. The lab's Global had Tenant
authority, an empty context list, an empty `CustomFooterText` and a null
`HelpURL`.

The update test journals both original strings (null included) in a bound
private recovery journal before writing, then sets `CustomFooterText` and
`HelpURL` to inert acceptance values in one `UpdateByIdentity`. It waits for a
persisted read-back in which the other 19 getters are unchanged, and compares
that read-back with the pinned getter. The journal handler then restores the
originals, sending `""` for the footer and an explicit JSON null for the help
link. An independent read confirms the complete original projection and is
compared again. Both application auth modes passed, so explicit-null persistence
of a string is qualified for `HelpURL`. Each run left exactly one
`meeting-configuration` journal entry, in the cleaned state. No licenses, roles,
users, groups or policy assignments were changed.

CI runs these tests as the `teams_meeting_configuration` partition. A
serialized `global-reads` step and an `update-restore` step run under the tenant
lock. The update step requires three pinned comparisons, the update and
restoration assertions and one cleaned restore entry. The partition uses its own
read-only oracle script, `read-meeting-configuration.ps1`, so it does not touch
the shared configuration oracle or lab kit. Registering the type in the shared
`policyrp` allow-list is the one shared-code change.

Five permanent sanitized response files come directly from these acceptance
reads, the update and the read-backs. The real 204 body is empty. Unit tests use
them for native routing, headers, typed getters, snapshot ownership, filter
selection, the Set preparation read and sparse PATCH bodies. The bodies cover
every setting, explicit null, false, positive integers, empty string, Force and every
context-list edit shape (on an in-memory copy of the actual response with a
populated list). Common denial propagation reuses the actual privacy-service 403
envelope; it is not a live denial claim for this service. Deliberately malformed
replies test decoding failures separately from successful fixtures.

Still unqualified:

- non-Global identities
- live writes of the other 19 settings
- live context-list edits and nonempty context lists (the lab list is empty, and
  contexts change meeting access controls)
- effective meeting behavior and propagation
- delegated authentication and other clouds

No create or delete methods are invented; the pinned module exports none.


## Pinned context-ID validation edge (2026-11-26)

Offline reflection of 8.1.0 `PublishedEntraAuthenticationContext.set_Id` (Core assembly SHA-256 `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`) checks null at IL3/8 and the regex at IL17–28; `.cctor` IL0 defines `^c([1-9]|[1-9][0-9])$`. Fresh controlled instances accepted and retained `c1` and `c1` followed by one LF; `c1` followed by CRLF and `c100` rejected. This is local setter evidence, not tenant acceptance.

Go's end anchor differs from the .NET rule. Validation now ignores exactly one final LF for regex comparison while retaining the original Id pointer/string and serialized request. Regression tests cover accepted final LF, rejected CRLF/repeated LF/out-of-range values, and an authored PATCH preserving the LF. This follows the existing banner-model validation convention; no general whitespace normalization or claim of server persistence is added.


## Source-confirmed continuation corrections

The [write review](pol91-meeting-write-source-review.md), eleven helper/serializer
outcomes and 24 positive-port setter/writer outcomes supersede the earlier
inferences that zero was a valid port body and Remove always failed. The real
initial combined read failure at 4e7a06d/34406 is retained, not relabeled a
fixture error. The corrected c9e096f/73712 full affected check passed 925 test
events across six test-bearing packages, plus provider compile-only;
targeted gosec/nolintlint passed. These are offline/source results, not new
live outcomes or qualification status changes.
