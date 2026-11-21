<!-- weave-backlog:msteams-powershell-parity:v1:CALL-03 -->
Backlog ID: `CALL-03` · Epic: [EPIC-CALL](EPIC-CALL.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Establish contracts for mainline_attendants, resource_accounts, schedules. This provides the research contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Own the exact command set below for families mainline_attendants, resource_accounts, schedules. Read the pinned package scripts, command configuration, help and binary metadata plus Microsoft documentation. Record operation routes/methods, authority/audience, discovery, parameters/defaults/validation, null/omission rules, response/error shapes, auth/role restrictions, pagination, async/retry/transfer semantics and local-only builders. Use sanitized PowerShell lab observations when available. Keep pinned command metadata under contracts/powershell. Store sanitized HTTP response fixtures from TestAcc directly under each service testdata directory and load them in TestUnit tests. Keep request and validation inputs in the tests; do not generate synthetic response captures. Resolve the canonical Go operations and API-family routing within the agreed architecture before downstream implementation starts. Do not infer a wire route from a command name or convert lack of evidence into a fabricated endpoint.

Assigned baseline commands:

- `Get-CsMainlineAttendantAppointmentBookingFlow` — 1 observed parameter sets.
- `Get-CsMainlineAttendantFlow` — 1 observed parameter sets.
- `Get-CsMainlineAttendantQuestionAnswerFlow` — 1 observed parameter sets.
- `Get-CsMainlineAttendantSpamDetectionTemplate` — 1 observed parameter sets.
- `Get-CsMainlineAttendantSupportedLanguages` — 1 observed parameter sets.
- `Get-CsMainlineAttendantSupportedVoices` — 1 observed parameter sets.
- `Get-CsMainlineAttendantTenantInformation` — 1 observed parameter sets.
- `New-CsMainlineAttendantAppointmentBookingFlow` — 1 observed parameter sets.
- `New-CsMainlineAttendantQuestionAnswerFlow` — 1 observed parameter sets.
- `New-CsMainlineAttendantSpamDetectionTemplate` — 1 observed parameter sets.
- `Remove-CsMainlineAttendantAppointmentBookingFlow` — 1 observed parameter sets.
- `Remove-CsMainlineAttendantQuestionAnswerFlow` — 1 observed parameter sets.
- `Remove-CsMainlineAttendantSpamDetectionTemplate` — 1 observed parameter sets.
- `Set-CsMainlineAttendantAppointmentBookingFlow` — 1 observed parameter sets.
- `Set-CsMainlineAttendantQuestionAnswerFlow` — 1 observed parameter sets.
- `Set-CsMainlineAttendantSpamDetectionTemplate` — 1 observed parameter sets.
- `Find-CsOnlineApplicationInstance` — 1 observed parameter sets.
- `Get-CsOnlineApplicationInstance` — 1 observed parameter sets.
- `Get-CsOnlineApplicationInstanceAssociation` — 1 observed parameter sets.
- `Get-CsOnlineApplicationInstanceAssociationStatus` — 1 observed parameter sets.
- `New-CsOnlineApplicationInstance` — 1 observed parameter sets.
- `New-CsOnlineApplicationInstanceAssociation` — 1 observed parameter sets.
- `Remove-CsOnlineApplicationInstanceAssociation` — 1 observed parameter sets.
- `Set-CsOnlineApplicationInstance` — 1 observed parameter sets.
- `Set-CsOnlineVoiceApplicationInstance` — 1 observed parameter sets.
- `Sync-CsOnlineApplicationInstance` — 1 observed parameter sets.
- `Export-CsAutoAttendantHolidays` — 1 observed parameter sets.
- `Get-CsAutoAttendantHolidays` — 1 observed parameter sets.
- `Get-CsOnlineSchedule` — 1 observed parameter sets.
- `Import-CsAutoAttendantHolidays` — 1 observed parameter sets.
- `New-CsOnlineDateTimeRange` — 1 observed parameter sets.
- `New-CsOnlineSchedule` — 3 observed parameter sets.
- `New-CsOnlineTimeRange` — 1 observed parameter sets.
- `Remove-CsOnlineSchedule` — 1 observed parameter sets.
- `Set-CsOnlineSchedule` — 1 observed parameter sets.

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
Component / scope owner: research; CALL workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `docs(call): establish contracts for mainline_attendants, resource_accounts, schedules`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
