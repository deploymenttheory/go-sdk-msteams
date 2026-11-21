<!-- weave-backlog:msteams-powershell-parity:v1:LAB-02 -->
Backlog ID: `LAB-02` · Epic: [EPIC-LAB](EPIC-LAB.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Compare SDK and pinned PowerShell behaviour with sanitized evidence. This provides the integration contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Implement a conformance harness that runs pinned 8.1.0 and SDK operations on equivalent isolated resources, normalizes only documented volatile fields and compares outcomes/read-back. Check the package hash, authentication mode, scenario and cleanup result. Store sanitized response fixtures directly in each service testdata directory for TestUnit reuse. Do not equate a skipped TestAcc case to a pass.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Deliberate field/permission/serialization differences are detected; normalized comparisons preserve meaningful ordering and omission/null distinctions.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Use controlled equal/different fixtures and failed-cleanup scenarios, then execute available tenant cases; review permanent response fixtures for credentials and identifiers before committing.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [BASE-03](BASE-03.md) — provides maintain the capability and verification coverage ledger.
- [LAB-01](LAB-01.md) — provides build tenant fixture and acceptance-test lifecycle.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M2 Representative proof.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: integration; LAB workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(lab): compare SDK and pinned PowerShell behaviour with sanitized evidence`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
