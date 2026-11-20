<!-- weave-backlog:msteams-powershell-parity:v1:DELTA-01 -->
Backlog ID: `DELTA-01` · Epic: [EPIC-DELTA](EPIC-DELTA.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Implement semantic baseline comparison and SDK impact reports. This provides the tooling contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Build tools/contractdiff over versioned inventories. Detect command/alias additions/removals, parameter-set/type/requiredness/default/validation/output differences and evidenced routing/auth changes. Classify documentation-only versus metadata versus observed behaviour changes. Link impacts to services, models, tests and stable issue IDs. Unknown fields and incomplete extraction are visible.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Identical snapshots produce no delta; each meaningful change has old/new evidence and an affected owner; incomplete snapshots cannot yield a clean result.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Use synthetic add/remove/rename/type/default/null fixtures and verified historical 8.1.0 changes, including CallQueue Statistics removal and new phone usage-order/parameter additions. Test stable versus preview versions.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [BASE-03](BASE-03.md) — provides maintain the capability and verification coverage ledger.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M4 Upgrade detection.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: tooling; DELTA workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(delta): implement semantic baseline comparison and SDK impact reports`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
