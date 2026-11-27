# Microsoft Teams Go SDK

A native Go SDK for Microsoft Teams administration, with Linux, macOS and Windows
support and portable authentication for CI/CD.

**Status:** active development. The SDK includes authentication providers, shared
transport, Microsoft Graph resources and Teams administration services. See each
resource's API documentation and the [GitHub backlog](https://github.com/deploymenttheory/go-sdk-msteams/issues/6)
for supported functionality and remaining work.

- [Documentation index](docs/README.md)
- [Authentication and permissions](docs/guides/authentication.md)
- [Client configuration](docs/guides/client-configuration.md)
- [Service development](docs/guides/service-development.md)
- [Unit and real-tenant acceptance testing](docs/guides/testing.md)

The architecture follows the [Jamf Pro Go SDK](https://github.com/deploymenttheory/go-sdk-jamfpro-v2):
a root client, shared transport, resource-specific services, typed models,
examples and acceptance tests. SDK consumers do not need PowerShell. Existing
acceptance comparisons may use MicrosoftTeams PowerShell 8.1.0 as a test oracle.

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

Use the Go version pinned in [go.mod](go.mod). Import the SDK entry point as
`github.com/deploymenttheory/go-sdk-msteams/msteams`.

```sh
go build ./...
go vet ./...
go test -count=1 -shuffle=on -run '^(TestUnit|Example)' ./...
go test -race -count=1 -shuffle=on -run '^(TestUnit|Example)' ./...
go tool govulncheck -test ./...
go run ./examples/client-construction
```

Unit tests use checked-in JSON fixtures and in-memory HTTP responses. They run
without tenant credentials. The client-construction example is also credential-free.
Real-tenant tests use `TestAcc`, explicit lab configuration, dedicated fixtures
and independently verified cleanup. See the [acceptance guide](tools/acceptance/README.md).

```text
msteams/                 SDK entry point and resource services
  client/, auth/         shared transport and portable authentication
  graph_api/             typed Microsoft Graph resources
  teams_admin_api/       typed administration services and local builders
tests/integration/lab/   real-tenant acceptance and recovery
examples/               compiled usage examples
```

Current delivery work covers CALL, CORE, EXT, POL and LAB. Voice administration
is deferred because the lab lacks licensing for its live tests. Missing
permissions, licences or fixtures remain documented limitations; skipped tests
do not establish successful tenant behaviour.

Use ordinary unit tests, JSON response fixtures and real-tenant acceptance to
verify changes. Standalone contract research, command/parameter-set ledgers and
CLR emulation are not part of the development or release workflow.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution conventions.

## Licence

See [LICENSE](LICENSE). This community project is not affiliated with Microsoft.
