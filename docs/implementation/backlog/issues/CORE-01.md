<!-- weave-backlog:msteams-powershell-parity:v1:CORE-01 -->
Backlog ID: `CORE-01` · Epic: [EPIC-CORE](EPIC-CORE.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Establish contracts for channel_members, channels, deployment_migration, operations, shared_channels, targeting, team_members, team_templates, teams. This provides the research contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Own the exact command set below for families channel_members, channels, deployment_migration, operations, shared_channels, targeting, team_members, team_templates, teams. Read the pinned package scripts, command configuration, help and binary metadata plus Microsoft documentation. Record operation routes/methods, authority/audience, discovery, parameters/defaults/validation, null/omission rules, response/error shapes, auth/role restrictions, pagination, async/retry/transfer semantics and local-only builders. Use sanitized PowerShell lab observations when available. Keep pinned command metadata under contracts/powershell. Store sanitized HTTP response fixtures from TestAcc directly under each service testdata directory and load them in TestUnit tests. Keep request and validation inputs in the tests; do not generate synthetic response captures. Resolve the canonical Go operations and API-family routing within the agreed architecture before downstream implementation starts. Do not infer a wire route from a command name or convert lack of evidence into a fabricated endpoint.

Assigned baseline commands:

- `Add-TeamChannelUser` — 1 observed parameter sets.
- `Get-TeamChannelUser` — 1 observed parameter sets.
- `Remove-TeamChannelUser` — 1 observed parameter sets.
- `Get-TeamAllChannel` — 1 observed parameter sets.
- `Get-TeamChannel` — 1 observed parameter sets.
- `Get-TeamIncomingChannel` — 1 observed parameter sets.
- `New-TeamChannel` — 1 observed parameter sets.
- `Remove-TeamChannel` — 1 observed parameter sets.
- `Set-TeamChannel` — 1 observed parameter sets.
- `Get-CsBatchTeamsDeploymentStatus` — 2 observed parameter sets.
- `Get-MultiGeoRegion` — 1 observed parameter sets.
- `Get-TenantPrivateChannelMigrationStatus` — 1 observed parameter sets.
- `New-CsBatchTeamsDeployment` — 1 observed parameter sets.
- `Get-Operation` — 1 observed parameter sets.
- `Get-AssociatedTeam` — 1 observed parameter sets.
- `Get-SharedWithTeam` — 1 observed parameter sets.
- `Get-SharedWithTeamUser` — 1 observed parameter sets.
- `Remove-SharedWithTeam` — 1 observed parameter sets.
- `Get-CsTeamsTargetingPolicy` — 2 observed parameter sets.
- `Get-TeamTargetingHierarchyStatus` — 1 observed parameter sets.
- `Remove-CsTeamsTargetingPolicy` — 1 observed parameter sets.
- `Remove-TeamTargetingHierarchy` — 1 observed parameter sets.
- `Set-CsTeamsTargetingPolicy` — 1 observed parameter sets.
- `Set-TeamTargetingHierarchy` — 1 observed parameter sets.
- `Add-TeamUser` — 1 observed parameter sets.
- `Get-TeamUser` — 1 observed parameter sets.
- `Remove-TeamUser` — 1 observed parameter sets.
- `Get-CsTeamTemplate` — 2 observed parameter sets.
- `Get-CsTeamTemplateList` — 1 observed parameter sets.
- `New-CsTeamsHiddenMeetingTemplate` — 1 observed parameter sets.
- `New-CsTeamsHiddenTemplate` — 1 observed parameter sets.
- `New-CsTeamTemplate` — 4 observed parameter sets.
- `Remove-CsTeamTemplate` — 2 observed parameter sets.
- `Update-CsTeamTemplate` — 4 observed parameter sets.
- `Get-Team` — 2 observed parameter sets.
- `New-Team` — 2 observed parameter sets.
- `Remove-Team` — 1 observed parameter sets.
- `Set-Team` — 1 observed parameter sets.
- `Set-TeamArchivedState` — 1 observed parameter sets.
- `Set-TeamPicture` — 1 observed parameter sets.

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
Component / scope owner: research; CORE workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `docs(core): establish contracts for channel_members, channels, deployment_migration, operations, shared_channels, targeting, team_members, team_templates, teams`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
