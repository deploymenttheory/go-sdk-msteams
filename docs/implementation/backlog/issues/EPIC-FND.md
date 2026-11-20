<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-FND -->
Backlog ID: `EPIC-FND` · Product baseline: [INDEX](INDEX.md)

## Outcome

Go module and cross-platform development foundation. The new module builds and tests on Linux, macOS and Windows with a root client and compiled examples. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [FND-01](FND-01.md) — ci; P0; Initialize the Go module and cross-platform quality checks.
- [ ] [FND-02](FND-02.md) — sdk; P0; Wire the root client, configuration and typed service composition.

## Epic acceptance

- [ ] The new module builds and tests on Linux, macOS and Windows with a root client and compiled examples.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M1 SDK foundations; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P0 — foundation or delivery gate.
Component / scope owner: FND workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
