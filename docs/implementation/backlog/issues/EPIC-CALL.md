<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-CALL -->
Backlog ID: `EPIC-CALL` · Product baseline: [INDEX](INDEX.md)

## Outcome

Queues, attendants, resource accounts and voice builders. Nested calling configurations round-trip through typed models and builders with operation-specific failure handling. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [CALL-01](CALL-01.md) — research; P0; Establish contracts for attendant_builders, audio_prompts, auto_attendants, call_queues.
- [ ] [CALL-02](CALL-02.md) — research; P0; Establish contracts for call_templates, compliance_applications.
- [ ] [CALL-03](CALL-03.md) — research; P0; Establish contracts for mainline_attendants, resource_accounts, schedules.
- [ ] [CALL-15](CALL-15.md) — verification; P0; Build call resources contract-test fixtures and service harness.
- [ ] [CALL-20](CALL-20.md) — sdk; P1; Implement attendant builders management parity.
- [ ] [CALL-21](CALL-21.md) — sdk; P1; Implement audio prompts management parity.
- [ ] [CALL-22](CALL-22.md) — sdk; P1; Implement auto attendants management parity.
- [ ] [CALL-23](CALL-23.md) — sdk; P1; Implement call queues management parity.
- [ ] [CALL-24](CALL-24.md) — sdk; P1; Implement call templates management parity.
- [ ] [CALL-25](CALL-25.md) — sdk; P1; Implement compliance applications management parity.
- [ ] [CALL-26](CALL-26.md) — sdk; P1; Implement mainline attendants management parity.
- [ ] [CALL-27](CALL-27.md) — sdk; P1; Implement resource accounts management parity.
- [ ] [CALL-28](CALL-28.md) — sdk; P1; Implement schedules management parity.
- [ ] [CALL-90](CALL-90.md) — verification; P0; Qualify integrated call resources compatibility coverage.

## Epic acceptance

- [ ] Nested calling configurations round-trip through typed models and builders with operation-specific failure handling.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M3 Complete management surface; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P1 — required release capability.
Component / scope owner: CALL workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
