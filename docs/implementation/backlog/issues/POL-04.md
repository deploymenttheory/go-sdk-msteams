<!-- weave-backlog:msteams-powershell-parity:v1:POL-04 -->
Backlog ID: `POL-04` · Epic: [EPIC-POL](EPIC-POL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Establish pol contracts from teams_custom_banner_text through teams_guest_messaging_configuration. This provides the research contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Own the exact command set below for families teams_custom_banner_text, teams_education_assignments_app_policy, teams_education_configuration, teams_enhanced_encryption_policy, teams_events_policy, teams_feedback_policy, teams_files_policy, teams_first_party_meeting_template_configuration, teams_guest_calling_configuration, teams_guest_meeting_configuration, teams_guest_messaging_configuration. Read the pinned package scripts, command configuration, help and binary metadata plus Microsoft documentation. Record operation routes/methods, authority/audience, discovery, parameters/defaults/validation, null/omission rules, response/error shapes, auth/role restrictions, pagination, async/retry/transfer semantics and local-only builders. Use sanitized PowerShell lab observations when available. Produce versioned contracts and representative request/response fixtures under contracts/endpoints, with source path/hash or documentation revision and confidence. Resolve the canonical Go operations and API-family routing within the agreed architecture before downstream implementation starts. Do not infer a wire route from a command name or convert lack of evidence into a fabricated endpoint.

Assigned baseline commands:

- `Get-CsTeamsCustomBannerText` — 3 observed parameter sets.
- `New-CsTeamsCustomBannerText` — 2 observed parameter sets.
- `Remove-CsTeamsCustomBannerText` — 2 observed parameter sets.
- `Set-CsTeamsCustomBannerText` — 2 observed parameter sets.
- `Get-CsTeamsEducationAssignmentsAppPolicy` — 2 observed parameter sets.
- `Set-CsTeamsEducationAssignmentsAppPolicy` — 1 observed parameter sets.
- `Get-CsTeamsEducationConfiguration` — 2 observed parameter sets.
- `Set-CsTeamsEducationConfiguration` — 1 observed parameter sets.
- `Get-CsTeamsEnhancedEncryptionPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsEnhancedEncryptionPolicy` — 3 observed parameter sets.
- `New-CsTeamsEnhancedEncryptionPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsEnhancedEncryptionPolicy` — 1 observed parameter sets.
- `Set-CsTeamsEnhancedEncryptionPolicy` — 1 observed parameter sets.
- `Get-CsTeamsEventsPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsEventsPolicy` — 4 observed parameter sets.
- `New-CsTeamsEventsPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsEventsPolicy` — 1 observed parameter sets.
- `Set-CsTeamsEventsPolicy` — 1 observed parameter sets.
- `Get-CsTeamsFeedbackPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsFeedbackPolicy` — 3 observed parameter sets.
- `New-CsTeamsFeedbackPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsFeedbackPolicy` — 1 observed parameter sets.
- `Set-CsTeamsFeedbackPolicy` — 1 observed parameter sets.
- `Get-CsTeamsFilesPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsFilesPolicy` — 4 observed parameter sets.
- `New-CsTeamsFilesPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsFilesPolicy` — 1 observed parameter sets.
- `Set-CsTeamsFilesPolicy` — 1 observed parameter sets.
- `Get-CsTeamsFirstPartyMeetingTemplateConfiguration` — 2 observed parameter sets.
- `Set-CsTeamsFirstPartyMeetingTemplateConfiguration` — 1 observed parameter sets.
- `Get-CsTeamsGuestCallingConfiguration` — 2 observed parameter sets.
- `Set-CsTeamsGuestCallingConfiguration` — 1 observed parameter sets.
- `Get-CsTeamsGuestMeetingConfiguration` — 2 observed parameter sets.
- `Set-CsTeamsGuestMeetingConfiguration` — 1 observed parameter sets.
- `Get-CsTeamsGuestMessagingConfiguration` — 2 observed parameter sets.
- `Set-CsTeamsGuestMessagingConfiguration` — 1 observed parameter sets.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Every assigned candidate has a settled contract or evidenced internal/conditional disposition; implementation prerequisites include all necessary types and special behaviours; unknown contracts keep this task open and block affected implementations.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Check every command and parameter set against the normalized snapshot; review positive, validation, authorization and failure examples. Confirm fixtures are evidence-derived and sanitized. Where the lab is unavailable, cite package/documentation evidence and state that live behaviour remains unobserved.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [ARCH-01](ARCH-01.md) — provides record client, authentication and backend architecture contracts.
- [BASE-02](BASE-02.md) — provides reconcile all exported and conditional command candidates.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M0 Baseline and contracts.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: research; POL workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `docs(pol): establish pol contracts from teams_custom_banner_text through teams_guest_messaging_configuration`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
