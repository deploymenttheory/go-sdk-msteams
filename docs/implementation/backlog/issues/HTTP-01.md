<!-- weave-backlog:msteams-powershell-parity:v1:HTTP-01 -->
Backlog ID: `HTTP-01` · Epic: [EPIC-HTTP](EPIC-HTTP.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Implement backend routing, request building and normalized errors. This provides the sdk contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Implement the shared client interface with Resty and cloud/audience-aware backend adapters. Build escaped paths/queries, preserve backend headers and expose raw response metadata plus typed errors with status, service code and request IDs. Preserve context and body cleanup. Resolve only allowlisted endpoint/discovery/continuation destinations and strip secrets from errors.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Graph and Teams envelopes decode without losing diagnostic identifiers; invalid discovery destinations cannot receive credentials; cancellation reaches the request.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Use httptest backends for JSON/non-JSON errors, empty bodies, malformed content, escaped identities, cross-host redirects, wrong token audience and cancellation.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [AUTH-01](AUTH-01.md) — provides implement resource-aware supplied tokens and refresh coordination.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M1 SDK foundations.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: sdk; HTTP workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(http): implement backend routing, request building and normalized errors`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
