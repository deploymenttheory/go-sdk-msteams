<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-ARCH -->
Backlog ID: `EPIC-ARCH` · Product baseline: [INDEX](INDEX.md)

## Outcome

Portable client and service architecture contracts. A representative Graph service, Teams administration service and local builder follow one documented public contract. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [ARCH-01](ARCH-01.md) — documentation; P0; Record client, authentication and backend architecture contracts.

## Epic acceptance

- [ ] A representative Graph service, Teams administration service and local builder follow one documented public contract.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M0 Baseline and contracts; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P0 — foundation or delivery gate.
Component / scope owner: ARCH workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
