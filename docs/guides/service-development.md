# Developing a service

Start from a concrete SDK operation or defect in the GitHub backlog. Read the
existing service, API documentation, JSON fixtures and acceptance scenarios.
Use documented or observed endpoints and payloads; do not invent routes or
constraints to fill out an interface.

## Implementation

Place resources under `msteams/graph_api/` or `msteams/teams_admin_api/` and wire
services through constructor injection. Services use the shared client boundary;
authentication and backend transport remain shared concerns.

Take `context.Context` first for remote operations. Return typed values, actual
Resty response metadata and errors. Keep read and write models separate, preserve
omission/null/false/empty distinctions where meaningful, and retain partial
operation results needed for recovery. Use existing CRUD/action naming and local
validation conventions. Pure constructors require no credentials or HTTP.

Keep cancellation, bounded paging/polling, trusted destinations and mutation
replay safety intact. Do not invent rollback or retry a mutation merely because
an HTTP response looks transient.

## Verification

- Extend ordinary unit tests using service `testdata/` JSON fixtures.
- Assert request construction, decoded fields, error receipts and relevant safety cases.
- Preserve successful/error response regressions; label synthetic cases clearly.
- Extend opt-in tenant acceptance for available real workflows, with read-back,
  independently verified cleanup and shared-state restoration.
- Run applicable build, vet, lint, unit/race, vulnerability and example checks.
- Update API docs and examples, including concrete permissions or licensing limits.

No separate contract research, command/parameter-set coverage registration or
CLR emulation is required. Use the architecture decisions in
[docs/decisions](../decisions) for client, authentication and backend boundaries.
