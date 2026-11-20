<!-- weave-backlog:msteams-powershell-parity:v1:AUTH-01 -->
Backlog ID: `AUTH-01` · Epic: [EPIC-AUTH](EPIC-AUTH.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Implement resource-aware supplied tokens and refresh coordination. This provides the sdk contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Define the auth provider request/result contract and adapters for audience-keyed supplied tokens and refresh callbacks. Key caches by tenant, cloud, identity and audience; serialize refresh and honor expiry/cancellation. Reject mixing delegated/application identity within a session. Do not forward a Graph token to a Teams endpoint. Static tokens report expiry instead of pretending to refresh.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Audience isolation, expiry, concurrent refresh and callback errors are deterministic; secrets never enter errors or debug logs.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Use fake clocks/token providers for near-expiry, simultaneous requests, cancelled refresh, wrong tenant/audience and missing token cases. Run the race detector.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [FND-02](FND-02.md) — provides wire the root client, configuration and typed service composition.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M1 SDK foundations.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: sdk; AUTH workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(auth): implement resource-aware supplied tokens and refresh coordination`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
