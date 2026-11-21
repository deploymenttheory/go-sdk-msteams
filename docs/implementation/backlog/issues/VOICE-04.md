<!-- weave-backlog:msteams-powershell-parity:v1:VOICE-04 -->
Backlog ID: `VOICE-04` · Epic: [EPIC-VOICE](EPIC-VOICE.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Establish contracts for emergency_policies, number_inventory, number_orders. This provides the research contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Own the exact command set below for families emergency_policies, number_inventory, number_orders. Read the pinned package scripts, command configuration, help and binary metadata plus Microsoft documentation. Record operation routes/methods, authority/audience, discovery, parameters/defaults/validation, null/omission rules, response/error shapes, auth/role restrictions, pagination, async/retry/transfer semantics and local-only builders. Use sanitized PowerShell lab observations when available. Keep pinned command metadata under contracts/powershell. Store sanitized HTTP response fixtures from TestAcc directly under each service testdata directory and load them in TestUnit tests. Keep request and validation inputs in the tests; do not generate synthetic response captures. Resolve the canonical Go operations and API-family routing within the agreed architecture before downstream implementation starts. Do not infer a wire route from a command name or convert lack of evidence into a fabricated endpoint.

Assigned baseline commands:

- `Get-CsLocationPolicy` — 2 observed parameter sets.
- `Get-CsTeamsCarrierEmergencyCallRoutingPolicy` — 2 observed parameter sets.
- `Get-CsTeamsEmergencyCallingPolicy` — 2 observed parameter sets.
- `Get-CsTeamsEmergencyCallRoutingPolicy` — 2 observed parameter sets.
- `Grant-CsTeamsCarrierEmergencyCallRoutingPolicy` — 4 observed parameter sets.
- `Grant-CsTeamsEmergencyCallingPolicy` — 3 observed parameter sets.
- `Grant-CsTeamsEmergencyCallRoutingPolicy` — 3 observed parameter sets.
- `New-CsLocationPolicy` — 1 observed parameter sets.
- `New-CsTeamsCarrierEmergencyCallRoutingPolicy` — 1 observed parameter sets.
- `New-CsTeamsEmergencyCallingExtendedNotification` — 1 observed parameter sets.
- `New-CsTeamsEmergencyCallingPolicy` — 1 observed parameter sets.
- `New-CsTeamsEmergencyCallRoutingPolicy` — 1 observed parameter sets.
- `New-CsTeamsEmergencyNumber` — 1 observed parameter sets.
- `Remove-CsLocationPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsCarrierEmergencyCallRoutingPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsEmergencyCallingPolicy` — 1 observed parameter sets.
- `Remove-CsTeamsEmergencyCallRoutingPolicy` — 1 observed parameter sets.
- `Set-CsLocationPolicy` — 1 observed parameter sets.
- `Set-CsTeamsCarrierEmergencyCallRoutingPolicy` — 1 observed parameter sets.
- `Set-CsTeamsEmergencyCallingPolicy` — 1 observed parameter sets.
- `Set-CsTeamsEmergencyCallRoutingPolicy` — 1 observed parameter sets.
- `Export-CsAcquiredPhoneNumber` — 1 observed parameter sets.
- `Get-CsExportAcquiredPhoneNumberStatus` — 2 observed parameter sets.
- `Get-CsHybridTelephoneNumber` — 3 observed parameter sets.
- `Get-CsOnlineTelephoneNumber` — 1 observed parameter sets.
- `Get-CsOnlineTelephoneNumberCountry` — 1 observed parameter sets.
- `Get-CsOnlineTelephoneNumberType` — 2 observed parameter sets.
- `New-CsHybridTelephoneNumber` — 2 observed parameter sets.
- `Remove-CsHybridTelephoneNumber` — 2 observed parameter sets.
- `Remove-CsOnlineTelephoneNumber` — 1 observed parameter sets.
- `Clear-CsOnlineTelephoneNumberOrder` — 1 observed parameter sets.
- `Complete-CsOnlineTelephoneNumberOrder` — 1 observed parameter sets.
- `Get-CsOnlineTelephoneNumberOrder` — 2 observed parameter sets.
- `New-CsOnlineDirectRoutingTelephoneNumberUploadOrder` — 3 observed parameter sets.
- `New-CsOnlineTelephoneNumberOrder` — 2 observed parameter sets.
- `New-CsOnlineTelephoneNumberReleaseOrder` — 3 observed parameter sets.
- `New-CsPhoneNumberUsageChangeOrder` — 2 observed parameter sets.

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
Component / scope owner: research; VOICE workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `docs(voice): establish contracts for emergency_policies, number_inventory, number_orders`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
