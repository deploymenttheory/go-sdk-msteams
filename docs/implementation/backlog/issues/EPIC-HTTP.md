<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-HTTP -->
Backlog ID: `EPIC-HTTP` · Product baseline: [INDEX](INDEX.md)

## Outcome

Reliable multi-backend REST transport. Backend adapters share tested cancellation, throttling and observability while retaining backend-specific paging, errors and operation semantics. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [HTTP-01](HTTP-01.md) — sdk; P0; Implement backend routing, request building and normalized errors.
- [ ] [HTTP-02](HTTP-02.md) — sdk; P0; Implement bounded retries, throttling and concurrency controls.
- [ ] [HTTP-03](HTTP-03.md) — sdk; P0; Implement backend-specific pagination and list traversal.
- [ ] [HTTP-04](HTTP-04.md) — sdk; P0; Implement long-running operations and eventual-consistency waits.
- [ ] [HTTP-05](HTTP-05.md) — sdk; P0; Implement binary, multipart and batch request behaviours.
- [ ] [HTTP-06](HTTP-06.md) — sdk; P0; Add redacted logging, tracing and transport configuration.

## Epic acceptance

- [ ] Backend adapters share tested cancellation, throttling and observability while retaining backend-specific paging, errors and operation semantics.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M1 SDK foundations; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P0 — foundation or delivery gate.
Component / scope owner: HTTP workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
