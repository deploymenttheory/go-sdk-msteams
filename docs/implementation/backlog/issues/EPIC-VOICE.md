<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-VOICE -->
Backlog ID: `EPIC-VOICE` · Product baseline: [INDEX](INDEX.md)

## Outcome

Voice users, numbering, routing and emergency administration. Voice operations retain their typed contracts, authorization rules, order semantics and explicit live-test limitations. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [VOICE-01](VOICE-01.md) — research; P0; Establish contracts for blocked_numbers, bulk_number_orders, calling_line_identity, calling_settings.
- [ ] [VOICE-02](VOICE-02.md) — research; P0; Establish contracts for conferencing, dial_plans.
- [ ] [VOICE-03](VOICE-03.md) — research; P0; Establish contracts for emergency_addresses, emergency_network.
- [ ] [VOICE-04](VOICE-04.md) — research; P0; Establish contracts for emergency_policies, number_inventory, number_orders.
- [ ] [VOICE-05](VOICE-05.md) — research; P0; Establish contracts for online_users, phone_assignments, phone_tags, pstn_gateways, survivability.
- [ ] [VOICE-06](VOICE-06.md) — research; P0; Establish contracts for tenant_network, video_interop, voice_routes.
- [ ] [VOICE-15](VOICE-15.md) — verification; P0; Build voice resources contract-test fixtures and service harness.
- [ ] [VOICE-20](VOICE-20.md) — sdk; P1; Implement blocked numbers management parity.
- [ ] [VOICE-21](VOICE-21.md) — sdk; P1; Implement bulk number orders management parity.
- [ ] [VOICE-22](VOICE-22.md) — sdk; P1; Implement calling line identity management parity.
- [ ] [VOICE-23](VOICE-23.md) — sdk; P1; Implement calling settings management parity.
- [ ] [VOICE-24](VOICE-24.md) — sdk; P1; Implement conferencing management parity.
- [ ] [VOICE-25](VOICE-25.md) — sdk; P1; Implement dial plans management parity.
- [ ] [VOICE-26](VOICE-26.md) — sdk; P1; Implement emergency addresses management parity.
- [ ] [VOICE-27](VOICE-27.md) — sdk; P1; Implement emergency network management parity.
- [ ] [VOICE-28](VOICE-28.md) — sdk; P1; Implement emergency policies management parity.
- [ ] [VOICE-29](VOICE-29.md) — sdk; P1; Implement number inventory management parity.
- [ ] [VOICE-30](VOICE-30.md) — sdk; P1; Implement number orders management parity.
- [ ] [VOICE-31](VOICE-31.md) — sdk; P1; Implement online users management parity.
- [ ] [VOICE-32](VOICE-32.md) — sdk; P1; Implement phone assignments management parity.
- [ ] [VOICE-33](VOICE-33.md) — sdk; P1; Implement phone tags management parity.
- [ ] [VOICE-34](VOICE-34.md) — sdk; P1; Implement pstn gateways management parity.
- [ ] [VOICE-35](VOICE-35.md) — sdk; P1; Implement survivability management parity.
- [ ] [VOICE-36](VOICE-36.md) — sdk; P1; Implement tenant network management parity.
- [ ] [VOICE-37](VOICE-37.md) — sdk; P1; Implement video interop management parity.
- [ ] [VOICE-38](VOICE-38.md) — sdk; P1; Implement voice routes management parity.
- [ ] [VOICE-90](VOICE-90.md) — verification; P0; Qualify integrated voice resources compatibility coverage.

## Epic acceptance

- [ ] Voice operations retain their typed contracts, authorization rules, order semantics and explicit live-test limitations.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M3 Complete management surface; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P1 — required release capability.
Component / scope owner: VOICE workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
