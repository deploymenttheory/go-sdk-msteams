# Contributing

Use the [implementation backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6) to select work whose direct prerequisites are satisfied. Read the ticket, its source contracts and the [implementation baseline](docs/implementation/plan.md) before changing an interface.

## Architecture and services

Follow the [service development guide](docs/guides/service-development.md) and numbered [client](docs/decisions/0001-client-and-service-composition.md), [authentication](docs/decisions/0002-portable-authentication.md) and [backend](docs/decisions/0003-backend-contracts-and-compatibility.md) decisions. Match the agreed Jamf-style structure and public conventions. Research unknown Teams routes and payloads before implementing them.

## Evidence and verification

Reproduce metadata with the [baseline tool](tools/baseline/README.md). Run its offline tests and backlog validator for discovery changes. Go build, lint, unit/race and example checks are introduced by FND-01; service work must use those gates once available. Keep contract-tested and live-tested evidence separate, and never commit tenant credentials or raw tenant captures.

Pull requests should name the ticket and concrete behaviour, include verification results, and explain compatibility impacts. Wait for green CI before merging. Fix a failing existing PR before accumulating more unfinished work. Do not close an epic solely because its children are closed; verify its integrated acceptance conditions.

Report bugs through [GitHub issues](https://github.com/deploymenttheory/go-sdk-msteams/issues). Follow the [code of conduct](CODE_OF_CONDUCT.md).
