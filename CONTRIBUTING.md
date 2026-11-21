# Contributing

Use the [implementation backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6) to select work whose direct prerequisites are satisfied. Read the ticket, its source contracts and the [implementation baseline](docs/implementation/plan.md) before changing an interface.

Keep SDK Go source and package tests under `msteams/`, with examples under `examples/`.
The repository root contains project metadata and configuration, not Go packages.

## Architecture and services

Follow the [service development guide](docs/guides/service-development.md) and numbered [client](docs/decisions/0001-client-and-service-composition.md), [authentication](docs/decisions/0002-portable-authentication.md) and [backend](docs/decisions/0003-backend-contracts-and-compatibility.md) decisions. Match the agreed Jamf-style structure and public conventions. Research unknown Teams routes and payloads before implementing them.

## Tests and verification

Reproduce metadata with the [baseline tool](tools/baseline/README.md). Run its offline tests and backlog validator for discovery changes. Run the Go build, vet, lint, unit/race and example checks before opening a PR. Unit tests start `TestUnit`; tenant acceptance tests start `TestAcc`. Successful response fixtures belong in each service's `testdata/` directory and must come from its acceptance tests. Sanitize tenant identifiers and credentials before committing fixtures. See the [acceptance test guide](tools/acceptance/README.md) for CI and fixture refresh instructions.

Pull requests should name the ticket and concrete behaviour, include verification results, and explain compatibility impacts. Wait for green CI before merging. Fix a failing existing PR before accumulating more unfinished work. Do not close an epic solely because its children are closed; verify its integrated acceptance conditions.

Report bugs through [GitHub issues](https://github.com/deploymenttheory/go-sdk-msteams/issues). Follow the [code of conduct](CODE_OF_CONDUCT.md).
