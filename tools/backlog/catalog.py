"""Reviewed work decomposition; text here is rendered into complete issue bodies."""

EPICS = {
    'BASE': ('Reproducible PowerShell compatibility baseline', 0, 'Every candidate has a sourced disposition, versioned metadata and a coverage owner; incomplete discovery cannot report success.'),
    'ARCH': ('Portable client and service architecture contracts', 0, 'A representative Graph service, Teams administration service and local builder follow one documented public contract.'),
    'FND': ('Go module and cross-platform development foundation', 1, 'The new module builds and tests on Linux, macOS and Windows with a root client and compiled examples.'),
    'AUTH': ('Portable delegated and automation authentication', 1, 'Each supported identity flow obtains the correct audience tokens with isolated caches, deterministic headless behaviour and documented permission restrictions.'),
    'HTTP': ('Reliable multi-backend REST transport', 1, 'Backend adapters share tested cancellation, throttling and observability while retaining backend-specific paging, errors and operation semantics.'),
    'LAB': ('Repeatable tenant conformance and fixture lifecycle', 2, 'Available tenant workflows compare SDK and pinned PowerShell results with sanitized evidence and verified cleanup.'),
    'CORE': ('Teams, channels, membership and templates', 3, 'Every assigned command is implemented and contract-tested; representative team and channel lifecycles are qualified against the available lab.'),
    'POL': ('Policies, configuration and assignment', 3, 'All assigned policy fields and assignment semantics are covered, including 8.1.0 additions, inheritance and batch outcomes.'),
    'VOICE': ('Voice users, numbering, routing and emergency administration', 3, 'Voice operations retain their typed contracts, authorization rules, order semantics and explicit live-test limitations.'),
    'CALL': ('Queues, attendants, resource accounts and voice builders', 3, 'Nested calling configurations round-trip through typed models and builders with operation-specific failure handling.'),
    'EXT': ('Applications, Shifts, devices and remaining management services', 3, 'All assigned extension families have implemented contracts; conditional/internal candidates have evidenced dispositions.'),
    'DELTA': ('Detect and review upstream compatibility changes', 4, 'Release and observed server changes produce reproducible impact reports and deduplicated draft PRs without silently adopting a baseline.'),
    'QUAL': ('Complete baseline coverage and first supported release', 5, 'All management operations are implemented and contract-tested, available lab cases pass and release documentation states evidence limitations.'),
}

MILESTONES = [
    ('M0 Baseline and contracts', 'Immutable 8.1.0 evidence, reconciled command ownership and agreed architecture. Family contract research precedes each implementation; no guessed wire routes.'),
    ('M1 SDK foundations', 'Portable Go client, authentication, backend adapters and cross-platform offline checks pass.'),
    ('M2 Representative proof', 'Team/channel/membership, policy lifecycle/assignment and voice read workflows have contract proof plus available lab evidence.'),
    ('M3 Complete management surface', 'Every in-scope resource and builder is implemented with tests, permissions, examples and coverage links.'),
    ('M4 Upgrade detection', 'Versioned and observed differences are classified and mapped to SDK impact; scheduled workflows maintain reviewable candidate draft PRs.'),
    ('M5 Release qualification', 'Full implementation and contract coverage, cross-platform gates, available lab regression and honest qualification documentation permit the first supported release.'),
]


def foundational_tasks(add):
    # ID, title, milestone, direct blockers, component, implementation, acceptance, verification.
    add('BASE-01', 'Build reproducible package acquisition and normalized baseline extraction', 0, [], 'tooling',
        'Promote the checked discovery seed into a repeatable tools/baseline workflow. Acquire exact Gallery package/metadata, verify SHA512, isolate imports, extract manifests/help/configuration/runtime signatures, and preserve provenance and schema version. Keep Microsoft binaries outside tracked artifacts. Preserve case-insensitive identities with original spellings and distinguish incomplete extraction from an empty surface.',
        'Repeated extraction from identical inputs produces the same normalized contract; checksum mismatch, import failure and incomplete metadata fail visibly; the 8.1.0 snapshot cannot be overwritten implicitly.',
        'Use the pinned archive and saved metadata plus checksum-corruption, truncated-download, import-error and empty-inventory fixtures. Compare two independent extractions byte-for-byte after removing explicitly volatile provenance fields.')
    add('BASE-02', 'Reconcile all exported and conditional command candidates', 0, ['BASE-01'], 'research',
        'Resolve the 730 candidates against actual exports, nested modules, platform/session conditions and documentation. The checked macOS observation contains 680 commands after case normalization. Investigate the other 50 individually. Retain every candidate in the ledger with evidence; only telemetry plumbing is already excluded. Record local builders, session equivalents, aliases and public management classifications. Add cross-platform observations when runners exist, without requiring Windows for SDK runtime.',
        'Every candidate has an evidenced disposition and ownership; no absence from one session is treated as proof of exclusion; additional public exports discovered later are added rather than hidden.',
        'Reconcile case-only names, commented preview declarations, conditional exports and internal telemetry examples; validate the union and explicit disposition totals against provenance. Publish the exact unresolved list if any contract research remains.')
    add('BASE-03', 'Maintain the capability and verification coverage ledger', 0, ['BASE-02'], 'tooling',
        'Define the ledger schema linking each command and parameter set to contract evidence, resource, Go symbol, ticket and tests. Track metadata-only, contract-tested and live-tested separately from implementation status. Include explicit common-parameter/session mappings and exclusions. Add a checker that fails on orphan commands, duplicate ownership, missing evidence and unsupported stubs counted as implementation.',
        'All baseline management commands have one implementation owner; evidence states cannot be conflated; a new unassigned command fails the check.',
        'Exercise removed owner, duplicate owner, stale symbol, excluded management command, skipped live test and unsupported-stub fixtures; require a clear machine-readable failure report.')
    add('ARCH-01', 'Record client, authentication and backend architecture contracts', 0, ['BASE-01'], 'documentation',
        'Turn the agreed architecture in docs/implementation/plan.md into numbered decisions and a contributor guide. Specify root client composition, constructor injection, Resty responses, functional options, Zap and optional OTel. Define resource-aware token acquisition, backend routing and error boundaries. Provide illustrative service signatures and a local builder example; distinguish Go API versioning from PowerShell versioning.',
        'The architecture explains Graph, Teams administration and local builders, with package dependencies free of import cycles and no runtime PowerShell/WAM requirement.',
        'Review the three representative call paths against the pinned Jamf reference and product rules; include a public-interface compatibility checklist.')
    add('FND-01', 'Initialize the Go module and cross-platform quality checks', 1, ['ARCH-01'], 'ci',
        'Replace template workload documentation with the planned SDK tree and module path. Pin a supported Go toolchain and compatible Resty/Zap/OTel/Microsoft identity dependencies. Configure build, vet, unit/race tests, lint, vulnerability checks and example compilation across Linux/macOS/Windows. Update template CI/CODEOWNERS links without merging unrelated dependency PRs.',
        'A minimal module and example compile on all platforms; ordinary CI needs no tenant secrets; lint configuration matches the selected tool version.',
        'Run go build ./..., go vet ./..., go test ./..., and platform matrix jobs. Verify race testing on supported runners and verify no test attempts authentication by default.')
    add('FND-02', 'Wire the root client, configuration and typed service composition', 1, ['FND-01'], 'sdk',
        'Add NewClient(authConfig, options...) and NewClientFromEnv with explicit authentication selection and functional options for HTTP settings. Wire typed services through constructor injection. Validate incompatible configuration locally, copy mutable configuration, keep tenant identity scoped per client, and expose session/cache cleanup without claiming Entra token revocation.',
        'Clients for different tenants cannot share mutable credentials or routes; construction errors are actionable; explicit configuration works without environment variables.',
        'Test environment parsing, explicit-over-environment behavior at the documented constructors, missing credentials, incompatible options, context lifecycle and independent concurrent clients.')
    add('AUTH-01', 'Implement resource-aware supplied tokens and refresh coordination', 1, ['FND-02'], 'sdk',
        'Define the auth provider request/result contract and adapters for audience-keyed supplied tokens and refresh callbacks. Key caches by tenant, cloud, identity and audience; serialize refresh and honor expiry/cancellation. Reject mixing delegated/application identity within a session. Do not forward a Graph token to a Teams endpoint. Static tokens report expiry instead of pretending to refresh.',
        'Audience isolation, expiry, concurrent refresh and callback errors are deterministic; secrets never enter errors or debug logs.',
        'Use fake clocks/token providers for near-expiry, simultaneous requests, cancelled refresh, wrong tenant/audience and missing token cases. Run the race detector.')
    add('AUTH-02', 'Add portable certificate and client-secret authentication', 1, ['AUTH-01'], 'sdk',
        'Use Microsoft authentication libraries for certificate/key material and client-secret client credentials. Support portable certificate loading and caller-provided key material; request each backend audience independently. Document app registration, consent and Teams role prerequisites per operation. No Windows store lookup or implicit interactive fallback.',
        'Both credential modes obtain audience-specific tokens on all platforms; invalid certificate/key pairs and expired credentials return redacted errors.',
        'Test PEM/PFX supported input contracts, mismatched key, expiry and token-service failures using isolated fixtures; run available commercial app-only read cases when credentials exist.')
    add('AUTH-03', 'Add portable delegated login and credential flows', 1, ['AUTH-01'], 'sdk',
        'Implement browser and device-code flows through Microsoft libraries, including caller callbacks and context cancellation. Support username/password only when Entra permits it; surface MFA/conditional-access requirements faithfully. Interactive modes are explicit and never selected by headless fallback. Document portable replacements for Windows-only connection switches.',
        'Browser/device flows have bounded cancellation and actionable errors; denied password flow does not downgrade authentication or retry indefinitely.',
        'Use mocked identity exchanges for authorization denial, MFA required, expired device code, timeout and user cancellation; qualify available delegated lab reads separately.')
    add('AUTH-04', 'Add managed identity and CI workload federation', 1, ['AUTH-01'], 'sdk',
        'Provide explicit managed-identity configuration using Azure Identity and federation using an assertion callback or file with Entra client assertion exchange. Support caller-selected identity where the environment exposes it. Refresh assertions/tokens at expiry, acquire backend audiences separately and never log assertion material. Examples show CI configuration without assuming a Windows runner.',
        'No identity endpoint produces a clear unavailable error; rotated assertion files and refresh callbacks work; audience and tenant boundaries are preserved.',
        'Mock identity endpoints and assertion exchanges for missing/expired assertions, wrong identity and concurrent refresh. Compile Linux/macOS/Windows examples; record live-unavailable identity environments honestly.')
    add('AUTH-05', 'Qualify permissions and authentication capability reporting', 1, ['AUTH-02','AUTH-03','AUTH-04','BASE-02'], 'verification',
        'Create the operation/authentication matrix using pinned evidence and current observations kept distinct. Cover module app-only exceptions, roles and consent requirements. Distinguish unauthenticated, forbidden, unsupported auth mode and unavailable tenant capability. Document supplied-token and federation additions without claiming the module directly exposes every provider.',
        'Each flow has portable documentation and each known command restriction has a contract test; authentication success alone is never reported as operation authorization.',
        'Exercise 401/403, missing role, delegated-only operation and same-type dual-token scenarios. Attach lab results where available and contract evidence elsewhere.')
    add('HTTP-01', 'Implement backend routing, request building and normalized errors', 1, ['AUTH-01'], 'sdk',
        'Implement the shared client interface with Resty and cloud/audience-aware backend adapters. Build escaped paths/queries, preserve backend headers and expose raw response metadata plus typed errors with status, service code and request IDs. Preserve context and body cleanup. Resolve only allowlisted endpoint/discovery/continuation destinations and strip secrets from errors.',
        'Graph and Teams envelopes decode without losing diagnostic identifiers; invalid discovery destinations cannot receive credentials; cancellation reaches the request.',
        'Use httptest backends for JSON/non-JSON errors, empty bodies, malformed content, escaped identities, cross-host redirects, wrong token audience and cancellation.')
    add('HTTP-02', 'Implement bounded retries, throttling and concurrency controls', 1, ['HTTP-01'], 'sdk',
        'Centralize request budgets, Retry-After handling, bounded backoff/jitter and per-client concurrency. Classify replay safety by operation rather than HTTP method alone. Refresh on eligible auth failures once without creating retry loops. Release concurrency slots on cancellation and prevent replay of non-rewindable uploads.',
        '429 and transient failures honor total budgets; unsafe mutations are not duplicated; waiting and in-flight cancellation release resources.',
        'Use deterministic clock/random injection for Retry-After date/seconds, invalid header, budget exhaustion, 401 loops, concurrent callers and ambiguous mutation failures; run race tests.')
    add('HTTP-03', 'Implement backend-specific pagination and list traversal', 1, ['HTTP-01'], 'sdk',
        'Provide page-wise and aggregate list helpers for evidenced continuation-link, continuation-token and offset patterns. Preserve original query semantics and trusted routing. Detect repeated continuations and bound traversal by context; expose partial failures rather than returning truncated success.',
        'Every observed pagination style handles empty/intermediate/final pages and cancellation; continuation links cannot change credentials or authority.',
        'Test multi-page fixtures, duplicate continuation loops, errors after a successful page, query encoding and cross-host continuation rejection.')
    add('HTTP-04', 'Implement long-running operations and eventual-consistency waits', 1, ['HTTP-02'], 'sdk',
        'Expose operation handles/status and bounded wait helpers for backend-specific async envelopes. Honor Retry-After and terminal success/failure states. Separate accepted requests from completed mutations; tolerate only evidenced transient propagation errors. Preserve remote operation identifiers after local cancellation.',
        '202 is not treated as completion; wait timeout returns recoverable operation information; terminal failures retain backend diagnostics.',
        'Test accepted/running/succeeded/failed/cancelled sequences, missing operation location, stalled jobs and eventual consistency boundaries.')
    add('HTTP-05', 'Implement binary, multipart and batch request behaviours', 1, ['HTTP-02'], 'sdk',
        'Add backend-specific binary upload/download, multipart and batch helpers needed by assigned contracts. Preserve content types, streaming ownership, size/error reporting and item-level results. Do not retry an entire partially successful batch or a consumed stream automatically.',
        'Binary bytes survive round-trip; partial batch outcomes remain available; failed streams are closed and unsafe replay is prevented.',
        'Test byte-for-byte fixtures, multipart boundaries, truncated bodies, upload errors and mixed-success batches with deterministic mock responses.')
    add('HTTP-06', 'Add redacted logging, tracing and transport configuration', 1, ['HTTP-01'], 'sdk',
        'Expose Jamf-style options for timeout/proxy/TLS/custom transport/headers/user agent/concurrency and Zap logging with optional OTel hooks. Redact tokens, assertions, passwords, certificate keys and sensitive payload fields in every logging mode. Tracing is opt-in and uses the caller lifecycle; secret headers cannot be overridden accidentally by generic options.',
        'Debug logs and spans remain sanitized; custom transport and proxy configuration work without bypassing auth/cancellation controls.',
        'Capture logs/spans with representative secret fixtures, TLS/proxy tests and custom transport tests; assert absence of raw secrets in success and error paths.')
    add('LAB-01', 'Build tenant fixture and acceptance-test lifecycle', 2, ['FND-02','HTTP-01'], 'integration',
        'Add opt-in acceptance configuration, tenant preflight, capability/permission inventory, unique test fixture names and cleanup journals. Separate read-only smoke tests from mutations and consequential scenarios. Track created IDs and restore snapshots of shared settings. No credentials or raw tenant dumps in committed fixtures or reports.',
        'Ordinary go test skips live work explicitly; interrupted acceptance runs leave recoverable cleanup records; available/unavailable scenarios are distinguishable.',
        'Test configuration and cleanup using fake backends, including failure midway through creation and restoration; run read-only lab preflight once credentials are provided.')
    add('LAB-02', 'Compare SDK and pinned PowerShell behaviour with sanitized evidence', 2, ['LAB-01','BASE-03'], 'integration',
        'Implement a conformance harness that runs pinned 8.1.0 and SDK operations on equivalent isolated resources, normalizes only documented volatile fields and compares outcomes/read-back. Capture package hash, auth mode, scenario, sanitized requests/responses and cleanup result. Store reviewed fixtures with provenance; do not equate a skipped live case to a pass.',
        'Deliberate field/permission/serialization differences are detected; normalized comparisons preserve meaningful ordering and omission/null distinctions.',
        'Use controlled equal/different fixtures and failed-cleanup scenarios, then execute available tenant cases; review captured output for credentials and identifiers before retention.')
    add('LAB-03', 'Prove the representative cross-backend workflows', 2, ['LAB-02'], 'verification',
        'Qualify team/channel/membership lifecycle, messaging-policy lifecycle and assignment, and an online voice-user read after their implementation prerequisites. Verify app-only and delegated restrictions and read-after-write behaviour. Missing live capability is recorded with contract evidence; unexplained differences remain defects.',
        'All selected workflows have passing contract comparisons and available live evidence, with cleanup verified independently of operation success.',
        'Run create/read/update/delete and membership/assignment scenarios against dedicated fixtures; attach SDK and PowerShell baseline identifiers, differences and cleanup records.')
    add('DELTA-01', 'Implement semantic baseline comparison and SDK impact reports', 4, ['BASE-03'], 'tooling',
        'Build tools/contractdiff over versioned inventories. Detect command/alias additions/removals, parameter-set/type/requiredness/default/validation/output differences and evidenced routing/auth changes. Classify documentation-only versus metadata versus observed behaviour changes. Link impacts to services, models, tests and stable issue IDs. Unknown fields and incomplete extraction are visible.',
        'Identical snapshots produce no delta; each meaningful change has old/new evidence and an affected owner; incomplete snapshots cannot yield a clean result.',
        'Use synthetic add/remove/rename/type/default/null fixtures and verified historical 8.1.0 changes, including CallQueue Statistics removal and new phone usage-order/parameter additions. Test stable versus preview versions.')
    add('DELTA-02', 'Schedule upstream discovery and maintain candidate draft PRs', 4, ['DELTA-01'], 'ci',
        'Add daily/manual release discovery, verified artifact acquisition and snapshot comparison. Keep stable/preview channels separate. Create or update one draft PR keyed by target version plus source checksum; include JSON/Markdown impact reports. Use least-required workflow permissions, safe fork behaviour and concurrency control. Never auto-merge or advance the accepted baseline.',
        'Repeated runs do not duplicate PRs; no-change runs do not write; download/extraction failure is visible and preserves the accepted 8.1.0 snapshot.',
        'Exercise no-change/new-release/checksum-change/prerelease/failure/retry scenarios with mocked Gallery/GitHub interfaces; verify draft metadata and idempotent update behaviour.')
    add('DELTA-03', 'Detect sampled server drift independently of module releases', 4, ['DELTA-01','LAB-02'], 'integration',
        'Add optional scheduled read-only probes against configured lab capabilities and compare sanitized response/behaviour observations with accepted contracts. Keep server drift reports separate from module version differences and map affected resources. Missing credentials or unsupported capabilities report skipped/unavailable without modifying the baseline.',
        'A changed sampled response shape is detected even with unchanged module version; probes never mutate tenant configuration; sampling limitations are explicit.',
        'Test added/removed/type-changed fields, transient service errors, unavailable probes and unchanged observations; verify report redaction and absence of mutation calls.')
    add('QUAL-01', 'Audit complete implementation and contract coverage', 5, ['BASE-03'], 'verification',
        'Audit every in-scope command, parameter set, builder and session equivalent against compiled Go symbols and executable contract tests. Reject unsupported stubs and missing models/permissions/examples. Reconcile all resource-family research outputs, conditional candidates and explicit exclusions. Publish a requirement-to-ticket-to-test coverage report.',
        'Every management capability has an implementation and contract-test evidence; excluded telemetry and portable session mappings are explicit; no unowned commands remain.',
        'Run the coverage checker and negative mutation fixtures, inspect representative per-family evidence and compare the audited command set to the immutable baseline.')
    add('QUAL-02', 'Qualify concurrency, security and available lab regressions', 5, ['QUAL-01','LAB-03','AUTH-05','HTTP-06'], 'verification',
        'Run full build/vet/unit/race/lint/vulnerability/example checks and available tenant scenarios across supported OS runners. Verify token isolation, redaction, credential rotation, unsafe replay protection and cleanup. Publish contract-tested versus live-tested matrices; unavailable licences/runners do not erase implemented scope.',
        'All automated gates pass; available acceptance cases pass with cleanup evidence; unavailable cases retain precise evidence limitations.',
        'Execute the planned Go checks, cross-platform jobs and opt-in lab suite; attach exact commit/toolchain/package hashes and distinguish failures from skipped capabilities.')
    add('QUAL-03', 'Publish SDK guides, CI recipes and compatibility documentation', 5, ['QUAL-01','AUTH-05','DELTA-02'], 'documentation',
        'Replace template README/docs with quick start, authentication/permission guides, cloud support, error/pagination/async usage, API-family examples and Linux CI federation/certificate recipes. Include PowerShell-to-Go mappings, version support, upgrade review, unsupported plumbing exclusions and live-evidence limitations. Explain that SDK use requires no PowerShell installation.',
        'A new consumer can authenticate and call both backend families from documented examples; every in-scope service has a discoverable example and permission reference.',
        'Compile all examples; check source links and command mappings; run the portable quick-start smoke case where lab access permits.')
    add('QUAL-04', 'Gate and prepare the first supported SDK release', 5, ['QUAL-02','QUAL-03','DELTA-03'], 'ci',
        'Add release automation, changelog/version rules and reproducible release checks. Require complete implementation/contract coverage and the published evidence matrix before a supported tag. Keep PowerShell baseline version separate from SDK semver. Document upgrade and rollback procedures; publishing the backlog does not authorize an SDK release tag.',
        'Release dry-run fails for missing capability evidence or failed gates and succeeds for a fully qualified commit; no partial build is labelled full parity.',
        'Exercise release dry-run with passing/failing gate fixtures and verify version/changelog metadata and baseline provenance without publishing a release.')
