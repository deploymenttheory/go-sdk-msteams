<!-- weave-backlog:msteams-powershell-parity:v1:INDEX -->
Backlog ID: `INDEX` · Source: agreed implementation plan; repository start `60aaa442defd7b3d5a2f7a11faac62c23a5bcfbb`

## Purpose and scope

Native Go SDK with complete MicrosoftTeams 8.1.0 management parity. This backlog implements the agreed plan; SDK work is not yet delivered.

## Product baseline

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


## Glossary

- Candidate: command found in a source, pending reconciled public/export classification.
- Contract-tested: implemented behaviour verified offline against evidence-derived contracts.
- Live-tested: an implemented scenario observed against a tenant.
- Baseline: immutable accepted upstream version; a candidate release PR does not change it.
- Direct blocker: an unavailable contract required to execute a task, independent of parent membership.

## Agreed architecture decisions

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


## Delivery map

| Milestone | Exit | Epics |
| --- | --- | --- |
| M0 Baseline and contracts | Immutable 8.1.0 evidence, reconciled command ownership and agreed architecture. Family contract research precedes each implementation; no guessed wire routes. | [EPIC-BASE](EPIC-BASE.md), [EPIC-ARCH](EPIC-ARCH.md) |
| M1 SDK foundations | Portable Go client, authentication, backend adapters and cross-platform offline checks pass. | [EPIC-FND](EPIC-FND.md), [EPIC-AUTH](EPIC-AUTH.md), [EPIC-HTTP](EPIC-HTTP.md) |
| M2 Representative proof | Team/channel/membership, policy lifecycle/assignment and voice read workflows have contract proof plus available lab evidence. | [EPIC-LAB](EPIC-LAB.md) |
| M3 Complete management surface | Every in-scope resource and builder is implemented with tests, permissions, examples and coverage links. | [EPIC-CORE](EPIC-CORE.md), [EPIC-POL](EPIC-POL.md), [EPIC-VOICE](EPIC-VOICE.md), [EPIC-CALL](EPIC-CALL.md), [EPIC-EXT](EPIC-EXT.md) |
| M4 Upgrade detection | Versioned and observed differences are classified and mapped to SDK impact; scheduled workflows maintain reviewable candidate draft PRs. | [EPIC-DELTA](EPIC-DELTA.md) |
| M5 Release qualification | Full implementation and contract coverage, cross-platform gates, available lab regression and honest qualification documentation permit the first supported release. | [EPIC-QUAL](EPIC-QUAL.md) |

## Requirement coverage

| Requirement | Owners |
| --- | --- |
| Pinned baseline and complete management coverage | [EPIC-BASE](EPIC-BASE.md), [QUAL-01](QUAL-01.md) |
| Jamf-style native Go client and repository structure | [EPIC-ARCH](EPIC-ARCH.md), [EPIC-FND](EPIC-FND.md) |
| Portable authentication and CI federation | [EPIC-AUTH](EPIC-AUTH.md) |
| REST behaviours, resilience and observability | [EPIC-HTTP](EPIC-HTTP.md) |
| Teams, channels and membership | [EPIC-CORE](EPIC-CORE.md) |
| Policies, configuration and assignment | [EPIC-POL](EPIC-POL.md) |
| Voice, numbering, routing and emergency administration | [EPIC-VOICE](EPIC-VOICE.md) |
| Queues, attendants and builders | [EPIC-CALL](EPIC-CALL.md) |
| Applications, Shifts and remaining services | [EPIC-EXT](EPIC-EXT.md) |
| Lab qualification and honest unavailable-feature evidence | [EPIC-LAB](EPIC-LAB.md), [QUAL-02](QUAL-02.md) |
| 8.2.0/new release and server change detection | [EPIC-DELTA](EPIC-DELTA.md) |
| Complete first supported release and consumer documentation | [EPIC-QUAL](EPIC-QUAL.md) |

All 730 candidate commands are enumerated in their research/implementation task bodies. 1 telemetry command is explicitly excluded; 9 helpers map to AUTH-06. The initial 680 exports are unauthenticated macOS evidence, not the final cross-platform surface. BASE-02 reconciles the other candidates without dropping them. Package SHA256: `78c5f73347a97b2e4810222e702f79b0fe6d5c8ec4f3b52b4b2891c70b445ab7`.

## Execution conventions

Type, priority, area and component labels are orthogonal. P0 denotes foundation/integrity/release gates; P1 denotes required release capabilities. Native task parents and direct blockers are required; supporting links do not add blocker edges. Milestones have no invented due dates. Research settles unknown wire contracts before dependent implementation. All tasks are initially open; metadata discovery is not SDK delivery. Existing dependency-update PRs are unrelated.

Source links: https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0; https://github.com/deploymenttheory/go-sdk-jamfpro-v2/tree/25714da557956da9e317f725ca3d22e8818ec77e. Baseline snapshots and the local plan must accompany implementation handoff.
