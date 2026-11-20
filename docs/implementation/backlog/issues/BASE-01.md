<!-- weave-backlog:msteams-powershell-parity:v1:BASE-01 -->
Backlog ID: `BASE-01` · Epic: [EPIC-BASE](EPIC-BASE.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Build reproducible package acquisition and normalized baseline extraction. This provides the tooling contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Promote the checked discovery seed into a repeatable tools/baseline workflow. Acquire exact Gallery package/metadata, verify SHA512, isolate imports, extract manifests/help/configuration/runtime signatures, and preserve provenance and schema version. Keep Microsoft binaries outside tracked artifacts. Preserve case-insensitive identities with original spellings and distinguish incomplete extraction from an empty surface.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Repeated extraction from identical inputs produces the same normalized contract; checksum mismatch, import failure and incomplete metadata fail visibly; the 8.1.0 snapshot cannot be overwritten implicitly.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Use the pinned archive and saved metadata plus checksum-corruption, truncated-download, import-error and empty-inventory fixtures. Compare two independent extractions byte-for-byte after removing explicitly volatile provenance fields.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

None; ready to start.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M0 Baseline and contracts.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: tooling; BASE workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(base): build reproducible package acquisition and normalized baseline extraction`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
