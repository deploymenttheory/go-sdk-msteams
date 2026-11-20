<!-- weave-backlog:msteams-powershell-parity:v1:DELTA-03 -->
Backlog ID: `DELTA-03` · Epic: [EPIC-DELTA](EPIC-DELTA.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Detect sampled server drift independently of module releases. This provides the integration contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Add optional scheduled read-only probes against configured lab capabilities and compare sanitized response/behaviour observations with accepted contracts. Keep server drift reports separate from module version differences and map affected resources. Missing credentials or unsupported capabilities report skipped/unavailable without modifying the baseline.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] A changed sampled response shape is detected even with unchanged module version; probes never mutate tenant configuration; sampling limitations are explicit.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Test added/removed/type-changed fields, transient service errors, unavailable probes and unchanged observations; verify report redaction and absence of mutation calls.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [DELTA-01](DELTA-01.md) — provides implement semantic baseline comparison and sdk impact reports.
- [LAB-02](LAB-02.md) — provides compare sdk and pinned powershell behaviour with sanitized evidence.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M4 Upgrade detection.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: integration; DELTA workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(delta): detect sampled server drift independently of module releases`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
