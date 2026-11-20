# Go development foundation

The module targets Go **1.27.4**, pinned exactly in `go.mod` and consumed by setup-go.
This version was verified against the [official Go download metadata](https://go.dev/dl/?mode=json).
CI uses that installed toolchain with `GOTOOLCHAIN=local` so an unexpected dependency
cannot silently switch the compiler. Go dependency resolution still checks the public
module checksum database, and `go.sum` is checked in.

## Dependencies

| Dependency | Pin | Purpose |
|---|---|---|
| Resty v2 | 2.17.2 | Shared HTTP request implementation and public response type |
| Zap | 1.28.0 | Structured logging |
| OpenTelemetry / trace | 1.47.0 | Optional tracing interfaces |
| MSAL Go | 1.10.1 | Portable delegated and application authentication adapters |
| Azure Identity | 1.15.0 | Managed identity and workload identity adapters |
| govulncheck | 1.10.0 | Reachability-based vulnerability check, pinned as a Go tool |
| golangci-lint | 2.15.0 | Pinned CI lint binary; version-2 configuration |

Versions were resolved from the [Go module proxy](https://proxy.golang.org/) and their
published `go.mod` files checked against the selected compiler. The [MSAL release](https://github.com/AzureAD/microsoft-authentication-library-for-go/releases/tag/v1.10.1)
and [golangci-lint release](https://github.com/golangci/golangci-lint/releases/tag/v2.15.0)
provide upstream provenance. Resty v2 retains the agreed public response contract.

`internal/dependencycheck/dependencies_test.go` references the portable dependency types
so they compile on all supported platforms before production adapters exist. These are
compile gates, not authentication behaviour tests. They do not construct credentials,
start browser sign-in, read tenant environment variables or contact Microsoft services.
As adapters are implemented, their meaningful tests replace this temporary compile gate.

Do not interpret a pinned dependency as an implemented provider. The root provides local client construction and session lifecycle. AUTH and HTTP
tickets own provider and transport behaviour. Resource packages are created with their
implementations, without placeholder methods that pretend to support operations.

## Local checks and CI

```sh
go mod download
go mod verify
go mod tidy
go build ./...
go vet ./...
go test -count=1 -shuffle=on ./...
go test -race -count=1 -shuffle=on ./...
golangci-lint run ./...
golangci-lint fmt --diff
go tool govulncheck -test ./...
go run ./examples/baseline
```

Use golangci-lint 2.15.0 locally. The lint workflow runs analysis, applies configured
formatters, and fails if they change tracked files. Analysis does not auto-fix findings.
Tests and examples are included; the old template exclusions and zero error-exit code
have been removed. The dedicated lint workflow replaces the template's duplicate
Super-Linter Go checks, ensuring one pinned configuration is authoritative.

Every PR receives build, vet, ordinary tests, race checks and example execution on
Linux/macOS/Windows. CI verifies `go mod tidy` produces no tracked changes. The Linux
security job runs the module-pinned `go tool govulncheck -test ./...`; a finding fails
the job. SAST runs through gosec in golangci-lint. All ordinary jobs use read-only
repository permissions and require no tenant credentials.

The baseline workflow separately requires PowerShell as a discovery oracle. SDK build
and unit tests have no PowerShell dependency. Future tenant tests must be explicitly
opted into and remain outside ordinary `go test ./...` execution.

Release Please now uses the Go release type. Tag/release publication remains disabled
with `skip-github-release` until QUAL establishes complete coverage and release readiness;
release-note PRs can still be prepared. This is independent of upstream PowerShell
versioning. No first-supported-release claim follows from these foundation checks.
