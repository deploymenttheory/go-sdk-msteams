# POL-30 recording messages evidence

Issue #103 owns five baseline commands and ten exact parameter sets: Get (Filter,
Identity, ParentAndRelativeKey), New/Set/Remove (Identity and ParentAndRelativeKey),
and the local localization builder (Identity). Native CRUD, explicit Set
composition and local builders now implement all ten mapped sets. The coverage
ledger remains in-progress and metadata-only, without claiming full command or
tenant qualification. The immutable runtime snapshot remains unchanged.

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

## Native implementation and qualification

TeamsAdminAPI.RecordingMessages exposes typed snapshots, pure localization/list
builders and Create/Get/List/Update/Delete with explicit Identity/Id entry points.
Create retains its chosen UUID/identity if submission fails. Update performs an
explicit read/local edit/PATCH composition; sparse native PATCH remains available.
Preparation and HTTP acceptance stay separate from persisted/effective state.
PatchByID requires an explicit BodyID rather than guessing the read-model value
from a potentially different route ID. UpdateByID supplies it from the read.

All ten sets map to compiling symbols and exact request/model/builder tests.
Tests compare fifteen deterministic provider cases in application/delegated
synthetic modes, all thirteen model observations and nine list observations.
The injected auto-ID fixture is supplemented by a native memory-sender test
checking fresh UUIDs, unbound Id omission, separate correlation IDs, retained
selection/acceptance on denial or malformed success, and exactly one send per
submission. Compositions keep a read-model Id that differs from the requested
route, preserve Before/response metadata, omit unbound fields and project known
DTO fields without forwarding future response data. No-HTTP builders, local
guards, status/redaction, mutation replay prevention and all seven equality fields
are covered. Structural reports retain coverage_claim false and live not-run.

On 2026-11-21, actual collection reads using the existing dedicated lab SPN
returned resource HTTP 403 in both client-secret and certificate modes. Bounded
identity reads and optional pinned PowerShell comparison were not reached. No
payload was retained and no message, policy, assignment, role or consent changed.
The Get ledger row records unavailable live qualification; writes remain not-run.
The read harness can compare at most five selected records with a tenant-bound,
hash-checked PowerShell projection when access is available.

Full shell/Newtonsoft coercion, presentation, mixed/nested shell collection
conversions, cloud/permission support, successful tenant writes/read-back/cleanup
and effective recording/transcription behavior remain unqualified under #103.
Native response fields require their Go types; raw presence is retained rather
than presenting every CLR default. Conservative shared identity/destination
validation is not a claim to accept every shell string. Future write acceptance
needs owned records and an authorized journal/cleanup plan.

Local verification on 2026-11-21 passed the full offline race/shuffle suite,
SDK/examples build, vet, scoped lint with zero issues, ledger validation and its
ten validator tests. All three recording PowerShell scripts parse. Exact-head
and cross-platform structural artifacts retain separate qualification flags;
these checks do not promote any command to a full parity or live pass.
