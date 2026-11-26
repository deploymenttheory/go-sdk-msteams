# POL91 Meeting write boundary review — source confirmed, execution pending

The [portable own probe](../../../scripts/discovery/observe-pol91-meeting-write-models.ps1) verifies the pinned Core and PolicyRp assembly hashes before import, then invokes only actual model constructors, the closed PSListModifierHelper, SetMeetingConfiguration's own comparator, GetNestedBoundParameters and BoundParametersContractResolver. No management lifecycle, authentication or HTTP runs.

[Eleven actual outcomes](evidence/pol91-meeting-write-models.json) settle the list boundary: bound null and empty modifiers clear to []; nullable Add and replacement entries remain [null]; default typed elements with unassigned Id serialize Id:null; nonnull Remove on an empty list succeeds without invoking the rejecting comparator; Remove on a populated list fails at that comparator. Null Remove fails even when the list is empty. These are source/helper outcomes, not server acceptance.

The same packet decodes a received context with PublishedName:37, Published:1 and Future:true, then serializes the actual typed DTO through the expanded selected-property resolver. The write has PublishedName:"37", Published:true, every known nullable property and no Future field. Source write serialization therefore uses the typed projection rather than resending raw received JSON. Owned RawJSON remains the read/marshal presentation and must not be repurposed as the write DTO.

The prior native write path had these concrete gaps: Add/Replace and primitive Patch reject source-permitted null members; the raw-preserving context MarshalJSON leaks unknown fields and pre-coercion scalar tokens into writes; Remove is not exposed even for the source-valid empty-list outcome. Existing tests label raw forwarding as source parity. Bound primitive Field.Null currently emits a null array despite the source bound-null modifier producing []; its explicit native/public-command boundary must be corrected or documented independently before qualification. The unassigned Id default is already supported by native Validate, and no blanket Id/default change is needed.

The approved correction now uses a dedicated seven-field typed writer, permits null container members, normalizes bound null/nil replacements to [], exposes ChangeAuthenticationContexts, and preserves helper-stage failures after preparation. Edit constructors/apply own their snapshots as a Go API guarantee, not a claim that the CLR helper clones. The updated tests preserve the standalone nil-element Validate error separately from container acceptance. The corrected combined read/write checkpoint passed as recorded below. Native input validation ordering must remain explicit: a null Remove is a helper-stage failure after preparation, while invalid assigned Id setter values may reject before credentials. No live or offline command status changes follow from this source packet alone.


## Positive UInt32 setter boundary

The first combined decoder check 34406 exposed a real resource-constraint gap:
36 own source cases (and their parent test) rejected converted zero in all six
port/range fields. The generic UInt32 conversion helper remains zero-capable;
only this resource's positiveInteger DTO imposes minimum one. The failed raw
result remains /private/tmp/msteams-pol91-last-compiled-4e7a06d-tests.json.

The [portable write probe](../../../scripts/discovery/observe-pol91-meeting-positive-ports.ps1)
and [24 actual outcomes](evidence/pol91-meeting-positive-ports.json) independently
confirm each field accepts null, 1 and 4294967295 but rejects zero at its DTO
setter with PropertyArgumentOutOfRangeException. Each public command parameter
setter is the eight-byte ldarg.0 / ldarg.1 / stfld / ret body recorded in the
packet; it does not impose that range before processing. Own UpdateProperties
invokes the six DTO setters at IL451/486/521/556/591/626. Shared BaseTpmSet
ProcessRecord obtains the preparation PolicyResult at IL277, invokes
UpdateProperties363, then starts the write task382. Reproduce those selected
method bodies with the [owning-context IL reader](pol04-primary-source-audit.md#portable-reproduction)
on the hash-verified module, without invoking any management lifecycle.

The native primitive Patch rejects zero before HTTP. Prepared Update counts a
bound zero as a setting, obtains the actual preparation, then returns ErrInput
with Before/read receipt and PatchInvoked=false/HTTPAccepted=false. It validates
before setting the mutation-invocation flag. The old zero-positive wire test is
corrected to a positive value and renamed, with its existing unqualified
contract-test reference updated; live references are untouched. Null still
means null for these nullable scalar fields, and no 65535 ceiling is invented.
The corrected write checkpoint passed at c9e096f/73712; see the exact execution record below.


## Corrected frozen execution

At `c9e096f36f268d18f48ad6f8acb0b07bf2ad466b`, full affected SDK/lab/provider,
root policy contracts and labkit verification passed 925 test events with no
test failure or skip. Six test-bearing packages passed; the provider compiled
without test files. The actual primitive/prepared zero guards, all own read
vectors, typed writes/null modifiers and both nine-exchange later-failure and
explicit caller recovery scenarios executed. Targeted gosec/nolintlint passed
with zero issues; source diff was zero and owned scratch was removed.
The [execution record](../../../contracts/coverage/evidence/pol91-last-compiled-execution.json)
preserves exact scopes, commands, raw hashes and the real prior failed run.
No command or live qualification follows from this record alone.
