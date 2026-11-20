<!-- weave-backlog:msteams-powershell-parity:v1:POL-92 -->
Backlog ID: `POL-92` · Epic: [EPIC-POL](EPIC-POL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Qualify integrated pol policy compatibility coverage. This provides the verification contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Audit and integrate the completed contracts, services, builders and fixtures for: application_access_policy, online_voicemail_policy, teams_app_permission_policy, teams_app_setup_policy, teams_audio_conferencing_policy, teams_call_hold_policy, teams_call_park_policy, teams_calling_policy, teams_channels_policy, teams_cortana_policy, teams_education_assignments_app_policy, teams_enhanced_encryption_policy, teams_events_policy, teams_feedback_policy, teams_files_policy, teams_media_connectivity_policy, teams_media_logging_policy, teams_meeting_broadcast_policy, teams_meeting_policy, teams_meeting_template_permission_policy, teams_messaging_policy, teams_mobility_policy, teams_network_roaming_policy, teams_notification_and_feeds_policy, teams_personal_attendant_policy, teams_recording_roll_out_policy, teams_room_video_tele_conferencing_policy, teams_shared_calling_routing_policy, teams_shifts_app_policy, teams_shifts_policy, teams_template_permission_policy, teams_update_management_policy, teams_upgrade_policy, teams_vdi_policy, teams_virtual_appointments_policy, teams_voice_applications_policy, teams_work_load_policy, teams_work_location_detection_policy, teamsai_policy, teamsbyod_and_desks_policy, teamsip_phone_policy. Verify shared identities/nested types and cross-service operations where applicable, compile examples together, and produce a signed-off coverage/evidence report consumed by QUAL-01. Check authorization failures and recovery across resource boundaries, not merely isolated child test counts. Missing live resources retain contract-tested status.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] All assigned family contracts are implemented and tested; integrated identity/type interactions work; the evidence report names every command, test and live limitation.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run the combined affected package tests, example compilation and coverage checker; compare resource command lists to BASE-03 and execute available acceptance scenarios with cleanup evidence.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [POL-20](POL-20.md) — provides implement application access policy management parity.
- [POL-25](POL-25.md) — provides implement online voicemail policy management parity.
- [POL-31](POL-31.md) — provides implement teams app permission policy management parity.
- [POL-32](POL-32.md) — provides implement teams app setup policy management parity.
- [POL-34](POL-34.md) — provides implement teams audio conferencing policy management parity.
- [POL-35](POL-35.md) — provides implement teams call hold policy management parity.
- [POL-36](POL-36.md) — provides implement teams call park policy management parity.
- [POL-37](POL-37.md) — provides implement teams calling policy management parity.
- [POL-38](POL-38.md) — provides implement teams channels policy management parity.
- [POL-40](POL-40.md) — provides implement teams cortana policy management parity.
- [POL-42](POL-42.md) — provides implement teams education assignments app policy management parity.
- [POL-44](POL-44.md) — provides implement teams enhanced encryption policy management parity.
- [POL-45](POL-45.md) — provides implement teams events policy management parity.
- [POL-46](POL-46.md) — provides implement teams feedback policy management parity.
- [POL-47](POL-47.md) — provides implement teams files policy management parity.
- [POL-52](POL-52.md) — provides implement teams media connectivity policy management parity.
- [POL-53](POL-53.md) — provides implement teams media logging policy management parity.
- [POL-55](POL-55.md) — provides implement teams meeting broadcast policy management parity.
- [POL-57](POL-57.md) — provides implement teams meeting policy management parity.
- [POL-59](POL-59.md) — provides implement teams meeting template permission policy management parity.
- [POL-61](POL-61.md) — provides implement teams messaging policy management parity.
- [POL-63](POL-63.md) — provides implement teams mobility policy management parity.
- [POL-65](POL-65.md) — provides implement teams network roaming policy management parity.
- [POL-66](POL-66.md) — provides implement teams notification and feeds policy management parity.
- [POL-67](POL-67.md) — provides implement teams personal attendant policy management parity.
- [POL-68](POL-68.md) — provides implement teams recording roll out policy management parity.
- [POL-69](POL-69.md) — provides implement teams room video tele conferencing policy management parity.
- [POL-70](POL-70.md) — provides implement teams shared calling routing policy management parity.
- [POL-71](POL-71.md) — provides implement teams shifts app policy management parity.
- [POL-72](POL-72.md) — provides implement teams shifts policy management parity.
- [POL-74](POL-74.md) — provides implement teams template permission policy management parity.
- [POL-76](POL-76.md) — provides implement teams update management policy management parity.
- [POL-78](POL-78.md) — provides implement teams upgrade policy management parity.
- [POL-79](POL-79.md) — provides implement teams vdi policy management parity.
- [POL-80](POL-80.md) — provides implement teams virtual appointments policy management parity.
- [POL-81](POL-81.md) — provides implement teams voice applications policy management parity.
- [POL-82](POL-82.md) — provides implement teams work load policy management parity.
- [POL-83](POL-83.md) — provides implement teams work location detection policy management parity.
- [POL-84](POL-84.md) — provides implement teamsai policy management parity.
- [POL-85](POL-85.md) — provides implement teamsbyod and desks policy management parity.
- [POL-86](POL-86.md) — provides implement teamsip phone policy management parity.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M3 Complete management surface.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: verification; POL workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(pol): qualify integrated pol policy compatibility coverage`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
