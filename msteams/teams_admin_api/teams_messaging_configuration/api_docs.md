# Teams messaging configuration

Use `sdk.TeamsAdminAPI.TeamsMessagingConfigurations` for the compiled public
MicrosoftTeams 8.1.0 Get and Set commands.

| Native operation | PowerShell behavior |
|---|---|
| `GetByIdentity` | Get Identity; explicit Global or full scoped identity |
| `List` | Get with omitted Identity; bare collection, exact Global first |
| `ListFiltered` | Get Filter; local Identity/Tag wildcard matching |
| `PatchByIdentity` | One sparse PATCH, with at least one bound setting |
| `UpdateByIdentity` | Set's preparation GET and sparse PATCH; local list edits |

The direct compiled provider sends `MPACmdlet: true` and command metadata without
modern ConfigAPI discovery. Scoped identities split at the first slash; the
remaining name stays one escaped segment. Empty identity/filter and malformed
filters fail before credentials; filter whitespace remains literal.

All 12 public settings are represented, including 8.1.0 `FileTypeBlockList` and
`FileTypeCustomizationSetting`. Scalar `policy.Field[T]` values distinguish
omission, explicit null and false/zero/empty values. `Storyline` is the pinned
Int32 enum (Disabled=0, Enabled=1); reads accept its named/numeric forms and retain
unknown numeric values. String settings have no invented enum restrictions.
A nonnil `policy.Strings(...)` replaces the full file block list; empty clears it.
`policy.ReplaceStrings` / `policy.ChangeStrings` on `UpdateRequest` compose local
replacement or Add/Remove with the preparation read. Nil omits the setting.
Removal deletes the first ordinal match, then additions append without deduplication.
Read/modify/write has no concurrency or ETag guarantee.

Bound Force retains the compiled string form. `BodyIdentity` can explicitly
control a lower-level PATCH body identity; the Set composition uses the selected
read identity. `UpdateResult` preserves preparation, mutation invocation and HTTP
acceptance separately; a 204 receipt is not a persisted-state or client-effect
claim. Mutations never replay automatically. There are no public create/delete
commands for this singleton in the assigned baseline.

All source getters have typed nullable scalars and a materialized file list.
`RawFields` and owned `RawJSON` retain unknown fields; marshaling a received model
returns the original snapshot. Build writes from request settings.

The SDK lab identity has Teams Administrator. Certificate and client-secret
acceptance tests read/list Global, compare all 12 getters with pinned PowerShell,
toggle `EnableVideoMessageCaptions`, verify all omitted getter values including
the populated file block list, and restore the original Boolean with an independent
read. Recovery journals the original value before mutation. CI serializes both
these Global reads and writes. Permanent sanitized acceptance `testdata` supplies
the unit response fixtures. File-list writes, explicit scalar-null persistence,
client propagation, delegated authentication and other clouds remain separately
unqualified; source-backed request tests cover their binding contracts.

See [qualification](../../../docs/implementation/policies/messaging-migration-configuration-lab-acceptance.md).
