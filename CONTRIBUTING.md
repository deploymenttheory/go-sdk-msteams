# Contributing

Select concrete functionality or test work from the [GitHub backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6).
Read the affected service implementation, API docs and existing JSON-backed tests.

Keep SDK source and package tests under `msteams/`, examples under `examples/`,
and real-tenant tests under `tests/integration/lab/`. The repository root holds
metadata and configuration, not Go packages.

## Architecture and services

Follow the [service development guide](docs/guides/service-development.md) and the
client, authentication and backend decisions in [docs/decisions](docs/decisions).
Use documented or observed API routes and payloads. Preserve typed models,
context propagation, response metadata, error classification and safe retries.

## Tests and verification

Unit tests start `TestUnit` and examples use Go example tests. Retain sanitized
JSON response fixtures in the affected service's `testdata/` directory. Test
request construction, decoded values, errors and relevant safety behaviour.
New synthetic edge cases must be labelled clearly rather than presented as live
captures. See the [testing guide](docs/guides/testing.md).

Tenant tests start `TestAcc` and use explicit credentials and dedicated fixtures.
Verify read-back, independent cleanup and shared-state restoration. Keep skips,
missing permissions and licensing restrictions explicit. Never commit credentials
or private tenant data. See the [acceptance guide](tools/acceptance/README.md).

Run build, vet, lint, unit/race, vulnerability and example checks appropriate to
the change. There is no separate contract-research or coverage-ledger gate.

Pull requests should name the ticket and concrete behaviour, include verification
results and explain compatibility impacts. Wait for green CI before merging.
Fix failing existing work before accumulating unfinished PRs. Assess an epic's
actual delivery criteria independently of its children's closure.

Report bugs through [GitHub issues](https://github.com/deploymenttheory/go-sdk-msteams/issues).
Follow the [code of conduct](CODE_OF_CONDUCT.md).
