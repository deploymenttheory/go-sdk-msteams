<!-- weave-backlog:msteams-powershell-parity:v1:FND-02 -->
Backlog ID: `FND-02` · Epic: [EPIC-FND](EPIC-FND.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Wire the root client, configuration and typed service composition. This provides the sdk contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Add NewClient(authConfig, options...) and NewClientFromEnv with explicit authentication selection and functional options for HTTP settings. Wire typed services through constructor injection. Validate incompatible configuration locally, copy mutable configuration, keep tenant identity scoped per client, and expose session/cache cleanup without claiming Entra token revocation.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Clients for different tenants cannot share mutable credentials or routes; construction errors are actionable; explicit configuration works without environment variables.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Test environment parsing, explicit-over-environment behavior at the documented constructors, missing credentials, incompatible options, context lifecycle and independent concurrent clients.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [FND-01](FND-01.md) — provides initialize the go module and cross-platform quality checks.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M1 SDK foundations.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: sdk; FND workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(fnd): wire the root client, configuration and typed service composition`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
