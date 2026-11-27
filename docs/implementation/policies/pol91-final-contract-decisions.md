# POL91 final reviewed implementation and contract decisions

Approved: restore exactly eighteen previously reviewed commands / twenty-seven sets, and separately complete Audio's two commands / three sets: twenty commands / thirty sets. The ledger applies only these reviewed decisions. Every live record remains unchanged; #185 stays open.

## Shared implementation/source boundaries

The original individual source reviews still supply own public command kinds, all assigned sets/parameters, concrete model/setter/control inventories, normalized resource names, explicit selector/header/audience routes, no-settings and sparse update rules, source list/filter/order and response/error semantics. Their canonical execution was preserved as bounded historical proof when qualifications were withdrawn. The received-JSON gap is now repaired by explicit opt-in wire types, not a global map rewrite.

Own pinned DTO evidence is the immutable 895-row qualified-read-scalar packet plus separate 68-row supplement and ten enum string edges: 973 actual own DTO observations across the nine resources. It includes nullable Boolean/string/UInt64/enum coercion, null/default/setter rejection, exact numeric lexemes and ordered readonly-list behavior. These source observations are local model/deserializer evidence, not management cmdlet executions or tenant responses.

ReadString/ReadBool are independently reviewed: exact numeric text, nullable empty-string Boolean behavior, source Boolean words/whitespace/NUL boundaries and source double conversion. Typed write requests stay separate. Six-family explicit wrappers retain per-occurrence URI setter validation, ordered list accumulation and raw snapshots. ReadUInt64 keeps integer tokens exact and uses actual double/ties-even conversion for float tokens; numeric strings are checked independently. Messaging's received enum wrapper reproduces Int64-to-Int32 wrap; typed-write enum parsing remains bounded and separate. Actual gosec G115 was confirmed and suppressed on that one source-required cast with explanation; targeted gosec+nolintlint then passed. No arithmetic rewrite changes source behavior.

## Revision-bound execution

- Client, Privacy, Education: corrected full check at 2f0b32339f242b76d5fe189171bd844a979daacf, root session 72991. 619 passing test events, no fail/skip; root SDK, three resource SDKs, three lab packages pass; policyrp compile-only. Raw /private/tmp/msteams-pol91-root-scalars-2f0b323-tests.json SHA b974abe75d1e89668ee561ac4ebff1cbef19239f15fa1f99016b428625ad24aa. Root records earlier 60428 failures as obsolete strict negatives/oracle coupling, not a weakened guard; LAB producer validation uses a strict defined alias.
- Other six: full check at af8514b53c1013812eb5a0e98e13a8df0ba64639, session 97339. 661 passing events, no fail/skip; all six SDKs, root policies and corresponding six lab package terminals pass, policyrp compile-only. Raw /private/tmp/msteams-pol91-six-read-fixed-tests.json SHA d61c53733823cc786ce827e712dec418d28c44d69c7e5ca752c61158e6719439. Post-run scoped diff zero, owned scratch removed. Earlier 15393 real readonly-list scalar-ignore failures are preserved. New own-vector tests execute 527 rows plus six actual root-client GET scenarios.
- Audio: full affected SDK/lab check at 8afc4aec0123d92138a4ebe6359143c0c5dd152b, session 99815, 103 passing events/no fail/skip, source-owned 62 read vectors. Raw /private/tmp/msteams-pol91-audio-decoder-tests.json and versioned execution record preserve exact scope/hash. Earlier 20189 canonical success and 51183 fixture-source failure remain distinct.
- 80e0086 is a comment-only source-required narrowing suppression descendant; single-package gosec+nolintlint passes. It is not a new full Go execution claim.

All local executions used Go 1.27.4, GOMAXPROCS=2, -p1/-count1, owned scratch and 14 GiB runtime guard. No TestAcc, live mutation, race or full-module run is asserted.

## Exactly eighteen restoration decisions

| Command | All sets | Complete command behavior basis and corrected read closure |
|---|---|---|
| Get-CsTeamsClientConfiguration | Identity, Filter | Own17-setting model, omitted/selected read and nonempty Filter; retained selected/list/filter/ownership/error cases plus actual own string/Boolean/Identity read profiles and root GET; raw view kept separate. |
| Set-CsTeamsClientConfiguration | Identity | All17 bound keys/types, sparse/null/false/omission, no-settings, prepared204/directtyped200, malformed/cancel/partial receipts and explicit recovery; corrected preparation decoder and complete own-vector execution. |
| Get-CsPrivacyConfiguration | Identity, Filter | Own4 nullable Boolean fields; Global-only authored selected/list/filter/whitespace, shared source order/selection, raw/error behavior and actual own numeric/quoted Boolean/Identity read semantics. |
| Set-CsPrivacyConfiguration | Identity | All4 bound keys, actual resolver null/false/omission, typed/malformed200, prepared204 and partial/recovery; corrected read preparation. Existing403 capture is not a successful capture. |
| Get-CsOnlineVoicemailValidationConfiguration | Identity, Filter | Actual Boolean/string/Identity read profiles, own URI setter null/regex/duplicates including ten actual URI supplement rows; bound-empty Filter correction, retained selected/list/filter and model/error receipts. Existing regex bytes were verified correct; no escaping rewrite. |
| Set-CsOnlineVoicemailValidationConfiguration | Identity | Both managed fields, sparse binding/no-settings/Force/URI constraints, typedPatch200/prepared204, malformed/read denial/cancel/partial and explicit recovery; source-correct preparation decoder. |
| Get-CsTeamsEducationConfiguration | Identity, Filter | All3 opaque strings, own getters/types/constraints, omitted/selected/list/filter, typed/raw ownership/errors and actual own scalar read profiles. |
| Set-CsTeamsEducationConfiguration | Identity | All3 source bound keys/plain setters, null/empty/omission, no-settings and control rules, actual typed/malformedPatch200 and retained-read/prepared204 with failure receipts; updated preparation codec. |
| Get-CsTeamsUpgradeConfiguration | Identity, Filter | Complete3-setting source model/defaults, own nullable/string/Identity read observations, retained selected/list/filter/readback and raw/error/guard handling. |
| Set-CsTeamsUpgradeConfiguration | Identity | All3 selected scalar settings/null/false/opaque/Force wire proofs, source bound-key/no-settings semantics, actual accepted204 and independent caller readback; corrected own-model preparation. |
| Get-CsTeamsSipDevicesConfiguration | Identity, Filter | Complete BulkSignIn source/string/null/default inventory, own received coercion/Identity profiles, retained get/list/filter/snapshots/errors and strict guard boundaries. |
| Set-CsTeamsSipDevicesConfiguration | Identity | Single bound string and null/empty/omitted/Force rules, accepted prepared204, raw snapshots/no replay and zero-request no-settings; corrected source preparation decoder. |
| Get-CsTeamsMultiTenantOrganizationConfiguration | Identity, Filter | Complete CopilotFromHomeTenant string/source inventory, own received scalar/Identity profiles, retained get/list/filter/owned snapshots/errors; no effective Copilot behavior claim. |
| Set-CsTeamsMultiTenantOrganizationConfiguration | Identity | Single bound string/null/empty/omission and concrete guard, prepared204/source control/receipt behavior plus corrected preparation codec. |
| Get-CsTeamsTenantAbuseConfiguration | Identity, Filter | Own10-setting model,7 nullable UInt64,2strings and nullable-entry readonly list; complete source received scalar/numeric/list/Identity profiles, exact/full-width/raw/order/default behavior, authored typed get/list/filter and overflow error receipts. |
| Set-CsTeamsTenantAbuseConfiguration | Identity | All10 fields/setters/bound keys, integer/null/zero/opaque/control request proofs, actual list modifier/null/ownership rules, prepared204/later403/earlier receipt preserved and caller restoration/readback; corrected preparation model. |
| Get-CsTeamsMessagingConfiguration | Identity, Filter | Complete12-setting model, actual enum/string/Boolean/readonly-list profiles and raw ownership; retained selected/filter plus direct omitted-Identity List, nullable entries/alias order/error receipts. |
| Set-CsTeamsMessagingConfiguration | Identity | All12 selected keys/setters/control semantics; source enum binder versus typed-write/received enum distinction explicit; source list clear/remove/add/nullable ownership, accepted204/explicit readback, malformed/read/no replay behavior and corrected preparation codec. |

## Separate Audio two-command / three-set decisions

| Command | Sets | Source and implementation closure |
|---|---|---|
| Get-CsTeamsAudioConferencingCustomPromptsConfiguration | Identity, Filter | Own closed normalization and Core/BaseTpm source, not a controlled400 route inference; actual model/source getter projection, typed source-authored selected/defaultList/filter/whitespace, all Prompt/Package fields, six nullable reference arrays/defaults, 62 own read vectors and raw/typed distinction. No claimed successful tenant read. |
| Set-CsTeamsAudioConferencingCustomPromptsConfiguration | Identity | Both own list-modifier bindings/actual helper predicates/expanded serializer; valid null/empty clears, nullAdd/replace allowed, nullRemove-empty guard after read, package nonnull empty removal versus comparison failure; complete typed direct200/malformed receipt, prepared204 then actual later403 and explicit caller restoration/readback. Source CLR versus native ErrInput presentation boundaries explicit. |

The 20-row helper/predicate/serializer probe includes actual GetNestedBoundParameters expansion at SetAsync151 before resolver 185, distinguishing read-mode DTO projection from sparse write-mode projection. No serializer change was inferred from an incomplete probe. Audio's directory alias caused the old denied-vector attribution hole; those are now correctly named request proofs, with real new success/failure branches rather than overwritten expectations or unconstructed patch callbacks.

## Limits and finite remainder

No live state changes. Minimum roles, tenant eligibility, real nondefault setting/client effects, oracle parity/restoration and cloud-specific availability remain separately unqualified. HTTP2xx acceptance and authored recovery are not tenant persistence. Canonical public field type corrections ([]*string for the two proven readonly lists), native path/early-error/raw-view distinctions and standard-JSON parsing boundary stay explicit. No conditional legacy/remoting branch is promoted.

With these twenty decisions approved, 19 commands / 29 sets remain: 7 compiled / 11 sets across FirstParty templates, GuestMeeting, Meeting, MeetingTemplate Get; 12 conditional / 18 sets with separate remote-source contracts. No child merge or source symbol closes #185 or QUAL01.

## Final enum closure (supersedes the prior Messaging-only final pointer)

Numeric-string trailing NUL is supported only in the checked numeric parse branch;
named NUL and numeric overflow still reject. Final full Messaging SDK/lab/root
policies execution at 201d6cbb102a9d7d59d4abacd1dc2fbc0871fd57, session 48041,
really replays 895+68+10 fixture inputs (Messaging count 205). 253 test/subtest
passes, zero fail/skip, all three package terminals PASS. Raw
/private/tmp/msteams-pol91-enum-all-edges-tests.json SHA
daec39afde8e26a71ec569f7054250c9a6fee306b193acfd32a3a4bb7b52ce81.
Targeted gosec+nolintlint then exit 0; scoped post-run diff zero and owned scratch
removed. The earlier 269b5b5/13260 full check also passed but contained only the
new dedicated NUL assertions, not the ten direct source-row replay added later.
The source 4af39 commit was metadata-only and is not mislabeled its execution.

The aggregate own source scope is 973 observations across nine families plus
Audio 62. Full six-family af8514b/97339 remains the previous coherent baseline;
final enum closure is the separately bound changed-family descendant. The approved ledger decisions restore only these reviewed offline qualifications;
all live states remain unchanged.
