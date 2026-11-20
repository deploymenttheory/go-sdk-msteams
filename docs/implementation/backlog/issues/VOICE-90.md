<!-- weave-backlog:msteams-powershell-parity:v1:VOICE-90 -->
Backlog ID: `VOICE-90` · Epic: [EPIC-VOICE](EPIC-VOICE.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Qualify integrated voice resources compatibility coverage. This provides the verification contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Audit and integrate the completed contracts, services, builders and fixtures for: blocked_numbers, bulk_number_orders, calling_line_identity, calling_settings, conferencing, dial_plans, emergency_addresses, emergency_network, emergency_policies, number_inventory, number_orders, online_users, phone_assignments, phone_tags, pstn_gateways, survivability, tenant_network, video_interop, voice_routes. Verify shared identities/nested types and cross-service operations where applicable, compile examples together, and produce a signed-off coverage/evidence report consumed by QUAL-01. Check authorization failures and recovery across resource boundaries, not merely isolated child test counts. Missing live resources retain contract-tested status.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] All assigned family contracts are implemented and tested; integrated identity/type interactions work; the evidence report names every command, test and live limitation.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run the combined affected package tests, example compilation and coverage checker; compare resource command lists to BASE-03 and execute available acceptance scenarios with cleanup evidence.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [VOICE-20](VOICE-20.md) — provides implement blocked numbers management parity.
- [VOICE-21](VOICE-21.md) — provides implement bulk number orders management parity.
- [VOICE-22](VOICE-22.md) — provides implement calling line identity management parity.
- [VOICE-23](VOICE-23.md) — provides implement calling settings management parity.
- [VOICE-24](VOICE-24.md) — provides implement conferencing management parity.
- [VOICE-25](VOICE-25.md) — provides implement dial plans management parity.
- [VOICE-27](VOICE-27.md) — provides implement emergency network management parity.
- [VOICE-28](VOICE-28.md) — provides implement emergency policies management parity.
- [VOICE-29](VOICE-29.md) — provides implement number inventory management parity.
- [VOICE-30](VOICE-30.md) — provides implement number orders management parity.
- [VOICE-31](VOICE-31.md) — provides implement online users management parity.
- [VOICE-32](VOICE-32.md) — provides implement phone assignments management parity.
- [VOICE-33](VOICE-33.md) — provides implement phone tags management parity.
- [VOICE-34](VOICE-34.md) — provides implement pstn gateways management parity.
- [VOICE-35](VOICE-35.md) — provides implement survivability management parity.
- [VOICE-36](VOICE-36.md) — provides implement tenant network management parity.
- [VOICE-37](VOICE-37.md) — provides implement video interop management parity.
- [VOICE-38](VOICE-38.md) — provides implement voice routes management parity.

Supporting contracts covered transitively by the direct blockers: [VOICE-26](VOICE-26.md). These are context, not additional native blockers.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M3 Complete management surface.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: verification; VOICE workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(voice): qualify integrated voice resources compatibility coverage`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
