# Audio conferencing custom prompts configuration

Access through `sdk.TeamsAdminAPI.AudioConferencingCustomPrompts`. Construction
uses the injected Teams scope and performs no authentication or network I/O.

| Native operation | Baseline behavior |
| --- | --- |
| `GetByIdentity` | Get Identity with an explicit full configuration identity |
| `List` | Get without Identity, bare collection, exact Global first |
| `ListFiltered` | Get Filter, local Identity/Tag wildcard selection |
| `PatchByIdentity` | One sparse PATCH with complete bound arrays |
| `UpdateByIdentity` | Explicit Set read/edit/PATCH composition |
| `NewPrompt`, `NewPromptPackage` | Local source-constructor equivalents |
| `Replace/Change/ClearPrompts`, `Replace/Change/ClearPackages` | Owned local edits |

The pinned command family exposes Get and Set. Configuration creation/deletion
and assignment routes need separate source contracts. Prompt/package creation
here means constructing list elements locally and including them in a bound
configuration update; it does not imply a separate server resource endpoint.

Get and Set use the direct compiled PolicyRp resource path, Teams administration
audience, MPACmdlet and command/correlation headers. The service shares session,
auth, error handling and operation replay controls with the client. It does not
introduce a separate authentication or retry loop.

`Configuration` has Prompts and Packages. `Prompt` includes Identity, Id, Name,
AudioPrompt, TextPrompt, Type and Locale. `PromptPackage` includes Identity, Id,
Name, DefaultLocale and the six Inbound/Outbound/Meeting Start/End recording
reference arrays. IDs use source GUID validation; Type/Locale accept empty or
future strings. Package reference strings retain null, empty and future values.
The parameterized native constructors use strings to express non-null inputs.
Source default models may contain missing/null getters, which remain readable.

Received readonly arrays ignore null and append repeated properties, including
casing variants. `RawJSON` retains original presence and duplicates; `RawFields`
retains the last occurrence of each exact key and unknown properties. Marshaling
a received model returns its raw snapshot. Native writes project the current
typed known fields, preserving nullable text and normalizing nil slices to `[]`.
Unknown raw properties do not enter a write implicitly.

In `Settings`, an unset `policy.Field` omits the list, `policy.Set(nilSlice)` emits
`[]`, and an explicit null array field fails locally. Empty Settings and Force
alone fail before credentials or HTTP. BodyIdentity may explicitly override the
route identity, including null. Bound Force false/true is sent as the source
string `False`/`True`; nil omits it.

In `UpdateRequest`, nil edits are unbound. Clear and empty operators explicitly
clear; Replace preserves order, duplicates and null elements. Change removes
before adding and does not deduplicate. Prompt removal deletes the first ordinal
match across Id, Name, AudioPrompt, TextPrompt, Type and Locale; Identity does not
participate. Comparing a null prompt fails. Package removal fails locally when
any existing entry is examined, as in the source; removal from an empty list
succeeds. Package add, replace and clear remain explicit supported operations.
Edits copy pointers, nested reference arrays and raw data at construction and
Apply, so caller mutations do not alter an edit's stored values.

Update reads with Set command metadata, retains Before/ReadResponse, applies edits
to owned copies, then sends only bound lists using the read Identity. A local
edit failure retains the read snapshot and sends no PATCH. PatchInvoked and
HTTPAccepted remain available on partial failures. After is a decoded PATCH
response, not a persisted read-back. No polling, ETag check, atomicity guarantee,
automatic restoration or implicit mutation replay is added.

Application and delegated synthetic cases test identical wire semantics.
Actual client-secret and certificate Global reads on 2026-11-21 returned 403.
Successful Teams roles, consent, cloud/feature availability, writes, read-back,
cleanup and effective playback remain unqualified. No privilege grants are
implicit in using this service. See the
[unit tests](service_test.go)
and [lab acceptance](../../../docs/acceptance/audio-conferencing-custom-prompts-lab-acceptance.md).
