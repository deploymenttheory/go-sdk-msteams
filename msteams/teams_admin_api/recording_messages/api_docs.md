# Recording and transcription custom messages

`Client.TeamsAdminAPI.RecordingMessages` exposes native Go CRUD with explicit
full-Identity and relative-Id methods. PowerShell is used only by development
oracles and separately enabled lab comparison.

| Go method | Baseline command / parameter set |
|---|---|
| GetByIdentity, List | Get-CsTeamsRecordingAndTranscriptionCustomMessage / Identity |
| GetByID | Get / ParentAndRelativeKey |
| ListFiltered | Get / Filter |
| CreateByIdentity | New / Identity |
| CreateByID, Create | New / ParentAndRelativeKey; bound or generated Id |
| UpdateByIdentity, PatchByIdentity | Set / Identity; composition or native sparse operation |
| UpdateByID, PatchByID | Set / ParentAndRelativeKey; composition or native sparse operation |
| DeleteByIdentity | Remove / Identity |
| DeleteByID | Remove / ParentAndRelativeKey |
| NewLocalization | New-CsTeamsRecordingAndTranscriptionLocalizationCustomMessage / Identity; local only |

Relative Id methods use `Global/{Id}`. Full Identity keeps the source's first-slash
scope/name escaping and the remainder as one opaque segment. List consumes a bare
array, places exact Global first and has no invented continuation. ListFiltered
performs local Identity/Tag wildcard matching; whitespace is literal. Use List
for omission, since a bound empty pattern rejects locally.

NewLocalization requires a nonempty Language, preserves whitespace and unknown
language names, and owns copies of optional text pointers. It exposes all six
text properties, with nullable defaults. No locale whitelist or text limit is
invented. Snapshots retain raw presence/future fields. Model decoding rejects an
invalid non-null Id, explicit null localization lists and explicit null Language
setters, including an earlier bad duplicate key. The source GUID regex accepts
one final LF; an absent Language can remain unset. Requests use known DTO fields,
without accidentally forwarding received future data.

ReplaceLocalizations, ChangeLocalizations and ClearLocalizations are pure local
builders. Clear represents a bound null/empty modifier; nil edit pointers mean
unbound. Removes delete the first ordinal match on all seven language/text fields,
ignoring Identity; additions preserve order and duplicates. Empty operators clear.
Null elements can be retained by replacement/addition; a null comparator operand
fails locally. Builders and Apply own their lists, pointers and raw future fields.

Settings uses policy.Set/Null for Description. Its bound localization slice is a
full replacement array; nil slices encode [], while an explicit null array field
rejects the source model's list rule. UpdateRequest uses local edits instead.
Set requires a setting; Force alone rejects before credentials. New permits empty
settings. Force, when supplied, is the compiled True/False string body value.

The source's write shapes remain distinct. CreateByID serializes null Id; Create
leaves Id unbound and omits that body key; CreateByIdentity inserts the extracted
second identity component as Id metadata. Create returns the generated UUID and
prepared identity even when the submission fails, and does not replay it. All
three issue a collection POST. The optional source Global read is ShouldProcess
presentation, so native Create adds no implicit comparison read.

Update reads with Set command metadata, applies local edits to a copy, and PATCHes
once. It takes body Identity/Id from the read snapshot while retaining the caller's
route. An Id-only Description PATCH omits Identity; binding the localization list
includes Identity via the source's recursive property resolver. Native PatchByID
requires explicit BodyID value/null to avoid guessing it from the route; use
UpdateByID for automatic read preparation. BodyIdentity defaults to the caller's
route in a native sparse call and is overridden by the preceding read in Update.
DeleteByID sends Id rather than Identity. No implicit GET, ETag, read-back, wait,
rollback or mutation replay is added to sparse native writes.

UpdateResult retains Before, ReadResponse, PatchInvoked and HTTPAccepted on partial
failure. A 204 has no fabricated Message; an accepted malformed reply retains HTTP
metadata and returns ErrDecode. Acceptance does not establish persistence or
effective recording behavior. Ordinary model/edit/result formatting is redacted;
raw responses and snapshots remain caller-owned sensitive data.

Requests use the selected Teams audience/origin and the shared compiled-policy
headers, with fresh UUID correlation and no generated forest or Graph lookup.
Synthetic application/delegated modes establish routing, not successful tenant
authorization. Owned ID/full-identity CRUD tests now verify read-back, omitted descriptions and
independent cleanup; see the lab guide for execution requirements and results.
Role/cloud availability and effective user behaviour remain environment-specific. See [unit tests](service_test.go)
and the [lab acceptance plan](../../../docs/acceptance/recording-messages-lab-acceptance.md).

The source null localization **modifier** clears the existing collection. Use
`ClearLocalizations` for that operation; nil `UpdateRequest.Localizations` means
omission. This differs from a null list in a raw model/Settings field, which fails
the model rule. Source no-settings Set returns an error; Force alone is not a
setting. Native Go reports that validation error before acquiring credentials.
