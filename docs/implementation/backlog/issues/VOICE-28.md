<!-- weave-backlog:msteams-powershell-parity:v1:VOICE-28 -->
Backlog ID: `VOICE-28` · Epic: [EPIC-VOICE](EPIC-VOICE.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Implement emergency policies management parity. This provides the sdk contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Implement the assigned commands using the settled contracts from VOICE-04. Add the planned emergency_policies resource package in the contract-selected graph_api or teams_admin_api family, constructor/root-client wiring, typed request/response models, context-aware methods, colocated mocks/tests, api_docs.md and compiling examples. Use shared transport controls rather than duplicating authentication/retry loops. Keep Go methods independent of PowerShell release numbers. Map parameter sets and local constructors explicitly; preserve omitted/null/zero/empty distinctions. Preserve all exposed resource fields, identity/scope rules and command-specific validation; implement grant/assignment or local builder behaviour when present in this command set. Update each command/parameter-set ledger row with Go symbols and verification evidence. Keep unrelated resource families and shared-transport changes outside this task; defects in shared contracts go to their owners.

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

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Every assigned public management command and parameter set is implemented, with typed models and evidence-backed validation; success/authorization/failure tests pass; examples compile and permissions are documented; contract-tested and live-tested status remain distinct. A stub or generic unsupported error does not satisfy parity.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run go test on the affected resource package and its mocks, compile the corresponding examples, and run the coverage-ledger check. Test exact method/path/query/body/headers against research fixtures; cover invalid identity, missing required arguments, auth denial, malformed/error responses and cancellation. For writes, verify non-default read-back, omission/null semantics and available cleanup; for local builders, prove no HTTP call occurs. Exercise evidenced paging/async/upload/batch behaviour where present. Add opt-in PowerShell/Go acceptance cases; record unavailable live prerequisites without dropping implementation.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [VOICE-04](VOICE-04.md) — provides establish contracts for emergency_policies, number_inventory, number_orders.
- [VOICE-15](VOICE-15.md) — provides build voice resources contract-test fixtures and service harness.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M3 Complete management surface.
Priority: P1 — required release capability.
Component / scope owner: sdk; VOICE workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(voice): implement emergency policies management parity`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
