# POL91 template command decisions — pending independent review

This packet proposes exactly three offline implementation/contract decisions,
covering five exact sets. No ledger status has changed. With independent
approval, POL91 would reach 23 commands / 35 sets, leaving 16 commands / 24 sets.
Every live status, reason, evidence link and live test reference stays separate.
Issue #185 remains open.

| Command | Complete sets | Proposed decision and behavior basis |
|---|---|---|
| Get-CsTeamsFirstPartyMeetingTemplateConfiguration | Identity, Filter | Implemented / contract passed: own direct compiled type and normalization; omitted collection versus selected/scoped identity; effective nonempty Filter and local wildcard/Tag selection; complete root/template/option models, own scalar conversions, defaults, ordered list aliases, per-occurrence setters, owned raw snapshots and request/error/context boundaries. Permanent populated primitive responses exercise actual typed getter branches. |
| Set-CsTeamsFirstPartyMeetingTemplateConfiguration | Identity | Implemented / contract passed: exactly Description:String and TeamsMeetingTemplates:PSListModifier<Template>; both concrete bound-key guards, sparse null/empty/omission/control semantics; own Name-keyed local helper and expanded serializer; complete prepared GET/PATCH behavior, receipt/error/acceptance separation, no automatic replay and explicit caller-selected restoration/readback after both later denial and accepted malformed data. |
| Get-CsTeamsMeetingTemplateConfiguration | Identity, Filter | Implemented / contract passed: own direct compiled type/normalization and sole exported Get surface; selected/omitted/list/filter, complete nullable/nested model inventory and corrected own read profile; typed/raw ownership, effective binder guards, backend/error/decode/context behavior. No configuration write surface is invented. |

## Source completeness

The [source review](pol91-template-source-review.md) and its portable
[own read packet](evidence/pol91-template-read-models.json) trace both public
Get implementing types and closed BaseTpmGet model arguments. FirstParty Set
is likewise a direct closed BaseTpmSet command. Own normalization is exactly
TeamsFirstPartyMeetingTemplateConfiguration or TeamsMeetingTemplateConfiguration.
There is no conditional ConfigAPI/remoting dispatch assigned to these commands.
The public metadata records every set/default and effective parameter type and
validation attribute. Identity is the default Get set; Filter is distinct, and
both inherit ValidateNotNullOrEmpty. Set has the one Identity set.

Both root DTOs contain Identity:String, Description:String and a getter-only
List<TeamsMeetingTemplateType>. Each template contains Identity, Description,
Name and Category strings and getter-only List<TeamsMeetingOption>; each option
contains Identity, Name and Value strings plus nullable IsLocked/IsHidden.
The 280 observations trace every own field under both public model declarations,
including the shared nested Core types. They prove exact scalar-to-string and
nullable-Boolean conversions, missing versus explicit-null required Name,
constructor defaults, nullable list members, primitive-ignore/object-reject,
case-alias accumulation and retained prior arrays. No constructor/default bug
was inferred: existing materialized list defaults already matched the source.

FirstParty's own AreParametersSpecified tests bound TeamsMeetingTemplates and
Description keys, independently of null/empty values; Force alone is not a
setting. UpdateProperties applies the actual generic modifier with its own
Name comparator, then assigns bound Description. The portable
[15-outcome helper packet](evidence/pol91-firstparty-template-models.json)
executes that closed helper/predicate, GetNestedBoundParameters and the actual
resolver. It establishes first ordinal match removal, case distinction,
unmatched and null-name behavior, empty/null clears, nullable Add/replacement,
null Remove and existing-null-entry failures, plus expanded root Description
serialization. Every typed template/option property is written once the list is
bound; unknown response fields are not copied into writes. The existing native
edit/resolve implementation already matched; the production correction is
limited to proven received-JSON behavior.

The [shared compiled endpoint/auth/result/control evidence](pol91-compiled-config-source-review.md#shared-endpoint-audience-response-and-control-evidence)
and [PolicyRp source investigation](pol04-primary-source-audit.md#compiled-bodies-and-grant-casing)
apply only after closing each own model and normalized name. These commands
use direct GET/PATCH under `/Skype.Policy/configurations/{normalized}` with
selected `/configuration/{identity}`, TeamsAdmin audience, MPACmdlet and exact
command metadata; they do not use generated forest discovery. Common source
Get selects collection on omission, exact Global first with stable group order,
and the Identity/Tag wildcard rule. No paging envelope is invented.
Shared HttpClientHelper uses broad 2xx success, followed by its typed decode;
empty successful bodies and nonempty malformed bodies are distinct. Public
Set host output does not promise an updated DTO: native direct Patch is an
explicit primitive, while prepared Update retains Before, ReadResponse, After,
PatchInvoked and HTTPAccepted. Shared no-settings WriteError, bound-key sparse
serialization, Force metadata and ReplayNever remain source-backed.

Native encoded-path safety and early errors, explicit context/confirmation,
typed primitive output and standard JSON parsing remain documented API
boundaries. Effective public validation is not inferred from native path
helpers. Shared compiled String binder evidence supports null/empty preservation;
its existing Client/Validation literal probes are not mislabeled as direct
FirstParty invocations. FirstParty own Description metadata/setter/bound-key
and actual serializer observations supply the concrete model basis. No new
server value, length/enum/GUID constraint or permission claim is invented.

## Actual success, failure and recovery execution

[Full decoder execution](../../../contracts/coverage/evidence/pol91-template-read-execution.json)
is frozen at `2d72bcd4e64b6e9413e3eae3a2744f7f4bca7f37`, session 6554:
416 passing events, zero failures/skips, both full SDK packages, shared labkit
and both corresponding lab packages pass. Both portable producers reran from
repository paths with exact parsed artifact and producer-hash equality.
The first sandbox cache denial occurred before compilation and is retained as
an admission failure, not a test result. The authorized run is the proof.

The independent root review reran both portable source probes and verified the
raw execution hash and all five package terminals. Canonical PowerShell oracle
validation remains strict through separate recursive producer types before
SDK setter validation; newly accepted backend scalar coercions do not weaken
producer rejection. The two prior Description:false malformed vectors now use
actual source-rejected objects rather than falsely rejecting a valid read.

[Full prepared-failure execution](../../../contracts/coverage/evidence/pol91-firstparty-composition-execution.json)
is separately frozen at `c3408096e986c3c4bcb5dd6f367d2e3b6047820c`, session
66609: 204 passing events, zero failures/skips, full FirstParty SDK, its lab
package and the complete root policy contract package pass. Both new branches
really execute through the root SDK: an earlier prepared 204, later prepared
PATCH403 or malformed PATCH200, independent same-Global MeetingTemplate read,
explicit original-null Description restoration and independent readback.
Assertions distinguish retained typed preparation and raw responses,
PatchInvoked, actual HTTPAccepted, absent After, separate earlier receipts and
owned snapshots. The exchange list proves there is no implicit retry/rollback.

Read bytes retain the permanent primitive acceptance provenance documented in
[FirstParty qualification notes](first-party-meeting-template-lab-acceptance.md)
and [meeting-template qualification notes](meeting-template-policies-lab-acceptance.md).
The compound sequence, denial replay, malformed accepted reply and restoration
are authored offline. Neither a reused readback nor a 204 establishes current
persistence, transactionality or effective meeting behavior. Both executions
used Go 1.27.4, GOMAXPROCS=2, `-p 1 -count=1`, an owned removed scratch directory
and the 14 GiB runtime guard. Scoped post-run source diffs were zero. No TestAcc,
race, full-module sweep, tenant request or permission change occurred.

## Finite remainder and live limits

Approval would add only these three commands / five sets to the prior reviewed
20 / 30. Four compiled commands / six sets remain: GuestMeeting Get/Set and
Meeting Get/Set. Twelve conditional commands / eighteen sets retain separate
unsettled remote-source branches. Their existence is not an exclusion, and
this packet does not settle those contracts.

Minimum role, live eligibility, nonempty custom meeting templates, first-party
template-list writes, effective client behavior, delegated access and other
clouds remain separately qualified or unavailable under the existing live
records. Historical selected-operation successes are not blanket live parity.
No child merge, status mapping or offline profile closes #185 or QUAL01.
