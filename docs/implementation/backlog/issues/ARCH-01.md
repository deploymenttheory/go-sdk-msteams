<!-- weave-backlog:msteams-powershell-parity:v1:ARCH-01 -->
Backlog ID: `ARCH-01` · Epic: [EPIC-ARCH](EPIC-ARCH.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Record client, authentication and backend architecture contracts. This provides the documentation contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Turn the agreed architecture in docs/implementation/plan.md into numbered decisions and a contributor guide. Specify root client composition, constructor injection, Resty responses, functional options, Zap and optional OTel. Define resource-aware token acquisition, backend routing and error boundaries. Provide illustrative service signatures and a local builder example; distinguish Go API versioning from PowerShell versioning.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] The architecture explains Graph, Teams administration and local builders, with package dependencies free of import cycles and no runtime PowerShell/WAM requirement.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Review the three representative call paths against the pinned Jamf reference and product rules; include a public-interface compatibility checklist.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [BASE-01](BASE-01.md) — provides build reproducible package acquisition and normalized baseline extraction.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M0 Baseline and contracts.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: documentation; ARCH workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `docs(arch): record client, authentication and backend architecture contracts`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
