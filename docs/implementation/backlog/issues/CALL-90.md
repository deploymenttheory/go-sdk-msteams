<!-- weave-backlog:msteams-powershell-parity:v1:CALL-90 -->
Backlog ID: `CALL-90` · Epic: [EPIC-CALL](EPIC-CALL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Qualify integrated call resources compatibility coverage. This provides the verification contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Audit and integrate the completed contracts, services, builders and fixtures for: attendant_builders, audio_prompts, auto_attendants, call_queues, call_templates, compliance_applications, mainline_attendants, resource_accounts, schedules. Verify shared identities/nested types and cross-service operations where applicable, compile examples together, and produce a signed-off coverage/evidence report consumed by QUAL-01. Check authorization failures and recovery across resource boundaries, not merely isolated child test counts. Missing live resources retain contract-tested status.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] All assigned family contracts are implemented and tested; integrated identity/type interactions work; the evidence report names every command, test and live limitation.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run the combined affected package tests, example compilation and coverage checker; compare resource command lists to BASE-03 and execute available acceptance scenarios with cleanup evidence.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [CALL-22](CALL-22.md) — provides implement auto attendants management parity.
- [CALL-23](CALL-23.md) — provides implement call queues management parity.
- [CALL-24](CALL-24.md) — provides implement call templates management parity.
- [CALL-25](CALL-25.md) — provides implement compliance applications management parity.
- [CALL-26](CALL-26.md) — provides implement mainline attendants management parity.

Supporting contracts covered transitively by the direct blockers: [CALL-20](CALL-20.md), [CALL-21](CALL-21.md), [CALL-27](CALL-27.md), [CALL-28](CALL-28.md). These are context, not additional native blockers.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M3 Complete management surface.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: verification; CALL workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(call): qualify integrated call resources compatibility coverage`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
