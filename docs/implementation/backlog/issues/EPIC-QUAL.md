<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-QUAL -->
Backlog ID: `EPIC-QUAL` · Product baseline: [INDEX](INDEX.md)

## Outcome

Complete baseline coverage and first supported release. All management operations are implemented and contract-tested, available lab cases pass and release documentation states evidence limitations. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [QUAL-01](QUAL-01.md) — verification; P0; Audit complete implementation and contract coverage.
- [ ] [QUAL-02](QUAL-02.md) — verification; P0; Qualify concurrency, security and available lab regressions.
- [ ] [QUAL-03](QUAL-03.md) — documentation; P0; Publish SDK guides, CI recipes and compatibility documentation.
- [ ] [QUAL-04](QUAL-04.md) — ci; P0; Gate and prepare the first supported SDK release.

## Epic acceptance

- [ ] All management operations are implemented and contract-tested, available lab cases pass and release documentation states evidence limitations.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M5 Release qualification; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P0 — foundation or delivery gate.
Component / scope owner: QUAL workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
