# POL91 meeting-template compiled frontier — pending command decisions

Scope: FirstParty Get (Identity, Filter), FirstParty Set (Identity), MeetingTemplate Get (Identity, Filter). No ledger or live promotion is proposed before complete source/control and execution review.

Own model evidence: [the portable template-read probe](../../../scripts/discovery/observe-pol91-template-read-models.ps1) verifies pinned Core/PolicyRp DLL hashes before import and records public implementing/base types, effective validation attributes, all parameter sets/default sets, own normalization and recursively declared DTO properties. Its [280 own observations](evidence/pol91-template-read-models.json) exercise all configuration/template/option scalar fields and both getter-only list properties under each public Get model. These are local data-model observations, not tenant captures.

The own read path accepts numeric/Boolean string conversions and nullable Boolean conversions; exact numeric lexemes survive string projection. Template/Option Name may be absent (default null) but explicit null invokes its rejecting setter. Missing/null readonly lists project as []; array aliases append and null aliases retain the prior items. Whole primitive list assignments are ignored; object assignments reject. Native defaults and per-occurrence Name rejection already match, so only explicit scalar read adapters and primitive-ignore list dispatch require changes. Raw snapshots and typed writes remain separate.

[FirstParty own write probe](../../../scripts/discovery/observe-pol91-firstparty-template-models.ps1) uses the actual PSListModifierHelper closed type, its own UpdateProperties Name comparator, GetNestedBoundParameters and BoundParametersContractResolver. [Fifteen outcomes](evidence/pol91-firstparty-template-models.json) confirm empty/null modifiers clear, nullable Add/replacement retain null entries, null Remove fails even against empty lists, and ordinary Name removal works. Duplicate removal affects only the first ordinal Name match; case differs, unmatched names remain, null names compare equal, and dereferencing an existing null entry fails. Nested expansion includes Description and therefore serializes retained root Description when the template list alone is bound. Existing production helper/resolve behavior matches; no speculative write correction is needed.

Independent control reflection: own AreParametersSpecified checks bound TeamsMeetingTemplates at IL11/16 and Description at IL34/39, not their values; UpdateProperties applies helper at IL66 and independently writes bound Description at IL82/87/101. Identity and Filter have effective ValidateNotNullOrEmpty, while Description and TeamsMeetingTemplates have no extra validation attributes. Both Gets default to Identity with separate Filter; FirstParty Set has one Identity set. Native path safety and early error/typed presentation remain explicit API boundaries.

The correction applies resource-local read adapters and replays full own typed projections independently of MarshalJSON's raw view while preserving RawFields and RawJSON. Two obsolete Description:false malformed vectors now use source-rejected objects. Canonical oracle producer validation stays strict with recursive producer DTOs ahead of SDK setter validation. The full affected SDK, shared labkit and corresponding lab check is recorded below; complete command decisions remain under review.

The correction is present in the continuation worktree after the prior combined continuation checkpoint passed. The frozen decoder checkpoint passed its bounded full affected check; complete per-command decisions remain pending. No management lifecycle, authentication, HTTP, tenant mutation, permission change or new TestAcc run has occurred.


## Frozen correction execution

At `2d72bcd4e64b6e9413e3eae3a2744f7f4bca7f37`, session 6554 passed the full
TestUnit/Example selection for both SDKs, shared labkit and both corresponding
lab packages: 416 passing events, zero failures/skips, five package terminals.
The [execution record](../../../contracts/coverage/evidence/pol91-template-read-execution.json)
retains exact source scope, raw-result hash, portable replay and admission detail.
Scoped source diff was zero and owned scratch was removed. No live test ran.
The existing prepared 204/readback and denial tests remain separately scoped.
The later prepared-write failure/accepted-malformed and explicit recovery gap
then passed at `c3408096e986c3c4bcb5dd6f367d2e3b6047820c`, session 66609:
204 passing events, no failures/skips across FirstParty SDK/lab and root policy
contracts. See the [separate execution record](../../../contracts/coverage/evidence/pol91-firstparty-composition-execution.json)
and [pending command decisions](pol91-template-contract-decisions.md).
