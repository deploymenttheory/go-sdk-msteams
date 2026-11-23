# Teams meeting configuration

Use `sdk.TeamsAdminAPI.TeamsMeetingConfigurations` for the compiled public
MicrosoftTeams 8.1.0 Get-CsTeamsMeetingConfiguration and
Set-CsTeamsMeetingConfiguration cmdlets.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; explicit Global or full scoped identity |
| `List` | Get with omitted Identity; bare collection, exact Global first |
| `ListFiltered` | Get Filter; local Identity/Tag wildcard matching |
| `PatchByIdentity` | One sparse PATCH with at least one bound setting |
| `UpdateByIdentity` | Set's preparation GET and sparse PATCH, retaining both receipts |

Both commands are compiled Core cmdlets (`BaseTpmGetCmdlet` /
`BaseTpmSetCmdlet` over `TeamsMeetingConfiguration`) that use the direct PolicyRP
provider: GET `/Skype.Policy/configurations/TeamsMeetingConfiguration[/configuration/{Identity}]`
and PATCH on the selected identity. Requests send `MPACmdlet: true` and command
metadata, with no modern ConfigAPI discovery. The module has no ConfigAPI proxy
entry for this type. Scoped identities split at the first slash. Empty or
invalid identities, empty or malformed filters, and canceled contexts fail
before credentials are acquired. Filter whitespace remains literal.

`Configuration` has all 21 getters as nullable values: `LogoURL`, `LegalURL`,
`HelpURL`, `CustomFooterText`, `DisableAnonymousJoin`,
`DisableAppInteractionForAnonymousUsers`, `EnableQoS`,
`EnableAttributedTranscripts`, `EnableGraphTranscriptAccess`,
`EnableWorkIQTranscriptAccess`, the six `uint32` client port and port-range
settings, `ClientMediaPortRangeEnabled`, `LimitPresenterRolePermissions`,
`FeedbackSurveyForAnonymousUsers`, `ReportMeeting` and
`PublishedEntraAuthenticationContexts`. The context list is
constructor-initialized in the module, so it is never nil after decoding. Each
element keeps its received JSON, so an Add edit re-sends existing elements
unchanged. `Configuration` also keeps unknown `RawFields` and an owned `RawJSON`
snapshot.

`ScalarSettings` uses `policy.Field[T]` for the 20 scalar Set parameters, so
omission, explicit null and false, zero or empty values stay distinct. The
compiled setters impose no URL, enum, length or port-range rule, so none is
added locally; strings such as `Enabled`/`Disabled` are sent verbatim.

`PublishedEntraAuthenticationContexts` is a PSListModifier in the module. In
`UpdateByIdentity` it is an `AuthenticationContextEdit` applied to the list read
during preparation, and the complete resulting list is sent:

- `ReplaceAuthenticationContexts(values...)` replaces the list; no values clears it.
- `AddAuthenticationContexts(values...)` appends without deduplication; no values
  is the module's empty modifier and clears the list.
- Remove is not offered: the pinned cmdlet's element comparer always throws
  `ErrorCannotRemoveElementsOf`, so a PowerShell Remove can never succeed.

An assigned element `Id` must match the pinned setter pattern
`^c([1-9]|[1-9][0-9])$` (`c1` to `c99`); nil elements are rejected.
`PatchByIdentity` takes a full replacement list directly.

Set requires an Identity and has no Global default. Bound Force keeps the
compiled string form. `UpdateResult` keeps the preparation read, mutation
invocation and HTTP acceptance separate. A 204 receipt does not show that a
setting persisted or took effect. Mutations never replay automatically. The
assigned baseline has no public create or delete command for this singleton.

The SDK lab identity has Teams Administrator. Certificate and client-secret
acceptance tests read and list Global and compare all 21 getters with pinned
PowerShell. They set `CustomFooterText` and `HelpURL`, verify that the other 19
getters are preserved, and restore the originals (an empty footer and an
explicit-null help link) with an independent read. Recovery journals the
originals before the mutation. CI serializes these Global reads and writes.
Permanent sanitized acceptance `testdata` supplies the unit response fixtures.

These tests do not qualify live writes of the other settings, authentication
context list edits (the lab list is empty), non-Global identities, effective
meeting behavior, delegated authentication or other clouds.

See [qualification](../../../docs/implementation/policies/meeting-configuration-lab-acceptance.md).
