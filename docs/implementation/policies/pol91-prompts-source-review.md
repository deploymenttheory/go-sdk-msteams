# POL91 Audio conferencing prompts source and branch repair

The directory alias audio_conferencing_prompts hid this resource from the initial
literal resource-directory overwrite sweep. Its old TypedProviderRequests suite
replaced success/decode expectations with controlled400 for every request. The
renamed rejected-provider/filter suites preserve those request proofs; the
preparation suite no longer retains unconstructed later-PATCH callbacks.

The [portable actual-source probe](../../../scripts/discovery/observe-pol91-prompts-models.ps1)
and [recorded evidence](evidence/pol91-prompts-models.json) hash-check Core and
PolicyRp before import. Public Get resolves to Core.GetTeamsAudioConferencingCustomPromptsConfigurationCmdlet,
with BaseTpmGetCmdlet closed over its own Core DTO. Actual closed
ConfigurationNameNormalization returns TeamsAudioConferencingCustomPromptsConfiguration.
This own normalization, not an old400 URL, establishes the resource segment.
The shared compiled selector/header/audience/response source documented in the
[compiled review](pol91-compiled-config-source-review.md#shared-endpoint-audience-response-and-control-evidence)
then applies: selected versus bare collection GET, X-MS-CmdletName/MPACmdlet,
fixed Teams backend and no generated forest discovery. No new endpoint is guessed.

Concrete UpdateProperties calls actual PSListModifierHelper with its actual
Prompts predicate at IL66 and Packages predicate137. The former compares Id,
Name, AudioPrompt, TextPrompt, Type and Locale. The latter explicitly throws an
ArgumentException after formatting ErrorCannotRemoveElementsOf("Packages");
the local probe can fail earlier in uninitialized resource formatting, so its
observed exception type is not asserted as the public host exception type.
Both reject comparisons/removals as recorded, before PATCH.

Twenty actual helper/predicate/serializer observations establish that bound null
and valid empty modifiers clear populated lists, nullable Add/replacement entries
are allowed, and nonnull package Remove from an empty target succeeds. Null
Remove fails even on an empty target: missing-item formatting dereferences null
at generic helper IL317. The native generic edit previously skipped all work
when the target was empty and incorrectly accepted this case. Its new guard
rejects only null removals, after the preparation read; nullable adds/replacements
stay valid. Native ErrInput versus CLR/host exception presentation is explicit.
No Clear behavior/documentation change follows from the earlier disproved null
modifier inference.

Write observations include the exact actual GetNestedBoundParameters expansion
called at SetAsync IL151, then actual resolver185 and serializer213. Omitting
that expansion misleadingly emits only selected nested Identity. After expansion,
full known Prompt/Package DTO fields serialize as the SDK already does; no
speculative serializer fix was made. The read fixture is separately projected
with normal DTO getter serialization, not a selected-write contract resolver.
Its hash and producer are recorded; it is primary-authored, not a captured200
service response, asset upload, or tenant fixture.

New tests execute typed selected/list/filter success, accepted prepared204,
actual later403 with partial/earlier receipts retained, explicit caller full-array
restoration/readback, and both empty-target null-Remove errors retaining Before
and the200 read without PATCH. Nullable Add and nonnull package empty-target
removal remain tested. Full frozen execution is pending. No implementation,
contract or live status is promoted by this checkpoint; #185 remains open.

## Own received-JSON coercion and default correction

The actual own BaseTpmGet path reaches Providers.HttpClientHelper.SendAsync's
state machine and calls plain JsonConvert.DeserializeObject<TPolicy> at IL452.
That differs from strongly typed CLR setter invocation and from a selected-write
contract resolver. The [portable read probe](../../../scripts/discovery/observe-pol91-prompts-read.ps1)
hash-checks Core/PolicyRp and executes that same deserializer over the actual
closed Audio DTO. Its [62 received-JSON vectors](../../../msteams/teams_admin_api/audio_conferencing_prompts/testdata/primary_read_vectors.json)
record exact accepted typed projections and rejected setters/shapes without
management lifecycle, authentication or HTTP.

Numbers/booleans become strings, preserving numeric lexemes exactly: large
integers beyond floating precision, exponent spelling, trailing decimal zeros,
and negative zero. The correction is resource-local and never parses those
strings through floating point. The six readonly reference arrays apply the
same nullable string conversion, append repeated/case aliases and ignore null
arrays without clearing earlier entries. Missing/null/empty package arrays
materialize empty typed lists; the mapped package constructor now does so too.
Top-level Prompt/Package arrays preserve their existing append/null behavior.

Id validation remains strict after string conversion, including the one final
LF allowance; CRLF/double LF still fail. Missing Type/Locale/DefaultLocale retain
null defaults, while explicit null setters fail. Optional Name null stays valid.
Objects/arrays are not converted to strings. RawFields/RawJSON preserve the
original received numeric/Boolean tokens independently of the coerced typed view.
Standard JSON parsing remains the native boundary; no unrelated family decoder
or global map normalization is relaxed.

Canonical full SDK/lab offline selection passed at fixture-corrected clean
`e62c819` / session20189; the earlier51183 run failed only an invalid fixture
provenance path. The separate decoder correction and own SDK GET golden-vector
check require their own frozen execution. No Audio command qualification follows
from the older canonical result or this source probe.

Full affected SDK/lab offline TestUnit/Example passed at frozen
`8afc4aec0123d92138a4ebe6359143c0c5dd152b`, session 99815: 103 passing
test/subtest events, zero fail/skip. Raw runner JSON was captured directly and
its hash/command are in [the execution record](../../../contracts/coverage/evidence/pol91-audio-read-execution.json).
Admission 30.36 GiB, 14 GiB runtime floor, Go 1.27.4 / GOMAXPROCS 2 / -p 1,
owned scratch removed; post-run selected/shared input diffs were zero before
root began its separate nine-family corrections. No Audio qualification or live
status changes follow from this execution checkpoint.
