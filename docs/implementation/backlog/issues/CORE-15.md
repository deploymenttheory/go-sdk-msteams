<!-- weave-backlog:msteams-powershell-parity:v1:CORE-15 -->
Backlog ID: `CORE-15` · Epic: [EPIC-CORE](EPIC-CORE.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Build core resources contract-test fixtures and service harness. This provides the verification contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Provide the reusable offline fixture runner, backend route/audience assertions, response/error comparators and ledger evidence writer consumed by resource tests for: channel_members, channels, deployment_migration, operations, shared_channels, targeting, team_members, team_templates, teams. Build the initial runner against a representative settled contract; resource tickets add their own exact fixtures. Preserve scope/identity, query semantics, omission/null and authorization errors; do not normalize meaningful ordering or hide unknown fields. Local builders use local-only assertions. This owns test infrastructure, not production resource CRUD.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Each family can register a deterministic case and receive exact wire/evidence comparison; a deliberate route, audience, serialization or error mismatch fails with a useful diff.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Execute representative positive/negative fixture cases without tenant credentials. Verify no HTTP call for local constructors, proper cancellation and no raw secrets in generated evidence.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [BASE-03](BASE-03.md) — provides maintain the capability and verification coverage ledger.
- [CORE-01](CORE-01.md) — provides establish contracts for channel_members, channels, deployment_migration, operations, shared_channels, targeting, team_members, team_templates, teams.
- [HTTP-02](HTTP-02.md) — provides implement bounded retries, throttling and concurrency controls.
- [HTTP-03](HTTP-03.md) — provides implement backend-specific pagination and list traversal.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M1 SDK foundations.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: verification; CORE workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(core): build core resources contract-test fixtures and service harness`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
