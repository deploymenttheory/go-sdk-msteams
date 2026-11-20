<!-- weave-backlog:msteams-powershell-parity:v1:FND-01 -->
Backlog ID: `FND-01` · Epic: [EPIC-FND](EPIC-FND.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Initialize the Go module and cross-platform quality checks. This provides the ci contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Replace template workload documentation with the planned SDK tree and module path. Pin a supported Go toolchain and compatible Resty/Zap/OTel/Microsoft identity dependencies. Configure build, vet, unit/race tests, lint, vulnerability checks and example compilation across Linux/macOS/Windows. Update template CI/CODEOWNERS links without merging unrelated dependency PRs.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] A minimal module and example compile on all platforms; ordinary CI needs no tenant secrets; lint configuration matches the selected tool version.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run go build ./..., go vet ./..., go test ./..., and platform matrix jobs. Verify race testing on supported runners and verify no test attempts authentication by default.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [ARCH-01](ARCH-01.md) — provides record client, authentication and backend architecture contracts.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M1 SDK foundations.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: ci; FND workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `ci(fnd): initialize the Go module and cross-platform quality checks`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
