<!-- weave-backlog:msteams-powershell-parity:v1:AUTH-06 -->
Backlog ID: `AUTH-06` · Epic: [EPIC-AUTH](EPIC-AUTH.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Map PowerShell connection and environment helpers to portable SDK sessions. This provides the sdk contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Map every assigned helper to explicit root-client configuration, cloud/region routing or local cache lifecycle. Resolve conditional OCE/region contexts in BASE-02 before reproducing public behaviour. Document portable certificate/browser equivalents and common-parameter mapping. Disconnect clears client-owned state, not globally shared identities; do not pretend it remotely revokes Entra access tokens.

Assigned baseline commands:

- `Clear-CsOCEContext` — not observed in the unauthenticated macOS import; reconcile before wire implementation.
- `Clear-CsRegionContext` — not observed in the unauthenticated macOS import; reconcile before wire implementation.
- `Clear-TeamsEnvironmentConfig` — 1 observed parameter sets.
- `Connect-MicrosoftTeams` — 4 observed parameter sets.
- `Disconnect-MicrosoftTeams` — 1 observed parameter sets.
- `Get-CsRegionContext` — not observed in the unauthenticated macOS import; reconcile before wire implementation.
- `Set-CsOCEContext` — not observed in the unauthenticated macOS import; reconcile before wire implementation.
- `Set-CsRegionContext` — not observed in the unauthenticated macOS import; reconcile before wire implementation.
- `Set-TeamsEnvironmentConfig` — 1 observed parameter sets.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Each assigned command has an explicit SDK equivalent or reviewed internal-context disposition; concurrent clients retain independent tenant/region settings.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Test client isolation, configuration replacement, cache clearing and invalid cloud/region settings. Compare session behaviour to 8.1.0 where observable without tenant writes.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [AUTH-01](AUTH-01.md) — provides implement resource-aware supplied tokens and refresh coordination.
- [BASE-02](BASE-02.md) — provides reconcile all exported and conditional command candidates.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M1 SDK foundations.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: sdk; AUTH workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(auth): map PowerShell connection and environment helpers to portable SDK sessions`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
