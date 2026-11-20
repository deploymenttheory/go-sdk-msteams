<!-- weave-backlog:msteams-powershell-parity:v1:CORE-28 -->
Backlog ID: `CORE-28` · Epic: [EPIC-CORE](EPIC-CORE.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Implement teams management parity. This provides the sdk contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Implement the assigned commands using the settled contracts from CORE-01. Add the planned teams resource package in the contract-selected graph_api or teams_admin_api family, constructor/root-client wiring, typed request/response models, context-aware methods, colocated mocks/tests, api_docs.md and compiling examples. Use shared transport controls rather than duplicating authentication/retry loops. Keep Go methods independent of PowerShell release numbers. Map parameter sets and local constructors explicitly; preserve omitted/null/zero/empty distinctions. Preserve group/team identity relationships, archive transitions and picture upload; observe asynchronous creation and resulting membership. Update each command/parameter-set ledger row with Go symbols and verification evidence. Keep unrelated resource families and shared-transport changes outside this task; defects in shared contracts go to their owners.

Assigned baseline commands:

- `Get-Team` — 2 observed parameter sets.
- `New-Team` — 2 observed parameter sets.
- `Remove-Team` — 1 observed parameter sets.
- `Set-Team` — 1 observed parameter sets.
- `Set-TeamArchivedState` — 1 observed parameter sets.
- `Set-TeamPicture` — 1 observed parameter sets.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Every assigned public management command and parameter set is implemented, with typed models and evidence-backed validation; success/authorization/failure tests pass; examples compile and permissions are documented; contract-tested and live-tested status remain distinct. A stub or generic unsupported error does not satisfy parity.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Run go test on the affected resource package and its mocks, compile the corresponding examples, and run the coverage-ledger check. Test exact method/path/query/body/headers against research fixtures; cover invalid identity, missing required arguments, auth denial, malformed/error responses and cancellation. For writes, verify non-default read-back, omission/null semantics and available cleanup; for local builders, prove no HTTP call occurs. Exercise evidenced paging/async/upload/batch behaviour where present. Add opt-in PowerShell/Go acceptance cases; record unavailable live prerequisites without dropping implementation.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [CORE-15](CORE-15.md) — provides build core resources contract-test fixtures and service harness.
- [HTTP-04](HTTP-04.md) — provides implement long-running operations and eventual-consistency waits.
- [HTTP-05](HTTP-05.md) — provides implement binary, multipart and batch request behaviours.

Supporting contracts covered transitively by the direct blockers: [CORE-01](CORE-01.md). These are context, not additional native blockers.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M2 Representative proof.
Priority: P1 — required release capability.
Component / scope owner: sdk; CORE workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(core): implement teams management parity`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
