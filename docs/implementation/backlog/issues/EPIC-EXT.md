<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-EXT -->
Backlog ID: `EPIC-EXT` · Product baseline: [INDEX](INDEX.md)

## Outcome

Applications, Shifts, devices and remaining management services. All assigned extension families have implemented contracts; conditional/internal candidates have evidenced dispositions. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [EXT-01](EXT-01.md) — research; P0; Establish contracts for ai_agents, app_catalog, cloud_call_data, configuration_diagnostics, devices.
- [ ] [EXT-02](EXT-02.md) — research; P0; Establish contracts for migration, provisioning_diagnostics, shifts_connections.
- [ ] [EXT-03](EXT-03.md) — research; P0; Establish contracts for shifts_sync, tenant_identity, unified_apps.
- [ ] [EXT-15](EXT-15.md) — verification; P0; Build ext resources contract-test fixtures and service harness.
- [ ] [EXT-20](EXT-20.md) — sdk; P1; Implement ai agents management parity.
- [ ] [EXT-21](EXT-21.md) — sdk; P1; Implement app catalog management parity.
- [ ] [EXT-22](EXT-22.md) — sdk; P1; Implement cloud call data management parity.
- [ ] [EXT-23](EXT-23.md) — sdk; P1; Implement configuration diagnostics management parity.
- [ ] [EXT-24](EXT-24.md) — sdk; P1; Implement devices management parity.
- [ ] [EXT-25](EXT-25.md) — sdk; P1; Implement migration management parity.
- [ ] [EXT-26](EXT-26.md) — sdk; P1; Implement provisioning diagnostics management parity.
- [ ] [EXT-27](EXT-27.md) — sdk; P1; Implement shifts connections management parity.
- [ ] [EXT-28](EXT-28.md) — sdk; P1; Implement shifts sync management parity.
- [ ] [EXT-29](EXT-29.md) — sdk; P1; Implement tenant identity management parity.
- [ ] [EXT-30](EXT-30.md) — sdk; P1; Implement unified apps management parity.
- [ ] [EXT-90](EXT-90.md) — verification; P0; Qualify integrated ext resources compatibility coverage.

## Epic acceptance

- [ ] All assigned extension families have implemented contracts; conditional/internal candidates have evidenced dispositions.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M3 Complete management surface; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P1 — required release capability.
Component / scope owner: EXT workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
