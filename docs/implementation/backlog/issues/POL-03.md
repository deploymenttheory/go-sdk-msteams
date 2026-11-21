<!-- weave-backlog:msteams-powershell-parity:v1:POL-03 -->
Backlog ID: `POL-03` · Epic: [EPIC-POL](EPIC-POL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Establish pol contracts from teams_app_setup_policy through teams_cortana_policy. This provides the research contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Own the exact command set below for families teams_app_setup_policy, teams_audio_conferencing_custom_prompts_configuration, teams_audio_conferencing_policy, teams_call_hold_policy, teams_call_park_policy, teams_calling_policy, teams_channels_policy, teams_client_configuration, teams_cortana_policy. Read the pinned package scripts, command configuration, help and binary metadata plus Microsoft documentation. Record operation routes/methods, authority/audience, discovery, parameters/defaults/validation, null/omission rules, response/error shapes, auth/role restrictions, pagination, async/retry/transfer semantics and local-only builders. Use sanitized PowerShell lab observations when available. Keep pinned command metadata under contracts/powershell. Store sanitized HTTP response fixtures from TestAcc directly under each service testdata directory and load them in TestUnit tests. Keep request and validation inputs in the tests; do not generate synthetic response captures. Resolve the canonical Go operations and API-family routing within the agreed architecture before downstream implementation starts. Do not infer a wire route from a command name or convert lack of evidence into a fabricated endpoint.

Assigned baseline commands:

- `Get-CsTeamsAppSetupPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsAppSetupPolicy` — 4 observed parameter sets.
- `New-CsTeamsAppSetupPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsAppSetupPolicy` — 1 observed parameter sets.
- `Set-CsTeamsAppSetupPolicy` — 1 observed parameter sets.
- `Get-CsTeamsAudioConferencingCustomPromptsConfiguration` — 2 observed parameter sets.
- `Set-CsTeamsAudioConferencingCustomPromptsConfiguration` — 1 observed parameter sets.
- `Get-CsTeamsAudioConferencingPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsAudioConferencingPolicy` — 3 observed parameter sets.
- `New-CsTeamsAudioConferencingPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsAudioConferencingPolicy` — 1 observed parameter sets.
- `Set-CsTeamsAudioConferencingPolicy` — 1 observed parameter sets.
- `Get-CsTeamsCallHoldPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsCallHoldPolicy` — 3 observed parameter sets.
- `New-CsTeamsCallHoldPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsCallHoldPolicy` — 1 observed parameter sets.
- `Set-CsTeamsCallHoldPolicy` — 1 observed parameter sets.
- `Get-CsTeamsCallParkPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsCallParkPolicy` — 3 observed parameter sets.
- `New-CsTeamsCallParkPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsCallParkPolicy` — 1 observed parameter sets.
- `Set-CsTeamsCallParkPolicy` — 1 observed parameter sets.
- `Get-CsTeamsCallingPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsCallingPolicy` — 4 observed parameter sets.
- `New-CsTeamsCallingPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsCallingPolicy` — 1 observed parameter sets.
- `Set-CsTeamsCallingPolicy` — 1 observed parameter sets.
- `Get-CsTeamsChannelsPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsChannelsPolicy` — 3 observed parameter sets.
- `New-CsTeamsChannelsPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsChannelsPolicy` — 1 observed parameter sets.
- `Set-CsTeamsChannelsPolicy` — 1 observed parameter sets.
- `Get-CsTeamsClientConfiguration` — 2 observed parameter sets.
- `Set-CsTeamsClientConfiguration` — 1 observed parameter sets.
- `Get-CsTeamsCortanaPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsCortanaPolicy` — 3 observed parameter sets.
- `New-CsTeamsCortanaPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsCortanaPolicy` — 1 observed parameter sets.
- `Set-CsTeamsCortanaPolicy` — 1 observed parameter sets.

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
Suggested PR title: `docs(pol): establish pol contracts from teams_app_setup_policy through teams_cortana_policy`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
