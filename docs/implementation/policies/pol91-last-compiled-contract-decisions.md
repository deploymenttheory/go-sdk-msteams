# POL91 last compiled command decisions — pending independent review

This packet proposes exactly four offline implementation/contract decisions,
covering six exact sets. It does not change the ledger. Approval would bring
POL91 to 27 commands / 41 sets, leaving twelve conditional commands / eighteen
sets. All live states, evidence and live-test references remain separate.
Issue #185 and full qualification remain open.

| Command | Every assigned set | Proposed implementation / contract decision |
|---|---|---|
| Get-CsTeamsGuestMeetingConfiguration | Identity, Filter | Implemented / passed: actual exported compiled type, own normalization, omitted/default collection versus selected identity, complete eight-setting model, source-compatible nullable/scalar projection and raw ownership, local filters, errors and context/path guards. |
| Set-CsTeamsGuestMeetingConfiguration | Identity | Implemented / passed: all eight bound keys/types, concrete presence guard, sparse null/false/empty/control handling, prepared GET/PATCH, empty acceptance versus typed output, later denial and accepted malformed response with retained preparation/receipts and explicit caller recovery. |
| Get-CsTeamsMeetingConfiguration | Identity, Filter | Implemented / passed: actual compiled type/normalization, all 21 settings and seven context properties, positive UInt32 and context-Id constraints, scalar coercion/defaults/alias population, selected/list/filter/order, raw ownership, errors and native input/context boundaries. |
| Set-CsTeamsMeetingConfiguration | Identity | Implemented / passed: all 21 selected keys/types, concrete presence guard, actual modifier/expanded serializer, seven-field typed wire projection, nullable container/clear behavior, positive-port setter timing, full prepared partial/error/acceptance receipts and caller-selected reconciliation/restoration/readback. |

## Complete concrete model and parameter inventory

The [725-row own read packet](evidence/pol91-last-compiled-read.json), produced by
[the portable observer](../../../scripts/discovery/observe-pol91-last-compiled-read.ps1),
records all four public implementing/base types, sets/defaults, effective
parameter types/validators, exact normalized names and every own model property.
These are local model/source observations, not tenant captures.

GuestMeeting has exactly six Nullable<Boolean> settings: AllowIPVideo,
AllowMultipleScreenshare, AllowMeetNow, AllowTranscription,
AllowParticipantGiveRequestControl and AllowExternalParticipantGiveRequestControl.
Its two String settings are ScreenSharingMode and LiveCaptionsEnabledType.
Identity is a separate received String. No nonnullable Boolean default or local
setting enum is synthesized. Historical generated-proxy declarations with six
settings are not the actual exported runtime surface; the own compiled eight-
setting declarations and observed models settle that distinction.

Meeting has the following complete twenty scalar settings plus its context list:

| Own types | Every setting |
|---|---|
| String | LogoURL, LegalURL, HelpURL, CustomFooterText, FeedbackSurveyForAnonymousUsers, ReportMeeting |
| Nullable<Boolean> | DisableAnonymousJoin, DisableAppInteractionForAnonymousUsers, EnableQoS, EnableAttributedTranscripts, EnableGraphTranscriptAccess, EnableWorkIQTranscriptAccess, ClientMediaPortRangeEnabled, LimitPresenterRolePermissions |
| Nullable<UInt32>, source positiveInteger | ClientAudioPort, ClientAudioPortRange, ClientVideoPort, ClientVideoPortRange, ClientAppSharingPort, ClientAppSharingPortRange |
| Getter-only List<PublishedEntraAuthenticationContext>; public PSListModifier | PublishedEntraAuthenticationContexts |

Each context has six String getters (Identity, Id, PublishedName,
PublishedDescription, EntraName, EntraDescription) and Nullable<Boolean>
Published. Missing/unassigned Id is a constructor default; explicit-null or
invalid assigned Id invokes the rejecting setter. The own pattern accepts c1
through c99 with one optional final LF. Native validation compares without that
one LF but retains the exact original request string; CRLF, repeated LF and
out-of-range names reject. No general whitespace/GUID normalization is added.

Own runtime deserialization establishes string and nullable Boolean conversions,
UInt32 numeric/string/Boolean/ties-even behavior, complete defaults and setter
errors. Only this resource's six port getters reject converted zero; the generic
UInt32 reader remains zero-capable. Null is allowed, and 1 through UInt32 max is
valid; a 65535 ceiling would be invented. Context arrays retain nullable entries,
append case aliases, ignore null/primitive assignments and reject objects.
Per-occurrence Id validation cannot be hidden by a later alias. Raw snapshots
retain original tokens/unknown fields independently of typed values.

## Versioned exact control flow and write rules

The new [35-method control packet](evidence/pol91-last-compiled-controls.json)
and [portable owning-context IL observer](../../../scripts/discovery/observe-pol91-last-compiled-controls.ps1)
version the previously private guard/order observations. It includes both
concrete AreParametersSpecified and UpdateProperties methods, every managed
public parameter setter and both closed BaseTpmSet ProcessRecord bodies. Assembly
and producer hashes plus method IL hashes are retained. Reflection only reads
method bodies; it invokes no management lifecycle, authentication or HTTP.

Guest's presence guard has eight keys at IL11/37/63/86/109/132/155/178;
Meeting has 21 at IL11/37/63/89/115/141/167/193/219/245/271/297/323/349/375/401/424/447/470/493/516.
Each uses BoundParameters.ContainsKey, not nullable HasValue/current property
values. Null, false, empty and bound zero therefore count as selected inputs;
Force alone does not. All source parameter setters are stores. The DTO setter
constraints apply when UpdateProperties assigns the selected values.

The [24-row positive-port writer packet](evidence/pol91-meeting-positive-ports.json)
independently calls each actual DTO setter and selected-property serializer for
null, 0, 1 and max. All six public port setters are eight-byte plain stores.
UpdateProperties calls the six DTO setters at IL451/486/521/556/591/626.
Shared BaseTpmSet obtains PolicyResult277, invokes UpdateProperties363, then
starts the write task382. Prepared native zero validation therefore runs after
GET, preserving Before/read receipt with PatchInvoked=false and no PATCH;
primitive Patch rejects zero before HTTP. Null remains different from omission.

The [eleven own modifier/serializer outcomes](evidence/pol91-meeting-write-models.json)
and [write review](pol91-meeting-write-source-review.md) settle nullable Add and
replacement, bound-null/empty clear to [], default typed null Id, nonnull Remove
on empty lists and rejecting populated/null Remove. Native ChangeAuthenticationContexts
implements remove-before-add; helper-stage failures occur after preparation.
Owned edit constructor/apply snapshots are a native Go guarantee, not a claim
that the CLR helper clones.

Source DTO serialization writes all seven known context fields. It converts
received PublishedName:37 to "37" and Published:1 to true and drops unknown
Future data. The native dedicated typed writer matches that projection while
keeping RawJSON for inspection. It does not forward raw response bytes into
writes. Primitive bound-null/nil replacements now send [], correcting the old
null-array wire claim; no backend JSON-null acceptance is inferred.

## Endpoint, selection, result and native boundaries

All four commands are direct closed Core/PolicyRp operations with own normalized
TeamsGuestMeetingConfiguration or TeamsMeetingConfiguration. They do not inherit
a conditional remoting blocker from old generated proxy declarations.
The [shared compiled source evidence](pol91-compiled-config-source-review.md#shared-endpoint-audience-response-and-control-evidence)
and [route/auth investigation](pol04-primary-source-audit.md#compiled-bodies-and-grant-casing)
apply after closing these concrete types: direct GET/PATCH under
`/Skype.Policy/configurations/{normalized}` with selected `/configuration/{id}`,
fixed TeamsAdmin destination/audience, MPACmdlet and exact public command metadata.
There is no generated forest target or invented paging/continuation envelope.

Get defaults to Identity, whose omission selects the bare collection; selected
identity uses the encoded path. The distinct Filter set requires a nonempty
string, with local Identity/Tag wildcard selection and literal whitespace.
The shared source preserves exact-Global-first stable group order. Set Identity
is mandatory and has no Global default. Native additional path guards and early
errors are explicit safety boundaries, not attributed to the public binder.

Common success is broad 2xx before typed decoding. Empty 204 receipt, typed data,
malformed-body ErrDecode and effective persistence are distinct. Public host
output is not a returned-updated-model promise; direct Patch is an explicit
native primitive, while Update exposes Before/ReadResponse/After/PatchInvoked/
HTTPAccepted. Force retains its source string metadata. Shared no-settings
WriteError is represented by native ErrInput. Context/confirmation, typed Go
arguments and PowerShell warning/output/error ordering remain explicit native
presentation boundaries. No automatic retry/rollback or effective setting effect
is claimed; mutation replay stays disabled.

## Actual complete affected execution

The [frozen execution record](../../../contracts/coverage/evidence/pol91-last-compiled-execution.json)
binds the corrected source to `c9e096f36f268d18f48ad6f8acb0b07bf2ad466b`, session
73712. All 925 test/subtest events passed, with zero test failures/skips; six
full test-bearing SDK/root/lab packages passed and the internal provider compiled
without tests. It covers all 725 own read rows, actual root-client reads, the
context typed/null/ownership/helper branches and all six primitive/prepared zero
stage checks. Targeted gosec/nolintlint session 79196 passed with zero issues.
Both owned scratch directories were removed and scoped source diff was zero.

The first combined 4e7a06d/34406 result was a real missing resource-positive read
guard: 36 own zero-conversion cases plus their parent failed. That artifact and
cause remain preserved. The corrected read wrapper and subsequent write-stage
correction are not mislabeled a fixture failure or an unchanged-component pass.

The new nine-exchange root scenarios execute both families through the actual
SDK. They retain an earlier Meeting prepared204, Guest preparation and later403
or accepted malformed200, raw receipts and actual acceptance flags. An explicit
Guest reconciliation read supplies an independently selected observed state;
the caller then restores/re-reads Guest and restores empty Meeting footer/null
HelpURL with independent readback. Earlier typed/raw snapshots remain unchanged.
Those compounds are authored offline; retained primitive read/update/readback
bytes keep their own [Guest](guest-meeting-configuration-lab-acceptance.md) and
[Meeting](meeting-configuration-lab-acceptance.md) provenance. A replayed readback
or accepted response is not current tenant persistence, atomicity or live parity.

All local checks used Go 1.27.4, GOMAXPROCS=2, -p1/-count1 and a 14 GiB runtime guard.
No local TestAcc, authentication, role change, tenant request, race or full-module
sweep ran. Current-head CI remains separately required. This decision does not
promote any live record or erase minimum-role, cloud/delegated, effective client
or full tenant lifecycle limits.

## Finite remainder

These four commands / six sets are the final compiled group. Approval would add
them to the prior 23/35, reaching 27/41. Twelve conditional commands / eighteen
sets remain. Their modern and remote contracts need their own source/behavior
work; legacy absence is neither exclusion nor proof of full parity. This packet
makes no conditional qualification inference and does not close #185 or QUAL01.
