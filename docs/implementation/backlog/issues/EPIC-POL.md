<!-- weave-backlog:msteams-powershell-parity:v1:EPIC-POL -->
Backlog ID: `EPIC-POL` · Product baseline: [INDEX](INDEX.md)

## Outcome

Policies, configuration and assignment. All assigned policy fields and assignment semantics are covered, including 8.1.0 additions, inheritance and batch outcomes. The starting SDK is template-only; this epic carries no implementation-complete claim.

## Agreed behavior

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Implementation tickets

- [ ] [POL-01](POL-01.md) — research; P0; Establish contracts for application_access_policy, application_meeting_configuration, assignment_migration, federation, meeting_branding, online_voicemail_policy.
- [ ] [POL-02](POL-02.md) — research; P0; Establish contracts for online_voicemail_validation_configuration, policy_assignments, policy_packages, privacy_configuration, recording_messages, teams_app_permission_policy.
- [ ] [POL-03](POL-03.md) — research; P0; Establish pol contracts from teams_app_setup_policy through teams_cortana_policy.
- [ ] [POL-04](POL-04.md) — research; P0; Establish pol contracts from teams_custom_banner_text through teams_guest_messaging_configuration.
- [ ] [POL-05](POL-05.md) — research; P0; Establish pol contracts from teams_media_connectivity_policy through teams_migration_configuration.
- [ ] [POL-06](POL-06.md) — research; P0; Establish pol contracts from teams_mobility_policy through teams_shifts_app_policy.
- [ ] [POL-07](POL-07.md) — research; P0; Establish pol contracts from teams_shifts_policy through teams_voice_applications_policy.
- [ ] [POL-08](POL-08.md) — research; P0; Establish contracts for teams_work_load_policy, teams_work_location_detection_policy, teamsai_policy, teamsbyod_and_desks_policy, teamsip_phone_policy, tenant_migration_configuration.
- [ ] [POL-15](POL-15.md) — verification; P0; Build pol compound contract-test fixtures and service harness.
- [ ] [POL-16](POL-16.md) — verification; P0; Build pol configuration contract-test fixtures and service harness.
- [ ] [POL-17](POL-17.md) — verification; P0; Build pol policy contract-test fixtures and service harness.
- [ ] [POL-20](POL-20.md) — sdk; P1; Implement application access policy management parity.
- [ ] [POL-21](POL-21.md) — sdk; P1; Implement application meeting configuration management parity.
- [ ] [POL-22](POL-22.md) — sdk; P1; Implement assignment migration management parity.
- [ ] [POL-23](POL-23.md) — sdk; P1; Implement federation management parity.
- [ ] [POL-24](POL-24.md) — sdk; P1; Implement meeting branding management parity.
- [ ] [POL-25](POL-25.md) — sdk; P1; Implement online voicemail policy management parity.
- [ ] [POL-26](POL-26.md) — sdk; P1; Implement online voicemail validation configuration management parity.
- [ ] [POL-27](POL-27.md) — sdk; P1; Implement policy assignments management parity.
- [ ] [POL-28](POL-28.md) — sdk; P1; Implement policy packages management parity.
- [ ] [POL-29](POL-29.md) — sdk; P1; Implement privacy configuration management parity.
- [ ] [POL-30](POL-30.md) — sdk; P1; Implement recording messages management parity.
- [ ] [POL-31](POL-31.md) — sdk; P1; Implement teams app permission policy management parity.
- [ ] [POL-32](POL-32.md) — sdk; P1; Implement teams app setup policy management parity.
- [ ] [POL-33](POL-33.md) — sdk; P1; Implement teams audio conferencing custom prompts configuration management parity.
- [ ] [POL-34](POL-34.md) — sdk; P1; Implement teams audio conferencing policy management parity.
- [ ] [POL-35](POL-35.md) — sdk; P1; Implement teams call hold policy management parity.
- [ ] [POL-36](POL-36.md) — sdk; P1; Implement teams call park policy management parity.
- [ ] [POL-37](POL-37.md) — sdk; P1; Implement teams calling policy management parity.
- [ ] [POL-38](POL-38.md) — sdk; P1; Implement teams channels policy management parity.
- [ ] [POL-39](POL-39.md) — sdk; P1; Implement teams client configuration management parity.
- [ ] [POL-40](POL-40.md) — sdk; P1; Implement teams cortana policy management parity.
- [ ] [POL-41](POL-41.md) — sdk; P1; Implement teams custom banner text management parity.
- [ ] [POL-42](POL-42.md) — sdk; P1; Implement teams education assignments app policy management parity.
- [ ] [POL-43](POL-43.md) — sdk; P1; Implement teams education configuration management parity.
- [ ] [POL-44](POL-44.md) — sdk; P1; Implement teams enhanced encryption policy management parity.
- [ ] [POL-45](POL-45.md) — sdk; P1; Implement teams events policy management parity.
- [ ] [POL-46](POL-46.md) — sdk; P1; Implement teams feedback policy management parity.
- [ ] [POL-47](POL-47.md) — sdk; P1; Implement teams files policy management parity.
- [ ] [POL-48](POL-48.md) — sdk; P1; Implement teams first party meeting template configuration management parity.
- [ ] [POL-49](POL-49.md) — sdk; P1; Implement teams guest calling configuration management parity.
- [ ] [POL-50](POL-50.md) — sdk; P1; Implement teams guest meeting configuration management parity.
- [ ] [POL-51](POL-51.md) — sdk; P1; Implement teams guest messaging configuration management parity.
- [ ] [POL-52](POL-52.md) — sdk; P1; Implement teams media connectivity policy management parity.
- [ ] [POL-53](POL-53.md) — sdk; P1; Implement teams media logging policy management parity.
- [ ] [POL-54](POL-54.md) — sdk; P1; Implement teams meeting broadcast configuration management parity.
- [ ] [POL-55](POL-55.md) — sdk; P1; Implement teams meeting broadcast policy management parity.
- [ ] [POL-56](POL-56.md) — sdk; P1; Implement teams meeting configuration management parity.
- [ ] [POL-57](POL-57.md) — sdk; P1; Implement teams meeting policy management parity.
- [ ] [POL-58](POL-58.md) — sdk; P1; Implement teams meeting template configuration management parity.
- [ ] [POL-59](POL-59.md) — sdk; P1; Implement teams meeting template permission policy management parity.
- [ ] [POL-60](POL-60.md) — sdk; P1; Implement teams messaging configuration management parity.
- [ ] [POL-61](POL-61.md) — sdk; P1; Implement teams messaging policy management parity.
- [ ] [POL-62](POL-62.md) — sdk; P1; Implement teams migration configuration management parity.
- [ ] [POL-63](POL-63.md) — sdk; P1; Implement teams mobility policy management parity.
- [ ] [POL-64](POL-64.md) — sdk; P1; Implement teams multi tenant organization configuration management parity.
- [ ] [POL-65](POL-65.md) — sdk; P1; Implement teams network roaming policy management parity.
- [ ] [POL-66](POL-66.md) — sdk; P1; Implement teams notification and feeds policy management parity.
- [ ] [POL-67](POL-67.md) — sdk; P1; Implement teams personal attendant policy management parity.
- [ ] [POL-68](POL-68.md) — sdk; P1; Implement teams recording roll out policy management parity.
- [ ] [POL-69](POL-69.md) — sdk; P1; Implement teams room video tele conferencing policy management parity.
- [ ] [POL-70](POL-70.md) — sdk; P1; Implement teams shared calling routing policy management parity.
- [ ] [POL-71](POL-71.md) — sdk; P1; Implement teams shifts app policy management parity.
- [ ] [POL-72](POL-72.md) — sdk; P1; Implement teams shifts policy management parity.
- [ ] [POL-73](POL-73.md) — sdk; P1; Implement teams sip devices configuration management parity.
- [ ] [POL-74](POL-74.md) — sdk; P1; Implement teams template permission policy management parity.
- [ ] [POL-75](POL-75.md) — sdk; P1; Implement teams tenant abuse configuration management parity.
- [ ] [POL-76](POL-76.md) — sdk; P1; Implement teams update management policy management parity.
- [ ] [POL-77](POL-77.md) — sdk; P1; Implement teams upgrade configuration management parity.
- [ ] [POL-78](POL-78.md) — sdk; P1; Implement teams upgrade policy management parity.
- [ ] [POL-79](POL-79.md) — sdk; P1; Implement teams vdi policy management parity.
- [ ] [POL-80](POL-80.md) — sdk; P1; Implement teams virtual appointments policy management parity.
- [ ] [POL-81](POL-81.md) — sdk; P1; Implement teams voice applications policy management parity.
- [ ] [POL-82](POL-82.md) — sdk; P1; Implement teams work load policy management parity.
- [ ] [POL-83](POL-83.md) — sdk; P1; Implement teams work location detection policy management parity.
- [ ] [POL-84](POL-84.md) — sdk; P1; Implement teamsai policy management parity.
- [ ] [POL-85](POL-85.md) — sdk; P1; Implement teamsbyod and desks policy management parity.
- [ ] [POL-86](POL-86.md) — sdk; P1; Implement teamsip phone policy management parity.
- [ ] [POL-87](POL-87.md) — sdk; P1; Implement tenant migration configuration management parity.
- [ ] [POL-90](POL-90.md) — verification; P0; Qualify integrated pol compound compatibility coverage.
- [ ] [POL-91](POL-91.md) — verification; P0; Qualify integrated pol configuration compatibility coverage.
- [ ] [POL-92](POL-92.md) — verification; P0; Qualify integrated pol policy compatibility coverage.

## Epic acceptance

- [ ] All assigned policy fields and assignment semantics are covered, including 8.1.0 additions, inheritance and batch outcomes.
- [ ] Integrated evidence covers interactions across children, including authorization, error recovery and applicable cleanup; contract/live evidence is distinguished.
- [ ] Coverage and documentation match the reconciled baseline; the maintainer evaluates these conditions independently of child closure.

## Planning

Milestone: M3 Complete management surface; this is the integrated epic exit. Children may deliver in earlier gates and do not wait for parent closure.
Priority: P1 — required release capability.
Component / scope owner: POL workstream; native child membership must match this checklist.
Source rationale: agreed implementation plan at starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`, pinned 8.1.0 metadata and INDEX. Direct blockers belong to child contracts, not milestone order.

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
