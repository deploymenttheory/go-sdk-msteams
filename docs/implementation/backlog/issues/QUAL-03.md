<!-- weave-backlog:msteams-powershell-parity:v1:QUAL-03 -->
Backlog ID: `QUAL-03` · Epic: [EPIC-QUAL](EPIC-QUAL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Publish SDK guides, CI recipes and compatibility documentation. This provides the documentation contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Replace template README/docs with quick start, authentication/permission guides, cloud support, error/pagination/async usage, API-family examples and Linux CI federation/certificate recipes. Include PowerShell-to-Go mappings, version support, upgrade review, unsupported plumbing exclusions and live-evidence limitations. Explain that SDK use requires no PowerShell installation.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] A new consumer can authenticate and call both backend families from documented examples; every in-scope service has a discoverable example and permission reference.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Compile all examples; check source links and command mappings; run the portable quick-start smoke case where lab access permits.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [AUTH-05](AUTH-05.md) — provides qualify permissions and authentication capability reporting.
- [DELTA-02](DELTA-02.md) — provides schedule upstream discovery and maintain candidate draft prs.
- [QUAL-01](QUAL-01.md) — provides audit complete implementation and contract coverage.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M5 Release qualification.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: documentation; QUAL workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `docs(qual): publish SDK guides, CI recipes and compatibility documentation`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
