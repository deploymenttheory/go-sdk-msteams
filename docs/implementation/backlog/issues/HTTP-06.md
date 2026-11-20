<!-- weave-backlog:msteams-powershell-parity:v1:HTTP-06 -->
Backlog ID: `HTTP-06` · Epic: [EPIC-HTTP](EPIC-HTTP.md) · Product baseline: [INDEX](INDEX.md)

## Outcome

Add redacted logging, tracing and transport configuration. This provides the sdk contract required for complete MicrosoftTeams 8.1.0 management parity. Starting SDK state: template-only; checked discovery metadata is evidence, not a delivered service.

## Implementation

Expose Jamf-style options for timeout/proxy/TLS/custom transport/headers/user agent/concurrency and Zap logging with optional OTel hooks. Redact tokens, assertions, passwords, certificate keys and sensitive payload fields in every logging mode. Tracing is opt-in and uses the caller lifecycle; secret headers cannot be overridden accidentally by generic options.

Source: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0); normalized command and parameter-set evidence in `contracts/powershell/8.1.0/commands.json` and `runtime.json`. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`. Planned paths are future implementation locations unless already present in the baseline preview.

## Acceptance criteria

- [ ] Debug logs and spans remain sanitized; custom transport and proxy configuration work without bypassing auth/cancellation controls.
- [ ] Evidence, limitations and command/test mappings are included with the implementation; no tenant secrets are retained.

## Verification

Capture logs/spans with representative secret fixtures, TLS/proxy tests and custom transport tests; assert absence of raw secrets in success and error paths.

Go checks and SDK fixture paths are planned foundation outputs, not claims of tests already run. The implementation PR records the actual command results and evidence revision.

## Dependencies

Direct blockers:

- [HTTP-01](HTTP-01.md) — provides implement backend routing, request building and normalized errors.

## Applicable product rules

Native Go on Linux/macOS/Windows; PowerShell is an oracle only. Target the pinned 8.1.0 management surface and portable CI authentication. Use evidenced contracts, never guessed routes. Implement lab-unavailable features from evidence, with contract-tested and live-tested states separate. Keep credentials/tenant data out of artifacts and preserve context, token isolation and operation replay safety. No partial implementation counts as full parity. Source: the agreed product baseline in INDEX and docs/implementation/plan.md, rules 1–11.

## Planning

Milestone: M1 SDK foundations.
Priority: P0 — foundation, contract integrity or delivery gate.
Component / scope owner: sdk; HTTP workstream. Assigned commands and owned contracts define this task; other resource families remain with their owners.
Source rationale: agreed plan and pinned baseline at repository starting revision `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`; see INDEX for the requirement map and exclusions.
Suggested PR title: `feat(http): add redacted logging, tracing and transport configuration`

## Architecture decision alignment

Follow the agreed root-client/service/transport separation, constructor injection, typed models, Resty responses, Zap and optional OTel. Cloud/audience routing and token lifecycle belong to shared client/auth boundaries; backend wire contracts remain explicit. No pre-existing numbered ADRs exist at the repository starting revision; ARCH-01 records these agreed decisions, not new permission to change them.
