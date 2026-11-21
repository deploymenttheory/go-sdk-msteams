# CORE-15 offline fixture harness

Work date: 2026-11-20. Ticket: [CORE-15 / #61](https://github.com/deploymenttheory/go-sdk-msteams/issues/61).

The shared [`internal/contracttest`](../../../internal/contracttest/README.md)
runner executes resource-owned callbacks through the actual root client, its
backend scopes and a rejecting in-memory transport. Synthetic credentials and
explicit client configuration keep all checks offline. No resource CRUD is
implemented by this ticket and no tenant success is claimed.

The implementation checks exact escaped destination, ordered query values,
selected headers, JSON presence/null/false/empty values, unknown fields, ordered
arrays, exact integer/number spelling, status, raw response metadata,
classification and normalized diagnostics. Duplicate JSON keys are rejected.
Missing/extra attempts and wrong bearer audience fail, including an audience
that is valid for a different exchange. Local-only constructors are checked for
zero HTTP and zero token acquisition. Pre-cancellation propagates without I/O.

Self-tests register cases for all nine CORE families. These are infrastructure
registration checks against synthetic requests, not implementation evidence for
those nine families. Deliberate route, method, audience, query-order, header,
omission/body, response-field/status and error mismatches produce structural
diffs. Unknown keys and value/token/environment canaries do not enter evidence.
Suite/evidence ordering is stable, and unexecuted or duplicate results cannot
produce a valid passing artifact.

## Representative settled contract

[`TestUnitPinnedTemplatePrimitiveFixtures`](../../../tests/contract/core/fixtures_test.go)
consumes five actual 8.1.0
[`generated-request-dispatch.json`](../../../contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json)
observations. The generated request URLs/bodies and response dispatch were
previously captured from the verified pinned module; this test adds real SDK
scope execution and compares them through the reusable runner.

| Captured case | Mapped command/set | Exercised boundary |
| --- | --- | --- |
| template-read-opaque-id | Get-CsTeamTemplate / Get | GET with escaped opaque OData identity |
| template-list | Get-CsTeamTemplateList / DefaultLocaleOverride | GET collection envelope |
| template-create | New-CsTeamTemplate / NewExpanded | POST sparse JSON and 201 response |
| template-update | Update-CsTeamTemplate / UpdateExpanded | PUT sparse JSON and 200 response |
| template-delete-empty | Remove-CsTeamTemplate / Delete | DELETE and genuinely empty 204 |

The mapping is checked against `contracts/powershell/8.1.0/runtime.json` during
the test. These are `scope: "primitive"` results; the callback directly invokes
the ConfigAPI scope. It does not claim template resource methods, default-locale
composition, forest discovery, wrapper warning/flattening behavior or all
parameter sets. Those remain CORE-27 implementation/qualification work.

## Artifact lifecycle

`WriteEvidence` creates a candidate JSON run report with the supplied full Git
revision. CI uses the checked-out revision during the existing race step and
uploads only the sanitized report on Linux, macOS and Windows. It does not
publish response payloads or general test directories. The immutable artifact
action is pinned to the official
[v7.0.2 release](https://github.com/actions/upload-artifact/releases/tag/v7.0.2).

Reports retain `coverage_claim: false` and `live: "not-run"`. No ledger command
or parameter-set state is promoted. A service owner must connect exact SDK
symbols, complete supported sets and reviewed run evidence before claiming
contract success. Lab acceptance remains separate and required for available
resources; the pending Teams role approval does not become a tenant pass.

Local scoped race checks passed. Full test/lint/build/vet and three-platform CI
results are recorded with this implementation PR; a passing fixture runner does
not close the remaining CORE-01 research or resource tickets.
