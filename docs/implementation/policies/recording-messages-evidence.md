# POL-30 recording messages evidence

Issue #103 owns five baseline commands and ten exact parameter sets: Get (Filter,
Identity, ParentAndRelativeKey), New/Set/Remove (Identity and ParentAndRelativeKey),
and the local localization builder (Identity). This source/capture checkpoint
establishes no native service, contract pass or tenant qualification. Coverage
states and the immutable runtime snapshot remain unchanged.

The accepted 8.1.0 package SHA256 is
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Core, PolicyRp and host assembly hashes are checked by both development scripts
and retained in the fixtures. The evidence comes from pinned Core command/model
types, their compiled base classes and actual helper calls.

## Protocol and selectors

The management commands derive from BaseTpmGet/New/Set/Remove with the
TeamsRecordingAndTranscriptionCustomMessage model. The resource uses direct
compiled ConfigAPI, without generated forest discovery or Graph. Collection
GET/POST use `/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage`;
item GET/PATCH/DELETE append `/configuration/` and the source's escaped identity.
The existing compiled origin/audience, MPA/command/correlation headers and
non-replayed mutation transport apply.

PrepareForScopeCheck maps an Id-only call to Identity `Global/{Id}`. A supplied
Identity containing `/` extracts only its second split component as Id metadata,
while the route keeps the entire identity and escapes the remainder after the
first slash as one segment. The New command initializes Id with a new GUID.
The oracle checks that default, then replaces it with a fixed fixture ID before
the unbound-auto preparation case, for deterministic capture. This does not
exercise the PowerShell parameter binder or prove an entire automatic flow.

The sixteen [provider observations](../../../contracts/endpoints/8.1.0/fixtures/recording-message-provider.json)
preserve distinctions that a generic CRUD body would erase:

- Create via Identity injects extracted Id metadata; via a bound Id it serializes
  the new DTO's null Id. With an unbound generated Id, Id is absent from the body.
  All three retain the prepared Identity. A non-GUID route Id is accepted by this
  helper path; it is not assigned through the model's validating Id setter.
- Set via Identity includes the read-model Identity. Set via Id includes the
  read-model Id, and includes Identity when the localization collection is bound:
  the recursive bound-property resolver brings in the child's Identity property.
  A description-only Id PATCH omits Identity. Route selection stays independent.
- Delete via Id sends Id and bound Force; Delete via Identity sends Identity and
  bound Force. Required-not-bound scope metadata is only injected by Create.
- A bound collection serializes full known localization DTO fields, including
  null optional strings and nested Identity. No server Add/Remove object is sent.
  New's empty-settings helper call is accepted; Set checks that at least one
  setting is bound. Force alone is not a Set setting.

[The provider oracle](../../../tools/contracts/oracle/RecordingMessageProviderCapture.cs)
installs a memory client before constructing any command, checks that it remains
installed, and invokes only actual scope preparation and direct HttpRequestHelper
methods. The host refuses credentials. Each observation has one memory send, zero
token requests, a GUID correlation ID, MPACmdlet true and no target-URI header.
Client/version/host state is restored. No BeginProcessing/ProcessRecord, command
runtime, telemetry, credentials, generated discovery or network runs. The New
base's optional Global comparison read is for ShouldProcess presentation; native
CRUD must distinguish that presentation behavior from resource state preparation.

## Models and local composition

[Thirteen model and nine list observations](../../../contracts/endpoints/8.1.0/fixtures/recording-message-models.json)
use actual pinned constructors/setters/Newtonsoft projections and the Core
PSListModifierHelper extension with the source's seven-field equality predicate.

The message has Identity, Id, Description and a localization list. Its default
list is empty; explicit null fails the list setter. Id allows null, rejects empty
or malformed strings, and accepts a canonical GUID with final LF because of the
source regex's `$` anchor. Description's apparent length checks do not throw;
the actual setter accepts the captured 2,000-character value. No artificial
1,040-character maximum should be inferred from those discarded constants.

A localization has Identity, Language, InitiatorImplicit, ParticipantImplicit,
InitiatorExplicit, ParticipantExplicitRequested, ParticipantExplicitProvided and
AgreementDialogue. Language rejects null through the setter, but the parameterless
deserializer can leave it absent. The direct constructor accepts empty, whitespace
and an unknown language string. The public builder's mandatory Language parameter
has separate ValidateNotNullOrEmpty binding; no supported-locale whitelist exists
in the observed model. Optional text fields preserve null and empty distinctions.

Null modifiers, empty operators and empty replacements clear the current list.
Replacement and addition preserve order and duplicates; the captured replacement
also retains a null item. Each Remove
deletes the first exact match, then Add appends without deduplication. Equality is
ordinal and includes all seven language/text fields; Identity is excluded. Thus
two entries with different Identity values can match, while en-US and en-us differ.
Null-element comparisons and mixed/nested shell conversions need further coverage.

## Native implementation sequence

Implement the typed snapshots and pure localization/list builders first, including
captured setter rules, received presence/future data and request DTO projection.
Then expose native Create/Get/List/Update/Delete with explicit Identity/Id entry
points and retained generated-ID preparation, rather than silently normalizing
their distinct wire bodies. Add explicit read/modify/PATCH composition for list
edits; keep sparse native operations available, with preparation and HTTP
acceptance separate from persisted/effective state.

Map all ten sets to compiling symbols and exact request/model/builder tests,
including no-HTTP builders, no mutation replay, failure/cancellation and partial
preparation. Add a bounded opt-in lab read and separate owned write/read-back/
cleanup scenarios. Full shell coercion, presentation, cloud/permission support,
successful tenant writes and effective recording/transcription behavior remain
unqualified. No lab request or permission change has run for this family.
