# Microsoft Teams SDK backlog publication preview

Target: `deploymenttheory/go-sdk-msteams`. Namespace: `msteams-powershell-parity`. Status: **draft, not published**.

This review set contains **193 issues: one baseline, 13 epics and 179 tasks**, covering 116 resource families and all 730 command candidates. It adds **25 labels** and **6 milestones**, with **456 direct blocker edges** and **179 native parent relationships**. Existing standard labels and dependency PRs are unchanged; no feature issues or milestones existed at inspection. No obsolete work, updates or identity conflicts are proposed.

Read the [agreed plan](../plan.md), [baseline issue](issues/INDEX.md), [command coverage](coverage.md), and [exact publication manifest](manifest.json). Full final draft bodies are linked below; they are not title-only placeholders.

Ready to start: [BASE-01](issues/BASE-01.md). Subsequent readiness follows direct blockers, not numbering. Research tickets deliberately block implementations with unsettled wire contracts.

## Publication contract

Approval covers these exact bodies, metadata, parent sets and blockers. Each body begins with a stable identity marker. Publication resolves local `ID.md` issue links to actual GitHub URLs, without changing scope. All other body content is preserved; no source revisions or successful tests are invented. Body SHA256 values in the manifest protect this preview against accidental changes.

Create missing metadata, baseline, epics and tasks in dependency order; then attach native sub-issues/blockers, resolve links and read everything back. Native API permissions are not yet exercised. If native links are unavailable, publication is incomplete even when textual links exist. Retain the issue-ID mapping and journal for recovery; do not duplicate ambiguous writes.

## Labels

All labels below are new orthogonal dimensions; existing `enhancement`, `documentation`, `dependencies` and `github_actions` retain their meanings. No synthetic blocked/status labels or GitHub Project changes are proposed.

| Label | Color | Meaning |
| --- | --- | --- |
| `type:backlog` | `#5319e7` | Compatibility baseline index |
| `type:epic` | `#5319e7` | Integrated behaviour and child work |
| `type:task` | `#5319e7` | Bounded implementation or research outcome |
| `priority:P0` | `#b60205` | Foundation, contract integrity or release gate |
| `priority:P1` | `#fbca04` | Required release capability |
| `area:base` | `#0052cc` | Reproducible PowerShell compatibility baseline |
| `area:arch` | `#0052cc` | Portable client and service architecture contracts |
| `area:fnd` | `#0052cc` | Go module and cross-platform development foundation |
| `area:auth` | `#0052cc` | Portable delegated and automation authentication |
| `area:http` | `#0052cc` | Reliable multi-backend REST transport |
| `area:lab` | `#0052cc` | Repeatable tenant conformance and fixture lifecycle |
| `area:core` | `#0052cc` | Teams, channels, membership and templates |
| `area:pol` | `#0052cc` | Policies, configuration and assignment |
| `area:voice` | `#0052cc` | Voice users, numbering, routing and emergency administration |
| `area:call` | `#0052cc` | Queues, attendants, resource accounts and voice builders |
| `area:ext` | `#0052cc` | Applications, Shifts, devices and remaining management services |
| `area:delta` | `#0052cc` | Detect and review upstream compatibility changes |
| `area:qual` | `#0052cc` | Complete baseline coverage and first supported release |
| `component:ci` | `#0e8a16` | Owns ci work |
| `component:documentation` | `#0e8a16` | Owns documentation work |
| `component:integration` | `#0e8a16` | Owns integration work |
| `component:research` | `#0e8a16` | Owns research work |
| `component:sdk` | `#0e8a16` | Owns sdk work |
| `component:tooling` | `#0e8a16` | Owns tooling work |
| `component:verification` | `#0e8a16` | Owns verification work |

## Delivery gates

- **M0 Baseline and contracts:** Immutable 8.1.0 evidence, reconciled command ownership and agreed architecture. Family contract research precedes each implementation; no guessed wire routes.
- **M1 SDK foundations:** Portable Go client, authentication, backend adapters and cross-platform offline checks pass.
- **M2 Representative proof:** Team/channel/membership, policy lifecycle/assignment and voice read workflows have contract proof plus available lab evidence.
- **M3 Complete management surface:** Every in-scope resource and builder is implemented with tests, permissions, examples and coverage links.
- **M4 Upgrade detection:** Versioned and observed differences are classified and mapped to SDK impact; scheduled workflows maintain reviewable candidate draft PRs.
- **M5 Release qualification:** Full implementation and contract coverage, cross-platform gates, available lab regression and honest qualification documentation permit the first supported release.

No due dates are proposed. Cross-milestone epics target their integrated exit; their M0 research and earlier implementation children remain independently actionable. DELTA work may start once its prerequisites exist.

## Complete issue bodies and direct blockers

### BASE

[Epic BASE](issues/EPIC-BASE.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[BASE-01] Build reproducible package acquisition and normalized baseline extraction](issues/BASE-01.md) | M0 | None; ready |
| [[BASE-02] Reconcile all exported and conditional command candidates](issues/BASE-02.md) | M0 | BASE-01 |
| [[BASE-03] Maintain the capability and verification coverage ledger](issues/BASE-03.md) | M0 | BASE-02 |

### ARCH

[Epic ARCH](issues/EPIC-ARCH.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[ARCH-01] Record client, authentication and backend architecture contracts](issues/ARCH-01.md) | M0 | BASE-01 |

### FND

[Epic FND](issues/EPIC-FND.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[FND-01] Initialize the Go module and cross-platform quality checks](issues/FND-01.md) | M1 | ARCH-01 |
| [[FND-02] Wire the root client, configuration and typed service composition](issues/FND-02.md) | M1 | FND-01 |

### AUTH

[Epic AUTH](issues/EPIC-AUTH.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[AUTH-01] Implement resource-aware supplied tokens and refresh coordination](issues/AUTH-01.md) | M1 | FND-02 |
| [[AUTH-02] Add portable certificate and client-secret authentication](issues/AUTH-02.md) | M1 | AUTH-01 |
| [[AUTH-03] Add portable delegated login and credential flows](issues/AUTH-03.md) | M1 | AUTH-01 |
| [[AUTH-04] Add managed identity and CI workload federation](issues/AUTH-04.md) | M1 | AUTH-01 |
| [[AUTH-05] Qualify permissions and authentication capability reporting](issues/AUTH-05.md) | M1 | AUTH-02, AUTH-03, AUTH-04, BASE-02 |
| [[AUTH-06] Map PowerShell connection and environment helpers to portable SDK sessions](issues/AUTH-06.md) | M1 | AUTH-01, BASE-02 |

### HTTP

[Epic HTTP](issues/EPIC-HTTP.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[HTTP-01] Implement backend routing, request building and normalized errors](issues/HTTP-01.md) | M1 | AUTH-01 |
| [[HTTP-02] Implement bounded retries, throttling and concurrency controls](issues/HTTP-02.md) | M1 | HTTP-01 |
| [[HTTP-03] Implement backend-specific pagination and list traversal](issues/HTTP-03.md) | M1 | HTTP-01 |
| [[HTTP-04] Implement long-running operations and eventual-consistency waits](issues/HTTP-04.md) | M1 | HTTP-02 |
| [[HTTP-05] Implement binary, multipart and batch request behaviours](issues/HTTP-05.md) | M1 | HTTP-02 |
| [[HTTP-06] Add redacted logging, tracing and transport configuration](issues/HTTP-06.md) | M1 | HTTP-01 |

### LAB

[Epic LAB](issues/EPIC-LAB.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[LAB-01] Build tenant fixture and acceptance-test lifecycle](issues/LAB-01.md) | M2 | HTTP-01 |
| [[LAB-02] Compare SDK and pinned PowerShell behaviour with sanitized evidence](issues/LAB-02.md) | M2 | BASE-03, LAB-01 |
| [[LAB-03] Prove the representative cross-backend workflows](issues/LAB-03.md) | M2 | CORE-20, CORE-26, LAB-02, POL-27, POL-61, VOICE-31 |

### CORE

[Epic CORE](issues/EPIC-CORE.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[CORE-01] Establish contracts for channel_members, channels, deployment_migration, operations, shared_channels, targeting, team_members, team_templates, teams](issues/CORE-01.md) | M0 | ARCH-01, BASE-02 |
| [[CORE-15] Build core resources contract-test fixtures and service harness](issues/CORE-15.md) | M1 | BASE-03, CORE-01, HTTP-02, HTTP-03 |
| [[CORE-20] Implement channel members management parity](issues/CORE-20.md) | M2 | CORE-21 |
| [[CORE-21] Implement channels management parity](issues/CORE-21.md) | M2 | CORE-28 |
| [[CORE-22] Implement deployment migration management parity](issues/CORE-22.md) | M3 | CORE-15, HTTP-04 |
| [[CORE-23] Implement operations management parity](issues/CORE-23.md) | M2 | CORE-15, HTTP-04 |
| [[CORE-24] Implement shared channels management parity](issues/CORE-24.md) | M3 | CORE-21 |
| [[CORE-25] Implement targeting management parity](issues/CORE-25.md) | M3 | CORE-15 |
| [[CORE-26] Implement team members management parity](issues/CORE-26.md) | M2 | CORE-28 |
| [[CORE-27] Implement team templates management parity](issues/CORE-27.md) | M3 | CORE-15 |
| [[CORE-28] Implement teams management parity](issues/CORE-28.md) | M2 | CORE-15, HTTP-04, HTTP-05 |
| [[CORE-90] Qualify integrated core resources compatibility coverage](issues/CORE-90.md) | M3 | CORE-20, CORE-22, CORE-23, CORE-24, CORE-25, CORE-26, CORE-27 |

### POL

[Epic POL](issues/EPIC-POL.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[POL-01] Establish contracts for application_access_policy, application_meeting_configuration, assignment_migration, federation, meeting_branding, online_voicemail_policy](issues/POL-01.md) | M0 | ARCH-01, BASE-02 |
| [[POL-02] Establish contracts for online_voicemail_validation_configuration, policy_assignments, policy_packages, privacy_configuration, recording_messages, teams_app_permission_policy](issues/POL-02.md) | M0 | ARCH-01, BASE-02 |
| [[POL-03] Establish pol contracts from teams_app_setup_policy through teams_cortana_policy](issues/POL-03.md) | M0 | ARCH-01, BASE-02 |
| [[POL-04] Establish pol contracts from teams_custom_banner_text through teams_guest_messaging_configuration](issues/POL-04.md) | M0 | ARCH-01, BASE-02 |
| [[POL-05] Establish pol contracts from teams_media_connectivity_policy through teams_migration_configuration](issues/POL-05.md) | M0 | ARCH-01, BASE-02 |
| [[POL-06] Establish pol contracts from teams_mobility_policy through teams_shifts_app_policy](issues/POL-06.md) | M0 | ARCH-01, BASE-02 |
| [[POL-07] Establish pol contracts from teams_shifts_policy through teams_voice_applications_policy](issues/POL-07.md) | M0 | ARCH-01, BASE-02 |
| [[POL-08] Establish contracts for teams_work_load_policy, teams_work_location_detection_policy, teamsai_policy, teamsbyod_and_desks_policy, teamsip_phone_policy, tenant_migration_configuration](issues/POL-08.md) | M0 | ARCH-01, BASE-02 |
| [[POL-15] Build pol compound contract-test fixtures and service harness](issues/POL-15.md) | M1 | BASE-03, HTTP-02, HTTP-03, POL-01 |
| [[POL-16] Build pol configuration contract-test fixtures and service harness](issues/POL-16.md) | M1 | BASE-03, HTTP-02, HTTP-03, POL-01 |
| [[POL-17] Build pol policy contract-test fixtures and service harness](issues/POL-17.md) | M1 | BASE-03, HTTP-02, HTTP-03, POL-01 |
| [[POL-20] Implement application access policy management parity](issues/POL-20.md) | M3 | POL-17 |
| [[POL-21] Implement application meeting configuration management parity](issues/POL-21.md) | M3 | POL-16 |
| [[POL-22] Implement assignment migration management parity](issues/POL-22.md) | M3 | HTTP-04, POL-15 |
| [[POL-23] Implement federation management parity](issues/POL-23.md) | M3 | POL-15 |
| [[POL-24] Implement meeting branding management parity](issues/POL-24.md) | M3 | HTTP-05, POL-15 |
| [[POL-25] Implement online voicemail policy management parity](issues/POL-25.md) | M3 | POL-17 |
| [[POL-26] Implement online voicemail validation configuration management parity](issues/POL-26.md) | M3 | POL-02, POL-16 |
| [[POL-27] Implement policy assignments management parity](issues/POL-27.md) | M2 | HTTP-04, POL-02, POL-15 |
| [[POL-28] Implement policy packages management parity](issues/POL-28.md) | M3 | POL-02, POL-15 |
| [[POL-29] Implement privacy configuration management parity](issues/POL-29.md) | M3 | POL-02, POL-16 |
| [[POL-30] Implement recording messages management parity](issues/POL-30.md) | M3 | POL-02, POL-15 |
| [[POL-31] Implement teams app permission policy management parity](issues/POL-31.md) | M3 | POL-02, POL-17 |
| [[POL-32] Implement teams app setup policy management parity](issues/POL-32.md) | M3 | POL-03, POL-17 |
| [[POL-33] Implement teams audio conferencing custom prompts configuration management parity](issues/POL-33.md) | M3 | POL-03, POL-16 |
| [[POL-34] Implement teams audio conferencing policy management parity](issues/POL-34.md) | M3 | POL-03, POL-17 |
| [[POL-35] Implement teams call hold policy management parity](issues/POL-35.md) | M3 | POL-03, POL-17 |
| [[POL-36] Implement teams call park policy management parity](issues/POL-36.md) | M3 | POL-03, POL-17 |
| [[POL-37] Implement teams calling policy management parity](issues/POL-37.md) | M3 | POL-03, POL-17 |
| [[POL-38] Implement teams channels policy management parity](issues/POL-38.md) | M3 | POL-03, POL-17 |
| [[POL-39] Implement teams client configuration management parity](issues/POL-39.md) | M3 | POL-03, POL-16 |
| [[POL-40] Implement teams cortana policy management parity](issues/POL-40.md) | M3 | POL-03, POL-17 |
| [[POL-41] Implement teams custom banner text management parity](issues/POL-41.md) | M3 | POL-04, POL-15 |
| [[POL-42] Implement teams education assignments app policy management parity](issues/POL-42.md) | M3 | POL-04, POL-17 |
| [[POL-43] Implement teams education configuration management parity](issues/POL-43.md) | M3 | POL-04, POL-16 |
| [[POL-44] Implement teams enhanced encryption policy management parity](issues/POL-44.md) | M3 | POL-04, POL-17 |
| [[POL-45] Implement teams events policy management parity](issues/POL-45.md) | M3 | POL-04, POL-17 |
| [[POL-46] Implement teams feedback policy management parity](issues/POL-46.md) | M3 | POL-04, POL-17 |
| [[POL-47] Implement teams files policy management parity](issues/POL-47.md) | M3 | POL-04, POL-17 |
| [[POL-48] Implement teams first party meeting template configuration management parity](issues/POL-48.md) | M3 | POL-04, POL-16 |
| [[POL-49] Implement teams guest calling configuration management parity](issues/POL-49.md) | M3 | POL-04, POL-16 |
| [[POL-50] Implement teams guest meeting configuration management parity](issues/POL-50.md) | M3 | POL-04, POL-16 |
| [[POL-51] Implement teams guest messaging configuration management parity](issues/POL-51.md) | M3 | POL-04, POL-16 |
| [[POL-52] Implement teams media connectivity policy management parity](issues/POL-52.md) | M3 | POL-05, POL-17 |
| [[POL-53] Implement teams media logging policy management parity](issues/POL-53.md) | M3 | POL-05, POL-17 |
| [[POL-54] Implement teams meeting broadcast configuration management parity](issues/POL-54.md) | M3 | POL-05, POL-16 |
| [[POL-55] Implement teams meeting broadcast policy management parity](issues/POL-55.md) | M3 | POL-05, POL-17 |
| [[POL-56] Implement teams meeting configuration management parity](issues/POL-56.md) | M3 | POL-05, POL-16 |
| [[POL-57] Implement teams meeting policy management parity](issues/POL-57.md) | M3 | POL-05, POL-17 |
| [[POL-58] Implement teams meeting template configuration management parity](issues/POL-58.md) | M3 | POL-05, POL-16 |
| [[POL-59] Implement teams meeting template permission policy management parity](issues/POL-59.md) | M3 | POL-05, POL-17 |
| [[POL-60] Implement teams messaging configuration management parity](issues/POL-60.md) | M3 | POL-05, POL-16 |
| [[POL-61] Implement teams messaging policy management parity](issues/POL-61.md) | M2 | POL-05, POL-17 |
| [[POL-62] Implement teams migration configuration management parity](issues/POL-62.md) | M3 | POL-05, POL-16 |
| [[POL-63] Implement teams mobility policy management parity](issues/POL-63.md) | M3 | POL-06, POL-17 |
| [[POL-64] Implement teams multi tenant organization configuration management parity](issues/POL-64.md) | M3 | POL-06, POL-16 |
| [[POL-65] Implement teams network roaming policy management parity](issues/POL-65.md) | M3 | POL-06, POL-17 |
| [[POL-66] Implement teams notification and feeds policy management parity](issues/POL-66.md) | M3 | POL-06, POL-17 |
| [[POL-67] Implement teams personal attendant policy management parity](issues/POL-67.md) | M3 | POL-06, POL-17 |
| [[POL-68] Implement teams recording roll out policy management parity](issues/POL-68.md) | M3 | POL-06, POL-17 |
| [[POL-69] Implement teams room video tele conferencing policy management parity](issues/POL-69.md) | M3 | POL-06, POL-17 |
| [[POL-70] Implement teams shared calling routing policy management parity](issues/POL-70.md) | M3 | POL-06, POL-17 |
| [[POL-71] Implement teams shifts app policy management parity](issues/POL-71.md) | M3 | POL-06, POL-17 |
| [[POL-72] Implement teams shifts policy management parity](issues/POL-72.md) | M3 | POL-07, POL-17 |
| [[POL-73] Implement teams sip devices configuration management parity](issues/POL-73.md) | M3 | POL-07, POL-16 |
| [[POL-74] Implement teams template permission policy management parity](issues/POL-74.md) | M3 | POL-07, POL-17 |
| [[POL-75] Implement teams tenant abuse configuration management parity](issues/POL-75.md) | M3 | POL-07, POL-16 |
| [[POL-76] Implement teams update management policy management parity](issues/POL-76.md) | M3 | POL-07, POL-17 |
| [[POL-77] Implement teams upgrade configuration management parity](issues/POL-77.md) | M3 | POL-07, POL-16 |
| [[POL-78] Implement teams upgrade policy management parity](issues/POL-78.md) | M3 | POL-07, POL-17 |
| [[POL-79] Implement teams vdi policy management parity](issues/POL-79.md) | M3 | POL-07, POL-17 |
| [[POL-80] Implement teams virtual appointments policy management parity](issues/POL-80.md) | M3 | POL-07, POL-17 |
| [[POL-81] Implement teams voice applications policy management parity](issues/POL-81.md) | M3 | POL-07, POL-17 |
| [[POL-82] Implement teams work load policy management parity](issues/POL-82.md) | M3 | POL-08, POL-17 |
| [[POL-83] Implement teams work location detection policy management parity](issues/POL-83.md) | M3 | POL-08, POL-17 |
| [[POL-84] Implement teamsai policy management parity](issues/POL-84.md) | M3 | POL-08, POL-17 |
| [[POL-85] Implement teamsbyod and desks policy management parity](issues/POL-85.md) | M3 | POL-08, POL-17 |
| [[POL-86] Implement teamsip phone policy management parity](issues/POL-86.md) | M3 | POL-08, POL-17 |
| [[POL-87] Implement tenant migration configuration management parity](issues/POL-87.md) | M3 | POL-08, POL-16 |
| [[POL-90] Qualify integrated pol compound compatibility coverage](issues/POL-90.md) | M3 | POL-22, POL-23, POL-24, POL-27, POL-28, POL-30, POL-41 |
| [[POL-91] Qualify integrated pol configuration compatibility coverage](issues/POL-91.md) | M3 | POL-21, POL-26, POL-29, POL-33, POL-39, POL-43, POL-48, POL-49, POL-50, POL-51, POL-54, POL-56, POL-58, POL-60, POL-62, POL-64, POL-73, POL-75, POL-77, POL-87 |
| [[POL-92] Qualify integrated pol policy compatibility coverage](issues/POL-92.md) | M3 | POL-20, POL-25, POL-31, POL-32, POL-34, POL-35, POL-36, POL-37, POL-38, POL-40, POL-42, POL-44, POL-45, POL-46, POL-47, POL-52, POL-53, POL-55, POL-57, POL-59, POL-61, POL-63, POL-65, POL-66, POL-67, POL-68, POL-69, POL-70, POL-71, POL-72, POL-74, POL-76, POL-78, POL-79, POL-80, POL-81, POL-82, POL-83, POL-84, POL-85, POL-86 |

### VOICE

[Epic VOICE](issues/EPIC-VOICE.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[VOICE-01] Establish contracts for blocked_numbers, bulk_number_orders, calling_line_identity, calling_settings](issues/VOICE-01.md) | M0 | ARCH-01, BASE-02 |
| [[VOICE-02] Establish contracts for conferencing, dial_plans](issues/VOICE-02.md) | M0 | ARCH-01, BASE-02 |
| [[VOICE-03] Establish contracts for emergency_addresses, emergency_network](issues/VOICE-03.md) | M0 | ARCH-01, BASE-02 |
| [[VOICE-04] Establish contracts for emergency_policies, number_inventory, number_orders](issues/VOICE-04.md) | M0 | ARCH-01, BASE-02 |
| [[VOICE-05] Establish contracts for online_users, phone_assignments, phone_tags, pstn_gateways, survivability](issues/VOICE-05.md) | M0 | ARCH-01, BASE-02 |
| [[VOICE-06] Establish contracts for tenant_network, video_interop, voice_routes](issues/VOICE-06.md) | M0 | ARCH-01, BASE-02 |
| [[VOICE-15] Build voice resources contract-test fixtures and service harness](issues/VOICE-15.md) | M1 | BASE-03, HTTP-02, HTTP-03, VOICE-01 |
| [[VOICE-20] Implement blocked numbers management parity](issues/VOICE-20.md) | M3 | VOICE-15 |
| [[VOICE-21] Implement bulk number orders management parity](issues/VOICE-21.md) | M3 | HTTP-04, HTTP-05, VOICE-15 |
| [[VOICE-22] Implement calling line identity management parity](issues/VOICE-22.md) | M3 | VOICE-15 |
| [[VOICE-23] Implement calling settings management parity](issues/VOICE-23.md) | M3 | VOICE-15 |
| [[VOICE-24] Implement conferencing management parity](issues/VOICE-24.md) | M3 | VOICE-02, VOICE-15 |
| [[VOICE-25] Implement dial plans management parity](issues/VOICE-25.md) | M3 | VOICE-02, VOICE-15 |
| [[VOICE-26] Implement emergency addresses management parity](issues/VOICE-26.md) | M3 | VOICE-03, VOICE-15 |
| [[VOICE-27] Implement emergency network management parity](issues/VOICE-27.md) | M3 | VOICE-26 |
| [[VOICE-28] Implement emergency policies management parity](issues/VOICE-28.md) | M3 | VOICE-04, VOICE-15 |
| [[VOICE-29] Implement number inventory management parity](issues/VOICE-29.md) | M3 | HTTP-05, VOICE-04, VOICE-15 |
| [[VOICE-30] Implement number orders management parity](issues/VOICE-30.md) | M3 | HTTP-04, VOICE-04, VOICE-15 |
| [[VOICE-31] Implement online users management parity](issues/VOICE-31.md) | M2 | VOICE-05, VOICE-15 |
| [[VOICE-32] Implement phone assignments management parity](issues/VOICE-32.md) | M3 | VOICE-05, VOICE-15 |
| [[VOICE-33] Implement phone tags management parity](issues/VOICE-33.md) | M3 | VOICE-05, VOICE-15 |
| [[VOICE-34] Implement pstn gateways management parity](issues/VOICE-34.md) | M3 | VOICE-05, VOICE-15 |
| [[VOICE-35] Implement survivability management parity](issues/VOICE-35.md) | M3 | VOICE-05, VOICE-15 |
| [[VOICE-36] Implement tenant network management parity](issues/VOICE-36.md) | M3 | VOICE-06, VOICE-15 |
| [[VOICE-37] Implement video interop management parity](issues/VOICE-37.md) | M3 | VOICE-06, VOICE-15 |
| [[VOICE-38] Implement voice routes management parity](issues/VOICE-38.md) | M3 | VOICE-06, VOICE-15 |
| [[VOICE-90] Qualify integrated voice resources compatibility coverage](issues/VOICE-90.md) | M3 | VOICE-20, VOICE-21, VOICE-22, VOICE-23, VOICE-24, VOICE-25, VOICE-27, VOICE-28, VOICE-29, VOICE-30, VOICE-31, VOICE-32, VOICE-33, VOICE-34, VOICE-35, VOICE-36, VOICE-37, VOICE-38 |

### CALL

[Epic CALL](issues/EPIC-CALL.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[CALL-01] Establish contracts for attendant_builders, audio_prompts, auto_attendants, call_queues](issues/CALL-01.md) | M0 | ARCH-01, BASE-02 |
| [[CALL-02] Establish contracts for call_templates, compliance_applications](issues/CALL-02.md) | M0 | ARCH-01, BASE-02 |
| [[CALL-03] Establish contracts for mainline_attendants, resource_accounts, schedules](issues/CALL-03.md) | M0 | ARCH-01, BASE-02 |
| [[CALL-15] Build call resources contract-test fixtures and service harness](issues/CALL-15.md) | M1 | BASE-03, CALL-01, HTTP-02, HTTP-03 |
| [[CALL-20] Implement attendant builders management parity](issues/CALL-20.md) | M3 | CALL-15 |
| [[CALL-21] Implement audio prompts management parity](issues/CALL-21.md) | M3 | CALL-15, HTTP-05 |
| [[CALL-22] Implement auto attendants management parity](issues/CALL-22.md) | M3 | CALL-20, CALL-21, CALL-27, CALL-28 |
| [[CALL-23] Implement call queues management parity](issues/CALL-23.md) | M3 | CALL-21, CALL-27 |
| [[CALL-24] Implement call templates management parity](issues/CALL-24.md) | M3 | CALL-02, CALL-15 |
| [[CALL-25] Implement compliance applications management parity](issues/CALL-25.md) | M3 | CALL-02, CALL-15 |
| [[CALL-26] Implement mainline attendants management parity](issues/CALL-26.md) | M3 | CALL-03, CALL-15 |
| [[CALL-27] Implement resource accounts management parity](issues/CALL-27.md) | M3 | CALL-03, CALL-15, HTTP-04 |
| [[CALL-28] Implement schedules management parity](issues/CALL-28.md) | M3 | CALL-03, CALL-15, HTTP-05 |
| [[CALL-90] Qualify integrated call resources compatibility coverage](issues/CALL-90.md) | M3 | CALL-22, CALL-23, CALL-24, CALL-25, CALL-26 |

### EXT

[Epic EXT](issues/EPIC-EXT.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[EXT-01] Establish contracts for ai_agents, app_catalog, cloud_call_data, configuration_diagnostics, devices](issues/EXT-01.md) | M0 | ARCH-01, BASE-02 |
| [[EXT-02] Establish contracts for migration, provisioning_diagnostics, shifts_connections](issues/EXT-02.md) | M0 | ARCH-01, BASE-02 |
| [[EXT-03] Establish contracts for shifts_sync, tenant_identity, unified_apps](issues/EXT-03.md) | M0 | ARCH-01, BASE-02 |
| [[EXT-15] Build ext resources contract-test fixtures and service harness](issues/EXT-15.md) | M1 | BASE-03, EXT-01, HTTP-02, HTTP-03 |
| [[EXT-20] Implement ai agents management parity](issues/EXT-20.md) | M3 | EXT-15 |
| [[EXT-21] Implement app catalog management parity](issues/EXT-21.md) | M3 | EXT-15, HTTP-05 |
| [[EXT-22] Implement cloud call data management parity](issues/EXT-22.md) | M3 | EXT-15 |
| [[EXT-23] Implement configuration diagnostics management parity](issues/EXT-23.md) | M3 | EXT-15 |
| [[EXT-24] Implement devices management parity](issues/EXT-24.md) | M3 | EXT-15, HTTP-04 |
| [[EXT-25] Implement migration management parity](issues/EXT-25.md) | M3 | EXT-02, EXT-15, HTTP-04 |
| [[EXT-26] Implement provisioning diagnostics management parity](issues/EXT-26.md) | M3 | EXT-02, EXT-15 |
| [[EXT-27] Implement shifts connections management parity](issues/EXT-27.md) | M3 | EXT-02, EXT-15, HTTP-04 |
| [[EXT-28] Implement shifts sync management parity](issues/EXT-28.md) | M3 | EXT-03, EXT-27, HTTP-05 |
| [[EXT-29] Implement tenant identity management parity](issues/EXT-29.md) | M3 | EXT-03, EXT-15 |
| [[EXT-30] Implement unified apps management parity](issues/EXT-30.md) | M3 | EXT-03, EXT-15 |
| [[EXT-90] Qualify integrated ext resources compatibility coverage](issues/EXT-90.md) | M3 | EXT-20, EXT-21, EXT-22, EXT-23, EXT-24, EXT-25, EXT-26, EXT-28, EXT-29, EXT-30 |

### DELTA

[Epic DELTA](issues/EPIC-DELTA.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[DELTA-01] Implement semantic baseline comparison and SDK impact reports](issues/DELTA-01.md) | M4 | BASE-03 |
| [[DELTA-02] Schedule upstream discovery and maintain candidate draft PRs](issues/DELTA-02.md) | M4 | DELTA-01 |
| [[DELTA-03] Detect sampled server drift independently of module releases](issues/DELTA-03.md) | M4 | DELTA-01, LAB-02 |

### QUAL

[Epic QUAL](issues/EPIC-QUAL.md)

| Task | Gate | Direct blockers |
| --- | --- | --- |
| [[QUAL-01] Audit complete implementation and contract coverage](issues/QUAL-01.md) | M5 | AUTH-06, CALL-90, CORE-90, EXT-90, POL-90, POL-91, POL-92, VOICE-90 |
| [[QUAL-02] Qualify concurrency, security and available lab regressions](issues/QUAL-02.md) | M5 | AUTH-05, HTTP-06, LAB-03, QUAL-01 |
| [[QUAL-03] Publish SDK guides, CI recipes and compatibility documentation](issues/QUAL-03.md) | M5 | AUTH-05, DELTA-02, QUAL-01 |
| [[QUAL-04] Gate and prepare the first supported SDK release](issues/QUAL-04.md) | M5 | DELTA-03, QUAL-02, QUAL-03 |

## Local validation

Run `python3 tools/backlog/build.py` to deterministically render the preview and `python3 tools/backlog/validate.py` to audit bodies, hashes, coverage, parent membership, references, cycles and milestone ordering. These checks verify planning artifacts, not SDK behaviour.
