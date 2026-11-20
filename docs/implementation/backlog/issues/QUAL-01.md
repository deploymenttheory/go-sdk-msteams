<!-- weave-backlog:msteams-powershell-parity:v1:QUAL-01 -->
Backlog ID: `QUAL-01` · Epic: [EPIC-QUAL](EPIC-QUAL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Audit complete implementation and contract coverage. This provides the verification contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Audit every in-scope command, parameter set, builder and session equivalent against compiled Go symbols and executable contract tests. Reject unsupported stubs and missing models/permissions/examples. Reconcile all resource-family research outputs, conditional candidates and explicit exclusions. Publish a requirement-to-ticket-to-test coverage report.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Every management capability has an implementation and contract-test evidence; excluded telemetry and portable session mappings are explicit; no unowned commands remain.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run the coverage checker and negative mutation fixtures, inspect representative per-family evidence and compare the audited command set to the immutable baseline.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [AUTH-06](AUTH-06.md) — provides map powershell connection and environment helpers to portable sdk sessions.
- [CALL-90](CALL-90.md) — provides qualify integrated call resources compatibility coverage.
- [CORE-90](CORE-90.md) — provides qualify integrated core resources compatibility coverage.
- [EXT-90](EXT-90.md) — provides qualify integrated ext resources compatibility coverage.
- [POL-90](POL-90.md) — provides qualify integrated pol compound compatibility coverage.
- [POL-91](POL-91.md) — provides qualify integrated pol configuration compatibility coverage.
- [POL-92](POL-92.md) — provides qualify integrated pol policy compatibility coverage.
- [VOICE-90](VOICE-90.md) — provides qualify integrated voice resources compatibility coverage.

Supporting contracts covered transitively by the direct blockers: [BASE-03](BASE-03.md). These are context, not additional native blockers.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M5 Release qualification.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: verification; QUAL workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(qual): audit complete implementation and contract coverage`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
