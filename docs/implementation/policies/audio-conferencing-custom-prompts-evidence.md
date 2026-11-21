# Audio conferencing custom prompts configuration: pinned evidence

Baseline: MicrosoftTeams 8.1.0. Scope: POL-33 / issue #106. Native implementation
uses the source checkpoint below; complete command qualification remains open.

The assigned management surface is Get (Identity/Filter) and Set (Identity).
The native configuration service exposes GetByIdentity, List, ListFiltered,
PatchByIdentity and explicit UpdateByIdentity (read, local edits, sparse PATCH).
Creation/deletion and assignments require separate evidenced contracts; neither
is present in this command family. Runtime execution remains native Go.

## Reproduction and provenance

The development capture scripts are
[provider capture](../../../tools/contracts/capture-audio-conferencing-prompts-provider.ps1)
and [model capture](../../../tools/contracts/capture-audio-conferencing-prompts-models.ps1).
Both require an explicit module/output path, reject an existing output, check the
8.1.0 manifest and SHA256 of Core, PolicyRp and host assemblies. They load the
actual pinned provider/model/list helper code. No Microsoft binaries or
decompiled source are redistributed.

```powershell
./tools/contracts/capture-audio-conferencing-prompts-provider.ps1 -ModulePath <8.1.0-module> -OutputPath <new-provider-output>
./tools/contracts/capture-audio-conferencing-prompts-models.ps1 -ModulePath <8.1.0-module> -OutputPath <new-model-output>
```

Package SHA256:
`78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.
Exact assembly hashes and host version accompany each fixture. Captured on
2026-11-21 using MicrosoftTeams 8.1.0 package helpers on the local PowerShell
host (the host version is recorded separately).

The provider adapter supplies a synthetic InvocationInfo, a host that refuses
credential acquisition, and a memory-only HTTP sender. It restores the private
client/module-version fields and host in finally. It supplies the factory's
MPACmdlet default. It does not construct/invoke command lifecycle methods,
authenticate, discover endpoints, access a tenant or send network requests.
The model adapter invokes constructors, setters, Newtonsoft deserialization and
the actual Core list helper, with comparators copied as facts from Set's source.
The compiler suppresses only assembly-version unification warning 1701 needed
for the pinned .NET Core assemblies on the development host.

## Observed provider behavior

[Provider fixture](../../../contracts/endpoints/8.1.0/fixtures/audio-conferencing-prompts-provider.json):
15 cases, each one memory send and zero token acquisitions. Four cases fail
as expected: null prompt Id, null package DefaultLocale, a list envelope and 403.

- GET collection: `/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration`.
  GET/PATCH item adds `/configuration/{escaped-identity}`. First slash separates
  scope; the remaining identity stays opaque. This uses the compiled direct
  PolicyRp path, not generated ConfigAPI forest discovery.
- Each request has MPACmdlet `true`, the actual command name and a GUID
  correlation header, without a target-URI header.
- Set helper emits only bound lists plus model Identity, which can differ from
  the route identity. Six nested package arrays and nullable element properties
  remain in bound DTOs. Bound empty lists serialize as `[]`.
- Bound Force false/true serializes as string `False`/`True`; omission is distinct.
- Empty and JSON-null reads return null. Collection results are bare arrays;
  `[]` and null produce the helper's false success flag without a transport
  exception. A `{value: []}` envelope fails. Native shared collection handling
  retains the established collection semantics.

Base Set requires at least one bound list (Force alone is insufficient), performs
Get using Set metadata, calls UpdateProperties on the returned object, then
PATCHes only bound settings. ShouldProcess presentation is separate from native
operations. Writes must retain the established no-replay policy. HTTP acceptance
does not prove persistence or an effective conferencing experience.

## Models, constructors and list edits

[Model fixture](../../../contracts/endpoints/8.1.0/fixtures/audio-conferencing-prompts-models.json):
36 model cases (15 expected failures), 12 constructors (7 expected failures) and
31 list edits (7 expected failures).

- Configuration has Identity plus readonly Prompts and Packages arrays. Null
  arrays are ignored; repeated properties append, including casing variants.
  Null elements survive. Constructor/getter projections emit empty arrays.
- Prompt has Identity, Id, Name, AudioPrompt, TextPrompt, Type and Locale.
  Explicit null Id/Type/Locale fails, but absent properties can stay null in a
  default-constructed/deserialized model. Id requires the source GUID regex,
  including its final-LF dollar-anchor behavior. Empty and future Type/Locale
  strings are accepted. Optional strings preserve null versus empty.
- Package has Identity, Id, Name, DefaultLocale and six readonly recording-prompt
  reference arrays (Inbound/Outbound/Meeting, Start/End). Id has the same regex;
  explicit null Id/DefaultLocale fails, while absent fields can stay null. Each
  array ignores null, appends duplicates, and retains null/empty/future string
  entries. No prompt-reference GUID restriction is inferred from package Id.
- Owned native constructors should require the parameterized constructor's
  non-null fields and preserve accepted empty/future strings. Raw received JSON
  and typed known-field writes must remain separate, so future data is retained
  for inspection without silently entering update requests.
- A null modifier or empty operators clear. Replace preserves duplicates and
  null elements. Add appends without deduplication. Prompt removal deletes the
  first ordinal match using Id, Name, AudioPrompt, TextPrompt, Type and Locale;
  Identity is ignored. All six differing-key cases retain the original entries.
  A null compared prompt fails, while replacement/add can retain null.
- Package removal's comparator throws unconditionally when an existing element
  is examined. Replacement, clear and add work. Removing from an empty collection
  does not invoke the comparator and succeeds. Mixed Remove/Add fails before
  Add when the collection is nonempty. A native edit must preserve this behavior
  and report a local validation error without dispatching PATCH.

## Qualification still required

Native root wiring, typed models, source-model/list tests, provider wire tests,
owned edit tests, partial update checks, examples, exact ledger mappings and
opt-in read-only lab support accompany this evidence. Native string constructor
arguments cannot represent source null inputs; seven representable constructor
cases execute natively while the five null-argument observations document the
source boundary. Both actual client-secret and certificate Global reads returned
403 on 2026-11-21. Filter/list and optional pinned comparison were not reached.
No successful read, mutation, read-back, restoration, prompt playback,
role/consent grant or PowerShell/Go tenant comparison is claimed here.
Arbitrary shell/Newtonsoft coercions, cloud/role/feature availability and all
write qualification remain open.

The full offline `go test -race -count=1 -shuffle=on ./...`, SDK/examples build,
vet, capability ledger validation and its ten tests passed on 2026-11-21.
Both development capture scripts executed successfully; they and the optional
read oracle also passed PowerShell syntax checks. These checks establish native
structural behavior separately from live qualification. The ledger retains
in-progress implementation and metadata-only contract states.
