<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-CORE -->
Backlog ID: `EPIC-CORE` · Product baseline: [INDEX](INDEX.md)

## Outcome

Teams, channels, membership and templates. Every assigned command is implemented and contract-tested; representative team and channel lifecycles are qualified against the available lab. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [CORE-01](CORE-01.md) — research; P0; Establish contracts for channel_members, channels, deployment_migration, operations, shared_channels, targeting, team_members, team_templates, teams.
- [ ] [CORE-15](CORE-15.md) — verification; P0; Build core resources contract-test fixtures and service harness.
- [ ] [CORE-20](CORE-20.md) — sdk; P1; Implement channel members management parity.
- [ ] [CORE-21](CORE-21.md) — sdk; P1; Implement channels management parity.
- [ ] [CORE-22](CORE-22.md) — sdk; P1; Implement deployment migration management parity.
- [ ] [CORE-23](CORE-23.md) — sdk; P1; Implement operations management parity.
- [ ] [CORE-24](CORE-24.md) — sdk; P1; Implement shared channels management parity.
- [ ] [CORE-25](CORE-25.md) — sdk; P1; Implement targeting management parity.
- [ ] [CORE-26](CORE-26.md) — sdk; P1; Implement team members management parity.
- [ ] [CORE-27](CORE-27.md) — sdk; P1; Implement team templates management parity.
- [ ] [CORE-28](CORE-28.md) — sdk; P1; Implement teams management parity.
- [ ] [CORE-90](CORE-90.md) — verification; P0; Qualify integrated core resources compatibility coverage.

## Epic acceptance

- [ ] Every assigned command is implemented and contract-tested; representative team and channel lifecycles are qualified against the available lab.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M3 Complete management surface; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P1 — required release capability.
Component / scope owner: CORE workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
