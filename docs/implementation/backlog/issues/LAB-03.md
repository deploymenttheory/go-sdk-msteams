<!-- weave-backlog:msteams-powershell-parity:v1:LAB-03 -->
Backlog ID: `LAB-03` · Epic: [EPIC-LAB](EPIC-LAB.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Prove the representative cross-backend workflows. This provides the verification contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Qualify team/channel/membership lifecycle, messaging-policy lifecycle and assignment, and an online voice-user read after their implementation prerequisites. Verify app-only and delegated restrictions and read-after-write behaviour. Missing live capability is recorded with contract evidence; unexplained differences remain defects.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] All selected workflows have passing contract comparisons and available live evidence, with cleanup verified independently of operation success.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run create/read/update/delete and membership/assignment scenarios against dedicated fixtures; attach SDK and PowerShell baseline identifiers, differences and cleanup records.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [CORE-20](CORE-20.md) — provides implement channel members management parity.
- [CORE-26](CORE-26.md) — provides implement team members management parity.
- [LAB-02](LAB-02.md) — provides compare sdk and pinned powershell behaviour with sanitized evidence.
- [POL-27](POL-27.md) — provides implement policy assignments management parity.
- [POL-61](POL-61.md) — provides implement teams messaging policy management parity.
- [VOICE-31](VOICE-31.md) — provides implement online users management parity.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M2 Representative proof.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: verification; LAB workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(lab): prove the representative cross-backend workflows`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
