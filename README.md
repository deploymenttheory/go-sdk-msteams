# Microsoft Teams Go SDK

A native Go SDK project targeting the management behaviour of MicrosoftTeams PowerShell **8.1.0**, with Linux, macOS and Windows support and portable CI/CD authentication.

**Status:** active development. The SDK includes portable authentication providers,
shared transport, Graph resources and Teams administration services. Complete
8.1.0 parity and release qualification remain outstanding. See the
[current coverage ledger](contracts/coverage/README.md) and
[GitHub backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6)
for command-level delivery and evidence limitations.

- [Agreed implementation plan](docs/implementation/plan.md)
- [Plan amendments and resource CRUD contract](docs/implementation/plan-amendments.md)
- [Published GitHub backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6)
- [Approved backlog preview](docs/implementation/backlog/README.md)
- [Publication verification](docs/implementation/backlog/publication/README.md)
- [Command coverage and ownership](docs/implementation/backlog/coverage.md)
- [Pinned PowerShell discovery evidence](contracts/powershell/8.1.0/README.md)
- [Reproduce the pinned baseline](tools/baseline/README.md)
- [Compare upstream contract changes](tools/contractdiff/README.md)
- [Discover releases and review candidate draft PRs](tools/releases/README.md)

The architecture follows the [Jamf Pro Go SDK](https://github.com/deploymenttheory/go-sdk-jamfpro-v2): a root client, shared transport, resource-specific services, typed models, examples and acceptance tests. PowerShell is a discovery and comparison tool, not a planned runtime dependency.

The first supported release requires complete implementation and contract-test coverage of the reconciled management baseline. Available tenant capabilities will also receive live conformance tests; unavailable live scenarios remain explicitly identified.

## Getting started

Use the Go version pinned in [go.mod](go.mod). Add the module to your application:

```sh
go get github.com/deploymenttheory/go-sdk-msteams
```

Configure an explicit authentication method using the
[authentication guide](docs/guides/authentication.md) and
[environment variable reference](docs/guides/client-configuration.md).
For example, `client-secret` requires `MSTEAMS_TENANT_ID`, `MSTEAMS_CLIENT_ID`
and `MSTEAMS_CLIENT_SECRET`. Supply credentials through your secret manager.
Authentication success does not establish permission to perform an operation.

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func main() {
    sdk, err := msteams.NewClientFromEnv()
    if err != nil {
        log.Fatal(err)
    }
    defer sdk.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    policy, _, err := sdk.TeamsAdminAPI.TeamsMeetingPolicies.GetByIdentity(ctx, "Global")
    if err != nil {
        log.Print(err)
        return
    }
    _ = policy // Typed policy; retain response metadata when diagnosing failures.
}
```

Construction performs local validation; the operation acquires its own backend token.
SDK consumers do not need PowerShell. For Graph operations, see the
[team API reference](msteams/graph_api/team/api_docs.md) and
[compiled examples](msteams/graph_api/team/example_test.go). The
[documentation index](docs/README.md) links both API families and their examples.

## Development

Requires the pinned Go version in [go.mod](go.mod) (currently 1.27.4). The module is
`github.com/deploymenttheory/go-sdk-msteams`; import the SDK entry point as
`github.com/deploymenttheory/go-sdk-msteams/msteams`.

```sh
go build ./...
go vet ./...
go test -race -count=1 -shuffle=on -run '^(TestUnit|Example)' ./...
go tool govulncheck -test ./...
go run ./examples/baseline
go run ./examples/client-construction
```

Both examples need no credentials. The client example constructs and closes a local session without signing in. The baseline example prints the target contract version;
it does not demonstrate an implemented Teams operation. See the
[foundation guide](docs/guides/development.md) and [client configuration guide](docs/guides/client-configuration.md) for pinned dependencies, CI gates and
package ownership. Offline quality CI runs on Linux, macOS and Windows without
tenant secrets. The separate [tenant acceptance workflow](tools/acceptance/README.md)
uses explicitly configured lab credentials and fixtures; ordinary unit checks do
not establish live qualification.

Repository layout, following the agreed Jamf composition:

```text
go.mod, go.sum           module metadata only; no Go source at repository root
msteams/                 SDK entry point, client composition and package tests
  client.go, doc.go       public client and package documentation
  client/                request boundary and shared transport
  auth/                  portable audience-aware providers
  config/, constants/    validated configuration and cloud definitions
  shared/                narrowly scoped wire/model primitives
  graph_api/<resource>/  typed Graph services
  teams_admin_api/<resource>/  typed administration services and local builders
  mocks/                offline doubles
tests/integration/lab/  tenant conformance, pinned comparisons and recovery
examples/                compiled usage examples
contracts/powershell/     immutable upstream evidence
contracts/coverage/       current command/parameter-set delivery ledger
internal/dependencycheck/ cross-platform dependency compilation gate
```

## Review and validate the backlog

```sh
python3 tools/backlog/build.py
python3 tools/backlog/validate.py
```

These commands operate locally and do not publish GitHub issues. The approved
publication preview is historical; GitHub issues/PRs and the coverage ledger track
current work. Do not regenerate the approved preview to represent implementation
progress.

```sh
python3 tools/coverage/ledger.py
# Release gate; expected to fail while implementation/qualification is incomplete:
python3 tools/coverage/ledger.py --require-complete
```

Current delivery work covers CALL, CORE, EXT, POL and LAB. Voice administration
(VOICE) is deferred from this work because the lab lacks licensing for its live
tests. This does not remove VOICE from the full-parity release requirements.
Unavailable live scenarios retain explicit limitations.

See [CONTRIBUTING.md](CONTRIBUTING.md) and the
[service development guide](docs/guides/service-development.md) for contribution
and verification conventions.

## Licence

See [LICENSE](LICENSE). This community project is not affiliated with Microsoft.
