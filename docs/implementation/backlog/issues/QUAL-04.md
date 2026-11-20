<!-- weave-backlog:msteams-powershell-parity:v1:QUAL-04 -->
Backlog ID: `QUAL-04` · Epic: [EPIC-QUAL](EPIC-QUAL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Gate and prepare the first supported SDK release. This provides the ci contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Add release automation, changelog/version rules and reproducible release checks. Require complete implementation/contract coverage and the published evidence matrix before a supported tag. Keep PowerShell baseline version separate from SDK semver. Document upgrade and rollback procedures; publishing the backlog does not authorize an SDK release tag.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Release dry-run fails for missing capability evidence or failed gates and succeeds for a fully qualified commit; no partial build is labelled full parity.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Exercise release dry-run with passing/failing gate fixtures and verify version/changelog metadata and baseline provenance without publishing a release.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [DELTA-03](DELTA-03.md) — provides detect sampled server drift independently of module releases.
- [QUAL-02](QUAL-02.md) — provides qualify concurrency, security and available lab regressions.
- [QUAL-03](QUAL-03.md) — provides publish sdk guides, ci recipes and compatibility documentation.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M5 Release qualification.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: ci; QUAL workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `ci(qual): gate and prepare the first supported SDK release`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
