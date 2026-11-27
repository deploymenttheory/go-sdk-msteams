# Messaging and migration configuration qualification

Owners: #133 (POL-60) and #135 (POL-62). Baseline: MicrosoftTeams 8.1.0.

The pinned public runtime metadata has Get Identity/Filter and Set Identity for
each resource. Messaging dispatches compiled Core cmdlets through direct PolicyRP
GET/PATCH. Migration dispatches public proxy functions through `CmdletConfig.json`
to `Get-CsConfigurationModern` / `Set-CsConfigurationModern` and the generated
ConfigAPI GET/PUT routes. The modern wrapper defaults Set Identity to Global;
its Boolean is not nullable and it exposes no Force. Older compiled migration
classes are not the selected public protocol.

Local primary sources: Core `TeamsMessagingConfiguration`, its Get/Set cmdlets,
`Storyline`, shared `BaseTpm*` and list modifiers; public proxy definitions,
`CmdletConfig.json`, merged modern helpers and generated ConfigAPI API methods.
Oracle checks pin these module assembly SHA256 values:

- Core: `68dbcb1f2392bc00d7f9f6ab457a6f49ef05dab387f448d39c23f455cf521e49`
- PolicyRP: `a5c8bc39cb6f43bd484bd6613549a6d0a2235f8d9b0b53b558828ab0aeb40e51`
- ConfigAPI: `8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6`

On 2026-11-22 the dedicated SDK service principal, already assigned Teams
Administrator, successfully read/listed Global and compared the typed public
getters with PowerShell 8.1.0. Both application authentication modes passed.
Messaging has 12 settings, including a nonempty actual file block list and a
named Storyline enum response. Migration has one Boolean setting.

Each update test reads all original getter values, refuses a missing/null selected
Boolean, persists its original value in a bound private recovery journal, then
toggles it. Messaging modifies only EnableVideoMessageCaptions and verifies every
other typed getter remains unchanged. Migration modifies only
EnableLegacyClientInterop. Both require persisted read-back and pinned getter
comparison, restore only the selected original Boolean, and independently read
and compare the complete original getter projection. Each test compares the
pinned module before, after and after restoration. Both application auth methods
passed the exact CI driver with both restore journal entries cleaned.

Global read tests use a serial read suite: they run under the same CI tenant lock
as mutations, but cannot enable mutations. CI also requires all three comparisons,
update/omission and restoration assertions, and exactly one cleaned entry per
update test. The generic bound recovery test registers both restoration kinds.
Existing repository secrets suffice; no additional roles or licenses changed.

Ten permanent sanitized response files come directly from the acceptance reads,
updates and restored read-backs. The real 204 bodies remain empty. Unit tests use
these responses for native routing, headers, typed getters, snapshot ownership,
filter selection, sparse updates, omission and local file-list edits. Common
PolicyRP denial propagation uses the actual privacy-service 403 envelope and is
explicitly not a live denial claim for these services. Deliberately malformed
replies test decoding failures separately from successful service fixtures.

Still unqualified: live messaging list replacement/Add/Remove/clear, explicit
scalar-null persistence, effective client behavior/propagation, delegated flows
and other clouds. Unit/SDK behaviours are kept separate from these live gaps.
Neither singleton invents create/delete methods absent from the pinned exports.
