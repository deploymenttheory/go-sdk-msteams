# AI-agent seven-command review packet

Review pending. No implementation, contract or live ledger state is changed.
This packet covers the seven accepted public commands and their seven declared
sets at frozen `b8540c1`, following the independently reviewed first repair
`b838d0f` / `1cb18c1`. It does not treat symbols, synthetic replies, upstream
binding observations or denials as successful tenant operations.

[Exact execution, hashes and seven metadata mappings](evidence/ai-agents-b8540c1-review.json)
include every existing per-set symbol/test reference and its immutable runtime
pointer. [Primary source and supplement manifest](evidence/ai-agents-source/manifest.json)
bind the portable pinned bodies/reproducers. The shared header proof is reused
only after tracing this family's own wrappers, proxies and pipeline arguments.
MCPS uses its independently audited factory and resource.

## Commands and complete declared sets

| Command / set | Native operation and source request | Input and output boundary |
| --- | --- | --- |
| Get-CsAgent / __AllParameterSets | List or GetByID; 200; Get-CsAgent_Get or Get-CsAgent_Get1 | Optional Id selects the path; Force is removed. Public list has no paging/filter parameters. Native ListOptions are a generated-internal extension. Config count is Int64; null collection becomes empty public output, null rows fail projection. |
| New-CsAgent / __AllParameterSets | Create; exactly 201; New-CsAgent_New | Three mandatory strings; optional target omitted when unbound, explicit empty retained. Body/parse failure preserves HTTP receipt; a null generated root yields nil model. No automatic replay or ownership inferred from missing output. |
| Set-CsAgent / __AllParameterSets | UpdateByID; exactly 200; Set-CsAgent_SetExpanded | Instance.Id selects path; actual wrapper/binder binds all four string fields and converts null to empty. Native typed request supplies the same replacement fields. No implicit read/merge. |
| Remove-CsAgent / __AllParameterSets | DeleteByID; exactly 200; Remove-CsAgent_Remove | Required Id; empty, whitespace or array success bodies yield null generated result without losing receipt; literal null/numeric root is a parser error. One DELETE exchange is asserted. |
| Get-CsAiAgents / Get | QueryPlatformAgents; 200; Get-CsAiAgents_Get | Required ProviderId; optional strings omit only empty/null, switches use True/False and signed Int32 maxResults is preserved. Caller controls continuation. Count stays Int32; catalog rows/string arrays retain nulls; description lazily defaults. |
| Get-AIGeneratedKnowledgeContainer / __AllParameterSets | ListKnowledgeContainers; any 2xx; public command header | Required ThreadId, optional nonempty CallId, encoded query; legacy parser retains null strings and exact numeric lexemes. Native raw nullable list differs from semicolon-joined Purview batch presentation. Empty/null body fails source Count access, while an actual empty array succeeds. |
| Remove-AIGeneratedKnowledge / __AllParameterSets | DeleteKnowledge; any 2xx; public command header | Same required thread/optional call addressing. Native returns receipt, source prints a fixed success message. Permanent operation, no automatic replay. No real knowledge deletion was performed. |

## Model, output and diagnostic semantics

Generated ConfigAPI JSON is exact-case/type, rejects exact duplicate keys
recursively even in unknown nested fields, and ignores wrong-type fields rather
than applying Newtonsoft coercion. All 56 duplicate-error observations and 25
accepted ordering observations now assert the selected actual getter projection
across all five model kinds. Nested source objects are canonicalized by semantic
JSON value with UseNumber, preserving numbers and ignoring irrelevant key order.
The generated nullable string arrays preserve every source position; MCPS uses
its separate legacy parser. Int64/Int32 ranges, fractional/exponent integer
failures and Double overflow/underflow are separately source-bound.

An absent/null AIAgent inside a response object projects an empty domain Agent.
Empty/whitespace/array roots return a null generated response. The immutable
primary's initial requirement for a Remove object is explicitly superseded by
the ordering supplement; no historical packet is rewritten. Config-list null
rows fail public projection, whereas nullable catalog rows survive. RawFields
and raw response bytes are native enrichment, not emitted PowerShell output.

Get/New/Remove invoke Write-AdminServiceDiagnostic: default/null/Info messages
use success output, Warning and Verbose use their streams. Set does not call
that helper. The SDK retains raw diagnostics/receipt without logging tenant
content or pretending its model-only result reproduces stream presentation.
MCPS output batching/headings/warnings and fixed delete message are explicitly
presentation boundaries. Legacy source retry/error flattening is not copied;
native typed errors, context and write replay controls remain deliberate.

## Common controls, identities and authentication

Source metadata exposes WhatIf/Confirm for Set and Remove-CsAgent; their custom
and internal wrappers forward common parameters. Native callers withhold the
mutation call when choosing not to act; there is no invented Go WhatIf flag or
interactive approval. The [bounded source gate supplement](evidence/ai-agents-source/controls-supplement.json)
resolves each actual ProcessRecord: ShouldProcess at IL50, false branch at IL57
skips the ProcessRecordAsync call at IL76. It withholds the remote action path;
start signals/BeginProcessing/session or telemetry hooks are not asserted
absent. This is exact static source proof, not executed PowerShell WhatIf. Other five public
commands have no WhatIf/Confirm declaration. Get's Force is removed. Shared
PowerShell stream/session controls map to the repository's native client/auth,
context, transport and error conventions, not extra resource fields.

Configuration IDs, provider IDs, agent IDs and target template IDs remain
distinct. Mutation path guards reject unsafe path parts independently of looser
source string binding; source-accepted optional query whitespace/control data
is encoded by url.Values, not reclassified as a path. Numeric query values are
not given an invented nonnegative constraint. Continuation is returned as data,
never followed to an arbitrary origin; the example has cycle/page bounds.
Recovery retains unresolved state for a nil list response rather than declaring
absence. No record returned by a read automatically grants delete ownership.

ConfigAPI uses the Teams Admin audience and discovered forest under the existing
trusted client. Historical dedicated app-only lab responses prove the owned
configuration lifecycle and empty catalog/BadArgument cases at their original
revisions; no general entitlement or specific permission grant is inferred.
MCPS independently requests the fixed application resource
6bc3b958-689b-49f5-9006-36d165f30e00/.default at the commercial Teams MCPS origin.
Its app-only 401 and pinned certificate pre-request KeyNotFoundException are
failures. Successful delegated MCPS access and populated catalog entries remain
external fixture/auth prerequisites, not successful evidence. No roles,
licenses, grants, fallback identity or tenant configuration was changed.

## Actual execution and remaining decision scope

Full AI SDK, AI lab and EXT integration TestUnit/Example selection at b8540c1:
three package terminals PASS, 166 test-pass events, zero failures/skips; no
TestAcc. The exact raw SHA/command are in the JSON packet. Earlier 07e7df2 failed
only the test's nested RawMessage key-order comparison; that raw artifact is
retained. Source expected values and production code were not changed to make
it pass. All three real Remove null-result service cases passed in both runs.

The two peer-review gaps are now covered. Final per-command approval still
requires review of this source/control/auth/native-boundary packet, including
the bounded static ShouldProcess gate supplement for Set/Remove. No automatic seven-command promotion follows from aggregate
passes. Live states for all 745 records remain unchanged, and #191 remains open.


## Catalog model follow-up under review

The first frozen repair left eleven known AiAgentDescription properties only in
RawFields. That was insufficient for full catalog qualification. The additive
follow-up types all 25 pinned properties, including developer privacy/terms,
extra, disclosures, supported locales, five tile URLs and webpage. The
[bounded actual description supplement](evidence/ai-agents-source/description-fields-supplement.json)
contains two real inert parser/getter observations: all eleven positive fields
(including both additional nullable arrays), and wrong-type/exact-case handling.
The original primary/five-packet manifest and a310c8e execution remain immutable.
Full affected execution at b8540c1 completed with three package terminals PASS
and 166 test-pass events, zero failures/skips, no TestAcc. Exact hashes and raw
provenance are in the successor JSON packet; the earlier a310c8e record is
preserved. Get-CsAiAgents and all six other command records remain unchanged
until independent review.
