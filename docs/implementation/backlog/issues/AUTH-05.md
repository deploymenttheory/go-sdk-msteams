<!-- weave-backlog:msteams-powershell-parity:v1:AUTH-05 -->
Backlog ID: `AUTH-05` · Epic: [EPIC-AUTH](EPIC-AUTH.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Qualify permissions and authentication capability reporting. This provides the verification contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Create the operation/authentication matrix using pinned evidence and current observations kept distinct. Cover module app-only exceptions, roles and consent requirements. Distinguish unauthenticated, forbidden, unsupported auth mode and unavailable tenant capability. Document supplied-token and federation additions without claiming the module directly exposes every provider.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Each flow has portable documentation and each known command restriction has a contract test; authentication success alone is never reported as operation authorization.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Exercise 401/403, missing role, delegated-only operation and same-type dual-token scenarios. Attach lab results where available and contract evidence elsewhere.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [AUTH-02](AUTH-02.md) — provides add portable certificate and client-secret authentication.
- [AUTH-03](AUTH-03.md) — provides add portable delegated login and credential flows.
- [AUTH-04](AUTH-04.md) — provides add managed identity and ci workload federation.
- [BASE-02](BASE-02.md) — provides reconcile all exported and conditional command candidates.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M1 SDK foundations.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: verification; AUTH workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `test(auth): qualify permissions and authentication capability reporting`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
