<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-LAB -->
Backlog ID: `EPIC-LAB` · Product baseline: [INDEX](INDEX.md)

## Outcome

Repeatable tenant conformance and fixture lifecycle. Available tenant workflows compare SDK and pinned PowerShell results with sanitized evidence and verified cleanup. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [LAB-01](LAB-01.md) — integration; P0; Build tenant fixture and acceptance-test lifecycle.
- [ ] [LAB-02](LAB-02.md) — integration; P0; Compare SDK and pinned PowerShell behaviour with sanitized evidence.
- [ ] [LAB-03](LAB-03.md) — verification; P0; Prove the representative cross-backend workflows.

## Epic acceptance

- [ ] Available tenant workflows compare SDK and pinned PowerShell results with sanitized evidence and verified cleanup.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M2 Representative proof; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P0 — foundation or delivery gate.
Component / scope owner: LAB workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
