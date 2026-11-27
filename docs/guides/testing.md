# Testing the SDK

The project uses ordinary Go unit tests, checked-in JSON response fixtures and
opt-in acceptance tests against a real tenant.

## Unit tests and JSON fixtures

Unit tests use `TestUnit`; compiled examples use Go example tests. Run them
without tenant credentials:

```sh
go test -count=1 -shuffle=on -run '^(TestUnit|Example)' ./...
go test -race -count=1 -shuffle=on -run '^(TestUnit|Example)' ./...
```

Keep captured, sanitized response JSON under the service's `testdata/` directory.
Verify typed values, original response metadata and error behaviour. Request
assertions should cover methods, paths, headers and JSON bodies where relevant.
Synthetic responses are useful for malformed payloads, cancellation, retry and
partial-failure cases; keep their origin clear. Local builders should not acquire
credentials or send HTTP requests.

Use the same Go test tools for these cases. No command inventory, parameter-set
mapping, provenance report or contract ledger is needed to register a test.

## Real-tenant acceptance

Acceptance tests use `TestAcc`. Configure the lab explicitly, select dedicated
fixtures and enable mutations only for the intended scenarios. Check operation
results through read-back and independently verify cleanup or restoration.
Retain recovery information when the outcome is uncertain. Never turn a skipped
or forbidden operation into a successful result.

Existing PowerShell comparisons remain part of selected tenant scenarios. The
comparison module is downloaded and checksum-verified by
`tools/acceptance/powershell.py`; it is not an SDK runtime dependency or research
inventory generator. See the [acceptance setup](../../tools/acceptance/README.md).

## Quality and release checks

Keep build, vet, lint, unit/race, vulnerability and compiled example checks.
Review functional support and actual tenant results before release. Record
missing permissions, fixtures and licences as limitations. Voice administration
is deferred until a licensed testing environment is available.
