<!-- weave-backlog:msteams-powershell-parity:v1:POL-91 -->
Backlog ID: `POL-91` · Epic: [EPIC-POL](EPIC-POL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Qualify integrated pol configuration compatibility coverage. This provides the verification contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Audit and integrate the completed contracts, services, builders and fixtures for: application_meeting_configuration, online_voicemail_validation_configuration, privacy_configuration, teams_audio_conferencing_custom_prompts_configuration, teams_client_configuration, teams_education_configuration, teams_first_party_meeting_template_configuration, teams_guest_calling_configuration, teams_guest_meeting_configuration, teams_guest_messaging_configuration, teams_meeting_broadcast_configuration, teams_meeting_configuration, teams_meeting_template_configuration, teams_messaging_configuration, teams_migration_configuration, teams_multi_tenant_organization_configuration, teams_sip_devices_configuration, teams_tenant_abuse_configuration, teams_upgrade_configuration, tenant_migration_configuration. Verify shared identities/nested types and cross-service operations where applicable, compile examples together, and produce a signed-off coverage/evidence report consumed by QUAL-01. Check authorization failures and recovery across resource boundaries, not merely isolated child test counts. Missing live resources retain contract-tested status.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] All assigned family contracts are implemented and tested; integrated identity/type interactions work; the evidence report names every command, test and live limitation.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run the combined affected package tests, example compilation and coverage checker; compare resource command lists to BASE-03 and execute available acceptance scenarios with cleanup evidence.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [POL-21](POL-21.md) — provides implement application meeting configuration management parity.
- [POL-26](POL-26.md) — provides implement online voicemail validation configuration management parity.
- [POL-29](POL-29.md) — provides implement privacy configuration management parity.
- [POL-33](POL-33.md) — provides implement teams audio conferencing custom prompts configuration management parity.
- [POL-39](POL-39.md) — provides implement teams client configuration management parity.
- [POL-43](POL-43.md) — provides implement teams education configuration management parity.
- [POL-48](POL-48.md) — provides implement teams first party meeting template configuration management parity.
- [POL-49](POL-49.md) — provides implement teams guest calling configuration management parity.
- [POL-50](POL-50.md) — provides implement teams guest meeting configuration management parity.
- [POL-51](POL-51.md) — provides implement teams guest messaging configuration management parity.
- [POL-54](POL-54.md) — provides implement teams meeting broadcast configuration management parity.
- [POL-56](POL-56.md) — provides implement teams meeting configuration management parity.
- [POL-58](POL-58.md) — provides implement teams meeting template configuration management parity.
- [POL-60](POL-60.md) — provides implement teams messaging configuration management parity.
- [POL-62](POL-62.md) — provides implement teams migration configuration management parity.
- [POL-64](POL-64.md) — provides implement teams multi tenant organization configuration management parity.
- [POL-73](POL-73.md) — provides implement teams sip devices configuration management parity.
- [POL-75](POL-75.md) — provides implement teams tenant abuse configuration management parity.
- [POL-77](POL-77.md) — provides implement teams upgrade configuration management parity.
- [POL-87](POL-87.md) — provides implement tenant migration configuration management parity.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M3 Complete management surface.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: verification; POL workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(pol): qualify integrated pol configuration compatibility coverage`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
