<!-- weave-backlog:msteams-powershell-parity:v1:BASE-02 -->
Backlog ID: `BASE-02` · Epic: [EPIC-BASE](EPIC-BASE.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Reconcile all exported and conditional command candidates. This provides the research contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Resolve the 730 candidates against actual exports, nested modules, platform/session conditions and documentation. The checked macOS observation contains 680 commands after case normalization. Investigate the other 50 individually. Retain every candidate in the ledger with evidence; only telemetry plumbing is already excluded. Record local builders, session equivalents, aliases and public management classifications. Add cross-platform observations when runners exist, without requiring Windows for SDK runtime.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Every candidate has an evidenced disposition and ownership; no absence from one session is treated as proof of exclusion; additional public exports discovered later are added rather than hidden.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Reconcile case-only names, commented preview declarations, conditional exports and internal telemetry examples; validate the union and explicit disposition totals against provenance. Publish the exact unresolved list if any contract research remains.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [BASE-01](BASE-01.md) — provides build reproducible package acquisition and normalized baseline extraction.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M0 Baseline and contracts.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: research; BASE workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `docs(base): reconcile all exported and conditional command candidates`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
