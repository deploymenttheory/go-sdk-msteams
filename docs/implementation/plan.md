# Microsoft Teams SDK implementation baseline

Status: agreed product and architecture plan; SDK implementation is not started. The accompanying backlog is a publication preview, not published GitHub issues.

## Source and destination

- Target: `deploymenttheory/go-sdk-msteams`; Go module `github.com/deploymenttheory/go-sdk-msteams`.
- Repository starting revision: `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`.
- PowerShell baseline: [MicrosoftTeams 8.1.0](https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0), published 20 November 2026. Package checksums and observed environment are in [provenance](../../contracts/powershell/8.1.0/provenance.json).
- Architecture reference: [Jamf Pro SDK](https://github.com/deploymenttheory/go-sdk-jamfpro-v2/tree/25714da557956da9e317f725ca3d22e8818ec77e), pinned at `25714da557956da9e317f725ca3d22e8818ec77e`.
- Ticket quality references: [feature 1531](https://github.com/weaveplatform/artifactweave/issues/1531) and [epic 1743](https://github.com/weaveplatform/artifactweave/issues/1743).
- Backlog namespace: `msteams-powershell-parity`; stable IDs are independent of eventual GitHub issue numbers.

## Product rules

1. **Native runtime:** the SDK works on Linux, macOS and Windows without PowerShell, subprocess wrappers, WAM or Windows certificate stores. PowerShell is a discovery and conformance oracle only.
2. **Management parity:** implement the public management behaviour, parameter sets, models and local object builders in 8.1.0. Explicitly map session/environment helpers to SDK configuration and lifecycle. PowerShell formatting, pipeline mechanics, confirmation UI and module telemetry are not SDK runtime features; retain a documented mapping for common parameters.
3. **Complete first release:** the first supported release requires complete implementation and contract coverage of the reconciled management surface. No partial release is presented as parity.
4. **Evidence before wire code:** use the pinned package, Microsoft documentation and sanitized observations. Do not invent endpoints or treat empty output metadata as proof of no response. Contract research produces a settled contract before its implementation ticket is ready.
5. **Lab limitations:** implement features unavailable in the lab from evidenced PowerShell behaviour and offline contracts. Distinguish metadata-only, contract-tested and live-tested evidence. Missing live fixtures do not remove features from scope; unknown behaviour remains a contract blocker.
6. **Portable authentication:** provide explicit browser, device code, username/password where permitted, certificate, client-secret, supplied-token, managed-identity and workload-federation flows. Headless use never silently prompts or changes identities. Certificate/key material is portable; federation accepts an assertion callback or file. Preserve service-specific application/delegated restrictions.
7. **Clouds:** commercial qualification first. Separate cloud authorities, audiences, endpoints and regional discovery; do not claim sovereign-cloud qualification from commercial tests.
8. **Version separation:** SDK semantic versioning is independent of PowerShell versioning. Keep 8.1.0 immutable; future versions are candidate baselines until reviewed.
9. **Transport integrity:** bound retries and polling, preserve cancellation, respect throttling and replay safety, isolate token audiences and tenants, and redact credentials and tenant data. Follow only trusted continuation/discovery/operation destinations.
10. **Lab lifecycle:** credentials arrive later through environment/secret configuration, never committed artifacts. Test-owned resources have unique identifiers and cleanup journals. Capture and restore shared settings; consequential tenant-wide and telephone-number workflows are explicitly enabled scenarios.
11. **Delivery evidence:** each resource supplies its contract, implementation, tests, examples, API documentation and ledger links. Documentation and validation belong with the change.
12. **Publication:** approve the complete issue/metadata/relationship preview before GitHub writes; native parent and direct-blocker relationships are required. Do not close epics automatically when children close.

## Architecture decisions

These are agreed decisions, not pre-existing numbered ADRs.

- **Client composition:** root `msteams.Client`, typed resource services, constructor-injected transport, functional options, explicit `NewClient(authConfig, options...)` and `NewClientFromEnv()`.
- **Jamf conventions:** Resty transport/response objects, Zap logging, optional OpenTelemetry, colocated models, service mocks, examples and separate acceptance packages. Services return typed results, `*resty.Response` and errors; all remote operations take `context.Context`.
- **API families:** `graph_api/<resource>` and `teams_admin_api/<resource>` packages. Internal adapters handle different Teams backend protocols. Contract research establishes each operation's routing; a package name never establishes an endpoint by itself.
- **Authentication boundary:** resource-aware provider requests include tenant, cloud, identity and audience. Use Microsoft MSAL Go/Azure Identity implementations behind this seam. Cache expiry and refresh are concurrency-safe; static supplied tokens cannot be refreshed without a callback. Disconnect clears local session/cache state, without claiming to revoke an Entra token remotely.
- **Models:** separate write/read types; distinguish omission, null, false, zero and empty collections where the upstream contract does. Local constructors become typed Go builders. Version method names only when actual API contracts require it.
- **REST adapters:** backend-specific error envelopes and pagination feed shared transport controls. Async operations expose status/wait operations rather than unbounded synchronous waits. Retry mutations only with evidenced replay safety; multipart, binary and batch semantics are implemented where observed.
- **Dependencies:** follow the reference design, not all of its dependencies. Use a supported Go toolchain and pinned compatible library versions established by the foundation task. Manual constructor wiring; no global service registry or DI container.

```text
msteams/
  msteams.go, with_options.go
  client/, auth/, config/, constants/, shared/, mocks/
  graph_api/<resource>/
  teams_admin_api/<resource>/
  acceptance/{graph_api,teams_admin_api}/
examples/{authentication,graph_api,teams_admin_api}/
recipes/
contracts/{powershell/8.1.0,endpoints,coverage}/
docs/{guides,research,decisions,implementation}/
tools/{baseline,contractdiff,coverage}/
scripts/{discovery,pipeline}/
.github/workflows/
```

## Delivery sequence

| Gate | Required outcome |
| --- | --- |
| M0 Baseline and contracts | Pin sources; reconcile exports; establish architecture and command ownership. Research tickets settle each family's wire contracts before its implementation can start. |
| M1 SDK foundations | Build portable client, authentication, routing, errors, retries, paging, operations, transfer support and observability with offline verification. |
| M2 Representative proof | Build the lab harness and prove team/channel/membership, one policy lifecycle/assignment and a voice administration read across backend families. |
| M3 Complete management surface | Implement every assigned resource family and builder with contract tests, documentation and examples. |
| M4 Upgrade detection | Compare normalized versions and observations; produce impact reports and one deduplicated draft PR per candidate version/checksum. |
| M5 Release qualification | Audit complete coverage, cross-platform checks, available tenant scenarios and evidence limitations; prepare the first supported release. |

The dependency graph controls readiness, not milestone number or ticket number. DELTA begins after the baseline schema is stable and can run alongside SDK development. Research tickets can proceed without lab credentials using package/documentation evidence; unknown behaviour cannot be silently converted into an implementation assumption.

## Baseline and change detection

The checked discovery contains 730 case-insensitive candidate commands and 680 observed runtime exports from an unauthenticated macOS import. Fifty Gallery candidates were not observed. The manifest scan is lexical and supplementary. BASE-02 must resolve conditional/platform/session exports and public versus internal classification. Preserve candidates until evidence resolves them.

Inventory records commands, aliases, parameter sets/types/requiredness/validation, defaults where evidenced, outputs, local constructors, endpoint families and auth constraints. Runtime reflection alone does not establish default values, server schemas or wire protocols.

Keep an immutable 8.1.0 snapshot with source checksums. Daily and manual workflows discover releases, verify downloads, normalize inventories and compare additions/removals, changed parameters/validation/defaults/outputs, routing/auth changes and available observations. Report extraction failures as failures, not empty deltas. Stable and preview releases have separate channels.

JSON/Markdown reports link changes to services, types, tests and issue identities. Automation creates or updates a candidate draft PR; it never auto-merges or promotes the baseline. Optional read-only tenant probes detect sampled server drift independently of package releases. No claim is made that static package diffs detect all server changes.

## Verification and release acceptance

The foundation will add `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./...`, lint, vulnerability checks and example compilation. These are planned checks, not currently passing SDK tests.

Offline cases cover exact serialization, omission/null distinctions, authorization errors, token expiry/concurrent refresh, paging loops, throttling budgets, replay safety, async terminal failure, partial batch failure, binary transfer and redaction. Live comparisons use equivalent isolated fixtures and pinned PowerShell 8.1.0. Broad destructive or chargeable scenarios are not default smoke tests.

The supported release requires every in-scope command to map to a Go operation/builder/session equivalent, passing contract tests, documented permissions and verification evidence. Live-unavailable cases retain a visible limitation. Exclusions require evidence and baseline review; a stub or unsupported error is not management parity.

## Backlog workflow

The [preview](backlog/README.md) links complete issue bodies, labels, milestones, hierarchy, direct dependencies, coverage and the ready frontier. Resource families have separate contract-research prerequisites where wire details are unknown. Each task uses all eight required sections and carries its command set inline.

Publication resolves stable IDs to GitHub URLs, creates native sub-issues and blockers, then reads back all bodies and relationships. Preserve the publication mapping and journal for recovery. Creating this backlog does not implement its SDK feature tickets.
